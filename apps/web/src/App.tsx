import { CheckForm } from "./components/CheckForm";
import { LatencyChart } from "./components/LatencyChart";
import { StatsTiles } from "./components/StatsTiles";
import { StatusGrid } from "./components/StatusGrid";
import { useStatusStream } from "./useStatusStream";
import { useUI } from "./store";

export function App() {
  const selectedUrl = useUI((s) => s.selectedUrl);
  useStatusStream();

  return (
    <div className="app">
      <header className="header">
        <h1>Sitemon</h1>
        <span className="muted">Site health · auto-refresh 5s</span>
      </header>

      <main>
        <section>
          <div className="section__head">
            <h2>Status</h2>
            <CheckForm />
          </div>
          <StatusGrid />
        </section>

        {selectedUrl && (
          <section>
            <h2 className="detail__title">{selectedUrl}</h2>
            <StatsTiles url={selectedUrl} />
            <LatencyChart url={selectedUrl} />
          </section>
        )}
      </main>
    </div>
  );
}
