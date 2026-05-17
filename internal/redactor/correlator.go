package redactor

import (
	"encoding/json"
	"sync"
	"time"
)

// CorrelatorConfig controls how log entries are correlated by a shared key.
type CorrelatorConfig struct {
	// KeyField is the JSON field used to group entries (e.g. "request_id").
	KeyField string
	// CorrelationField is the field written to mark correlated entries.
	CorrelationField string
	// TTL is how long a key is remembered before expiry.
	TTL time.Duration
	// MaxKeys caps the number of tracked keys to avoid unbounded memory growth.
	MaxKeys int
}

// DefaultCorrelatorConfig returns sensible defaults.
func DefaultCorrelatorConfig() CorrelatorConfig {
	return CorrelatorConfig{
		KeyField:         "request_id",
		CorrelationField: "correlated",
		TTL:              5 * time.Minute,
		MaxKeys:          10_000,
	}
}

type correlatorEntry struct {
	count   int
	expires time.Time
}

// Correlator annotates log entries that share a key field, marking them as
// part of the same logical request or trace.
type Correlator struct {
	cfg   CorrelatorConfig
	mu    sync.Mutex
	store map[string]*correlatorEntry
}

// NewCorrelator creates a Correlator with the given config.
func NewCorrelator(cfg CorrelatorConfig) *Correlator {
	return &Correlator{
		cfg:   cfg,
		store: make(map[string]*correlatorEntry),
	}
}

// Process implements pipeline.Processor. It annotates the entry with a
// correlation count when the key field is present.
func (c *Correlator) Process(entry *Entry) (bool, error) {
	if entry == nil {
		return false, nil
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(entry.Body, &obj); err != nil {
		return true, nil // non-JSON passes through unchanged
	}

	keyVal, ok := obj[c.cfg.KeyField]
	if !ok {
		return true, nil
	}
	key, ok := keyVal.(string)
	if !ok || key == "" {
		return true, nil
	}

	count := c.record(key)
	if count > 1 {
		obj[c.cfg.CorrelationField] = count
		b, err := json.Marshal(obj)
		if err != nil {
			return true, err
		}
		entry.Body = b
	}

	return true, nil
}

func (c *Correlator) record(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// Evict expired entries when we're at capacity.
	if len(c.store) >= c.cfg.MaxKeys {
		for k, e := range c.store {
			if now.After(e.expires) {
				delete(c.store, k)
			}
		}
	}

	e, exists := c.store[key]
	if !exists || now.After(e.expires) {
		c.store[key] = &correlatorEntry{count: 1, expires: now.Add(c.cfg.TTL)}
		return 1
	}
	e.count++
	return e.count
}
