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
