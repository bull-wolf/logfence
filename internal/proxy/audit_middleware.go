package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/yourorg/logfence/internal/redactor"
)

// AuditMiddleware wraps an http.Handler and emits an AuditEvent for every
// field that was redacted or dropped in the proxied log payload.
type AuditMiddleware struct {
	next     http.Handler
	audit    *redactor.AuditLog
	filter   []string // field names considered sensitive
}

// NewAuditMiddleware creates middleware that records field-level audit events.
func NewAuditMiddleware(next http.Handler, al *redactor.AuditLog, sensitiveFields []string) *AuditMiddleware {
	return &AuditMiddleware{
		next:   next,
		audit:  al,
		filter: sensitiveFields,
	}
}

func (m *AuditMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.Body == nil {
		m.next.ServeHTTP(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	// Inspect JSON fields before forwarding.
	var entry map[string]any
	if json.Unmarshal(body, &entry) == nil {
		m.inspectFields(entry)
	}

	// Restore body for downstream handler.
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.ContentLength = int64(len(body))
	m.next.ServeHTTP(w, r)
}

// inspectFields records audit events for any field matching the sensitive list.
func (m *AuditMiddleware) inspectFields(entry map[string]any) {
	sensitiveSet := make(map[string]struct{}, len(m.filter))
	for _, f := range m.filter {
		sensitiveSet[f] = struct{}{}
	}
	for key := range entry {
		if _, found := sensitiveSet[key]; found {
			m.audit.RecordFieldDrop(key)
		}
	}
}
