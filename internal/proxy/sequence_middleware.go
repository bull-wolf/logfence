package proxy

import (
	"bytes"
	"io"
	"net/http"

	"github.com/yourorg/logfence/internal/redactor"
)

// NewSequenceMiddleware wraps next with a middleware that injects a per-key
// monotonic sequence number into every JSON log entry.
func NewSequenceMiddleware(cfg redactor.SequencerConfig, next http.Handler) http.Handler {
	seq := redactor.NewSequencer(cfg)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()

		isJSON := isJSONContent(r.Header.Get("Content-Type"))
		entry := &redactor.Entry{
			Raw:    body,
			IsJSON: isJSON,
		}

		result, ok := seq.Process(entry)
		if !ok || result == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(result.Raw))
		r.ContentLength = int64(len(result.Raw))
		next.ServeHTTP(w, r)
	})
}
