import { useMemo, useRef, useState } from 'react';
import {
  hashDeg,
  buildRing,
  ownerOf,
  loadDistribution,
  movedIndices,
} from '../sim-lib/consistent-hash';

/**
 * Interactive consistent-hashing ring — the hero proof.
 * Replaces 3 static D2 SVGs (overview / node_removed / hotspot) with one thing
 * you can poke. Ownership is drawn as colored arcs so "which node owns what" is
 * obvious; keys that move on a topology change flash; hovering a key draws its
 * route to the owning node. Ring math lives in sim-lib/consistent-hash (pure +
 * unit tested); this file keeps only the SVG geometry and interaction.
 */

const PALETTE = ['#4F9DF7', '#2ECC71', '#E67E22', '#A66CFF', '#E74C3C', '#1ABC9C', '#F1C40F'];
const CX = 190;
const CY = 190;
const R = 140; // ring radius (arcs sit here)
const KEY_R = R - 30; // keys sit just inside

function polar(deg: number, radius = R) {
  const t = (deg / 360) * 2 * Math.PI - Math.PI / 2;
  return { x: CX + radius * Math.cos(t), y: CY + radius * Math.sin(t) };
}

function arcPath(startDeg: number, endDeg: number, radius: number) {
  const s = polar(startDeg, radius);
  const e = polar(endDeg, radius);
  const delta = (endDeg - startDeg + 360) % 360;
  const large = delta > 180 ? 1 : 0;
  return `M ${s.x} ${s.y} A ${radius} ${radius} 0 ${large} 1 ${e.x} ${e.y}`;
}

const initialKeys = () => Array.from({ length: 14 }, (_, i) => hashDeg(`key_${i}`));

