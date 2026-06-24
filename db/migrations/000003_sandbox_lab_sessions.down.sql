DROP TABLE sandbox_sessions;

CREATE TABLE sandbox_sessions (
  id uuid PRIMARY KEY,
  issue_id uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  run_id uuid NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  provider text NOT NULL CHECK (length(trim(provider)) > 0),
  provider_session_id text NOT NULL CHECK (length(trim(provider_session_id)) > 0),
  state text NOT NULL CHECK (state IN ('active', 'closed')),
  created_at timestamptz NOT NULL DEFAULT now(),
  closed_at timestamptz
);

CREATE INDEX sandbox_sessions_issue_id_idx ON sandbox_sessions(issue_id);
CREATE INDEX sandbox_sessions_run_id_idx ON sandbox_sessions(run_id);
