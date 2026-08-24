import {
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import { useHistory } from "../hooks";

export function LatencyChart({ url }: { url: string }) {
  const { data } = useHistory(url);
  const points = (data?.results ?? [])
    .map((r) => ({
      t: new Date(r.timestamp).toLocaleTimeString(),
      ms: r.response_time_ms,
    }))
    .reverse();

  if (points.length === 0) return <p className="muted">No history yet.</p>;

  return (
    <ResponsiveContainer width="100%" height={220}>
      <LineChart data={points} margin={{ top: 8, right: 16, bottom: 0, left: -8 }}>
        <XAxis dataKey="t" tick={{ fontSize: 11 }} minTickGap={32} />
        <YAxis tick={{ fontSize: 11 }} unit="ms" width={48} />
        <Tooltip />
        <Line
          type="monotone"
          dataKey="ms"
          stroke="var(--accent)"
          strokeWidth={2}
          dot={false}
          isAnimationActive={false}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}
