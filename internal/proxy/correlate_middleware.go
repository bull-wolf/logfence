package proxy

import (
	"bytes"
	"io"
	"net/http"

	"github.com/yourorg/logfence/internal/redactor"
)

// NewCorrelateMiddleware returns an HTTP middleware that annotates JSON log
// entries sharing the same key field with a correlation count, making it easy
// to trace repeated log lines back to a single request or operation.
func NewCorrelateMiddleware(next http.Handler) http.Handler {
	cfg := redactor.DefaultCorrelatorConfig()
	return NewCorrelateMiddlewareWithConfig(next, cfg)
}

// NewCorrelateMiddlewareWithConfig is like NewCorrelateMiddleware but accepts
// a custom CorrelatorConfig.
func NewCorrelateMiddlewareWithConfig(next http.Handler, cfg redactor.CorrelatorConfig) http.Handler {
	correlator := redactor.NewCorrelator(cfg)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Body == nil {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}
		r.Body.Close()

		entry := &redactor.Entry{Body: body}
		keep, err := correlator.Process(entry)
		if err != nil || !keep {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(entry.Body))
		r.ContentLength = int64(len(entry.Body))
		next.ServeHTTP(w, r)
	})
}
