package redactor

import (
	"encoding/json"
	"strings"
)

// ScrubberConfig controls which JSON keys are fully removed from log entries.
type ScrubberConfig struct {
	// Keys is a list of exact or suffix-matched JSON field names to scrub.
	Keys []string
	// Replacement is the value substituted for scrubbed fields (default: "[SCRUBBED]").
	Replacement string
}

// DefaultScrubberConfig returns a ScrubberConfig with common sensitive keys.
func DefaultScrubberConfig() ScrubberConfig {
	return ScrubberConfig{
		Keys: []string{
			"password",
			"passwd",
			"secret",
			"api_key",
			"apikey",
			"token",
			"private_key",
			"ssn",
			"credit_card",
		},
		Replacement: "[SCRUBBED]",
	}
}

// Scrubber replaces sensitive JSON field values with a placeholder.
type Scrubber struct {
	cfg ScrubberConfig
}

// NewScrubber creates a Scrubber from the given config.
func NewScrubber(cfg ScrubberConfig) *Scrubber {
	if cfg.Replacement == "" {
		cfg.Replacement = "[SCRUBBED]"
	}
	return &Scrubber{cfg: cfg}
}

// ScrubJSON parses raw JSON, replaces matched field values, and re-encodes it.
// Returns the original input unchanged if it is not valid JSON.
func (s *Scrubber) ScrubJSON(input []byte) []byte {
	var obj map[string]interface{}
	if err := json.Unmarshal(input, &obj); err != nil {
		return input
	}
	s.scrubMap(obj)
	out, err := json.Marshal(obj)
	if err != nil {
		return input
	}
	return out
}

func (s *Scrubber) scrubMap(m map[string]interface{}) {
	for k, v := range m {
		if s.matchKey(k) {
			m[k] = s.cfg.Replacement
			continue
		}
		switch child := v.(type) {
		case map[string]interface{}:
			s.scrubMap(child)
		case []interface{}:
			s.scrubSlice(child)
		}
	}
}

func (s *Scrubber) scrubSlice(arr []interface{}) {
	for _, item := range arr {
		if child, ok := item.(map[string]interface{}); ok {
			s.scrubMap(child)
		}
	}
}

func (s *Scrubber) matchKey(key string) bool {
	lower := strings.ToLower(key)
	for _, k := range s.cfg.Keys {
		if lower == k || strings.HasSuffix(lower, "_"+k) {
			return true
		}
	}
	return false
}
