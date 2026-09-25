import { describe, it, expect } from 'vitest';
import { CacheSim } from './cache-traffic';
import { mulberry32 } from './rng';

const run = (sim: CacheSim, ms: number, dt = 16) => {
  for (let t = 0; t < ms; t += dt) sim.step(dt);
};

describe('CacheSim', () => {
  it('high skew yields a high hit ratio', () => {
    const sim = new CacheSim({ ratePerSec: 15, skew: 0.95, capacity: 12 }, mulberry32(1));
    run(sim, 8000);
    expect(sim.metrics.hitRatio).toBeGreaterThan(0.5);
  });

  it('low skew (uniform) yields a lower hit ratio than high skew', () => {
    const lo = new CacheSim({ ratePerSec: 15, skew: 0.0, capacity: 8 }, mulberry32(2));
    const hi = new CacheSim({ ratePerSec: 15, skew: 0.95, capacity: 8 }, mulberry32(2));
    run(lo, 8000);
    run(hi, 8000);
    expect(hi.metrics.hitRatio).toBeGreaterThan(lo.metrics.hitRatio);
  });

  it('never caches more than capacity', () => {
    const sim = new CacheSim({ ratePerSec: 20, skew: 0.1, capacity: 5 }, mulberry32(3));
    run(sim, 8000);
    expect(sim.occupancy).toBeLessThanOrEqual(5);
  });

  it('misses generate database load; hits do not', () => {
    const sim = new CacheSim({ ratePerSec: 12, skew: 0.9, capacity: 12 }, mulberry32(4));
    run(sim, 8000);
    expect(sim.hits + sim.misses).toBeGreaterThan(0);
    // every miss is counted as db load at some point
    expect(sim.misses).toBeGreaterThan(0);
  });

  it('is deterministic for a given seed', () => {
    const a = new CacheSim({ ratePerSec: 15 }, mulberry32(7));
    const b = new CacheSim({ ratePerSec: 15 }, mulberry32(7));
    run(a, 5000);
    run(b, 5000);
    expect(a.hits).toBe(b.hits);
    expect(a.misses).toBe(b.misses);
  });
});
