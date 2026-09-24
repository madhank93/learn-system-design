import { useEffect, useRef, useState } from 'react';
import { RlSim } from '../sim-lib/rl-traffic';
import { useCanvasSim, usePrefersReducedMotion } from './useCanvasSim';
import { Slider, Metric, SimLegend, LiveSummary } from './simControls';

/**
 * Sam Rose–style rate-limiting flow. Requests stream toward a token-bucket gate;
 * each arrival spends a token (passes through green) or is rejected with a red
 * 429 bounce. The bucket tank shows the live token level. Burst past the refill
 * rate and watch the bucket drain and denials begin. Decision logic reuses the
 * TokenBucket engine via sim-lib/rl-traffic.
 */

const HEIGHT = 320;

export default function RateLimiterFlow() {
  const simRef = useRef<RlSim>(null as unknown as RlSim);
  if (!simRef.current) simRef.current = new RlSim();

  const [rate, setRate] = useState(6);
  const [capacity, setCapacity] = useState(10);
  const [refill, setRefill] = useState(6);
  const [speed, setSpeed] = useState(1);
  const reduced = usePrefersReducedMotion();
  const [playing, setPlaying] = useState(true);
  useEffect(() => { if (reduced) setPlaying(false); }, [reduced]);
  const [, setTick] = useState(0);

  useEffect(() => {
    simRef.current.setConfig({ ratePerSec: rate, capacity, refillPerSec: refill });
  }, [rate, capacity, refill]);

  useEffect(() => {
    const t = setInterval(() => setTick((n) => n + 1), 200);
    return () => clearInterval(t);
  }, []);

  const step = (dtMs: number) => {
    simRef.current.step(dtMs);
  };

  const draw = (ctx: CanvasRenderingContext2D, w: number, h: number) => {
    const sim = simRef.current;
    const srcX = 46;
    const midY = h / 2;
    const gateX = Math.round(w * 0.5);
    const sinkX = w - 46;

    // lane guide
    ctx.strokeStyle = 'rgba(255,255,255,0.05)';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(srcX, midY);
    ctx.lineTo(sinkX, midY);
    ctx.stroke();

    // source
    drawBox(ctx, srcX - 26, midY - 22, 52, 44, '#2a3142', '#4F9DF7');
    label(ctx, 'clients', srcX, midY + 38, '#cfd6e4', 'center');
    label(ctx, `${sim.cfg.ratePerSec}/s`, srcX, midY + 4, '#cfd6e4', 'center');

    // token bucket tank at the gate
    const tankW = 60;
    const tankH = h - 80;
    const tankX = gateX - tankW / 2;
    const tankY = 40;
    const fill = Math.max(0, Math.min(1, sim.tokens / sim.cfg.capacity));
    ctx.fillStyle = '#11151d';
    drawRoundRect(ctx, tankX, tankY, tankW, tankH, 8);
    ctx.fill();
    // liquid
    const fh = tankH * fill;
    ctx.fillStyle = '#2ECC71';
    ctx.globalAlpha = 0.85;
    drawRoundRect(ctx, tankX, tankY + (tankH - fh), tankW, fh, 8);
    ctx.fill();
    ctx.globalAlpha = 1;
    ctx.strokeStyle = '#3a4150';
    ctx.lineWidth = 1.5;
    drawRoundRect(ctx, tankX, tankY, tankW, tankH, 8);
    ctx.stroke();
    label(ctx, 'tokens', gateX, tankY - 8, '#9aa3b2', 'center');
    label(ctx, `${Math.floor(sim.tokens)}/${sim.cfg.capacity}`, gateX, tankY + tankH / 2, '#ffffff', 'center', '700 13px');
    label(ctx, `refill ${sim.cfg.refillPerSec}/s`, gateX, tankY + tankH + 16, '#9aa3b2', 'center');

    // sink
    drawBox(ctx, sinkX - 26, midY - 22, 52, 44, '#16241c', '#2ECC71');
    label(ctx, 'service', sinkX, midY + 38, '#cfd6e4', 'center');

    // requests
    for (const r of sim.reqs) {
      const lane = ((r.id * 53) % 80) - 40;
      if (r.state === 'flight') {
        const x = srcX + 24 + (gateX - tankW / 2 - (srcX + 24)) * r.flight;
        ctx.fillStyle = '#4F9DF7';
        dot(ctx, x, midY + lane * (1 - r.flight) * 0.6, 5);
      } else if (r.state === 'pass') {
        const x = gateX + tankW / 2 + (sinkX - (gateX + tankW / 2)) * r.post;
        ctx.fillStyle = `rgba(46,204,113,${1 - r.post * 0.5})`;
        dot(ctx, x, midY, 5);
      } else {
        // deny — bounce up and fade, with a 429 tag
        const a = 1 - r.post;
        ctx.fillStyle = `rgba(231,76,60,${a})`;
        const y = midY - r.post * 70;
        dot(ctx, gateX - tankW / 2 - 6, y, 5);
        if (r.post < 0.6) label(ctx, '429', gateX - tankW / 2 - 6, y - 10, `rgba(231,76,60,${a})`, 'center', '700 10px');
      }
    }
  };

  const canvasRef = useCanvasSim({
    height: HEIGHT,
    playing,
    timeScale: speed,
    ariaLabel: 'Animated rate-limiter simulation: requests metered by a token bucket at the gate.',
    step,
    draw,
  });
  const m = simRef.current.metrics;

  return (
    <div className="lsd-diagram not-content">
      <p className="lsd-hint">
        Each request spends a token at the gate; an empty bucket means <strong>429</strong>. The bucket refills
        steadily, so a quiet period banks a burst allowance. <strong>Raise the rate above the refill</strong> and
        watch the tank drain and denials (red) begin.
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
        text={`Allowed ${m.allowed}, denied ${m.denied}, allow rate ${Math.round(m.allowRate * 100)} percent, ${Math.floor(m.tokens)} tokens in the bucket.`}
      />
      <SimLegend
        items={[
          { glyph: '●', color: '#4F9DF7', label: 'request arriving' },
          { glyph: '●', color: '#2ECC71', label: 'allowed — spent a token' },
          { glyph: '✕', color: '#E74C3C', label: 'denied — 429, bucket empty' },
        ]}
      />

      <div className="lsd-slider-grid">
        <Slider label="request rate" value={rate} min={1} max={40} step={1} suffix="/s" onChange={setRate} />
        <Slider label="bucket capacity" value={capacity} min={1} max={30} step={1} onChange={setCapacity} />
        <Slider label="refill rate" value={refill} min={1} max={15} step={1} suffix="/s" onChange={setRefill} />
      </div>

      <div className="lsd-metric-row">
        <Metric label="tokens" value={`${Math.floor(m.tokens)}/${capacity}`} />
        <Metric label="allowed" value={`${m.allowed}`} />
        <Metric label="denied (429)" value={`${m.denied}`} warn={m.denied > 0} />
        <Metric label="allow rate" value={`${Math.round(m.allowRate * 100)}%`} warn={m.allowRate < 0.6} />
      </div>
    </div>
  );
}

