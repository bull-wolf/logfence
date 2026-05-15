package redactor

import (
	"sync"
	"time"
)

// RateLimiterConfig controls per-key rate limiting of log entries.
type RateLimiterConfig struct {
	// MaxPerWindow is the maximum number of log entries allowed per key per window.
	// Zero disables rate limiting.
	MaxPerWindow int
	// Window is the duration of each rate limiting window.
	Window time.Duration
}

// DefaultRateLimiterConfig returns a sensible default: 100 entries per 10 seconds.
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		MaxPerWindow: 100,
		Window:       10 * time.Second,
	}
}

type windowCounter struct {
	count     int
	windowEnd time.Time
}

// RateLimiter tracks log entry counts per key and signals when a key exceeds
// its allowed rate.
type RateLimiter struct {
	cfg     RateLimiterConfig
	mu      sync.Mutex
	counters map[string]*windowCounter
}

// NewRateLimiter constructs a RateLimiter with the given config.
// If cfg.MaxPerWindow is zero, Allow always returns true.
func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		cfg:      cfg,
		counters: make(map[string]*windowCounter),
	}
}

// Allow returns true if the entry for the given key is within the rate limit,
// and false if the entry should be dropped. It is safe for concurrent use.
func (r *RateLimiter) Allow(key string) bool {
	if r.cfg.MaxPerWindow <= 0 {
		return true
	}

	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	wc, ok := r.counters[key]
	if !ok || now.After(wc.windowEnd) {
		r.counters[key] = &windowCounter{
			count:     1,
			windowEnd: now.Add(r.cfg.Window),
		}
		return true
	}

	wc.count++
	return wc.count <= r.cfg.MaxPerWindow
}

// Stats returns the current count and window-end time for a key.
// Returns zeros if the key has not been seen.
func (r *RateLimiter) Stats(key string) (count int, windowEnd time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if wc, ok := r.counters[key]; ok {
		return wc.count, wc.windowEnd
	}
	return 0, time.Time{}
}
