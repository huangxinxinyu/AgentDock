package runtime

import "context"

type Runner interface {
	Start(context.Context, StartRequest) (Result, error)
}

type StartRequest struct {
	RunID  string
	Prompt string
}

type Result struct {
	Runtime string
	Summary string
}

type NoopRunner struct{}

func (NoopRunner) Start(context.Context, StartRequest) (Result, error) {
	return Result{
		Runtime: "noop",
		Summary: "runtime execution is not implemented in Sprint 1",
	}, nil
}
