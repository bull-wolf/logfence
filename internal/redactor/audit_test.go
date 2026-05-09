package redactor

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestAuditLog_RecordRedaction(t *testing.T) {
	var buf bytes.Buffer
	al := NewAuditLog(&buf)
	al.RecordRedaction("email", "email-pattern")

	var ev AuditEvent
	if err := json.NewDecoder(&buf).Decode(&ev); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if ev.EventType != "redaction" {
		t.Errorf("expected event_type=redaction, got %q", ev.EventType)
	}
	if ev.Field != "email" {
		t.Errorf("expected field=email, got %q", ev.Field)
	}
	if ev.RuleName != "email-pattern" {
		t.Errorf("expected rule=email-pattern, got %q", ev.RuleName)
	}
	if ev.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestAuditLog_RecordSampledDrop(t *testing.T) {
	var buf bytes.Buffer
	al := NewAuditLog(&buf)
	al.RecordSampledDrop("service=auth")

	var ev AuditEvent
	if err := json.NewDecoder(&buf).Decode(&ev); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if ev.EventType != "sampled_drop" {
		t.Errorf("expected event_type=sampled_drop, got %q", ev.EventType)
	}
	if ev.SamplingKey != "service=auth" {
		t.Errorf("expected sampling_key=service=auth, got %q", ev.SamplingKey)
	}
}

func TestAuditLog_RecordFieldDrop(t *testing.T) {
	var buf bytes.Buffer
	al := NewAuditLog(&buf)
	al.RecordFieldDrop("password")

	var ev AuditEvent
	if err := json.NewDecoder(&buf).Decode(&ev); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if ev.EventType != "field_drop" {
		t.Errorf("expected event_type=field_drop, got %q", ev.EventType)
	}
	if ev.Field != "password" {
		t.Errorf("expected field=password, got %q", ev.Field)
	}
}

func TestAuditLog_Stats(t *testing.T) {
	al := NewAuditLog(io.Discard)
	al.RecordRedaction("token", "bearer")
	al.RecordRedaction("email", "email-pattern")
	al.RecordSampledDrop("svc=api")
	al.RecordFieldDrop("secret")

	r, d, f := al.Stats()
	if r != 2 {
		t.Errorf("expected 2 redactions, got %d", r)
	}
	if d != 1 {
		t.Errorf("expected 1 drop, got %d", d)
	}
	if f != 1 {
		t.Errorf("expected 1 field drop, got %d", f)
	}
}

func TestAuditLog_Discard(t *testing.T) {
	al := NewAuditLog(io.Discard)
	// Should not panic when writing to discard.
	for i := 0; i < 10; i++ {
		al.RecordRedaction("f", "r")
	}
	r, _, _ := al.Stats()
	if r != 10 {
		t.Errorf("expected 10, got %d", r)
	}
}

func TestAuditLog_MultipleEvents(t *testing.T) {
	var buf bytes.Buffer
	al := NewAuditLog(&buf)
	al.RecordRedaction("email", "email-rule")
	al.RecordFieldDrop("ssn")

	dec := json.NewDecoder(&buf)
	var events []AuditEvent
	for {
		var ev AuditEvent
		if err := dec.Decode(&ev); err != nil {
			break
		}
		events = append(events, ev)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	types := []string{events[0].EventType, events[1].EventType}
	joined := strings.Join(types, ",")
	if joined != "redaction,field_drop" {
		t.Errorf("unexpected event order/types: %s", joined)
	}
}
