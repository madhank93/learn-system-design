/** Shared control + metric primitives for the canvas flow sims. */

export function Slider(props: {
  label: string;
  value: number;
  min: number;
  max: number;
  step: number;
  suffix?: string;
  pct?: boolean;
  onChange: (v: number) => void;
}) {
  const shown = props.pct ? `${Math.round(props.value * 100)}%` : `${props.value}${props.suffix ?? ''}`;
  return (
    <label className="lsd-slider">
      <span className="lsd-slider-label">
        {props.label}
        <strong>{shown}</strong>
      </span>
      <input
        className="lsd-range"
        type="range"
        min={props.min}
        max={props.max}
        step={props.step}
        value={props.value}
        onChange={(e) => props.onChange(parseFloat(e.target.value))}
      />
    </label>
  );
}

export function Metric({ label, value, warn }: { label: string; value: string; warn?: boolean }) {
  return (
    <div className="lsd-metric">
      <span className="lsd-metric-label">{label}</span>
      <span className={warn ? 'lsd-metric-value is-warn' : 'lsd-metric-value'}>
        {value}
      </span>
    </div>
  );
}

export type LegendItem = { glyph: string; color: string; label: string };

/**
 * Legend for a canvas sim. State is encoded by glyph AND colour (never colour
 * alone) so it survives colour-blindness and greyscale. The glyph mirrors what
 * the canvas draws (●, ✕, ▲, ○…).
 */
export function SimLegend({ items }: { items: LegendItem[] }) {
  return (
    <ul className="lsd-simlegend" aria-label="Legend">
      {items.map((it) => (
        <li key={it.label}>
          <span className="lsd-simlegend-glyph" style={{ color: it.color }} aria-hidden>
            {it.glyph}
          </span>
          {it.label}
        </li>
      ))}
    </ul>
  );
}

/**
 * Visually-hidden, polite live region. Gives screen-reader users the running
 * state of a canvas sim (which is otherwise invisible to them). Update it from
 * the same ~200ms tick the metrics panel uses.
 */
export function LiveSummary({ text }: { text: string }) {
  return (
    <p className="sr-only" aria-live="polite" role="status">
      {text}
    </p>
  );
}
