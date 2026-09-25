import { useEffect, useRef, useState } from 'react';

/**
 * True when the user has `prefers-reduced-motion: reduce`. Flow components use
 * it to default to paused (no autoplay) — the sim is still fully interactive via
 * Play / step, it just doesn't move on its own.
 */
export function usePrefersReducedMotion(): boolean {
  const [reduced, setReduced] = useState(false);
  useEffect(() => {
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)');
    setReduced(mq.matches);
    const on = () => setReduced(mq.matches);
    mq.addEventListener('change', on);
    return () => mq.removeEventListener('change', on);
  }, []);
  return reduced;
}

/**
 * Shared canvas + requestAnimationFrame harness for the flow simulations.
 * Each frame: `step(dtMs)` advances the sim (clamped + time-scaled), then
 * `draw(ctx, w, h)` paints it. Handles devicePixelRatio scaling and width
 * via ResizeObserver so components are just simulation + drawing.
 *
 * `step`/`draw` are read from refs each frame, so passing fresh closures (that
 * capture current config/state) every render is fine — no loop restart.
 *
 * The loop only runs while the canvas is on screen; scrolled away, the sim
 * freezes rather than burning frames nobody sees.
 *
 * Accessibility: the canvas is exposed as `role="img"` with `ariaLabel`; pair it
 * with a visually-hidden `aria-live` summary in the component for live metrics,
 * and default `playing` to false when `usePrefersReducedMotion()` is true.
 */
export function useCanvasSim(opts: {
  height: number;
  playing: boolean;
  timeScale?: number;
  ariaLabel?: string;
  step: (dtMs: number) => void;
  draw: (ctx: CanvasRenderingContext2D, w: number, h: number) => void;
}) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const ref = useRef(opts);
  ref.current = opts;

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    if (ref.current.ariaLabel) {
      canvas.setAttribute('role', 'img');
      canvas.setAttribute('aria-label', ref.current.ariaLabel);
    }

    let raf = 0;
    let last = performance.now();
    let cssW = canvas.clientWidth || 600;
    const { height } = ref.current;

    const resize = () => {
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      cssW = canvas.clientWidth || 600;
      canvas.width = Math.round(cssW * dpr);
      canvas.height = Math.round(height * dpr);
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    };
    resize();
    const ro = new ResizeObserver(resize);
    ro.observe(canvas);

    let visible = true;
    const frame = (now: number) => {
      if (!visible) {
        raf = 0;
        return;
      }
      const o = ref.current;
      let dt = now - last;
      last = now;
      dt = Math.min(dt, 50); // clamp tab-switch jumps
      if (o.playing) o.step(dt * (o.timeScale ?? 1));
      ctx.clearRect(0, 0, cssW, height);
      o.draw(ctx, cssW, height);
      raf = requestAnimationFrame(frame);
    };
    raf = requestAnimationFrame(frame);

    const io = new IntersectionObserver(([entry]) => {
      visible = entry.isIntersecting;
      if (visible && raf === 0) {
        last = performance.now();
        raf = requestAnimationFrame(frame);
      }
    });
    io.observe(canvas);

    return () => {
      cancelAnimationFrame(raf);
      ro.disconnect();
      io.disconnect();
    };
  }, []);

  return canvasRef;
}
