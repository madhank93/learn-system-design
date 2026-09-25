import { describe, it, expect } from 'vitest';
import { LbSim } from './lb-traffic';
import { mulberry32 } from './rng';

const run = (sim: LbSim, ms: number, dt = 16) => {
  for (let t = 0; t < ms; t += dt) sim.step(dt);
};

describe('LbSim', () => {
  it('conserves requests: spawned = served + dropped + active', () => {
    const sim = new LbSim({ ratePerSec: 20 }, 3, mulberry32(1));
    run(sim, 8000);
    expect(sim.spawned).toBe(sim.served + sim.dropped + sim.requests.length);
  });

  it('drops requests when arrival rate far exceeds capacity', () => {
    const sim = new LbSim({ ratePerSec: 40, algo: 'rr' }, 2, mulberry32(2));
    run(sim, 8000);
    expect(sim.dropped).toBeGreaterThan(0);
    expect(sim.metrics.dropRate).toBeGreaterThan(0);
    expect(sim.metrics.dropRate).toBeLessThanOrEqual(1);
  });

  it('rarely drops when capacity comfortably exceeds load', () => {
    const sim = new LbSim({ ratePerSec: 3, powerVariance: 0 }, 4, mulberry32(3));
    run(sim, 8000);
    expect(sim.metrics.dropRate).toBeLessThan(0.05);
  });

  it('reports monotonic latency percentiles', () => {
    const sim = new LbSim({ ratePerSec: 12 }, 3, mulberry32(4));
    run(sim, 8000);
    const { p50, p95, p99 } = sim.metrics;
    expect(p50).toBeLessThanOrEqual(p95);
    expect(p95).toBeLessThanOrEqual(p99);
  });

  it('round-robin spreads load across servers', () => {
    const sim = new LbSim({ ratePerSec: 6, algo: 'rr', powerVariance: 0 }, 3, mulberry32(5));
    run(sim, 12000);
    const totals = sim.servers.map((s) => s.total);
    const max = Math.max(...totals);
    const min = Math.min(...totals);
    expect(min).toBeGreaterThan(0);
    expect(max - min).toBeLessThanOrEqual(max * 0.5 + 2); // roughly balanced
  });

  it('is deterministic for a given seed', () => {
    const a = new LbSim({ ratePerSec: 15 }, 3, mulberry32(7));
    const b = new LbSim({ ratePerSec: 15 }, 3, mulberry32(7));
    run(a, 5000);
    run(b, 5000);
    expect(a.served).toBe(b.served);
    expect(a.dropped).toBe(b.dropped);
  });

  it('drops everything when all servers are unhealthy', () => {
    const sim = new LbSim({ ratePerSec: 10 }, 2, mulberry32(8));
    sim.toggleHealth(0);
    sim.toggleHealth(1);
    run(sim, 4000);
    expect(sim.served).toBe(0);
    expect(sim.dropped).toBeGreaterThan(0);
  });
});
