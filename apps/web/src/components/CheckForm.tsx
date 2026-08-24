import { useState } from "react";
import { useCheck } from "../hooks";

export function CheckForm() {
  const [url, setUrl] = useState("");
  const check = useCheck();

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = url.trim();
    if (!trimmed) return;
    check.mutate([trimmed]);
    setUrl("");
  };

  return (
    <form className="checkform" onSubmit={submit}>
      <input
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        placeholder="example.com"
        aria-label="URL to check"
      />
      <button type="submit" disabled={check.isPending}>
        {check.isPending ? "Checking…" : "Check"}
      </button>
      {check.error && <span className="error">{(check.error as Error).message}</span>}
    </form>
  );
}
