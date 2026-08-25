import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, Sparkles } from "lucide-react";
import { Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { api } from "../lib/api";
import { useHistory, useStats } from "../lib/hooks";
import { Panel } from "../components/Panel";
import { Tabs } from "../components/Tabs";
import { KpiStat } from "../components/KpiStat";
import { Button } from "../components/Button";
import { Empty, Spinner } from "../components/Spinner";

export function MonitorDetail() {
  // React Router already decodes path params — decoding again throws URIError on
  // any URL containing a literal '%'.
  const { url: target = "" } = useParams();
  const [tab, setTab] = useState("overview");

  return (
    <div className="mx-auto max-w-5xl px-6 py-6">
      <Link to="/" className="mb-4 inline-flex items-center gap-1.5 text-sm text-muted hover:text-text">
        <ArrowLeft size={15} /> Monitors
      </Link>
      <h1 className="mb-1 break-all font-display text-xl font-bold">{target}</h1>
      <div className="mb-5">
        <Tabs
          tabs={[
            { id: "overview", label: "Overview" },
            { id: "ssl", label: "SSL" },
            { id: "ai", label: "AI Analysis" },
          ]}
          active={tab}
          onChange={setTab}
        />
      </div>

      {tab === "overview" && <Overview url={target} />}
      {tab === "ssl" && <SslTab url={target} />}
      {tab === "ai" && <AiTab url={target} />}
    </div>
  );
}

function Overview({ url }: { url: string }) {
  const history = useHistory(url);
  const stats = useStats(url);
  const rows = (history.data?.results ?? []).slice().reverse();
  const points = rows.map((r) => ({
    t: r.timestamp ? new Date(r.timestamp).toLocaleTimeString() : "",
    ms: r.response_time_ms,
  }));
  const s = stats.data;
  const err = (history.error || stats.error) as Error | null;

  return (
    <div className="flex flex-col gap-4">
      {err && <p className="text-sm text-down">Couldn't load metrics: {err.message}</p>}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <KpiStat label="Checks" value={s ? String(s.total_checks) : "—"} />
        <KpiStat label="Uptime" value={s ? `${s.uptime_percentage.toFixed(1)}%` : "—"} accent="var(--color-up)" />
        <KpiStat label="Failed" value={s ? String(s.failed) : "—"} accent={s && s.failed > 0 ? "var(--color-down)" : undefined} />
        <KpiStat label="Avg latency" value={s ? `${Math.round(s.avg_latency_ms)}ms` : "—"} />
      </div>

      <Panel title="Latency">
        {points.length < 2 ? (
          <Empty>Not enough history yet.</Empty>
        ) : (
          <ResponsiveContainer width="100%" height={240}>
            <LineChart data={points} margin={{ top: 8, right: 12, bottom: 0, left: -12 }}>
              <XAxis dataKey="t" tick={{ fontSize: 10, fill: "var(--color-muted)" }} minTickGap={40} />
              <YAxis tick={{ fontSize: 10, fill: "var(--color-muted)" }} unit="ms" width={46} />
              <Tooltip
                contentStyle={{
                  background: "var(--color-panel2)",
                  border: "1px solid var(--color-border)",
                  borderRadius: 8,
                  fontSize: 12,
                }}
              />
              <Line type="monotone" dataKey="ms" stroke="var(--color-accent)" strokeWidth={2} dot={false} isAnimationActive={false} />
            </LineChart>
          </ResponsiveContainer>
        )}
      </Panel>
    </div>
  );
}

function SslTab({ url }: { url: string }) {
  const q = useQuery({ queryKey: ["ssl", url], queryFn: () => api.ssl(url) });
  if (q.isLoading) return <Spinner label="Fetching certificate…" />;
  if (q.error) return <Empty>{(q.error as Error).message}</Empty>;
  const c = q.data!;
  const dayColor = c.days_remaining < 7 ? "var(--color-down)" : c.days_remaining < 30 ? "var(--color-warn)" : "var(--color-up)";
  const rows = [
    ["Issuer", c.issuer],
    ["Subject", c.subject],
    ["Protocol", c.protocol],
    ["Valid from", new Date(c.valid_from).toLocaleDateString()],
    ["Valid until", new Date(c.valid_until).toLocaleDateString()],
  ];
  return (
    <div className="grid gap-4 sm:grid-cols-3">
      <Panel title="Days remaining" className="sm:col-span-1">
        <div className="flex flex-col items-center py-4">
          <div className="font-mono text-5xl font-bold tabular-nums" style={{ color: dayColor }}>
            {c.days_remaining}
          </div>
          <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-panel2">
            <div className="h-full rounded-full" style={{ width: `${Math.min(100, (c.days_remaining / 90) * 100)}%`, background: dayColor }} />
          </div>
          <div className="mt-2 text-xs text-muted">{c.is_valid ? "valid" : "invalid"}</div>
        </div>
      </Panel>
      <Panel title="Certificate" className="sm:col-span-2">
        <dl className="divide-y divide-border">
          {rows.map(([k, v]) => (
            <div key={k} className="flex justify-between gap-4 py-2 text-sm">
              <dt className="text-muted">{k}</dt>
              <dd className="truncate text-right font-mono">{v}</dd>
            </div>
          ))}
        </dl>
      </Panel>
    </div>
  );
}

function AiTab({ url }: { url: string }) {
  const [run, setRun] = useState(false);
  const q = useQuery({ queryKey: ["analyze", url], queryFn: () => api.analyze(url), enabled: run });
  const sevColor: Record<string, string> = {
    ok: "var(--color-up)",
    warn: "var(--color-warn)",
    critical: "var(--color-down)",
  };

  if (!run) {
    return (
      <Panel title="AI incident analysis">
        <div className="flex flex-col items-start gap-3">
          <p className="text-sm text-muted">
            Analyze recent history for anomalies and a written incident summary.
          </p>
          <Button onClick={() => setRun(true)}>
            <Sparkles size={15} /> Analyze
          </Button>
        </div>
      </Panel>
    );
  }
  if (q.isLoading) return <Spinner label="Analyzing…" />;
  if (q.error) return <Empty>{(q.error as Error).message}</Empty>;
  const a = q.data!;
  return (
    <div className="flex flex-col gap-4">
      <Panel>
        <div className="flex items-center justify-between">
          <div>
            <div className="font-mono text-[11px] uppercase tracking-wide text-muted">Severity</div>
            <div className="mt-1 font-display text-2xl font-bold capitalize" style={{ color: sevColor[a.severity] }}>
              {a.severity}
            </div>
          </div>
          <div className="text-right font-mono text-sm text-muted">
            <div>{a.uptime_pct.toFixed(1)}% uptime</div>
            <div>{Math.round(a.avg_ms)}ms avg · {Math.round(a.p95_ms)}ms p95</div>
          </div>
        </div>
      </Panel>
      {a.summary && (
        <Panel title="Summary">
          <p className="text-sm leading-relaxed">{a.summary}</p>
        </Panel>
      )}
      <Panel title="Anomalies">
        {a.anomalies.length === 0 ? (
          <p className="text-sm text-muted">None detected in the last {a.window} checks.</p>
        ) : (
          <ul className="flex flex-col gap-2">
            {a.anomalies.map((x) => (
              <li key={x} className="flex items-center gap-2 text-sm">
                <span className="h-1.5 w-1.5 rounded-full bg-warn" /> {x}
              </li>
            ))}
          </ul>
        )}
      </Panel>
      {!a.summary && (
        <p className="text-xs text-muted">Set ANTHROPIC_API_KEY on the ai service for a written summary.</p>
      )}
    </div>
  );
}
