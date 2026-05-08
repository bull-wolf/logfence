package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/logfence/internal/redactor"
)

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	r, err := redactor.New(redactor.DefaultRules())
	if err != nil {
		t.Fatalf("failed to create redactor: %v", err)
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return NewHandler(r, next)
}

func TestHandler_RedactsJSONBody(t *testing.T) {
	h := newTestHandler(t)

	payload := map[string]interface{}{
		"message": "user logged in",
		"email":   "alice@example.com",
		"token":   "Bearer supersecret123",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	var captured map[string]interface{}
	h.next = http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(data, &captured)
	})

	h.ServeHTTP(rec, req)

	if captured["email"] == "alice@example.com" {
		t.Error("expected email to be redacted")
	}
	if captured["token"] == "Bearer supersecret123" {
		t.Error("expected token to be redacted")
	}
	if captured["message"] != "user logged in" {
		t.Errorf("expected message to be unchanged, got %v", captured["message"])
	}
}

func TestHandler_RedactsPlainTextBody(t *testing.T) {
	h := newTestHandler(t)

	rawLog := "2024-01-01 INFO user=alice@example.com action=login"
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString(rawLog))
	rec := httptest.NewRecorder()

	var capturedBody string
	h.next = http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		capturedBody = string(data)
	})

	h.ServeHTTP(rec, req)

	if bytes.Contains([]byte(capturedBody), []byte("alice@example.com")) {
		t.Error("expected email to be redacted from plain text body")
	}
}

func TestHandler_PassthroughNonPost(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	called := false
	h.next = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	h.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called for non-POST request")
	}
}
