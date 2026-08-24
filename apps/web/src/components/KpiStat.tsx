export function KpiStat({ label, value, sub, accent }: {
  label: string;
  value: string;
  sub?: string;
  accent?: string;
}) {
  return (
    <div className="rounded-xl border border-border bg-panel p-4">
      <div className="font-mono text-[11px] uppercase tracking-wide text-muted">{label}</div>
      <div className="mt-1 font-mono text-2xl font-semibold tabular-nums" style={{ color: accent }}>
        {value}
      </div>
      {sub && <div className="mt-0.5 text-xs text-muted">{sub}</div>}
    </div>
  );
}
