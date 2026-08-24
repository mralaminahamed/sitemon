import { useToasts } from "../lib/toast";

const toneColor: Record<string, string> = {
  up: "var(--color-up)",
  down: "var(--color-down)",
  warn: "var(--color-warn)",
  accent: "var(--color-accent)",
};

export function Toaster() {
  const { toasts, dismiss } = useToasts();
  return (
    <div className="pointer-events-none fixed bottom-4 right-4 z-50 flex w-80 flex-col gap-2">
      {toasts.map((t) => (
        <button
          key={t.id}
          onClick={() => dismiss(t.id)}
          className="pointer-events-auto rounded-lg border border-border bg-panel2 p-3 text-left shadow-lg"
          style={{ borderLeftColor: toneColor[t.tone], borderLeftWidth: 3 }}
        >
          <div className="text-sm font-medium text-text">{t.title}</div>
          {t.detail && <div className="mt-0.5 font-mono text-xs text-muted">{t.detail}</div>}
        </button>
      ))}
    </div>
  );
}
