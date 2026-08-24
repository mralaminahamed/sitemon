// Signature motif: a live heartbeat line, colored by status.
export function Pulse({ color = "var(--color-accent)", width = 120, height = 28 }: {
  color?: string;
  width?: number;
  height?: number;
}) {
  return (
    <svg width={width} height={height} viewBox="0 0 120 28" fill="none" aria-hidden>
      <polyline
        className="pulse-line"
        points="0,14 26,14 34,6 46,22 58,3 68,14 120,14"
        stroke={color}
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
