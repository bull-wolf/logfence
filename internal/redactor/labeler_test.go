package redactor

import (
	"encoding/json"
	"testing"
)

func TestLabeler_NonJSONPassthrough(t *testing.T) {
	l, _ := NewLabeler(DefaultLabelerConfig())
	input := []byte("plain text log line")
	out, err := l.Process(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != string(input) {
		t.Errorf("expected passthrough, got %q", out)
	}
}

func TestLabeler_NoRulesNoChange(t *testing.T) {
	l, _ := NewLabeler(DefaultLabelerConfig())
	input := []byte(`{"level":"error","msg":"boom"}`)
	out, err := l.Process(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != string(input) {
		t.Errorf("expected no change, got %q", out)
	}
}

func TestLabeler_MatchAddsLabel(t *testing.T) {
	cfg := LabelerConfig{
		LabelField: "_labels",
		Rules: []LabelConfig{
			{Key: "msg", Patterns: []string{"error", "fail"}, Label: "alert"},
		},
	}
	l, err := NewLabeler(cfg)
	if err != nil {
		t.Fatalf("NewLabeler: %v", err)
	}
	input := []byte(`{"msg":"disk error occurred"}`)
	out, err := l.Process(input)
	if err != nil {
		t.Fatalf("Process: %v", err)
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(out, &obj); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	labels, ok := obj["_labels"].([]interface{})
	if !ok || len(labels) == 0 {
		t.Fatalf("expected _labels field with entries, got %v", obj["_labels"])
	}
	if labels[0] != "alert" {
		t.Errorf("expected label 'alert', got %v", labels[0])
	}
}

func TestLabeler_NoMatchNoLabel(t *testing.T) {
	cfg := LabelerConfig{
		LabelField: "_labels",
		Rules: []LabelConfig{
			{Key: "msg", Patterns: []string{"panic"}, Label: "critical"},
		},
	}
	l, _ := NewLabeler(cfg)
	input := []byte(`{"msg":"all systems nominal"}`)
	out, _ := l.Process(input)
	var obj map[string]interface{}
	json.Unmarshal(out, &obj)
	if _, exists := obj["_labels"]; exists {
		t.Errorf("expected no _labels field, but found one")
	}
}

func TestLabeler_MultipleRulesMultipleLabels(t *testing.T) {
	cfg := LabelerConfig{
		LabelField: "_labels",
		Rules: []LabelConfig{
			{Key: "level", Patterns: []string{"error"}, Label: "alert"},
			{Key: "service", Patterns: []string{"payments"}, Label: "pci"},
		},
	}
	l, _ := NewLabeler(cfg)
	input := []byte(`{"level":"error","service":"payments-api"}`)
	out, _ := l.Process(input)
	var obj map[string]interface{}
	json.Unmarshal(out, &obj)
	labels, _ := obj["_labels"].([]interface{})
	if len(labels) != 2 {
		t.Errorf("expected 2 labels, got %d: %v", len(labels), labels)
	}
}

func TestNewLabeler_InvalidRule(t *testing.T) {
	cfg := LabelerConfig{
		Rules: []LabelConfig{
			{Key: "", Patterns: []string{"x"}, Label: "lbl"},
		},
	}
	_, err := NewLabeler(cfg)
	if err == nil {
		t.Error("expected error for empty key, got nil")
	}
}
