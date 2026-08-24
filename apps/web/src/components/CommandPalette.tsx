import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUI } from "../lib/store";
import { useStatus } from "../lib/hooks";

export function CommandPalette() {
  const { cmdkOpen, setCmdk } = useUI();
  const { data } = useStatus();
  const nav = useNavigate();
  const [q, setQ] = useState("");

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setCmdk(!cmdkOpen);
      }
      if (e.key === "Escape") setCmdk(false);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [cmdkOpen, setCmdk]);

  if (!cmdkOpen) return null;

  const pages = [
    { label: "Dashboard", to: "/" },
    { label: "Load Test", to: "/loadtest" },
    { label: "SSL Inspector", to: "/ssl" },
    { label: "Settings", to: "/settings" },
  ];
  const monitors = (data?.results ?? []).map((r) => ({
    label: r.url.replace(/^https?:\/\//, ""),
    to: `/monitors/${encodeURIComponent(r.url)}`,
  }));
  const all = [...pages, ...monitors].filter((i) =>
    i.label.toLowerCase().includes(q.toLowerCase()),
  );

  const go = (to: string) => {
    setCmdk(false);
    setQ("");
    nav(to);
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center bg-black/50 pt-[15vh]"
      onClick={() => setCmdk(false)}
    >
      <div
        className="w-full max-w-lg overflow-hidden rounded-xl border border-border bg-panel shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <input
          autoFocus
          value={q}
          onChange={(e) => setQ(e.target.value)}
          placeholder="Jump to a page or monitor…"
          className="w-full border-b border-border bg-transparent px-4 py-3 text-sm outline-none placeholder:text-muted"
        />
        <div className="max-h-72 overflow-y-auto p-2">
          {all.length === 0 && <div className="px-3 py-6 text-center text-sm text-muted">No matches</div>}
          {all.map((i) => (
            <button
              key={i.to}
              onClick={() => go(i.to)}
              className="block w-full rounded-lg px-3 py-2 text-left text-sm text-text hover:bg-panel2"
            >
              {i.label}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
