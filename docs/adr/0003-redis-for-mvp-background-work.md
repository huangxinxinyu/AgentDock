# Use Redis for MVP Background Work

AgentDock will use Redis for MVP background work instead of introducing a dedicated message queue such as NATS JetStream. This keeps the initial infrastructure smaller while still supporting run dispatch, cancellation hints, cleanup jobs, lightweight locking, rate limits, and realtime fanout.

**Consequences**

Postgres remains the source of truth for runs, run events, sandbox sessions, artifacts, and eval metadata. Redis is an operational coordination layer; workers must treat Redis messages as delivery hints and reload authoritative state from Postgres before changing run state.
