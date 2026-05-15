package redactor

import (
	"encoding/json"
	"testing"
)

func TestTagger_NonJSONPassthrough(t *testing.T) {
	tagger, _ := NewTagger(DefaultTaggerConfig())
	input := []byte("plain text log line")
	out, err := tagger.Process(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != string(input) {
		t.Errorf("expected passthrough, got %s", out)
	}
}

func TestTagger_NoRulesNoChange(t *testing.T) {
	tagger, _ := NewTagger(DefaultTaggerConfig())
	input := []byte(`{"level":"info","msg":"hello"}`)
	out, err := tagger.Process(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var obj map[string]interface{}
	json.Unmarshal(out, &obj)
	if _, ok := obj["tags"]; ok {
		t.Error("expected no tags field when no rules match")
	}
}

func TestTagger_MatchAddsTags(t *testing.T) {
	cfg := TaggerConfig{
		TagField: "tags",
		Rules: []TaggerRule{
			{Field: "level", Contains: "error", Tags: []string{"alert", "ops"}},
		},
	}
	tagger, _ := NewTagger(cfg)
	input := []byte(`{"level":"error","msg":"disk full"}`)
	out, err := tagger.Process(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var obj map[string]interface{}
	json.Unmarshal(out, &obj)
	tags, ok := obj["tags"].([]interface{})
	if !ok || len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %v", obj["tags"])
	}
	if tags[0] != "alert" || tags[1] != "ops" {
		t.Errorf("unexpected tags: %v", tags)
	}
}

func TestTagger_NoMatchNoTags(t *testing.T) {
	cfg := TaggerConfig{
		TagField: "tags",
		Rules: []TaggerRule{
			{Field: "level", Contains: "error", Tags: []string{"alert"}},
		},
	}
	tagger, _ := NewTagger(cfg)
	input := []byte(`{"level":"info","msg":"all good"}`)
	out, err := tagger.Process(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var obj map[string]interface{}
	json.Unmarshal(out, &obj)
	if _, ok := obj["tags"]; ok {
		t.Error("expected no tags field")
	}
}

func TestTagger_MultipleRulesAccumulateTags(t *testing.T) {
	cfg := TaggerConfig{
		TagField: "tags",
		Rules: []TaggerRule{
			{Field: "level", Contains: "warn", Tags: []string{"warning"}},
			{Field: "msg", Contains: "timeout", Tags: []string{"network"}},
		},
	}
	tagger, _ := NewTagger(cfg)
	input := []byte(`{"level":"warn","msg":"connection timeout"}`)
	out, err := tagger.Process(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var obj map[string]interface{}
	json.Unmarshal(out, &obj)
	tags, ok := obj["tags"].([]interface{})
	if !ok || len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %v", obj["tags"])
	}
}

func TestTagger_MergesExistingTags(t *testing.T) {
	cfg := TaggerConfig{
		TagField: "tags",
		Rules: []TaggerRule{
			{Field: "level", Contains: "error", Tags: []string{"new-tag"}},
		},
	}
	tagger, _ := NewTagger(cfg)
	input := []byte(`{"level":"error","tags":["existing"]}`)
	out, err := tagger.Process(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var obj map[string]interface{}
	json.Unmarshal(out, &obj)
	tags, ok := obj["tags"].([]interface{})
	if !ok || len(tags) != 2 {
		t.Fatalf("expected 2 tags (existing + new), got %v", obj["tags"])
	}
}
