package redactor

import (
	"encoding/json"
)

// Pipeline chains field filtering, sampling, and redaction together
// for processing structured log entries.
type Pipeline struct {
	filter  *FieldFilter
	sampler *Sampler
	redactor *Redactor
}

// PipelineConfig holds configuration for building a Pipeline.
type PipelineConfig struct {
	FieldFilter  FieldFilterConfig
	Sampler      SamplerConfig
	RedactorRules []RuleConfig
}

// NewPipeline constructs a Pipeline from the given config.
func NewPipeline(cfg PipelineConfig) (*Pipeline, error) {
	filter := NewFieldFilter(cfg.FieldFilter)

	sampler, err := NewSampler(cfg.Sampler)
	if err != nil {
		return nil, err
	}

	rules, err := CompileRules(cfg.RedactorRules)
	if err != nil {
		return nil, err
	}

	r := &Redactor{rules: rules}

	return &Pipeline{
		filter:   filter,
		sampler:  sampler,
		redactor: r,
	}, nil
}

// DefaultPipeline returns a Pipeline with sensible defaults.
func DefaultPipeline() (*Pipeline, error) {
	return NewPipeline(PipelineConfig{
		FieldFilter:  DefaultFieldFilterConfig(),
		Sampler:      DefaultSamplerConfig(),
		RedactorRules: DefaultRules(),
	})
}

// ProcessJSON takes a raw JSON log line, applies field filtering,
// sampling, and redaction, returning the processed JSON or nil if
// the entry should be dropped.
func (p *Pipeline) ProcessJSON(data []byte) ([]byte, error) {
	var entry map[string]interface{}
	if err := json.Unmarshal(data, &entry); err != nil {
		// Not valid JSON — redact as plain text and return.
		redacted := p.redactor.RedactString(string(data))
		return []byte(redacted), nil
	}

	// Determine sampling key (use "level" or "service" if present).
	key := samplingKeyFromEntry(entry)
	if !p.sampler.Allow(key) {
		return nil, nil
	}

	// Apply field-level redaction and filtering.
	processed := p.redactor.RedactFields(entry)
	p.filter.Filter(processed)

	out, err := json.Marshal(processed)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func samplingKeyFromEntry(entry map[string]interface{}) string {
	for _, k := range []string{"service", "level", "logger"} {
		if v, ok := entry[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return "default"
}
