package redactor

import (
	"strings"
	"testing"
)

func TestMask_FullReplace(t *testing.T) {
	cfg := DefaultMaskConfig()
	result := Mask("secret", cfg)
	if result != "******" {
		t.Errorf("expected '******', got %q", result)
	}
}

func TestMask_ShowPrefix(t *testing.T) {
	cfg := MaskConfig{ShowPrefix: 2, MaskChar: '*', MinMaskLen: 3}
	result := Mask("abcdef", cfg)
	if !strings.HasPrefix(result, "ab") {
		t.Errorf("expected prefix 'ab', got %q", result)
	}
	if !strings.Contains(result, "***") {
		t.Errorf("expected at least 3 mask chars, got %q", result)
	}
}

func TestMask_ShowSuffix(t *testing.T) {
	cfg := MaskConfig{ShowSuffix: 2, MaskChar: '*', MinMaskLen: 3}
	result := Mask("abcdef", cfg)
	if !strings.HasSuffix(result, "ef") {
		t.Errorf("expected suffix 'ef', got %q", result)
	}
}

func TestMask_ShowPrefixAndSuffix(t *testing.T) {
	cfg := MaskConfig{ShowPrefix: 1, ShowSuffix: 1, MaskChar: '-', MinMaskLen: 2}
	result := Mask("hello", cfg)
	if result != "h---o" {
		t.Errorf("expected 'h---o', got %q", result)
	}
}

func TestMask_TooShort_MasksAll(t *testing.T) {
	cfg := MaskConfig{ShowPrefix: 3, ShowSuffix: 3, MaskChar: '*', MinMaskLen: 4}
	// value has only 4 runes, can't honour prefix+suffix=6
	result := Mask("abcd", cfg)
	if result != "****" {
		t.Errorf("expected '****', got %q", result)
	}
}

func TestMask_DefaultMaskChar(t *testing.T) {
	cfg := MaskConfig{MinMaskLen: 4} // MaskChar zero value
	result := Mask("secret", cfg)
	for _, ch := range result {
		if ch != '*' {
			t.Errorf("expected all '*', got %q", result)
			break
		}
	}
}

func TestMaskEmail_Standard(t *testing.T) {
	result := MaskEmail("john.doe@example.com")
	if !strings.HasPrefix(result, "j") {
		t.Errorf("expected preserved first char 'j', got %q", result)
	}
	if !strings.HasSuffix(result, "@example.com") {
		t.Errorf("expected domain preserved, got %q", result)
	}
	if strings.Contains(result, "ohn") {
		t.Errorf("local part should be masked, got %q", result)
	}
}

func TestMaskEmail_NoAtSign(t *testing.T) {
	// Should fall back to full masking.
	result := MaskEmail("notanemail")
	for _, ch := range result {
		if ch != '*' {
			t.Errorf("expected full mask for invalid email, got %q", result)
			break
		}
	}
}

func TestMaskEmail_SingleCharLocal(t *testing.T) {
	result := MaskEmail("a@example.com")
	if !strings.HasSuffix(result, "@example.com") {
		t.Errorf("expected domain preserved, got %q", result)
	}
}
