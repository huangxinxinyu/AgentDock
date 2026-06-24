package sandbox

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var ErrProviderNotConfigured = errors.New("sandbox provider not configured")

type CommandRunner interface {
	Run(context.Context, string, ...string) (string, error)
}

type DockerConfig struct {
	AgentOSImage   string
	DefaultWorkdir string
	Network        string
	VolumePrefix   string
	Runner         CommandRunner
}

type DockerProvider struct {
	image          string
	defaultWorkdir string
	network        string
	volumePrefix   string
	runner         CommandRunner
}

func NewDockerProvider(config DockerConfig) DockerProvider {
	defaultWorkdir := strings.TrimSpace(config.DefaultWorkdir)
	if defaultWorkdir == "" {
		defaultWorkdir = "/workspace"
	}
	volumePrefix := strings.TrimSpace(config.VolumePrefix)
	if volumePrefix == "" {
		volumePrefix = "agentdock"
	}
	runner := config.Runner
	if runner == nil {
		runner = execCommandRunner{}
	}
	return DockerProvider{
		image:          strings.TrimSpace(config.AgentOSImage),
		defaultWorkdir: defaultWorkdir,
		network:        strings.TrimSpace(config.Network),
		volumePrefix:   volumePrefix,
		runner:         runner,
	}
}

func (provider DockerProvider) CreateSession(ctx context.Context, request CreateSessionRequest) (Session, error) {
	if provider.image == "" {
		return Session{}, ErrProviderNotConfigured
	}
	workdir := strings.TrimSpace(request.DefaultWorkdir)
	if workdir == "" {
		workdir = provider.defaultWorkdir
	}
	name := "agentdock-" + sanitizeName(request.Name)
	volume := provider.volumePrefix + "-" + sanitizeName(request.Name)
	args := []string{
		"run", "-d",
		"--name", name,
		"-e", "AGENTDOCK_SANDBOX_NAME=" + request.Name,
		"-w", workdir,
		"-v", volume + ":" + workdir,
	}
	if provider.network != "" {
		args = append(args, "--network", provider.network)
	}
	args = append(args, provider.image, "sleep", "infinity")
	output, err := provider.runner.Run(ctx, "docker", args...)
	if err != nil {
		return Session{}, fmt.Errorf("docker run agentos: %w", err)
	}
	return Session{
		ID:             strings.TrimSpace(output),
		Provider:       "local-docker",
		DefaultWorkdir: workdir,
		Metadata:       fmt.Sprintf(`{"container_name":%q,"volume":%q}`, name, volume),
		State:          "ready",
	}, nil
}

func (provider DockerProvider) PauseSession(ctx context.Context, ref SessionRef) (SessionObservation, error) {
	if err := provider.runDocker(ctx, "pause", ref.ProviderSessionID); err != nil {
		return SessionObservation{}, err
	}
	return SessionObservation{State: "paused"}, nil
}

func (provider DockerProvider) ResumeSession(ctx context.Context, ref SessionRef) (SessionObservation, error) {
	if err := provider.runDocker(ctx, "unpause", ref.ProviderSessionID); err != nil {
		return SessionObservation{}, err
	}
	return SessionObservation{State: "ready"}, nil
}

func (provider DockerProvider) CloseSession(ctx context.Context, ref SessionRef) (SessionObservation, error) {
	if err := provider.runDocker(ctx, "stop", ref.ProviderSessionID); err != nil {
		return SessionObservation{}, err
	}
	return SessionObservation{State: "closed"}, nil
}

func (provider DockerProvider) InspectSession(ctx context.Context, ref SessionRef) (SessionObservation, error) {
	output, err := provider.runner.Run(ctx, "docker", "inspect", "--format", "{{.State.Status}}", ref.ProviderSessionID)
	if err != nil {
		return SessionObservation{}, fmt.Errorf("docker inspect sandbox: %w", err)
	}
	return SessionObservation{State: dockerStatusToSandboxState(strings.TrimSpace(output))}, nil
}

func (provider DockerProvider) runDocker(ctx context.Context, args ...string) error {
	if provider.image == "" {
		return ErrProviderNotConfigured
	}
	if _, err := provider.runner.Run(ctx, "docker", args...); err != nil {
		return fmt.Errorf("docker %s sandbox: %w", args[0], err)
	}
	return nil
}

func dockerStatusToSandboxState(status string) string {
	switch status {
	case "running":
		return "ready"
	case "paused":
		return "paused"
	case "created", "exited", "dead", "removing":
		return "closed"
	default:
		return "failed"
	}
}

func sanitizeName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
			continue
		}
		if r == '_' || r == ' ' || r == '.' {
			builder.WriteRune('-')
		}
	}
	name := strings.Trim(builder.String(), "-")
	if name == "" {
		return "sandbox"
	}
	return name
}

type execCommandRunner struct{}

func (execCommandRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	return string(output), err
}
