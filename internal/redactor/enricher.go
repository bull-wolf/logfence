package redactor

import (
	"encoding/json"
	"time"
)

// EnricherConfig controls which metadata fields are injected into log entries.
type EnricherConfig struct {
	// AddTimestamp injects a "_logfence_ts" field with the current UTC time.
	AddTimestamp bool `yaml:"add_timestamp"`
	// AddHostname injects a "_logfence_host" field.
	AddHostname string `yaml:"hostname"`
	// AddService injects a "_logfence_service" field.
	AddService string `yaml:"service"`
	// TimestampField overrides the default timestamp key name.
	TimestampField string `yaml:"timestamp_field"`
}

// DefaultEnricherConfig returns a no-op enricher configuration.
func DefaultEnricherConfig() EnricherConfig {
	return EnricherConfig{
		AddTimestamp:   false,
		TimestampField: "_logfence_ts",
	}
}

// Enricher injects metadata fields into JSON log entries.
type Enricher struct {
	cfg EnricherConfig
}

// NewEnricher creates an Enricher from the provided config.
func NewEnricher(cfg EnricherConfig) *Enricher {
	if cfg.TimestampField == "" {
		cfg.TimestampField = "_logfence_ts"
	}
	return &Enricher{cfg: cfg}
}

// Enrich attempts to parse payload as JSON and inject configured metadata
// fields. Non-JSON payloads are returned unchanged.
func (e *Enricher) Enrich(payload []byte) []byte {
	var entry map[string]interface{}
	if err := json.Unmarshal(payload, &entry); err != nil {
		return payload
	}

	if e.cfg.AddTimestamp {
		entry[e.cfg.TimestampField] = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if e.cfg.AddHostname != "" {
		entry["_logfence_host"] = e.cfg.AddHostname
	}
	if e.cfg.AddService != "" {
		entry["_logfence_service"] = e.cfg.AddService
	}

	out, err := json.Marshal(entry)
	if err != nil {
		return payload
	}
	return out
}
