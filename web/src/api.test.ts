import { describe, expect, it, vi } from "vitest";

import { fetchBackendHealth } from "./api";

describe("fetchBackendHealth", () => {
  it("loads backend health from the configured API base URL", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ service: "agentdock-api", status: "ok" }),
    });

    const health = await fetchBackendHealth("http://127.0.0.1:8080", fetchMock);

    expect(fetchMock).toHaveBeenCalledWith("http://127.0.0.1:8080/healthz", {
      headers: { Accept: "application/json" },
    });
    expect(health).toEqual({ service: "agentdock-api", status: "ok" });
  });

  it("returns a degraded health result when the backend cannot be reached", async () => {
    const fetchMock = vi.fn().mockRejectedValue(new Error("network down"));

    const health = await fetchBackendHealth("http://127.0.0.1:8080", fetchMock);

    expect(health.status).toBe("degraded");
    expect(health.message).toContain("network down");
  });
});
