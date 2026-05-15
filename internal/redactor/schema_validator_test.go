package redactor

import (
	"testing"
)

func TestSchemaValidator_ValidEntry(t *testing.T) {
	v, err := NewSchemaValidator(DefaultSchemaValidatorConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	input := []byte(`{"level":"info","message":"hello world"}`)
	out, keep := v.Process(input)
	if !keep {
		t.Fatal("expected entry to be kept")
	}
	if string(out) != string(input) {
		t.Errorf("expected output unchanged, got %s", out)
	}
}

func TestSchemaValidator_MissingRequiredField_PassThrough(t *testing.T) {
	cfg := DefaultSchemaValidatorConfig()
	cfg.DropInvalid = false
	v, _ := NewSchemaValidator(cfg)

	input := []byte(`{"level":"info"}`)
	out, keep := v.Process(input)
	if !keep {
		t.Fatal("expected pass-through when DropInvalid=false")
	}
	if string(out) != string(input) {
		t.Errorf("output should be unchanged")
	}
}

func TestSchemaValidator_MissingRequiredField_Drop(t *testing.T) {
	cfg := DefaultSchemaValidatorConfig()
	cfg.DropInvalid = true
	v, _ := NewSchemaValidator(cfg)

	_, keep := v.Process([]byte(`{"level":"info"}`))
	if keep {
		t.Fatal("expected entry to be dropped")
	}
}

func TestSchemaValidator_PatternMismatch_Drop(t *testing.T) {
	cfg := DefaultSchemaValidatorConfig()
	cfg.DropInvalid = true
	v, _ := NewSchemaValidator(cfg)

	// "verbose" is not a valid level.
	_, keep := v.Process([]byte(`{"level":"verbose","message":"hi"}`))
	if keep {
		t.Fatal("expected entry with invalid level to be dropped")
	}
}

func TestSchemaValidator_NonJSONPassthrough(t *testing.T) {
	v, _ := NewSchemaValidator(DefaultSchemaValidatorConfig())
	input := []byte("plain text log line")
	out, keep := v.Process(input)
	if !keep {
		t.Fatal("non-JSON should always pass through")
	}
	if string(out) != string(input) {
		t.Errorf("non-JSON output should be unchanged")
	}
}

func TestNewSchemaValidator_InvalidPattern(t *testing.T) {
	cfg := SchemaValidatorConfig{
		FieldPatterns: map[string]string{"level": `[invalid`},
	}
	_, err := NewSchemaValidator(cfg)
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}

func TestSchemaValidator_PatternMismatch_PassThrough(t *testing.T) {
	cfg := DefaultSchemaValidatorConfig()
	cfg.DropInvalid = false
	v, _ := NewSchemaValidator(cfg)

	input := []byte(`{"level":"verbose","message":"hi"}`)
	out, keep := v.Process(input)
	if !keep {
		t.Fatal("expected pass-through when DropInvalid=false")
	}
	if string(out) != string(input) {
		t.Errorf("output should be unchanged")
	}
}
