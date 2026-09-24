import { useEffect, useRef, useState } from 'react';
import { type Algo, type Server, chooseTarget, stickyTarget } from '../sim-lib/load-balancer';

/**
 * Interactive load-balancer simulator. Pick an algorithm, fire traffic, and
 * watch where requests land — active connections build and drain, you can take
 * a server down (health check) and see traffic reroute, and Sticky mode shows
 * each client pinned to one server. Replaces the static "four algorithms" +
 * health-check + sticky-failover D2 SVGs with one thing you can drive.
 * Routing decisions live in sim-lib/load-balancer (pure + unit tested).
 */

const PALETTE = ['#4F9DF7', '#2ECC71', '#E67E22', '#A66CFF', '#E74C3C', '#1ABC9C'];
const CLIENTS = ['alice', 'bob', 'carol', 'dave', 'erin'];

const ALGO_INFO: Record<Algo, { label: string; blurb: string }> = {
  rr: { label: 'Round Robin', blurb: 'Cycle through healthy servers in order. Simple and fair — but ignores how busy each server actually is.' },
  least: { label: 'Least Connections', blurb: 'Send each request to the server with the fewest active connections. Adapts when request durations vary.' },
  random: { label: 'Random', blurb: 'Pick a healthy server at random. Converges to even over time; zero state to track.' },
  sticky: { label: 'Sticky (hash)', blurb: 'Hash the client key so the same client always lands on the same server — needed for in-memory sessions. Killing a server re-pins only its clients.' },
};

