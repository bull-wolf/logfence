package proxy

import (
	"bytes"
	"io"
	"net/http"

	"github.com/yourorg/logfence/internal/redactor"
)

// NewAlertMiddleware wraps next with an Alerter that scans request bodies for
// patterns defined in cfg, annotates matching JSON payloads with a severity
// label, and forwards the (potentially modified) body downstream.
func NewAlertMiddleware(cfg redactor.AlerterConfig, next http.Handler) (http.Handler, error) {
	a, err := redactor.NewAlerter(cfg)
	if err != nil {
		return nil, err
	}
	return &alertMiddleware{alerter: a, next: next}, nil
}

type alertMiddleware struct {
	aleriter *redactor.Alerter
	next     http.Handler
}

func (m *alertMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.Body == nil {
		m.next.ServeHTTP(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}

	processed := m.alerter.Process(body)
	r.Body = io.NopCloser(bytes.NewReader(processed))
	r.ContentLength = int64(len(processed))

	m.next.ServeHTTP(w, r)
}
