import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { HealthResult } from "./types";
import { getApiKey } from "./api";
import { useToasts } from "./toast";
import { useConn } from "./conn";
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
    let attempt = 0;
    let timer: ReturnType<typeof setTimeout>;
    const setConnected = useConn.getState().setConnected;

    const connect = () => {
      sock = new WebSocket(`${proto}://${location.host}/ws${qs}`);
      sock.onopen = () => {
        attempt = 0;
        setConnected(true);
      };
      sock.onerror = () => {
        // Surfaced via onclose, which schedules the retry.
      };
      sock.onmessage = (e) => {
        let msg: Msg;
        try {
          msg = JSON.parse(e.data);
        } catch {
          return; // ignore malformed frames rather than throwing
        }
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
        setConnected(false);
        if (closed) return;
        // Exponential backoff with jitter, capped at 30s, so a downed gateway
        // isn't hammered with a reconnect every 3s.
        const delay = Math.min(30000, 1000 * 2 ** attempt) + Math.random() * 1000;
        attempt++;
        timer = setTimeout(connect, delay);
      };
    };
    connect();

    return () => {
      closed = true;
      clearTimeout(timer);
      sock?.close();
    };
  }, [qc, push]);
}
