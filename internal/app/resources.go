package app

import (
	"context"
	"errors"

	"github.com/huangxinxinyu/agentdock/internal/domain"
	"github.com/huangxinxinyu/agentdock/internal/httpapi"
	"github.com/huangxinxinyu/agentdock/internal/store"
)

type ResourceService struct {
	store *store.PostgresStore
}

func NewResourceService(store *store.PostgresStore) *ResourceService {
	return &ResourceService{store: store}
}

func (svc *ResourceService) CreateWorkspace(ctx context.Context, req httpapi.CreateWorkspaceRequest) (domain.Workspace, error) {
	return svc.store.CreateWorkspace(ctx, store.CreateWorkspaceParams{Name: req.Name})
}

func (svc *ResourceService) GetWorkspace(ctx context.Context, id string) (domain.Workspace, error) {
	workspace, err := svc.store.GetWorkspace(ctx, id)
	return workspace, mapStoreError(err)
}

func (svc *ResourceService) CreateRepository(ctx context.Context, workspaceID string, req httpapi.CreateRepositoryRequest) (domain.Repository, error) {
	return svc.store.CreateRepository(ctx, store.CreateRepositoryParams{
		WorkspaceID: workspaceID,
		Name:        req.Name,
		URL:         req.URL,
	})
}

func (svc *ResourceService) CreateAgent(ctx context.Context, workspaceID string, req httpapi.CreateAgentRequest) (domain.Agent, error) {
	runtimeKey := req.RuntimeKey
	if runtimeKey == "" {
		runtimeKey = "noop"
	}
	return svc.store.CreateAgent(ctx, store.CreateAgentParams{
		WorkspaceID: workspaceID,
		Name:        req.Name,
		RuntimeKey:  runtimeKey,
	})
}

func (svc *ResourceService) CreateIssue(ctx context.Context, workspaceID string, req httpapi.CreateIssueRequest) (domain.Issue, error) {
	return svc.store.CreateIssue(ctx, store.CreateIssueParams{
		WorkspaceID:  workspaceID,
		RepositoryID: req.RepositoryID,
		AgentID:      req.AgentID,
		Title:        req.Title,
		Prompt:       req.Prompt,
	})
}

func (svc *ResourceService) GetIssue(ctx context.Context, id string) (domain.Issue, error) {
	issue, err := svc.store.GetIssue(ctx, id)
	return issue, mapStoreError(err)
}

func (svc *ResourceService) CreateRun(ctx context.Context, issueID string, req httpapi.CreateRunRequest) (domain.Run, error) {
	return svc.store.CreateRun(ctx, store.CreateRunParams{
		IssueID:        issueID,
		IdempotencyKey: req.IdempotencyKey,
	})
}

func (svc *ResourceService) GetRun(ctx context.Context, id string) (domain.Run, error) {
	run, err := svc.store.GetRun(ctx, id)
	return run, mapStoreError(err)
}

func (svc *ResourceService) ListRunEvents(ctx context.Context, runID string) ([]domain.RunEvent, error) {
	return svc.store.ListRunEvents(ctx, runID)
}

func mapStoreError(err error) error {
	if errors.Is(err, store.ErrNotFound) {
		return httpapi.ErrNotFound
	}
	return err
}
