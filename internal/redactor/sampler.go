package redactor

import (
	"sync"
	"time"
)

// SamplerConfig controls log sampling to reduce high-volume noise.
type SamplerConfig struct {
	// Rate is the fraction of logs to allow through (0.0 to 1.0).
	// A value of 1.0 means all logs pass; 0.1 means ~10%.
	Rate float64 `yaml:"rate" json:"rate"`
	// BurstLimit is the max number of identical messages per window before sampling kicks in.
	BurstLimit int `yaml:"burst_limit" json:"burst_limit"`
	// Window is the duration over which burst counting is tracked.
	Window time.Duration `yaml:"window" json:"window"`
}

// DefaultSamplerConfig returns a conservative sampler that allows all logs through.
func DefaultSamplerConfig() SamplerConfig {
	return SamplerConfig{
		Rate:       1.0,
		BurstLimit: 100,
		Window:     10 * time.Second,
	}
}

type msgEntry struct {
	count     int
	windowEnd time.Time
}

// Sampler decides whether a given log message should be forwarded.
type Sampler struct {
	cfg     SamplerConfig
	mu      sync.Mutex
	counter map[string]*msgEntry
	clock   func() time.Time
}

// NewSampler creates a Sampler from the provided config.
func NewSampler(cfg SamplerConfig) *Sampler {
	if cfg.Rate <= 0 {
		cfg.Rate = 0
	}
	if cfg.Rate > 1 {
		cfg.Rate = 1
	}
	if cfg.BurstLimit <= 0 {
		cfg.BurstLimit = 1
	}
	if cfg.Window <= 0 {
		cfg.Window = 10 * time.Second
	}
	return &Sampler{
		cfg:     cfg,
		counter: make(map[string]*msgEntry),
		clock:   time.Now,
	}
}

// Allow returns true if the message with the given key should be forwarded.
func (s *Sampler) Allow(key string) bool {
	if s.cfg.Rate >= 1.0 {
		return true
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	entry, ok := s.counter[key]
	if !ok || now.After(entry.windowEnd) {
		s.counter[key] = &msgEntry{count: 1, windowEnd: now.Add(s.cfg.Window)}
		return true
	}

	entry.count++
	if entry.count <= s.cfg.BurstLimit {
		return true
	}

	// Beyond burst limit: apply rate sampling deterministically via count modulo.
	threshold := int(1.0 / s.cfg.Rate)
	return entry.count%threshold == 0
}
