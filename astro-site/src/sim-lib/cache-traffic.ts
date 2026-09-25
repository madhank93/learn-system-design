/**
 * Continuous cache-aside traffic simulation (pure, no DOM).
 * Requests for keys arrive at a rate; each checks an LRU+TTL cache. A hit
 * returns fast; a miss falls through to the database (slow), then populates the
 * cache. Skew concentrates traffic on hot keys (raising hit ratio); capacity and
 * TTL bound what stays cached. The renderer maps request stages → pixels.
 */

export type CacheStage = 'in' | 'hitback' | 'todb' | 'dbback' | 'out';

export type CacheReq = {
  id: number;
  key: number;
  stage: CacheStage;
  prog: number; // 0..1 within current stage
  age: number; // ms since spawn
  hit?: boolean;
};

export type CacheEvent = { type: 'hit' | 'miss' | 'evict' };

export type CacheConfig = {
  ratePerSec: number;
  capacity: number; // cache slots
  skew: number; // 0..1 hot-key concentration
  ttlMs: number;
};

const KEY_COUNT = 50;
const HOT = Math.max(1, Math.round(KEY_COUNT * 0.15));
const LEG_SEC = 0.4; // travel time per hop
const LAT_WINDOW = 200;

export class CacheSim {
  cfg: CacheConfig;
  reqs: CacheReq[] = [];
  private rng: () => number;
  private nextId = 1;
  private spawnAcc = 0;
  private now = 0;
  private cache = new Map<number, number>(); // key -> expiry (ms)
  private lru: number[] = []; // keys, front = least-recent
  private latencies: number[] = [];
  private missTimes: number[] = [];
  hits = 0;
  misses = 0;

  constructor(cfg: Partial<CacheConfig> = {}, rng: () => number = Math.random) {
    this.cfg = { ratePerSec: 8, capacity: 8, skew: 0.6, ttlMs: 6000, ...cfg };
    this.rng = rng;
  }

  setConfig(p: Partial<CacheConfig>): void {
    this.cfg = { ...this.cfg, ...p };
    this.evictToCapacity(null);
  }

  reset(): void {
    this.reqs = [];
    this.cache.clear();
    this.lru = [];
    this.latencies = [];
    this.missTimes = [];
    this.hits = 0;
    this.misses = 0;
    this.spawnAcc = 0;
    this.now = 0;
  }

  get occupancy(): number {
    return this.cache.size;
  }

  get metrics() {
    const total = this.hits + this.misses;
    return {
      hitRatio: total ? this.hits / total : 0,
      dbLoad: this.missTimes.length, // misses in the last sim-second
      occupancy: this.cache.size,
      p50: this.percentile(50),
      p95: this.percentile(95),
      latencies: this.latencies,
    };
  }

  percentile(p: number): number {
    if (!this.latencies.length) return 0;
    const s = [...this.latencies].sort((a, b) => a - b);
    return s[Math.min(s.length - 1, Math.floor((p / 100) * s.length))];
  }

  private pickKey(): number {
    if (this.rng() < this.cfg.skew) return Math.floor(this.rng() * HOT);
    return Math.floor(this.rng() * KEY_COUNT);
  }

  private touch(key: number): void {
    const i = this.lru.indexOf(key);
    if (i >= 0) this.lru.splice(i, 1);
    this.lru.push(key);
  }

  private evictToCapacity(events: CacheEvent[] | null): void {
    while (this.cache.size > this.cfg.capacity && this.lru.length) {
      const k = this.lru.shift()!;
      if (this.cache.delete(k) && events) events.push({ type: 'evict' });
    }
  }

  step(dtMs: number): CacheEvent[] {
    const dt = dtMs / 1000;
    this.now += dtMs;
    const events: CacheEvent[] = [];

    // expire TTLs lazily handled on lookup; spawn arrivals
    this.spawnAcc += this.cfg.ratePerSec * dt;
    while (this.spawnAcc >= 1) {
      this.spawnAcc -= 1;
      this.reqs.push({ id: this.nextId++, key: this.pickKey(), stage: 'in', prog: 0, age: 0 });
    }

    const alive: CacheReq[] = [];
    for (const r of this.reqs) {
      r.age += dtMs;
      r.prog += dt / LEG_SEC;
      if (r.prog < 1) {
        alive.push(r);
        continue;
      }
      r.prog = 0;
      this.advance(r, events, alive);
    }
    this.reqs = alive;

    const cutoff = this.now - 1000;
    while (this.missTimes.length && this.missTimes[0] < cutoff) this.missTimes.shift();
    return events;
  }

  private advance(r: CacheReq, events: CacheEvent[], alive: CacheReq[]): void {
    switch (r.stage) {
      case 'in': {
        const exp = this.cache.get(r.key);
        const hit = exp !== undefined && exp > this.now;
        if (hit) {
          this.hits++;
          this.touch(r.key);
          r.hit = true;
          r.stage = 'hitback';
          events.push({ type: 'hit' });
        } else {
          if (exp !== undefined) this.cache.delete(r.key); // expired
          this.misses++;
          this.missTimes.push(this.now);
          r.hit = false;
          r.stage = 'todb';
          events.push({ type: 'miss' });
        }
        alive.push(r);
        break;
      }
      case 'todb':
        r.stage = 'dbback';
        alive.push(r);
        break;
      case 'dbback':
        // populate cache on the way back
        this.cache.set(r.key, this.now + this.cfg.ttlMs);
        this.touch(r.key);
        this.evictToCapacity(events);
        r.stage = 'out';
        alive.push(r);
        break;
      case 'hitback':
      case 'out':
        this.recordDone(r);
        break;
    }
  }

  private recordDone(r: CacheReq): void {
    this.latencies.push(r.age);
    if (this.latencies.length > LAT_WINDOW) this.latencies.shift();
  }
}
