import { Link } from "react-router-dom";
import { Trash2 } from "lucide-react";
import type { HealthResult } from "../lib/types";
import { ago, ms, statusColor } from "../lib/format";
import { StatusBadge } from "./StatusDot";

export function MonitorCard({ result, onDelete }: {
  result: HealthResult;
  onDelete?: (url: string) => void;
}) {
  const c = statusColor(result.status);
  return (
    <Link
      to={`/monitors/${encodeURIComponent(result.url)}`}
      className="group relative block overflow-hidden rounded-xl border border-border bg-panel p-4 transition hover:border-accent"
    >
      <span className="absolute inset-x-0 top-0 h-0.5" style={{ background: c }} />
      <div className="flex items-center justify-between">
        <StatusBadge status={result.status} />
        <span className="font-mono text-xs text-muted">{result.status_code || "—"}</span>
      </div>
      <div className="mt-3 truncate font-medium" title={result.url}>
        {result.url.replace(/^https?:\/\//, "")}
      </div>
      <div className="mt-2 flex items-center justify-between font-mono text-xs">
        <span className="tabular-nums text-text">{ms(result.response_time_ms)}</span>
        <span className="text-muted">{ago(result.timestamp)}</span>
      </div>
      {onDelete && (
        <button
          onClick={(e) => {
            e.preventDefault();
            onDelete(result.url);
          }}
          className="absolute bottom-3 right-3 rounded p-1 text-muted opacity-0 transition hover:text-down focus-visible:opacity-100 group-hover:opacity-100"
          aria-label="Remove monitor"
        >
          <Trash2 size={14} />
        </button>
      )}
    </Link>
  );
}
