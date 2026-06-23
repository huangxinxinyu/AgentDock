import { useEffect, useMemo, useState } from "react";

import { type BackendHealth, fetchBackendHealth } from "./api";
import "./styles.css";

type AppProps = {
  apiBaseUrl?: string;
  fetcher?: typeof fetch;
};

const navItems = [
  "Workspaces",
  "Issues",
  "Runs",
  "Agents",
  "Settings",
] as const;

const boundaryItems = [
  "Sandbox provider",
  "Agent runtime",
  "Repository integration",
  "Checks",
  "Eval",
] as const;

export function App({ apiBaseUrl, fetcher }: AppProps) {
  const resolvedApiBaseUrl = useMemo(
    () =>
      apiBaseUrl ??
      import.meta.env.VITE_API_BASE_URL ??
      "http://127.0.0.1:8080",
    [apiBaseUrl],
  );
  const [health, setHealth] = useState<BackendHealth>({
    service: "agentdock-api",
    status: "degraded",
    message: "checking",
  });

  useEffect(() => {
    let active = true;

    void fetchBackendHealth(resolvedApiBaseUrl, fetcher).then((nextHealth) => {
      if (active) {
        setHealth(nextHealth);
      }
    });

    return () => {
      active = false;
    };
  }, [fetcher, resolvedApiBaseUrl]);

  const isOnline = health.status === "ok";

  return (
    <main className="app-shell">
      <aside className="sidebar" aria-label="Primary navigation">
        <div className="brand-lockup">
          <div className="brand-mark" aria-hidden="true">
            AD
          </div>
          <div>
            <h1>AgentDock</h1>
            <p>Control plane</p>
          </div>
        </div>
        <nav>
          {navItems.map((item) => (
            <a href={`#${item.toLowerCase()}`} key={item}>
              {item}
            </a>
          ))}
        </nav>
      </aside>

      <section className="workspace">
        <header className="topbar">
          <div>
            <p className="eyebrow">Sprint 1 skeleton</p>
            <h2>Work Management Shell</h2>
          </div>
          <div
            className={`health-pill ${isOnline ? "ok" : "degraded"}`}
            role="status"
          >
            <span aria-hidden="true" />
            {isOnline ? "Backend online" : "Backend unavailable"}
          </div>
        </header>

        <section className="status-grid" aria-label="Skeleton status">
          <article>
            <p className="section-label">Backend</p>
            <h3>Remote Run Engine</h3>
            <dl>
              <div>
                <dt>Service</dt>
                <dd>{health.service}</dd>
              </div>
              <div>
                <dt>Health</dt>
                <dd>{health.status}</dd>
              </div>
            </dl>
          </article>

          <article>
            <p className="section-label">Scaffolding</p>
            <h3>Provider boundaries</h3>
            <ul>
              {boundaryItems.map((item) => (
                <li key={item}>{item}</li>
              ))}
            </ul>
          </article>
        </section>

        <section className="route-band" aria-label="Route placeholders">
          {navItems.slice(0, 4).map((item) => (
            <div key={item}>
              <span>{item}</span>
              <strong>Placeholder route</strong>
            </div>
          ))}
        </section>
      </section>
    </main>
  );
}
