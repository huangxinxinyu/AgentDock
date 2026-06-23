# AgentDock Backend Engineering Standards

`AGENTS.md` defines the agent workflow. This file defines the backend code
quality bar for AgentDock: how services, APIs, storage, background work,
consistency, caching, observability, and extensibility should be designed.

AgentDock is a Go-first control plane for remote coding-agent runs. The backend
must be reliable under retries, concurrent workers, partial failures, duplicated
events, stale caches, and sandbox/provider outages.

## Core Principles

- Postgres is the source of truth for durable product state.
- Redis is an operational coordination layer for queues, locks, rate limits,
  fanout, and cache. It must not own durable state.
- Every state transition must be explicit, validated, and observable.
- Every cross-process call must have a timeout, cancellation path, retry policy,
  and structured error handling.
- Every background job, event consumer, and retryable API path must be
  idempotent.
- The normal failure model is at-least-once delivery, duplicated work, delayed
  messages, and out-of-order observations.
- Prefer boring, inspectable designs over clever distributed protocols.

## Domain Boundaries

Keep product concepts distinct:

- `Issue` describes user-visible work.
- `Run` describes one execution attempt.
- `Run Event` is the append-only execution trace.
- `Sandbox Session` is the remote environment attached to an active issue.
- `Patch Version` is a concrete diff produced by a run.
- `Check` is a deterministic command result.
- `Eval` is a product-level assessment of work quality.

Do not collapse these concepts for implementation convenience. Mixing lifecycle
state across these boundaries makes retries, review, cleanup, and evals harder
to reason about.

Provider-specific terms must stay behind provider packages. Core services should
speak AgentDock language: run, sandbox session, repository preparation, command,
event, patch, artifact, cleanup.

## API Contracts

Design APIs around stable resources and standard operations:

- Use resource-oriented paths and nouns for durable objects.
- Keep API resources stable even if database tables change.
- Prefer standard CRUD semantics where they fit.
- Use explicit custom actions only when the operation is not naturally CRUD,
  such as `cancel`, `continue`, `apply_patch`, or `retry`.
- Use consistent JSON field names and avoid leaking internal table names.
- Use cursor pagination for list endpoints that can grow.
- Make filters and sort fields explicit allowlists.
- Return stable error shapes with machine-readable codes and human-readable
  messages.
- Include request IDs or trace IDs in responses and logs.
- Preserve backwards compatibility once an API shape is used by the UI or CLI.

Mutation APIs must be retry-safe:

- Prefer caller-supplied resource IDs or idempotency keys for create operations.
- If using server-generated IDs, accept an idempotency key for retried `POST`
  requests.
- Treat repeated requests with the same idempotency key and same parameters as
  the same operation.
- Reject repeated requests with the same idempotency key and different
  parameters.
- Store enough response metadata to return a semantically equivalent response on
  retry.
- Define the idempotency retention window explicitly.

Asynchronous APIs should return a resource the client can poll or subscribe to:

- Long-running run creation returns a run ID.
- Patch application returns a patch state or apply operation state.
- Cancellation returns the accepted target state and records a durable event.

## Idempotency

Idempotency is required for:

- Run creation and dispatch.
- Sandbox provisioning and cleanup.
- Repository preparation.
- Agent command launch.
- Patch export and patch version creation.
- Patch application.
- Check and eval recording.
- Webhook handling.
- Queue and event consumers.
- Retryable external API calls.

Use one or more of these mechanisms:

- Unique database constraints for natural idempotency.
- Idempotency-key tables for external requests.
- Event IDs and processed-event tables for consumers.
- Compare-and-swap updates for state transitions.
- Leases with fencing tokens for workers.
- `INSERT ... ON CONFLICT` when duplicate creation is expected.

Side effects and idempotency records must be committed atomically whenever they
share a correctness boundary. If they cannot be committed atomically, use an
outbox or make the downstream side effect independently idempotent.

Never implement idempotency only in memory.

## State Machines

State machines are correctness boundaries.

- Define allowed `from -> to` transitions in code.
- Reject invalid transitions instead of silently ignoring them.
- Record who or what caused each transition.
- Persist timestamps for important lifecycle points.
- Keep coarse product states stable; put detailed progress in append-only
  events.
- Separate execution lifecycle from review lifecycle.

For workers:

- Claim work with a lease and fencing token.
- Renew leases explicitly.
- Treat expired leases as ambiguous, not proof that no side effect happened.
- On recovery, reload authoritative state from Postgres before acting.
- Make cleanup safe to repeat.

## Concurrency

Every goroutine must have an owner and a shutdown path.

- Pass `context.Context` through request, worker, database, Redis, provider, and
  external API calls.
