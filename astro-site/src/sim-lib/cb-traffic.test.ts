import { describe, it, expect } from 'vitest';
import { CbSim } from './cb-traffic';
import { mulberry32 } from './rng';

const run = (sim: CbSim, ms: number, dt = 16) => {
  for (let t = 0; t < ms; t += dt) sim.step(dt);
};

describe('CbSim', () => {
  it('stays CLOSED and passes everything when downstream is healthy', () => {
    const sim = new CbSim({ ratePerSec: 10, failRate: 0 }, mulberry32(1));
    run(sim, 5000);
    expect(sim.state).toBe('CLOSED');
    expect(sim.passed).toBeGreaterThan(0);
    expect(sim.failed).toBe(0);
    expect(sim.blocked).toBe(0);
  });

  it('trips OPEN and fails fast when downstream is failing', () => {
    const sim = new CbSim({ ratePerSec: 12, failRate: 1 }, mulberry32(2));
    run(sim, 6000);
    expect(sim.blocked).toBeGreaterThan(0);
    // most requests should be blocked (fast-failed), not sent downstream
    expect(sim.blocked).toBeGreaterThan(sim.failed);
  });

  it('limits downstream load once tripped', () => {
    const sim = new CbSim({ ratePerSec: 20, failRate: 1 }, mulberry32(3));
    run(sim, 6000);
    // downstream calls per second stay far below the arrival rate
    expect(sim.metrics.dsLoad).toBeLessThan(5);
  });

  it('is deterministic for a given seed', () => {
    const a = new CbSim({ ratePerSec: 15, failRate: 0.7 }, mulberry32(8));
    const b = new CbSim({ ratePerSec: 15, failRate: 0.7 }, mulberry32(8));
    run(a, 5000);
    run(b, 5000);
    expect(a.passed).toBe(b.passed);
    expect(a.blocked).toBe(b.blocked);
  });
});
