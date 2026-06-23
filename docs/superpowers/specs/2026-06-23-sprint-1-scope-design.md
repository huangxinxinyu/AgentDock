# Sprint 1 Scope Design

Date: 2026-06-23
Status: proposed for implementation after user review

## Summary

Sprint 1 will build AgentDock's first end-to-end vertical slice:

```text
create issue
-> create run
-> worker drives sandbox provider
-> append run events
-> stream live trace
-> persist summary and patch placeholder
-> show result in the web console
```

The sprint should prove AgentDock's core thesis: a Go control plane owns the run
lifecycle while execution happens behind a sandbox provider boundary. The first
implementation may use a fake sandbox provider for local deterministic
verification, with a Daytona provider boundary in place for the first real
remote smoke test.

## Goals

- Create the initial application skeleton for a greenfield repo.
- Implement the minimum Work Management Shell needed to start and inspect one
  coding-agent run.
- Implement the Remote Run Engine's first durable lifecycle path.
- Persist authoritative state in Postgres.
- Use Redis only as an operational coordination boundary where needed.
- Store run trace as append-only durable events.
- Stream run events to the browser with run-scoped SSE.
- Represent patch review and eval as lightweight MVP records without building
  the full review or eval lab experience.

## Non-Goals

- GitHub App installation, OAuth, webhooks, pull request creation, or merge
  automation.
- Applying patches back to GitHub branches.
- Full patch review workflow beyond displaying the latest produced patch
  placeholder.
- Sandbox reuse, Continue Work, or full sandbox TTL management.
- Multi-agent routing, multi-provider agent runtimes, squads, autopilots, or
  marketplace skills.
- Billing, organization RBAC, mobile app, self-host installer, or collaboration
  features.
- External eval SDK integrations, LLM judges, golden test suites, or Eval Lab.
- Copying implementation code from `multica-upstream/`.

## Product Surface

Sprint 1 should expose a thin operational shell:

- Workspace seed or create flow sufficient for local development.
- Repository record with a hardcoded or manually entered repo URL.
- Agent record for one Claude-based coding agent profile.
- Issue list and issue detail.
- Run detail with state, live trace, summary, patch placeholder, checks, and
  lightweight eval result.

This keeps the UI focused on operating and observing remote runs rather than
becoming a broad issue tracker.

## Architecture

### Backend

The backend should be a Go-first service with:

- HTTP API for workspaces, repositories, agents, issues, runs, and run events.
- Embedded worker loop in the same deployable binary for Sprint 1.
- Postgres migrations and repository layer for durable state.
- Narrow sandbox provider interface owned by AgentDock domain language.
- Fake sandbox provider for deterministic local tests.
- Daytona provider package behind the same interface, allowed to be incomplete
  until credentials and remote smoke testing are available.
- Structured logging around request ID, workspace ID, issue ID, run ID, and
  sandbox session ID where applicable.

The embedded worker is an MVP deployment choice, not a permanent architectural
lock-in. Package boundaries should still allow a separate worker process later.

### Frontend

The frontend should be a TypeScript/React web console with:

- Issue list/detail.
- Create issue action.
- Start run action.
- Run detail panel.
- Live trace viewer backed by SSE.
- Summary, patch placeholder, check result, and eval result panels.

The UI should be utilitarian and operational: dense enough to inspect run state,
with no marketing landing page.

### Sandbox Provider Boundary

Core services should speak AgentDock terms:

- create sandbox session
- prepare repository
- start command
- stream provider output into normalized run events
- collect summary
- export patch placeholder or patch content
- cleanup or mark cleanup pending

Only provider packages should speak Daytona-specific terms. Fake and Daytona
providers must satisfy the same consumer-facing interface.

## State Model

Sprint 1 should use coarse run states with rich events:

```text
queued
provisioning
preparing_workspace
running
awaiting_review
completed
failed
cancelled
```

`completed` is the successful terminal execution state for Sprint 1. This
resolves the existing document inconsistency where `SCOPE.md` uses `succeeded`
and ADR-0004 uses `completed`.

Run state and patch state remain separate. Sprint 1 patch states:

```text
pending
superseded
applied
rejected
conflict
```

Only `pending` needs to be reachable in the first vertical slice. The remaining
states should exist in the model if that is cheap, but their full workflows are
not in Sprint 1.

## Data Scope

Initial durable tables should cover:

- `workspaces`
- `repositories`
- `agents`
- `issues`
- `runs`
- `run_events`
- `sandbox_sessions`
- `patch_versions`
- `checks`
- `eval_results`

Important invariants:

- Postgres is the source of truth.
- Run events are append-only and ordered per run.
- Run transitions are validated in code.
- Worker claims and retries must be safe to repeat.
- Sandbox provider side effects must be represented by durable run state and
  events before the UI depends on them.

## API Scope

Sprint 1 API should include:

- Create/list/get workspace as needed for the local shell.
- Create/list/get repository as needed for a hardcoded or manually entered repo.
- Create/list/get agent for one Claude agent profile.
- Create/list/get issue.
- Create run for an issue.
- Get run detail.
- List run events with cursor or sequence offset.
- Subscribe to run events with SSE.

Mutation APIs that create issues or runs should use idempotency keys or a clear
retry-safe strategy.

## Runtime Behavior

The first run path should be:

1. User creates an issue.
2. User starts a run for the issue.
3. API creates a durable `queued` run.
4. Worker claims the run.
5. Worker transitions through `provisioning`, `preparing_workspace`, and
   `running`.
6. Provider emits normalized events.
7. Worker appends events to `run_events`.
8. UI receives events over SSE and can reload from REST if SSE disconnects.
9. Worker persists summary, patch placeholder, check result, and eval result.
10. Worker transitions to `awaiting_review` when there is a reviewable patch
    placeholder, or `completed` when there is only a summary result.
11. Failure paths transition to `failed` with structured error events.

For Sprint 1, a fake provider can simulate the full path. A real Daytona smoke
test is optional unless credentials and remote access are available.

## Eval Scope

Eval is included as a lightweight MVP record, not an eval platform.

The first evaluator should persist simple deterministic fields such as:

- run reached a terminal state
- run produced summary
- run produced patch placeholder or patch content
- checks passed, failed, or were not run
- optional user-visible verdict label

External eval infrastructure and LLM judging are out of scope.

## Testing And Verification

Sprint 1 implementation should include:

- Unit tests for run state transitions.
- Unit tests for append-only event sequencing.
- Unit tests for idempotent run creation or retry-safe run creation behavior.
- Worker tests with fake sandbox provider covering success and failure.
- API tests for create issue, create run, get run, and list events.
- Frontend smoke test for issue creation, run start, trace display, and final
  result display.
- Migration verification.

If Daytona credentials are available, add one manual or scripted smoke test that
creates a sandbox, runs a simple command, streams output, and records events.

## Acceptance Criteria

Sprint 1 is complete when:

- A developer can start the stack locally.
- A user can create an issue from the web console.
- A user can start a run for that issue.
- The run moves through durable states.
- The worker appends ordered run events.
- The browser shows live trace updates.
- The run stores a summary and patch placeholder.
- The issue/run view shows the final state, summary, patch placeholder, check
  result, and lightweight eval result.
- Tests cover the core state machine, worker, API, and UI smoke path.
- The implementation does not copy code from `multica-upstream/`.

## Follow-Up After Sprint 1

Likely Sprint 2 candidates:

- Real Daytona provider hardening.
- GitHub App connection and issue branch creation.
- Patch export and apply guard.
- Sandbox TTL and Continue Work.
- First real Claude SDK adapter inside the sandbox.
- More useful eval scoring and persisted feedback.
