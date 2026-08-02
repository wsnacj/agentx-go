package tools

import (
	"strings"
	"unicode/utf8"
)

func trimToMaxChars(text string, max int) (string, bool) {
	if max <= 0 {
		return text, false
	}
	if utf8.ValidString(text) {
		if utf8.RuneCountInString(text) <= max {
			return text, false
		}
		runes := []rune(text)
		return string(runes[:max]), true
	}
	if len(text) <= max {
		return text, false
	}
	return text[:max], true
}

func truncateToolText(text string, max int) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" || max <= 0 {
		return ""
	}
	if out, truncated := trimToMaxChars(trimmed, max); !truncated {
		return out
	}
	if max <= 3 {
		out, _ := trimToMaxChars(trimmed, max)
		return out
	}
	out, _ := trimToMaxChars(trimmed, max-3)
	return strings.TrimSpace(out) + "..."
}
