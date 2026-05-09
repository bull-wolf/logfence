package redactor

import "strings"

// MaskConfig controls how matched values are masked.
type MaskConfig struct {
	// ShowPrefix is the number of leading characters to preserve.
	ShowPrefix int
	// ShowSuffix is the number of trailing characters to preserve.
	ShowSuffix int
	// MaskChar is the character used for masking (default '*').
	MaskChar rune
	// MinMaskLen is the minimum number of mask characters to emit.
	MinMaskLen int
}

// DefaultMaskConfig returns a sensible default masking configuration.
func DefaultMaskConfig() MaskConfig {
	return MaskConfig{
		ShowPrefix:  0,
		ShowSuffix:  0,
		MaskChar:    '*',
		MinMaskLen:  4,
	}
}

// Mask replaces the interior of value according to cfg, preserving optional
// prefix/suffix characters and replacing the rest with MaskChar.
// If the value is too short to honour both prefix and suffix, the entire
// value is replaced with MinMaskLen mask characters.
func Mask(value string, cfg MaskConfig) string {
	if cfg.MaskChar == 0 {
		cfg.MaskChar = '*'
	}
	runes := []rune(value)
	n := len(runes)

	if n <= cfg.ShowPrefix+cfg.ShowSuffix {
		// Not enough characters — mask everything.
		maskLen := cfg.MinMaskLen
		if maskLen <= 0 {
			maskLen = 4
		}
		return strings.Repeat(string(cfg.MaskChar), maskLen)
	}

	var sb strings.Builder
	if cfg.ShowPrefix > 0 {
		sb.WriteString(string(runes[:cfg.ShowPrefix]))
	}

	maskLen := n - cfg.ShowPrefix - cfg.ShowSuffix
	if maskLen < cfg.MinMaskLen {
		maskLen = cfg.MinMaskLen
	}
	sb.WriteString(strings.Repeat(string(cfg.MaskChar), maskLen))

	if cfg.ShowSuffix > 0 {
		sb.WriteString(string(runes[n-cfg.ShowSuffix:]))
	}
	return sb.String()
}

// MaskEmail masks an email address, showing the first character of the local
// part and the domain suffix (e.g. "j****@example.com").
func MaskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return Mask(email, DefaultMaskConfig())
	}
	local := parts[0]
	domain := parts[1]
	maskedLocal := Mask(local, MaskConfig{
		ShowPrefix: 1,
		ShowSuffix: 0,
		MaskChar:   '*',
		MinMaskLen: 3,
	})
	return maskedLocal + "@" + domain
}
