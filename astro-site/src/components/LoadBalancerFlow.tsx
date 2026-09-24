import { useEffect, useMemo, useRef, useState } from 'react';
import { LbSim, type LbEvent } from '../sim-lib/lb-traffic';
import type { Algo } from '../sim-lib/load-balancer';
import { useCanvasSim, usePrefersReducedMotion } from './useCanvasSim';
import { Slider, Metric, SimLegend, LiveSummary } from './simControls';

/**
 * Sam Rose–style load-balancing flow. Requests are balls that fly to servers,
 * wait in visible queues (colour drifts cool→warm as they wait), get served
 * (clock-shrink), and drop when a queue fills or times out. Drag the sliders
 * and watch drop % / latency percentiles react in real time. Pure timing/state
 * lives in sim-lib/lb-traffic; this file only maps state → canvas pixels.
 */

const HEIGHT = 360;
const TIMEOUT_MS = 5000;
const COOL = [79, 157, 247]; // #4F9DF7
const WARM = [231, 76, 60]; // #E74C3C
const ALGOS: { id: Algo; label: string }[] = [
  { id: 'rr', label: 'Round robin' },
  { id: 'least', label: 'Least conn' },
  { id: 'random', label: 'Random' },
  { id: 'sticky', label: 'Sticky' },
];

type Effect = { x: number; y: number; vy: number; life: number; color: string; kind: 'drop' | 'done' };

function ageColor(age: number, alpha = 1): string {
  const t = Math.max(0, Math.min(1, age / TIMEOUT_MS));
  const c = COOL.map((lo, i) => Math.round(lo + (WARM[i] - lo) * t));
  return `rgba(${c[0]},${c[1]},${c[2]},${alpha})`;
}

