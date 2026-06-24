# Sprint 4 Local Docker AgentOS Lifecycle Design

Date: 2026-06-24
Status: proposed for implementation after user review

## Summary

Sprint 4 will make Sandbox Lab lifecycle operations real through a local Docker provider that starts an externally supplied AgentOS image.

AgentDock does not build or define the sandbox image. AgentDock starts AgentOS, tracks lifecycle state, and exposes controls. AgentOS owns the runtime contents inside the sandbox.

```text
Sandbox Lab create
-> local Docker provider starts AgentOS image
-> AgentDock records provider container id
-> user can pause, resume, close, and inspect state
```

Daytona remains the target remote provider, but is not implemented in this sprint.

## Goals

- Implement a `local-docker` sandbox provider behind the Sprint 3 lifecycle interface.
- Start a configured AgentOS image as the sandbox runtime.
- Persist the Docker container id as `provider_session_id`.
- Support create, pause, resume, close, and inspect.
- Preserve sandbox records and work data after close unless a future cleanup policy says otherwise.
- Surface provider configuration errors clearly in the API and UI.
- Add opt-in Docker integration tests.

## Non-Goals

- Adding a Dockerfile to the AgentDock repository.
- Building AgentOS from AgentDock.
- Installing Python, Claude Agent SDK, or toolchain dependencies from the Go backend.
- Calling Daytona.
- Running agent tasks.
- Implementing terminal, file browser, or task log streaming.
- Automatic TTL cleanup or deletion of Docker volumes.
- Binding sandboxes to repositories, issues, or runs.

## Boundary Between AgentDock and AgentOS

AgentDock owns:

- sandbox records and lifecycle state;
- lifecycle API and UI controls;
- provider abstraction;
- provider configuration validation;
- status and error reporting.

AgentOS owns:

- Python and Claude Agent SDK installation;
- runtime tools;
- runner entrypoints;
- filesystem layout inside the sandbox;
- behavior of future task execution.

AgentDock may pass configuration into AgentOS, such as sandbox id, default workdir, callback URL, and safe environment variables. It should not bake AgentOS implementation details into the Go service.

## Configuration

The Docker provider should be disabled unless configured.

Required configuration:

```text
AGENTDOCK_SANDBOX_PROVIDER=local-docker
AGENTDOCK_AGENTOS_IMAGE=<image reference>
```

Optional configuration may include:

```text
AGENTDOCK_AGENTOS_DEFAULT_WORKDIR=/workspace
AGENTDOCK_DOCKER_NETWORK=<network name>
AGENTDOCK_DOCKER_VOLUME_PREFIX=agentdock
```

If `local-docker` is selected but no AgentOS image is configured, sandbox creation should fail with a provider-not-configured error. The UI should show the error from `last_error` or the API response.

## Provider Contract

The provider interface should support:

```text
CreateSession(ctx, request) -> provider session id, default workdir, metadata
PauseSession(ctx, sandbox)
ResumeSession(ctx, sandbox)
CloseSession(ctx, sandbox)
InspectSession(ctx, sandbox) -> observed state and metadata
```

Implementation details can choose Docker CLI or Docker Engine API. The service boundary should keep that decision inside the provider package.

## Lifecycle Semantics

State transitions remain those from Sprint 3:

```text
creating
ready
paused
closing
closed
failed
```

Expected behavior:

- create starts the configured AgentOS image and moves the sandbox to ready when the provider observes a running container;
- pause moves ready to paused;
- resume moves paused to ready;
- close stops the container and marks the sandbox closed;
- inspect reconciles database state with Docker-observed state when possible;
- provider failures record `last_error` and move the sandbox to failed when the requested operation cannot complete.

Idempotent actions:

- pause on paused returns paused;
- resume on ready returns ready;
- close on closed returns closed.

The exact Docker primitive for pause/resume can be selected during implementation. The product semantics matter more than whether the provider uses Docker pause/unpause or stop/start.

## Data Model Changes

Sprint 4 should reuse Sprint 3's `sandbox_sessions` table.

Expected field usage:

- `provider = 'local-docker'`
- `provider_session_id = <container id>`
- `agentos_image = <configured image reference>`
- `default_workdir = <AgentOS workdir, default /workspace>`
- `metadata` may include safe Docker observations such as image digest, container name, network name, and volume name.
- `last_started_at`, `last_paused_at`, and `closed_at` are updated by lifecycle actions.

No new task tables are required in Sprint 4.

## API and UI Changes

The Sprint 3 endpoints remain the user-facing lifecycle API:

```text
POST /sandboxes
GET  /sandboxes
GET  /sandboxes/{id}
POST /sandboxes/{id}/pause
POST /sandboxes/{id}/resume
POST /sandboxes/{id}/close
```

Sprint 4 adds inspect/refresh:

```text
POST /sandboxes/{id}/inspect
```

The UI should show:

- provider configuration errors;
- Docker container id;
- AgentOS image;
- observed status after inspect;
- last lifecycle timestamps;
- last error.

## Acceptance Criteria

- With `AGENTDOCK_SANDBOX_PROVIDER=local-docker` and `AGENTDOCK_AGENTOS_IMAGE` set, a developer can create a sandbox backed by a Docker container.
- The created container uses the configured AgentOS image.
- The sandbox record stores the Docker container id in `provider_session_id`.
- Pause, resume, close, and inspect work against the Docker-backed sandbox.
- Repeated pause, resume, and close actions are safe.
- Provider errors are persisted in `last_error` and exposed through the API/UI.
- Missing AgentOS image configuration produces a clear provider-not-configured error.
- The AgentDock repository does not add or require an AgentOS Dockerfile.
- Work data is not automatically deleted by close.
- Normal CI does not require Docker.
- Docker integration tests can be run explicitly with an environment flag.

## Verification

Always run:

```sh
make ci
```

Opt-in local Docker verification:

```sh
AGENTDOCK_DOCKER_TESTS=1 make go-test
```

The opt-in tests should be skipped by default when Docker or `AGENTDOCK_AGENTOS_IMAGE` is unavailable.

## Deferred Complexity

- Daytona provider.
- AgentOS task execution.
- Container log streaming.
- Terminal access.
- File browser.
- Automatic cleanup or garbage collection.
- Credentials and secret injection.
- Workspace, repository, issue, and run attachment.
- Production sandbox network policy.
