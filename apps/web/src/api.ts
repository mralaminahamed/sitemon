import type { HealthResult, Results, Stats } from "./types";

const base = "/api";

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(base + path, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
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
