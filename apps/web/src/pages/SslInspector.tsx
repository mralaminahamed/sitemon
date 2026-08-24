import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { ShieldCheck } from "lucide-react";
import { api } from "../lib/api";
import type { CertificateInfo } from "../lib/types";
import { Panel } from "../components/Panel";
import { Button } from "../components/Button";
import { Input } from "../components/Field";
import { Spinner } from "../components/Spinner";

export function SslInspector() {
  const [url, setUrl] = useState("");
  const q = useMutation({ mutationFn: (u: string) => api.ssl(u) });

  return (
    <div className="mx-auto max-w-3xl px-6 py-6">
      <h1 className="mb-1 font-display text-xl font-bold">SSL Inspector</h1>
      <p className="mb-5 text-sm text-muted">Check any host's TLS certificate.</p>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (url.trim()) q.mutate(url.trim());
        }}
        className="mb-5 flex gap-2"
      >
        <Input placeholder="example.com" value={url} onChange={(e) => setUrl(e.target.value)} className="flex-1" />
        <Button type="submit" disabled={q.isPending}>
          <ShieldCheck size={15} /> Inspect
        </Button>
      </form>

      {q.isPending && <Spinner label="Connecting…" />}
      {q.error && <p className="text-sm text-down">{(q.error as Error).message}</p>}
      {q.data && <Cert info={q.data} />}
    </div>
  );
}

function Cert({ info }: { info: CertificateInfo }) {
  const dayColor = info.days_remaining < 7 ? "var(--color-down)" : info.days_remaining < 30 ? "var(--color-warn)" : "var(--color-up)";
  const rows = [
    ["Host", info.host],
    ["Issuer", info.issuer],
    ["Subject", info.subject],
    ["Protocol", info.protocol],
    ["Valid from", new Date(info.valid_from).toLocaleString()],
    ["Valid until", new Date(info.valid_until).toLocaleString()],
    ["Days remaining", String(info.days_remaining)],
  ];
  return (
    <Panel title={info.is_valid ? "Valid certificate" : "Invalid certificate"}>
      <div className="mb-4 flex items-baseline gap-2">
        <span className="font-mono text-4xl font-bold tabular-nums" style={{ color: dayColor }}>
          {info.days_remaining}
        </span>
        <span className="text-sm text-muted">days left</span>
      </div>
      <dl className="divide-y divide-border">
        {rows.map(([k, v]) => (
          <div key={k} className="flex justify-between gap-4 py-2 text-sm">
            <dt className="text-muted">{k}</dt>
            <dd className="truncate text-right font-mono">{v}</dd>
          </div>
        ))}
      </dl>
    </Panel>
  );
}
