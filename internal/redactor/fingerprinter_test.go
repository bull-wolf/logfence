package redactor

import (
	"encoding/json"
	"testing"
)

func TestFingerprinter_AnnotatesJSON(t *testing.T) {
	fp, err := NewFingerprinter(DefaultFingerprinterConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	input := `{"level":"error","msg":"disk full","service":"storage"}`
	out, ok := fp.Process([]byte(input))
	if !ok {
		t.Fatal("expected ok=true")
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(out, &obj); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	v, exists := obj["_fingerprint"]
	if !exists {
		t.Fatal("expected _fingerprint field")
	}
	if len(v.(string)) != 16 {
		t.Errorf("expected 16-char fingerprint, got %q", v)
	}
}

func TestFingerprinter_StableAcrossCalls(t *testing.T) {
	fp, _ := NewFingerprinter(DefaultFingerprinterConfig())
	input := []byte(`{"level":"warn","msg":"retry","service":"api"}`)

	out1, _ := fp.Process(input)
	out2, _ := fp.Process(input)

	var o1, o2 map[string]interface{}
	json.Unmarshal(out1, &o1)
	json.Unmarshal(out2, &o2)

	if o1["_fingerprint"] != o2["_fingerprint"] {
		t.Errorf("fingerprint not stable: %v vs %v", o1["_fingerprint"], o2["_fingerprint"])
	}
}

func TestFingerprinter_DifferentFieldsDifferentHash(t *testing.T) {
	fp, _ := NewFingerprinter(DefaultFingerprinterConfig())

	a, _ := fp.Process([]byte(`{"level":"error","msg":"disk full","service":"storage"}`))
	b, _ := fp.Process([]byte(`{"level":"info","msg":"started","service":"api"}`))

	var oa, ob map[string]interface{}
	json.Unmarshal(a, &oa)
	json.Unmarshal(b, &ob)

	if oa["_fingerprint"] == ob["_fingerprint"] {
		t.Error("expected different fingerprints for different entries")
	}
}

func TestFingerprinter_NonJSONPassthrough(t *testing.T) {
	fp, _ := NewFingerprinter(DefaultFingerprinterConfig())
	input := []byte("plain text log line")
	out, ok := fp.Process(input)
	if !ok {
		t.Fatal("expected ok=true for passthrough")
	}
	if string(out) != string(input) {
		t.Errorf("expected passthrough, got %q", out)
	}
}

func TestFingerprinter_NilEntryReturnsFalse(t *testing.T) {
	fp, _ := NewFingerprinter(DefaultFingerprinterConfig())
	out, ok := fp.Process(nil)
	if ok {
		t.Error("expected ok=false for nil entry")
	}
	if out != nil {
		t.Errorf("expected nil output, got %v", out)
	}
}

func TestFingerprinter_CustomField(t *testing.T) {
	cfg := FingerprinterConfig{
		Fields:           []string{"msg"},
		FingerprintField: "_fp",
	}
	fp, _ := NewFingerprinter(cfg)
	out, _ := fp.Process([]byte(`{"msg":"hello"}`))

	var obj map[string]interface{}
	json.Unmarshal(out, &obj)

	if _, ok := obj["_fp"]; !ok {
		t.Error("expected custom fingerprint field _fp")
	}
}
