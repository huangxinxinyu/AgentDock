import { useEffect, useMemo, useState } from "react";

import {
  type BackendHealth,
  type SandboxSession,
  type SandboxTask,
  type SandboxTaskEvent,
  createSandbox,
  createSandboxTask,
  fetchBackendHealth,
  listSandboxTaskEvents,
  listSandboxes,
  sandboxAction,
} from "./api";
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
  const [sandboxes, setSandboxes] = useState<SandboxSession[]>([]);
  const [sandboxName, setSandboxName] = useState("");
  const [sandboxError, setSandboxError] = useState("");
  const [taskPrompt, setTaskPrompt] = useState("");
  const [tasks, setTasks] = useState<SandboxTask[]>([]);
  const [taskEvents, setTaskEvents] = useState<SandboxTaskEvent[]>([]);

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

  useEffect(() => {
    let active = true;

    void listSandboxes(resolvedApiBaseUrl, fetcher)
      .then((nextSandboxes) => {
        if (active) {
          setSandboxes(nextSandboxes);
          setSandboxError("");
        }
      })
      .catch((error: unknown) => {
        if (active) {
          setSandboxError(
            error instanceof Error ? error.message : "unknown error",
          );
        }
      });

    return () => {
      active = false;
    };
  }, [fetcher, resolvedApiBaseUrl]);

  const isOnline = health.status === "ok";
  const selectedSandbox = sandboxes[0];

  async function handleCreateSandbox(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const name = sandboxName.trim();
    if (name === "") {
      return;
    }
    try {
      const sandbox = await createSandbox(
        resolvedApiBaseUrl,
        { name, agentos_image: "" },
        fetcher,
      );
      setSandboxes((current) => [sandbox, ...current]);
      setSandboxName("");
      setSandboxError("");
    } catch (error) {
      setSandboxError(error instanceof Error ? error.message : "unknown error");
    }
  }

  async function handleSandboxAction(
    sandbox: SandboxSession,
    action: "pause" | "resume" | "close",
  ) {
    try {
      const updated = await sandboxAction(
        resolvedApiBaseUrl,
        sandbox.id,
        action,
        fetcher,
      );
      setSandboxes((current) =>
        current.map((candidate) =>
          candidate.id === updated.id ? updated : candidate,
        ),
      );
      setSandboxError("");
    } catch (error) {
      setSandboxError(error instanceof Error ? error.message : "unknown error");
    }
  }

  async function handleRunTask(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const prompt = taskPrompt.trim();
    if (!selectedSandbox || prompt === "") {
      return;
    }
    try {
      const task = await createSandboxTask(
        resolvedApiBaseUrl,
        selectedSandbox.id,
        { prompt },
        fetcher,
      );
      setTasks((current) => [task, ...current]);
      setTaskPrompt("");
      const events = await listSandboxTaskEvents(
        resolvedApiBaseUrl,
        task.id,
        fetcher,
      );
      setTaskEvents(events);
      setSandboxError("");
    } catch (error) {
      setSandboxError(error instanceof Error ? error.message : "unknown error");
    }
  }

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
            <p className="eyebrow">Sandbox Lab</p>
            <h2>Sandbox control surface</h2>
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

        <section className="sandbox-lab" aria-label="Sandbox Lab">
          <div className="sandbox-panel">
            <div>
              <p className="section-label">Sandbox Lab</p>
              <h3>Sessions</h3>
            </div>
            <form className="sandbox-form" onSubmit={handleCreateSandbox}>
              <label>
                Sandbox name
                <input
                  value={sandboxName}
                  onChange={(event) => setSandboxName(event.target.value)}
                  placeholder="Scratch"
                />
              </label>
              <button type="submit" disabled={sandboxName.trim() === ""}>
                Create sandbox
              </button>
            </form>
            {sandboxError && <p className="error-text">{sandboxError}</p>}
            <div className="sandbox-list">
              {sandboxes.length === 0 ? (
                <p className="empty-state">No sandboxes yet.</p>
              ) : (
                sandboxes.map((sandbox) => (
                  <button
                    className="sandbox-row"
                    key={sandbox.id}
                    type="button"
                    aria-label={`Select ${sandbox.name}`}
                  >
                    <span>
                      <strong>{sandbox.name}</strong>
                      <small>{sandbox.provider}</small>
                    </span>
                    <span className={`state-badge ${sandbox.state}`}>
                      {sandbox.state}
                    </span>
                  </button>
                ))
              )}
            </div>
          </div>

          <div className="sandbox-panel detail">
            <p className="section-label">Details</p>
            {selectedSandbox ? (
              <>
                <h3>{selectedSandbox.name}</h3>
                <dl>
                  <div>
                    <dt>Provider</dt>
                    <dd>{selectedSandbox.provider}</dd>
                  </div>
                  <div>
                    <dt>Session</dt>
                    <dd>{selectedSandbox.provider_session_id ?? "pending"}</dd>
                  </div>
                  <div>
                    <dt>Workdir</dt>
                    <dd>{selectedSandbox.default_workdir ?? "/workspace"}</dd>
                  </div>
                  <div>
                    <dt>AgentOS</dt>
                    <dd>{selectedSandbox.agentos_image || "not configured"}</dd>
                  </div>
                </dl>
                {selectedSandbox.last_error && (
                  <p className="error-text">{selectedSandbox.last_error}</p>
                )}
                <div className="sandbox-actions">
                  <button
                    type="button"
                    disabled={selectedSandbox.state !== "ready"}
                    onClick={() =>
                      void handleSandboxAction(selectedSandbox, "pause")
                    }
                  >
                    Pause
                  </button>
                  <button
                    type="button"
                    disabled={selectedSandbox.state !== "paused"}
                    onClick={() =>
                      void handleSandboxAction(selectedSandbox, "resume")
                    }
                  >
                    Resume
                  </button>
                  <button
                    type="button"
                    disabled={selectedSandbox.state === "closed"}
                    onClick={() =>
                      void handleSandboxAction(selectedSandbox, "close")
                    }
                  >
                    Close
                  </button>
                </div>
                <div className="task-panel">
                  <p className="section-label">AgentOS task</p>
                  <form className="task-form" onSubmit={handleRunTask}>
                    <label>
                      Task prompt
                      <textarea
                        value={taskPrompt}
                        onChange={(event) => setTaskPrompt(event.target.value)}
                        rows={3}
                        placeholder="Create a small Python project"
                      />
                    </label>
                    <button
                      type="submit"
                      disabled={
                        selectedSandbox.state !== "ready" ||
                        taskPrompt.trim() === ""
                      }
                    >
                      Run task
                    </button>
                  </form>
                  <div className="task-list">
                    {tasks.length === 0 ? (
                      <p className="empty-state">No tasks yet.</p>
                    ) : (
                      tasks.map((task) => (
                        <article className="task-row" key={task.id}>
                          <span className={`state-badge ${task.state}`}>
                            {task.state}
                          </span>
                          {task.summary && <strong>{task.summary}</strong>}
                          {task.output_ref && <small>{task.output_ref}</small>}
                        </article>
                      ))
                    )}
                  </div>
                  <div className="task-events">
                    {taskEvents.map((event) => (
                      <div key={`${event.sandbox_task_id}-${event.sequence}`}>
                        <span>{event.sequence}</span>
                        <strong>{event.type}</strong>
                        {event.message && <small>{event.message}</small>}
                      </div>
                    ))}
                  </div>
                </div>
              </>
            ) : (
              <p className="empty-state">Select a sandbox to inspect it.</p>
            )}
          </div>
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
