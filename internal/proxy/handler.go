package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/yourorg/logfence/internal/redactor"
)

// Handler is an HTTP handler that redacts incoming log payloads.
type Handler struct {
	redactor *redactor.Redactor
	next     http.Handler
}

// NewHandler creates a new proxy Handler wrapping the given next handler.
func NewHandler(r *redactor.Redactor, next http.Handler) *Handler {
	return &Handler{redactor: r, next: next}
}

// ServeHTTP reads the request body, redacts sensitive fields, and forwards
// the sanitised payload to the next handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.next.ServeHTTP(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		// Not JSON — attempt string-level redaction and forward as-is.
		redacted := h.redactor.RedactString(string(body))
		r.Body = io.NopCloser(bytes.NewBufferString(redacted))
		r.ContentLength = int64(len(redacted))
		log.Printf("[logfence] non-JSON body redacted (%d bytes)", len(redacted))
		h.next.ServeHTTP(w, r)
		return
	}

	h.redactor.RedactFields(payload)

	redacted, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to marshal redacted payload", http.StatusInternalServerError)
		return
	}

	log.Printf("[logfence] JSON body redacted (%d bytes)", len(redacted))
	r.Body = io.NopCloser(bytes.NewBuffer(redacted))
	r.ContentLength = int64(len(redacted))
	h.next.ServeHTTP(w, r)
}
