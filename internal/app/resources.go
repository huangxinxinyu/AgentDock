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

func (svc *ResourceService) CreateSandboxTask(ctx context.Context, sandboxID string, req httpapi.CreateSandboxTaskRequest) (domain.SandboxTask, error) {
	session, err := svc.store.GetSandbox(ctx, sandboxID)
	if err != nil {
		return domain.SandboxTask{}, mapStoreError(err)
	}
	if session.State != domain.SandboxStateReady {
		return domain.SandboxTask{}, errors.New("sandbox is not ready")
	}
	workdir := strings.TrimSpace(req.Workdir)
	if workdir == "" {
		workdir = session.DefaultWorkdir
	}
	task, err := svc.store.CreateSandboxTask(ctx, store.CreateSandboxTaskParams{
		SandboxSessionID: sandboxID,
		Prompt:           req.Prompt,
		Entrypoint:       req.Entrypoint,
		Workdir:          workdir,
	})
	if err != nil {
		return domain.SandboxTask{}, err
	}
	if _, err := svc.store.AppendSandboxTaskEvent(ctx, task.ID, domain.SandboxTaskEventQueued, "task queued", "{}"); err != nil {
		return domain.SandboxTask{}, err
	}
	task, err = svc.store.UpdateSandboxTaskState(ctx, task.ID, domain.SandboxTaskStateQueued, domain.SandboxTaskStateStarting, "", "", "")
	if err != nil {
		return domain.SandboxTask{}, err
	}
	if _, err := svc.store.AppendSandboxTaskEvent(ctx, task.ID, domain.SandboxTaskEventStarting, "task starting", "{}"); err != nil {
		return domain.SandboxTask{}, err
	}
	task, err = svc.store.UpdateSandboxTaskState(ctx, task.ID, domain.SandboxTaskStateStarting, domain.SandboxTaskStateRunning, "", "", "")
	if err != nil {
		return domain.SandboxTask{}, err
	}
	if _, err := svc.store.AppendSandboxTaskEvent(ctx, task.ID, domain.SandboxTaskEventRunning, "task running", "{}"); err != nil {
		return domain.SandboxTask{}, err
	}
	result, err := svc.provider.RunTask(ctx, sandbox.TaskRequest{
		TaskID:  task.ID,
		Prompt:  task.Prompt,
		Workdir: task.Workdir,
		Session: sandbox.SessionRef{ProviderSessionID: session.ProviderSessionID, State: string(session.State)},
	})
	if err != nil {
		failed, updateErr := svc.store.UpdateSandboxTaskState(ctx, task.ID, domain.SandboxTaskStateRunning, domain.SandboxTaskStateFailed, "", "", err.Error())
		if updateErr != nil {
			return domain.SandboxTask{}, updateErr
		}
		_, _ = svc.store.AppendSandboxTaskEvent(ctx, task.ID, domain.SandboxTaskEventFailed, err.Error(), "{}")
		return failed, nil
	}
	task, err = svc.store.UpdateSandboxTaskState(ctx, task.ID, domain.SandboxTaskStateRunning, domain.SandboxTaskStateSucceeded, result.Summary, result.OutputRef, "")
	if err != nil {
		return domain.SandboxTask{}, err
	}
	if result.OutputRef != "" {
		if _, err := svc.store.AppendSandboxTaskEvent(ctx, task.ID, domain.SandboxTaskEventOutput, result.OutputRef, "{}"); err != nil {
			return domain.SandboxTask{}, err
		}
	}
	if _, err := svc.store.AppendSandboxTaskEvent(ctx, task.ID, domain.SandboxTaskEventSucceeded, "task succeeded", "{}"); err != nil {
		return domain.SandboxTask{}, err
	}
	return task, nil
}

func (svc *ResourceService) ListSandboxTasks(ctx context.Context, sandboxID string) ([]domain.SandboxTask, error) {
	return svc.store.ListSandboxTasks(ctx, sandboxID)
}

func (svc *ResourceService) GetSandboxTask(ctx context.Context, id string) (domain.SandboxTask, error) {
	task, err := svc.store.GetSandboxTask(ctx, id)
	return task, mapStoreError(err)
}

func (svc *ResourceService) ListSandboxTaskEvents(ctx context.Context, taskID string) ([]domain.SandboxTaskEvent, error) {
	return svc.store.ListSandboxTaskEvents(ctx, taskID)
}

func (svc *ResourceService) CancelSandboxTask(ctx context.Context, id string) (domain.SandboxTask, error) {
	task, err := svc.store.GetSandboxTask(ctx, id)
	if err != nil {
		return domain.SandboxTask{}, mapStoreError(err)
	}
	if task.State == domain.SandboxTaskStateCancelled {
		return task, nil
	}
	if task.State == domain.SandboxTaskStateSucceeded || task.State == domain.SandboxTaskStateFailed {
		return task, nil
	}
	session, err := svc.store.GetSandbox(ctx, task.SandboxSessionID)
	if err != nil {
		return domain.SandboxTask{}, mapStoreError(err)
	}
	_ = svc.provider.CancelTask(ctx, sandbox.TaskRef{
		TaskID:  task.ID,
		Session: sandbox.SessionRef{ProviderSessionID: session.ProviderSessionID, State: string(session.State)},
	})
	cancelled, err := svc.store.UpdateSandboxTaskState(ctx, task.ID, task.State, domain.SandboxTaskStateCancelled, "", "", "")
	if err != nil {
		return domain.SandboxTask{}, err
	}
	if _, err := svc.store.AppendSandboxTaskEvent(ctx, task.ID, domain.SandboxTaskEventCancelled, "task cancelled", "{}"); err != nil {
		return domain.SandboxTask{}, err
	}
	return cancelled, nil
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
