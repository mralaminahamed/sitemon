
import { useStatus } from "../hooks";
import { useUI } from "../store";
import { StatusBadge } from "./StatusBadge";

export function StatusGrid() {
  const { data, isLoading, error } = useStatus();
  const { selectedUrl, select } = useUI();

  if (isLoading) return <p className="muted">Loading status…</p>;
  if (error) return <p className="error">{(error as Error).message}</p>;

  const results = data?.results ?? [];
  if (results.length === 0)
    return <p className="muted">No monitored URLs yet. Run a check below.</p>;

  return (
    <div className="grid">
      {results.map((r) => (
        <button
          key={r.url}
          className={`card ${selectedUrl === r.url ? "card--active" : ""}`}
          onClick={() => select(r.url)}
        >
          <div className="card__top">
            <StatusBadge status={r.status} />
            <span className="code">{r.status_code || "—"}</span>
          </div>
          <div className="card__url" title={r.url}>
            {r.url}
          </div>
          <div className="card__meta">
            <span>{r.response_time_ms} ms</span>
            <span className="muted">
              {new Date(r.timestamp).toLocaleTimeString()}
            </span>
          </div>
        </button>
      ))}
    </div>
  );
}
