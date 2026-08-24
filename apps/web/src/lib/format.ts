import type { Status } from "./types";

export const statusToken: Record<string, string> = {
  UP: "up",
  WARNING: "warn",
  REDIRECT: "redirect",
  DOWN: "down",
  UNKNOWN: "muted",
};

export function statusColor(s: string): string {
  return `var(--color-${statusToken[s] ?? "muted"})`;
}

export function ms(v: number): string {
  if (v >= 1000) return `${(v / 1000).toFixed(2)}s`;
  return `${Math.round(v)}ms`;
}

export function ago(iso: string): string {
  const d = Date.parse(iso);
  if (Number.isNaN(d)) return "—";
  const s = Math.max(0, Math.floor((Date.now() - d) / 1000));
  if (s < 60) return `${s}s ago`;
  if (s < 3600) return `${Math.floor(s / 60)}m ago`;
  if (s < 86400) return `${Math.floor(s / 3600)}h ago`;
  return `${Math.floor(s / 86400)}d ago`;
}

export function isDown(s: Status | string): boolean {
  return s === "DOWN" || s === "WARNING";
}
