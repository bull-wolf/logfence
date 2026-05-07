package redactor_test

import (
	"testing"

	"github.com/yourorg/logfence/internal/redactor"
)

func TestRedactString_Email(t *testing.T) {
	r, err := redactor.New(redactor.DefaultRules())
	if err != nil {
		t.Fatalf("failed to create redactor: %v", err)
	}

	input := "user login from alice@example.com succeeded"
	got := r.RedactString(input)
	if contains(got, "alice@example.com") {
		t.Errorf("email not redacted, got: %s", got)
	}
}

func TestRedactString_BearerToken(t *testing.T) {
	r, err := redactor.New(redactor.DefaultRules())
	if err != nil {
		t.Fatalf("failed to create redactor: %v", err)
	}

	input := "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.payload.sig"
	got := r.RedactString(input)
	if contains(got, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Errorf("bearer token not redacted, got: %s", got)
	}
}

func TestRedactFields_MixedTypes(t *testing.T) {
	r, err := redactor.New(redactor.DefaultRules())
	if err != nil {
		t.Fatalf("failed to create redactor: %v", err)
	}

	fields := map[string]interface{}{
		"message": "contact bob@corp.io for details",
		"count":   42,
		"active":  true,
	}
	result := r.RedactFields(fields)

	msg, _ := result["message"].(string)
	if contains(msg, "bob@corp.io") {
		t.Errorf("email not redacted in fields, got: %s", msg)
	}
	if result["count"] != 42 {
		t.Errorf("non-string field mutated")
	}
}

func TestNew_InvalidPattern(t *testing.T) {
	_, err := redactor.New([]redactor.RuleConfig{
		{Name: "bad", Pattern: `[invalid`},
	})
	if err == nil {
		t.Fatal("expected error for invalid pattern, got nil")
	}
}

func TestRedactString_CustomReplacement(t *testing.T) {
	r, err := redactor.New([]redactor.RuleConfig{
		{Name: "secret", Pattern: `secret-\w+`, Replacement: "***"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := r.RedactString("token=secret-abc123")
	if contains(got, "secret-abc123") {
		t.Errorf("custom pattern not redacted, got: %s", got)
	}
	if !contains(got, "***") {
		t.Errorf("expected replacement *** in output, got: %s", got)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		})())
}