// --- canvas helpers ---
function dot(ctx: CanvasRenderingContext2D, x: number, y: number, r: number) {
  ctx.beginPath();
  ctx.arc(x, y, r, 0, Math.PI * 2);
  ctx.fill();
}
function label(
  ctx: CanvasRenderingContext2D,
  text: string,
  x: number,
  y: number,
  color: string,
  align: CanvasTextAlign = 'left',
  font = '600 11px',
) {
  ctx.fillStyle = color;
  ctx.font = `${font} ui-sans-serif, system-ui, sans-serif`;
  ctx.textAlign = align;
  ctx.textBaseline = 'middle';
  ctx.fillText(text, x, y);
}
function drawBox(ctx: CanvasRenderingContext2D, x: number, y: number, w: number, h: number, fill: string, stroke: string) {
  ctx.fillStyle = fill;
  ctx.strokeStyle = stroke;
  ctx.lineWidth = 1.5;
  drawRoundRect(ctx, x, y, w, h, 10);
  ctx.fill();
  ctx.stroke();
}
function drawRoundRect(ctx: CanvasRenderingContext2D, x: number, y: number, w: number, h: number, r: number) {
  const rr = Math.min(r, w / 2, h / 2);
  ctx.beginPath();
  ctx.moveTo(x + rr, y);
  ctx.arcTo(x + w, y, x + w, y + h, rr);
  ctx.arcTo(x + w, y + h, x, y + h, rr);
  ctx.arcTo(x, y + h, x, y, rr);
  ctx.arcTo(x, y, x + w, y, rr);
  ctx.closePath();
}
