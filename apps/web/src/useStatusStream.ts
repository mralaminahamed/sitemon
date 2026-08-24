import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { HealthResult, Results } from "./types";

interface Msg {
  type: "snapshot" | "result";
  results?: HealthResult[];
  result?: HealthResult;
}

// Live status via WebSocket; TanStack Query polling remains as a fallback.
export function useStatusStream() {
  const qc = useQueryClient();
  useEffect(() => {
    const proto = location.protocol === "https:" ? "wss" : "ws";
    const sock = new WebSocket(`${proto}://${location.host}/ws`);

    sock.onmessage = (e) => {
      const msg: Msg = JSON.parse(e.data);
      if (msg.type === "snapshot") {
        qc.setQueryData<Results>(["status"], { results: msg.results ?? [] });
      } else if (msg.type === "result" && msg.result) {
        const r = msg.result;
        qc.setQueryData<Results>(["status"], (old) => {
          const rest = (old?.results ?? []).filter((x) => x.url !== r.url);
          return { results: [...rest, r] };
        });
      }
    };

    return () => sock.close();
  }, [qc]);
}
