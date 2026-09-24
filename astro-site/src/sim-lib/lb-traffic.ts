/**
 * Continuous load-balancing traffic simulation (pure, no DOM).
 *
 * Requests spawn at `ratePerSec`, fly to a server chosen by the active algorithm
 * (reusing `chooseTarget` from load-balancer.ts), wait in a bounded queue, and
 * are served at a rate set by each server's power. Requests are dropped when a
 * queue is full or after they time out. The renderer reads `servers` / live
 * `requests` (with normalized positions + age) each frame and maps them to
 * pixels; this module owns all timing/state so it stays unit-testable.
 */
import { chooseTarget, type Algo, type Server } from './load-balancer';

export type ReqState = 'flight' | 'queue' | 'serving';

export type Req = {
  id: number;
  server: number;
  state: ReqState;
  flight: number; // 0..1 travel progress
  cost: number; // remaining work units
  cost0: number; // initial work (for clock-shrink rendering)
  age: number; // ms since spawn (== wait time until served)
  timedOut: boolean;
};

export type LbEvent =
  | { type: 'done'; server: number; latency: number }
  | { type: 'drop'; server: number; reason: 'full' | 'timeout' | 'unhealthy' };

export type LbConfig = {
  ratePerSec: number;
  powerVariance: number; // 0..1 spread of per-server power
  costVariance: number; // 0..1 spread of per-request cost
  algo: Algo;
};

type Srv = { id: number; healthy: boolean; serving: Req | null; queue: Req[]; total: number; dropped: number };

const FLIGHT_SEC = 0.55;
const TIMEOUT_MS = 5000;
const QUEUE_CAP = 8;
const BASE_POWER = 3; // work units / sec per server at variance 0
const BASE_COST = 1; // work units per request at variance 0
const CLIENTS = ['alice', 'bob', 'carol', 'dave', 'erin'];
const LAT_WINDOW = 200;

export class LbSim {
  cfg: LbConfig;
  servers: Srv[] = [];
  private rng: () => number;
  private rr = 0;
  private now = 0;
  private nextId = 1;
  private spawnAcc = 0;
  private clientIdx = 0;
  private latencies: number[] = [];
  private doneTimes: number[] = []; // sim-ms timestamps for throughput window
  served = 0;
  dropped = 0;

  /** total requests ever spawned (each gets an id) — for conservation checks. */
  get spawned(): number {
    return this.nextId - 1;
  }

  constructor(cfg: Partial<LbConfig> = {}, nServers = 3, rng: () => number = Math.random) {
    this.cfg = { ratePerSec: 8, powerVariance: 0.3, costVariance: 0.4, algo: 'rr', ...cfg };
    this.rng = rng;
    this.setServers(nServers);
  }

  setServers(n: number): void {
    this.servers = Array.from({ length: n }, (_, id) => ({
      id,
      healthy: true,
      serving: null,
      queue: [],
      total: 0,
      dropped: 0,
    }));
  }

  setConfig(partial: Partial<LbConfig>): void {
    this.cfg = { ...this.cfg, ...partial };
  }

  addServer(): void {
    if (this.servers.length >= 6) return;
    this.servers.push({ id: this.servers.length, healthy: true, serving: null, queue: [], total: 0, dropped: 0 });
  }

  toggleHealth(id: number): void {
    const s = this.servers[id];
    if (!s) return;
    s.healthy = !s.healthy;
    if (!s.healthy) {
      // drain: in-flight/queued on a downed server are dropped
      s.queue = [];
      s.serving = null;
    }
  }

  reset(): void {
    const n = this.servers.length;
    this.setServers(n);
    this.rr = 0;
    this.now = 0;
    this.spawnAcc = 0;
    this.latencies = [];
    this.doneTimes = [];
    this.served = 0;
    this.dropped = 0;
  }

  /** power (work/sec) for server `i` given current variance — deterministic, no rng. */
  power(i: number): number {
    const n = this.servers.length;
    const spread = n > 1 ? (i / (n - 1)) * 2 - 1 : 0; // -1..1
    return BASE_POWER * (1 + this.cfg.powerVariance * spread);
  }

  /** All active requests (flight + queued + serving) for rendering. */
  get requests(): Req[] {
    const out: Req[] = [];
    for (const s of this.servers) {
      if (s.serving) out.push(s.serving);
      out.push(...s.queue);
    }
    return out.concat(this.inFlight);
  }
  private inFlight: Req[] = [];

