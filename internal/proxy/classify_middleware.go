package proxy

import (
	"bytes"
	"io"
	"net/http"

	"github.com/yourorg/logfence/internal/redactor"
)

// ClassifyMiddleware annotates incoming JSON log bodies with a sensitivity
// label before forwarding them to the next handler.
type ClassifyMiddleware struct {
	classifier *redactor.Classifier
	next       http.Handler
}

// NewClassifyMiddleware constructs a ClassifyMiddleware using the default
// classifier configuration.
func NewClassifyMiddleware(next http.Handler) (*ClassifyMiddleware, error) {
	c, err := redactor.NewClassifier(redactor.DefaultClassifierConfig())
	if err != nil {
		return nil, err
	}
	return &ClassifyMiddleware{classifier: c, next: next}, nil
}

// NewClassifyMiddlewareWithClassifier constructs a ClassifyMiddleware with a
// caller-supplied Classifier.
func NewClassifyMiddlewareWithClassifier(c *redactor.Classifier, next http.Handler) *ClassifyMiddleware {
	return &ClassifyMiddleware{classifier: c, next: next}
}

func (m *ClassifyMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !isJSONContent(r.Header.Get("Content-Type")) {
		m.next.ServeHTTP(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	annotated, err := m.classifier.Process(body)
	if err != nil || annotated == nil {
		annotated = body
	}

	r.Body = io.NopCloser(bytes.NewReader(annotated))
	r.ContentLength = int64(len(annotated))
	m.next.ServeHTTP(w, r)
}
