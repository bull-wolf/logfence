package redactor

import (
	"encoding/json"
	"testing"
)

func TestAlerter_NoRulesPassthrough(t *testing.T) {
	a, err := NewAlerter(DefaultAlerterConfig())
	if err != nil {
		t.Fatal(err)
	}
	input := []byte(`{"msg":"hello world"}`)
	out := a.Process(input)
	if string(out) != string(input) {
		t.Errorf("expected passthrough, got %s", out)
	}
}

func TestAlerter_MatchAnnotatesJSON(t *testing.T) {
	cfg := DefaultAlerterConfig()
	cfg.Rules = []AlertRuleConfig{
		{Name: "sql-injection", Pattern: `(?i)select\s+\*`, Severity: "high"},
	}
	a, err := NewAlerter(cfg)
	if err != nil {
		t.Fatal(err)
	}
	input := []byte(`{"query":"SELECT * FROM users"}`)
	out := a.Process(input)

	var obj map[string]interface{}
	if err := json.Unmarshal(out, &obj); err != nil {
		t.Fatalf("output not valid JSON: %s", out)
	}
	if obj["_alert_severity"] != "high" {
		t.Errorf("expected severity high, got %v", obj["_alert_severity"])
	}
}

func TestAlerter_NoMatchNoLabel(t *testing.T) {
	cfg := DefaultAlerterConfig()
	cfg.Rules = []AlertRuleConfig{
		{Name: "secret", Pattern: `password=\S+`, Severity: "critical"},
	}
	a, _ := NewAlerter(cfg)
	input := []byte(`{"msg":"everything is fine"}`)
	out := a.Process(input)

	var obj map[string]interface{}
	json.Unmarshal(out, &obj)
	if _, ok := obj["_alert_severity"]; ok {
		t.Error("expected no alert label on non-matching entry")
	}
}

func TestAlerter_HighestSeverityWins(t *testing.T) {
	cfg := DefaultAlerterConfig()
	cfg.Rules = []AlertRuleConfig{
		{Name: "r1", Pattern: `foo`, Severity: "low"},
		{Name: "r2", Pattern: `foo`, Severity: "critical"},
		{Name: "r3", Pattern: `foo`, Severity: "medium"},
	}
	a, _ := NewAlerter(cfg)
	out := a.Process([]byte(`{"msg":"foo bar"}`))

	var obj map[string]interface{}
	json.Unmarshal(out, &obj)
	if obj["_alert_severity"] != "critical" {
		t.Errorf("expected critical, got %v", obj["_alert_severity"])
	}
}

func TestAlerter_OnAlertCallback(t *testing.T) {
	var events []AlertEvent
	cfg := DefaultAlerterConfig()
	cfg.Rules = []AlertRuleConfig{
		{Name: "bearer", Pattern: `Bearer\s+\S+`, Severity: "high"},
	}
	cfg.OnAlert = func(e AlertEvent) { events = append(events, e) }

	a, _ := NewAlerter(cfg)
	a.Process([]byte(`{"auth":"Bearer abc123"}`))

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Rule != "bearer" || events[0].Severity != "high" {
		t.Errorf("unexpected event: %+v", events[0])
	}
}

func TestAlerter_NonJSONPlainTextScanned(t *testing.T) {
	var fired bool
	cfg := DefaultAlerterConfig()
	cfg.Rules = []AlertRuleConfig{
		{Name: "panic", Pattern: `panic:`, Severity: "critical"},
	}
	cfg.OnAlert = func(AlertEvent) { fired = true }

	a, _ := NewAlerter(cfg)
	out := a.Process([]byte(`panic: runtime error: index out of range`))
	if !fired {
		t.Error("expected alert to fire on plain-text match")
	}
	// plain text should be returned unchanged (no JSON annotation possible)
	if string(out) != `panic: runtime error: index out of range` {
		t.Errorf("plain text mutated: %s", out)
	}
}

func TestAlerter_InvalidPatternReturnsError(t *testing.T) {
	cfg := DefaultAlerterConfig()
	cfg.Rules = []AlertRuleConfig{
		{Name: "bad", Pattern: `[invalid`, Severity: "low"},
	}
	_, err := NewAlerter(cfg)
	if err == nil {
		t.Error("expected error for invalid regex pattern")
	}
}

func TestAlerter_CustomLabelField(t *testing.T) {
	cfg := DefaultAlerterConfig()
	cfg.LabelField = "alert_level"
	cfg.Rules = []AlertRuleConfig{
		{Name: "r", Pattern: `error`, Severity: "medium"},
	}
	a, _ := NewAlerter(cfg)
	out := a.Process([]byte(`{"msg":"an error occurred"}`))

	var obj map[string]interface{}
	json.Unmarshal(out, &obj)
	if obj["alert_level"] != "medium" {
		t.Errorf("expected medium in alert_level, got %v", obj["alert_level"])
	}
}
