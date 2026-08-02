package tools

import (
	"strings"
	"unicode/utf8"
)

const maxPDFTokenCount = 128

func runeLen(text string) int { return len([]rune(text)) }

func trimToMaxChars(text string, max int) (string, bool) {
	if max <= 0 {
		return text, false
	}
	if utf8.ValidString(text) {
		runes := []rune(text)
		if len(runes) <= max {
			return text, false
		}
		return string(runes[:max]), true
	}
	if len(text) <= max {
		return text, false
	}
	return text[:max], true
}

func tokenizeQuery(query string) []string {
	normalized := strings.ToLower(strings.TrimSpace(query))
	if normalized == "" {
		return nil
	}
	out := make([]string, 0, 16)
	seen := map[string]bool{}
	add := func(token string) {
		token = strings.TrimSpace(token)
		if token == "" || seen[token] || len(out) >= maxPDFTokenCount {
			return
		}
		seen[token] = true
		out = append(out, token)
	}
	for _, part := range strings.Fields(normalized) {
		add(part)
	}
	for _, part := range strings.FieldsFunc(normalized, func(r rune) bool {
		return !isPDFAlphaNumeric(r) && !isPDFCJKRune(r)
	}) {
		add(part)
	}
	for _, segment := range pdfCJKSegments(normalized) {
		runes := []rune(segment)
		add(segment)
		for i := 0; i+2 <= len(runes); i++ {
			add(string(runes[i : i+2]))
		}
	}
	if len(out) == 0 {
		return []string{normalized}
	}
	return out
}

func pdfCJKSegments(text string) []string {
	runes := []rune(text)
	segments := make([]string, 0, 4)
	start := -1
	for index, r := range runes {
		if isPDFCJKRune(r) {
			if start < 0 {
				start = index
			}
			continue
		}
		if start >= 0 {
			segments = append(segments, string(runes[start:index]))
			start = -1
		}
	}
	if start >= 0 {
		segments = append(segments, string(runes[start:]))
	}
	return segments
}

func isPDFAlphaNumeric(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_'
}

func isPDFCJKRune(r rune) bool {
	return r >= 0x4E00 && r <= 0x9FFF || r >= 0x3400 && r <= 0x4DBF ||
		r >= 0x20000 && r <= 0x2EBEF || r >= 0xF900 && r <= 0xFAFF ||
		r >= 0x3040 && r <= 0x30FF || r >= 0xAC00 && r <= 0xD7AF
}
