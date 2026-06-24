import { describe, expect, it, vi } from "vitest";

import {
  createSandbox,
  fetchBackendHealth,
  inspectSandbox,
  listSandboxes,
} from "./api";

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

describe("sandbox api", () => {
  it("creates a sandbox through the configured API base URL", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        id: "sandbox-1",
        name: "Scratch",
        provider: "noop",
        state: "ready",
        default_workdir: "/workspace",
      }),
    });

    const sandbox = await createSandbox(
      "http://127.0.0.1:8080",
      { name: "Scratch", agentos_image: "agentos:test" },
      fetchMock,
    );

    expect(fetchMock).toHaveBeenCalledWith("http://127.0.0.1:8080/sandboxes", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name: "Scratch", agentos_image: "agentos:test" }),
    });
    expect(sandbox.state).toBe("ready");
  });

  it("lists sandboxes", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => [
        { id: "sandbox-1", name: "Scratch", provider: "noop", state: "ready" },
      ],
    });

    const sandboxes = await listSandboxes("http://127.0.0.1:8080", fetchMock);

    expect(fetchMock).toHaveBeenCalledWith("http://127.0.0.1:8080/sandboxes", {
      headers: { Accept: "application/json" },
    });
    expect(sandboxes).toHaveLength(1);
  });

  it("inspects a sandbox", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        id: "sandbox-1",
        name: "Scratch",
        provider: "local-docker",
        state: "ready",
      }),
    });

    const sandbox = await inspectSandbox(
      "http://127.0.0.1:8080",
      "sandbox-1",
      fetchMock,
    );

    expect(fetchMock).toHaveBeenCalledWith(
      "http://127.0.0.1:8080/sandboxes/sandbox-1/inspect",
      {
        method: "POST",
        headers: { Accept: "application/json" },
      },
    );
    expect(sandbox.provider).toBe("local-docker");
  });
});
