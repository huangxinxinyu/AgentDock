package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Dependencies struct {
	ServiceName       string
	StartedAt         time.Time
	CORSAllowedOrigin string
	Readiness         func(context.Context) Readiness
}

type DependencyStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type Readiness struct {
	Status       string                      `json:"status"`
	Dependencies map[string]DependencyStatus `json:"dependencies"`
}

func NewRouter(deps Dependencies) http.Handler {
	if deps.ServiceName == "" {
		deps.ServiceName = "agentdock-api"
	}
	if deps.StartedAt.IsZero() {
		deps.StartedAt = time.Now().UTC()
	}
	if deps.Readiness == nil {
		deps.Readiness = defaultReadiness
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"service": deps.ServiceName,
			"status":  "ok",
		})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		readiness := deps.Readiness(r.Context())
		status := http.StatusOK
		if readiness.Status != "ok" {
			status = http.StatusServiceUnavailable
		}
		writeJSON(w, status, readiness)
	})

	return requestIDMiddleware(corsMiddleware(deps.CORSAllowedOrigin, mux))
}

func defaultReadiness(context.Context) Readiness {
	return Readiness{
		Status: "degraded",
		Dependencies: map[string]DependencyStatus{
			"postgres": {Status: "not_configured", Message: "readiness probe not wired in Sprint 1"},
			"redis":    {Status: "not_configured", Message: "readiness probe not wired in Sprint 1"},
		},
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
