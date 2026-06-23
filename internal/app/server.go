package app

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/huangxinxinyu/agentdock/internal/config"
	"github.com/huangxinxinyu/agentdock/internal/httpapi"
)

func NewServer(cfg config.Config, logger *slog.Logger) (*http.Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	logger.Info("starting agentdock api", "service", cfg.ServiceName, "config", cfg.RedactedValues())

	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewRouter(httpapi.Dependencies{ServiceName: cfg.ServiceName, CORSAllowedOrigin: cfg.CORSAllowedOrigin}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}, nil
}
