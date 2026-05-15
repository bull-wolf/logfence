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

func newTagMiddleware(t *testing.T, cfg redactor.TaggerConfig) http.Handler {
	t.Helper()
	var captured []byte
	sink := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	mw, err := NewTagMiddleware(cfg, sink)
	if err != nil {
		t.Fatalf("NewTagMiddleware: %v", err)
	}
	_ = captured
	return mw
}

func TestTagMiddleware_NonPostPassthrough(t *testing.T) {
	cfg := redactor.DefaultTaggerConfig()
	var body []byte
	sink := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	mw, _ := NewTagMiddleware(cfg, sink)

	req := httptest.NewRequest(http.MethodGet, "/logs", bytes.NewBufferString(`{"level":"error"}`))
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	_ = body
}

func TestTagMiddleware_TagsAppliedToJSON(t *testing.T) {
	cfg := redactor.TaggerConfig{
		TagField: "tags",
		Rules: []redactor.TaggerRule{
			{Field: "level", Contains: "error", Tags: []string{"critical"}},
		},
	}
	var captured []byte
	sink := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	mw, err := NewTagMiddleware(cfg, sink)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := `{"level":"error","msg":"crash"}`
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(captured, &obj); err != nil {
		t.Fatalf("invalid JSON in captured body: %v", err)
	}
	tags, ok := obj["tags"].([]interface{})
	if !ok || len(tags) == 0 {
		t.Errorf("expected tags to be set, got %v", obj["tags"])
	}
}

func TestTagMiddleware_PlainTextUnchanged(t *testing.T) {
	cfg := redactor.TaggerConfig{
		TagField: "tags",
		Rules: []redactor.TaggerRule{
			{Field: "level", Contains: "error", Tags: []string{"critical"}},
		},
	}
	var captured []byte
	sink := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	mw, _ := NewTagMiddleware(cfg, sink)

	plain := "this is a plain log line"
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString(plain))
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if string(captured) != plain {
		t.Errorf("expected plain text passthrough, got %q", captured)
	}
}
