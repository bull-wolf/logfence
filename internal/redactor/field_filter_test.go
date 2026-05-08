package redactor

import (
	"testing"
)

func TestFieldFilter_ExactMatch(t *testing.T) {
	filter := NewFieldFilter(DefaultFieldFilterConfig())

	cases := []struct {
		field    string
		expected bool
	}{
		{"password", true},
		{"PASSWORD", true},
		{"token", true},
		{"api_key", true},
		{"authorization", true},
		{"username", false},
		{"email", false},
		{"message", false},
	}

	for _, tc := range cases {
		t.Run(tc.field, func(t *testing.T) {
			got := filter.ShouldRedact(tc.field)
			if got != tc.expected {
				t.Errorf("ShouldRedact(%q) = %v, want %v", tc.field, got, tc.expected)
			}
		})
	}
}

func TestFieldFilter_SuffixMatch(t *testing.T) {
	filter := NewFieldFilter(DefaultFieldFilterConfig())

	cases := []struct {
		field    string
		expected bool
	}{
		{"db_password", true},
		{"github_token", true},
		{"aws_secret", true},
		{"encryption_key", true},
		{"user_id", false},
		{"created_at", false},
	}

	for _, tc := range cases {
		t.Run(tc.field, func(t *testing.T) {
			got := filter.ShouldRedact(tc.field)
			if got != tc.expected {
				t.Errorf("ShouldRedact(%q) = %v, want %v", tc.field, got, tc.expected)
			}
		})
	}
}

func TestFieldFilter_CustomConfig(t *testing.T) {
	filter := NewFieldFilter(FieldFilterConfig{
		ExactNames: []string{"pin", "otp"},
		Suffixes:   []string{"_code"},
	})

	if !filter.ShouldRedact("pin") {
		t.Error("expected 'pin' to be redacted")
	}
	if !filter.ShouldRedact("invite_code") {
		t.Error("expected 'invite_code' to be redacted")
	}
	if filter.ShouldRedact("password") {
		t.Error("expected 'password' NOT to be redacted with custom config")
	}
}

func TestNewFieldFilter_EmptyConfig(t *testing.T) {
	filter := NewFieldFilter(FieldFilterConfig{})
	if filter.ShouldRedact("password") {
		t.Error("expected no redaction with empty config")
	}
}
