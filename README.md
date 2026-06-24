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

## Secrets

Do not commit real secrets. `.env.local` is ignored. Sprint 1 does not require
Daytona, OpenRouter, Anthropic, Claude, or GitHub provider keys; those variables
are documented in `.env.example` for future smoke-test work only.
