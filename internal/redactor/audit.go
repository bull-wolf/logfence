package redactor

import (
	"encoding/json"
	"io"
	"sync/atomic"
	"time"
)

// AuditEvent records a single redaction or sampling decision.
type AuditEvent struct {
	Timestamp  time.Time `json:"timestamp"`
	EventType  string    `json:"event_type"` // "redaction" | "sampled_drop" | "field_drop"
	Field      string    `json:"field,omitempty"`
	RuleName   string    `json:"rule,omitempty"`
	SamplingKey string   `json:"sampling_key,omitempty"`
}

// AuditLog collects audit events and writes them to a writer.
type AuditLog struct {
	writer  io.Writer
	encoder *json.Encoder

	redactions  atomic.Int64
	drops       atomic.Int64
	fieldDrops  atomic.Int64
}

// NewAuditLog creates an AuditLog that writes JSON events to w.
// Pass io.Discard to disable output while still collecting counters.
func NewAuditLog(w io.Writer) *AuditLog {
	return &AuditLog{
		writer:  w,
		encoder: json.NewEncoder(w),
	}
}

// RecordRedaction logs a pattern-based redaction event.
func (a *AuditLog) RecordRedaction(field, rule string) {
	a.redactions.Add(1)
	_ = a.encoder.Encode(AuditEvent{
		Timestamp: time.Now().UTC(),
		EventType: "redaction",
		Field:     field,
		RuleName:  rule,
	})
}

// RecordSampledDrop logs that a log entry was dropped by the sampler.
func (a *AuditLog) RecordSampledDrop(key string) {
	a.drops.Add(1)
	_ = a.encoder.Encode(AuditEvent{
		Timestamp:   time.Now().UTC(),
		EventType:   "sampled_drop",
		SamplingKey: key,
	})
}

// RecordFieldDrop logs that a field was removed by the field filter.
func (a *AuditLog) RecordFieldDrop(field string) {
	a.fieldDrops.Add(1)
	_ = a.encoder.Encode(AuditEvent{
		Timestamp: time.Now().UTC(),
		EventType: "field_drop",
		Field:     field,
	})
}

// Stats returns cumulative counters for redactions, sampled drops, and field drops.
func (a *AuditLog) Stats() (redactions, drops, fieldDrops int64) {
	return a.redactions.Load(), a.drops.Load(), a.fieldDrops.Load()
}
