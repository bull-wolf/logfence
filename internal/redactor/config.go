package redactor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the full redactor configuration loaded from a YAML file.
type Config struct {
	Rules       []RuleConfig      `yaml:"rules"`
	FieldFilter FieldFilterConfig `yaml:"field_filter"`
}

// RuleConfig is a single regex-based redaction rule.
type RuleConfig struct {
	Name        string `yaml:"name"`
	Pattern     string `yaml:"pattern"`
	Replacement string `yaml:"replacement"`
}

// LoadConfigFile reads and parses a YAML config file into a Config struct.
func LoadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}
	return &cfg, nil
}

// BuildRedactor constructs a Redactor and FieldFilter from a Config.
// If cfg is nil, defaults are used.
func BuildRedactor(cfg *Config) (*Redactor, *FieldFilter, error) {
	var ruleCfgs []RuleConfig
	var ffCfg FieldFilterConfig

	if cfg != nil {
		ruleCfgs = cfg.Rules
		ffCfg = cfg.FieldFilter
	} else {
		ffCfg = DefaultFieldFilterConfig()
	}

	if len(ruleCfgs) == 0 {
		ruleCfgs = toRuleConfigs(DefaultRules)
	}

	compiled, err := CompileRules(toRuleConfigsFromInternal(ruleCfgs))
	if err != nil {
		return nil, nil, fmt.Errorf("compiling rules: %w", err)
	}

	r, err := New(compiled)
	if err != nil {
		return nil, nil, fmt.Errorf("creating redactor: %w", err)
	}

	return r, NewFieldFilter(ffCfg), nil
}

func ext(path string) string {
	return strings.ToLower(filepath.Ext(path))
}

func lowercase(s string) string {
	return strings.ToLower(s)
}

func toRuleConfigs(rules []DefaultRule) []RuleConfig {
	out := make([]RuleConfig, len(rules))
	for i, r := range rules {
		out[i] = RuleConfig{Name: r.Name, Pattern: r.Pattern, Replacement: r.Replacement}
	}
	return out
}

func toRuleConfigsFromInternal(cfgs []RuleConfig) []RuleConfig {
	return cfgs
}
