package redactor

import (
	"encoding/json"
	"testing"
)

func defaultTestPipeline(t *testing.T) *Pipeline {
	t.Helper()
	p, err := DefaultPipeline()
	if err != nil {
		t.Fatalf("DefaultPipeline() error: %v", err)
	}
	return p
}

func TestPipeline_RedactsEmailInJSON(t *testing.T) {
	p := defaultTestPipeline(t)
	input := `{"level":"info","msg":"user logged in","email":"alice@example.com"}`

	out, err := p.ProcessJSON([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected output, got nil (entry dropped)")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if v, ok := result["email"]; ok {
		if s, ok := v.(string); ok && s == "alice@example.com" {
			t.Errorf("email was not redacted, got %q", s)
		}
	}
}

func TestPipeline_PlainTextFallback(t *testing.T) {
	p := defaultTestPipeline(t)
	input := "Bearer eyJhbGciOiJIUzI1NiJ9.payload.sig"

	out, err := p.ProcessJSON([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) == input {
		t.Errorf("plain text token was not redacted")
	}
}

func TestPipeline_SamplingDropsEntries(t *testing.T) {
	p, err := NewPipeline(PipelineConfig{
		FieldFilter:  DefaultFieldFilterConfig(),
		Sampler:      SamplerConfig{Rate: 0.0, BurstSize: 0},
		RedactorRules: DefaultRules(),
	})
	if err != nil {
		t.Fatalf("NewPipeline() error: %v", err)
	}

	input := `{"level":"debug","msg":"heartbeat"}`
	out, err := p.ProcessJSON([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected entry to be dropped, got: %s", out)
	}
}

func TestPipeline_InvalidPattern(t *testing.T) {
	_, err := NewPipeline(PipelineConfig{
		FieldFilter:  DefaultFieldFilterConfig(),
		Sampler:      DefaultSamplerConfig(),
		RedactorRules: []RuleConfig{{Pattern: "[", Replacement: "REDACTED"}},
	})
	if err == nil {
		t.Error("expected error for invalid regex pattern, got nil")
	}
}

func TestSamplingKeyFromEntry(t *testing.T) {
	cases := []struct {
		entry    map[string]interface{}
		expected string
	}{
		{map[string]interface{}{"service": "auth"}, "auth"},
		{map[string]interface{}{"level": "error"}, "error"},
		{map[string]interface{}{"other": "val"}, "default"},
		{map[string]interface{}{}, "default"},
	}
	for _, tc := range cases {
		got := samplingKeyFromEntry(tc.entry)
		if got != tc.expected {
			t.Errorf("samplingKeyFromEntry(%v) = %q, want %q", tc.entry, got, tc.expected)
		}
	}
}
