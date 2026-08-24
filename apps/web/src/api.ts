import type { HealthResult, Results, Stats } from "./types";

const base = "/api";
const apiKey = import.meta.env.VITE_API_KEY as string | undefined;

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (apiKey) headers["X-API-Key"] = apiKey;
  const res = await fetch(base + path, { headers, ...init });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `HTTP ${res.status}`);
  }
  return res.json();
}

export const api = {
  status: () => req<Results>("/status"),
  history: (url: string, limit = 50) =>
    req<Results>(`/history?url=${encodeURIComponent(url)}&limit=${limit}`),
  stats: (url: string) => req<Stats>(`/stats?url=${encodeURIComponent(url)}`),
  check: (urls: string[]) =>
    req<Results>("/checks", { method: "POST", body: JSON.stringify({ urls }) }),
};

export function nsToMs(ns: number): number {
  return Math.round(ns / 1e6);
}

export type { HealthResult, Stats };
