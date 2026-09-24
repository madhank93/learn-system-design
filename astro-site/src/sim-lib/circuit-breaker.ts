/**
 * Pure circuit-breaker logic, split from the React view (CircuitBreaker).
 * CLOSED → OPEN (trips on failure rate) → HALF_OPEN (probe after cooldown) → CLOSED.
 */

export type CBState = 'CLOSED' | 'OPEN' | 'HALF_OPEN';

export type CBConfig = {
  /** rolling window size (last N calls) */
  window: number;
  /** failure fraction above which the breaker trips */
  threshold: number;
  /** minimum calls in-window before the rate is evaluated */
  minCalls: number;
  /** OPEN cooldown before a probe is allowed (ms) */
  openMs: number;
};

export const DEFAULT_CB: CBConfig = { window: 10, threshold: 0.5, minCalls: 5, openMs: 5000 };

/** Append a call result, keeping only the last `window` entries. */
export function recordCall(win: boolean[], ok: boolean, window: number): boolean[] {
  return [...win, ok].slice(-window);
}

/** Has the in-window failure rate crossed the trip threshold? */
export function shouldTrip(win: boolean[], minCalls: number, threshold: number): boolean {
  if (win.length < minCalls) return false;
  const fails = win.filter((x) => !x).length;
  return fails / win.length > threshold;
}

export type CBOutcome = 'pass-ok' | 'pass-fail' | 'block' | 'probe-ok' | 'probe-fail';

/** Canonical engine — drives tests and is reusable outside React. */
export class CircuitBreaker {
  state: CBState = 'CLOSED';
  win: boolean[] = [];
  openUntil = 0;

  constructor(public cfg: CBConfig = DEFAULT_CB) {}

  /** Process one request given downstream health `ok` at time `now` (ms). */
  request(ok: boolean, now: number): CBOutcome {
    if (this.state === 'OPEN') return 'block';
    if (this.state === 'HALF_OPEN') {
      if (ok) {
        this.state = 'CLOSED';
        this.win = [];
        return 'probe-ok';
      }
      this.trip(now);
      return 'probe-fail';
    }
    // CLOSED
    this.win = recordCall(this.win, ok, this.cfg.window);
    if (shouldTrip(this.win, this.cfg.minCalls, this.cfg.threshold)) this.trip(now);
    return ok ? 'pass-ok' : 'pass-fail';
  }

  trip(now: number): void {
    this.openUntil = now + this.cfg.openMs;
    this.state = 'OPEN';
  }

  /** Advance the clock; OPEN → HALF_OPEN once the cooldown elapses. */
  tick(now: number): boolean {
    if (this.state === 'OPEN' && now >= this.openUntil) {
      this.state = 'HALF_OPEN';
      return true;
    }
    return false;
  }

  get failureRate(): number {
    return this.win.length ? this.win.filter((x) => !x).length / this.win.length : 0;
  }
}
