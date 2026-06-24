DROP TABLE sandbox_sessions;

CREATE TABLE sandbox_sessions (
  id uuid PRIMARY KEY,
  name text NOT NULL CHECK (length(trim(name)) > 0),
  provider text NOT NULL CHECK (length(trim(provider)) > 0),
  provider_session_id text,
  state text NOT NULL CHECK (state IN ('creating', 'ready', 'paused', 'closing', 'closed', 'failed')),
  default_workdir text NOT NULL DEFAULT '/workspace' CHECK (length(trim(default_workdir)) > 0),
  agentos_image text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  last_error text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  last_started_at timestamptz,
  last_paused_at timestamptz,
  closed_at timestamptz
);

CREATE UNIQUE INDEX sandbox_sessions_provider_session_id_idx ON sandbox_sessions(provider, provider_session_id)
  WHERE provider_session_id IS NOT NULL;
CREATE INDEX sandbox_sessions_created_at_idx ON sandbox_sessions(created_at);
CREATE INDEX sandbox_sessions_state_updated_at_idx ON sandbox_sessions(state, updated_at);
