package proxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/yourusername/logfence/internal/redactor"
)

func newSequenceMiddleware(t *testing.T) http.Handler {
	t.Helper()
	seq, err := redactor.NewSequencer(redactor.DefaultSequencerConfig())
	if err != nil {
		t.Fatalf("NewSequencer: %v", err)
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return NewSequenceMiddleware(seq, next)
}

func TestSequenceMiddleware_NonPostPassthrough(t *testing.T) {
	h := newSequenceMiddleware(t)

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSequenceMiddleware_AnnotatesJSONBody(t *testing.T) {
	h := newSequenceMiddleware(t)

	body := `{"level":"info","msg":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/logs",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Capture what the next handler receives by wrapping.
	var captured []byte
	seq, err := redactor.NewSequencer(redactor.DefaultSequencerConfig())
	if err != nil {
		t.Fatalf("NewSequencer: %v", err)
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		captured = buf.Bytes()
		w.WriteHeader(http.StatusOK)
	})
	h = NewSequenceMiddleware(seq, next)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(captured, &out); err != nil {
		t.Fatalf("unmarshal captured body: %v", err)
	}

	cfg := redactor.DefaultSequencerConfig()
	if _, ok := out[cfg.Field]; !ok {
		t.Errorf("expected field %q in output, got %v", cfg.Field, out)
	}
}

func TestSequenceMiddleware_SequenceMonotonicallyIncreasing(t *testing.T) {
	seq, err := redactor.NewSequencer(redactor.DefaultSequencerConfig())
	if err != nil {
		t.Fatalf("NewSequencer: %v", err)
	}

	var sequences []int64
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		var out map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		cfg := redactor.DefaultSequencerConfig()
		raw, ok := out[cfg.Field]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var n int64
		switch v := raw.(type) {
		case float64:
			n = int64(v)
		case string:
			n, _ = strconv.ParseInt(v, 10, 64)
		}
		sequences = append(sequences, n)
		w.WriteHeader(http.StatusOK)
	})
	h := NewSequenceMiddleware(seq, next)

	for i := 0; i < 5; i++ {
		body := `{"msg":"entry"}`
		req := httptest.NewRequest(http.MethodPost, "/logs",
			bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}

	if len(sequences) != 5 {
		t.Fatalf("expected 5 sequences, got %d", len(sequences))
	}
	for i := 1; i < len(sequences); i++ {
		if sequences[i] <= sequences[i-1] {
			t.Errorf("sequence not monotonically increasing: %v", sequences)
		}
	}
}

func TestSequenceMiddleware_PlainTextPassthrough(t *testing.T) {
	h := newSequenceMiddleware(t)

	body := "plain text log line"
	req := httptest.NewRequest(http.MethodPost, "/logs",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
