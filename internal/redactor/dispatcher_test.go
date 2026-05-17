package redactor

import (
	"encoding/json"
	"testing"
)

func newTestDispatcher() *Dispatcher {
	return NewDispatcher(DefaultDispatcherConfig())
}

func jsonEntryDispatch(fields map[string]interface{}) []byte {
	b, _ := json.Marshal(fields)
	return b
}

func TestDispatcher_DefaultOutputForPlainText(t *testing.T) {
	d := newTestDispatcher()
	out := d.Dispatch([]byte("plain text log line"))
	if out != "default" {
		t.Errorf("expected default, got %q", out)
	}
}

func TestDispatcher_DefaultOutputWhenFieldMissing(t *testing.T) {
	d := newTestDispatcher()
	entry := jsonEntryDispatch(map[string]interface{}{"msg": "hello"})
	out := d.Dispatch(entry)
	if out != "default" {
		t.Errorf("expected default, got %q", out)
	}
}

func TestDispatcher_ErrorLevelRoutesToErrors(t *testing.T) {
	d := newTestDispatcher()
	entry := jsonEntryDispatch(map[string]interface{}{"level": "error", "msg": "boom"})
	out := d.Dispatch(entry)
	if out != "errors" {
		t.Errorf("expected errors, got %q", out)
	}
}

func TestDispatcher_WarnLevelRoutesToWarnings(t *testing.T) {
	d := newTestDispatcher()
	entry := jsonEntryDispatch(map[string]interface{}{"level": "WARNING", "msg": "watch out"})
	out := d.Dispatch(entry)
	if out != "warnings" {
		t.Errorf("expected warnings, got %q", out)
	}
}

func TestDispatcher_InfoLevelRoutesToDefault(t *testing.T) {
	d := newTestDispatcher()
	entry := jsonEntryDispatch(map[string]interface{}{"level": "info", "msg": "ok"})
	out := d.Dispatch(entry)
	if out != "default" {
		t.Errorf("expected default, got %q", out)
	}
}

func TestDispatcher_CustomFieldAndRule(t *testing.T) {
	d := NewDispatcher(DispatcherConfig{
		FieldName:     "service",
		DefaultOutput: "misc",
		Rules: []DispatchRule{
			{Contains: "auth", Output: "auth-logs"},
		},
	})
	entry := jsonEntryDispatch(map[string]interface{}{"service": "auth-service", "msg": "login"})
	out := d.Dispatch(entry)
	if out != "auth-logs" {
		t.Errorf("expected auth-logs, got %q", out)
	}
}

func TestDispatcher_FirstMatchingRuleWins(t *testing.T) {
	d := NewDispatcher(DispatcherConfig{
		FieldName:     "level",
		DefaultOutput: "default",
		Rules: []DispatchRule{
			{Contains: "error", Output: "errors"},
			{Contains: "err", Output: "other-errors"},
		},
	})
	entry := jsonEntryDispatch(map[string]interface{}{"level": "error"})
	out := d.Dispatch(entry)
	if out != "errors" {
		t.Errorf("expected errors (first match), got %q", out)
	}
}
