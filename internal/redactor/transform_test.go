package redactor

import (
	"testing"
)

func TestTransform_StripControlChars(t *testing.T) {
	cfg := DefaultTransformConfig()
	input := "hello\x00world\x01"
	got := Transform(input, cfg)
	want := "helloworld"
	if got != want {
		t.Errorf("StripControlChars: got %q, want %q", got, want)
	}
}

func TestTransform_PreservesTabNewline(t *testing.T) {
	cfg := DefaultTransformConfig()
	input := "line1\nline2\ttabbed"
	got := Transform(input, cfg)
	if got != input {
		t.Errorf("expected tabs/newlines preserved, got %q", got)
	}
}

func TestTransform_Lowercase(t *testing.T) {
	cfg := TransformConfig{Lowercase: true, StripControlChars: false}
	got := Transform("Hello WORLD", cfg)
	if got != "hello world" {
		t.Errorf("Lowercase: got %q", got)
	}
}

func TestTransform_Truncate(t *testing.T) {
	cfg := TransformConfig{TruncateAt: 5, StripControlChars: false}
	got := Transform("abcdefgh", cfg)
	if got != "abcde" {
		t.Errorf("Truncate: got %q, want %q", got, "abcde")
	}
}

func TestTransform_TruncateNoOp(t *testing.T) {
	cfg := TransformConfig{TruncateAt: 20, StripControlChars: false}
	got := Transform("short", cfg)
	if got != "short" {
		t.Errorf("TruncateNoOp: got %q", got)
	}
}

func TestTransform_Combined(t *testing.T) {
	cfg := TransformConfig{
		Lowercase:         true,
		TruncateAt:        6,
		StripControlChars: true,
	}
	got := Transform("HELLO\x00WORLD", cfg)
	if got != "hellow" {
		t.Errorf("Combined: got %q, want %q", got, "hellow")
	}
}

func TestTransformFields_StringValues(t *testing.T) {
	cfg := TransformConfig{Lowercase: true, StripControlChars: false}
	entry := map[string]any{
		"msg":   "Hello World",
		"level": "INFO",
		"count": 42,
	}
	result := TransformFields(entry, cfg)
	if result["msg"] != "hello world" {
		t.Errorf("msg: got %v", result["msg"])
	}
	if result["level"] != "info" {
		t.Errorf("level: got %v", result["level"])
	}
	if result["count"] != 42 {
		t.Errorf("count should be unchanged, got %v", result["count"])
	}
}

func TestTransformFields_EmptyEntry(t *testing.T) {
	cfg := DefaultTransformConfig()
	result := TransformFields(map[string]any{}, cfg)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestDefaultTransformConfig(t *testing.T) {
	cfg := DefaultTransformConfig()
	if cfg.Lowercase {
		t.Error("Lowercase should default to false")
	}
	if cfg.TruncateAt != 0 {
		t.Error("TruncateAt should default to 0")
	}
	if !cfg.StripControlChars {
		t.Error("StripControlChars should default to true")
	}
}
