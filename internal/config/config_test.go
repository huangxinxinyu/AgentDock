package config

import "testing"

func TestLoadUsesDefaultsAndEnvOverrides(t *testing.T) {
	cfg, err := Load(map[string]string{
		"AGENTDOCK_HTTP_ADDR":           ":9090",
		"AGENTDOCK_DATABASE_URL":        "postgres://agentdock:agentdock@localhost:5432/agentdock?sslmode=disable",
		"AGENTDOCK_REDIS_URL":           "redis://localhost:6379/1",
		"AGENTDOCK_APP_SECRET":          "test-app-secret",
		"AGENTDOCK_ENCRYPTION_KEY":      "test-encryption-key",
		"AGENTDOCK_CORS_ALLOWED_ORIGIN": "http://localhost:5173",
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
	if cfg.DatabaseURL == "" || cfg.RedisURL == "" {
		t.Fatalf("expected database and redis URLs to be loaded: %#v", cfg)
	}
	if cfg.CORSAllowedOrigin != "http://localhost:5173" {
		t.Fatalf("CORSAllowedOrigin = %q", cfg.CORSAllowedOrigin)
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
		"AGENTDOCK_REDIS_URL":      "redis://:secret@localhost:6379/0",
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
