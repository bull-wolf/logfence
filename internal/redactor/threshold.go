package redactor

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ThresholdAction defines what to do when a threshold is breached.
type ThresholdAction string

const (
	ThresholdActionDrop    ThresholdAction = "drop"
	ThresholdActionAnnotate ThresholdAction = "annotate"
)

// ThresholdRule defines a numeric field comparison.
type ThresholdRule struct {
	Field    string          `json:"field"`
	Operator string          `json:"operator"` // "gt", "lt", "gte", "lte", "eq"
	Value    float64         `json:"value"`
	Action   ThresholdAction `json:"action"`
	Label    string          `json:"label"`
}

// ThresholdConfig configures the Threshold processor.
type ThresholdConfig struct {
	Rules      []ThresholdRule `json:"rules"`
	LabelField string          `json:"label_field"`
}

// DefaultThresholdConfig returns a safe default configuration.
func DefaultThresholdConfig() ThresholdConfig {
	return ThresholdConfig{
		LabelField: "threshold_alert",
	}
}

// Threshold evaluates numeric fields against configured rules.
type Threshold struct {
	cfg ThresholdConfig
}

// NewThreshold creates a Threshold processor from the given config.
func NewThreshold(cfg ThresholdConfig) (*Threshold, error) {
	for _, r := range cfg.Rules {
		switch r.Operator {
		case "gt", "lt", "gte", "lte", "eq":
		default:
			return nil, fmt.Errorf("threshold: unknown operator %q", r.Operator)
		}
		if r.Field == "" {
			return nil, fmt.Errorf("threshold: rule has empty field")
		}
	}
	return &Threshold{cfg: cfg}, nil
}

// Process evaluates threshold rules against a log entry.
// Returns (entry, false) if the entry should be dropped, (entry, true) otherwise.
func (t *Threshold) Process(entry map[string]any) (map[string]any, bool) {
	if entry == nil {
		return nil, false
	}
	for _, rule := range t.cfg.Rules {
		v, ok := entry[rule.Field]
		if !ok {
			continue
		}
		num, err := toFloat64(v)
		if err != nil {
			continue
		}
		if matches(num, rule.Operator, rule.Value) {
			if rule.Action == ThresholdActionDrop {
				return entry, false
			}
			// annotate
			labelField := t.cfg.LabelField
			if labelField == "" {
				labelField = "threshold_alert"
			}
			if rule.Label != "" {
				entry[labelField] = rule.Label
			} else {
				entry[labelField] = fmt.Sprintf("%s_%s_%v", rule.Field, rule.Operator, rule.Value)
			}
		}
	}
	return entry, true
}

// ProcessBytes processes a raw JSON log line.
func (t *Threshold) ProcessBytes(data []byte) ([]byte, bool) {
	var entry map[string]any
	if err := json.Unmarshal(data, &entry); err != nil {
		return data, true
	}
	out, keep := t.Process(entry)
	if !keep {
		return nil, false
	}
	b, err := json.Marshal(out)
	if err != nil {
		return data, true
	}
	return b, true
}

func toFloat64(v any) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case json.Number:
		return val.Float64()
	case string:
		return strconv.ParseFloat(val, 64)
	}
	return 0, fmt.Errorf("not numeric")
}

func matches(actual float64, op string, threshold float64) bool {
	switch op {
	case "gt":
		return actual > threshold
	case "lt":
		return actual < threshold
	case "gte":
		return actual >= threshold
	case "lte":
		return actual <= threshold
	case "eq":
		return actual == threshold
	}
	return false
}
