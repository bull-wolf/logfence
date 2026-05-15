package redactor

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// SchemaValidatorConfig controls which fields are required and what patterns they must match.
type SchemaValidatorConfig struct {
	// RequiredFields is a list of top-level JSON keys that must be present.
	RequiredFields []string `yaml:"required_fields"`
	// FieldPatterns maps field names to regex patterns the value must satisfy.
	FieldPatterns map[string]string `yaml:"field_patterns"`
	// DropInvalid drops the log entry when validation fails instead of passing it through.
	DropInvalid bool `yaml:"drop_invalid"`
}

// DefaultSchemaValidatorConfig returns a config that requires a "level" and "message" field.
func DefaultSchemaValidatorConfig() SchemaValidatorConfig {
	return SchemaValidatorConfig{
		RequiredFields: []string{"level", "message"},
		FieldPatterns:  map[string]string{"level": `^(debug|info|warn|error|fatal)$`},
		DropInvalid:    false,
	}
}

// SchemaValidator validates JSON log entries against a required-field and pattern schema.
type SchemaValidator struct {
	cfg      SchemaValidatorConfig
	patterns map[string]*regexp.Regexp
}

// NewSchemaValidator compiles the configured field patterns and returns a SchemaValidator.
func NewSchemaValidator(cfg SchemaValidatorConfig) (*SchemaValidator, error) {
	compiled := make(map[string]*regexp.Regexp, len(cfg.FieldPatterns))
	for field, pat := range cfg.FieldPatterns {
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, fmt.Errorf("schema_validator: invalid pattern for field %q: %w", field, err)
		}
		compiled[field] = re
	}
	return &SchemaValidator{cfg: cfg, patterns: compiled}, nil
}

// Process validates the entry. If validation fails and DropInvalid is true the
// second return value is false (drop). Otherwise the original payload is returned.
func (v *SchemaValidator) Process(data []byte) ([]byte, bool) {
	var entry map[string]interface{}
	if err := json.Unmarshal(data, &entry); err != nil {
		// Not JSON – pass through unchanged.
		return data, true
	}

	for _, field := range v.cfg.RequiredFields {
		if _, ok := entry[field]; !ok {
			if v.cfg.DropInvalid {
				return nil, false
			}
			return data, true
		}
	}

	for field, re := range v.patterns {
		val, ok := entry[field]
		if !ok {
			continue
		}
		str, ok := val.(string)
		if !ok || !re.MatchString(str) {
			if v.cfg.DropInvalid {
				return nil, false
			}
			return data, true
		}
	}

	return data, true
}
