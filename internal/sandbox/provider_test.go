package sandbox

import (
	"context"
	"testing"
)

func TestNoopProviderSatisfiesProviderBoundary(t *testing.T) {
	var provider Provider = NoopProvider{}
	session, err := provider.CreateSession(context.Background(), CreateSessionRequest{
		Name:           "scratch",
		DefaultWorkdir: "/workspace",
		AgentOSImage:   "agentos:test",
	})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}
	if session.Provider != "noop" {
		t.Fatalf("Provider = %q, want noop", session.Provider)
	}
	if session.DefaultWorkdir != "/workspace" {
		t.Fatalf("DefaultWorkdir = %q, want /workspace", session.DefaultWorkdir)
	}

	inspected, err := provider.InspectSession(context.Background(), SessionRef{
		ProviderSessionID: session.ID,
		State:             "ready",
	})
	if err != nil {
		t.Fatalf("InspectSession returned error: %v", err)
	}
	if inspected.State != "ready" {
		t.Fatalf("Inspect state = %q, want ready", inspected.State)
	}

	result, err := provider.RunTask(context.Background(), TaskRequest{
		TaskID:  "task-1",
		Prompt:  "create a file",
		Workdir: "/workspace",
		Session: SessionRef{ProviderSessionID: session.ID, State: "ready"},
	})
	if err != nil {
		t.Fatalf("RunTask returned error: %v", err)
	}
	if result.Summary == "" || result.OutputRef == "" {
		t.Fatalf("RunTask result = %#v, want summary and output ref", result)
	}
}
