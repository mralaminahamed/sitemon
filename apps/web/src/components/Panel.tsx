import type { ReactNode } from "react";

export function Panel({ title, action, children, className = "" }: {
  title?: string;
  action?: ReactNode;
  children: ReactNode;
  className?: string;
}) {
  return (
    <section className={`rounded-xl border border-border bg-panel ${className}`}>
      {(title || action) && (
        <header className="flex items-center justify-between border-b border-border px-4 py-3">
          {title && (
            <h2 className="font-mono text-[11px] font-semibold uppercase tracking-[0.14em] text-muted">
              {title}
            </h2>
          )}
          {action}
        </header>
      )}
      <div className="p-4">{children}</div>
    </section>
  );
}
