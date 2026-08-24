export type Status = "UP" | "WARNING" | "REDIRECT" | "DOWN";

export interface HealthResult {
  url: string;
  status: Status;
  status_code: number;
  response_time_ms: number;
  timestamp: string;
  error?: string;
}

export interface Stats {
  url: string;
  total_checks: number;
  successful: number;
  failed: number;
  uptime_percentage: number;
  avg_latency_ms: number;
}

export interface Monitor {
  url: string;
  created_at: string;
}

export interface CertificateInfo {
  url: string;
  host: string;
  issuer: string;
  subject: string;
  valid_from: string;
  valid_until: string;
  days_remaining: number;
  is_valid: boolean;
  protocol: string;
  cipher_suite: string;
}

export interface RequestStats {
  total_requests: number;
  successful: number;
  failed: number;
  duration_ms: number;
  requests_per_sec: number;
  success_rate: number;
  avg_latency_ms: number;
  min_latency_ms: number;
  max_latency_ms: number;
  p50_latency_ms: number;
  p90_latency_ms: number;
  p95_latency_ms: number;
  p99_latency_ms: number;
  status_codes: Record<string, number>;
}

export interface Analysis {
  url: string;
  window: number;
  uptime_pct: number;
  error_rate: number;
  avg_ms: number;
  p95_ms: number;
  current_status: string;
  anomalies: string[];
  severity: "ok" | "warn" | "critical";
  summary?: string;
}
