# Use NATS JetStream for Background Work

Status: superseded by ADR-0003

AgentDock will use NATS JetStream as the platform message queue for background work such as run dispatch, cancellation hints, sandbox cleanup, artifact processing, and later eval fanout. NATS is Go-native, lightweight compared with Kafka or Temporal, and provides durable streams, acknowledgements, redelivery, and consumer groups without making workflow infrastructure the center of the MVP.

**Considered Options**

- NATS JetStream.
- Redis Streams.
- RabbitMQ.
- Temporal.
- Kafka.

**Consequences**

Postgres remains the source of truth for runs, run events, sandbox sessions, and artifacts. NATS messages are delivery hints and work items; consumers must be idempotent and reload authoritative state from Postgres before mutating anything.
