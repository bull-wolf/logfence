package redactor

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// FingerprinterConfig controls how log entry fingerprints are generated.
type FingerprinterConfig struct {
	// Fields to include in the fingerprint. If empty, all fields are used.
	Fields []string `yaml:"fields"`
	// FingerprintField is the JSON key where the fingerprint is written.
	FingerprintField string `yaml:"fingerprint_field"`
}

// DefaultFingerprinterConfig returns a sensible default configuration.
func DefaultFingerprinterConfig() FingerprinterConfig {
	return FingerprinterConfig{
		Fields:           []string{"level", "msg", "service"},
		FingerprintField: "_fingerprint",
	}
}

// Fingerprinter annotates log entries with a stable hash derived from
// selected fields, enabling downstream deduplication and correlation.
type Fingerprinter struct {
	cfg FingerprinterConfig
}

// NewFingerprinter creates a Fingerprinter with the given config.
func NewFingerprinter(cfg FingerprinterConfig) (*Fingerprinter, error) {
	if cfg.FingerprintField == "" {
		cfg.FingerprintField = "_fingerprint"
	}
	return &Fingerprinter{cfg: cfg}, nil
}

// Process annotates the entry with a fingerprint field and returns it.
// Non-JSON payloads are passed through unchanged.
func (f *Fingerprinter) Process(entry []byte) ([]byte, bool) {
	if entry == nil {
		return entry, false
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(entry, &obj); err != nil {
		return entry, true
	}

	fingerprint := f.computeFingerprint(obj)
	obj[f.cfg.FingerprintField] = fingerprint

	out, err := json.Marshal(obj)
	if err != nil {
		return entry, true
	}
	return out, true
}

func (f *Fingerprinter) computeFingerprint(obj map[string]interface{}) string {
	keys := f.cfg.Fields
	if len(keys) == 0 {
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)
	}

	h := sha256.New()
	for _, k := range keys {
		v, ok := obj[k]
		if !ok {
			continue
		}
		h.Write([]byte(k))
		h.Write([]byte("="))
		h.Write([]byte(toString(v)))
		h.Write([]byte(";;"))
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}
