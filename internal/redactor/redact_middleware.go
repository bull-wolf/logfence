package redactor

import (
	"bytes"
	"encoding/json"
	"io"
)

// RedactMiddleware applies regex-based redaction to log entries passing
// through the pipeline. It supports both JSON and plain-text payloads.
type RedactMiddleware struct {
	redactor *Redactor
}

// NewRedactMiddleware constructs a RedactMiddleware using the provided Redactor.
func NewRedactMiddleware(r *Redactor) *RedactMiddleware {
	return &RedactMiddleware{redactor: r}
}

// Process redacts sensitive values from the entry payload.
// For JSON payloads it redacts string field values; for plain text it
// redacts the raw body. Returns (nil, false) to signal the entry should
// be dropped only when an internal error prevents safe processing.
func (m *RedactMiddleware) Process(entry *Entry) (*Entry, bool) {
	if entry == nil {
		return nil, false
	}

	if !isJSON(entry.Payload) {
		redacted := m.redactor.RedactString(string(entry.Payload))
		out := *entry
		out.Payload = []byte(redacted)
		return &out, true
	}

	var fields map[string]interface{}
	if err := json.Unmarshal(entry.Payload, &fields); err != nil {
		// Not valid JSON despite content-type hint; redact as plain text.
		redacted := m.redactor.RedactString(string(entry.Payload))
		out := *entry
		out.Payload = []byte(redacted)
		return &out, true
	}

	m.redactor.RedactFields(fields)

	buf, err := json.Marshal(fields)
	if err != nil {
		return entry, true
	}

	out := *entry
	out.Payload = buf
	return &out, true
}

// isJSON returns true when the payload appears to be a JSON object or array.
func isJSON(p []byte) bool {
	trimmed := bytes.TrimSpace(p)
	if len(trimmed) == 0 {
		return false
	}
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	token, err := dec.Token()
	if err != nil {
		return false
	}
	delim, ok := token.(json.Delim)
	return ok && (delim == '{' || delim == '[')
}

// Entry is a minimal log entry flowing through a redactor pipeline stage.
// Re-declared here for packages that import only this file; the canonical
// definition lives in pipeline.go.
var _ io.Reader = (*bytes.Reader)(nil) // compile-time import anchor
