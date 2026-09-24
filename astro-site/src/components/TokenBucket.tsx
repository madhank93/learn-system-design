import { useEffect, useRef, useState } from 'react';
import { refill } from '../sim-lib/token-bucket';

/**
 * Interactive token-bucket rate limiter. Tokens refill at a steady rate up to a
 * capacity; each request spends one. Watch an idle bucket bank tokens, fire a
 * burst that drains them, then get throttled to the refill rate. Replaces the
 * static "four algorithms" + burst_allowed D2 SVGs. Refill math lives in
 * sim-lib/token-bucket (pure + unit tested).
 */
export default function TokenBucket() {
  const [capacity, setCapacity] = useState(10);
  const [rate, setRate] = useState(2); // tokens/sec
  const [tokens, setTokens] = useState(10);
  const [allowed, setAllowed] = useState(0);
  const [denied, setDenied] = useState(0);
  const [auto, setAuto] = useState(false);
  const [log, setLog] = useState<{ n: number; ok: boolean }[]>([]);
  const tokensRef = useRef(tokens);
  tokensRef.current = tokens;
  const reqN = useRef(0);

  // continuous refill
  useEffect(() => {
    const step = 100; // ms
    const t = setInterval(() => {
      setTokens((x) => refill(x, capacity, rate, step));
    }, step);
    return () => clearInterval(t);
  }, [rate, capacity]);

  function send() {
    const n = ++reqN.current;
    if (tokensRef.current >= 1) {
      setTokens((x) => x - 1);
      setAllowed((a) => a + 1);
      setLog((l) => [{ n, ok: true }, ...l].slice(0, 8));
    } else {
      setDenied((d) => d + 1);
      setLog((l) => [{ n, ok: false }, ...l].slice(0, 8));
    }
  }

  useEffect(() => {
    if (!auto) return;
    const t = setInterval(send, 180);
    return () => clearInterval(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [auto]);

  function reset() {
    setAuto(false);
    setTokens(capacity);
    setAllowed(0);
    setDenied(0);
    setLog([]);
    reqN.current = 0;
  }

  const pct = Math.max(0, Math.min(100, (tokens / capacity) * 100));
  const whole = Math.floor(tokens + 1e-9);
  const total = allowed + denied;

  return (
    <div className="lsd-diagram not-content">
      <p className="lsd-hint">
        The bucket holds up to <strong>{capacity}</strong> tokens and refills <strong>{rate}/sec</strong>.
        Each request spends one; an empty bucket returns <strong>429</strong>. Let it fill, then hit{' '}
        <strong>Burst ×10</strong> to drain it — afterwards you're throttled to the refill rate.
      </p>

      <div className="lsd-controls">
        <button onClick={send}>▶ Send 1</button>
        <button onClick={() => { for (let i = 0; i < 10; i++) setTimeout(send, i * 30); }}>⚡ Burst ×10</button>
        <button onClick={() => setAuto((a) => !a)}>{auto ? '⏸ Stop steady' : '⏩ Steady traffic'}</button>
        <button onClick={reset}>↺ Reset</button>
      </div>
      <div className="lsd-controls">
        <span>capacity:</span>
        <span className="lsd-seg">
          {[5, 10, 20].map((c) => (
            <button key={c} aria-pressed={capacity === c} onClick={() => { setCapacity(c); setTokens((t) => Math.min(c, t)); }}>{c}</button>
          ))}
        </span>
        <span>refill/sec:</span>
        <span className="lsd-seg">
          {[1, 2, 5].map((r) => (
            <button key={r} aria-pressed={rate === r} onClick={() => setRate(r)}>{r}</button>
          ))}
        </span>
      </div>

      <div style={{ display: 'flex', gap: '1.5rem', alignItems: 'flex-end', marginTop: '0.75rem' }}>
        {/* bucket */}
        <div style={{ textAlign: 'center' }}>
          <div style={{ position: 'relative', width: 90, height: 160, border: '2px solid var(--sl-color-gray-4)', borderTop: 'none', borderRadius: '0 0 12px 12px', overflow: 'hidden', background: 'var(--sl-color-gray-6)' }}>
            <div style={{ position: 'absolute', bottom: 0, left: 0, right: 0, height: `${pct}%`, background: 'linear-gradient(#5fb0ff,#3a7bd5)', transition: 'height .1s linear' }} />
            <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 700, color: '#fff', textShadow: '0 1px 2px #000' }}>
              {whole}
            </div>
          </div>
          <div className="lsd-stat" style={{ marginTop: 4 }}>{whole} / {capacity} tokens</div>
        </div>

        <div style={{ flex: 1, minWidth: 200 }}>
          <p className="lsd-stat">Allowed: <strong style={{ color: '#2ECC71' }}>{allowed}</strong> · Denied (429): <strong style={{ color: '#E74C3C' }}>{denied}</strong></p>
          {total > 0 && (
            <div className="lsd-bar-track" style={{ margin: '4px 0 10px', height: 14 }}>
              <span style={{ display: 'inline-block', height: '100%', width: `${(allowed / total) * 100}%`, background: '#2ECC71' }} />
              <span style={{ display: 'inline-block', height: '100%', width: `${(denied / total) * 100}%`, background: '#E74C3C' }} />
            </div>
          )}
          <p className="lsd-stat"><strong>Recent</strong></p>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
            {log.length === 0 && <span className="lsd-stat" style={{ opacity: 0.5 }}>send a request…</span>}
            {log.map((e) => (
              <span key={e.n} style={{ fontSize: '0.72rem', padding: '1px 6px', borderRadius: 999, background: e.ok ? 'rgba(46,204,113,.18)' : 'rgba(231,76,60,.18)', color: e.ok ? '#2ECC71' : '#E74C3C', border: `1px solid ${e.ok ? '#2ECC71' : '#E74C3C'}` }}>
                #{e.n} {e.ok ? '200' : '429'}
              </span>
            ))}
          </div>
        </div>
      </div>
      <p className="lsd-caption">
        Idle a moment (bucket refills), then <strong>Burst ×10</strong>: the banked tokens absorb the spike,
        the rest get 429. Lower <strong>capacity</strong> to shrink the allowed burst; raise{' '}
        <strong>refill/sec</strong> to recover faster.
      </p>
    </div>
  );
}
