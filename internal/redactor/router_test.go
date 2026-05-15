package redactor

import (
	"testing"
)

func TestRouter_DefaultPipelineWhenNoRoutes(t *testing.T) {
	cfg := RouterConfig{DefaultPipeline: "default"}
	r, err := NewRouter(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := r.Route("some-container")
	if got != "default" {
		t.Errorf("expected 'default', got %q", got)
	}
}

func TestRouter_MatchesFirstRoute(t *testing.T) {
	cfg := RouterConfig{
		DefaultPipeline: "default",
		Routes: []RouteRule{
			{SourcePattern: "^nginx", Pipeline: "nginx-pipeline"},
			{SourcePattern: "^api", Pipeline: "api-pipeline"},
		},
	}
	r, err := NewRouter(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := r.Route("nginx-proxy")
	if got != "nginx-pipeline" {
		t.Errorf("expected 'nginx-pipeline', got %q", got)
	}
}

func TestRouter_SecondRouteMatches(t *testing.T) {
	cfg := RouterConfig{
		DefaultPipeline: "default",
		Routes: []RouteRule{
			{SourcePattern: "^nginx", Pipeline: "nginx-pipeline"},
			{SourcePattern: "^api", Pipeline: "api-pipeline"},
		},
	}
	r, err := NewRouter(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := r.Route("api-service")
	if got != "api-pipeline" {
		t.Errorf("expected 'api-pipeline', got %q", got)
	}
}

func TestRouter_FallsBackToDefault(t *testing.T) {
	cfg := RouterConfig{
		DefaultPipeline: "fallback",
		Routes: []RouteRule{
			{SourcePattern: "^nginx", Pipeline: "nginx-pipeline"},
		},
	}
	r, err := NewRouter(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := r.Route("unknown-service")
	if got != "fallback" {
		t.Errorf("expected 'fallback', got %q", got)
	}
}

func TestRouter_InvalidPatternReturnsError(t *testing.T) {
	cfg := RouterConfig{
		DefaultPipeline: "default",
		Routes: []RouteRule{
			{SourcePattern: "[invalid(", Pipeline: "some-pipeline"},
		},
	}
	_, err := NewRouter(cfg)
	if err == nil {
		t.Fatal("expected error for invalid regex pattern, got nil")
	}
}

func TestRouter_EmptyDefaultUsesBuiltin(t *testing.T) {
	cfg := RouterConfig{DefaultPipeline: ""}
	r, err := NewRouter(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := r.Route("anything")
	if got != "default" {
		t.Errorf("expected built-in 'default', got %q", got)
	}
}

func TestDefaultRouterConfig(t *testing.T) {
	cfg := DefaultRouterConfig()
	if cfg.DefaultPipeline != "default" {
		t.Errorf("expected default pipeline 'default', got %q", cfg.DefaultPipeline)
	}
	if len(cfg.Routes) != 0 {
		t.Errorf("expected empty routes, got %d", len(cfg.Routes))
	}
}
