/**
 * Pure load-balancing logic, used by the LbSim traffic engine.
 * No DOM, no timers — just the routing decisions, so they can be unit tested
 * and reused by any renderer.
 */

export type Algo = 'rr' | 'least' | 'random' | 'sticky';
export type Server = { id: number; healthy: boolean; conns: number; total: number };

/** FNV-1a hash → unsigned 32-bit. Stable across runs for a given string. */
export function hashStr(s: string): number {
  let h = 2166136261;
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i);
    h = Math.imul(h, 16777619);
  }
  return h >>> 0;
}

/**
 * Sticky (consistent-hash style) target: hash the client to a starting slot,
 * then walk clockwise to the first healthy server. Killing a server re-pins
 * only the clients that were on it. Returns null when nothing is healthy.
 */
export function stickyTarget(client: string, list: Server[]): number | null {
  if (!list.length) return null;
  const start = hashStr(client) % list.length;
  for (let i = 0; i < list.length; i++) {
    const cand = list[(start + i) % list.length];
    if (cand.healthy) return cand.id;
  }
  return null;
}

/**
 * Pick a target server for `client` under `algo`. Round-robin is stateful, so
 * the caller threads `rr` (the rotating index) in and gets the next value back.
 * `rand` is injectable for deterministic tests of the random/`random` path.
 */
export function chooseTarget(
  algo: Algo,
  client: string,
  list: Server[],
  rr: number,
  rand: () => number = Math.random,
): { target: number | null; rr: number } {
  const healthy = list.filter((s) => s.healthy);
  if (!healthy.length) return { target: null, rr };

  if (algo === 'sticky') return { target: stickyTarget(client, list), rr };
  if (algo === 'random') return { target: healthy[Math.floor(rand() * healthy.length)].id, rr };
  if (algo === 'least') {
    return { target: healthy.reduce((a, b) => (b.conns < a.conns ? b : a)).id, rr };
  }
  // round robin over healthy servers
  const next = (rr + 1) % healthy.length;
  return { target: healthy[next].id, rr: next };
}
