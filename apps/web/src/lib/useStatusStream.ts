import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { HealthResult } from "./types";
import { getApiKey } from "./api";
import { useToasts } from "./toast";
import { isDown } from "./format";

interface Msg {
  type: "snapshot" | "result";
  results?: HealthResult[];
  result?: HealthResult;
}

type Cache = { results: HealthResult[] };

// Live status over WebSocket → query cache, with a toast on any UP<->down flip.
export function useStatusStream() {
  const qc = useQueryClient();
  const push = useToasts((s) => s.push);

  useEffect(() => {
    const proto = location.protocol === "https:" ? "wss" : "ws";
    const key = getApiKey();
    const qs = key ? `?api_key=${encodeURIComponent(key)}` : "";
    let sock: WebSocket;
    let closed = false;

    const connect = () => {
      sock = new WebSocket(`${proto}://${location.host}/ws${qs}`);
      sock.onmessage = (e) => {
        const msg: Msg = JSON.parse(e.data);
        if (msg.type === "snapshot") {
          qc.setQueryData<Cache>(["status"], { results: msg.results ?? [] });
        } else if (msg.type === "result" && msg.result) {
          const r = msg.result;
          const prev = qc
            .getQueryData<Cache>(["status"])
            ?.results.find((x) => x.url === r.url);
          if (prev && isDown(prev.status) !== isDown(r.status)) {
            const recovered = !isDown(r.status);
            push({
              title: `${r.url.replace(/^https?:\/\//, "")} ${recovered ? "recovered" : "is down"}`,
              detail: `${r.status} · ${r.status_code || "—"}`,
              tone: recovered ? "up" : "down",
            });
          }
          qc.setQueryData<Cache>(["status"], (old) => {
            const rest = (old?.results ?? []).filter((x) => x.url !== r.url);
            return { results: [...rest, r] };
          });
        }
      };
      sock.onclose = () => {
        if (!closed) setTimeout(connect, 3000);
      };
    };
    connect();

    return () => {
      closed = true;
      sock?.close();
    };
  }, [qc, push]);
}