  get metrics() {
    return {
      throughput: this.doneTimes.length, // completions in the last sim-second
      dropRate: this.served + this.dropped > 0 ? this.dropped / (this.served + this.dropped) : 0,
      p50: this.percentile(50),
      p95: this.percentile(95),
      p99: this.percentile(99),
      latencies: this.latencies,
    };
  }

  percentile(p: number): number {
    if (!this.latencies.length) return 0;
    const sorted = [...this.latencies].sort((a, b) => a - b);
    const idx = Math.min(sorted.length - 1, Math.floor((p / 100) * sorted.length));
    return sorted[idx];
  }

  private serverView(): Server[] {
    return this.servers.map((s) => ({
      id: s.id,
      healthy: s.healthy,
      conns: s.queue.length + (s.serving ? 1 : 0),
      total: s.total,
    }));
  }

  /** Advance the simulation by `dtMs` real milliseconds (already time-scaled). */
  step(dtMs: number): LbEvent[] {
    const dt = dtMs / 1000;
    this.now += dtMs;
    const events: LbEvent[] = [];

    // 1. spawn new requests
    this.spawnAcc += this.cfg.ratePerSec * dt;
    while (this.spawnAcc >= 1) {
      this.spawnAcc -= 1;
      this.spawn(events);
    }

    // 2. advance in-flight travel; enqueue on arrival
    const stillFlying: Req[] = [];
    for (const r of this.inFlight) {
      r.age += dtMs;
      r.flight += dt / FLIGHT_SEC;
      if (r.flight < 1) {
        stillFlying.push(r);
        continue;
      }
      const s = this.servers[r.server];
      if (!s || !s.healthy) {
        this.dropped++;
        if (s) s.dropped++;
        events.push({ type: 'drop', server: r.server, reason: 'unhealthy' });
      } else if (s.queue.length >= QUEUE_CAP) {
        this.dropped++;
        s.dropped++;
        events.push({ type: 'drop', server: r.server, reason: 'full' });
      } else {
        r.state = 'queue';
        r.flight = 1;
        s.queue.push(r);
      }
    }
    this.inFlight = stillFlying;

    // 3. per-server: age queue, drop timeouts, serve
    for (const s of this.servers) {
      // age + timeout queued
      s.queue = s.queue.filter((r) => {
        r.age += dtMs;
        if (r.age > TIMEOUT_MS) {
          this.dropped++;
          s.dropped++;
          events.push({ type: 'drop', server: s.id, reason: 'timeout' });
          return false;
        }
        if (r.age > TIMEOUT_MS * 0.8) r.timedOut = true;
        return true;
      });

      if (!s.healthy) continue;
      // pull next into service
      if (!s.serving && s.queue.length) {
        s.serving = s.queue.shift()!;
        s.serving.state = 'serving';
      }
      // serve
      if (s.serving) {
        s.serving.age += dtMs;
        s.serving.cost -= this.power(s.id) * dt;
        if (s.serving.cost <= 0) {
          this.served++;
          s.total++;
          this.latencies.push(s.serving.age);
          if (this.latencies.length > LAT_WINDOW) this.latencies.shift();
          this.doneTimes.push(this.now);
          events.push({ type: 'done', server: s.id, latency: s.serving.age });
          s.serving = null;
        }
      }
    }

    // 4. trim throughput window to the last sim-second
    const cutoff = this.now - 1000;
    while (this.doneTimes.length && this.doneTimes[0] < cutoff) this.doneTimes.shift();

    return events;
  }

  private spawn(events: LbEvent[]): void {
    const client = CLIENTS[this.clientIdx++ % CLIENTS.length];
    const picked = chooseTarget(this.cfg.algo, client, this.serverView(), this.rr, this.rng);
    this.rr = picked.rr;
    if (picked.target === null) {
      this.dropped++;
      events.push({ type: 'drop', server: -1, reason: 'unhealthy' });
      return;
    }
    const cost = Math.max(0.15, BASE_COST * (1 + this.cfg.costVariance * (this.rng() * 2 - 1)));
    this.inFlight.push({
      id: this.nextId++,
      server: picked.target,
      state: 'flight',
      flight: 0,
      cost,
      cost0: cost,
      age: 0,
      timedOut: false,
    });
  }
}
