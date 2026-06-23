# Sprint 1 Scope Design

Date: 2026-06-23
Status: proposed for implementation after user review

## Summary

Sprint 1 will build AgentDock's project skeleton, not the first remote
execution loop.

The goal is to turn the documentation-only repository into a runnable,
testable, greenfield application foundation:

```text
repo layout
-> Go backend skeleton
-> TypeScript/React frontend skeleton
-> local Postgres/Redis dev environment
-> configuration and secret conventions
-> provider/runtime interface placeholders
-> CI and verification commands
```

This sprint should make the next implementation sprint straightforward without
silently committing product behavior too early.

## Goals

- Create the initial monorepo structure for AgentDock.
- Add a Go backend skeleton with clear package boundaries.
- Add a TypeScript/React web console skeleton.
- Add local development infrastructure for Postgres and Redis.
- Add configuration loading and `.env.example` conventions without committing
  real secrets.
- Add placeholder boundaries for sandbox providers, agent runtimes, repository
  integration, run orchestration, checks, and eval.
- Add basic health/status endpoints and frontend app shell only.
- Add formatting, linting, testing, and CI commands so future work has a stable
  engineering loop.

## Non-Goals

- Creating issues from the UI.
- Creating runs or executing a worker lifecycle.
- Provisioning Daytona sandboxes.
- Calling OpenRouter, Anthropic, Claude SDK, or Claude Code.
- Persisting the full product data model.
- Streaming run events over SSE or WebSocket.
- Implementing patch review, checks, or eval behavior.
- Implementing GitHub App installation, OAuth, webhooks, branch creation, pull
  requests, or patch apply.
- Implementing authentication, billing, RBAC, collaboration, marketplace skills,
  or multi-agent routing.
- Copying implementation code from `multica-upstream/`.

## Repository Shape

Sprint 1 should establish a structure close to:

```text
cmd/
  agentdock-api/
    main.go
internal/
  app/
  config/
  httpapi/
  domain/
  store/
  worker/
  sandbox/
  runtime/
  repo/
  checks/
  eval/
web/
  src/
  package.json
db/
  migrations/
deploy/
  docker-compose.yml
scripts/
docs/
```

Exact names can change during implementation if the local Go/React tooling makes
a better convention obvious, but the boundary intent should remain.

## Backend Skeleton

The Go backend should include:

- `cmd/agentdock-api` entrypoint.
- Configuration loading from environment variables.
- Structured logger initialization with secret redaction rules documented.
- HTTP server setup with request context, timeouts, request IDs, and graceful
  shutdown.
- `GET /healthz` for process health.
- `GET /readyz` for dependency readiness, allowed to report degraded database or
  Redis state during local development.
- Package boundaries for domain models, storage, API handlers, background
  workers, sandbox providers, agent runtimes, repository integrations, checks,
  and eval.
- Interfaces or placeholder types for future provider/runtime boundaries, with
  no real external calls.
- Unit tests that prove config loading, health handlers, and basic package wiring
  compile and run.

The backend should not implement the run state machine yet. It may define type
names or package locations that make the later run state machine natural.

## Frontend Skeleton

The frontend should include:

- TypeScript/React app setup.
- Routing shell for future workspace, issue list, issue detail, run detail, and
  settings pages.
- A first screen that shows AgentDock operational shell chrome and backend health
  status.
- API client foundation with typed response handling.
- Basic empty/loading/error states.
- Formatting, linting, typecheck, and test scripts.

The frontend should not implement issue creation, run creation, live trace,
patch review, or eval views beyond placeholder routes.

## Local Development

Sprint 1 should make local setup predictable:

- `deploy/docker-compose.yml` or equivalent for Postgres and Redis.
- `.env.example` documenting required local variables.
- `.env.local` ignored by git.
- No real Daytona, OpenRouter, Anthropic, GitHub, or encryption keys committed.
- Scripts or Make targets for:
  - installing dependencies
  - starting local dependencies
  - running backend
  - running frontend
  - running tests
  - running format/lint/typecheck

Local development should work without any external provider key. Provider keys
are introduced only in later smoke-test work.

## Configuration And Secret Boundaries

Sprint 1 should document and scaffold these secret categories:

- Platform infrastructure secrets: database URL, Redis URL, app/session secret,
  encryption key.
- Sandbox provider secrets: Daytona API key and related endpoint settings.
- Model provider secrets: user-supplied OpenRouter, Anthropic, or later runtime
  keys.
- GitHub integration secrets: future GitHub App credentials.

The skeleton should establish this principle: frontend never receives provider
secrets, committed files never contain real secrets, and backend logs must be
designed to redact credentials.

No separate AI middleware repository is part of Sprint 1. Agent/model routing
should remain a future backend boundary unless a later sprint proves that a
separate deployment is needed.

## Database And Migrations

Sprint 1 should add migration tooling and either:

- an empty first migration that validates the tooling, or
- minimal infrastructure tables required by the skeleton, such as schema version
  tracking if the chosen migration tool needs it.

It should not create the full AgentDock product schema. Future schema work must
still start from access patterns and invariants.

## CI And Verification

Sprint 1 should add a baseline verification loop:

- Go formatting and tests.
- Frontend formatting, linting, typechecking, and tests.
- Optional migration validation if the tooling supports it locally.
- CI workflow that runs the same commands without requiring external secrets.

The CI path must not require Daytona, OpenRouter, Anthropic, or GitHub keys.

## Acceptance Criteria

Sprint 1 is complete when:

- The repository has a clear backend/frontend/deploy/script structure.
- A developer can start Postgres and Redis locally.
- A developer can run the backend locally and hit `GET /healthz`.
- A developer can run the frontend locally and see the AgentDock app shell.
- The frontend can call the backend health endpoint.
- Backend tests pass.
- Frontend lint/typecheck/test commands pass.
- CI runs the baseline checks without external provider secrets.
- `.env.example` documents configuration without leaking real credentials.
- Provider/runtime/repository/check/eval package boundaries exist as scaffolding
  for later sprints.
- No implementation code is copied from `multica-upstream/`.

## Follow-Up After Sprint 1

Likely Sprint 2 candidates:

- Minimal domain schema and migrations for workspace, repository, agent, issue,
  run, and run event records.
- First fake-provider run lifecycle through the Go backend.
- Run-scoped SSE trace.
- Daytona smoke test path.
- OpenRouter or Claude runtime adapter behind the runtime interface.
