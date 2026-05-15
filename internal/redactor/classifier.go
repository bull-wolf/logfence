package redactor

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ClassifierRule maps a regex pattern to a sensitivity label.
type ClassifierRule struct {
	Pattern   *regexp.Regexp
	Label     string
}

// ClassifierConfig holds configuration for the Classifier.
type ClassifierConfig struct {
	// Rules is an ordered list of pattern→label mappings applied to field values.
	Rules []struct {
		Pattern string `yaml:"pattern" json:"pattern"`
		Label   string `yaml:"label"   json:"label"`
	} `yaml:"rules" json:"rules"`
	// LabelField is the JSON key written into the log entry (default: "_sensitivity").
	LabelField string `yaml:"label_field" json:"label_field"`
}

// DefaultClassifierConfig returns a config that tags PII and secrets.
func DefaultClassifierConfig() ClassifierConfig {
	return ClassifierConfig{
		LabelField: "_sensitivity",
		Rules: []struct {
			Pattern string `yaml:"pattern" json:"pattern"`
			Label   string `yaml:"label"   json:"label"`
		}{
			{Pattern: `[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`, Label: "pii"},
			{Pattern: `(?i)bearer\s+[a-zA-Z0-9\-._~+/]+=*`, Label: "secret"},
			{Pattern: `\b\d{4}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`, Label: "pci"},
		},
	}
}

// Classifier scans log entry field values and annotates the entry with a
// sensitivity label when a pattern matches.
type Classifier struct {
	rules      []ClassifierRule
	labelField string
}

// NewClassifier compiles the given config into a Classifier.
func NewClassifier(cfg ClassifierConfig) (*Classifier, error) {
	if cfg.LabelField == "" {
		cfg.LabelField = "_sensitivity"
	}
	rules := make([]ClassifierRule, 0, len(cfg.Rules))
	for _, r := range cfg.Rules {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, err
		}
		rules = append(rules, ClassifierRule{Pattern: re, Label: r.Label})
	}
	return &Classifier{rules: rules, labelField: cfg.LabelField}, nil
}

// Process annotates a JSON log line with a sensitivity label if any field
// value matches a configured pattern. Non-JSON input is returned unchanged.
func (c *Classifier) Process(line []byte) ([]byte, error) {
	var entry map[string]interface{}
	if err := json.Unmarshal(line, &entry); err != nil {
		return line, nil
	}

	label := c.classify(entry)
	if label == "" {
		return line, nil
	}
	entry[c.labelField] = label
	out, err := json.Marshal(entry)
	if err != nil {
		return line, nil
	}
	return out, nil
}

// classify returns the first matching label found across all string values.
func (c *Classifier) classify(entry map[string]interface{}) string {
	for _, v := range entry {
		s := ""
		switch val := v.(type) {
		case string:
			s = val
		case map[string]interface{}:
			if lbl := c.classify(val); lbl != "" {
				return lbl
			}
			continue
		default:
			continue
		}
		for _, rule := range c.rules {
			if rule.Pattern.MatchString(strings.TrimSpace(s)) {
				return rule.Label
			}
		}
	}
	return ""
}
