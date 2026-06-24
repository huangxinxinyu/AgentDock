package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/huangxinxinyu/agentdock/internal/domain"
)

var ErrNotFound = errors.New("resource not found")

type ResourceService interface {
	CreateWorkspace(context.Context, CreateWorkspaceRequest) (domain.Workspace, error)
	GetWorkspace(context.Context, string) (domain.Workspace, error)
	CreateRepository(context.Context, string, CreateRepositoryRequest) (domain.Repository, error)
	CreateAgent(context.Context, string, CreateAgentRequest) (domain.Agent, error)
	CreateIssue(context.Context, string, CreateIssueRequest) (domain.Issue, error)
	GetIssue(context.Context, string) (domain.Issue, error)
	CreateRun(context.Context, string, CreateRunRequest) (domain.Run, error)
	GetRun(context.Context, string) (domain.Run, error)
	ListRunEvents(context.Context, string) ([]domain.RunEvent, error)
}

type CreateWorkspaceRequest struct {
	Name string `json:"name"`
}

type CreateRepositoryRequest struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type CreateAgentRequest struct {
	Name       string `json:"name"`
	RuntimeKey string `json:"runtime_key"`
}

type CreateIssueRequest struct {
	RepositoryID string `json:"repository_id"`
	AgentID      string `json:"agent_id"`
	Title        string `json:"title"`
	Prompt       string `json:"prompt"`
}

type CreateRunRequest struct {
	IdempotencyKey string `json:"-"`
}

func registerResourceRoutes(mux *http.ServeMux, svc ResourceService) {
	if svc == nil {
		return
	}

	mux.HandleFunc("POST /workspaces", func(w http.ResponseWriter, r *http.Request) {
		var req CreateWorkspaceRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		workspace, err := svc.CreateWorkspace(r.Context(), req)
		writeResourceResponse(w, http.StatusCreated, workspace, err)
	})

	mux.HandleFunc("GET /workspaces/{id}", func(w http.ResponseWriter, r *http.Request) {
		workspace, err := svc.GetWorkspace(r.Context(), r.PathValue("id"))
		writeResourceResponse(w, http.StatusOK, workspace, err)
	})

	mux.HandleFunc("POST /workspaces/{workspace_id}/repositories", func(w http.ResponseWriter, r *http.Request) {
		var req CreateRepositoryRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		repository, err := svc.CreateRepository(r.Context(), r.PathValue("workspace_id"), req)
		writeResourceResponse(w, http.StatusCreated, repository, err)
	})

	mux.HandleFunc("POST /workspaces/{workspace_id}/agents", func(w http.ResponseWriter, r *http.Request) {
		var req CreateAgentRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		agent, err := svc.CreateAgent(r.Context(), r.PathValue("workspace_id"), req)
		writeResourceResponse(w, http.StatusCreated, agent, err)
	})

	mux.HandleFunc("POST /workspaces/{workspace_id}/issues", func(w http.ResponseWriter, r *http.Request) {
		var req CreateIssueRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		issue, err := svc.CreateIssue(r.Context(), r.PathValue("workspace_id"), req)
		writeResourceResponse(w, http.StatusCreated, issue, err)
	})

	mux.HandleFunc("GET /issues/{id}", func(w http.ResponseWriter, r *http.Request) {
		issue, err := svc.GetIssue(r.Context(), r.PathValue("id"))
		writeResourceResponse(w, http.StatusOK, issue, err)
	})

	mux.HandleFunc("POST /issues/{id}/runs", func(w http.ResponseWriter, r *http.Request) {
		var req CreateRunRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		req.IdempotencyKey = strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		run, err := svc.CreateRun(r.Context(), r.PathValue("id"), req)
		writeResourceResponse(w, http.StatusAccepted, run, err)
	})

	mux.HandleFunc("GET /runs/{id}", func(w http.ResponseWriter, r *http.Request) {
		run, err := svc.GetRun(r.Context(), r.PathValue("id"))
		writeResourceResponse(w, http.StatusOK, run, err)
	})

	mux.HandleFunc("GET /runs/{id}/events", func(w http.ResponseWriter, r *http.Request) {
		events, err := svc.ListRunEvents(r.Context(), r.PathValue("id"))
		writeResourceResponse(w, http.StatusOK, events, err)
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return false
	}
	return true
}

func writeResourceResponse(w http.ResponseWriter, status int, body any, err error) {
	if err == nil {
		writeJSON(w, status, body)
		return
	}
	if errors.Is(err, ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
}
