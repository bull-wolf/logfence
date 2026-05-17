package redactor

import (
	"encoding/json"
	"testing"
)

func newTestThreshold(t *testing.T, rules []ThresholdRule) *Threshold {
	t.Helper()
	cfg := DefaultThresholdConfig()
	cfg.Rules = rules
	th, err := NewThreshold(cfg)
	if err != nil {
		t.Fatalf("NewThreshold: %v", err)
	}
	return th
}

func TestThreshold_AnnotatesWhenMatched(t *testing.T) {
	th := newTestThreshold(t, []ThresholdRule{
		{Field: "latency_ms", Operator: "gt", Value: 500, Action: ThresholdActionAnnotate, Label: "high_latency"},
	})
	entry := map[string]any{"latency_ms": float64(600), "msg": "slow request"}
	out, keep := th.Process(entry)
	if !keep {
		t.Fatal("expected entry to be kept")
	}
	if out["threshold_alert"] != "high_latency" {
		t.Errorf("expected label 'high_latency', got %v", out["threshold_alert"])
	}
}

func TestThreshold_NoMatchNoAnnotation(t *testing.T) {
	th := newTestThreshold(t, []ThresholdRule{
		{Field: "latency_ms", Operator: "gt", Value: 500, Action: ThresholdActionAnnotate, Label: "high_latency"},
	})
	entry := map[string]any{"latency_ms": float64(100)}
	out, keep := th.Process(entry)
	if !keep {
		t.Fatal("expected entry to be kept")
	}
	if _, ok := out["threshold_alert"]; ok {
		t.Error("expected no label on non-matching entry")
	}
}

func TestThreshold_DropAction(t *testing.T) {
	th := newTestThreshold(t, []ThresholdRule{
		{Field: "error_count", Operator: "gte", Value: 10, Action: ThresholdActionDrop},
	})
	entry := map[string]any{"error_count": float64(10)}
	_, keep := th.Process(entry)
	if keep {
		t.Fatal("expected entry to be dropped")
	}
}

func TestThreshold_MissingFieldSkipped(t *testing.T) {
	th := newTestThreshold(t, []ThresholdRule{
		{Field: "score", Operator: "lt", Value: 0, Action: ThresholdActionDrop},
	})
	entry := map[string]any{"msg": "no score here"}
	_, keep := th.Process(entry)
	if !keep {
		t.Fatal("expected entry without field to pass through")
	}
}

func TestThreshold_ProcessBytes_JSONAnnotated(t *testing.T) {
	th := newTestThreshold(t, []ThresholdRule{
		{Field: "cpu", Operator: "gte", Value: 90, Action: ThresholdActionAnnotate, Label: "cpu_high"},
	})
	input := []byte(`{"cpu":95,"host":"web-1"}`)
	out, keep := th.ProcessBytes(input)
	if !keep {
		t.Fatal("expected entry to be kept")
	}
	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result["threshold_alert"] != "cpu_high" {
		t.Errorf("expected cpu_high label, got %v", result["threshold_alert"])
	}
}

func TestThreshold_ProcessBytes_NonJSONPassthrough(t *testing.T) {
	th := newTestThreshold(t, []ThresholdRule{
		{Field: "cpu", Operator: "gt", Value: 50, Action: ThresholdActionDrop},
	})
	input := []byte("plain text log line")
	out, keep := th.ProcessBytes(input)
	if !keep {
		t.Fatal("expected plain text to pass through")
	}
	if string(out) != string(input) {
		t.Errorf("expected unchanged output, got %q", out)
	}
}

func TestNewThreshold_InvalidOperator(t *testing.T) {
	cfg := DefaultThresholdConfig()
	cfg.Rules = []ThresholdRule{
		{Field: "score", Operator: "between", Value: 10, Action: ThresholdActionAnnotate},
	}
	_, err := NewThreshold(cfg)
	if err == nil {
		t.Fatal("expected error for invalid operator")
	}
}

func TestThreshold_NilEntryReturnsFalse(t *testing.T) {
	th := newTestThreshold(t, nil)
	_, keep := th.Process(nil)
	if keep {
		t.Fatal("expected nil entry to return false")
	}
}
