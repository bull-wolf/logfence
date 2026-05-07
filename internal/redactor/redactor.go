package redactor

import (
	"regexp"
	"strings"
)

// Rule defines a single redaction rule with a pattern and replacement.
type Rule struct {
	Name        string
	Pattern     *regexp.Regexp
	Replacement string
}

// Redactor applies a set of redaction rules to log field values.
type Redactor struct {
	rules []Rule
}

// New creates a Redactor from a slice of RuleConfig definitions.
func New(configs []RuleConfig) (*Redactor, error) {
	rules := make([]Rule, 0, len(configs))
	for _, cfg := range configs {
		re, err := regexp.Compile(cfg.Pattern)
		if err != nil {
			return nil, &InvalidPatternError{Name: cfg.Name, Err: err}
		}
		replacement := cfg.Replacement
		if replacement == "" {
			replacement = "[REDACTED]"
		}
		rules = append(rules, Rule{
			Name:        cfg.Name,
			Pattern:     re,
			Replacement: replacement,
		})
	}
	return &Redactor{rules: rules}, nil
}

// RedactString applies all rules to a plain string value.
func (r *Redactor) RedactString(value string) string {
	for _, rule := range r.rules {
		value = rule.Pattern.ReplaceAllString(value, rule.Replacement)
	}
	return value
}

// RedactFields applies redaction rules to a map of log fields.
// Only string values are inspected; other types are passed through unchanged.
func (r *Redactor) RedactFields(fields map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(fields))
	for k, v := range fields {
		switch s := v.(type) {
		case string:
			result[k] = r.RedactString(s)
		default:
			result[k] = v
		}
	}
	return result
}

// RuleConfig is the serialisable configuration for a single redaction rule.
type RuleConfig struct {
	Name        string `yaml:"name" json:"name"`
	Pattern     string `yaml:"pattern" json:"pattern"`
	Replacement string `yaml:"replacement,omitempty" json:"replacement,omitempty"`
}

// InvalidPatternError is returned when a rule pattern fails to compile.
type InvalidPatternError struct {
	Name string
	Err  error
}

func (e *InvalidPatternError) Error() string {
	return "logfence: invalid pattern for rule \"" + e.Name + "\": " + e.Err.Error()
}

// DefaultRules returns a set of built-in redaction rules for common sensitive data.
func DefaultRules() []RuleConfig {
	return []RuleConfig{
		{
			Name:        "credit-card",
			Pattern:     `\b(?:\d[ -]?){13,16}\b`,
			Replacement: "[CARD-REDACTED]",
		},
		{
			Name:        "email",
			Pattern:     `[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`,
			Replacement: "[EMAIL-REDACTED]",
		},
		{
			Name:        "bearer-token",
			Pattern:     `(?i)bearer\s+[A-Za-z0-9\-._~+/]+=*`,
			Replacement: "Bearer [TOKEN-REDACTED]",
		},
	}
	_ = strings.ToLower // suppress unused import
}
