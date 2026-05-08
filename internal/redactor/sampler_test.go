package redactor

import (
	"fmt"
	"testing"
	"time"
)

func TestSampler_AllowAll(t *testing.T) {
	s := NewSampler(SamplerConfig{Rate: 1.0, BurstLimit: 5, Window: time.Second})
	for i := 0; i < 200; i++ {
		if !s.Allow("key") {
			t.Fatalf("expected Allow=true at iteration %d with rate=1.0", i)
		}
	}
}

func TestSampler_BlockAll(t *testing.T) {
	s := NewSampler(SamplerConfig{Rate: 0.0, BurstLimit: 1, Window: time.Second})
	// First message in window always passes (burst=1).
	s.Allow("key")
	allowed := 0
	for i := 0; i < 100; i++ {
		if s.Allow("key") {
			allowed++
		}
	}
	// With rate=0 clamped to 0, threshold = 1/0 → division avoided; rate stays 0.
	// After burst, no messages should pass.
	if allowed > 0 {
		t.Fatalf("expected 0 allowed beyond burst with rate=0, got %d", allowed)
	}
}

func TestSampler_BurstAllowedThenSampled(t *testing.T) {
	cfg := SamplerConfig{Rate: 0.1, BurstLimit: 5, Window: 10 * time.Second}
	s := NewSampler(cfg)

	// First 5 calls should all be allowed (burst).
	for i := 0; i < 5; i++ {
		if !s.Allow("msg") {
			t.Fatalf("expected burst allow at i=%d", i)
		}
	}

	// Beyond burst, only ~10% should pass.
	allowed := 0
	total := 100
	for i := 0; i < total; i++ {
		if s.Allow("msg") {
			allowed++
		}
	}
	if allowed == 0 || allowed > total/2 {
		t.Fatalf("expected ~10%% sampling beyond burst, got %d/%d", allowed, total)
	}
}

func TestSampler_WindowReset(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	s := NewSampler(SamplerConfig{Rate: 0.5, BurstLimit: 2, Window: 5 * time.Second})
	s.clock = func() time.Time { return now }

	s.Allow("k")
	s.Allow("k")
	s.Allow("k") // beyond burst

	// Advance past window.
	now = now.Add(6 * time.Second)
	if !s.Allow("k") {
		t.Fatal("expected Allow=true after window reset")
	}
}

func TestSampler_IndependentKeys(t *testing.T) {
	s := NewSampler(SamplerConfig{Rate: 0.5, BurstLimit: 1, Window: time.Minute})
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key-%d", i)
		if !s.Allow(key) {
			t.Fatalf("first call for %s should always be allowed", key)
		}
	}
}

func TestDefaultSamplerConfig(t *testing.T) {
	cfg := DefaultSamplerConfig()
	if cfg.Rate != 1.0 {
		t.Fatalf("expected default rate 1.0, got %f", cfg.Rate)
	}
	if cfg.BurstLimit != 100 {
		t.Fatalf("expected default burst 100, got %d", cfg.BurstLimit)
	}
}
