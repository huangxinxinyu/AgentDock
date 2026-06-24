package store

import (
	"context"
	"os"
	"testing"

	"github.com/huangxinxinyu/agentdock/internal/domain"
	"github.com/huangxinxinyu/agentdock/internal/runtime"
	"github.com/huangxinxinyu/agentdock/internal/sandbox"
	"github.com/huangxinxinyu/agentdock/internal/worker"
)

func TestPostgresStoreCreatesRunIdempotently(t *testing.T) {
	databaseURL := os.Getenv("AGENTDOCK_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set AGENTDOCK_TEST_DATABASE_URL to run Postgres integration tests")
	}

	ctx := context.Background()
	store, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer store.Close()

	if err := ApplyMigrations(ctx, store.DB(), "../../db/migrations"); err != nil {
		t.Fatalf("ApplyMigrations returned error: %v", err)
	}
	if err := store.TruncateForTest(ctx); err != nil {
		t.Fatalf("TruncateForTest returned error: %v", err)
	}

	workspace, err := store.CreateWorkspace(ctx, CreateWorkspaceParams{Name: "Acme"})
	if err != nil {
		t.Fatalf("CreateWorkspace returned error: %v", err)
	}
	repository, err := store.CreateRepository(ctx, CreateRepositoryParams{
		WorkspaceID: workspace.ID,
		Name:        "agentdock",
		URL:         "https://github.com/example/agentdock",
	})
	if err != nil {
		t.Fatalf("CreateRepository returned error: %v", err)
	}
	agent, err := store.CreateAgent(ctx, CreateAgentParams{
		WorkspaceID: workspace.ID,
		Name:        "Noop Agent",
		RuntimeKey:  "noop",
	})
	if err != nil {
		t.Fatalf("CreateAgent returned error: %v", err)
	}
	issue, err := store.CreateIssue(ctx, CreateIssueParams{
		WorkspaceID:  workspace.ID,
		RepositoryID: repository.ID,
		AgentID:      agent.ID,
		Title:        "Explain this code",
		Prompt:       "Explain the code without changing files.",
	})
	if err != nil {
		t.Fatalf("CreateIssue returned error: %v", err)
	}

	firstRun, err := store.CreateRun(ctx, CreateRunParams{IssueID: issue.ID, IdempotencyKey: "retry-1"})
	if err != nil {
		t.Fatalf("CreateRun first returned error: %v", err)
	}
	secondRun, err := store.CreateRun(ctx, CreateRunParams{IssueID: issue.ID, IdempotencyKey: "retry-1"})
	if err != nil {
		t.Fatalf("CreateRun retry returned error: %v", err)
	}
	if firstRun.ID != secondRun.ID {
		t.Fatalf("idempotent run IDs differ: %s != %s", firstRun.ID, secondRun.ID)
	}

	events, err := store.ListRunEvents(ctx, firstRun.ID)
	if err != nil {
		t.Fatalf("ListRunEvents returned error: %v", err)
	}
	if len(events) != 1 || events[0].Type != "run_queued" {
		t.Fatalf("events = %#v, want run_queued event", events)
	}
}

func TestPostgresStoreSupportsWorkerLifecycle(t *testing.T) {
	databaseURL := os.Getenv("AGENTDOCK_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set AGENTDOCK_TEST_DATABASE_URL to run Postgres integration tests")
	}

	ctx := context.Background()
	store, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer store.Close()

	if err := ApplyMigrations(ctx, store.DB(), "../../db/migrations"); err != nil {
		t.Fatalf("ApplyMigrations returned error: %v", err)
	}
	if err := store.TruncateForTest(ctx); err != nil {
		t.Fatalf("TruncateForTest returned error: %v", err)
	}

	_, _, _, issue := createRunFixture(t, ctx, store)
	run, err := store.CreateRun(ctx, CreateRunParams{IssueID: issue.ID, IdempotencyKey: "worker-1"})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}

	processor := worker.NewProcessor(store, sandbox.NoopProvider{}, runtime.NoopRunner{})
	processed, err := processor.ProcessNext(ctx)
	if err != nil {
		t.Fatalf("ProcessNext returned error: %v", err)
	}
	if !processed {
		t.Fatal("ProcessNext did not claim run")
	}

	completed, err := store.GetRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRun returned error: %v", err)
	}
	if completed.State != domain.RunStateCompleted {
		t.Fatalf("run state = %s, want %s", completed.State, domain.RunStateCompleted)
	}
	if completed.ResultSummary == "" {
		t.Fatal("result summary was empty")
	}

	events, err := store.ListRunEvents(ctx, run.ID)
	if err != nil {
		t.Fatalf("ListRunEvents returned error: %v", err)
	}
	if len(events) < 5 {
		t.Fatalf("events = %#v, want queued plus lifecycle events", events)
	}
}

func createRunFixture(t *testing.T, ctx context.Context, store *PostgresStore) (domain.Workspace, domain.Repository, domain.Agent, domain.Issue) {
	t.Helper()
	workspace, err := store.CreateWorkspace(ctx, CreateWorkspaceParams{Name: "Acme"})
	if err != nil {
		t.Fatalf("CreateWorkspace returned error: %v", err)
	}
	repository, err := store.CreateRepository(ctx, CreateRepositoryParams{
		WorkspaceID: workspace.ID,
		Name:        "agentdock",
		URL:         "https://github.com/example/agentdock",
	})
	if err != nil {
		t.Fatalf("CreateRepository returned error: %v", err)
	}
	agent, err := store.CreateAgent(ctx, CreateAgentParams{
		WorkspaceID: workspace.ID,
		Name:        "Noop Agent",
		RuntimeKey:  "noop",
	})
	if err != nil {
		t.Fatalf("CreateAgent returned error: %v", err)
	}
	issue, err := store.CreateIssue(ctx, CreateIssueParams{
		WorkspaceID:  workspace.ID,
		RepositoryID: repository.ID,
		AgentID:      agent.ID,
		Title:        "Explain this code",
		Prompt:       "Explain the code without changing files.",
	})
	if err != nil {
		t.Fatalf("CreateIssue returned error: %v", err)
	}
	return workspace, repository, agent, issue
}