export default function LoadBalancerFlow() {
  const simRef = useRef<LbSim>(null as unknown as LbSim);
  if (!simRef.current) simRef.current = new LbSim();
  const effects = useRef<Effect[]>([]);

  const [algo, setAlgo] = useState<Algo>('rr');
  const [rate, setRate] = useState(8);
  const [powerVar, setPowerVar] = useState(0.3);
  const [costVar, setCostVar] = useState(0.4);
  const [speed, setSpeed] = useState(1);
  const reduced = usePrefersReducedMotion();
  const [playing, setPlaying] = useState(true);
  const [, setTick] = useState(0);

  // honour reduced-motion: don't autoplay (still fully usable via Play)
  useEffect(() => {
    if (reduced) setPlaying(false);
  }, [reduced]);

  // push control changes into the running sim
  useEffect(() => {
    simRef.current.setConfig({ algo, ratePerSec: rate, powerVariance: powerVar, costVariance: costVar });
  }, [algo, rate, powerVar, costVar]);

  // refresh the DOM panels (metrics, server buttons) a few times a second
  useEffect(() => {
    const t = setInterval(() => setTick((n) => n + 1), 220);
    return () => clearInterval(t);
  }, []);

  const geom = (w: number, n: number) => {
    const boxW = 132;
    const boxH = Math.min(64, (HEIGHT - 28) / n - 10);
    const x0 = w - boxW - 22;
    const totalH = n * boxH + (n - 1) * 10;
    const top = (HEIGHT - totalH) / 2;
    return { boxW, boxH, x0, top, srcX: 46, srcY: HEIGHT / 2 };
  };

  const step = (dtMs: number) => {
    const sim = simRef.current;
    const evs = sim.step(dtMs);
    spawnEffects(evs, dtMs);
    // advance effects
    for (const e of effects.current) {
      e.y += e.vy * (dtMs / 1000);
      e.vy += 320 * (dtMs / 1000);
      e.life -= dtMs;
    }
    effects.current = effects.current.filter((e) => e.life > 0);
  };

  const spawnEffects = (evs: LbEvent[], _dt: number) => {
    const sim = simRef.current;
    // we don't know canvas width here; place drops near the server box using a
    // cached width set in draw. Fallback to a nominal width.
    const w = lastW.current;
    const n = sim.servers.length;
    const g = geom(w, n);
    for (const e of evs) {
      const i = e.server < 0 ? 0 : e.server;
      const y = g.top + i * (g.boxH + 10) + g.boxH / 2;
      const x = e.server < 0 ? g.srcX : g.x0 + 8;
      if (e.type === 'drop') {
        effects.current.push({ x, y, vy: -40, life: 700, color: '#E74C3C', kind: 'drop' });
      }
    }
  };

  const lastW = useRef(640);

  const draw = (ctx: CanvasRenderingContext2D, w: number, h: number) => {
    lastW.current = w;
    const sim = simRef.current;
    const n = sim.servers.length;
    const g = geom(w, n);

    // source emitter
    ctx.fillStyle = '#2a3142';
    ctx.strokeStyle = '#4F9DF7';
    ctx.lineWidth = 1.5;
    roundRect(ctx, g.srcX - 26, g.srcY - 26, 52, 52, 10);
    ctx.fill();
    ctx.stroke();
    ctx.fillStyle = '#cfd6e4';
    ctx.font = '600 11px ui-sans-serif, system-ui, sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText('clients', g.srcX, g.srcY + 44);
    ctx.fillText(`${sim.cfg.ratePerSec}/s`, g.srcX, g.srcY + 4);

    // server boxes
    sim.servers.forEach((s, i) => {
      const y = g.top + i * (g.boxH + 10);
      const cy = y + g.boxH / 2;
      ctx.lineWidth = 1.5;
      ctx.strokeStyle = s.healthy ? '#3a4150' : '#E74C3C';
      ctx.fillStyle = s.healthy ? '#10131a' : 'rgba(231,76,60,0.08)';
      roundRect(ctx, g.x0, y, g.boxW, g.boxH, 8);
      ctx.fill();
      ctx.stroke();

      // power bar (relative capacity)
      const pw = sim.power(i);
      const pwMax = 3 * (1 + sim.cfg.powerVariance);
      ctx.fillStyle = '#2ECC71';
      ctx.globalAlpha = 0.5;
      ctx.fillRect(g.x0, y + g.boxH - 3, (g.boxW * pw) / (pwMax || 1), 3);
      ctx.globalAlpha = 1;

      // label
      ctx.fillStyle = s.healthy ? '#cfd6e4' : '#E74C3C';
      ctx.font = '600 11px ui-sans-serif, system-ui, sans-serif';
      ctx.textAlign = 'left';
      ctx.fillText(`S${s.id}${s.healthy ? '' : ' ✕'}`, g.x0 + 8, y + 14);
      ctx.fillStyle = '#7d8597';
      ctx.textAlign = 'right';
      ctx.fillText(`${s.total}`, g.x0 + g.boxW - 8, y + 14);

      // serving (clock-shrink)
      if (s.serving) {
        const r = 6 * Math.max(0.25, s.serving.cost / s.serving.cost0);
        ctx.fillStyle = ageColor(s.serving.age);
        dot(ctx, g.x0 + 16, cy + 6, r);
      }
      // queue dots
      s.queue.forEach((q, k) => {
        const x = g.x0 + 32 + k * 12;
        if (x > g.x0 + g.boxW - 8) return;
        drawReq(ctx, x, cy + 6, q.age, q.timedOut);
      });
    });

    // in-flight balls
    for (const r of sim.requests) {
      if (r.state !== 'flight') continue;
      const i = r.server;
      const y = g.top + i * (g.boxH + 10) + g.boxH / 2;
      const x = g.srcX + 26 + (g.x0 - (g.srcX + 26)) * r.flight;
      const yy = g.srcY + (y - g.srcY) * r.flight;
      drawReq(ctx, x, yy, r.age, false);
    }

    // effects (drops fall + fade)
    for (const e of effects.current) {
      const a = Math.max(0, e.life / 700);
      ctx.fillStyle = `rgba(231,76,60,${a})`;
      dot(ctx, e.x, e.y, 5);
      ctx.strokeStyle = `rgba(231,76,60,${a})`;
      ctx.lineWidth = 1.5;
      ctx.beginPath();
      ctx.moveTo(e.x - 6, e.y - 6);
      ctx.lineTo(e.x + 6, e.y + 6);
      ctx.moveTo(e.x + 6, e.y - 6);
      ctx.lineTo(e.x - 6, e.y + 6);
      ctx.stroke();
    }
  };

  const canvasRef = useCanvasSim({
    height: HEIGHT,
    playing,
    timeScale: speed,
    ariaLabel: 'Animated load-balancing simulation: requests flowing from clients to a pool of servers.',
    step,
    draw,
  });

  const m = simRef.current.metrics;
  const servers = simRef.current.servers;
  const hist = useMemo(() => histogram(m.latencies), [m.latencies.length, m.p99]);

  return (
    <div className="lsd-diagram not-content">
      <p className="lsd-hint">
        Requests fly in from the left and queue at each server; a ball reddens the longer it waits.
        <strong> Push the rate past total capacity</strong> to watch queues back up and drops (✕) climb — then
        switch to <strong>Least conn</strong> or even out <strong>server power</strong> and watch p99 recover.
      </p>

      <div className="lsd-controls">
        <span className="lsd-seg">
          {ALGOS.map((a) => (
            <button key={a.id} aria-pressed={algo === a.id} onClick={() => setAlgo(a.id)}>
              {a.label}
            </button>
          ))}
        </span>
        <button onClick={() => setPlaying((p) => !p)}>{playing ? '⏸ Pause' : '▶ Play'}</button>
        <button onClick={() => { simRef.current.addServer(); setTick((n) => n + 1); }}>＋ Server</button>
        <button onClick={() => { simRef.current.reset(); effects.current = []; setTick((n) => n + 1); }}>↺ Reset</button>
        <span className="lsd-seg" style={{ marginLeft: 'auto' }}>
          {[0.5, 1, 2].map((s) => (
            <button key={s} aria-pressed={speed === s} onClick={() => setSpeed(s)}>{s}×</button>
          ))}
        </span>
      </div>

      <canvas ref={canvasRef} className="lsd-canvas" style={{ height: HEIGHT }} />
      <LiveSummary
        text={`Throughput ${m.throughput} per second, drop rate ${Math.round(m.dropRate * 100)} percent, p99 latency ${Math.round(m.p99)} milliseconds, ${servers.length} servers.`}
      />
      <SimLegend
        items={[
          { glyph: '●', color: '#4F9DF7', label: 'in-flight request' },
          { glyph: '●', color: '#E67E22', label: 'waiting in queue (reddens as it waits)' },
          { glyph: '✕', color: '#E74C3C', label: 'dropped — queue full or timed out' },
        ]}
      />

      <div className="lsd-controls" style={{ marginTop: '0.6rem' }}>
        {servers.map((s) => (
          <button
            key={s.id}
            aria-pressed={!s.healthy}
            onClick={() => { simRef.current.toggleHealth(s.id); setTick((n) => n + 1); }}
            title={s.healthy ? 'Take server down' : 'Revive server'}
          >
            {s.healthy ? `⏻ S${s.id}` : `↺ S${s.id}`}
          </button>
        ))}
      </div>

      <div className="lsd-slider-grid">
        <Slider label="request rate" value={rate} min={1} max={40} step={1} suffix="/s" onChange={setRate} />
        <Slider label="server power variance" value={powerVar} min={0} max={0.8} step={0.05} pct onChange={setPowerVar} />
        <Slider label="request cost variance" value={costVar} min={0} max={0.9} step={0.05} pct onChange={setCostVar} />
      </div>

      <div className="lsd-metric-row">
        <Metric label="throughput" value={`${m.throughput}/s`} />
        <Metric label="drop rate" value={`${Math.round(m.dropRate * 100)}%`} warn={m.dropRate > 0.1} />
        <Metric label="p50" value={`${Math.round(m.p50)}ms`} />
        <Metric label="p95" value={`${Math.round(m.p95)}ms`} />
        <Metric label="p99" value={`${Math.round(m.p99)}ms`} warn={m.p99 > 2000} />
      </div>

      <div className="lsd-hist" aria-hidden>
        {hist.bars.map((v, i) => (
          <span key={i} className="lsd-hist-bar" style={{ height: `${Math.round(v * 100)}%` }} />
        ))}
        <span className="lsd-hist-axis">latency distribution → {hist.max}ms</span>
      </div>
    </div>
  );
}

