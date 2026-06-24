CREATE TABLE sandbox_tasks (
  id uuid PRIMARY KEY,
  sandbox_session_id uuid NOT NULL REFERENCES sandbox_sessions(id) ON DELETE CASCADE,
  prompt text NOT NULL CHECK (length(trim(prompt)) > 0),
  state text NOT NULL CHECK (state IN ('queued', 'starting', 'running', 'succeeded', 'failed', 'cancelled')),
  entrypoint text NOT NULL CHECK (length(trim(entrypoint)) > 0),
  workdir text NOT NULL CHECK (length(trim(workdir)) > 0),
  summary text NOT NULL DEFAULT '',
  output_ref text NOT NULL DEFAULT '',
  last_error text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  started_at timestamptz,
  completed_at timestamptz
);

CREATE INDEX sandbox_tasks_sandbox_session_id_created_at_idx ON sandbox_tasks(sandbox_session_id, created_at);
CREATE INDEX sandbox_tasks_state_updated_at_idx ON sandbox_tasks(state, updated_at);

CREATE TABLE sandbox_task_events (
  id uuid PRIMARY KEY,
  sandbox_task_id uuid NOT NULL REFERENCES sandbox_tasks(id) ON DELETE CASCADE,
  sequence integer NOT NULL CHECK (sequence > 0),
  type text NOT NULL CHECK (length(trim(type)) > 0),
  message text NOT NULL DEFAULT '',
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (sandbox_task_id, sequence)
);

CREATE INDEX sandbox_task_events_task_sequence_idx ON sandbox_task_events(sandbox_task_id, sequence);
