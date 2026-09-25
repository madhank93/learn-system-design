import { describe, it, expect } from 'vitest';
import { ReplSim } from './repl-traffic';
import { mulberry32 } from './rng';

const run = (sim: ReplSim, ms: number, dt = 16) => {
  for (let t = 0; t < ms; t += dt) sim.step(dt);
};

describe('ReplSim', () => {
  it('higher lag produces more stale reads', () => {
    const lo = new ReplSim({ writeRate: 5, readRate: 12, lagMs: 100 }, mulberry32(1));
    const hi = new ReplSim({ writeRate: 5, readRate: 12, lagMs: 2500 }, mulberry32(1));
    run(lo, 9000);
    run(hi, 9000);
    expect(hi.metrics.staleRate).toBeGreaterThan(lo.metrics.staleRate);
  });

  it('no writes means no stale reads', () => {
    const sim = new ReplSim({ writeRate: 0, readRate: 12, lagMs: 2000 }, mulberry32(2));
    run(sim, 6000);
    expect(sim.staleReads).toBe(0);
    expect(sim.freshReads).toBeGreaterThan(0);
  });

  it('replicas converge to the primary version once writes stop', () => {
    const sim = new ReplSim({ writeRate: 6, readRate: 4, lagMs: 500 }, mulberry32(3));
    run(sim, 4000);
    sim.setConfig({ writeRate: 0 });
    run(sim, 4000); // let WAL drain
    expect(sim.metrics.maxLagVersions).toBe(0);
    expect(sim.metrics.version).toBeGreaterThan(0);
  });

  it('is deterministic for a given seed', () => {
    const a = new ReplSim({ writeRate: 4, readRate: 10, lagMs: 1000 }, mulberry32(9));
    const b = new ReplSim({ writeRate: 4, readRate: 10, lagMs: 1000 }, mulberry32(9));
    run(a, 5000);
    run(b, 5000);
    expect(a.freshReads).toBe(b.freshReads);
    expect(a.staleReads).toBe(b.staleReads);
  });
});
