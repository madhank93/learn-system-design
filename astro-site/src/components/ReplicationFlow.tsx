import { useEffect, useRef, useState } from 'react';
import { ReplSim } from '../sim-lib/repl-traffic';
import { useCanvasSim, usePrefersReducedMotion } from './useCanvasSim';
import { Slider, Metric, SimLegend, LiveSummary } from './simControls';
import { box, dot, label, lerp, laneOf, roundRect } from './canvasDraw';

/**
 * Sam Rose–style single-leader replication flow. Writes (orange) go to the
 * primary, which streams a WAL record to each replica; the record arrives after
 * the replication lag. Reads (blue) hit a replica and flash green (fresh) or
 * amber (stale) — stale when that replica hasn't caught up to writes committed
 * before the read. Logic in sim-lib/repl-traffic.
 */

const HEIGHT = 340;
const BLUE = '#4F9DF7';
const ORANGE = '#E67E22';
const GREEN = '#2ECC71';
const AMBER = '#F1C40F';

export default function ReplicationFlow() {
  const simRef = useRef<ReplSim>(null as unknown as ReplSim);
  if (!simRef.current) simRef.current = new ReplSim();

  const [writeRate, setWriteRate] = useState(3);
  const [readRate, setReadRate] = useState(8);
  const [lag, setLag] = useState(1200);
  const [replicas, setReplicas] = useState(2);
  const [speed, setSpeed] = useState(1);
  const reduced = usePrefersReducedMotion();
  const [playing, setPlaying] = useState(true);
  useEffect(() => { if (reduced) setPlaying(false); }, [reduced]);
  const [, setTick] = useState(0);

  useEffect(() => {
    simRef.current.setConfig({ writeRate, readRate, lagMs: lag, replicas });
  }, [writeRate, readRate, lag, replicas]);

  useEffect(() => {
    const t = setInterval(() => setTick((n) => n + 1), 200);
    return () => clearInterval(t);
  }, []);

  const step = (dtMs: number) => {
    simRef.current.step(dtMs);
  };

  const draw = (ctx: CanvasRenderingContext2D, w: number, h: number) => {
    const sim = simRef.current;
    const clientX = 46;
    const pX = Math.round(w * 0.4);
    const rX = w - 92;
    const midY = h / 2;
    const n = sim.applied.length;
    const top = 44;
    const bot = h - 44;
    const ry = (i: number) => (n === 1 ? midY : top + (i / (n - 1)) * (bot - top));

    // client
    box(ctx, clientX - 26, midY - 22, 52, 44, '#2a3142', BLUE);
    label(ctx, 'clients', clientX, midY + 38, '#cfd6e4', 'center');

    // wal links
    sim.applied.forEach((_, i) => {
      ctx.strokeStyle = 'rgba(230,126,34,0.18)';
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.moveTo(pX + 30, midY);
      ctx.lineTo(rX - 30, ry(i));
      ctx.stroke();
    });

    // primary
    const pw = 66;
    const ph = 58;
    box(ctx, pX - pw / 2, midY - ph / 2, pw, ph, '#241b16', ORANGE);
    label(ctx, 'Primary', pX, midY - 6, '#e6b483', 'center', '700 11px');
    label(ctx, `v${sim.version}`, pX, midY + 12, '#cfd6e4', 'center', '700 12px');

    // replicas
    sim.applied.forEach((av, i) => {
      const y = ry(i);
      const behind = sim.version - av;
      const stroke = behind > 0 ? AMBER : GREEN;
      box(ctx, rX - 34, y - 20, 68, 40, '#10131a', stroke);
      label(ctx, `R${i}`, rX - 24, y - 6, '#cfd6e4', 'left', '700 11px');
      label(ctx, `v${av}`, rX - 24, y + 9, behind > 0 ? AMBER : '#9aa3b2', 'left', '600 10px');
      if (behind > 0) label(ctx, `-${behind}`, rX + 24, y, AMBER, 'right', '600 10px');
    });

    // WAL packets (squares streaming primary → replica)
    for (const p of sim.wal) {
      const y = ry(p.replica);
      const x = lerp(pX + 30, rX - 30, p.prog);
      const yy = lerp(midY, y, p.prog);
      ctx.fillStyle = ORANGE;
      roundRect(ctx, x - 4, yy - 4, 8, 8, 2);
      ctx.fill();
    }

    // requests
    for (const r of sim.reqs) {
      if (r.kind === 'write') {
        const x = lerp(clientX + 24, pX - pw / 2, r.prog);
        const y = midY + laneOf(r.id) * (1 - r.prog) * 0.6;
        ctx.fillStyle = ORANGE;
        dot(ctx, x, y, 5);
      } else {
        const y0 = ry(r.replica);
        if (r.stage === 'flash') {
          const a = 1 - r.flash;
          const col = r.fresh ? GREEN : AMBER;
          ctx.globalAlpha = a;
          ctx.strokeStyle = col;
          ctx.lineWidth = 2;
          ctx.beginPath();
          ctx.arc(rX - 34 + 34, y0, 16 + r.flash * 10, 0, Math.PI * 2);
          ctx.stroke();
          ctx.globalAlpha = 1;
          if (!r.fresh) label(ctx, 'stale', rX, y0 + 30, AMBER, 'center', '700 10px');
        } else {
          const x = lerp(clientX + 24, rX - 34, r.prog);
          const y = lerp(midY, y0, r.prog);
          ctx.fillStyle = BLUE;
          dot(ctx, x, y, 5);
        }
      }
    }
  };

  const canvasRef = useCanvasSim({
    height: HEIGHT,
    playing,
    timeScale: speed,
    ariaLabel: 'Animated replication simulation: writes streaming from the primary to replicas, reads returning fresh or stale.',
    step,
    draw,
  });
  const m = simRef.current.metrics;

  return (
    <div className="lsd-diagram not-content">
      <p className="lsd-hint">
        <strong style={{ color: ORANGE }}>Writes</strong> hit the primary, which streams a WAL record to each
        replica after the <strong>replication lag</strong>. <strong style={{ color: BLUE }}>Reads</strong> hit a
        replica and flash <strong style={{ color: GREEN }}>fresh</strong> or{' '}
        <strong style={{ color: AMBER }}>stale</strong>. Raise the lag (or the write rate) to watch stale reads climb.
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
        text={`Stale reads ${Math.round(m.staleRate * 100)} percent, replicas up to ${m.maxLagVersions} writes behind, primary at version ${m.version}.`}
      />
      <SimLegend
        items={[
          { glyph: '●', color: '#E67E22', label: 'write to primary' },
          { glyph: '■', color: '#E67E22', label: 'WAL record streaming to a replica' },
          { glyph: '●', color: '#4F9DF7', label: 'read hitting a replica' },
          { glyph: '○', color: '#2ECC71', label: 'fresh read' },
          { glyph: '○', color: '#F1C40F', label: 'stale read (replica behind)' },
        ]}
      />

      <div className="lsd-slider-grid">
        <Slider label="write rate" value={writeRate} min={0} max={20} step={1} suffix="/s" onChange={setWriteRate} />
        <Slider label="read rate" value={readRate} min={1} max={30} step={1} suffix="/s" onChange={setReadRate} />
        <Slider label="replication lag" value={lag} min={100} max={4000} step={100} suffix="ms" onChange={setLag} />
        <Slider label="replicas" value={replicas} min={1} max={4} step={1} onChange={setReplicas} />
      </div>

      <div className="lsd-metric-row">
        <Metric label="stale reads" value={`${Math.round(m.staleRate * 100)}%`} warn={m.staleRate > 0.1} />
        <Metric label="replication lag" value={`${lag}ms`} />
        <Metric label="max lag" value={`${m.maxLagVersions} writes`} warn={m.maxLagVersions > 3} />
        <Metric label="primary version" value={`v${m.version}`} />
      </div>
    </div>
  );
}
