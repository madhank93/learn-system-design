import { useEffect, useRef, useState } from 'react';
import { CbSim, type CbReq } from '../sim-lib/cb-traffic';
import type { CBState } from '../sim-lib/circuit-breaker';
import { useCanvasSim, usePrefersReducedMotion } from './useCanvasSim';
import { Slider, Metric, SimLegend, LiveSummary } from './simControls';
import { box, dot, label, lerp, laneOf } from './canvasDraw';

/**
 * Sam Rose–style circuit-breaker flow. Requests pass through the breaker to a
 * flaky downstream (green ok / red fail). When the failure rate crosses the
 * threshold the breaker trips OPEN and fails fast (orange, no downstream call);
 * after a cooldown it probes HALF_OPEN. Decisions reuse the CircuitBreaker
 * engine via sim-lib/cb-traffic.
 */

const HEIGHT = 320;
const STATE_COLOR: Record<CBState, string> = { CLOSED: '#2ECC71', OPEN: '#E74C3C', HALF_OPEN: '#E67E22' };
const GREEN = '#2ECC71';
const RED = '#E74C3C';
const RED_TEXT = '#ff7b72'; // RED reads below 4.5:1 as small text on the panel
const ORANGE = '#E67E22';
const BLUE = '#4F9DF7';

export default function CircuitBreakerFlow() {
  const simRef = useRef<CbSim>(null as unknown as CbSim);
  if (!simRef.current) simRef.current = new CbSim();

  const [rate, setRate] = useState(8);
  const [failRate, setFailRate] = useState(0);
  const [speed, setSpeed] = useState(1);
  const reduced = usePrefersReducedMotion();
  const [playing, setPlaying] = useState(true);
  useEffect(() => { if (reduced) setPlaying(false); }, [reduced]);
  const [, setTick] = useState(0);

  useEffect(() => {
    simRef.current.setConfig({ ratePerSec: rate, failRate });
  }, [rate, failRate]);

  useEffect(() => {
    const t = setInterval(() => setTick((n) => n + 1), 200);
    return () => clearInterval(t);
  }, []);

  const step = (dtMs: number) => {
    simRef.current.step(dtMs);
  };

  const reqPos = (r: CbReq, clientX: number, bx: number, dsx: number, midY: number) => {
    const lane = laneOf(r.id);
    if (r.stage === 'in') return { x: lerp(clientX + 24, bx - 36, r.prog), y: midY + lane * (1 - r.prog) * 0.7, c: BLUE };
    if (r.stage === 'block') return { x: bx - 12, y: midY - Math.sin(r.prog * Math.PI) * 52, c: ORANGE };
    // ok / fail: out to downstream and back
    const tri = r.prog < 0.5 ? r.prog * 2 : (1 - r.prog) * 2;
    return { x: lerp(bx + 36, dsx - 28, tri), y: midY + 12, c: r.stage === 'ok' ? GREEN : RED };
  };

  const draw = (ctx: CanvasRenderingContext2D, w: number, h: number) => {
    const sim = simRef.current;
    const clientX = 46;
    const bx = Math.round(w * 0.44);
    const dsx = w - 66;
    const midY = h / 2;
    const st = sim.state;
    const sc = STATE_COLOR[st];

    // client
    box(ctx, clientX - 26, midY - 22, 52, 44, '#2a3142', BLUE);
    label(ctx, 'clients', clientX, midY + 38, '#cfd6e4', 'center');
    label(ctx, `${sim.cfg.ratePerSec}/s`, clientX, midY, '#cfd6e4', 'center');

    // breaker
    const bw = 84;
    const bh = 92;
    ctx.save();
    ctx.shadowColor = sc;
    ctx.shadowBlur = 16;
    box(ctx, bx - bw / 2, midY - bh / 2, bw, bh, '#10131a', sc);
    ctx.restore();
    label(ctx, 'breaker', bx, midY - bh / 2 + 14, '#9aa3b2', 'center');
    label(ctx, st, bx, midY - 2, sc, 'center', '700 13px');
    if (st === 'OPEN') label(ctx, `probe ${(sim.metrics.openLeft / 1000).toFixed(1)}s`, bx, midY + 16, '#9aa3b2', 'center', '600 9px');
    else label(ctx, `${Math.round(sim.metrics.failureRate * 100)}% fail`, bx, midY + 16, '#9aa3b2', 'center', '600 9px');

    // downstream
    const dw = 58;
    const dh = 70;
    box(ctx, dsx - dw / 2, midY - dh / 2, dw, dh, '#1a1f29', '#3a4150');
    label(ctx, 'downstream', dsx, midY + dh / 2 + 12, '#9aa3b2', 'center', '600 9px');
    label(ctx, `${sim.metrics.dsLoad}/s`, dsx, midY, '#cfd6e4', 'center', '700 12px');

    // requests
    for (const r of sim.reqs) {
      const p = reqPos(r, clientX, bx, dsx, midY);
      ctx.fillStyle = p.c;
      dot(ctx, p.x, p.y, 5);
    }
  };

  const canvasRef = useCanvasSim({
    height: HEIGHT,
    playing,
    timeScale: speed,
    ariaLabel: 'Animated circuit-breaker simulation: calls passing, failing, or fast-failing as the breaker trips.',
    step,
    draw,
  });
  const m = simRef.current.metrics;

  return (
    <div className="lsd-diagram not-content">
      <p className="lsd-hint">
        Each request calls a flaky downstream (<strong style={{ color: GREEN }}>ok</strong> /{' '}
        <strong style={{ color: RED_TEXT }}>fail</strong>). <strong>Push the failure rate up</strong>: once failures
        cross 50% the breaker trips <strong style={{ color: RED_TEXT }}>OPEN</strong> and fails fast (orange) — sending
        zero load downstream — then probes <strong style={{ color: ORANGE }}>HALF_OPEN</strong> after a cooldown.
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
        text={`Breaker ${m.state}, window failure ${Math.round(m.failureRate * 100)} percent, downstream load ${m.dsLoad} per second, ${m.blocked} calls fast-failed.`}
      />
      <SimLegend
        items={[
          { glyph: '●', color: '#2ECC71', label: 'call succeeded' },
          { glyph: '●', color: '#E74C3C', label: 'call failed downstream' },
          { glyph: '●', color: '#E67E22', label: 'fast-failed (breaker OPEN, no downstream call)' },
        ]}
      />

      <div className="lsd-slider-grid">
        <Slider label="request rate" value={rate} min={1} max={30} step={1} suffix="/s" onChange={setRate} />
        <Slider label="downstream failure rate" value={failRate} min={0} max={1} step={0.05} pct onChange={setFailRate} />
      </div>

      <div className="lsd-metric-row">
        <Metric label="state" value={m.state} warn={m.state !== 'CLOSED'} />
        <Metric label="window fail" value={`${Math.round(m.failureRate * 100)}%`} warn={m.failureRate > 0.5} />
        <Metric label="downstream load" value={`${m.dsLoad}/s`} />
        <Metric label="fast-failed" value={`${m.blocked}`} warn={m.blocked > 0} />
      </div>
    </div>
  );
}
