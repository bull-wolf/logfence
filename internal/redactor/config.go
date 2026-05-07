package redactor

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the full redaction configuration loaded from a file.
type Config struct {
	// UseDefaults merges the built-in rule set before any custom rules.
	UseDefaults bool         `yaml:"use_defaults" json:"use_defaults"`
	Rules       []RuleConfig `yaml:"rules"        json:"rules"`
}

// LoadConfigFile reads a YAML or JSON config file and returns a Config.
func LoadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("logfence: reading config %q: %w", path, err)
	}

	var cfg Config
	switch ext(path) {
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("logfence: parsing JSON config: %w", err)
		}
	default: // .yaml / .yml and anything else
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("logfence: parsing YAML config: %w", err)
		}
	}
	return &cfg, nil
}

// BuildRedactor constructs a Redactor from a Config, optionally prepending
// the built-in default rules when cfg.UseDefaults is true.
func BuildRedactor(cfg *Config) (*Redactor, error) {
	rules := make([]RuleConfig, 0)
	if cfg.UseDefaults {
		rules = append(rules, DefaultRules()...)
	}
	rules = append(rules, cfg.Rules...)
	return New(rules)
}

// ext returns the lowercase file extension including the leading dot.
func ext(path string) string {
	for i := len(path) - 1; i >= 0 && path[i] != '/'; i-- {
		if path[i] == '.' {
			return lowercase(path[i:])
		}
	}
	return ""
}

func lowercase(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}
