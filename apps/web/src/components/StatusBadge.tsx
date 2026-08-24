import type { Status } from "../types";

const colors: Record<Status, string> = {
  UP: "var(--ok)",
  WARNING: "var(--warn)",
  REDIRECT: "var(--info)",
  DOWN: "var(--bad)",
};

export function StatusBadge({ status }: { status: Status }) {
  return (
    <span className="badge" style={{ background: colors[status] }}>
      {status}
    </span>
  );
}
