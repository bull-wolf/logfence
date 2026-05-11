package redactor

import (
	"strings"
	"unicode"
)

// TransformConfig holds configuration for log field value transformations.
type TransformConfig struct {
	// Lowercase converts string values to lowercase before further processing.
	Lowercase bool `yaml:"lowercase" json:"lowercase"`
	// TruncateAt truncates string values to the given length. 0 means no truncation.
	TruncateAt int `yaml:"truncate_at" json:"truncate_at"`
	// StripControlChars removes non-printable control characters from string values.
	StripControlChars bool `yaml:"strip_control_chars" json:"strip_control_chars"`
}

// DefaultTransformConfig returns a TransformConfig with safe defaults.
func DefaultTransformConfig() TransformConfig {
	return TransformConfig{
		Lowercase:         false,
		TruncateAt:        0,
		StripControlChars: true,
	}
}

// Transform applies the configured transformations to a string value.
func Transform(s string, cfg TransformConfig) string {
	if cfg.StripControlChars {
		s = stripControl(s)
	}
	if cfg.Lowercase {
		s = strings.ToLower(s)
	}
	if cfg.TruncateAt > 0 && len(s) > cfg.TruncateAt {
		s = s[:cfg.TruncateAt]
	}
	return s
}

// TransformFields applies Transform to all string values in a log entry map.
func TransformFields(entry map[string]any, cfg TransformConfig) map[string]any {
	result := make(map[string]any, len(entry))
	for k, v := range entry {
		switch val := v.(type) {
		case string:
			result[k] = Transform(val, cfg)
		default:
			result[k] = v
		}
	}
	return result
}

func stripControl(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\t' || r == '\n' || r == '\r' || !unicode.IsControl(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
