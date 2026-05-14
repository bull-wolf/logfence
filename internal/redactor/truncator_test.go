package redactor

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTruncator_ShortValueUnchanged(t *testing.T) {
	tr := NewTruncator(TruncatorConfig{MaxBytes: 64, Suffix: "..."})
	input := `{"msg":"hello"}`
	out := tr.Apply([]byte(input))
	if string(out) != input {
		t.Errorf("expected unchanged, got %s", out)
	}
}

func TestTruncator_LongValueTruncated(t *testing.T) {
	tr := NewTruncator(TruncatorConfig{MaxBytes: 10, Suffix: "...[truncated]"})
	long := strings.Repeat("a", 100)
	input, _ := json.Marshal(map[string]string{"msg": long})
	out := tr.Apply(input)

	var result map[string]string
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if !strings.HasSuffix(result["msg"], "...[truncated]") {
		t.Errorf("expected truncation suffix, got: %s", result["msg"])
	}
	if len(result["msg"]) > 10+len("...[truncated]") {
		t.Errorf("value too long after truncation: %d chars", len(result["msg"]))
	}
}

func TestTruncator_NonJSONPassthrough(t *testing.T) {
	tr := NewTruncator(DefaultTruncatorConfig())
	input := []byte("plain text log line")
	out := tr.Apply(input)
	if string(out) != string(input) {
		t.Errorf("non-JSON should pass through unchanged")
	}
}

func TestTruncator_ZeroMaxBytesDisabled(t *testing.T) {
	tr := NewTruncator(TruncatorConfig{MaxBytes: 0, Suffix: "..."})
	long := strings.Repeat("x", 1000)
	input, _ := json.Marshal(map[string]string{"data": long})
	out := tr.Apply(input)

	var result map[string]string
	json.Unmarshal(out, &result)
	if result["data"] != long {
		t.Error("expected no truncation when MaxBytes is 0")
	}
}

func TestTruncator_SpecificFieldsOnly(t *testing.T) {
	tr := NewTruncator(TruncatorConfig{
		MaxBytes: 5,
		Suffix:   "...",
		Fields:   []string{"secret"},
	})
	input, _ := json.Marshal(map[string]string{
		"secret": "toolongvalue",
		"msg":    "toolongvalue",
	})
	out := tr.Apply(input)

	var result map[string]string
	json.Unmarshal(out, &result)
	if !strings.HasSuffix(result["secret"], "...") {
		t.Errorf("expected 'secret' to be truncated, got: %s", result["secret"])
	}
	if result["msg"] != "toolongvalue" {
		t.Errorf("expected 'msg' to be unchanged, got: %s", result["msg"])
	}
}

func TestTruncator_DefaultConfig(t *testing.T) {
	cfg := DefaultTruncatorConfig()
	if cfg.MaxBytes != 512 {
		t.Errorf("expected default MaxBytes=512, got %d", cfg.MaxBytes)
	}
	if cfg.Suffix == "" {
		t.Error("expected non-empty default suffix")
	}
}

func TestSafeTruncate_UTF8Boundary(t *testing.T) {
	// "é" is 2 bytes (0xC3 0xA9); truncating at 1 byte should not split the rune.
	s := "aé"
	result := safeTruncate(s, 2)
	if result != "a" {
		t.Errorf("expected 'a', got %q", result)
	}
}
