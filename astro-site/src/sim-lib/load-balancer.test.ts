import { describe, it, expect } from 'vitest';
import { hashStr, stickyTarget, chooseTarget, type Server } from './load-balancer';

const mk = (ids: number[], down: number[] = []): Server[] =>
  ids.map((id) => ({ id, healthy: !down.includes(id), conns: 0, total: 0 }));

describe('hashStr', () => {
  it('is deterministic and unsigned 32-bit', () => {
    const h = hashStr('alice');
    expect(h).toBe(hashStr('alice'));
    expect(h).toBeGreaterThanOrEqual(0);
    expect(h).toBeLessThanOrEqual(0xffffffff);
  });
  it('differs for different inputs', () => {
    expect(hashStr('alice')).not.toBe(hashStr('bob'));
  });
});

describe('stickyTarget', () => {
  it('pins a client to the same server across calls', () => {
    const list = mk([0, 1, 2]);
    expect(stickyTarget('carol', list)).toBe(stickyTarget('carol', list));
  });
  it('re-pins only to a healthy server', () => {
    const list = mk([0, 1, 2]);
    const home = stickyTarget('carol', list)!;
    const downed = mk([0, 1, 2], [home]);
    const next = stickyTarget('carol', downed);
    expect(next).not.toBe(home);
    expect(downed.find((s) => s.id === next)!.healthy).toBe(true);
  });
  it('returns null when nothing is healthy', () => {
    expect(stickyTarget('x', mk([0, 1], [0, 1]))).toBeNull();
  });
});

describe('chooseTarget', () => {
  it('round-robin cycles through healthy servers', () => {
    const list = mk([0, 1, 2]);
    let rr = 0;
    const seq: (number | null)[] = [];
    for (let i = 0; i < 4; i++) {
      const r = chooseTarget('rr', 'c', list, rr);
      rr = r.rr;
      seq.push(r.target);
    }
    expect(seq).toEqual([1, 2, 0, 1]);
  });
  it('round-robin skips unhealthy servers', () => {
    const list = mk([0, 1, 2], [1]);
    let rr = 0;
    const seen = new Set<number | null>();
    for (let i = 0; i < 5; i++) {
      const r = chooseTarget('rr', 'c', list, rr);
      rr = r.rr;
      seen.add(r.target);
    }
    expect(seen.has(1)).toBe(false);
  });
  it('least-connections picks the idlest server', () => {
    const list = mk([0, 1, 2]);
    list[0].conns = 5;
    list[1].conns = 1;
    list[2].conns = 3;
    expect(chooseTarget('least', 'c', list, 0).target).toBe(1);
  });
  it('random uses the injected RNG', () => {
    const list = mk([0, 1, 2]);
    expect(chooseTarget('random', 'c', list, 0, () => 0).target).toBe(0);
    expect(chooseTarget('random', 'c', list, 0, () => 0.99).target).toBe(2);
  });
  it('returns null when all servers are down', () => {
    expect(chooseTarget('rr', 'c', mk([0, 1], [0, 1]), 0).target).toBeNull();
  });
});
