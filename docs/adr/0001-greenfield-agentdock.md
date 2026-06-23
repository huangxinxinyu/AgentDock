# Build AgentDock Greenfield Instead of Forking Multica

AgentDock will be implemented as a greenfield system while keeping `multica-upstream/` only as a product and design reference. Multica's core execution path is built around a local daemon on the user's machine, while AgentDock's core path is a Go control plane provisioning remote Daytona sandboxes and running agents there; adapting the upstream daemon/runtime surface would carry more product and data-model baggage than it saves.

**Considered Options**

- Fork and aggressively trim Multica.
- Build greenfield and borrow only concepts such as agents, issues, skills, runtimes, and trace viewing.

**Consequences**

AgentDock can model `runs`, `run_events`, and `sandbox_sessions` directly instead of inheriting Multica's `agent_task_queue` shape. The cost is that web UI, API, and persistence must be rebuilt rather than copied.
