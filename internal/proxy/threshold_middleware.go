package proxy

import (
	"bytes"
	"io"
	"net/http"

	"github.com/yourorg/logfence/internal/redactor"
)

// ThresholdMiddleware applies numeric threshold rules to incoming log entries.
type ThresholdMiddleware struct {
	threshold *redactor.Threshold
	next      http.Handler
}

// NewThresholdMiddleware creates a middleware that evaluates numeric field
// thresholds on each log entry, either annotating or dropping entries that
// breach configured limits.
func NewThresholdMiddleware(next http.Handler, rules []redactor.ThresholdRule) (http.Handler, error) {
	cfg := redactor.DefaultThresholdConfig()
	cfg.Rules = rules
	th, err := redactor.NewThreshold(cfg)
	if err != nil {
		return nil, err
	}
	return &ThresholdMiddleware{threshold: th, next: next}, nil
}

func (m *ThresholdMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.Body == nil {
		m.next.ServeHTTP(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	processed, keep := m.threshold.ProcessBytes(body)
	if !keep {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	r.Body = io.NopCloser(bytes.NewReader(processed))
	r.ContentLength = int64(len(processed))
	m.next.ServeHTTP(w, r)
}
