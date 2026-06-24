package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/huangxinxinyu/agentdock/internal/config"
	"github.com/huangxinxinyu/agentdock/internal/httpapi"
	"github.com/huangxinxinyu/agentdock/internal/runtime"
	"github.com/huangxinxinyu/agentdock/internal/sandbox"
	"github.com/huangxinxinyu/agentdock/internal/store"
	"github.com/huangxinxinyu/agentdock/internal/worker"
)

type Application struct {
	HTTPServer *http.Server
	store      *store.PostgresStore
	worker     *worker.Processor
}

func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*Application, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	logger.Info("starting agentdock api", "service", cfg.ServiceName, "config", cfg.RedactedValues())

	postgresStore, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := store.ApplyMigrations(ctx, postgresStore.DB(), "db/migrations"); err != nil {
		_ = postgresStore.Close()
		return nil, err
	}

	sandboxProvider := sandbox.NoopProvider{}
	resourceService := NewResourceService(postgresStore, sandboxProvider)
	deps := httpapi.Dependencies{
		ServiceName:       cfg.ServiceName,
		CORSAllowedOrigin: cfg.CORSAllowedOrigin,
		Readiness: func(ctx context.Context) httpapi.Readiness {
			status := httpapi.DependencyStatus{Status: "ok"}
			if err := postgresStore.DB().PingContext(ctx); err != nil {
				status = httpapi.DependencyStatus{Status: "error", Message: err.Error()}
			}
			overall := "ok"
			if status.Status != "ok" {
				overall = "degraded"
			}
			return httpapi.Readiness{
				Status: overall,
				Dependencies: map[string]httpapi.DependencyStatus{
					"postgres": status,
				},
			}
		},
		Resources: resourceService,
	}

	return &Application{
		HTTPServer: newHTTPServer(cfg, httpapi.NewRouter(deps)),
		store:      postgresStore,
		worker:     worker.NewProcessor(postgresStore, sandboxProvider, runtime.NoopRunner{}),
	}, nil
}

func NewServer(cfg config.Config, logger *slog.Logger) (*http.Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	logger.Info("starting agentdock api", "service", cfg.ServiceName, "config", cfg.RedactedValues())

	return newHTTPServer(cfg, httpapi.NewRouter(httpapi.Dependencies{ServiceName: cfg.ServiceName, CORSAllowedOrigin: cfg.CORSAllowedOrigin})), nil
}

func (app *Application) StartWorker(ctx context.Context) error {
	return app.worker.Start(ctx, 500*time.Millisecond)
}

func (app *Application) Close() error {
	return app.store.Close()
}

func newHTTPServer(cfg config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
