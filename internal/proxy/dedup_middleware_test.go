package proxy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/logfence/internal/redactor"
)

func newDedupMiddleware(t *testing.T) *DedupMiddleware {
	t.Helper()
	cfg := redactor.DedupConfig{WindowSize: 5 * time.Second, MaxEntries: 64}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return NewDedupMiddleware(inner, cfg)
}

func TestDedupMiddleware_FirstRequestPasses(t *testing.T) {
	m := newDedupMiddleware(t)
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString(`{"msg":"hello"}`))
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestDedupMiddleware_DuplicateBodyDropped(t *testing.T) {
	m := newDedupMiddleware(t)
	body := `{"msg":"duplicate line"}`

	req1 := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString(body))
	m.ServeHTTP(httptest.NewRecorder(), req1)

	req2 := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString(body))
	rec2 := httptest.NewRecorder()
	m.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for duplicate, got %d", rec2.Code)
	}
}

func TestDedupMiddleware_DifferentBodiesPass(t *testing.T) {
	m := newDedupMiddleware(t)

	for i, body := range []string{`{"msg":"a"}`, `{"msg":"b"}`, `{"msg":"c"}`} {
		req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		m.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, rec.Code)
		}
	}
}

func TestDedupMiddleware_NonPostPassthrough(t *testing.T) {
	m := newDedupMiddleware(t)
	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected GET to pass through, got %d", rec.Code)
	}
}
