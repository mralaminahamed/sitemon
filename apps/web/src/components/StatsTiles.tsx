import { useStats } from "../hooks";

export function StatsTiles({ url }: { url: string }) {
  const { data } = useStats(url);
  if (!data) return null;

  const tiles = [
    { label: "Checks", value: data.total_checks },
    { label: "Uptime", value: `${data.uptime_percentage.toFixed(1)}%` },
    { label: "Failed", value: data.failed },
    { label: "Avg latency", value: `${data.avg_latency_ms.toFixed(0)} ms` },
  ];

  return (
    <div className="tiles">
      {tiles.map((t) => (
        <div key={t.label} className="tile">
          <div className="tile__value">{t.value}</div>
          <div className="tile__label">{t.label}</div>
        </div>
      ))}
    </div>
  );
}
