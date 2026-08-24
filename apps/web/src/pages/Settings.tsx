import { useState } from "react";
import { Check } from "lucide-react";
import { useUI } from "../lib/store";
import { Panel } from "../components/Panel";
import { Button } from "../components/Button";
import { Field, Input } from "../components/Field";

export function Settings() {
  const { theme, toggleTheme } = useUI();
  const [key, setKey] = useState(() => {
    try {
      return localStorage.getItem("sitemon.apiKey") ?? "";
    } catch {
      return "";
    }
  });
  const [saved, setSaved] = useState(false);

  const save = () => {
    try {
      if (key) localStorage.setItem("sitemon.apiKey", key);
      else localStorage.removeItem("sitemon.apiKey");
    } catch {
      /* ignore */
    }
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div className="mx-auto max-w-2xl px-6 py-6">
      <h1 className="mb-5 font-display text-xl font-bold">Settings</h1>
      <div className="flex flex-col gap-4">
        <Panel title="API key">
          <div className="flex items-end gap-2">
            <div className="flex-1">
              <Field label="Sent as X-API-Key to the gateway">
                <Input type="password" value={key} onChange={(e) => setKey(e.target.value)} placeholder="unset (open gateway)" />
              </Field>
            </div>
            <Button onClick={save}>{saved ? <><Check size={15} /> Saved</> : "Save"}</Button>
          </div>
          <p className="mt-2 text-xs text-muted">Stored in this browser only. Applies to API calls and the live stream.</p>
        </Panel>

        <Panel title="Appearance">
          <div className="flex items-center justify-between">
            <span className="text-sm">Theme</span>
            <Button variant="ghost" onClick={toggleTheme}>{theme === "dark" ? "Dark" : "Light"}</Button>
          </div>
        </Panel>
      </div>
    </div>
  );
}
