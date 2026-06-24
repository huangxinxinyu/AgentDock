package config

import (
	"errors"
	"fmt"
	"strings"
)

const (
	envHTTPAddr          = "AGENTDOCK_HTTP_ADDR"
	envDatabaseURL       = "AGENTDOCK_DATABASE_URL"
	envAppSecret         = "AGENTDOCK_APP_SECRET"
	envEncryptionKey     = "AGENTDOCK_ENCRYPTION_KEY"
	envCORSAllowedOrigin = "AGENTDOCK_CORS_ALLOWED_ORIGIN"
)

type Config struct {
	ServiceName       string
	HTTPAddr          string
	DatabaseURL       string
	AppSecret         string
	EncryptionKey     string
	CORSAllowedOrigin string
}

func Load(env map[string]string) (Config, error) {
	cfg := Config{
		ServiceName:       "agentdock-api",
		HTTPAddr:          valueOrDefault(env, envHTTPAddr, ":8080"),
		DatabaseURL:       strings.TrimSpace(env[envDatabaseURL]),
		AppSecret:         strings.TrimSpace(env[envAppSecret]),
		EncryptionKey:     strings.TrimSpace(env[envEncryptionKey]),
		CORSAllowedOrigin: valueOrDefault(env, envCORSAllowedOrigin, "http://localhost:5173"),
	}

	var missing []string
	for _, item := range []struct {
		key   string
		value string
	}{
		{envDatabaseURL, cfg.DatabaseURL},
		{envAppSecret, cfg.AppSecret},
		{envEncryptionKey, cfg.EncryptionKey},
	} {
		if item.value == "" {
			missing = append(missing, item.key)
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func (cfg Config) Validate() error {
	if cfg.ServiceName == "" {
		return errors.New("service name is required")
	}
	if cfg.HTTPAddr == "" {
		return errors.New("http address is required")
	}
	return nil
}

func (cfg Config) RedactedValues() map[string]string {
	return map[string]string{
		envHTTPAddr:          cfg.HTTPAddr,
		envDatabaseURL:       "[redacted]",
		envAppSecret:         "[redacted]",
		envEncryptionKey:     "[redacted]",
		envCORSAllowedOrigin: cfg.CORSAllowedOrigin,
	}
}

func valueOrDefault(env map[string]string, key string, fallback string) string {
	value := strings.TrimSpace(env[key])
	if value == "" {
		return fallback
	}
	return value
}
