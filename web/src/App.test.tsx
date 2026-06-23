import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { App } from "./App";

describe("App", () => {
  it("renders the AgentDock operational shell and backend health", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ service: "agentdock-api", status: "ok" }),
    });

    render(<App apiBaseUrl="http://127.0.0.1:8080" fetcher={fetchMock} />);

    expect(
      screen.getByRole("heading", { name: "AgentDock" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Work Management Shell")).toBeInTheDocument();
    expect(screen.getByText("Remote Run Engine")).toBeInTheDocument();
    expect(screen.getByText("Provider boundaries")).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText("Backend online")).toBeInTheDocument();
    });
  });
});
