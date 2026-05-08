package proxy

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"

	"github.com/yourorg/logfence/internal/redactor"
)

// SamplingHandler wraps an http.Handler and drops log requests that exceed
// the configured sampling policy.
type SamplingHandler struct {
	next    http.Handler
	sampler *redactor.Sampler
}

// NewSamplingHandler creates a SamplingHandler with the given sampler, delegating
// allowed requests to next.
func NewSamplingHandler(next http.Handler, cfg redactor.SamplerConfig) *SamplingHandler {
	return &SamplingHandler{
		next:    next,
		sampler: redactor.NewSampler(cfg),
	}
}

func (s *SamplingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := samplingKey(r)
	if !s.sampler.Allow(key) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	s.next.ServeHTTP(w, r)
}

// samplingKey derives a stable key from the request path and a short hash of
// the body so that identical log lines are grouped together for sampling.
func samplingKey(r *http.Request) string {
	if r.Body == nil {
		return r.URL.Path
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 512))
	if err != nil || len(body) == 0 {
		return r.URL.Path
	}
	// Restore body for downstream handlers.
	r.Body = io.NopCloser(io.MultiReader(
		newBytesReader(body),
		r.Body,
	))
	hash := md5.Sum(body) //nolint:gosec // non-cryptographic use
	return fmt.Sprintf("%s#%x", r.URL.Path, hash[:4])
}

// newBytesReader wraps a byte slice as an io.Reader (avoids importing bytes).
func newBytesReader(b []byte) io.Reader {
	return &bytesReader{data: b}
}

type bytesReader struct {
	data []byte
	pos  int
}

func (br *bytesReader) Read(p []byte) (int, error) {
	if br.pos >= len(br.data) {
		return 0, io.EOF
	}
	n := copy(p, br.data[br.pos:])
	br.pos += n
	return n, nil
}
