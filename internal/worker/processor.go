package worker

import (
	"context"
	"errors"
	"time"

	"github.com/huangxinxinyu/agentdock/internal/domain"
	"github.com/huangxinxinyu/agentdock/internal/runtime"
	"github.com/huangxinxinyu/agentdock/internal/sandbox"
)

var (
	ErrRunNotFound     = errors.New("run not found")
	ErrRunStateChanged = errors.New("run state changed")
)

type RunStore interface {
	ClaimQueuedRun(context.Context) (domain.Run, bool, error)
	AdvanceRun(context.Context, string, domain.RunState, domain.RunState) (domain.Run, error)
	CompleteRun(context.Context, string, string) (domain.Run, error)
	AppendRunEvent(context.Context, string, domain.RunEventType, string) (domain.RunEvent, error)
	RecordSandboxSession(context.Context, RecordSandboxSessionParams) (domain.SandboxSession, error)
}

type RecordSandboxSessionParams = domain.RecordSandboxSessionParams

type Processor struct {
	store    RunStore
	provider sandbox.Provider
	runner   runtime.Runner
}

func NewProcessor(store RunStore, provider sandbox.Provider, runner runtime.Runner) *Processor {
	return &Processor{store: store, provider: provider, runner: runner}
}

func (processor *Processor) ProcessNext(ctx context.Context) (bool, error) {
	run, ok, err := processor.store.ClaimQueuedRun(ctx)
	if err != nil || !ok {
		return ok, err
	}
	if _, err := processor.store.AppendRunEvent(ctx, run.ID, domain.RunEventProvisioning, "run provisioning started"); err != nil {
		return true, err
	}
	session, err := processor.provider.CreateSession(ctx, sandbox.CreateSessionRequest{
		Name:           "run-" + run.ID,
		DefaultWorkdir: "/workspace",
	})
	if err != nil {
		return true, err
	}
	if _, err := processor.store.RecordSandboxSession(ctx, RecordSandboxSessionParams{
		Name:              "run-" + run.ID,
		Provider:          session.Provider,
		ProviderSessionID: session.ID,
		State:             domain.SandboxStateReady,
		DefaultWorkdir:    session.DefaultWorkdir,
		Metadata:          session.Metadata,
	}); err != nil {
		return true, err
	}

	if run, err = processor.transition(ctx, run, domain.RunStateProvisioning, domain.RunStatePreparingWorkspace, domain.RunEventPreparingWorkspace, "workspace preparation started"); err != nil {
		return true, err
	}
	if run, err = processor.transition(ctx, run, domain.RunStatePreparingWorkspace, domain.RunStateRunning, domain.RunEventRunning, "runtime execution started"); err != nil {
		return true, err
	}
	result, err := processor.runner.Start(ctx, runtime.StartRequest{RunID: run.ID, Prompt: run.Prompt})
	if err != nil {
		return true, err
	}
	if _, err := processor.store.CompleteRun(ctx, run.ID, result.Summary); err != nil {
		return true, err
	}
	if _, err := processor.store.AppendRunEvent(ctx, run.ID, domain.RunEventCompleted, "run completed"); err != nil {
		return true, err
	}
	return true, nil
}

func (processor *Processor) Start(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if _, err := processor.ProcessNext(ctx); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (processor *Processor) transition(ctx context.Context, run domain.Run, from domain.RunState, to domain.RunState, eventType domain.RunEventType, message string) (domain.Run, error) {
	next, err := processor.store.AdvanceRun(ctx, run.ID, from, to)
	if err != nil {
		return domain.Run{}, err
	}
	if _, err := processor.store.AppendRunEvent(ctx, run.ID, eventType, message); err != nil {
		return domain.Run{}, err
	}
	return next, nil
}
