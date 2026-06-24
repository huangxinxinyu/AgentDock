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
