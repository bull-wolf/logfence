package proxy

import (
	"net/http"

	"github.com/yourorg/logfence/internal/redactor"
)

// NewQuotaMiddleware returns an http.Handler that enforces per-key byte quotas
// on incoming log entries before forwarding to next.
// Entries that exceed the quota receive a 429 response; others are forwarded.
func NewQuotaMiddleware(quota *redactor.Quota, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		body, r2, err := drainBody(r)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}

		entry := &redactor.Entry{
			Raw:    body,
			Fields: extractJSONFields(body),
		}

		_, ok := quota.Process(entry)
		if !ok {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r2)
	})
}
