package worker

import (
	"context"
	"testing"

	"github.com/huangxinxinyu/agentdock/internal/domain"
	"github.com/huangxinxinyu/agentdock/internal/runtime"
	"github.com/huangxinxinyu/agentdock/internal/sandbox"
)

type memoryRunStore struct {
	run      domain.Run
	events   []domain.RunEvent
	sessions []domain.SandboxSession
}

func (store *memoryRunStore) ClaimQueuedRun(context.Context) (domain.Run, bool, error) {
	if store.run.State != domain.RunStateQueued {
		return domain.Run{}, false, nil
	}
	store.run.State = domain.RunStateProvisioning
	return store.run, true, nil
}

func (store *memoryRunStore) AdvanceRun(_ context.Context, runID string, from domain.RunState, to domain.RunState) (domain.Run, error) {
	if store.run.ID != runID {
		return domain.Run{}, ErrRunNotFound
	}
	if store.run.State != from {
		return domain.Run{}, ErrRunStateChanged
	}
	store.run.State = to
	return store.run, nil
}

func (store *memoryRunStore) CompleteRun(_ context.Context, runID string, summary string) (domain.Run, error) {
	if store.run.ID != runID {
		return domain.Run{}, ErrRunNotFound
	}
	store.run.State = domain.RunStateCompleted
	store.run.ResultSummary = summary
	return store.run, nil
}

func (store *memoryRunStore) AppendRunEvent(_ context.Context, runID string, eventType domain.RunEventType, message string) (domain.RunEvent, error) {
	event := domain.RunEvent{RunID: runID, Sequence: len(store.events) + 1, Type: eventType, Message: message}
	store.events = append(store.events, event)
	return event, nil
}

func (store *memoryRunStore) RecordSandboxSession(_ context.Context, params RecordSandboxSessionParams) (domain.SandboxSession, error) {
	session := domain.SandboxSession{
		IssueID:           params.IssueID,
		RunID:             params.RunID,
		Provider:          params.Provider,
		ProviderSessionID: params.ProviderSessionID,
		State:             "active",
	}
	store.sessions = append(store.sessions, session)
	return session, nil
}

func TestProcessNextCompletesQueuedRun(t *testing.T) {
	store := &memoryRunStore{
		run: domain.Run{
			ID:      "run-1",
			IssueID: "issue-1",
			State:   domain.RunStateQueued,
			Prompt:  "explain the code",
		},
	}
	processor := NewProcessor(store, sandbox.NoopProvider{}, runtime.NoopRunner{})

	processed, err := processor.ProcessNext(context.Background())
	if err != nil {
		t.Fatalf("ProcessNext returned error: %v", err)
	}
	if !processed {
		t.Fatal("ProcessNext did not process queued run")
	}
	if store.run.State != domain.RunStateCompleted {
		t.Fatalf("run state = %s, want %s", store.run.State, domain.RunStateCompleted)
	}
	if store.run.ResultSummary == "" {
		t.Fatal("run summary was not recorded")
	}
	if len(store.sessions) != 1 || store.sessions[0].Provider != "noop" {
		t.Fatalf("sandbox sessions = %#v, want recorded noop session", store.sessions)
	}
	if len(store.events) < 4 {
		t.Fatalf("events = %#v, want lifecycle events", store.events)
	}
}
