# Sprint 5 AgentOS Runner Smoke Loop Design

Date: 2026-06-24
Status: proposed for implementation after user review

## Summary

Sprint 5 will prove that AgentDock can ask AgentOS to run a minimal Python Claude Agent SDK task inside a ready sandbox and observe the result.

This is still a Sandbox Lab workflow. It does not connect tasks to issues, subagents, runs, GitHub, or patch review.

```text
ready sandbox
-> user submits prompt
-> AgentDock creates sandbox task
-> provider asks AgentOS to run task
-> AgentOS runs Python Claude Agent SDK
-> AgentDock records task events, final state, summary, and output reference
```

## Goals

- Add a Sandbox Lab task model.
- Start a task in a ready sandbox through an AgentOS runner contract.
- Record task lifecycle state.
- Record append-only task events.
- Show task status, logs/events, summary, and output reference in the UI.
- Prove a small Python Claude Agent SDK task can create files in the sandbox workdir through AgentOS.

## Non-Goals

- Binding tasks to issues, subagents, or runs.
- Creating patch versions, diffs, pull requests, or branch commits.
- GitHub integration.
- Daytona provider.
- Terminal, file browser, or interactive shell.
- Full artifact storage and download APIs.
- Long-running AgentOS service architecture unless implementation proves it is required.
- Seamless task continuation after sandbox pause.

## AgentOS Runner Boundary

AgentDock should not know Claude Agent SDK internals. It should call a stable AgentOS runner entrypoint and observe results.

The exact command or API can be finalized during implementation, but the contract must support:

```text
run task id
read prompt from a file or structured request
execute in workdir
emit logs/events or expose logs for collection
write outputs under workdir
return final status and summary
```

Illustrative command shape:

```text
agentos run --task-id <id> --prompt-file <path> --workdir <path>
```

This command shape is not a committed CLI. It exists to define the minimum data contract between AgentDock and AgentOS.

## API Shape

```text
POST /sandboxes/{id}/tasks
GET  /sandboxes/{id}/tasks
GET  /sandbox-tasks/{id}
GET  /sandbox-tasks/{id}/events
POST /sandbox-tasks/{id}/cancel
```

Task creation accepts:

- `prompt`;
- optional workdir override;
- optional entrypoint key, defaulting to the Python Claude Agent SDK runner exposed by AgentOS.

The default workdir comes from `sandbox_sessions.default_workdir`.

## Data Model

```sql
sandbox_tasks
  id uuid primary key
  sandbox_session_id uuid not null references sandbox_sessions(id)
  prompt text not null
  state text not null
  entrypoint text not null
  workdir text not null
  summary text not null default ''
  output_ref text not null default ''
  last_error text not null default ''
  created_at timestamptz not null
  updated_at timestamptz not null
  started_at timestamptz
  completed_at timestamptz
```

Task state values:

```text
queued
starting
running
succeeded
failed
cancelled
```

Constraints and indexes:

```sql
CHECK (state IN ('queued', 'starting', 'running', 'succeeded', 'failed', 'cancelled'))

sandbox_tasks_sandbox_session_id_created_at_idx
sandbox_tasks_state_updated_at_idx
```

Task events:

```sql
sandbox_task_events
  id uuid primary key
  sandbox_task_id uuid not null references sandbox_tasks(id)
  sequence integer not null
  type text not null
  message text not null default ''
  payload jsonb not null default '{}'
  created_at timestamptz not null
```

Constraints and indexes:

```sql
UNIQUE (sandbox_task_id, sequence)

sandbox_task_events_task_sequence_idx
```

Initial event types:

```text
task_queued
task_starting
task_running
log
output
task_succeeded
task_failed
task_cancelled
```

`output_ref` may point to a path or manifest produced by AgentOS. Sprint 5 should not add a full artifact table.

## Backend Design

Add a task service that:

- validates the sandbox exists and is ready before task start;
- creates a task in `queued`;
- moves task through `starting`, `running`, and a terminal state;
- appends task events with monotonic sequence numbers;
- invokes the provider/AgentOS runner contract;
- records final summary, output reference, and last error.

The first implementation may run tasks synchronously or with a small bounded worker, but it must not start unowned goroutines. Any asynchronous worker must have a context and shutdown path.

Cancellation should be represented in the API and state machine even if the first AgentOS contract can only best-effort stop the process.

## UI Design

Extend Sandbox Lab with a task panel for ready sandboxes.

Required UI elements:

- prompt input;
- start task action;
- task list for the selected sandbox;
- task detail/status;
- event/log list;
- summary;
- output reference;
- cancel action when task is cancellable.

The UI should not claim full file browsing or diff review. It should show where AgentOS says output was written.

## Pause and Close During Tasks

Sprint 5 should use conservative lifecycle rules:

- a running task blocks close, or close requires cancellation first;
- pausing a sandbox with a running task may leave the task interrupted or unknown;
- after resume, the UI may allow inspection and rerun, but must not promise transparent task continuation.

These rules can be refined once AgentOS task behavior is stable.

## Acceptance Criteria

- A developer can create a task against a ready sandbox through the API.
- Task creation is rejected for closed, paused, failed, or missing sandboxes.
- A task records state transitions from queued to terminal state.
- Task events are append-only and ordered per task.
- AgentDock invokes the AgentOS runner contract without embedding Claude Agent SDK logic in the Go backend.
- A smoke task can ask AgentOS to create a small Python project or file in the sandbox workdir.
- The UI shows task status, events/logs, summary, and output reference.
- Task failure records `last_error` and a failure event.
- Cancel is represented in API, state, and UI even if provider support is best effort.
- Normal CI uses fake provider/AgentOS tests and does not require Docker or Claude credentials.
- Opt-in integration tests can exercise Docker plus AgentOS when configured.

## Verification

Always run:

```sh
make ci
```

Opt-in integration verification:

```sh
AGENTDOCK_DOCKER_TESTS=1 AGENTDOCK_AGENTOS_IMAGE=<image> make go-test
```

Normal test coverage should include:

- task state transition tests;
- task event sequence tests;
- API validation tests;
- fake AgentOS success and failure tests;
- frontend task creation and display tests.

## Deferred Complexity

- Issue/subagent/run integration.
- GitHub repository checkout.
- Patch review and diff generation.
- Artifact table and file download API.
- Realtime streaming over SSE or WebSocket.
- Durable task queue infrastructure.
- Daytona execution.
- Secret management for model credentials.
- Transparent resume of interrupted tasks.
