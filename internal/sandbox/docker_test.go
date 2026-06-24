package sandbox

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeCommandRunner struct {
	calls  []string
	output string
	err    error
}

func (runner *fakeCommandRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	runner.calls = append(runner.calls, name+" "+strings.Join(args, " "))
	return runner.output, runner.err
}

func TestDockerProviderRequiresAgentOSImage(t *testing.T) {
	provider := NewDockerProvider(DockerConfig{})

	_, err := provider.CreateSession(context.Background(), CreateSessionRequest{Name: "scratch"})

	if !errors.Is(err, ErrProviderNotConfigured) {
		t.Fatalf("error = %v, want ErrProviderNotConfigured", err)
	}
}

func TestDockerProviderCreatesAgentOSContainer(t *testing.T) {
	runner := &fakeCommandRunner{output: "container-123\n"}
	provider := NewDockerProvider(DockerConfig{
		AgentOSImage:   "agentos:test",
		DefaultWorkdir: "/workspace",
		VolumePrefix:   "agentdock",
		Runner:         runner,
	})

	session, err := provider.CreateSession(context.Background(), CreateSessionRequest{Name: "Scratch"})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}

	if session.ID != "container-123" {
		t.Fatalf("session id = %q, want container-123", session.ID)
	}
	if session.Provider != "local-docker" {
		t.Fatalf("provider = %q, want local-docker", session.Provider)
	}
	if len(runner.calls) != 1 || !strings.Contains(runner.calls[0], "run -d") || !strings.Contains(runner.calls[0], "agentos:test") {
		t.Fatalf("docker calls = %#v, want docker run with image", runner.calls)
	}
}

func TestDockerProviderLifecycleCommands(t *testing.T) {
	runner := &fakeCommandRunner{output: "running\n"}
	provider := NewDockerProvider(DockerConfig{
		AgentOSImage: "agentos:test",
		Runner:       runner,
	})
	ref := SessionRef{ProviderSessionID: "container-123", State: "ready"}

	if _, err := provider.PauseSession(context.Background(), ref); err != nil {
		t.Fatalf("PauseSession returned error: %v", err)
	}
	if _, err := provider.ResumeSession(context.Background(), ref); err != nil {
		t.Fatalf("ResumeSession returned error: %v", err)
	}
	if _, err := provider.CloseSession(context.Background(), ref); err != nil {
		t.Fatalf("CloseSession returned error: %v", err)
	}
	observed, err := provider.InspectSession(context.Background(), ref)
	if err != nil {
		t.Fatalf("InspectSession returned error: %v", err)
	}

	if observed.State != "ready" {
		t.Fatalf("observed state = %q, want ready", observed.State)
	}
	joined := strings.Join(runner.calls, "\n")
	for _, want := range []string{"pause container-123", "unpause container-123", "stop container-123", "inspect"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("docker calls = %q, missing %q", joined, want)
		}
	}
}
