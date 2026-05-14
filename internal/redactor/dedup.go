package redactor

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// DedupConfig controls log-line deduplication behaviour.
type DedupConfig struct {
	// WindowSize is how long a seen hash is remembered.
	WindowSize time.Duration
	// MaxEntries caps the in-memory hash set to avoid unbounded growth.
	MaxEntries int
}

// DefaultDedupConfig returns sensible defaults.
func DefaultDedupConfig() DedupConfig {
	return DedupConfig{
		WindowSize: 10 * time.Second,
		MaxEntries: 4096,
	}
}

type entry struct {
	expiresAt time.Time
}

// Deduplicator suppresses repeated identical log lines within a time window.
type Deduplicator struct {
	mu      sync.Mutex
	seen    map[string]entry
	cfg     DedupConfig
	nowFunc func() time.Time
}

// NewDeduplicator creates a Deduplicator with the given config.
func NewDeduplicator(cfg DedupConfig) *Deduplicator {
	if cfg.WindowSize <= 0 {
		cfg.WindowSize = DefaultDedupConfig().WindowSize
	}
	if cfg.MaxEntries <= 0 {
		cfg.MaxEntries = DefaultDedupConfig().MaxEntries
	}
	return &Deduplicator{
		seen:    make(map[string]entry, cfg.MaxEntries),
		cfg:     cfg,
		nowFunc: time.Now,
	}
}

// IsDuplicate returns true if the payload has been seen within the window.
// It records the payload if it is new or has expired.
func (d *Deduplicator) IsDuplicate(payload []byte) bool {
	h := hashBytes(payload)
	now := d.nowFunc()

	d.mu.Lock()
	defer d.mu.Unlock()

	if e, ok := d.seen[h]; ok && now.Before(e.expiresAt) {
		return true
	}

	// Evict expired entries when at capacity.
	if len(d.seen) >= d.cfg.MaxEntries {
		d.evictLocked(now)
	}

	d.seen[h] = entry{expiresAt: now.Add(d.cfg.WindowSize)}
	return false
}

func (d *Deduplicator) evictLocked(now time.Time) {
	for k, e := range d.seen {
		if now.After(e.expiresAt) {
			delete(d.seen, k)
		}
	}
}

func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
