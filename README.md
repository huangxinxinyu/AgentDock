# AgentDock

AgentDock is a Go-first control plane for managed coding-agent runs in remote
sandboxes. Sprint 1 establishes the project skeleton: Go backend, React web
console, local Postgres dependency, configuration conventions, provider
boundary placeholders, and CI verification.

## Local Setup

```sh
cp .env.example .env.local
make install
make deps-up
make backend
```

In another terminal:

```sh
make frontend
```

Backend health is available at:

```sh
curl http://127.0.0.1:8080/healthz
```

## Verification

```sh
make ci
```

This runs Go formatting/tests/build, frontend format/lint/typecheck/tests/build,
and migration layout checks.

Postgres-backed integration tests are opt-in so CI does not require a local
database:

```sh
AGENTDOCK_TEST_DATABASE_URL="postgres://agentdock:agentdock@localhost:5432/agentdock?sslmode=disable" \
  make go-test
```

Sprint 2 backend API resources:

```text
POST /workspaces
GET  /workspaces/{id}
POST /workspaces/{workspace_id}/repositories
POST /workspaces/{workspace_id}/agents
POST /workspaces/{workspace_id}/issues
GET  /issues/{id}
POST /issues/{id}/runs
GET  /runs/{id}
GET  /runs/{id}/events
```

`POST /issues/{id}/runs` accepts `Idempotency-Key`; other mutation endpoints
will need full idempotency coverage in a later sprint.

## Secrets

Do not commit real secrets. `.env.local` is ignored. The current local backend
requires Postgres but does not require Daytona, OpenRouter, Anthropic, Claude,
or GitHub provider keys; those variables are documented in `.env.example` for
future smoke-test work only.
