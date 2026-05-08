package redactor

import "regexp"

// Rule defines a single redaction rule with a compiled pattern and replacement.
type Rule struct {
	Name        string
	Pattern     *regexp.Regexp
	Replacement string
}

// RuleConfig holds the raw configuration for a redaction rule before compilation.
type RuleConfig struct {
	Name        string `yaml:"name" json:"name"`
	Pattern     string `yaml:"pattern" json:"pattern"`
	Replacement string `yaml:"replacement" json:"replacement"`
}

// defaultRuleConfigs contains the built-in redaction rule definitions.
var defaultRuleConfigs = []RuleConfig{
	{
		Name:        "email",
		Pattern:     `[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`,
		Replacement: "[REDACTED_EMAIL]",
	},
	{
		Name:        "bearer_token",
		Pattern:     `(?i)bearer\s+[a-zA-Z0-9\-._~+/]+=*`,
		Replacement: "[REDACTED_TOKEN]",
	},
	{
		Name:        "ipv4",
		Pattern:     `\b(?:(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\b`,
		Replacement: "[REDACTED_IP]",
	},
	{
		Name:        "credit_card",
		Pattern:     `\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|6(?:011|5[0-9]{2})[0-9]{12})\b`,
		Replacement: "[REDACTED_CC]",
	},
	{
		Name:        "aws_key",
		Pattern:     `(?i)(aws_?(access_key_id|secret_access_key|session_token)[\s:=]+)[\w/+]{16,}`,
		Replacement: "[REDACTED_AWS_KEY]",
	},
}

// CompileRules compiles a slice of RuleConfig into ready-to-use Rule instances.
// Returns an error if any pattern fails to compile.
func CompileRules(configs []RuleConfig) ([]Rule, error) {
	rules := make([]Rule, 0, len(configs))
	for _, cfg := range configs {
		re, err := regexp.Compile(cfg.Pattern)
		if err != nil {
			return nil, err
		}
		replacement := cfg.Replacement
		if replacement == "" {
			replacement = "[REDACTED]"
		}
		rules = append(rules, Rule{
			Name:        cfg.Name,
			Pattern:     re,
			Replacement: replacement,
		})
	}
	return rules, nil
}
