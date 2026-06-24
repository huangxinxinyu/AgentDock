# AgentDock Scope

## Working Title

AgentDock: a Go-first, remote-sandbox managed coding agents platform inspired by Multica.

## Reference Project

The upstream reference code is checked out at:

- `multica-upstream/`

Multica's useful product model:

- agents as assignable teammates
- issue board and task lifecycle
- reusable skills
- runtime abstraction
- live execution progress
- workspace isolation

AgentDock should learn from that model, but it should not be a direct fork for implementation. The main architectural difference is that AgentDock runs coding agents inside remote sandboxes instead of a local daemon on the user's machine.

## Core Thesis

Build a managed agents platform where a user can assign a coding task to an agent, and the Go control plane provisions a Daytona sandbox, prepares the repository, runs a Claude-based agent, streams trace events, and returns a patch or task summary.

## Primary Difference From Multica

Multica:

```text
Web app -> Go backend -> local agent daemon -> user's machine executes agent CLI
```

AgentDock:

```text
Web app -> Go control plane -> Daytona sandbox -> remote environment executes agent
```

This makes sandboxing, traceability, reproducibility, and evals first-class parts of the system.

## MVP User Flow

1. User creates a workspace.
2. User connects a GitHub repository or supplies a repo URL.
3. User creates an issue/task.
4. User assigns the task to a configured Claude coding agent.
5. Go control plane creates a Daytona sandbox.
6. The sandbox clones or receives the repository.
7. The runtime starts the agent with the task prompt, skills, and allowed tools.
8. The backend streams run events to the web console.
9. The agent produces a summary and, when code changed, a patch.
10. The issue is marked succeeded, failed, cancelled, or awaiting review.

## MVP Components

### Go Control Plane

Responsibilities:

- workspace, agent, issue, run, and event APIs
- run lifecycle state machine
- Daytona sandbox orchestration
- event streaming over SSE or WebSocket
- policy checks before dangerous operations
- audit log and trace persistence
- cancellation, timeout, retry, and cleanup

Initial run states:

```text
queued
provisioning
preparing_workspace
running
awaiting_review
succeeded
failed
cancelled
```

### Daytona Sandbox Provider

First and only sandbox provider for MVP.

Responsibilities:

- create sandbox
- clone or upload repo
- install runtime dependencies
- inject agent configuration and secrets
- run agent process
- stream stdout, stderr, tool events, and file change metadata
- snapshot or export patch
- destroy or retain sandbox based on run policy

The provider interface should remain abstract enough to add E2B or local Docker later.

### Local Docker Development Provider

Before Daytona is integrated, Sandbox Lab sprints may use a local Docker provider
to validate sandbox lifecycle and AgentOS execution on a developer machine. This
is a development substrate, not a change to the managed remote-sandbox product
direction.

AgentDock should not maintain the runtime image contents in this repository.
The local Docker provider starts an externally supplied AgentOS image or package.
AgentOS owns Python, Claude Agent SDK dependencies, runner entrypoints, toolchain
setup, and filesystem conventions inside the sandbox.

### Agent Runtime

MVP should support one agent runtime:

- Claude Agent SDK or Claude Code-based execution inside the sandbox

Non-goals for MVP:

- Codex, Gemini, Copilot, Cursor Agent, Kimi, and other CLIs
- multi-agent routing
- squads/autopilots

### Web Console

MVP screens:

- workspace selector
- issue board
- issue detail
- agent settings
- run detail
- live trace viewer
- sandbox status

The web app should focus on operational clarity rather than a broad product surface.

### Skills

MVP skill format:

```text
.agentdock/
  skills/
    code-review/
      SKILL.md
      manifest.json
```

Skill responsibilities:

- provide task-specific instructions
- declare required tools
- declare sandbox setup steps if needed
- declare eval fixtures later

### Eval Lab

Eval is a phase-two feature, not required for the first clickable loop.

Planned responsibilities:

- fixed repo fixtures
- golden tasks
- test-pass scoring
- patch diff scoring
- LLM judge
- trace regression
- skill regression

Python is the preferred language for eval and analysis.

## Suggested Stack

### Required

- Go for backend control plane
- PostgreSQL for durable state
- Daytona for remote sandboxes
- TypeScript/React for web console
- Claude Agent SDK or Claude Code runtime inside sandbox

### Later

- Python eval runner
- OpenTelemetry traces
- Redis or NATS for queues/events
- object storage for artifacts
- GitHub App integration

## Data Model Draft

Core tables:

- `workspaces`
- `repositories`
- `agents`
- `issues`
- `runs`
- `run_events`
- `sandbox_sessions`
- `skills`
- `audit_logs`

Phase-two tables:

- `eval_suites`
- `eval_cases`
- `eval_runs`
- `eval_results`
- `memories`

## First Milestone

Build the smallest remote execution loop:

```text
Create issue
-> assign to agent
-> create Daytona sandbox
-> clone repo
-> run a simple agent command
-> stream logs
-> save summary
-> update issue status
```

This milestone can use a hardcoded repository and a hardcoded agent prompt. The important part is proving that Go can own the run lifecycle while execution happens remotely.

## Explicit Non-Goals For MVP

- local daemon compatibility
- hosted billing
- organization-level RBAC
- self-host installer
- mobile app
- multi-provider agent CLI compatibility
- marketplace for skills
- custom Firecracker implementation
- full GitHub PR automation
- complex memory system

## Engineering Questions To Resolve Next

- Which Daytona SDK/API path is best for Go integration?
- Should the Go backend call Daytona directly, or wrap Daytona behind a small internal service?
- Should live trace use SSE first or WebSocket first?
- Should the first web app be built from scratch or adapted around Multica's UI concepts?
- Should the agent runtime use Claude Agent SDK directly or shell out to Claude Code in the sandbox?
- What is the minimum repo fixture for first evals?

## License Note

The upstream Multica checkout is for study and comparison. AgentDock should keep its own implementation and architecture decisions separate unless we deliberately review license implications for copied code or assets.
