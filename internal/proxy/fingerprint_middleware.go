package proxy

import (
	"bytes"
	"io"
	"net/http"

	"github.com/yourorg/logfence/internal/redactor"
)

// FingerprintMiddleware annotates incoming JSON log payloads with a stable
// fingerprint field before forwarding to the next handler.
type FingerprintMiddleware struct {
	fp   *redactor.Fingerprinter
	next http.Handler
}

// NewFingerprintMiddleware constructs the middleware with default config.
func NewFingerprintMiddleware(next http.Handler) (*FingerprintMiddleware, error) {
	return NewFingerprintMiddlewareWithConfig(next, redactor.DefaultFingerprinterConfig())
}

// NewFingerprintMiddlewareWithConfig constructs the middleware with a custom config.
func NewFingerprintMiddlewareWithConfig(next http.Handler, cfg redactor.FingerprinterConfig) (*FingerprintMiddleware, error) {
	fp, err := redactor.NewFingerprinter(cfg)
	if err != nil {
		return nil, err
	}
	return &FingerprintMiddleware{fp: fp, next: next}, nil
}

func (m *FingerprintMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		m.next.ServeHTTP(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	annotated, _ := m.fp.Process(body)

	r2 := r.Clone(r.Context())
	r2.Body = io.NopCloser(bytes.NewReader(annotated))
	r2.ContentLength = int64(len(annotated))

	m.next.ServeHTTP(w, r2)
}
