import { useEffect, useRef, useState } from 'react';
import { type CBState as State, DEFAULT_CB, recordCall, shouldTrip } from '../sim-lib/circuit-breaker';

/**
 * Interactive circuit-breaker state machine. Toggle the downstream health, fire
 * traffic, and watch CLOSED → OPEN (trips on failure rate) → HALF_OPEN (probe
 * after cooldown) → CLOSED. Replaces the static state-machine + tripped D2 SVGs.
 * Trip/window logic lives in sim-lib/circuit-breaker (pure + unit tested).
 */
const { window: N, threshold: THRESHOLD, minCalls: MIN_CALLS, openMs: OPEN_MS } = DEFAULT_CB;

const COLOR: Record<State, string> = { CLOSED: '#2ECC71', OPEN: '#E74C3C', HALF_OPEN: '#E67E22' };
const DESC: Record<State, string> = {
  CLOSED: 'Normal — calls pass through, failures counted.',
  OPEN: 'Tripped — calls fail fast with a fallback; downstream gets a break.',
  HALF_OPEN: 'Probing — one test call decides whether to close or re-open.',
};

export default function CircuitBreaker() {
  const [state, setState] = useState<State>('CLOSED');
  const [win, setWin] = useState<boolean[]>([]);
  const [up, setUp] = useState(true);
  const [auto, setAuto] = useState(false);
  const [log, setLog] = useState<{ n: number; msg: string; kind: 'ok' | 'fail' | 'block' | 'state' }[]>([]);
  const [openLeft, setOpenLeft] = useState(0);
  const stateRef = useRef(state);
  stateRef.current = state;
  const winRef = useRef(win);
  winRef.current = win;
  const upRef = useRef(up);
  upRef.current = up;
  const openUntil = useRef(0);
  const reqN = useRef(0);

  function push(msg: string, kind: 'ok' | 'fail' | 'block' | 'state') {
    setLog((l) => [{ n: ++reqN.current, msg, kind }, ...l].slice(0, 8));
  }

  function trip() {
    openUntil.current = Date.now() + OPEN_MS;
    setState('OPEN');
    push('failure rate > 50% → OPEN (fail fast)', 'state');
  }

  function send() {
    const s = stateRef.current;
    if (s === 'OPEN') {
      push('OPEN → 503 fallback (no downstream call)', 'block');
      return;
    }
    const ok = upRef.current;
    if (s === 'HALF_OPEN') {
      if (ok) {
        setState('CLOSED');
        setWin([]);
        push('probe OK → CLOSED (recovered)', 'state');
      } else {
        trip();
        push('probe failed → OPEN again', 'fail');
      }
      return;
    }
    // CLOSED
    push(ok ? 'downstream 200 OK' : 'downstream 5xx / timeout', ok ? 'ok' : 'fail');
    const next = recordCall(winRef.current, ok, N);
    setWin(next);
    if (shouldTrip(next, MIN_CALLS, THRESHOLD)) trip();
  }

  // OPEN cooldown → HALF_OPEN
  useEffect(() => {
    if (state !== 'OPEN') {
      setOpenLeft(0);
      return;
    }
    const t = setInterval(() => {
      const left = openUntil.current - Date.now();
      setOpenLeft(Math.max(0, left));
      if (left <= 0) {
        clearInterval(t);
        setState('HALF_OPEN');
        push('cooldown elapsed → HALF_OPEN (probe next)', 'state');
      }
    }, 200);
    return () => clearInterval(t);
  }, [state]);

  useEffect(() => {
    if (!auto) return;
    const t = setInterval(send, 800);
    return () => clearInterval(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [auto]);

  function reset() {
    setAuto(false);
    setState('CLOSED');
    setWin([]);
    setLog([]);
    setOpenLeft(0);
    reqN.current = 0;
  }

  const fails = win.filter((x) => !x).length;
  const rate = win.length ? Math.round((fails / win.length) * 100) : 0;

  return (
    <div className="lsd-diagram not-content">
      <p className="lsd-hint">
        Toggle the downstream to <strong>Down</strong>, start <strong>Auto traffic</strong>, and watch the
        breaker trip <strong>OPEN</strong> once failures cross 50%. After a {OPEN_MS / 1000}s cooldown it goes{' '}
        <strong>HALF_OPEN</strong> and sends one probe — set downstream <strong>Up</strong> to let the probe
        close it.
      </p>

      <div className="lsd-controls">
        <span>downstream:</span>
        <span className="lsd-seg">
          <button aria-pressed={up} onClick={() => setUp(true)}>● Up</button>
          <button aria-pressed={!up} onClick={() => setUp(false)}>● Down</button>
        </span>
        <button onClick={send}>▶ Send request</button>
        <button onClick={() => setAuto((a) => !a)}>{auto ? '⏸ Stop traffic' : '⏩ Auto traffic'}</button>
        <button onClick={reset}>↺ Reset</button>
      </div>

      <div style={{ display: 'flex', gap: '1.5rem', flexWrap: 'wrap', alignItems: 'center', marginTop: '0.75rem' }}>
        {/* state badge */}
        <div style={{ textAlign: 'center', minWidth: 150 }}>
          <div style={{ display: 'inline-block', padding: '0.6rem 1.1rem', borderRadius: 10, fontWeight: 800, fontSize: '1.1rem', color: '#fff', background: COLOR[state], boxShadow: `0 0 0 4px ${COLOR[state]}33` }}>
            {state}
          </div>
          {state === 'OPEN' && (
            <div className="lsd-stat" style={{ marginTop: 6 }}>probe in {(openLeft / 1000).toFixed(1)}s</div>
          )}
          <p className="lsd-stat" style={{ marginTop: 6, maxWidth: 180 }}>{DESC[state]}</p>
        </div>

        <div style={{ flex: 1, minWidth: 220 }}>
          <p className="lsd-stat">
            Window (last {N}): <strong>{rate}%</strong> failures{win.length < MIN_CALLS ? ` (need ${MIN_CALLS} calls to evaluate)` : ''}
          </p>
          <div style={{ display: 'flex', gap: 4, margin: '4px 0 10px' }}>
            {Array.from({ length: N }).map((_, i) => {
              const v = win[i];
              return (
                <span key={i} style={{ width: 18, height: 18, borderRadius: 4, display: 'inline-flex', alignItems: 'center', justifyContent: 'center', fontSize: 11, color: '#fff', background: v === undefined ? 'var(--sl-color-gray-6)' : v ? '#2ECC71' : '#E74C3C' }}>
                  {v === undefined ? '' : v ? '✓' : '✗'}
                </span>
              );
            })}
          </div>
          <p className="lsd-stat"><strong>Event log</strong></p>
          <div style={{ fontFamily: 'var(--sl-font-mono, monospace)', fontSize: '0.76rem' }}>
            {log.length === 0 && <span className="lsd-stat" style={{ opacity: 0.5 }}>send a request…</span>}
            {log.map((e) => (
              <div key={e.n} style={{ color: e.kind === 'ok' ? '#2ECC71' : e.kind === 'fail' ? '#E74C3C' : e.kind === 'block' ? '#E67E22' : 'var(--sl-color-white)' }}>
                {e.kind === 'state' ? '⚡ ' : '· '}{e.msg}
              </div>
            ))}
          </div>
        </div>
      </div>
      <p className="lsd-caption">
        CLOSED counts failures over the last {N} calls. Above {THRESHOLD * 100}% it trips OPEN and fails fast
        (no downstream load). After cooldown it half-opens and risks exactly one probe — success closes it,
        failure re-opens. That's the whole pattern.
      </p>
    </div>
  );
}
