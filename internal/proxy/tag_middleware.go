package proxy

import (
	"bytes"
	"io"
	"net/http"

	"github.com/yourorg/logfence/internal/redactor"
)

// tagMiddleware applies the Tagger to incoming JSON log bodies.
type tagMiddleware struct {
	tagger *redactor.Tagger
	next   http.Handler
}

// NewTagMiddleware wraps next with a tagger built from cfg.
func NewTagMiddleware(cfg redactor.TaggerConfig, next http.Handler) (http.Handler, error) {
	t, err := redactor.NewTagger(cfg)
	if err != nil {
		return nil, err
	}
	return &tagMiddleware{tagger: t, next: next}, nil
}

func (m *tagMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	tagged, err := m.tagger.Process(body)
	if err != nil {
		tagged = body
	}

	r.Body = io.NopCloser(bytes.NewReader(tagged))
	r.ContentLength = int64(len(tagged))
	m.next.ServeHTTP(w, r)
}
