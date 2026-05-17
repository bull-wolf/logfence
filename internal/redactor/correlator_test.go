package redactor

import (
	"encoding/json"
	"testing"
	"time"
)

func newTestCorrelator() *Correlator {
	cfg := DefaultCorrelatorConfig()
	cfg.TTL = 200 * time.Millisecond
	return NewCorrelator(cfg)
}

func jsonEntry(t *testing.T, fields map[string]interface{}) *Entry {
	t.Helper()
	b, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return &Entry{Body: b}
}

func TestCorrelator_FirstSeenNoAnnotation(t *testing.T) {
	c := newTestCorrelator()
	e := jsonEntry(t, map[string]interface{}{"request_id": "abc", "msg": "hello"})
	keep, err := c.Process(e)
	if err != nil || !keep {
		t.Fatalf("unexpected keep=%v err=%v", keep, err)
	}
	var obj map[string]interface{}
	json.Unmarshal(e.Body, &obj)
	if _, ok := obj["correlated"]; ok {
		t.Error("first entry should not have correlated field")
	}
}

func TestCorrelator_SecondSeenAnnotated(t *testing.T) {
	c := newTestCorrelator()
	e1 := jsonEntry(t, map[string]interface{}{"request_id": "xyz"})
	c.Process(e1)

	e2 := jsonEntry(t, map[string]interface{}{"request_id": "xyz", "msg": "second"})
	keep, err := c.Process(e2)
	if err != nil || !keep {
		t.Fatalf("unexpected keep=%v err=%v", keep, err)
	}
	var obj map[string]interface{}
	json.Unmarshal(e2.Body, &obj)
	count, ok := obj["correlated"]
	if !ok {
		t.Fatal("expected correlated field on second entry")
	}
	if count.(float64) != 2 {
		t.Errorf("expected count 2, got %v", count)
	}
}

func TestCorrelator_DifferentKeysIndependent(t *testing.T) {
	c := newTestCorrelator()
	c.Process(jsonEntry(t, map[string]interface{}{"request_id": "aaa"}))
	e := jsonEntry(t, map[string]interface{}{"request_id": "bbb"})
	c.Process(e)
	var obj map[string]interface{}
	json.Unmarshal(e.Body, &obj)
	if _, ok := obj["correlated"]; ok {
		t.Error("different key should not be annotated")
	}
}

func TestCorrelator_ExpiryResetsCount(t *testing.T) {
	c := newTestCorrelator()
	c.Process(jsonEntry(t, map[string]interface{}{"request_id": "ttl-key"}))
	time.Sleep(250 * time.Millisecond)
	e := jsonEntry(t, map[string]interface{}{"request_id": "ttl-key"})
	c.Process(e)
	var obj map[string]interface{}
	json.Unmarshal(e.Body, &obj)
	if _, ok := obj["correlated"]; ok {
		t.Error("entry after TTL expiry should not be annotated")
	}
}

func TestCorrelator_NonJSONPassthrough(t *testing.T) {
	c := newTestCorrelator()
	e := &Entry{Body: []byte("plain text log line")}
	keep, err := c.Process(e)
	if err != nil || !keep {
		t.Errorf("non-JSON should pass through: keep=%v err=%v", keep, err)
	}
	if string(e.Body) != "plain text log line" {
		t.Error("non-JSON body should be unchanged")
	}
}

func TestCorrelator_NilEntryReturnsFalse(t *testing.T) {
	c := newTestCorrelator()
	keep, err := c.Process(nil)
	if keep || err != nil {
		t.Errorf("nil entry: expected keep=false err=nil, got keep=%v err=%v", keep, err)
	}
}
