/**
 * Continuous single-leader replication simulation (pure, no DOM).
 * Writes go to the primary, which bumps a version and streams a WAL record to
 * each replica; the record applies after the replication lag. Reads hit a
 * replica and are "stale" if that replica hasn't yet applied everything that
 * committed before the read was issued — the read-your-writes hazard.
 */

export type ReplStage = 'in' | 'flash';

export type ReplReq = {
  id: number;
  kind: 'write' | 'read';
  replica: number; // target replica for reads (-1 for writes)
  expected: number; // primary version at read issue time
  stage: ReplStage;
  prog: number;
  flash: number;
  fresh?: boolean;
};

export type Wal = { id: number; replica: number; version: number; prog: number };

export type ReplConfig = { writeRate: number; readRate: number; lagMs: number; replicas: number };

const LEG_SEC = 0.4;

export class ReplSim {
  cfg: ReplConfig;
  version = 0; // primary's latest committed version
  applied: number[] = []; // per-replica applied version
  reqs: ReplReq[] = [];
  wal: Wal[] = [];
  freshReads = 0;
  staleReads = 0;
  private rng: () => number;
  private nextId = 1;
  private nextWal = 1;
  private wAcc = 0;
  private rAcc = 0;
  private now = 0;

  constructor(cfg: Partial<ReplConfig> = {}, rng: () => number = Math.random) {
    this.cfg = { writeRate: 3, readRate: 8, lagMs: 1200, replicas: 2, ...cfg };
    this.rng = rng;
    this.setReplicas(this.cfg.replicas);
  }

  setReplicas(n: number): void {
    this.applied = Array.from({ length: n }, (_, i) => this.applied[i] ?? this.version);
  }

  setConfig(p: Partial<ReplConfig>): void {
    const prevN = this.cfg.replicas;
    this.cfg = { ...this.cfg, ...p };
    if (this.cfg.replicas !== prevN) this.setReplicas(this.cfg.replicas);
  }

  reset(): void {
    this.version = 0;
    this.applied = this.applied.map(() => 0);
    this.reqs = [];
    this.wal = [];
    this.freshReads = 0;
    this.staleReads = 0;
    this.wAcc = 0;
    this.rAcc = 0;
    this.now = 0;
  }

  get metrics() {
    const reads = this.freshReads + this.staleReads;
    const minApplied = this.applied.length ? Math.min(...this.applied) : this.version;
    return {
      staleRate: reads ? this.staleReads / reads : 0,
      maxLagVersions: this.version - minApplied,
      version: this.version,
      reads,
    };
  }

  step(dtMs: number): void {
    const dt = dtMs / 1000;
    this.now += dtMs;
    const walSec = Math.max(50, this.cfg.lagMs) / 1000;

    // spawn writes + reads
    this.wAcc += this.cfg.writeRate * dt;
    while (this.wAcc >= 1) {
      this.wAcc -= 1;
      this.reqs.push({ id: this.nextId++, kind: 'write', replica: -1, expected: 0, stage: 'in', prog: 0, flash: 0 });
    }
    this.rAcc += this.cfg.readRate * dt;
    while (this.rAcc >= 1) {
      this.rAcc -= 1;
      const replica = Math.floor(this.rng() * this.applied.length);
      this.reqs.push({ id: this.nextId++, kind: 'read', replica, expected: this.version, stage: 'in', prog: 0, flash: 0 });
    }

    // advance requests
    const alive: ReplReq[] = [];
    for (const r of this.reqs) {
      if (r.stage === 'flash') {
        r.flash += dt / 0.4;
        if (r.flash < 1) alive.push(r);
        continue;
      }
      r.prog += dt / LEG_SEC;
      if (r.prog < 1) {
        alive.push(r);
        continue;
      }
      if (r.kind === 'write') {
        this.version++;
        for (let i = 0; i < this.applied.length; i++) {
          this.wal.push({ id: this.nextWal++, replica: i, version: this.version, prog: 0 });
        }
        // write committed at primary; request leaves
      } else {
        r.fresh = this.applied[r.replica] >= r.expected;
        if (r.fresh) this.freshReads++;
        else this.staleReads++;
        r.stage = 'flash';
        r.prog = 1;
        alive.push(r);
      }
    }
    this.reqs = alive;

    // advance WAL records; apply on arrival
    const walAlive: Wal[] = [];
    for (const p of this.wal) {
      p.prog += dt / walSec;
      if (p.prog >= 1) {
        if (this.applied[p.replica] < p.version) this.applied[p.replica] = p.version;
      } else {
        walAlive.push(p);
      }
    }
    this.wal = walAlive;
  }
}
