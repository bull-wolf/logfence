package redactor

import (
	"encoding/json"
	"unicode/utf8"
)

// TruncatorConfig controls how large field values are truncated.
type TruncatorConfig struct {
	// MaxBytes is the maximum byte length of any string value. 0 means no limit.
	MaxBytes int `yaml:"max_bytes"`
	// Suffix is appended when a value is truncated.
	Suffix string `yaml:"suffix"`
	// Fields lists specific JSON field names to truncate; empty means all fields.
	Fields []string `yaml:"fields"`
}

// DefaultTruncatorConfig returns sensible defaults.
func DefaultTruncatorConfig() TruncatorConfig {
	return TruncatorConfig{
		MaxBytes: 512,
		Suffix:   "...[truncated]",
	}
}

// Truncator truncates oversized string values in JSON log entries.
type Truncator struct {
	cfg    TruncatorConfig
	fields map[string]struct{}
}

// NewTruncator creates a Truncator from the given config.
func NewTruncator(cfg TruncatorConfig) *Truncator {
	f := make(map[string]struct{}, len(cfg.Fields))
	for _, name := range cfg.Fields {
		f[name] = struct{}{}
	}
	return &Truncator{cfg: cfg, fields: f}
}

// Apply truncates string values in the JSON object and returns the modified JSON.
// Non-JSON input is returned unchanged.
func (t *Truncator) Apply(data []byte) []byte {
	if t.cfg.MaxBytes <= 0 {
		return data
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return data
	}

	changed := false
	for k, v := range obj {
		if len(t.fields) > 0 {
			if _, ok := t.fields[k]; !ok {
				continue
			}
		}
		str, ok := v.(string)
		if !ok {
			continue
		}
		if len(str) > t.cfg.MaxBytes {
			truncated := safeTruncate(str, t.cfg.MaxBytes)
			obj[k] = truncated + t.cfg.Suffix
			changed = true
		}
	}

	if !changed {
		return data
	}

	out, err := json.Marshal(obj)
	if err != nil {
		return data
	}
	return out
}

// safeTruncate trims s to at most maxBytes without splitting a UTF-8 rune.
func safeTruncate(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	// Walk back until we find a valid rune boundary.
	for maxBytes > 0 && !utf8.RuneStart(s[maxBytes]) {
		maxBytes--
	}
	return s[:maxBytes]
}