function histogram(lat: number[]): { bars: number[]; max: number } {
  const buckets = 24;
  if (!lat.length) return { bars: new Array(buckets).fill(0), max: 0 };
  const hi = Math.max(200, Math.max(...lat));
  const bars = new Array(buckets).fill(0);
  for (const v of lat) {
    const b = Math.min(buckets - 1, Math.floor((v / hi) * buckets));
    bars[b]++;
  }
  const peak = Math.max(...bars) || 1;
  return { bars: bars.map((b) => b / peak), max: Math.round(hi) };
}

// --- canvas helpers ---
function dot(ctx: CanvasRenderingContext2D, x: number, y: number, r: number) {
  ctx.beginPath();
  ctx.arc(x, y, r, 0, Math.PI * 2);
  ctx.fill();
}
function drawReq(ctx: CanvasRenderingContext2D, x: number, y: number, age: number, hollow: boolean) {
  if (hollow) {
    ctx.strokeStyle = ageColor(age);
    ctx.lineWidth = 1.6;
    ctx.beginPath();
    ctx.arc(x, y, 5, 0, Math.PI * 2);
    ctx.stroke();
  } else {
    ctx.fillStyle = ageColor(age);
    dot(ctx, x, y, 5);
  }
}
function roundRect(ctx: CanvasRenderingContext2D, x: number, y: number, w: number, h: number, r: number) {
  ctx.beginPath();
  ctx.moveTo(x + r, y);
  ctx.arcTo(x + w, y, x + w, y + h, r);
  ctx.arcTo(x + w, y + h, x, y + h, r);
  ctx.arcTo(x, y + h, x, y, r);
  ctx.arcTo(x, y, x + w, y, r);
  ctx.closePath();
}
