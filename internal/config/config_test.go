package config

import "testing"

func TestLoadUsesDefaultsAndEnvOverrides(t *testing.T) {
	cfg, err := Load(map[string]string{
		"AGENTDOCK_HTTP_ADDR":           ":9090",
		"AGENTDOCK_DATABASE_URL":        "postgres://agentdock:agentdock@localhost:5432/agentdock?sslmode=disable",
		"AGENTDOCK_APP_SECRET":          "test-app-secret",
		"AGENTDOCK_ENCRYPTION_KEY":      "test-encryption-key",
		"AGENTDOCK_CORS_ALLOWED_ORIGIN": "http://localhost:5173",
		"AGENTDOCK_SANDBOX_PROVIDER":    "local-docker",
		"AGENTDOCK_AGENTOS_IMAGE":       "agentos:test",
	})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.ServiceName != "agentdock-api" {
		t.Fatalf("ServiceName = %q, want agentdock-api", cfg.ServiceName)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr = %q, want :9090", cfg.HTTPAddr)
	}
	if cfg.DatabaseURL == "" {
		t.Fatalf("expected database URL to be loaded: %#v", cfg)
	}
	if cfg.CORSAllowedOrigin != "http://localhost:5173" {
		t.Fatalf("CORSAllowedOrigin = %q", cfg.CORSAllowedOrigin)
	}
	if cfg.SandboxProvider != "local-docker" {
		t.Fatalf("SandboxProvider = %q, want local-docker", cfg.SandboxProvider)
	}
	if cfg.AgentOSImage != "agentos:test" {
		t.Fatalf("AgentOSImage = %q, want agentos:test", cfg.AgentOSImage)
	}
	if cfg.AgentOSDefaultWorkdir != "/workspace" {
		t.Fatalf("AgentOSDefaultWorkdir = %q, want /workspace", cfg.AgentOSDefaultWorkdir)
	}
}

func TestLoadDoesNotRequireRedisURL(t *testing.T) {
	_, err := Load(map[string]string{
		"AGENTDOCK_DATABASE_URL":   "postgres://agentdock:agentdock@localhost:5432/agentdock?sslmode=disable",
		"AGENTDOCK_APP_SECRET":     "test-app-secret",
		"AGENTDOCK_ENCRYPTION_KEY": "test-encryption-key",
	})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
}

func TestLoadRejectsMissingRequiredSecrets(t *testing.T) {
	_, err := Load(map[string]string{})
	if err == nil {
		t.Fatal("Load succeeded without required secrets")
	}
}

func TestRedactedValuesMaskSecrets(t *testing.T) {
	cfg, err := Load(map[string]string{
		"AGENTDOCK_DATABASE_URL":   "postgres://agentdock:secret@localhost:5432/agentdock?sslmode=disable",
		"AGENTDOCK_APP_SECRET":     "test-app-secret",
		"AGENTDOCK_ENCRYPTION_KEY": "test-encryption-key",
	})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	redacted := cfg.RedactedValues()
	for key, value := range redacted {
		if value == "test-app-secret" || value == "test-encryption-key" {
			t.Fatalf("%s was not redacted", key)
		}
	}
	if redacted["AGENTDOCK_APP_SECRET"] != "[redacted]" {
		t.Fatalf("app secret redaction = %q", redacted["AGENTDOCK_APP_SECRET"])
	}
}
