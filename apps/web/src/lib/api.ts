import type {
  Analysis,
  CertificateInfo,
  HealthResult,
  Monitor,
  RequestStats,
  Stats,
} from "./types";

const base = "/api";

export function getApiKey(): string {
  try {
    return localStorage.getItem("sitemon.apiKey") || (import.meta.env.VITE_API_KEY as string) || "";
  } catch {
    return (import.meta.env.VITE_API_KEY as string) || "";
  }
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  const key = getApiKey();
  if (key) headers["X-API-Key"] = key;
  const res = await fetch(base + path, { headers, ...init });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `HTTP ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export interface LoadTestInput {
  url: string;
  method?: string;
  workers?: number;
  rps?: number;
  count?: number;
  duration_ms?: number;
}

export const api = {
  status: () => req<{ results: HealthResult[] }>("/status"),
  history: (url: string, limit = 60) =>
    req<{ results: HealthResult[] }>(`/history?url=${encodeURIComponent(url)}&limit=${limit}`),
  stats: (url: string) => req<Stats>(`/stats?url=${encodeURIComponent(url)}`),
  analyze: (url: string) => req<Analysis>(`/analyze?url=${encodeURIComponent(url)}`),
  ssl: (url: string) => req<CertificateInfo>(`/ssl?url=${encodeURIComponent(url)}`),
  monitors: () => req<{ monitors: Monitor[] }>("/monitors"),
  addMonitor: (url: string) =>
    req<Monitor>("/monitors", { method: "POST", body: JSON.stringify({ url }) }),
  deleteMonitor: (url: string) =>
    req<void>(`/monitors?url=${encodeURIComponent(url)}`, { method: "DELETE" }),
  check: (urls: string[]) =>
    req<{ results: HealthResult[] }>("/checks", { method: "POST", body: JSON.stringify({ urls }) }),
  loadtest: (input: LoadTestInput) =>
    req<RequestStats>("/loadtest", { method: "POST", body: JSON.stringify(input) }),
};
