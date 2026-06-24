import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { App } from "./App";

describe("App", () => {
  it("renders the AgentDock operational shell and backend health", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ service: "agentdock-api", status: "ok" }),
    });
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ service: "agentdock-api", status: "ok" }),
    });
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => [],
    });

    render(<App apiBaseUrl="http://127.0.0.1:8080" fetcher={fetchMock} />);

    expect(
      screen.getByRole("heading", { name: "AgentDock" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Sandbox control surface")).toBeInTheDocument();
    expect(screen.getByText("Remote Run Engine")).toBeInTheDocument();
    expect(screen.getByText("Provider boundaries")).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText("Backend online")).toBeInTheDocument();
    });
  });

  it("shows Sandbox Lab sessions and creates a sandbox", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ service: "agentdock-api", status: "ok" }),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => [
          {
            id: "sandbox-1",
            name: "Existing",
            provider: "noop",
            state: "ready",
            default_workdir: "/workspace",
          },
        ],
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: "sandbox-2",
          name: "Scratch",
          provider: "noop",
          state: "ready",
          default_workdir: "/workspace",
        }),
      });

    render(<App apiBaseUrl="http://127.0.0.1:8080" fetcher={fetchMock} />);

    expect(await screen.findAllByText("Existing")).toHaveLength(2);
    fireEvent.change(screen.getByLabelText("Sandbox name"), {
      target: { value: "Scratch" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Create sandbox" }));

    expect(await screen.findAllByText("Scratch")).toHaveLength(2);
  });
});
