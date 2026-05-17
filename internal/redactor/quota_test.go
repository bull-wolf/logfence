package redactor

import (
	"fmt"
	"testing"
	"time"
)

func newTestQuota(maxBytes int64, window time.Duration, keyField string) *Quota {
	q := NewQuota(QuotaConfig{
		MaxBytesPerWindow: maxBytes,
		Window:            window,
		KeyField:          keyField,
	})
	return q
}

func makeEntry(service string, raw string) *Entry {
	e := &Entry{
		Raw:    []byte(raw),
		Fields: map[string]interface{}{},
	}
	if service != "" {
		e.Fields["service"] = service
	}
	return e
}

func TestQuota_AllowsUnderLimit(t *testing.T) {
	q := newTestQuota(1000, time.Minute, "service")
	e := makeEntry("svc-a", "hello")
	_, ok := q.Process(e)
	if !ok {
		t.Fatal("expected entry to pass under quota")
	}
}

func TestQuota_BlocksOverLimit(t *testing.T) {
	q := newTestQuota(10, time.Minute, "service")
	e := makeEntry("svc-a", "this message is definitely longer than ten bytes")
	// First entry fills the bucket beyond limit
	_, ok := q.Process(e)
	if !ok {
		t.Fatal("expected first oversized entry to pass (it resets bucket)")
	}
	// Second entry should be dropped
	_, ok = q.Process(makeEntry("svc-a", "x"))
	if ok {
		t.Fatal("expected second entry to be dropped after quota exceeded")
	}
}

func TestQuota_WindowReset(t *testing.T) {
	q := newTestQuota(10, 50*time.Millisecond, "service")
	fixedNow := time.Now()
	q.now = func() time.Time { return fixedNow }

	e := makeEntry("svc-b", "1234567890") // exactly 10 bytes
	q.Process(e)

	// Advance past window
	q.now = func() time.Time { return fixedNow.Add(100 * time.Millisecond) }
	_, ok := q.Process(makeEntry("svc-b", "new"))
	if !ok {
		t.Fatal("expected entry to pass after window reset")
	}
}

func TestQuota_IndependentKeys(t *testing.T) {
	q := newTestQuota(20, time.Minute, "service")
	for i := 0; i < 3; i++ {
		svc := fmt.Sprintf("svc-%d", i)
		_, ok := q.Process(makeEntry(svc, "short"))
		if !ok {
			t.Fatalf("svc %s should pass independently", svc)
		}
	}
}

func TestQuota_ZeroMaxDisabled(t *testing.T) {
	q := newTestQuota(0, time.Minute, "service")
	for i := 0; i < 100; i++ {
		_, ok := q.Process(makeEntry("svc", "payload that would exceed any limit if enabled"))
		if !ok {
			t.Fatal("quota should be disabled when MaxBytesPerWindow is 0")
		}
	}
}

func TestQuota_NilEntryReturnsFalse(t *testing.T) {
	q := newTestQuota(1000, time.Minute, "service")
	_, ok := q.Process(nil)
	if ok {
		t.Fatal("nil entry should return false")
	}
}
