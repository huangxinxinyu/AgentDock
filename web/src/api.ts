export type BackendHealth = {
  service: string;
  status: "ok" | "degraded";
  message?: string;
};

export type SandboxState =
  | "creating"
  | "ready"
  | "paused"
  | "closing"
  | "closed"
  | "failed";

export type SandboxSession = {
  id: string;
  name: string;
  provider: string;
  provider_session_id?: string;
  state: SandboxState;
  default_workdir?: string;
  agentos_image?: string;
  last_error?: string;
};

export type CreateSandboxInput = {
  name: string;
  provider?: string;
  default_workdir?: string;
  agentos_image?: string;
};

export type SandboxTaskState =
  | "queued"
  | "starting"
  | "running"
  | "succeeded"
  | "failed"
  | "cancelled";

export type SandboxTask = {
  id: string;
  sandbox_session_id: string;
  prompt?: string;
  state: SandboxTaskState;
  entrypoint?: string;
  workdir?: string;
  summary?: string;
  output_ref?: string;
  last_error?: string;
};

export type SandboxTaskEvent = {
  id?: string;
  sandbox_task_id: string;
  sequence: number;
  type: string;
  message?: string;
  payload?: string;
};

export type CreateSandboxTaskInput = {
  prompt: string;
  entrypoint?: string;
  workdir?: string;
};

type Fetcher = typeof fetch;

export async function fetchBackendHealth(
  apiBaseUrl: string,
  fetcher: Fetcher = fetch,
): Promise<BackendHealth> {
  const baseUrl = apiBaseUrl.replace(/\/+$/, "");

  try {
    const response = await fetcher(`${baseUrl}/healthz`, {
      headers: { Accept: "application/json" },
    });
    if (!response.ok) {
      return {
        service: "agentdock-api",
        status: "degraded",
        message: `health check returned ${response.status}`,
      };
    }

    const body = (await response.json()) as Partial<BackendHealth>;
    return {
      service: body.service ?? "agentdock-api",
      status: body.status === "ok" ? "ok" : "degraded",
      message: body.message,
    };
  } catch (error) {
    const message = error instanceof Error ? error.message : "unknown error";
    return {
      service: "agentdock-api",
      status: "degraded",
      message,
    };
  }
}

export async function listSandboxes(
  apiBaseUrl: string,
  fetcher: Fetcher = fetch,
): Promise<SandboxSession[]> {
  const baseUrl = apiBaseUrl.replace(/\/+$/, "");
  const response = await fetcher(`${baseUrl}/sandboxes`, {
    headers: { Accept: "application/json" },
  });
  if (!response.ok) {
    throw new Error(`list sandboxes returned ${response.status}`);
  }
  return (await response.json()) as SandboxSession[];
}

export async function createSandbox(
  apiBaseUrl: string,
  input: CreateSandboxInput,
  fetcher: Fetcher = fetch,
): Promise<SandboxSession> {
  const baseUrl = apiBaseUrl.replace(/\/+$/, "");
  const response = await fetcher(`${baseUrl}/sandboxes`, {
    method: "POST",
    headers: { Accept: "application/json", "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw new Error(`create sandbox returned ${response.status}`);
  }
  return (await response.json()) as SandboxSession;
}

export async function sandboxAction(
  apiBaseUrl: string,
  sandboxID: string,
  action: "pause" | "resume" | "close",
  fetcher: Fetcher = fetch,
): Promise<SandboxSession> {
  const baseUrl = apiBaseUrl.replace(/\/+$/, "");
  const response = await fetcher(
    `${baseUrl}/sandboxes/${sandboxID}/${action}`,
    {
      method: "POST",
      headers: { Accept: "application/json" },
    },
  );
  if (!response.ok) {
    throw new Error(`${action} sandbox returned ${response.status}`);
  }
  return (await response.json()) as SandboxSession;
}

export async function inspectSandbox(
  apiBaseUrl: string,
  sandboxID: string,
  fetcher: Fetcher = fetch,
): Promise<SandboxSession> {
  const baseUrl = apiBaseUrl.replace(/\/+$/, "");
  const response = await fetcher(`${baseUrl}/sandboxes/${sandboxID}/inspect`, {
    method: "POST",
    headers: { Accept: "application/json" },
  });
  if (!response.ok) {
    throw new Error(`inspect sandbox returned ${response.status}`);
  }
  return (await response.json()) as SandboxSession;
}

export async function createSandboxTask(
  apiBaseUrl: string,
  sandboxID: string,
  input: CreateSandboxTaskInput,
  fetcher: Fetcher = fetch,
): Promise<SandboxTask> {
  const baseUrl = apiBaseUrl.replace(/\/+$/, "");
  const response = await fetcher(`${baseUrl}/sandboxes/${sandboxID}/tasks`, {
    method: "POST",
    headers: { Accept: "application/json", "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw new Error(`create sandbox task returned ${response.status}`);
  }
  return (await response.json()) as SandboxTask;
}

export async function listSandboxTasks(
  apiBaseUrl: string,
  sandboxID: string,
  fetcher: Fetcher = fetch,
): Promise<SandboxTask[]> {
  const baseUrl = apiBaseUrl.replace(/\/+$/, "");
  const response = await fetcher(`${baseUrl}/sandboxes/${sandboxID}/tasks`, {
    headers: { Accept: "application/json" },
  });
  if (!response.ok) {
    throw new Error(`list sandbox tasks returned ${response.status}`);
  }
  return (await response.json()) as SandboxTask[];
}

export async function listSandboxTaskEvents(
  apiBaseUrl: string,
  taskID: string,
  fetcher: Fetcher = fetch,
): Promise<SandboxTaskEvent[]> {
  const baseUrl = apiBaseUrl.replace(/\/+$/, "");
  const response = await fetcher(`${baseUrl}/sandbox-tasks/${taskID}/events`, {
    headers: { Accept: "application/json" },
  });
  if (!response.ok) {
    throw new Error(`list sandbox task events returned ${response.status}`);
  }
  return (await response.json()) as SandboxTaskEvent[];
}

export async function cancelSandboxTask(
  apiBaseUrl: string,
  taskID: string,
  fetcher: Fetcher = fetch,
): Promise<SandboxTask> {
  const baseUrl = apiBaseUrl.replace(/\/+$/, "");
  const response = await fetcher(`${baseUrl}/sandbox-tasks/${taskID}/cancel`, {
    method: "POST",
    headers: { Accept: "application/json" },
  });
  if (!response.ok) {
    throw new Error(`cancel sandbox task returned ${response.status}`);
  }
  return (await response.json()) as SandboxTask;
}
