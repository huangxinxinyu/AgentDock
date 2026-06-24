package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthzReturnsServiceStatus(t *testing.T) {
	router := NewRouter(Dependencies{
		ServiceName: "agentdock-api",
		StartedAt:   time.Unix(10, 0).UTC(),
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["service"] != "agentdock-api" || body["status"] != "ok" {
		t.Fatalf("body = %#v", body)
	}
}

func TestReadyzReportsDependencyFailures(t *testing.T) {
	router := NewRouter(Dependencies{
		ServiceName: "agentdock-api",
		StartedAt:   time.Now(),
		Readiness: func(context.Context) Readiness {
			return Readiness{
				Status: "degraded",
				Dependencies: map[string]DependencyStatus{
					"postgres": {Status: "error", Message: "connection refused"},
					"redis":    {Status: "ok"},
				},
			}
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body Readiness
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Dependencies["postgres"].Status != "error" {
		t.Fatalf("postgres readiness = %#v", body.Dependencies["postgres"])
	}
}

func TestDefaultReadyzDoesNotIncludeRedis(t *testing.T) {
	router := NewRouter(Dependencies{
		ServiceName: "agentdock-api",
		StartedAt:   time.Now(),
	})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var body Readiness
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := body.Dependencies["redis"]; ok {
		t.Fatalf("default readiness includes redis dependency: %#v", body.Dependencies)
	}
}

func TestRouterSetsCORSForAllowedOrigin(t *testing.T) {
	router := NewRouter(Dependencies{
		ServiceName:       "agentdock-api",
		StartedAt:         time.Now(),
		CORSAllowedOrigin: "http://localhost:5173",
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}
