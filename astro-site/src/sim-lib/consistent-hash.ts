/**
 * Pure consistent-hashing ring math, split from the React view
 * (ConsistentHashingRing). Positions are in degrees (0–360) standing in for the
 * 0 → 2³²−1 hash space; ownership is the first ring position clockwise.
 */

export type VPos = { deg: number; node: number };

/** FNV-1a hash mapped onto the 0–359° ring. */
export function hashDeg(s: string): number {
  let h = 2166136261;
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i);
    h = Math.imul(h, 16777619);
  }
  return (h >>> 0) % 360;
}

/** Place `nodeCount` servers × `vnodes` virtual nodes, sorted clockwise. */
export function buildRing(nodeCount: number, vnodes: number): VPos[] {
  const ring: VPos[] = [];
  for (let n = 0; n < nodeCount; n++)
    for (let v = 0; v < vnodes; v++) ring.push({ deg: hashDeg(`S${n}#${v}`), node: n });
  return ring.sort((a, b) => a.deg - b.deg);
}

/** Owner of a key = first ring position at or clockwise of the key, wrapping. */
export function ownerOf(keyDeg: number, ring: VPos[]): number {
  for (const v of ring) if (v.deg >= keyDeg) return v.node;
  return ring.length ? ring[0].node : -1;
}

/** Count keys owned per node id (index = node id). */
export function loadDistribution(ownership: number[], nodeCount: number): number[] {
  const counts = new Array(nodeCount).fill(0);
  ownership.forEach((o) => o >= 0 && o < nodeCount && counts[o]++);
  return counts;
}

/** Indices of keys whose owner changes when moving from `before` to `after`. */
export function movedIndices(keys: number[], before: VPos[], after: VPos[]): Set<number> {
  const moved = new Set<number>();
  keys.forEach((k, i) => {
    if (ownerOf(k, before) !== ownerOf(k, after)) moved.add(i);
  });
  return moved;
}
