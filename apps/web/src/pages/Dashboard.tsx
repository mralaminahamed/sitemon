import { useState } from "react";
import { Plus } from "lucide-react";
import { useAddMonitor, useDeleteMonitor, useMonitors, useStatus } from "../lib/hooks";
import type { HealthResult } from "../lib/types";
import { isDown } from "../lib/format";
import { useToasts } from "../lib/toast";
import { KpiStat } from "../components/KpiStat";
import { MonitorCard } from "../components/MonitorCard";
import { Button } from "../components/Button";
import { Input } from "../components/Field";
import { Empty, Spinner } from "../components/Spinner";

export function Dashboard() {
  const status = useStatus();
  const monitors = useMonitors();
  const addMon = useAddMonitor();
  const delMon = useDeleteMonitor();
  const push = useToasts((s) => s.push);
  const [url, setUrl] = useState("");
  const [filter, setFilter] = useState("");

  const remove = (u: string) => {
    if (!window.confirm(`Stop monitoring ${u.replace(/^https?:\/\//, "")}?`)) return;
    delMon.mutate(u, {
      onError: (e) => push({ title: "Delete failed", detail: (e as Error).message, tone: "down" }),
    });
  };

  const results = status.data?.results ?? [];
  const byUrl = new Map(results.map((r) => [r.url, r]));

  // Union of monitored URLs and anything that already has a status.
  const urls = new Set<string>([
    ...(monitors.data?.monitors ?? []).map((m) => m.url),
    ...results.map((r) => r.url),
  ]);
  const rows: HealthResult[] = [...urls]
    .map(
      (u) =>
        byUrl.get(u) ?? {
          url: u,
          status: "UNKNOWN" as HealthResult["status"],
          status_code: 0,
          response_time_ms: 0,
          timestamp: "",
        },
    )
    .filter((r) => r.url.toLowerCase().includes(filter.toLowerCase()))
    .sort((a, b) => Number(isDown(b.status)) - Number(isDown(a.status)) || a.url.localeCompare(b.url));

  const total = urls.size;
  const down = results.filter((r) => isDown(r.status)).length;
  const up = results.filter((r) => r.status === "UP").length;
  const avg =
    results.length > 0
      ? Math.round(results.reduce((s, r) => s + r.response_time_ms, 0) / results.length)
      : 0;
  const worst = results.reduce((m, r) => Math.max(m, r.response_time_ms), 0);

  const add = (e: React.FormEvent) => {
    e.preventDefault();
    const v = url.trim();
    if (v) addMon.mutate(v, { onSuccess: () => setUrl("") });
  };

  const loadError = (status.error || monitors.error) as Error | null;

  return (
    <div className="mx-auto max-w-6xl px-6 py-6">
      <div className="mb-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
        <KpiStat label="Monitors" value={String(total)} sub={`${up} up · ${down} down`} />
        <KpiStat label="Uptime now" value={total ? `${Math.round((up / Math.max(total, 1)) * 100)}%` : "—"} accent="var(--color-up)" />
        <KpiStat label="Avg latency" value={avg ? `${avg}ms` : "—"} />
        <KpiStat label="Worst latency" value={worst ? `${Math.round(worst)}ms` : "—"} accent={worst > 1000 ? "var(--color-warn)" : undefined} />
      </div>

      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <h1 className="font-display text-xl font-bold">Monitors</h1>
        <div className="flex items-center gap-2">
          <Input placeholder="Filter…" value={filter} onChange={(e) => setFilter(e.target.value)} className="w-40" />
          <form onSubmit={add} className="flex items-center gap-2">
            <Input placeholder="example.com" value={url} onChange={(e) => setUrl(e.target.value)} className="w-52" />
            <Button type="submit" disabled={addMon.isPending}>
              <Plus size={15} /> Add
            </Button>
          </form>
        </div>
      </div>

      {addMon.error && <p className="mb-3 text-sm text-down">{(addMon.error as Error).message}</p>}
      {loadError && (
        <p className="mb-3 text-sm text-down">Couldn't reach the API: {loadError.message}</p>
      )}

      {status.isLoading && monitors.isLoading ? (
        <Spinner label="Loading monitors…" />
      ) : rows.length === 0 ? (
        <Empty>
          {loadError
            ? "Couldn't load monitors. Retrying…"
            : "No monitors yet. Add a URL above to start watching it."}
        </Empty>
      ) : (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {rows.map((r) => (
            <MonitorCard key={r.url} result={r} onDelete={remove} />
          ))}
        </div>
      )}
    </div>
  );
}
