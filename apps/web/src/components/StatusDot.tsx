import { statusColor } from "../lib/format";

export function StatusDot({ status, live = false }: { status: string; live?: boolean }) {
  const c = statusColor(status);
  return (
    <span
      className={`inline-block h-2.5 w-2.5 rounded-full ${live ? "blip" : ""}`}
      style={{ background: c, boxShadow: `0 0 8px ${c}` }}
    />
  );
}

export function StatusBadge({ status }: { status: string }) {
  const c = statusColor(status);
  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 font-mono text-[11px] font-semibold uppercase tracking-wide"
      style={{ color: c, background: `color-mix(in oklab, ${c} 14%, transparent)` }}
    >
      <span className="h-1.5 w-1.5 rounded-full" style={{ background: c }} />
      {status}
    </span>
  );
}
