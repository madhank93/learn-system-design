import { describe, it, expect } from 'vitest';
import { RlSim } from './rl-traffic';
import { mulberry32 } from './rng';

const run = (sim: RlSim, ms: number, dt = 16) => {
  for (let t = 0; t < ms; t += dt) sim.step(dt);
};

describe('RlSim', () => {
  it('allows a burst up to capacity, then throttles to the refill rate', () => {
    // high arrival rate, small refill: most requests should be denied once the
    // initial bucket drains
    const sim = new RlSim({ ratePerSec: 30, capacity: 10, refillPerSec: 2 }, mulberry32(1));
    run(sim, 6000);
    expect(sim.allowed).toBeGreaterThan(0);
    expect(sim.denied).toBeGreaterThan(0);
    expect(sim.metrics.allowRate).toBeLessThan(0.5);
  });

  it('allows nearly everything when arrival rate stays under the refill rate', () => {
    const sim = new RlSim({ ratePerSec: 2, capacity: 10, refillPerSec: 5 }, mulberry32(2));
    run(sim, 6000);
    expect(sim.metrics.allowRate).toBeGreaterThan(0.9);
  });

  it('keeps tokens within [0, capacity]', () => {
    const sim = new RlSim({ ratePerSec: 12, capacity: 8, refillPerSec: 3 }, mulberry32(3));
    run(sim, 4000);
    expect(sim.tokens).toBeGreaterThanOrEqual(0);
    expect(sim.tokens).toBeLessThanOrEqual(8);
  });

  it('throughput of allowed requests approaches the refill rate under overload', () => {
    const sim = new RlSim({ ratePerSec: 50, capacity: 5, refillPerSec: 4 }, mulberry32(4));
    run(sim, 10000);
    // ~4 tokens/sec over ~10s ≈ 40 allowed (plus the initial 5), generous bounds
    expect(sim.allowed).toBeGreaterThan(25);
    expect(sim.allowed).toBeLessThan(60);
  });

  it('is deterministic for a given seed', () => {
    const a = new RlSim({ ratePerSec: 20 }, mulberry32(9));
    const b = new RlSim({ ratePerSec: 20 }, mulberry32(9));
    run(a, 4000);
    run(b, 4000);
    expect(a.allowed).toBe(b.allowed);
    expect(a.denied).toBe(b.denied);
  });
});