- Do not start fire-and-forget goroutines.
- Bound worker pools, queues, channel buffers, and fanout.
- Use backpressure or load shedding instead of unbounded memory growth.
- Protect shared mutable state with explicit synchronization or avoid sharing.
- Do not hold locks while making network calls.
- Keep lock ordering stable to avoid deadlocks.
- Prefer database constraints and transactional updates over distributed locks
  for durable correctness.

Use Redis locks only for coordination, not as the only guard for durable state.
Any Redis lock that protects durable state needs a Postgres-side validation or
fencing mechanism.

## Timeouts, Retries, and Backoff

All cross-process calls need timeouts:

- HTTP clients.
- GitHub API calls.
- Daytona/provider calls.
- Agent command supervision.
- Redis calls.
- Database calls where the caller has a bounded lifecycle.

Retries must be deliberate:

- Retry only operations that are idempotent or read-only.
- Retry at one layer in a call stack, not every layer.
- Use capped exponential backoff with jitter.
- Bound total retry duration by the caller's deadline.
- Use retry budgets or token buckets for high-volume paths.
- Do not retry validation errors, authorization errors, or deterministic
  conflicts.
- Log final failure with the number of attempts and last error class.

Fallbacks are risky. A fallback path must have the same quality bar as the
primary path and must not hide data loss, stale state, or permission failures.

## Database Design

Schema design starts from access patterns:

- Document expected read/write paths before adding tables.
- Estimate growth for high-volume tables.
- Use UUID or stable opaque IDs for externally visible identifiers.
- Use foreign keys for durable relationships unless a clear scaling reason says
  otherwise.
- Use unique constraints for invariants.
- Use check constraints for local validity rules.
- Store timestamps for lifecycle and audit needs.
- Avoid storing static enumerations in mutable database tables unless they are
  user-configurable data.

Indexes must match queries:

- Add indexes for common `WHERE`, `JOIN`, `ORDER BY`, and pagination paths.
- Prefer composite indexes that match real access patterns.
- Avoid redundant indexes.
- Validate non-trivial queries with `EXPLAIN` or `EXPLAIN ANALYZE` once realistic
  data exists.
- Never ship an unbounded table scan on a path expected to grow.

Migrations must be production-minded:

- Make schema migrations reversible where possible.
- Split risky changes into expand/migrate/contract phases.
- Backfill large tables in batches.
- Avoid long transactions and table-wide locks.
- Add indexes on large existing tables concurrently when supported.
- Keep destructive changes separate and documented.
- Include rollback or recovery notes for data migrations.

## Transactions and Locking

Transactions should be small and purposeful:

- Put all writes that define one invariant in the same transaction.
- Do not include slow network calls inside database transactions.
- Use row-level locks only when they protect a named invariant.
- Prefer optimistic concurrency for user-facing updates where conflict feedback
  is acceptable.
- Use serializable isolation only for workflows that need it; otherwise encode
  invariants with constraints and compare-and-swap updates.
- Treat deadlocks and serialization failures as retryable only when the
  transaction is idempotent.

State transition updates should include the expected current state:

```sql
UPDATE runs
SET state = 'running'
WHERE id = $1 AND state = 'preparing_workspace';
```

The affected row count is part of the correctness check.

## Events and Consistency

Use append-only events for traceability:

- Run events are durable and ordered per run.
- Event sequence numbers must be monotonic inside the run.
- Event payloads should be structured JSON with stable event types.
- Store enough metadata to replay or debug a run without the live sandbox.

Use the transactional outbox pattern when a database write must cause an
external message or side effect:

- Write business state and outbox event in the same Postgres transaction.
- Relay outbox records asynchronously.
- Make the relay safe to retry.
- Make consumers idempotent by event ID.
- Preserve per-aggregate ordering when ordering matters.

Do not claim exactly-once delivery across process boundaries. Design for
at-least-once delivery with duplicate detection and idempotent effects.

## Cache Consistency

Cache is an optimization, not the source of truth.

- Every cached value must be reconstructable from Postgres or an external source
  of truth.
- Define cache key shape, TTL, invalidation trigger, and allowed staleness.
- Include version, updated timestamp, or ETag-style metadata where stale writes
  are dangerous.
- Do not let older cache fills overwrite newer invalidations.
- Prefer cache-aside for simple read-heavy paths.
- Use write-through or explicit invalidation only when the mutation path can be
  tested and observed.
- Do not use Redis to mask missing database constraints.
- Provide a bypass or rebuild path for operational recovery.

Critical cache paths need observability:

- Hit/miss rate.
- Stale or version-mismatch count.
- Invalidation lag.
- Rebuild failures.
- Redis latency and error rate.

## High Availability and Overload

Availability comes from bounded work and graceful degradation.

