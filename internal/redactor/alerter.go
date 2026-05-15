package redactor

import (
	"encoding/json"
	"regexp"
	"strings"
)

// AlertRule defines a condition that triggers an alert when matched.
type AlertRule struct {
	Name    string
	Pattern *regexp.Regexp
	Severity string // "low", "medium", "high", "critical"
}

// AlertEvent is emitted when a log entry matches an alert rule.
type AlertEvent struct {
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Matched  string `json:"matched"`
}

// AlerterConfig holds configuration for the Alerter.
type AlerterConfig struct {
	Rules      []AlertRuleConfig
	OnAlert    func(AlertEvent)
	LabelField string // JSON field to write severity label into, e.g. "_alert"
}

// AlertRuleConfig is the serialisable form of an AlertRule.
type AlertRuleConfig struct {
	Name     string `yaml:"name"     json:"name"`
	Pattern  string `yaml:"pattern"  json:"pattern"`
	Severity string `yaml:"severity" json:"severity"`
}

// DefaultAlerterConfig returns a safe no-op configuration.
func DefaultAlerterConfig() AlerterConfig {
	return AlerterConfig{
		LabelField: "_alert_severity",
	}
}

// Alerter scans log entries for patterns and emits alert events.
type Alerter struct {
	cfg   AlerterConfig
	rules []AlertRule
}

// NewAlerter compiles alert rules and returns an Alerter.
func NewAlerter(cfg AlerterConfig) (*Alerter, error) {
	if cfg.LabelField == "" {
		cfg.LabelField = DefaultAlerterConfig().LabelField
	}
	var rules []AlertRule
	for _, rc := range cfg.Rules {
		re, err := regexp.Compile(rc.Pattern)
		if err != nil {
			return nil, err
		}
		sev := strings.ToLower(rc.Severity)
		if sev == "" {
			sev = "low"
		}
		rules = append(rules, AlertRule{Name: rc.Name, Pattern: re, Severity: sev})
	}
	return &Alerter{cfg: cfg, rules: rules}, nil
}

// Process checks the entry against all alert rules.
// Matching entries are annotated with the highest severity found and the
// OnAlert callback (if set) is invoked for each match.
// Non-JSON payloads are scanned as plain text but not annotated.
func (a *Alerter) Process(entry []byte) []byte {
	if len(a.rules) == 0 {
		return entry
	}

	var obj map[string]interface{}
	isJSON := json.Unmarshal(entry, &obj) == nil

	raw := string(entry)
	highest := ""
	for _, rule := range a.rules {
		if rule.Pattern.MatchString(raw) {
			if a.cfg.OnAlert != nil {
				a.cfg.OnAlert(AlertEvent{
					Rule:     rule.Name,
					Severity: rule.Severity,
					Matched:  rule.Pattern.FindString(raw),
				})
			}
			if severityRank(rule.Severity) > severityRank(highest) {
				highest = rule.Severity
			}
		}
	}

	if highest != "" && isJSON {
		obj[a.cfg.LabelField] = highest
		if out, err := json.Marshal(obj); err == nil {
			return out
		}
	}
	return entry
}

func severityRank(s string) int {
	switch s {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	}
	return 0
}
