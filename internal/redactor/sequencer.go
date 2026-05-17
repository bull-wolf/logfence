package redactor

import (
	"encoding/json"
	"sync"
)

// SequencerConfig controls per-key sequence number injection.
type SequencerConfig struct {
	// Field is the JSON key written with the sequence number.
	Field string
	// KeyField is the JSON key used to partition counters (e.g. "service").
	// If empty, a single global counter is used.
	KeyField string
}

// DefaultSequencerConfig returns a sensible default configuration.
func DefaultSequencerConfig() SequencerConfig {
	return SequencerConfig{
		Field:    "_seq",
		KeyField: "service",
	}
}

// Sequencer annotates each log entry with a monotonically increasing sequence
// number, optionally partitioned by a key field.
type Sequencer struct {
	cfg     SequencerConfig
	mu      sync.Mutex
	counters map[string]uint64
}

// NewSequencer constructs a Sequencer with the given config.
func NewSequencer(cfg SequencerConfig) *Sequencer {
	if cfg.Field == "" {
		cfg.Field = "_seq"
	}
	return &Sequencer{
		cfg:      cfg,
		counters: make(map[string]uint64),
	}
}

// Process implements the pipeline stage interface.
// It injects a sequence number into JSON entries and passes plain text through.
func (s *Sequencer) Process(entry *Entry) (*Entry, bool) {
	if entry == nil {
		return nil, false
	}
	if !entry.IsJSON {
		return entry, true
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(entry.Raw, &obj); err != nil {
		return entry, true
	}

	partitionKey := ""
	if s.cfg.KeyField != "" {
		if v, ok := obj[s.cfg.KeyField]; ok {
			if str, ok := v.(string); ok {
				partitionKey = str
			}
		}
	}

	s.mu.Lock()
	s.counters[partitionKey]++
	seq := s.counters[partitionKey]
	s.mu.Unlock()

	obj[s.cfg.Field] = seq

	b, err := json.Marshal(obj)
	if err != nil {
		return entry, true
	}
	entry.Raw = b
	return entry, true
}
