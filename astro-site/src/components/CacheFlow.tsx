import { useEffect, useRef, useState } from 'react';
import { CacheSim, type CacheReq } from '../sim-lib/cache-traffic';
import { useCanvasSim, usePrefersReducedMotion } from './useCanvasSim';
import { Slider, Metric, SimLegend, LiveSummary } from './simControls';
import { box, dot, label, lerp, laneOf, roundRect } from './canvasDraw';

/**
 * Sam Rose–style cache-aside flow. Requests fly to the cache: a hit bounces
 * straight back (green, fast); a miss falls through to the database (orange,
 * slow) and populates the cache on the way back. Raise key skew to watch the
 * hit ratio climb and database load fall. Logic in sim-lib/cache-traffic.
 */

const HEIGHT = 320;
const BLUE = '#4F9DF7';
const GREEN = '#2ECC71';
const ORANGE = '#E67E22';

export default function CacheFlow() {
  const simRef = useRef<CacheSim>(null as unknown as CacheSim);
  if (!simRef.current) simRef.current = new CacheSim();

  const [rate, setRate] = useState(8);
  const [capacity, setCapacity] = useState(8);
  const [skew, setSkew] = useState(0.6);
  const [ttl, setTtl] = useState(6);
  const [speed, setSpeed] = useState(1);
  const reduced = usePrefersReducedMotion();
  const [playing, setPlaying] = useState(true);
  useEffect(() => { if (reduced) setPlaying(false); }, [reduced]);
  const [, setTick] = useState(0);

  useEffect(() => {
    simRef.current.setConfig({ ratePerSec: rate, capacity, skew, ttlMs: ttl * 1000 });
  }, [rate, capacity, skew, ttl]);

  useEffect(() => {
    const t = setInterval(() => setTick((n) => n + 1), 220);
    return () => clearInterval(t);
  }, []);

  const step = (dtMs: number) => {
    simRef.current.step(dtMs);
  };

  const reqPos = (r: CacheReq, cx: number, dbx: number, clientX: number, midY: number) => {
    const lane = laneOf(r.id);
    const cacheL = cx - 46;
    const cacheR = cx + 46;
    switch (r.stage) {
      case 'in':
        return { x: lerp(clientX + 24, cacheL, r.prog), y: midY + lane * (1 - r.prog) * 0.7, c: BLUE };
      case 'hitback':
        return { x: lerp(cacheL, clientX + 24, r.prog), y: midY - 16 + lane * r.prog * 0.7, c: GREEN };
      case 'todb':
        return { x: lerp(cacheR, dbx - 28, r.prog), y: midY + 14, c: ORANGE };
      case 'dbback':
        return { x: lerp(dbx - 28, cacheR, r.prog), y: midY + 28, c: ORANGE };
      case 'out':
        return { x: lerp(cacheL, clientX + 24, r.prog), y: midY - 16 + lane * r.prog * 0.7, c: ORANGE };
    }
  };

  const draw = (ctx: CanvasRenderingContext2D, w: number, h: number) => {
    const sim = simRef.current;
    const clientX = 46;
    const cx = Math.round(w * 0.46);
    const dbx = w - 64;
    const midY = h / 2;

    // client
    box(ctx, clientX - 26, midY - 22, 52, 44, '#2a3142', BLUE);
    label(ctx, 'clients', clientX, midY + 38, '#cfd6e4', 'center');
    label(ctx, `${sim.cfg.ratePerSec}/s`, clientX, midY, '#cfd6e4', 'center');

    // cache box with capacity slots
    const cacheW = 92;
    const cacheH = 96;
    box(ctx, cx - cacheW / 2, midY - cacheH / 2, cacheW, cacheH, '#10131a', BLUE);
    label(ctx, 'Cache', cx, midY - cacheH / 2 + 14, '#cfd6e4', 'center');
    const cols = 6;
    const slot = 11;
    const gx = cx - (cols * slot) / 2 + 2;
    const gy = midY - cacheH / 2 + 26;
    for (let i = 0; i < sim.cfg.capacity; i++) {
      const r = Math.floor(i / cols);
      const c = i % cols;
      ctx.fillStyle = i < sim.occupancy ? GREEN : '#222834';
      roundRect(ctx, gx + c * slot, gy + r * slot, slot - 3, slot - 3, 2);
      ctx.fill();
    }
    label(ctx, `${sim.occupancy}/${sim.cfg.capacity}`, cx, midY + cacheH / 2 - 12, '#9aa3b2', 'center');

    // database (slow origin)
    const dbw = 56;
    const dbh = 64;
    box(ctx, dbx - dbw / 2, midY - dbh / 2, dbw, dbh, '#241b16', ORANGE);
    label(ctx, 'DB', dbx, midY - 4, '#e6b483', 'center', '700 12px');
    label(ctx, '~10ms', dbx, midY + 12, '#9aa3b2', 'center', '600 9px');

    // requests
    for (const r of sim.reqs) {
      const p = reqPos(r, cx, dbx, clientX, midY);
      if (!p) continue;
      ctx.fillStyle = p.c;
      dot(ctx, p.x, p.y, 5);
    }
  };

  const canvasRef = useCanvasSim({
    height: HEIGHT,
    playing,
    timeScale: speed,
    ariaLabel: 'Animated cache-aside simulation: reads served from cache or falling through to the database.',
    step,
    draw,
  });
  const m = simRef.current.metrics;

  return (
    <div className="lsd-diagram not-content">
      <p className="lsd-hint">
        A <strong style={{ color: GREEN }}>hit</strong> returns from cache in ~1ms; a{' '}
        <strong style={{ color: ORANGE }}>miss</strong> falls through to the database (~10ms) and fills the cache.
        <strong> Raise key skew</strong> (hot keys) to watch the hit ratio climb and database load drop.
      </p>

      <div className="lsd-controls">
        <button onClick={() => setPlaying((p) => !p)}>{playing ? '⏸ Pause' : '▶ Play'}</button>
        <button onClick={() => { simRef.current.reset(); setTick((n) => n + 1); }}>↺ Reset</button>
        <span className="lsd-seg" style={{ marginLeft: 'auto' }}>
          {[0.5, 1, 2].map((s) => (
            <button key={s} aria-pressed={speed === s} onClick={() => setSpeed(s)}>{s}×</button>
          ))}
        </span>
      </div>

      <canvas ref={canvasRef} className="lsd-canvas" style={{ height: HEIGHT }} />
      <LiveSummary
        text={`Hit ratio ${Math.round(m.hitRatio * 100)} percent, database load ${m.dbLoad} per second, cache ${m.occupancy} of ${capacity} slots used.`}
      />
      <SimLegend
        items={[
          { glyph: '●', color: '#4F9DF7', label: 'request heading to cache' },
          { glyph: '●', color: '#2ECC71', label: 'cache hit — fast return' },
          { glyph: '●', color: '#E67E22', label: 'miss — falls through to the database' },
        ]}
      />

      <div className="lsd-slider-grid">
        <Slider label="request rate" value={rate} min={1} max={40} step={1} suffix="/s" onChange={setRate} />
        <Slider label="cache capacity" value={capacity} min={2} max={24} step={1} onChange={setCapacity} />
        <Slider label="key skew (hot keys)" value={skew} min={0} max={1} step={0.05} pct onChange={setSkew} />
        <Slider label="TTL" value={ttl} min={1} max={20} step={1} suffix="s" onChange={setTtl} />
      </div>

      <div className="lsd-metric-row">
        <Metric label="hit ratio" value={`${Math.round(m.hitRatio * 100)}%`} warn={m.hitRatio < 0.5} />
        <Metric label="db load" value={`${m.dbLoad}/s`} warn={m.dbLoad > 8} />
        <Metric label="cache" value={`${m.occupancy}/${capacity}`} />
        <Metric label="p50 (rel)" value={`${(m.p50 / 400).toFixed(1)}×`} />
        <Metric label="p95 (rel)" value={`${(m.p95 / 400).toFixed(1)}×`} />
      </div>
    </div>
  );
}
