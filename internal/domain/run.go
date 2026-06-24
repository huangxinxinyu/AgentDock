package domain

import (
	"fmt"
	"time"
)

type RunState string

const (
	RunStateQueued             RunState = "queued"
	RunStateProvisioning       RunState = "provisioning"
	RunStatePreparingWorkspace RunState = "preparing_workspace"
	RunStateRunning            RunState = "running"
	RunStateCompleted          RunState = "completed"
	RunStateFailed             RunState = "failed"
	RunStateCancelled          RunState = "cancelled"
)

func IsRunState(value string) bool {
	switch RunState(value) {
	case RunStateQueued,
		RunStateProvisioning,
		RunStatePreparingWorkspace,
		RunStateRunning,
		RunStateCompleted,
		RunStateFailed,
		RunStateCancelled:
		return true
	default:
		return false
	}
}

func ValidateRunTransition(from RunState, to RunState) error {
	allowed := map[RunState][]RunState{
		RunStateQueued:             {RunStateProvisioning, RunStateCancelled},
		RunStateProvisioning:       {RunStatePreparingWorkspace, RunStateFailed, RunStateCancelled},
		RunStatePreparingWorkspace: {RunStateRunning, RunStateFailed, RunStateCancelled},
		RunStateRunning:            {RunStateCompleted, RunStateFailed, RunStateCancelled},
	}
	for _, candidate := range allowed[from] {
		if candidate == to {
			return nil
		}
	}
	return fmt.Errorf("invalid run transition %s -> %s", from, to)
}

type RunEventType string

const (
	RunEventQueued             RunEventType = "run_queued"
	RunEventProvisioning       RunEventType = "run_provisioning"
	RunEventPreparingWorkspace RunEventType = "run_preparing_workspace"
	RunEventRunning            RunEventType = "run_running"
	RunEventCompleted          RunEventType = "run_completed"
	RunEventFailed             RunEventType = "run_failed"
)

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type Repository struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

type Agent struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	RuntimeKey  string    `json:"runtime_key"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

type IssueStatus string

const (
	IssueStatusOpen IssueStatus = "open"
)

type Issue struct {
	ID           string      `json:"id"`
	WorkspaceID  string      `json:"workspace_id"`
	RepositoryID string      `json:"repository_id"`
	AgentID      string      `json:"agent_id"`
	Title        string      `json:"title"`
	Prompt       string      `json:"prompt"`
	Status       IssueStatus `json:"status"`
	CreatedAt    time.Time   `json:"created_at,omitempty"`
	UpdatedAt    time.Time   `json:"updated_at,omitempty"`
}

type Run struct {
	ID               string    `json:"id"`
	IssueID          string    `json:"issue_id"`
	RepositoryID     string    `json:"repository_id,omitempty"`
	AgentID          string    `json:"agent_id,omitempty"`
	Prompt           string    `json:"prompt,omitempty"`
	State            RunState  `json:"state"`
	ResultSummary    string    `json:"result_summary,omitempty"`
	IdempotencyKey   string    `json:"-"`
	CreatedAt        time.Time `json:"created_at,omitempty"`
	UpdatedAt        time.Time `json:"updated_at,omitempty"`
	StartedAt        time.Time `json:"started_at,omitempty"`
	CompletedAt      time.Time `json:"completed_at,omitempty"`
	LastTransitionAt time.Time `json:"last_transition_at,omitempty"`
}

type RunEvent struct {
	ID        string       `json:"id,omitempty"`
	RunID     string       `json:"run_id"`
	Sequence  int          `json:"sequence"`
	Type      RunEventType `json:"type"`
	Message   string       `json:"message"`
	CreatedAt time.Time    `json:"created_at,omitempty"`
}

type SandboxSession struct {
	ID                string    `json:"id"`
	IssueID           string    `json:"issue_id"`
	RunID             string    `json:"run_id"`
	Provider          string    `json:"provider"`
	ProviderSessionID string    `json:"provider_session_id"`
	State             string    `json:"state"`
	CreatedAt         time.Time `json:"created_at,omitempty"`
}

type RecordSandboxSessionParams struct {
	IssueID           string
	RunID             string
	Provider          string
	ProviderSessionID string
}
