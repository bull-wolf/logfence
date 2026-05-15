package redactor

import (
	"testing"
	"time"
)

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{MaxPerWindow: 3, Window: time.Second})
	for i := 0; i < 3; i++ {
		if !rl.Allow("svc") {
			t.Fatalf("expected Allow=true on call %d", i+1)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{MaxPerWindow: 2, Window: time.Second})
	rl.Allow("svc")
	rl.Allow("svc")
	if rl.Allow("svc") {
		t.Fatal("expected Allow=false after exceeding limit")
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{MaxPerWindow: 1, Window: 50 * time.Millisecond})
	if !rl.Allow("svc") {
		t.Fatal("expected first Allow=true")
	}
	if rl.Allow("svc") {
		t.Fatal("expected second Allow=false within window")
	}
	time.Sleep(60 * time.Millisecond)
	if !rl.Allow("svc") {
		t.Fatal("expected Allow=true after window reset")
	}
}

func TestRateLimiter_IndependentKeys(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{MaxPerWindow: 1, Window: time.Second})
	if !rl.Allow("a") {
		t.Fatal("expected Allow=true for key a")
	}
	if !rl.Allow("b") {
		t.Fatal("expected Allow=true for key b")
	}
	if rl.Allow("a") {
		t.Fatal("expected Allow=false for key a after limit")
	}
}

func TestRateLimiter_ZeroMaxDisabled(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{MaxPerWindow: 0, Window: time.Second})
	for i := 0; i < 1000; i++ {
		if !rl.Allow("svc") {
			t.Fatalf("expected Allow=true when disabled, failed on call %d", i+1)
		}
	}
}

func TestRateLimiter_Stats(t *testing.T) {
	rl := NewRateLimiter(DefaultRateLimiterConfig())
	count, end := rl.Stats("unknown")
	if count != 0 || !end.IsZero() {
		t.Fatal("expected zero stats for unseen key")
	}

	rl.Allow("known")
	rl.Allow("known")
	count, end = rl.Stats("known")
	if count != 2 {
		t.Fatalf("expected count=2, got %d", count)
	}
	if end.IsZero() {
		t.Fatal("expected non-zero window end")
	}
}

func TestDefaultRateLimiterConfig(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	if cfg.MaxPerWindow != 100 {
		t.Fatalf("expected MaxPerWindow=100, got %d", cfg.MaxPerWindow)
	}
	if cfg.Window != 10*time.Second {
		t.Fatalf("expected Window=10s, got %v", cfg.Window)
	}
}