export default function ConsistentHashingRing() {
  const [nodeCount, setNodeCount] = useState(3);
  const [vnodes, setVnodes] = useState(1);
  const [keys, setKeys] = useState<number[]>(initialKeys);
  const [remap, setRemap] = useState<number | null>(null);
  const [moved, setMoved] = useState<Set<number>>(new Set());
  const [hover, setHover] = useState<number | null>(null);
  const svgRef = useRef<SVGSVGElement>(null);
  const movedTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const ring = useMemo(() => buildRing(nodeCount, vnodes), [nodeCount, vnodes]);
  const ownership = useMemo(() => keys.map((k) => ownerOf(k, ring)), [keys, ring]);

  const load = useMemo(() => loadDistribution(ownership, nodeCount), [ownership, nodeCount]);

  const arcs = useMemo(() => {
    if (ring.length === 0) return [];
    return ring.map((v, i) => {
      const prev = ring[(i - 1 + ring.length) % ring.length];
      return { from: prev.deg, to: v.deg, node: v.node };
    });
  }, [ring]);

  function flashMoved(indices: Set<number>) {
    setMoved(indices);
    if (movedTimer.current) clearTimeout(movedTimer.current);
    movedTimer.current = setTimeout(() => setMoved(new Set()), 2600);
  }

  function changeTopology(nextNodes: number, nextV: number) {
    const nextRing = buildRing(nextNodes, nextV);
    const movedSet = movedIndices(keys, ring, nextRing);
    setRemap(keys.length ? Math.round((movedSet.size / keys.length) * 100) : 0);
    flashMoved(movedSet);
    setNodeCount(nextNodes);
    setVnodes(nextV);
  }

  function addKeyAt(evt: React.MouseEvent<SVGSVGElement>) {
    const svg = svgRef.current;
    if (!svg) return;
    const box = svg.getBoundingClientRect();
    const x = ((evt.clientX - box.left) / box.width) * 380;
    const y = ((evt.clientY - box.top) / box.height) * 380;
    let deg = (Math.atan2(y - CY, x - CX) * 180) / Math.PI + 90;
    deg = ((deg % 360) + 360) % 360;
    setKeys((p) => [...p, deg]);
    setRemap(null);
  }

  function reset() {
    setNodeCount(3);
    setVnodes(1);
    setKeys(initialKeys());
    setRemap(null);
    setMoved(new Set());
  }

  const modN = Math.round(((nodeCount - 1) / Math.max(1, nodeCount)) * 100);
  const balance = load.length > 1 ? Math.max(...load) - Math.min(...load) : 0;

  return (
    <div className="lsd-diagram not-content">
      <p className="lsd-hint">
        <strong>S0, S1, S2 are your servers</strong> (cache/storage nodes) placed on the ring; colored
        dots are keys, each owned by the first server clockwise (its colored arc).
        <br />
        👆 <strong>Click the ring</strong> to drop a key. <strong>Add a node</strong> and watch how few
        keys move (they flash). Raise <strong>vnodes</strong> to even out the load bars. Hover a key to
        see where it routes.
      </p>

      <div className="lsd-controls">
        <button onClick={() => changeTopology(nodeCount + 1, vnodes)} disabled={nodeCount >= 7}>
          ＋ Add node
        </button>
        <button onClick={() => changeTopology(nodeCount - 1, vnodes)} disabled={nodeCount <= 1}>
          － Remove node
        </button>
        <span style={{ opacity: 0.8 }}>vnodes/node:</span>
        <span className="lsd-seg">
          {[1, 3, 8].map((v) => (
            <button key={v} aria-pressed={vnodes === v} onClick={() => changeTopology(nodeCount, v)}>
              {v}
            </button>
          ))}
        </span>
        <button onClick={() => { setKeys(Array.from({ length: 14 }, (_, i) => hashDeg(`k${i}_${Math.random()}`))); setRemap(null); }}>
          ↻ Reshuffle keys
        </button>
        <button onClick={reset}>↺ Reset</button>
      </div>

      <div style={{ display: 'flex', gap: '1.25rem', flexWrap: 'wrap', alignItems: 'flex-start' }}>
        <svg
          ref={svgRef}
          viewBox="0 0 380 380"
          width="380"
          height="380"
          onClick={addKeyAt}
          style={{ cursor: 'crosshair', maxWidth: '100%', flex: '0 0 auto' }}
        >
          <circle cx={CX} cy={CY} r={R} fill="none" stroke="#3a3f4b" strokeWidth={14} />
          {/* ownership arcs */}
          {arcs.map((a, i) => (
            <path
              key={`a${i}`}
              d={arcPath(a.from, a.to, R)}
              fill="none"
              stroke={PALETTE[a.node % PALETTE.length]}
              strokeWidth={14}
              strokeLinecap="butt"
              opacity={0.85}
            />
          ))}
          {/* node markers */}
          {ring.map((v, i) => {
            const p = polar(v.deg, R);
            const lp = polar(v.deg, R + 20);
            return (
              <g key={`n${i}`}>
                <title>Server {v.node}</title>
                <circle cx={p.x} cy={p.y} r={6} fill={PALETTE[v.node % PALETTE.length]} stroke="#10131a" strokeWidth={2} />
                {vnodes === 1 && (
                  <text x={lp.x} y={lp.y + 4} textAnchor="middle" fontSize="12" fontWeight={700} fill={PALETTE[v.node % PALETTE.length]}>
                    S{v.node}
                  </text>
                )}
              </g>
            );
          })}
          {/* routing line on hover */}
          {hover !== null && ownership[hover] >= 0 && (() => {
            const kp = polar(keys[hover], KEY_R);
            // nearest owning vnode marker clockwise
            const ownerVnode = ring.find((v) => v.deg >= keys[hover] && v.node === ownership[hover]) ?? ring.find((v) => v.node === ownership[hover]);
            const np = ownerVnode ? polar(ownerVnode.deg, R) : kp;
            return <line x1={kp.x} y1={kp.y} x2={np.x} y2={np.y} stroke="#fff" strokeWidth={1.5} strokeDasharray="3 3" />;
          })()}
          {/* keys */}
          {keys.map((k, i) => {
            const p = polar(k, KEY_R);
            const owner = ownership[i];
            const color = owner >= 0 ? PALETTE[owner % PALETTE.length] : '#999';
            return (
              <g key={`k${i}`} onMouseEnter={() => setHover(i)} onMouseLeave={() => setHover(null)} style={{ cursor: 'pointer' }}>
                {moved.has(i) && <circle cx={p.x} cy={p.y} r={5} fill={color} style={{ animation: 'lsd-pulse 1.3s ease-out 2' }} />}
                <circle cx={p.x} cy={p.y} r={hover === i ? 7 : 5} fill={color} stroke={moved.has(i) ? '#fff' : '#10131a'} strokeWidth={moved.has(i) ? 2 : 1} />
              </g>
            );
          })}
          <text x={CX} y={CY - 2} textAnchor="middle" fill="#8b93a3" fontSize="13">hash ring</text>
          <text x={CX} y={CY + 16} textAnchor="middle" fill="#5c6473" fontSize="10">0 → 2³²−1, clockwise</text>
        </svg>

        <div style={{ minWidth: 220, flex: 1 }}>
          <p className="lsd-stat" style={{ fontSize: '0.95rem' }}>
            On the last change, <strong style={{ color: remap !== null && remap <= modN / 2 ? '#2ECC71' : undefined }}>
              {remap === null ? '—' : `${remap}%`}
            </strong> of keys moved.
          </p>
          <p className="lsd-stat" style={{ marginTop: -6 }}>
            Plain <code>hash % N</code> would have moved ≈<strong>{modN}%</strong>.
          </p>

          <p className="lsd-stat" style={{ marginTop: '0.75rem' }}>
            <strong>Keys owned per server</strong> &nbsp;(spread max−min: {balance})
          </p>
          <div className="lsd-legend" style={{ marginTop: 4 }}>
            {load.map((c, n) => {
              const pct = keys.length ? Math.round((c / keys.length) * 100) : 0;
              return (
                <div className="lsd-legend-row" key={n}>
                  <span className="lsd-dot" style={{ background: PALETTE[n % PALETTE.length] }} />
                  <span className="lsd-stat" style={{ width: 26 }}>S{n}</span>
                  <span className="lsd-bar-track">
                    <span className="lsd-bar-fill" style={{ width: `${pct}%`, background: PALETTE[n % PALETTE.length] }} />
                  </span>
                  <span className="lsd-stat" style={{ width: 56 }}>{c} ({pct}%)</span>
                </div>
              );
            })}
          </div>
          <p className="lsd-stat" style={{ marginTop: '0.75rem', opacity: 0.8 }}>
            {keys.length} keys · {nodeCount} servers · {vnodes} vnode(s)/server = {ring.length} ring positions
          </p>
        </div>
      </div>
    </div>
  );
}
