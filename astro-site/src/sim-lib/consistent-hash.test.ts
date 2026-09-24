import { describe, it, expect } from 'vitest';
import {
  hashDeg,
  buildRing,
  ownerOf,
  loadDistribution,
  movedIndices,
} from './consistent-hash';

describe('hashDeg', () => {
  it('is deterministic and within 0–359', () => {
    const d = hashDeg('key_0');
    expect(d).toBe(hashDeg('key_0'));
    expect(d).toBeGreaterThanOrEqual(0);
    expect(d).toBeLessThan(360);
  });
});

describe('buildRing', () => {
  it('creates nodeCount * vnodes positions, sorted clockwise', () => {
    const ring = buildRing(3, 4);
    expect(ring.length).toBe(12);
    for (let i = 1; i < ring.length; i++) {
      expect(ring[i].deg).toBeGreaterThanOrEqual(ring[i - 1].deg);
    }
  });
});

describe('ownerOf', () => {
  const ring = [
    { deg: 90, node: 0 },
    { deg: 180, node: 1 },
    { deg: 270, node: 2 },
  ];
  it('assigns to the first position clockwise', () => {
    expect(ownerOf(100, ring)).toBe(1);
    expect(ownerOf(180, ring)).toBe(1);
  });
  it('wraps past the last position to the first', () => {
    expect(ownerOf(300, ring)).toBe(0);
  });
  it('returns -1 for an empty ring', () => {
    expect(ownerOf(10, [])).toBe(-1);
  });
});

describe('loadDistribution', () => {
  it('counts keys per node id', () => {
    expect(loadDistribution([0, 0, 1, 2, 2, 2], 3)).toEqual([2, 1, 3]);
  });
});

describe('movedIndices', () => {
  it('flags only keys whose owner changes', () => {
    const before = buildRing(3, 1);
    const after = buildRing(4, 1);
    const keys = Array.from({ length: 20 }, (_, i) => hashDeg(`key_${i}`));
    const moved = movedIndices(keys, before, after);
    // consistent hashing moves far fewer than all keys when adding one node
    expect(moved.size).toBeLessThan(keys.length);
  });
  it('moves nothing when topology is unchanged', () => {
    const ring = buildRing(3, 1);
    const keys = [10, 100, 200];
    expect(movedIndices(keys, ring, ring).size).toBe(0);
  });
});
