package redactor

import (
	"strings"
)

// FieldFilter determines which JSON fields should be fully redacted
// regardless of their value, based on field name matching.
type FieldFilter struct {
	exactNames map[string]struct{}
	suffixes   []string
}

// FieldFilterConfig holds configuration for field-based redaction.
type FieldFilterConfig struct {
	// ExactNames are field names that will always be redacted.
	ExactNames []string `yaml:"exact_names"`
	// Suffixes are field name suffixes that trigger redaction (e.g. "_secret", "_token").
	Suffixes []string `yaml:"suffixes"`
}

// DefaultFieldFilterConfig returns a sensible default field filter config.
func DefaultFieldFilterConfig() FieldFilterConfig {
	return FieldFilterConfig{
		ExactNames: []string{
			"password", "passwd", "secret", "api_key", "apikey",
			"token", "access_token", "refresh_token", "private_key",
			"authorization", "credit_card", "ssn",
		},
		Suffixes: []string{"_secret", "_token", "_key", "_password"},
	}
}

// NewFieldFilter constructs a FieldFilter from config.
func NewFieldFilter(cfg FieldFilterConfig) *FieldFilter {
	exact := make(map[string]struct{}, len(cfg.ExactNames))
	for _, name := range cfg.ExactNames {
		exact[strings.ToLower(name)] = struct{}{}
	}
	return &FieldFilter{
		exactNames: exact,
		suffixes:   cfg.Suffixes,
	}
}

// ShouldRedact returns true if the given field name should be fully redacted.
func (f *FieldFilter) ShouldRedact(fieldName string) bool {
	lower := strings.ToLower(fieldName)
	if _, ok := f.exactNames[lower]; ok {
		return true
	}
	for _, suffix := range f.suffixes {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}
	return false
}
