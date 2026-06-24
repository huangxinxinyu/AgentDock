package domain

import (
	"fmt"
	"time"
)

type SandboxTaskState string

const (
	SandboxTaskStateQueued    SandboxTaskState = "queued"
	SandboxTaskStateStarting  SandboxTaskState = "starting"
	SandboxTaskStateRunning   SandboxTaskState = "running"
	SandboxTaskStateSucceeded SandboxTaskState = "succeeded"
	SandboxTaskStateFailed    SandboxTaskState = "failed"
	SandboxTaskStateCancelled SandboxTaskState = "cancelled"
)

func ValidateSandboxTaskTransition(from SandboxTaskState, to SandboxTaskState) error {
	if from == to {
		return nil
	}
	allowed := map[SandboxTaskState][]SandboxTaskState{
		SandboxTaskStateQueued:   {SandboxTaskStateStarting, SandboxTaskStateCancelled, SandboxTaskStateFailed},
		SandboxTaskStateStarting: {SandboxTaskStateRunning, SandboxTaskStateCancelled, SandboxTaskStateFailed},
		SandboxTaskStateRunning:  {SandboxTaskStateSucceeded, SandboxTaskStateFailed, SandboxTaskStateCancelled},
	}
	for _, candidate := range allowed[from] {
		if candidate == to {
			return nil
		}
	}
	return fmt.Errorf("invalid sandbox task transition %s -> %s", from, to)
}

type SandboxTaskEventType string

const (
	SandboxTaskEventQueued    SandboxTaskEventType = "task_queued"
	SandboxTaskEventStarting  SandboxTaskEventType = "task_starting"
	SandboxTaskEventRunning   SandboxTaskEventType = "task_running"
	SandboxTaskEventLog       SandboxTaskEventType = "log"
	SandboxTaskEventOutput    SandboxTaskEventType = "output"
	SandboxTaskEventSucceeded SandboxTaskEventType = "task_succeeded"
	SandboxTaskEventFailed    SandboxTaskEventType = "task_failed"
	SandboxTaskEventCancelled SandboxTaskEventType = "task_cancelled"
)

type SandboxTask struct {
	ID               string           `json:"id"`
	SandboxSessionID string           `json:"sandbox_session_id"`
	Prompt           string           `json:"prompt,omitempty"`
	State            SandboxTaskState `json:"state"`
	Entrypoint       string           `json:"entrypoint"`
	Workdir          string           `json:"workdir"`
	Summary          string           `json:"summary,omitempty"`
	OutputRef        string           `json:"output_ref,omitempty"`
	LastError        string           `json:"last_error,omitempty"`
	CreatedAt        time.Time        `json:"created_at,omitempty"`
	UpdatedAt        time.Time        `json:"updated_at,omitempty"`
	StartedAt        time.Time        `json:"started_at,omitempty"`
	CompletedAt      time.Time        `json:"completed_at,omitempty"`
}

type SandboxTaskEvent struct {
	ID            string               `json:"id,omitempty"`
	SandboxTaskID string               `json:"sandbox_task_id"`
	Sequence      int                  `json:"sequence"`
	Type          SandboxTaskEventType `json:"type"`
	Message       string               `json:"message"`
	Payload       string               `json:"payload,omitempty"`
	CreatedAt     time.Time            `json:"created_at,omitempty"`
}
