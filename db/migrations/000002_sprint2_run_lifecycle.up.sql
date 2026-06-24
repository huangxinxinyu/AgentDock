CREATE TABLE workspaces (
  id uuid PRIMARY KEY,
  name text NOT NULL CHECK (length(trim(name)) > 0),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE repositories (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  name text NOT NULL CHECK (length(trim(name)) > 0),
  url text NOT NULL CHECK (length(trim(url)) > 0),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX repositories_workspace_id_idx ON repositories(workspace_id);

CREATE TABLE agents (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  name text NOT NULL CHECK (length(trim(name)) > 0),
  runtime_key text NOT NULL CHECK (length(trim(runtime_key)) > 0),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX agents_workspace_id_idx ON agents(workspace_id);

CREATE TABLE issues (
  id uuid PRIMARY KEY,
  workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  repository_id uuid NOT NULL REFERENCES repositories(id) ON DELETE RESTRICT,
  agent_id uuid NOT NULL REFERENCES agents(id) ON DELETE RESTRICT,
  title text NOT NULL CHECK (length(trim(title)) > 0),
  prompt text NOT NULL CHECK (length(trim(prompt)) > 0),
  status text NOT NULL CHECK (status IN ('open')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX issues_workspace_id_idx ON issues(workspace_id);
CREATE INDEX issues_repository_id_idx ON issues(repository_id);
CREATE INDEX issues_agent_id_idx ON issues(agent_id);

CREATE TABLE runs (
  id uuid PRIMARY KEY,
  issue_id uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  repository_id uuid NOT NULL REFERENCES repositories(id) ON DELETE RESTRICT,
  agent_id uuid NOT NULL REFERENCES agents(id) ON DELETE RESTRICT,
  prompt text NOT NULL,
  state text NOT NULL CHECK (state IN ('queued', 'provisioning', 'preparing_workspace', 'running', 'completed', 'failed', 'cancelled')),
  result_summary text NOT NULL DEFAULT '',
  idempotency_key text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  started_at timestamptz,
  completed_at timestamptz,
  last_transition_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX runs_issue_id_idx ON runs(issue_id);
CREATE INDEX runs_state_created_at_idx ON runs(state, created_at);
CREATE UNIQUE INDEX runs_issue_idempotency_key_idx ON runs(issue_id, idempotency_key)
  WHERE idempotency_key IS NOT NULL;

CREATE TABLE run_events (
  id uuid PRIMARY KEY,
  run_id uuid NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  sequence integer NOT NULL CHECK (sequence > 0),
  type text NOT NULL CHECK (length(trim(type)) > 0),
  message text NOT NULL DEFAULT '',
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (run_id, sequence)
);

CREATE INDEX run_events_run_id_sequence_idx ON run_events(run_id, sequence);

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
