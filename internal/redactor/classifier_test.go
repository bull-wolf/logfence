package redactor

import (
	"encoding/json"
	"testing"
)

func TestClassifier_EmailLabelledPII(t *testing.T) {
	c, err := NewClassifier(DefaultClassifierConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	input := `{"msg":"contact user@example.com for details"}`
	out, err := c.Process([]byte(input))
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}
	var entry map[string]interface{}
	if err := json.Unmarshal(out, &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if entry["_sensitivity"] != "pii" {
		t.Errorf("expected _sensitivity=pii, got %v", entry["_sensitivity"])
	}
}

func TestClassifier_BearerTokenLabelledSecret(t *testing.T) {
	c, err := NewClassifier(DefaultClassifierConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	input := `{"auth":"Bearer eyJhbGciOiJIUzI1NiJ9.payload.sig"}`
	out, _ := c.Process([]byte(input))
	var entry map[string]interface{}
	json.Unmarshal(out, &entry)
	if entry["_sensitivity"] != "secret" {
		t.Errorf("expected secret, got %v", entry["_sensitivity"])
	}
}

func TestClassifier_NoMatchNoLabel(t *testing.T) {
	c, _ := NewClassifier(DefaultClassifierConfig())
	input := `{"msg":"all clear, nothing sensitive here"}`
	out, _ := c.Process([]byte(input))
	var entry map[string]interface{}
	json.Unmarshal(out, &entry)
	if _, ok := entry["_sensitivity"]; ok {
		t.Errorf("expected no _sensitivity label, but found one")
	}
}

func TestClassifier_NonJSONPassthrough(t *testing.T) {
	c, _ := NewClassifier(DefaultClassifierConfig())
	input := []byte("plain text log line with user@example.com")
	out, err := c.Process(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != string(input) {
		t.Errorf("expected passthrough for non-JSON, got %q", out)
	}
}

func TestClassifier_CustomLabelField(t *testing.T) {
	cfg := DefaultClassifierConfig()
	cfg.LabelField = "sensitivity_level"
	c, _ := NewClassifier(cfg)
	input := `{"token":"Bearer abc123"}`
	out, _ := c.Process([]byte(input))
	var entry map[string]interface{}
	json.Unmarshal(out, &entry)
	if _, ok := entry["sensitivity_level"]; !ok {
		t.Errorf("expected custom label field 'sensitivity_level'")
	}
}

func TestClassifier_InvalidPatternReturnsError(t *testing.T) {
	cfg := ClassifierConfig{
		Rules: []struct {
			Pattern string `yaml:"pattern" json:"pattern"`
			Label   string `yaml:"label"   json:"label"`
		}{
			{Pattern: `(?Pinvalid`, Label: "bad"},
		},
	}
	_, err := NewClassifier(cfg)
	if err == nil {
		t.Error("expected error for invalid pattern, got nil")
	}
}

func TestClassifier_NestedFieldMatch(t *testing.T) {
	c, _ := NewClassifier(DefaultClassifierConfig())
	input := `{"user":{"email":"nested@example.com"}}`
	out, _ := c.Process([]byte(input))
	var entry map[string]interface{}
	json.Unmarshal(out, &entry)
	if entry["_sensitivity"] != "pii" {
		t.Errorf("expected pii from nested field, got %v", entry["_sensitivity"])
	}
}
