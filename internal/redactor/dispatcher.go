package redactor

import (
	"encoding/json"
	"strings"
)

// DispatcherConfig controls how log entries are dispatched to named outputs.
type DispatcherConfig struct {
	// Rules maps a field value pattern to an output name.
	Rules []DispatchRule `json:"rules"`
	// DefaultOutput is used when no rule matches.
	DefaultOutput string `json:"default_output"`
	// FieldName is the JSON field to match against (e.g. "level", "service").
	FieldName string `json:"field_name"`
}

// DispatchRule associates a field value substring with an output name.
type DispatchRule struct {
	Contains string `json:"contains"`
	Output   string `json:"output"`
}

// DefaultDispatcherConfig returns a sensible default config.
func DefaultDispatcherConfig() DispatcherConfig {
	return DispatcherConfig{
		FieldName:     "level",
		DefaultOutput: "default",
		Rules: []DispatchRule{
			{Contains: "error", Output: "errors"},
			{Contains: "warn", Output: "warnings"},
		},
	}
}

// Dispatcher routes a log entry to a named output based on a field value.
type Dispatcher struct {
	cfg DispatcherConfig
}

// NewDispatcher creates a Dispatcher from the given config.
func NewDispatcher(cfg DispatcherConfig) *Dispatcher {
	return &Dispatcher{cfg: cfg}
}

// Dispatch returns the output name for the given raw log entry (JSON or plain).
// It never drops the entry; it always returns a non-empty output name.
func (d *Dispatcher) Dispatch(entry []byte) string {
	var fields map[string]interface{}
	if err := json.Unmarshal(entry, &fields); err != nil {
		return d.cfg.DefaultOutput
	}

	val, ok := fields[d.cfg.FieldName]
	if !ok {
		return d.cfg.DefaultOutput
	}

	str, ok := val.(string)
	if !ok {
		return d.cfg.DefaultOutput
	}
	lower := strings.ToLower(str)

	for _, rule := range d.cfg.Rules {
		if strings.Contains(lower, strings.ToLower(rule.Contains)) {
			return rule.Output
		}
	}
	return d.cfg.DefaultOutput
}
