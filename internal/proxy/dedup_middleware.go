package proxy

import (
	"bytes"
	"io"
	"net/http"

	"github.com/yourorg/logfence/internal/redactor"
)

// DedupMiddleware drops HTTP log-forwarding requests whose body is an exact
// duplicate of a recently seen body, within a configurable time window.
type DedupMiddleware struct {
	next  http.Handler
	dedup *redactor.Deduplicator
}

// NewDedupMiddleware wraps next with deduplication using cfg.
func NewDedupMiddleware(next http.Handler, cfg redactor.DedupConfig) *DedupMiddleware {
	return &DedupMiddleware{
		next:  next,
		dedup: redactor.NewDeduplicator(cfg),
	}
}

func (m *DedupMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.Body == nil {
		m.next.ServeHTTP(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	if m.dedup.IsDuplicate(body) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	m.next.ServeHTTP(w, r)
}
