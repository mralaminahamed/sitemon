export type Status = "UP" | "WARNING" | "REDIRECT" | "DOWN";

export interface HealthResult {
  url: string;
  status: Status;
  status_code: number;
  response_time_ms: number; // milliseconds
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

export interface Results {
  results: HealthResult[];
}
