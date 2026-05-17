package redactor

import (
	"sync"
	"time"
)

// QuotaConfig controls per-key byte-based quotas over a rolling window.
type QuotaConfig struct {
	// MaxBytesPerWindow is the maximum number of bytes allowed per key per window.
	// Zero disables quota enforcement.
	MaxBytesPerWindow int64
	// Window is the duration of the rolling window.
	Window time.Duration
	// KeyField is the JSON field used to derive the quota key (e.g. "service").
	KeyField string
}

// DefaultQuotaConfig returns a sensible default configuration.
func DefaultQuotaConfig() QuotaConfig {
	return QuotaConfig{
		MaxBytesPerWindow: 1 << 20, // 1 MiB
		Window:            time.Minute,
		KeyField:          "service",
	}
}

type quotaBucket struct {
	bytes     int64
	resetAt   time.Time
}

// Quota enforces per-key byte quotas, dropping entries that exceed the limit.
type Quota struct {
	cfg     QuotaConfig
	mu      sync.Mutex
	buckets map[string]*quotaBucket
	now     func() time.Time
}

// NewQuota creates a Quota from the given config.
func NewQuota(cfg QuotaConfig) *Quota {
	return &Quota{
		cfg:     cfg,
		buckets: make(map[string]*quotaBucket),
		now:     time.Now,
	}
}

// Process checks whether the entry's byte size fits within the quota for its key.
// Returns (entry, false) when the entry is dropped.
func (q *Quota) Process(entry *Entry) (*Entry, bool) {
	if entry == nil {
		return nil, false
	}
	if q.cfg.MaxBytesPerWindow <= 0 {
		return entry, true
	}

	key := "_default"
	if q.cfg.KeyField != "" {
		if v, ok := entry.Fields[q.cfg.KeyField]; ok {
			if s, ok := v.(string); ok && s != "" {
				key = s
			}
		}
	}

	size := int64(len(entry.Raw))
	now := q.now()

	q.mu.Lock()
	defer q.mu.Unlock()

	b, exists := q.buckets[key]
	if !exists || now.After(b.resetAt) {
		q.buckets[key] = &quotaBucket{
			bytes:   size,
			resetAt: now.Add(q.cfg.Window),
		}
		return entry, true
	}

	if b.bytes+size > q.cfg.MaxBytesPerWindow {
		return entry, false
	}
	b.bytes += size
	return entry, true
}
