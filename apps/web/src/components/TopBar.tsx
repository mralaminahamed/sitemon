import { Moon, Search, Sun } from "lucide-react";
import { useStatus } from "../lib/hooks";
import { useUI } from "../lib/store";
import { useConn } from "../lib/conn";
import { isDown } from "../lib/format";
import { StatusDot } from "./StatusDot";

export function TopBar() {
  const { data } = useStatus();
  const { theme, toggleTheme, setCmdk } = useUI();
  const connected = useConn((s) => s.connected);
  const results = data?.results ?? [];
  const down = results.filter((r) => isDown(r.status)).length;
  const overall = results.length === 0 ? "UNKNOWN" : down > 0 ? "DOWN" : "UP";

  return (
    <header className="flex items-center justify-between border-b border-border px-6 py-3">
      <div className="flex items-center gap-2 text-sm">
        <StatusDot status={overall} live />
        <span className="text-muted">
          {results.length === 0
            ? "no monitors"
            : down > 0
              ? `${down} of ${results.length} down`
              : `all ${results.length} operational`}
        </span>
        {!connected && (
          <span className="font-mono text-[10px] uppercase tracking-wide text-warn">reconnecting…</span>
        )}
      </div>

      <div className="flex items-center gap-2">
        <button
          onClick={() => setCmdk(true)}
          className="flex items-center gap-2 rounded-lg border border-border bg-panel2 px-3 py-1.5 text-sm text-muted hover:border-accent"
        >
          <Search size={14} />
          <span>Search</span>
          <kbd className="rounded bg-panel px-1.5 py-0.5 font-mono text-[10px]">⌘K</kbd>
        </button>
        <button
          onClick={toggleTheme}
          className="rounded-lg border border-border bg-panel2 p-2 text-muted hover:border-accent hover:text-text"
          aria-label="Toggle theme"
        >
          {theme === "dark" ? <Sun size={16} /> : <Moon size={16} />}
        </button>
      </div>
    </header>
  );
}
