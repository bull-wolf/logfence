package redactor

import (
	"encoding/json"
	"strings"
	"testing"
)

func newTestRedactMiddleware(t *testing.T) *RedactMiddleware {
	t.Helper()
	r, err := New(DefaultRules())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return NewRedactMiddleware(r)
}

func TestRedactMiddleware_PlainTextRedactsEmail(t *testing.T) {
	m := newTestRedactMiddleware(t)
	entry := &Entry{Payload: []byte("contact us at user@example.com for help")}
	out, ok := m.Process(entry)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if strings.Contains(string(out.Payload), "user@example.com") {
		t.Errorf("email not redacted: %s", out.Payload)
	}
}

func TestRedactMiddleware_JSONRedactsEmailValue(t *testing.T) {
	m := newTestRedactMiddleware(t)
	body, _ := json.Marshal(map[string]string{"msg": "hello user@example.com", "level": "info"})
	entry := &Entry{Payload: body}
	out, ok := m.Process(entry)
	if !ok {
		t.Fatal("expected ok=true")
	}
	var fields map[string]interface{}
	if err := json.Unmarshal(out.Payload, &fields); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if strings.Contains(fields["msg"].(string), "user@example.com") {
		t.Errorf("email not redacted in JSON field: %v", fields["msg"])
	}
}

func TestRedactMiddleware_JSONRedactsBearerToken(t *testing.T) {
	m := newTestRedactMiddleware(t)
	body, _ := json.Marshal(map[string]string{"auth": "Bearer eyJhbGciOiJIUzI1NiJ9.payload.sig"})
	entry := &Entry{Payload: body}
	out, ok := m.Process(entry)
	if !ok {
		t.Fatal("expected ok=true")
	}
	var fields map[string]interface{}
	if err := json.Unmarshal(out.Payload, &fields); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if strings.Contains(fields["auth"].(string), "eyJ") {
		t.Errorf("bearer token not redacted: %v", fields["auth"])
	}
}

func TestRedactMiddleware_NilEntryReturnsFalse(t *testing.T) {
	m := newTestRedactMiddleware(t)
	out, ok := m.Process(nil)
	if ok || out != nil {
		t.Errorf("expected (nil, false) for nil entry, got (%v, %v)", out, ok)
	}
}

func TestRedactMiddleware_EmptyPayloadPassthrough(t *testing.T) {
	m := newTestRedactMiddleware(t)
	entry := &Entry{Payload: []byte{}}
	out, ok := m.Process(entry)
	if !ok {
		t.Fatal("expected ok=true for empty payload")
	}
	if len(out.Payload) != 0 {
		t.Errorf("expected empty payload, got %q", out.Payload)
	}
}
