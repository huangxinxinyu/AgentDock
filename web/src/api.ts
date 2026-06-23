export type BackendHealth = {
  service: string;
  status: "ok" | "degraded";
  message?: string;
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
