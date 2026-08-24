import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Zap } from "lucide-react";
import { api } from "../lib/api";
import type { RequestStats } from "../lib/types";
import { Panel } from "../components/Panel";
import { Button } from "../components/Button";
import { Field, Input } from "../components/Field";
import { KpiStat } from "../components/KpiStat";
import { Spinner } from "../components/Spinner";

export function LoadTest() {
  const [url, setUrl] = useState("");
  const [method, setMethod] = useState("GET");
  const [workers, setWorkers] = useState(20);
  const [rps, setRps] = useState(100);
  const [count, setCount] = useState(200);

  const run = useMutation({
    mutationFn: () => api.loadtest({ url, method, workers, rps, count }),
  });

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    if (url.trim()) run.mutate();
  };

  return (
    <div className="mx-auto max-w-4xl px-6 py-6">
      <h1 className="mb-1 font-display text-xl font-bold">Load Test</h1>
      <p className="mb-5 text-sm text-muted">Fire concurrent requests and read the latency distribution.</p>

      <Panel title="Configuration">
        <form onSubmit={submit} className="grid grid-cols-2 gap-4 sm:grid-cols-3">
          <div className="col-span-2 sm:col-span-3">
            <Field label="Target URL">
              <Input placeholder="example.com" value={url} onChange={(e) => setUrl(e.target.value)} />
            </Field>
          </div>
          <Field label="Method">
            <select
              value={method}
              onChange={(e) => setMethod(e.target.value)}
              className="rounded-lg border border-border bg-panel2 px-3 py-2 text-sm outline-none focus:border-accent"
            >
              {["GET", "HEAD", "POST"].map((m) => <option key={m}>{m}</option>)}
            </select>
          </Field>
          <Field label="Workers"><Input type="number" value={workers} onChange={(e) => setWorkers(+e.target.value)} /></Field>
          <Field label="RPS"><Input type="number" value={rps} onChange={(e) => setRps(+e.target.value)} /></Field>
          <Field label="Count"><Input type="number" value={count} onChange={(e) => setCount(+e.target.value)} /></Field>
          <div className="col-span-2 flex items-end sm:col-span-3">
            <Button type="submit" disabled={run.isPending}>
              <Zap size={15} /> {run.isPending ? "Running…" : "Run test"}
            </Button>
          </div>
        </form>
      </Panel>

      {run.isPending && <div className="mt-4"><Spinner label="Sending requests…" /></div>}
      {run.error && <p className="mt-4 text-sm text-down">{(run.error as Error).message}</p>}
      {run.data && <Results stats={run.data} />}
    </div>
  );
}

function Results({ stats }: { stats: RequestStats }) {
  const pct = [
    ["p50", stats.p50_latency_ms],
    ["p90", stats.p90_latency_ms],
    ["p95", stats.p95_latency_ms],
    ["p99", stats.p99_latency_ms],
  ] as const;
  const max = Math.max(...pct.map(([, v]) => v), 1);
  return (
    <div className="mt-5 flex flex-col gap-4">
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <KpiStat label="Requests" value={String(stats.total_requests)} />
        <KpiStat label="Success" value={`${stats.success_rate.toFixed(1)}%`} accent="var(--color-up)" />
        <KpiStat label="Throughput" value={`${stats.requests_per_sec.toFixed(0)}/s`} />
        <KpiStat label="Avg" value={`${stats.avg_latency_ms}ms`} />
      </div>
      <Panel title="Latency percentiles">
        <div className="flex flex-col gap-2">
          {pct.map(([k, v]) => (
            <div key={k} className="flex items-center gap-3">
              <span className="w-10 font-mono text-xs text-muted">{k}</span>
              <div className="h-4 flex-1 overflow-hidden rounded bg-panel2">
                <div className="h-full rounded bg-accent" style={{ width: `${(v / max) * 100}%` }} />
              </div>
              <span className="w-16 text-right font-mono text-xs tabular-nums">{v}ms</span>
            </div>
          ))}
        </div>
      </Panel>
      <Panel title="Status codes">
        <div className="flex flex-wrap gap-2">
          {Object.entries(stats.status_codes).map(([code, n]) => (
            <span key={code} className="rounded-lg border border-border bg-panel2 px-2.5 py-1 font-mono text-xs">
              {code}: <span className="tabular-nums">{n}</span>
            </span>
          ))}
        </div>
      </Panel>
    </div>
  );
}
