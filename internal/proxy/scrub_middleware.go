package proxy

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/yourorg/logfence/internal/redactor"
)

// ScrubMiddleware wraps an http.Handler and scrubs sensitive JSON fields
// from incoming request bodies before passing them downstream.
type ScrubMiddleware struct {
	scrubber *redactor.Scrubber
	next     http.Handler
}

// NewScrubMiddleware creates a ScrubMiddleware with the given scrubber.
func NewScrubMiddleware(s *redactor.Scrubber, next http.Handler) *ScrubMiddleware {
	return &ScrubMiddleware{scrubber: s, next: next}
}

func (m *ScrubMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && isJSONContent(r) && r.Body != nil {
		body, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err == nil {
			scrubbed := m.scrubber.ScrubJSON(body)
			r.Body = io.NopCloser(bytes.NewReader(scrubbed))
			r.ContentLength = int64(len(scrubbed))
		} else {
			r.Body = io.NopCloser(bytes.NewReader(body))
		}
	}
	m.next.ServeHTTP(w, r)
}

func isJSONContent(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return strings.Contains(ct, "application/json")
}
