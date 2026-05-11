package redactor

import (
	"encoding/json"
	"testing"
)

func TestScrubber_ExactKeyMatch(t *testing.T) {
	s := NewScrubber(DefaultScrubberConfig())
	input := `{"username":"alice","password":"s3cr3t"}`
	out := s.ScrubJSON([]byte(input))

	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if m["password"] != "[SCRUBBED]" {
		t.Errorf("expected password to be scrubbed, got %v", m["password"])
	}
	if m["username"] != "alice" {
		t.Errorf("expected username to be preserved, got %v", m["username"])
	}
}

func TestScrubber_SuffixKeyMatch(t *testing.T) {
	s := NewScrubber(DefaultScrubberConfig())
	input := `{"db_password":"hunter2","host":"localhost"}`
	out := s.ScrubJSON([]byte(input))

	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if m["db_password"] != "[SCRUBBED]" {
		t.Errorf("expected db_password to be scrubbed, got %v", m["db_password"])
	}
	if m["host"] != "localhost" {
		t.Errorf("expected host to be preserved")
	}
}

func TestScrubber_NestedJSON(t *testing.T) {
	s := NewScrubber(DefaultScrubberConfig())
	input := `{"user":{"name":"bob","token":"abc123"}}`
	out := s.ScrubJSON([]byte(input))

	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	user := m["user"].(map[string]interface{})
	if user["token"] != "[SCRUBBED]" {
		t.Errorf("expected nested token to be scrubbed, got %v", user["token"])
	}
	if user["name"] != "bob" {
		t.Errorf("expected name to be preserved")
	}
}

func TestScrubber_NonJSONPassthrough(t *testing.T) {
	s := NewScrubber(DefaultScrubberConfig())
	input := []byte("plain text log line")
	out := s.ScrubJSON(input)
	if string(out) != string(input) {
		t.Errorf("expected non-JSON to pass through unchanged")
	}
}

func TestScrubber_CustomReplacement(t *testing.T) {
	s := NewScrubber(ScrubberConfig{
		Keys:        []string{"secret"},
		Replacement: "***",
	})
	input := `{"secret":"mysecret"}`
	out := s.ScrubJSON([]byte(input))

	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if m["secret"] != "***" {
		t.Errorf("expected custom replacement, got %v", m["secret"])
	}
}

func TestScrubber_ArrayOfObjects(t *testing.T) {
	s := NewScrubber(DefaultScrubberConfig())
	input := `{"users":[{"name":"alice","password":"p1"},{"name":"bob","password":"p2"}]}`
	out := s.ScrubJSON([]byte(input))

	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	users := m["users"].([]interface{})
	for _, u := range users {
		obj := u.(map[string]interface{})
		if obj["password"] != "[SCRUBBED]" {
			t.Errorf("expected password in array element to be scrubbed, got %v", obj["password"])
		}
	}
}
