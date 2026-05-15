package redactor

import (
	"encoding/json"
	"testing"
)

func TestNormalizer_LowercaseKeys(t *testing.T) {
	cfg := DefaultNormalizerConfig()
	cfg.LowercaseKeys = true
	n := NewNormalizer(cfg)

	input := `{"UserName":"alice","EMAIL":"alice@example.com"}`
	out := n.NormalizeJSON([]byte(input))

	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if _, ok := m["username"]; !ok {
		t.Error("expected key 'username' to be present")
	}
	if _, ok := m["email"]; !ok {
		t.Error("expected key 'email' to be present")
	}
	if _, ok := m["UserName"]; ok {
		t.Error("expected original key 'UserName' to be absent")
	}
}

func TestNormalizer_TrimSpace(t *testing.T) {
	cfg := DefaultNormalizerConfig()
	cfg.TrimSpace = true
	n := NewNormalizer(cfg)

	input := `{"msg":"  hello world  "}`
	out := n.NormalizeJSON([]byte(input))

	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if m["msg"] != "hello world" {
		t.Errorf("expected trimmed value, got %q", m["msg"])
	}
}

func TestNormalizer_CollapseWhitespace(t *testing.T) {
	cfg := NormalizerConfig{LowercaseKeys: false, TrimSpace: false, CollapseWhitespace: true}
	n := NewNormalizer(cfg)

	input := `{"msg":"hello   world  again"}`
	out := n.NormalizeJSON([]byte(input))

	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if m["msg"] != "hello world again" {
		t.Errorf("expected collapsed whitespace, got %q", m["msg"])
	}
}

func TestNormalizer_NonJSONPassthrough(t *testing.T) {
	n := NewNormalizer(DefaultNormalizerConfig())
	input := []byte("plain text log line")
	out := n.NormalizeJSON(input)
	if string(out) != string(input) {
		t.Errorf("expected passthrough for non-JSON, got %q", out)
	}
}

func TestNormalizer_NestedObject(t *testing.T) {
	cfg := DefaultNormalizerConfig()
	n := NewNormalizer(cfg)

	input := `{"Outer":{"InnerKey":"  value  "}}`
	out := n.NormalizeJSON([]byte(input))

	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	outer, ok := m["outer"].(map[string]interface{})
	if !ok {
		t.Fatal("expected nested object at key 'outer'")
	}
	if outer["innerkey"] != "value" {
		t.Errorf("expected trimmed nested value, got %q", outer["innerkey"])
	}
}

func TestNormalizer_ArrayValues(t *testing.T) {
	cfg := DefaultNormalizerConfig()
	n := NewNormalizer(cfg)

	input := `{"tags":["  go  ","  logs  "]}`
	out := n.NormalizeJSON([]byte(input))

	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	tags, ok := m["tags"].([]interface{})
	if !ok || len(tags) != 2 {
		t.Fatal("expected tags array with 2 elements")
	}
	if tags[0] != "go" || tags[1] != "logs" {
		t.Errorf("expected trimmed array values, got %v", tags)
	}
}
