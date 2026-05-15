package redactor

import (
	"encoding/json"
	"strings"
)

// TaggerConfig defines rules for appending tags to log entries.
type TaggerConfig struct {
	// TagField is the JSON field name to write tags into (default: "tags").
	TagField string
	// Rules maps a source field value substring to a list of tags to add.
	Rules []TaggerRule
}

// TaggerRule matches a field value and applies tags.
type TaggerRule struct {
	Field    string
	Contains string
	Tags     []string
}

// DefaultTaggerConfig returns a TaggerConfig with sensible defaults.
func DefaultTaggerConfig() TaggerConfig {
	return TaggerConfig{
		TagField: "tags",
		Rules:    []TaggerRule{},
	}
}

// Tagger appends structured tags to JSON log entries based on field content.
type Tagger struct {
	cfg TaggerConfig
}

// NewTagger creates a Tagger from the provided config.
func NewTagger(cfg TaggerConfig) (*Tagger, error) {
	if cfg.TagField == "" {
		cfg.TagField = "tags"
	}
	return &Tagger{cfg: cfg}, nil
}

// Process evaluates each TaggerRule against the entry and appends matching tags.
// Non-JSON input is returned unchanged.
func (t *Tagger) Process(entry []byte) ([]byte, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal(entry, &obj); err != nil {
		return entry, nil
	}

	var collected []string
	for _, rule := range t.cfg.Rules {
		val, ok := obj[rule.Field]
		if !ok {
			continue
		}
		str, ok := val.(string)
		if !ok {
			continue
		}
		if strings.Contains(str, rule.Contains) {
			collected = append(collected, rule.Tags...)
		}
	}

	if len(collected) == 0 {
		return entry, nil
	}

	// Merge with existing tags if present.
	existing, _ := obj[t.cfg.TagField].([]interface{})
	for _, tag := range collected {
		existing = append(existing, tag)
	}
	obj[t.cfg.TagField] = existing

	out, err := json.Marshal(obj)
	if err != nil {
		return entry, err
	}
	return out, nil
}
