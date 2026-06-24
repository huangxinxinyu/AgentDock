# Sprint 3 Sandbox Lab Foundation Design

Date: 2026-06-24
Status: proposed for implementation after user review

## Summary

Sprint 3 will build AgentDock's Sandbox Lab foundation instead of extending the issue/run workflow.

The goal is to make sandbox sessions independent, durable, and visible in the web console before any real Docker, Daytona, GitHub, or agent execution work is introduced:

```text
create sandbox record
-> track lifecycle state
-> expose Sandbox Lab APIs
-> show and control sandbox state in the UI
```

This sprint deliberately creates the product and engineering surface that Sprint 4's local Docker provider and Sprint 5's AgentOS runner will use.

## Goals

- Make `sandbox_sessions` an independent lifecycle resource.
- Remove the current assumption that a sandbox session must belong to a run.
- Add a global Sandbox Lab API.
- Add a Sandbox Lab UI for listing, creating, viewing, pausing, resuming, and closing sandbox sessions.
- Add a provider interface for sandbox lifecycle operations.
- Use a noop or deterministic stub provider so API and UI behavior can be tested without Docker or Daytona.
- Preserve the current issue/run APIs without expanding them.

## Non-Goals

- Creating real Docker containers.
- Calling Daytona.
- Starting AgentOS.
- Running Claude Agent SDK tasks.
- Binding sandboxes to workspaces, repositories, issues, or runs.
- Implementing GitHub integration, patch review, branch apply, file browsing, terminal access, or live task logs.
- Adding TTL cleanup or automatic retention policies.

## Product Decisions

Sandbox Lab is a temporary but real product surface for proving sandbox infrastructure. It is not the final issue/subagent assignment workflow.

For Sprint 3, sandbox sessions are global resources. `workspace_id`, `repo_id`, and `issue_id` are intentionally deferred. This avoids pulling product organization and repository decisions into the sandbox substrate before the lifecycle model is proven.

The future issue workflow should attach an issue to one current sandbox session, and many runs may later reference that same session. A sandbox session should not store a single `run_id`.

## API Shape

```text
POST /sandboxes
GET  /sandboxes
GET  /sandboxes/{id}
POST /sandboxes/{id}/pause
POST /sandboxes/{id}/resume
POST /sandboxes/{id}/close
```

`POST /sandboxes` accepts a name and optional safe provider settings. Provider secrets must not be accepted in request bodies or persisted in database fields.

The lifecycle actions should return the current sandbox representation. Repeated actions should be safe:

- pausing an already paused sandbox returns paused;
- resuming an already ready sandbox returns ready;
- closing an already closed sandbox returns closed.

## Data Model

Sprint 3 should replace or migrate the current run-bound `sandbox_sessions` shape to:

```sql
sandbox_sessions
  id uuid primary key
  name text not null
  provider text not null
  provider_session_id text
  state text not null
  default_workdir text not null default '/workspace'
  agentos_image text
  metadata jsonb not null default '{}'
  last_error text not null default ''
  created_at timestamptz not null
  updated_at timestamptz not null
  last_started_at timestamptz
  last_paused_at timestamptz
  closed_at timestamptz
```

State values:

```text
creating
ready
paused
closing
closed
failed
```

Constraints and indexes:

```sql
CHECK (state IN ('creating', 'ready', 'paused', 'closing', 'closed', 'failed'))

UNIQUE (provider, provider_session_id)
WHERE provider_session_id IS NOT NULL

sandbox_sessions_created_at_idx
sandbox_sessions_state_updated_at_idx
```

Field notes:

- `default_workdir` is the default directory inside AgentOS where work should happen. It is not a host path and is not necessarily a repository checkout.
- `agentos_image` records the requested AgentOS image or package id for provider startup. Sprint 3 may store it even though the noop provider does not use it.
- `metadata` is for provider-observed facts that are safe to persist, not secrets.
- `provider_session_id` may be empty before a real provider creates backing infrastructure.

Fields intentionally not included:

- `workspace_id`: deferred until the product workspace boundary returns.
- `repo_id`: sandboxes can initially work in an empty directory.
- `issue_id`: issue attachment is deferred.
- `run_id`: future runs should reference the sandbox session, not the reverse.
- TTL fields: lifecycle is manual for now.
- secret values: credentials belong in backend environment or a later secret provider.

## Backend Design

Add a sandbox service that owns validation, state transitions, and provider calls. Provider-specific identifiers and errors stay behind `internal/sandbox`.

The provider interface should be small enough for a noop implementation:

```text
CreateSession(ctx, request) -> session metadata
PauseSession(ctx, sandbox)
ResumeSession(ctx, sandbox)
CloseSession(ctx, sandbox)
InspectSession(ctx, sandbox) -> observed state and metadata
```

Sprint 3 may implement lifecycle operations synchronously because the provider is a stub. The service should still model `creating`, `closing`, and `failed` so Sprint 4 can use the same state machine when Docker calls fail.

## Frontend Design

Add a Sandbox Lab page to the web console.

Required UI elements:

- sandbox list;
- create sandbox form;
- sandbox detail area or page;
- state badge;
- provider and provider session id;
- default workdir;
- AgentOS image;
- last error;
- pause, resume, and close buttons with state-based disabling.

The page should feel like an operational control surface, not a marketing page. It should not include terminal, file browser, task prompt, or log streaming in Sprint 3.

## Acceptance Criteria

- A developer can create a sandbox through `POST /sandboxes` without Docker, Daytona, or AgentOS configured.
- A developer can list and retrieve sandbox sessions through the API.
- Sandbox lifecycle actions update durable state according to the allowed state transitions.
- Repeated pause, resume, and close actions are safe and return the current state.
- Invalid transitions return a stable error shape instead of silently mutating state.
- The web console includes a Sandbox Lab surface for create, list, detail, pause, resume, and close.
- The UI disables impossible actions based on sandbox state.
- `sandbox_sessions` no longer requires `issue_id` or `run_id`.
- No provider secrets are accepted or persisted.
- Existing issue/run tests continue to pass.
- No code or assets are copied from `multica-upstream/`.

## Verification

Run:

```sh
make ci
```

Expected coverage:

- domain state transition tests;
- store create/get/list/update tests;
- HTTP API tests;
- frontend API client tests;
- basic Sandbox Lab interaction tests;
- migration layout checks.

Docker, Daytona, and AgentOS integration tests are not required in Sprint 3.

## Deferred Complexity

- Workspace ownership.
- Repository checkout and repo binding.
- Issue attachment.
- Run-to-session foreign key.
- Realtime events.
- Sandbox TTL and cleanup jobs.
- Secret references and encrypted credentials.
- Daytona provider behavior.
- AgentOS task execution.
