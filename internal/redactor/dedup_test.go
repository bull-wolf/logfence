package redactor

import (
	"testing"
	"time"
)

func newTestDedup(window time.Duration) *Deduplicator {
	cfg := DedupConfig{WindowSize: window, MaxEntries: 8}
	d := NewDeduplicator(cfg)
	return d
}

func TestDedup_FirstSeenNotDuplicate(t *testing.T) {
	d := newTestDedup(5 * time.Second)
	if d.IsDuplicate([]byte("hello world")) {
		t.Fatal("expected first occurrence to not be a duplicate")
	}
}

func TestDedup_SecondSeenIsDuplicate(t *testing.T) {
	d := newTestDedup(5 * time.Second)
	payload := []byte("repeated log line")
	d.IsDuplicate(payload)
	if !d.IsDuplicate(payload) {
		t.Fatal("expected second occurrence to be a duplicate")
	}
}

func TestDedup_DifferentPayloadsNotDuplicate(t *testing.T) {
	d := newTestDedup(5 * time.Second)
	d.IsDuplicate([]byte("line one"))
	if d.IsDuplicate([]byte("line two")) {
		t.Fatal("different payloads should not be duplicates of each other")
	}
}

func TestDedup_ExpiryAllowsReplay(t *testing.T) {
	now := time.Now()
	d := newTestDedup(2 * time.Second)
	d.nowFunc = func() time.Time { return now }

	payload := []byte("expiring entry")
	d.IsDuplicate(payload)

	// Advance time beyond window.
	d.nowFunc = func() time.Time { return now.Add(3 * time.Second) }
	if d.IsDuplicate(payload) {
		t.Fatal("entry should have expired and not be a duplicate")
	}
}

func TestDedup_EvictsAtCapacity(t *testing.T) {
	d := NewDeduplicator(DedupConfig{WindowSize: time.Minute, MaxEntries: 4})

	for i := 0; i < 10; i++ {
		payload := []byte{byte(i)}
		d.IsDuplicate(payload)
	}

	d.mu.Lock()
	size := len(d.seen)
	d.mu.Unlock()

	if size > 4 {
		t.Fatalf("expected at most 4 entries, got %d", size)
	}
}

func TestDedup_DefaultConfig(t *testing.T) {
	cfg := DefaultDedupConfig()
	if cfg.WindowSize <= 0 {
		t.Fatal("expected positive window size")
	}
	if cfg.MaxEntries <= 0 {
		t.Fatal("expected positive max entries")
	}
}