export default function LoadBalancerSim() {
  const [algo, setAlgo] = useState<Algo>('rr');
  const [servers, setServers] = useState<Server[]>(() =>
    [0, 1, 2].map((id) => ({ id, healthy: true, conns: 0, total: 0 })),
  );
  const [auto, setAuto] = useState(false);
  const [log, setLog] = useState<{ n: number; client: string; server: number | null }[]>([]);
  const [flash, setFlash] = useState<number | null>(null);
  const rr = useRef(0);
  const reqN = useRef(0);
  const serversRef = useRef(servers);
  serversRef.current = servers;

  function send() {
    const client = CLIENTS[Math.floor(Math.random() * CLIENTS.length)];
    const list = serversRef.current;
    const picked = chooseTarget(algo, client, list, rr.current);
    rr.current = picked.rr;
    const target = picked.target;
    const n = ++reqN.current;
    setLog((l) => [{ n, client, server: target }, ...l].slice(0, 6));
    if (target === null) return;
    setFlash(target);
    setTimeout(() => setFlash((f) => (f === target ? null : f)), 350);
    setServers((prev) => prev.map((s) => (s.id === target ? { ...s, conns: s.conns + 1, total: s.total + 1 } : s)));
    // request finishes after a variable duration → connection drains
    const dur = 1100 + Math.random() * 2400;
    setTimeout(() => {
      setServers((prev) => prev.map((s) => (s.id === target ? { ...s, conns: Math.max(0, s.conns - 1) } : s)));
    }, dur);
  }

  useEffect(() => {
    if (!auto) return;
    const t = setInterval(send, 600);
    return () => clearInterval(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [auto, algo]);

  function toggleHealth(id: number) {
    setServers((prev) => prev.map((s) => (s.id === id ? { ...s, healthy: !s.healthy, conns: s.healthy ? 0 : s.conns } : s)));
  }
  function addServer() {
    setServers((prev) => (prev.length >= 6 ? prev : [...prev, { id: prev.length, healthy: true, conns: 0, total: 0 }]));
  }
  function reset() {
    setAuto(false);
    rr.current = 0;
    reqN.current = 0;
    setLog([]);
    setServers([0, 1, 2].map((id) => ({ id, healthy: true, conns: 0, total: 0 })));
  }

  const totalServed = servers.reduce((a, s) => a + s.total, 0);
  const maxConn = Math.max(1, ...servers.map((s) => s.conns));

  return (
    <div className="lsd-diagram not-content">
      <div className="lsd-controls">
        <span className="lsd-seg">
          {(Object.keys(ALGO_INFO) as Algo[]).map((a) => (
            <button key={a} aria-pressed={algo === a} onClick={() => setAlgo(a)}>
              {ALGO_INFO[a].label}
            </button>
          ))}
        </span>
      </div>
      <p className="lsd-hint">{ALGO_INFO[algo].blurb}</p>

      <div className="lsd-controls">
        <button onClick={send}>▶ Send 1 request</button>
        <button onClick={() => setAuto((a) => !a)}>{auto ? '⏸ Stop traffic' : '⏩ Auto traffic'}</button>
        <button onClick={addServer} disabled={servers.length >= 6}>＋ Add server</button>
        <button onClick={reset}>↺ Reset</button>
        <span className="lsd-stat" style={{ marginLeft: 'auto' }}>Total served: <strong>{totalServed}</strong></span>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: `repeat(${Math.min(servers.length, 3)}, 1fr)`, gap: '0.6rem', marginTop: '0.5rem' }}>
        {servers.map((s) => {
          const color = PALETTE[s.id % PALETTE.length];
          const pct = totalServed ? Math.round((s.total / totalServed) * 100) : 0;
          return (
            <div
              key={s.id}
              className={flash === s.id ? 'lsd-card-flash' : ''}
              style={{
                border: `1px solid ${s.healthy ? 'var(--sl-color-gray-5)' : 'var(--sl-color-red)'}`,
                borderRadius: '0.5rem',
                padding: '0.6rem',
                opacity: s.healthy ? 1 : 0.55,
                background: 'var(--sl-color-black)',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 6 }}>
                <span className="lsd-dot" style={{ background: color }} />
                <strong style={{ color: 'var(--sl-color-white)' }}>Server {s.id}</strong>
                <button onClick={() => toggleHealth(s.id)} style={{ marginLeft: 'auto', fontSize: '0.72rem', padding: '0.15rem 0.45rem' }}>
                  {s.healthy ? '⏻ Take down' : '↺ Revive'}
                </button>
              </div>
              <div className="lsd-stat">active connections: <strong>{s.conns}</strong></div>
              <div className="lsd-bar-track" style={{ margin: '3px 0 6px' }}>
                <span className="lsd-bar-fill" style={{ width: `${(s.conns / maxConn) * 100}%`, background: color }} />
              </div>
              <div className="lsd-stat">served: <strong>{s.total}</strong> ({pct}%)</div>
              {!s.healthy && <div className="lsd-stat" style={{ color: 'var(--sl-color-red)' }}>● failed health check</div>}
            </div>
          );
        })}
      </div>

      {algo === 'sticky' && (
        <div style={{ marginTop: '0.75rem' }}>
          <p className="lsd-stat"><strong>Client → server pinning</strong> (kill a server to watch only its clients re-pin):</p>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6, marginTop: 4 }}>
            {CLIENTS.map((c) => {
              const t = stickyTarget(c, servers);
              return (
                <span key={c} style={{ display: 'inline-flex', alignItems: 'center', gap: 5, border: '1px solid var(--sl-color-gray-5)', borderRadius: 999, padding: '2px 8px', fontSize: '0.8rem' }}>
                  {c}
                  <span style={{ opacity: 0.6 }}>→</span>
                  <span className="lsd-dot" style={{ background: t === null ? '#888' : PALETTE[t % PALETTE.length] }} />
                  {t === null ? 'none' : `S${t}`}
                </span>
              );
            })}
          </div>
        </div>
      )}

      <div style={{ marginTop: '0.75rem' }}>
        <p className="lsd-stat"><strong>Recent routes</strong></p>
        <div style={{ fontFamily: 'var(--sl-font-mono, monospace)', fontSize: '0.78rem', color: 'var(--sl-color-gray-2)' }}>
          {log.length === 0 && <div style={{ opacity: 0.5 }}>send a request…</div>}
          {log.map((e) => (
            <div key={e.n}>
              req #{e.n} ({e.client}) →{' '}
              {e.server === null ? <span style={{ color: 'var(--sl-color-red)' }}>503 no healthy server</span> : <span style={{ color: PALETTE[e.server % PALETTE.length] }}>Server {e.server}</span>}
            </div>
          ))}
        </div>
      </div>
      <p className="lsd-caption">
        Tip: switch to <strong>Round Robin</strong>, start Auto traffic, then take a server down — traffic
        rebalances. Switch to <strong>Least Connections</strong> to see it favor idle servers. <strong>Sticky</strong>
        {' '}keeps each client on one server until that server fails.
      </p>
    </div>
  );
}
