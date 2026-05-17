package redactor

import (
	"encoding/json"
	"sync"
	"testing"
)

func newTestSequencer() *Sequencer {
	return NewSequencer(DefaultSequencerConfig())
}

func TestSequencer_NonJSONPassthrough(t *testing.T) {
	s := newTestSequencer()
	entry := &Entry{Raw: []byte("plain text log"), IsJSON: false}
	out, ok := s.Process(entry)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if string(out.Raw) != "plain text log" {
		t.Fatalf("unexpected mutation: %s", out.Raw)
	}
}

func TestSequencer_NilEntryReturnsFalse(t *testing.T) {
	s := newTestSequencer()
	_, ok := s.Process(nil)
	if ok {
		t.Fatal("expected ok=false for nil entry")
	}
}

func TestSequencer_AnnotatesWithSeq(t *testing.T) {
	s := newTestSequencer()
	entry := &Entry{Raw: []byte(`{"level":"info"}`), IsJSON: true}
	out, ok := s.Process(entry)
	if !ok {
		t.Fatal("expected ok=true")
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(out.Raw, &obj); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, exists := obj["_seq"]; !exists {
		t.Fatal("expected _seq field")
	}
}

func TestSequencer_MonotonicallyIncreasing(t *testing.T) {
	s := newTestSequencer()
	var prev float64
	for i := 0; i < 5; i++ {
		entry := &Entry{Raw: []byte(`{"service":"api"}`), IsJSON: true}
		out, _ := s.Process(entry)
		var obj map[string]interface{}
		_ = json.Unmarshal(out.Raw, &obj)
		curr := obj["_seq"].(float64)
		if curr <= prev {
			t.Fatalf("sequence not increasing: got %v after %v", curr, prev)
		}
		prev = curr
	}
}

func TestSequencer_PartitionedByKeyField(t *testing.T) {
	s := newTestSequencer()

	for i := 0; i < 3; i++ {
		s.Process(&Entry{Raw: []byte(`{"service":"api"}`), IsJSON: true})
	}
	for i := 0; i < 2; i++ {
		s.Process(&Entry{Raw: []byte(`{"service":"worker"}`), IsJSON: true})
	}

	apiEntry := &Entry{Raw: []byte(`{"service":"api"}`), IsJSON: true}
	out, _ := s.Process(apiEntry)
	var obj map[string]interface{}
	_ = json.Unmarshal(out.Raw, &obj)
	if obj["_seq"].(float64) != 4 {
		t.Fatalf("expected api seq=4, got %v", obj["_seq"])
	}
}

func TestSequencer_ConcurrentSafe(t *testing.T) {
	s := newTestSequencer()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Process(&Entry{Raw: []byte(`{"service":"svc"}`), IsJSON: true})
		}()
	}
	wg.Wait()

	entry := &Entry{Raw: []byte(`{"service":"svc"}`), IsJSON: true}
	out, _ := s.Process(entry)
	var obj map[string]interface{}
	_ = json.Unmarshal(out.Raw, &obj)
	if obj["_seq"].(float64) != 51 {
		t.Fatalf("expected seq=51 after 50 concurrent writes, got %v", obj["_seq"])
	}
}
