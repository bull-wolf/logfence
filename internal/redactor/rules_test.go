package redactor

import (
	"testing"
)

func TestCompileRules_Valid(t *testing.T) {
	configs := []RuleConfig{
		{Name: "test", Pattern: `\d+`, Replacement: "[NUM]"},
		{Name: "word", Pattern: `\w+`, Replacement: "[WORD]"},
	}
	rules, err := CompileRules(configs)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].Name != "test" {
		t.Errorf("expected rule name 'test', got '%s'", rules[0].Name)
	}
}

func TestCompileRules_InvalidPattern(t *testing.T) {
	configs := []RuleConfig{
		{Name: "bad", Pattern: `[invalid`, Replacement: "[X]"},
	}
	_, err := CompileRules(configs)
	if err == nil {
		t.Fatal("expected error for invalid pattern, got nil")
	}
}

func TestCompileRules_DefaultReplacement(t *testing.T) {
	configs := []RuleConfig{
		{Name: "no-replace", Pattern: `\d+`, Replacement: ""},
	}
	rules, err := CompileRules(configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rules[0].Replacement != "[REDACTED]" {
		t.Errorf("expected default replacement '[REDACTED]', got '%s'", rules[0].Replacement)
	}
}

func TestDefaultRuleConfigs_Compile(t *testing.T) {
	rules, err := CompileRules(defaultRuleConfigs)
	if err != nil {
		t.Fatalf("default rule configs failed to compile: %v", err)
	}
	if len(rules) != len(defaultRuleConfigs) {
		t.Errorf("expected %d rules, got %d", len(defaultRuleConfigs), len(rules))
	}
}

func TestDefaultRuleConfigs_EmailMatch(t *testing.T) {
	rules, _ := CompileRules(defaultRuleConfigs)
	var emailRule Rule
	for _, r := range rules {
		if r.Name == "email" {
			emailRule = r
			break
		}
	}
	if emailRule.Pattern == nil {
		t.Fatal("email rule not found")
	}
	input := "contact user@example.com for info"
	result := emailRule.Pattern.ReplaceAllString(input, emailRule.Replacement)
	expected := "contact [REDACTED_EMAIL] for info"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestDefaultRuleConfigs_CreditCardMatch(t *testing.T) {
	rules, _ := CompileRules(defaultRuleConfigs)
	var ccRule Rule
	for _, r := range rules {
		if r.Name == "credit_card" {
			ccRule = r
			break
		}
	}
	if ccRule.Pattern == nil {
		t.Fatal("credit_card rule not found")
	}
	input := "card number 4111111111111111 was charged"
	result := ccRule.Pattern.ReplaceAllString(input, ccRule.Replacement)
	expected := "card number [REDACTED_CC] was charged"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}
