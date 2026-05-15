package redactor

import (
	"encoding/json"
	"strings"
	"unicode"
)

// NormalizerConfig controls how log field keys and string values are normalized.
type NormalizerConfig struct {
	// LowercaseKeys normalizes all JSON object keys to lowercase.
	LowercaseKeys bool `yaml:"lowercase_keys"`
	// TrimSpace strips leading and trailing whitespace from string values.
	TrimSpace bool `yaml:"trim_space"`
	// CollapseWhitespace replaces runs of internal whitespace with a single space.
	CollapseWhitespace bool `yaml:"collapse_whitespace"`
}

// DefaultNormalizerConfig returns a NormalizerConfig with sensible defaults.
func DefaultNormalizerConfig() NormalizerConfig {
	return NormalizerConfig{
		LowercaseKeys:      true,
		TrimSpace:          true,
		CollapseWhitespace: false,
	}
}

// Normalizer applies key and value normalization to JSON log entries.
type Normalizer struct {
	cfg NormalizerConfig
}

// NewNormalizer creates a Normalizer with the given config.
func NewNormalizer(cfg NormalizerConfig) *Normalizer {
	return &Normalizer{cfg: cfg}
}

// NormalizeJSON parses a JSON object, applies normalization, and re-encodes it.
// Non-JSON input is returned unchanged.
func (n *Normalizer) NormalizeJSON(input []byte) []byte {
	var obj map[string]interface{}
	if err := json.Unmarshal(input, &obj); err != nil {
		return input
	}
	normalized := n.normalizeMap(obj)
	out, err := json.Marshal(normalized)
	if err != nil {
		return input
	}
	return out
}

func (n *Normalizer) normalizeMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		key := k
		if n.cfg.LowercaseKeys {
			key = strings.ToLower(k)
		}
		result[key] = n.normalizeValue(v)
	}
	return result
}

func (n *Normalizer) normalizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		return n.normalizeString(val)
	case map[string]interface{}:
		return n.normalizeMap(val)
	case []interface{}:
		out := make([]interface{}, len(val))
		for i, elem := range val {
			out[i] = n.normalizeValue(elem)
		}
		return out
	}
	return v
}

func (n *Normalizer) normalizeString(s string) string {
	if n.cfg.TrimSpace {
		s = strings.TrimSpace(s)
	}
	if n.cfg.CollapseWhitespace {
		s = collapseWS(s)
	}
	return s
}

func collapseWS(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) && r != '\n' && r != '\t' {
			if !inSpace {
				b.WriteRune(' ')
				inSpace = true
			}
		} else {
			b.WriteRune(r)
			inSpace = false
		}
	}
	return b.String()
}
