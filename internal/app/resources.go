package app

import (
	"context"
	"errors"
	"strings"

	"github.com/huangxinxinyu/agentdock/internal/domain"
	"github.com/huangxinxinyu/agentdock/internal/httpapi"
	"github.com/huangxinxinyu/agentdock/internal/sandbox"
	"github.com/huangxinxinyu/agentdock/internal/store"
)

type ResourceService struct {
	store    *store.PostgresStore
	provider sandbox.Provider
}

func NewResourceService(store *store.PostgresStore, provider sandbox.Provider) *ResourceService {
	if provider == nil {
		provider = sandbox.NoopProvider{}
	}
	return &ResourceService{store: store, provider: provider}
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

func (svc *ResourceService) CreateSandbox(ctx context.Context, req httpapi.CreateSandboxRequest) (domain.SandboxSession, error) {
	providerName := strings.TrimSpace(req.Provider)
	if providerName == "" {
		providerName = "noop"
	}
	defaultWorkdir := strings.TrimSpace(req.DefaultWorkdir)
	if defaultWorkdir == "" {
		defaultWorkdir = "/workspace"
	}
	session, err := svc.provider.CreateSession(ctx, sandbox.CreateSessionRequest{
		Name:           req.Name,
		DefaultWorkdir: defaultWorkdir,
		AgentOSImage:   req.AgentOSImage,
	})
	if err != nil {
		if errors.Is(err, sandbox.ErrProviderNotConfigured) {
			return domain.SandboxSession{}, httpapi.ErrProviderNotConfigured
		}
		return domain.SandboxSession{}, err
	}
	if session.Provider != "" {
		providerName = session.Provider
	}
	return svc.store.CreateSandbox(ctx, store.CreateSandboxParams{
		Name:              req.Name,
		Provider:          providerName,
		ProviderSessionID: session.ID,
		State:             domain.SandboxState(session.State),
		DefaultWorkdir:    session.DefaultWorkdir,
		AgentOSImage:      req.AgentOSImage,
		Metadata:          session.Metadata,
	})
}

func (svc *ResourceService) ListSandboxes(ctx context.Context) ([]domain.SandboxSession, error) {
	return svc.store.ListSandboxes(ctx)
}

func (svc *ResourceService) GetSandbox(ctx context.Context, id string) (domain.SandboxSession, error) {
	session, err := svc.store.GetSandbox(ctx, id)
	return session, mapStoreError(err)
}

func (svc *ResourceService) PauseSandbox(ctx context.Context, id string) (domain.SandboxSession, error) {
	return svc.transitionSandbox(ctx, id, domain.SandboxStatePaused, svc.provider.PauseSession)
}

func (svc *ResourceService) ResumeSandbox(ctx context.Context, id string) (domain.SandboxSession, error) {
	return svc.transitionSandbox(ctx, id, domain.SandboxStateReady, svc.provider.ResumeSession)
}

func (svc *ResourceService) CloseSandbox(ctx context.Context, id string) (domain.SandboxSession, error) {
	return svc.transitionSandbox(ctx, id, domain.SandboxStateClosed, svc.provider.CloseSession)
}

func (svc *ResourceService) InspectSandbox(ctx context.Context, id string) (domain.SandboxSession, error) {
	session, err := svc.store.GetSandbox(ctx, id)
	if err != nil {
		return domain.SandboxSession{}, mapStoreError(err)
	}
	observed, err := svc.provider.InspectSession(ctx, sandbox.SessionRef{
		ProviderSessionID: session.ProviderSessionID,
		State:             string(session.State),
	})
	if err != nil {
		if errors.Is(err, sandbox.ErrProviderNotConfigured) {
			return domain.SandboxSession{}, httpapi.ErrProviderNotConfigured
		}
		return domain.SandboxSession{}, err
	}
	if observed.State != "" {
		session.State = domain.SandboxState(observed.State)
	}
	if observed.Metadata != "" {
		session.Metadata = observed.Metadata
	}
	return session, nil
}

func (svc *ResourceService) transitionSandbox(ctx context.Context, id string, target domain.SandboxState, callProvider func(context.Context, sandbox.SessionRef) (sandbox.SessionObservation, error)) (domain.SandboxSession, error) {
	session, err := svc.store.GetSandbox(ctx, id)
	if err != nil {
		return domain.SandboxSession{}, mapStoreError(err)
	}
	if session.State == target {
		return session, nil
	}
	if err := domain.ValidateSandboxTransition(session.State, target); err != nil {
		return domain.SandboxSession{}, err
	}
	observed, err := callProvider(ctx, sandbox.SessionRef{
		ProviderSessionID: session.ProviderSessionID,
		State:             string(session.State),
	})
	if err != nil {
		if errors.Is(err, sandbox.ErrProviderNotConfigured) {
			return domain.SandboxSession{}, httpapi.ErrProviderNotConfigured
		}
		return domain.SandboxSession{}, err
	}
	nextState := target
	if observed.State != "" {
		nextState = domain.SandboxState(observed.State)
	}
	return svc.store.UpdateSandboxState(ctx, id, session.State, nextState, "")
}

func mapStoreError(err error) error {
	if errors.Is(err, store.ErrNotFound) {
		return httpapi.ErrNotFound
	}
	return err
}
