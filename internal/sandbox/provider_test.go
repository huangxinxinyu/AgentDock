package sandbox

import (
	"context"
	"testing"
)

func TestNoopProviderSatisfiesProviderBoundary(t *testing.T) {
	var provider Provider = NoopProvider{}
	session, err := provider.CreateSession(context.Background(), CreateSessionRequest{
		WorkspaceID: "workspace_123",
		IssueID:     "issue_123",
	})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}
	if session.Provider != "noop" {
		t.Fatalf("Provider = %q, want noop", session.Provider)
	}
}