- Bound request body size, page size, queue depth, concurrent runs, and provider
  calls.
- Use admission control for run dispatch.
- Separate user-facing request latency from long-running worker execution.
- Prefer quick `202 Accepted` plus durable run state for slow operations.
- Make cancellation best-effort but durable and visible.
- Do not let one noisy workspace starve others.
- Use per-workspace or per-repository quotas where needed.
- Make cleanup and reconciliation periodic, idempotent, and safe after crashes.

Avoid cascading failure:

- Do not fan out unboundedly from user requests.
- Do not synchronously call optional dependencies on critical paths.
- Do not retry aggressively when a dependency is overloaded.
- Record dependency health and fail fast when continuing would only amplify
  load.

## Observability

Every important operation needs correlation:

- Request ID.
- Workspace ID.
- Repository ID when applicable.
- Issue ID when applicable.
- Run ID when applicable.
- Sandbox session ID when applicable.
- Patch version ID when applicable.
- External provider request ID when available.

Use structured logs for:

- State transitions.
- Worker claims and lease renewals.
- Retry exhaustion.
- Idempotency hits and mismatches.
- Provider operations.
- Patch apply conflicts.
- Webhook receipt and deduplication.
- Cache invalidation and rebuild failures.

Metrics should cover:

- Request rate, latency, and error rate.
- Queue depth and worker lag.
- Run state durations.
- Sandbox provision, prepare, command, export, and cleanup duration.
- Retry counts by dependency.
- Lease expiration and recovery counts.
- Database query latency for hot paths.
- Redis latency and error rate.
- Idempotency key reuse and mismatch counts.
- Cache hit/miss/stale rates.

Traces should cross API, worker, provider, database, Redis, and event relay
boundaries when available.

## Security and Data Safety

- Treat repository contents, prompts, traces, patches, logs, and artifacts as
  user data.
- Never log secrets, access tokens, environment variables, or full authorization
  headers.
- Redact provider credentials and sandbox secrets before persistence.
- Keep GitHub writes in the backend, not inside sandbox processes, unless an ADR
  explicitly changes this boundary.
- Enforce authorization before repository, issue, run, patch, artifact, and
  sandbox access.
- Prefer deny-by-default permission checks.
- Make destructive actions explicit, audited, and reversible where possible.
- Store enough audit context to answer who initiated a run, who applied a patch,
  and which credentials/provider were used.

## Extensibility

Extensibility must be paid for by a known second implementation or a clear
boundary in the current design.

- Keep provider interfaces narrow and product-oriented.
- Do not leak Daytona, GitHub, Redis, or agent-runtime vendor concepts into core
  domain types.
- Keep sandbox provider, repository integration, agent runtime, evaluator, and
  notification boundaries separate.
- Prefer composition over global registries.
- Avoid configuration formats that require code changes for routine provider
  additions.
- Do not add generic plugin machinery before the first stable concrete
  integration proves the boundary.

## Go Code Quality

- Follow idiomatic Go and `gofmt`.
- Prefer small packages with clear ownership.
- Keep interfaces at consumer boundaries.
- Accept interfaces, return concrete types when practical.
- Do not pass pointers to interfaces.
- Copy maps and slices at ownership boundaries when mutation would be unsafe.
- Avoid mutable global state.
- Handle each error once: either wrap and return it, translate it, or log it at
  the boundary.
- Do not panic for expected runtime errors.
- Use typed errors or error predicates for control flow.
- Use `time.Time`, `time.Duration`, and monotonic time-aware APIs correctly.
- Make zero values useful where reasonable.
- Tests should cover invariants, state transitions, idempotency, and failure
  paths, not just happy paths.

## Backend Change Checklist

For every backend change, answer the relevant questions before committing:

- What durable invariant does this change introduce or rely on?
- Is the operation safe to retry?
- What prevents duplicate side effects?
- What happens if the worker crashes after the database write but before the
  external side effect?
- What happens if the external side effect succeeds but the process times out?
- Which database constraints enforce correctness?
- Which query paths need indexes or query-plan review?
- Does this migration lock, rewrite, or scan a table that can grow large?
- What is the maximum concurrency and how is it bounded?
- Can one workspace, repository, or run starve others?
- What is the allowed consistency window?
- If cache is wrong, how is it detected and rebuilt?
- Are events safe under duplicate, delayed, or out-of-order delivery?
- What metrics, logs, or traces will prove the system is healthy?
- Does this change preserve the product language in `CONTEXT.md`?

## Reference Baseline

These standards are informed by production backend guidance from Google SRE,
AWS Builders Library, Stripe idempotency practices, Google AIPs, Microsoft API
Guidelines, GitLab database guidelines, PostgreSQL documentation, Meta cache
consistency work, and the transactional outbox pattern.
