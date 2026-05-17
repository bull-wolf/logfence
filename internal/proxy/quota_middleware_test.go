package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/logfence/internal/redactor"
)

func newQuotaMiddleware(maxBytes int64) http.Handler {
	q := redactor.NewQuota(redactor.QuotaConfig{
		MaxBytesPerWindow: maxBytes,
		Window:            time.Minute,
		KeyField:          "service",
	})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return NewQuotaMiddleware(q, next)
}

func TestQuotaMiddleware_AllowsUnderLimit(t *testing.T) {
	h := newQuotaMiddleware(10000)
	body := `{"service":"api","msg":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/logs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestQuotaMiddleware_BlocksOverLimit(t *testing.T) {
	h := newQuotaMiddleware(5)
	body := `{"service":"api","msg":"this is definitely over five bytes"}`
	// First request resets and fills bucket
	req1 := httptest.NewRequest(http.MethodPost, "/logs", strings.NewReader(body))
	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, req1)

	// Second request should be rate-limited
	req2 := httptest.NewRequest(http.MethodPost, "/logs", strings.NewReader(body))
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec2.Code)
	}
}

func TestQuotaMiddleware_NonPostPassthrough(t *testing.T) {
	h := newQuotaMiddleware(1) // extremely tight quota
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected GET to pass through, got %d", rec.Code)
	}
}
