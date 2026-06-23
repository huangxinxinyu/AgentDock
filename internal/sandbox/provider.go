package sandbox

import "context"

type Provider interface {
	CreateSession(context.Context, CreateSessionRequest) (Session, error)
}

type CreateSessionRequest struct {
	WorkspaceID string
	IssueID     string
}

type Session struct {
	ID       string
	Provider string
}

type NoopProvider struct{}

func (NoopProvider) CreateSession(context.Context, CreateSessionRequest) (Session, error) {
	return Session{
		ID:       "noop-session",
		Provider: "noop",
	}, nil
}
