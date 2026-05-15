package redactor

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEnricher_NonJSONPassthrough(t *testing.T) {
	e := NewEnricher(DefaultEnricherConfig())
	input := []byte("plain text log line")
	out := e.Enrich(input)
	if string(out) != string(input) {
		t.Fatalf("expected passthrough, got %q", out)
	}
}

func TestEnricher_NoOpConfig(t *testing.T) {
	e := NewEnricher(DefaultEnricherConfig())
	input := []byte(`{"msg":"hello"}`)
	out := e.Enrich(input)

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if _, ok := result["_logfence_ts"]; ok {
		t.Error("timestamp should not be present with default config")
	}
	if _, ok := result["_logfence_host"]; ok {
		t.Error("host should not be present with default config")
	}
}

func TestEnricher_AddTimestamp(t *testing.T) {
	cfg := DefaultEnricherConfig()
	cfg.AddTimestamp = true
	e := NewEnricher(cfg)

	out := e.Enrich([]byte(`{"level":"info"}`))

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	ts, ok := result["_logfence_ts"].(string)
	if !ok || ts == "" {
		t.Error("expected non-empty _logfence_ts field")
	}
}

func TestEnricher_CustomTimestampField(t *testing.T) {
	cfg := EnricherConfig{AddTimestamp: true, TimestampField: "ingested_at"}
	e := NewEnricher(cfg)

	out := e.Enrich([]byte(`{"x":1}`))
	var result map[string]interface{}
	json.Unmarshal(out, &result)

	if _, ok := result["ingested_at"]; !ok {
		t.Error("expected ingested_at field")
	}
}

func TestEnricher_AddHostnameAndService(t *testing.T) {
	cfg := EnricherConfig{
		AddHostname:    "node-1",
		AddService:     "auth-svc",
		TimestampField: "_logfence_ts",
	}
	e := NewEnricher(cfg)

	out := e.Enrich([]byte(`{"msg":"login"}`))
	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["_logfence_host"] != "node-1" {
		t.Errorf("expected _logfence_host=node-1, got %v", result["_logfence_host"])
	}
	if result["_logfence_service"] != "auth-svc" {
		t.Errorf("expected _logfence_service=auth-svc, got %v", result["_logfence_service"])
	}
}

func TestEnricher_PreservesExistingFields(t *testing.T) {
	cfg := EnricherConfig{AddHostname: "host-a", TimestampField: "_logfence_ts"}
	e := NewEnricher(cfg)

	out := e.Enrich([]byte(`{"level":"warn","msg":"disk full"}`))
	if !strings.Contains(string(out), `"level"`) {
		t.Error("original fields should be preserved")
	}
	if !strings.Contains(string(out), `"msg"`) {
		t.Error("original msg field should be preserved")
	}
}
