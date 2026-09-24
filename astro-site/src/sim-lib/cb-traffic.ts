/**
 * Continuous circuit-breaker traffic simulation (pure, no DOM).
 * Requests flow to a breaker that fronts a flaky downstream. Decisions reuse the
 * CircuitBreaker engine: CLOSED passes calls through (success/failure by the
 * downstream's failure probability), trips OPEN on a high failure rate (then
 * fails fast with zero downstream load), and after a cooldown probes HALF_OPEN.
 */
import { CircuitBreaker, type CBState, DEFAULT_CB } from './circuit-breaker';

export type CbStage = 'in' | 'ok' | 'fail' | 'block';

export type CbReq = {
  id: number;
  stage: CbStage;
  prog: number; // 0..1 within stage
  age: number;
};

export type CbEvent = { type: 'ok' | 'fail' | 'block' | 'trip' | 'probe' };

export type CbConfig = { ratePerSec: number; failRate: number }; // failRate 0..1 downstream error prob

const LEG_SEC = 0.45;

export class CbSim {
  cfg: CbConfig;
  cb: CircuitBreaker;
  reqs: CbReq[] = [];
  private rng: () => number;
  private nextId = 1;
  private spawnAcc = 0;
  private now = 0;
  passed = 0; // reached downstream
  failed = 0; // downstream errors
  blocked = 0; // failed fast while OPEN
  private dsTimes: number[] = []; // downstream-call timestamps (load window)

  constructor(cfg: Partial<CbConfig> = {}, rng: () => number = Math.random) {
    this.cfg = { ratePerSec: 8, failRate: 0, ...cfg };
    this.cb = new CircuitBreaker(DEFAULT_CB);
    this.rng = rng;
  }

  setConfig(p: Partial<CbConfig>): void {
    this.cfg = { ...this.cfg, ...p };
  }

  reset(): void {
    this.cb = new CircuitBreaker(DEFAULT_CB);
    this.reqs = [];
    this.passed = 0;
    this.failed = 0;
    this.blocked = 0;
    this.dsTimes = [];
    this.spawnAcc = 0;
    this.now = 0;
  }

  get state(): CBState {
    return this.cb.state;
  }

  get metrics() {
    return {
      state: this.cb.state,
      failureRate: this.cb.failureRate,
      dsLoad: this.dsTimes.length, // downstream calls in the last sim-second
      blocked: this.blocked,
      failed: this.failed,
      passed: this.passed,
      openLeft: this.cb.state === 'OPEN' ? Math.max(0, this.cb.openUntil - this.now) : 0,
    };
  }

  step(dtMs: number): CbEvent[] {
    const dt = dtMs / 1000;
    this.now += dtMs;
    const events: CbEvent[] = [];

    const before = this.cb.state;
    if (this.cb.tick(this.now) && this.cb.state === 'HALF_OPEN') events.push({ type: 'probe' });

    this.spawnAcc += this.cfg.ratePerSec * dt;
    while (this.spawnAcc >= 1) {
      this.spawnAcc -= 1;
      this.reqs.push({ id: this.nextId++, stage: 'in', prog: 0, age: 0 });
    }

    const alive: CbReq[] = [];
    for (const r of this.reqs) {
      r.age += dtMs;
      r.prog += dt / LEG_SEC;
      if (r.prog < 1) {
        alive.push(r);
        continue;
      }
      r.prog = 0;
      if (r.stage === 'in') {
        this.decide(r, events);
        alive.push(r);
      }
      // 'ok' | 'fail' | 'block' second leg completes → request leaves
    }
    this.reqs = alive;

    if (this.cb.state === 'OPEN' && before !== 'OPEN') events.push({ type: 'trip' });

    const cutoff = this.now - 1000;
    while (this.dsTimes.length && this.dsTimes[0] < cutoff) this.dsTimes.shift();
    return events;
  }

  private decide(r: CbReq, events: CbEvent[]): void {
    if (this.cb.state === 'OPEN') {
      this.cb.request(false, this.now); // no-op while OPEN (returns 'block')
      this.blocked++;
      r.stage = 'block';
      events.push({ type: 'block' });
      return;
    }
    // CLOSED or HALF_OPEN → actually call downstream
    const ok = this.rng() >= this.cfg.failRate;
    const wasOpenBefore = this.cb.state;
    this.cb.request(ok, this.now);
    this.dsTimes.push(this.now);
    if (ok) {
      this.passed++;
      r.stage = 'ok';
      events.push({ type: 'ok' });
    } else {
      this.failed++;
      r.stage = 'fail';
      events.push({ type: 'fail' });
    }
    void wasOpenBefore;
  }
}
