import type { InputHTMLAttributes, ReactNode } from "react";

export function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="flex flex-col gap-1.5">
      <span className="font-mono text-[11px] uppercase tracking-wide text-muted">{label}</span>
      {children}
    </label>
  );
}

export function Input(props: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      {...props}
      className={`rounded-lg border border-border bg-panel2 px-3 py-2 text-sm text-text outline-none placeholder:text-muted focus:border-accent ${props.className ?? ""}`}
    />
  );
}
