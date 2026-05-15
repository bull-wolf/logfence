package redactor

import (
	"encoding/json"
	"fmt"
	"strings"
)

// LabelConfig defines a single label rule: if the field named Key contains
// a value matching any of the Patterns (substring match), the entry gets
// the given Label added under the configured label field.
type LabelConfig struct {
	Key      string   `yaml:"key"`
	Patterns []string `yaml:"patterns"`
	Label    string   `yaml:"label"`
}

// LabelerConfig holds configuration for the Labeler.
type LabelerConfig struct {
	LabelField string        `yaml:"label_field"` // destination field, default "_labels"
	Rules      []LabelConfig `yaml:"rules"`
}

// DefaultLabelerConfig returns a sensible default configuration.
func DefaultLabelerConfig() LabelerConfig {
	return LabelerConfig{
		LabelField: "_labels",
		Rules:      []LabelConfig{},
	}
}

// Labeler annotates structured log entries with derived labels based on
// field value patterns.
type Labeler struct {
	cfg LabelerConfig
}

// NewLabeler creates a Labeler from the provided config.
func NewLabeler(cfg LabelerConfig) (*Labeler, error) {
	if cfg.LabelField == "" {
		cfg.LabelField = "_labels"
	}
	for i, r := range cfg.Rules {
		if r.Key == "" {
			return nil, fmt.Errorf("labeler rule %d: key must not be empty", i)
		}
		if r.Label == "" {
			return nil, fmt.Errorf("labeler rule %d: label must not be empty", i)
		}
	}
	return &Labeler{cfg: cfg}, nil
}

// Process annotates the log entry (JSON or plain text) with matching labels.
// Plain-text entries are returned unchanged.
func (l *Labeler) Process(entry []byte) ([]byte, error) {
	if len(l.cfg.Rules) == 0 {
		return entry, nil
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(entry, &obj); err != nil {
		// Not JSON — pass through.
		return entry, nil
	}

	matched := map[string]struct{}{}
	for _, rule := range l.cfg.Rules {
		val, ok := obj[rule.Key]
		if !ok {
			continue
		}
		str, ok := val.(string)
		if !ok {
			continue
		}
		for _, pat := range rule.Patterns {
			if strings.Contains(str, pat) {
				matched[rule.Label] = struct{}{}
				break
			}
		}
	}

	if len(matched) == 0 {
		return entry, nil
	}

	existing, _ := obj[l.cfg.LabelField].([]interface{})
	for lbl := range matched {
		existing = append(existing, lbl)
	}
	obj[l.cfg.LabelField] = existing

	out, err := json.Marshal(obj)
	if err != nil {
		return entry, nil
	}
	return out, nil
}
