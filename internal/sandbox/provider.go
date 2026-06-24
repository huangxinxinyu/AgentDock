package sandbox

import "context"

type Provider interface {
	CreateSession(context.Context, CreateSessionRequest) (Session, error)
	PauseSession(context.Context, SessionRef) (SessionObservation, error)
	ResumeSession(context.Context, SessionRef) (SessionObservation, error)
	CloseSession(context.Context, SessionRef) (SessionObservation, error)
	InspectSession(context.Context, SessionRef) (SessionObservation, error)
}

type CreateSessionRequest struct {
	Name           string
	DefaultWorkdir string
	AgentOSImage   string
}

type Session struct {
	ID             string
	Provider       string
	DefaultWorkdir string
	Metadata       string
	State          string
}

type SessionRef struct {
	ProviderSessionID string
	State             string
}

type SessionObservation struct {
	State    string
	Metadata string
}

type NoopProvider struct{}

func (NoopProvider) CreateSession(_ context.Context, request CreateSessionRequest) (Session, error) {
	defaultWorkdir := request.DefaultWorkdir
	if defaultWorkdir == "" {
		defaultWorkdir = "/workspace"
	}
	return Session{
		ID:             "noop-session",
		Provider:       "noop",
		DefaultWorkdir: defaultWorkdir,
		State:          "ready",
	}, nil
}

func (NoopProvider) PauseSession(context.Context, SessionRef) (SessionObservation, error) {
	return SessionObservation{State: "paused"}, nil
}

func (NoopProvider) ResumeSession(context.Context, SessionRef) (SessionObservation, error) {
	return SessionObservation{State: "ready"}, nil
}

func (NoopProvider) CloseSession(context.Context, SessionRef) (SessionObservation, error) {
	return SessionObservation{State: "closed"}, nil
}

func (NoopProvider) InspectSession(_ context.Context, ref SessionRef) (SessionObservation, error) {
	return SessionObservation{State: ref.State}, nil
}
