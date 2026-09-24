/**
 * Continuous rate-limiting traffic simulation (pure, no DOM).
 * Requests stream toward a token-bucket gate; each arrival spends a token
 * (allowed) or is rejected with 429 (denied). Wraps the existing TokenBucket
 * engine so the refill/consume rules stay in one place.
 */
import { TokenBucket } from './token-bucket';

export type RlState = 'flight' | 'pass' | 'deny';

export type RlReq = {
  id: number;
  flight: number; // 0..1 travel to the gate
  state: RlState;
  post: number; // 0..1 post-decision animation progress
  age: number;
};

export type RlEvent = { type: 'allow' | 'deny' };

export type RlConfig = { ratePerSec: number; capacity: number; refillPerSec: number };

const FLIGHT_SEC = 0.5;

export class RlSim {
  cfg: RlConfig;
  bucket: TokenBucket;
  reqs: RlReq[] = [];
  allowed = 0;
  denied = 0;
  private rng: () => number;
  private nextId = 1;
  private spawnAcc = 0;

  constructor(cfg: Partial<RlConfig> = {}, rng: () => number = Math.random) {
    this.cfg = { ratePerSec: 8, capacity: 10, refillPerSec: 2, ...cfg };
    this.bucket = new TokenBucket(this.cfg.capacity, this.cfg.refillPerSec);
    this.rng = rng;
  }

  setConfig(partial: Partial<RlConfig>): void {
    this.cfg = { ...this.cfg, ...partial };
    this.bucket.capacity = this.cfg.capacity;
    this.bucket.rate = this.cfg.refillPerSec;
    if (this.bucket.tokens > this.cfg.capacity) this.bucket.tokens = this.cfg.capacity;
  }

  reset(): void {
    this.bucket = new TokenBucket(this.cfg.capacity, this.cfg.refillPerSec);
    this.reqs = [];
    this.allowed = 0;
    this.denied = 0;
    this.spawnAcc = 0;
  }

  get tokens(): number {
    return this.bucket.tokens;
  }

  get metrics() {
    const total = this.allowed + this.denied;
    return {
      allowed: this.allowed,
      denied: this.denied,
      allowRate: total ? this.allowed / total : 1,
      tokens: this.bucket.tokens,
    };
  }

  step(dtMs: number): RlEvent[] {
    const dt = dtMs / 1000;
    this.bucket.refill(dtMs);
    const events: RlEvent[] = [];

    this.spawnAcc += this.cfg.ratePerSec * dt;
    while (this.spawnAcc >= 1) {
      this.spawnAcc -= 1;
      this.reqs.push({ id: this.nextId++, flight: 0, state: 'flight', post: 0, age: 0 });
    }

    const alive: RlReq[] = [];
    for (const r of this.reqs) {
      r.age += dtMs;
      if (r.state === 'flight') {
        r.flight += dt / FLIGHT_SEC;
        if (r.flight >= 1) {
          r.flight = 1;
          if (this.bucket.tryConsume()) {
            r.state = 'pass';
            this.allowed++;
            events.push({ type: 'allow' });
          } else {
            r.state = 'deny';
            this.denied++;
            events.push({ type: 'deny' });
          }
        }
        alive.push(r);
      } else {
        r.post += dt / 0.45;
        if (r.post < 1) alive.push(r);
      }
    }
    this.reqs = alive;
    return events;
  }
}
