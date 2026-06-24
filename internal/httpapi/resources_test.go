package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/huangxinxinyu/agentdock/internal/domain"
)

type recordingResourceService struct {
	createWorkspaceReq CreateWorkspaceRequest
	createRunReq       CreateRunRequest
	createSandboxReq   CreateSandboxRequest
	pauseSandboxID     string
	inspectSandboxID   string
}

func (svc *recordingResourceService) CreateWorkspace(_ context.Context, req CreateWorkspaceRequest) (domain.Workspace, error) {
	svc.createWorkspaceReq = req
	return domain.Workspace{ID: "workspace-1", Name: req.Name}, nil
}

func (svc *recordingResourceService) GetWorkspace(context.Context, string) (domain.Workspace, error) {
	return domain.Workspace{}, ErrNotFound
}

func (svc *recordingResourceService) CreateRepository(context.Context, string, CreateRepositoryRequest) (domain.Repository, error) {
	return domain.Repository{}, nil
}

func (svc *recordingResourceService) CreateAgent(context.Context, string, CreateAgentRequest) (domain.Agent, error) {
	return domain.Agent{}, nil
}

func (svc *recordingResourceService) CreateIssue(context.Context, string, CreateIssueRequest) (domain.Issue, error) {
	return domain.Issue{}, nil
}

func (svc *recordingResourceService) GetIssue(context.Context, string) (domain.Issue, error) {
	return domain.Issue{}, ErrNotFound
}

func (svc *recordingResourceService) CreateRun(_ context.Context, issueID string, req CreateRunRequest) (domain.Run, error) {
	svc.createRunReq = req
	return domain.Run{ID: "run-1", IssueID: issueID, State: domain.RunStateQueued}, nil
}

func (svc *recordingResourceService) GetRun(context.Context, string) (domain.Run, error) {
	return domain.Run{}, ErrNotFound
}

func (svc *recordingResourceService) ListRunEvents(context.Context, string) ([]domain.RunEvent, error) {
	return nil, nil
}

func (svc *recordingResourceService) CreateSandbox(_ context.Context, req CreateSandboxRequest) (domain.SandboxSession, error) {
	svc.createSandboxReq = req
	return domain.SandboxSession{ID: "sandbox-1", Name: req.Name, Provider: "noop", State: domain.SandboxStateReady, DefaultWorkdir: "/workspace"}, nil
}

func (svc *recordingResourceService) ListSandboxes(context.Context) ([]domain.SandboxSession, error) {
	return []domain.SandboxSession{{ID: "sandbox-1", Name: "scratch", Provider: "noop", State: domain.SandboxStateReady}}, nil
}

func (svc *recordingResourceService) GetSandbox(context.Context, string) (domain.SandboxSession, error) {
	return domain.SandboxSession{}, ErrNotFound
}

func (svc *recordingResourceService) PauseSandbox(_ context.Context, id string) (domain.SandboxSession, error) {
	svc.pauseSandboxID = id
	return domain.SandboxSession{ID: id, Name: "scratch", Provider: "noop", State: domain.SandboxStatePaused}, nil
}

func (svc *recordingResourceService) ResumeSandbox(context.Context, string) (domain.SandboxSession, error) {
	return domain.SandboxSession{}, nil
}

func (svc *recordingResourceService) CloseSandbox(context.Context, string) (domain.SandboxSession, error) {
	return domain.SandboxSession{}, nil
}

func (svc *recordingResourceService) InspectSandbox(_ context.Context, id string) (domain.SandboxSession, error) {
	svc.inspectSandboxID = id
	return domain.SandboxSession{ID: id, Name: "scratch", Provider: "noop", State: domain.SandboxStateReady}, nil
}

func TestCreateWorkspaceEndpoint(t *testing.T) {
	svc := &recordingResourceService{}
	router := NewRouter(Dependencies{Resources: svc})

	body := bytes.NewBufferString(`{"name":"Acme"}`)
	req := httptest.NewRequest(http.MethodPost, "/workspaces", body)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if svc.createWorkspaceReq.Name != "Acme" {
		t.Fatalf("workspace request = %#v", svc.createWorkspaceReq)
	}

	var workspace domain.Workspace
	if err := json.NewDecoder(rec.Body).Decode(&workspace); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if workspace.ID != "workspace-1" {
		t.Fatalf("workspace response = %#v", workspace)
	}
}

func TestCreateRunEndpointPassesIdempotencyKey(t *testing.T) {
	svc := &recordingResourceService{}
	router := NewRouter(Dependencies{Resources: svc})

	req := httptest.NewRequest(http.MethodPost, "/issues/issue-1/runs", bytes.NewBufferString(`{}`))
	req.Header.Set("Idempotency-Key", "retry-key-1")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if svc.createRunReq.IdempotencyKey != "retry-key-1" {
		t.Fatalf("idempotency key = %q", svc.createRunReq.IdempotencyKey)
	}
}

func TestCreateSandboxEndpoint(t *testing.T) {
	svc := &recordingResourceService{}
	router := NewRouter(Dependencies{Resources: svc})

	req := httptest.NewRequest(http.MethodPost, "/sandboxes", bytes.NewBufferString(`{"name":"Scratch","agentos_image":"agentos:test"}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if svc.createSandboxReq.Name != "Scratch" {
		t.Fatalf("sandbox request = %#v", svc.createSandboxReq)
	}

	var sandbox domain.SandboxSession
	if err := json.NewDecoder(rec.Body).Decode(&sandbox); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if sandbox.State != domain.SandboxStateReady {
		t.Fatalf("sandbox state = %s, want ready", sandbox.State)
	}
}

func TestPauseSandboxEndpoint(t *testing.T) {
	svc := &recordingResourceService{}
	router := NewRouter(Dependencies{Resources: svc})

	req := httptest.NewRequest(http.MethodPost, "/sandboxes/sandbox-1/pause", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.pauseSandboxID != "sandbox-1" {
		t.Fatalf("pause sandbox id = %q", svc.pauseSandboxID)
	}
}

func TestInspectSandboxEndpoint(t *testing.T) {
	svc := &recordingResourceService{}
	router := NewRouter(Dependencies{Resources: svc})

	req := httptest.NewRequest(http.MethodPost, "/sandboxes/sandbox-1/inspect", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.inspectSandboxID != "sandbox-1" {
		t.Fatalf("inspect sandbox id = %q", svc.inspectSandboxID)
	}
}
