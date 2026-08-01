package controlcontract

import "strings"

func stripAutoDelegationJSONFence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if inner, ok := firstAutoDelegationCodeFenceBlock(s); ok {
		s = inner
	}
	s = strings.TrimLeft(s, "\uFEFF")
	s = stripAutoDelegationJSONComments(s)
	if out, ok := findFirstAutoDelegationBalancedJSON(s); ok {
		return strings.TrimSpace(out)
	}
	return strings.TrimSpace(s)
}

func firstAutoDelegationCodeFenceBlock(s string) (string, bool) {
	i := strings.Index(s, "```")
	if i < 0 {
		return "", false
	}
	j := strings.Index(s[i+3:], "\n")
	if j < 0 {
		return "", false
	}
	start := i + 3 + j + 1
	k := strings.Index(s[start:], "```")
	if k < 0 {
		return "", false
	}
	return s[start : start+k], true
}

func stripAutoDelegationJSONComments(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inString := false
	inLineComment := false
	inBlockComment := false
	escaped := false
	for i := 0; i < len(s); i++ {
		current := s[i]
		var next byte
		if i+1 < len(s) {
			next = s[i+1]
		}
		if inLineComment {
			if current == '\n' || current == '\r' {
				inLineComment = false
				b.WriteByte(current)
			}
			continue
		}
		if inBlockComment {
			if current == '*' && next == '/' {
				inBlockComment = false
				i++
			}
			continue
		}
		if inString {
			b.WriteByte(current)
			if escaped {
				escaped = false
				continue
			}
			if current == '\\' {
				escaped = true
				continue
			}
			if current == '"' {
				inString = false
			}
			continue
		}
		switch current {
		case '"':
			inString = true
			b.WriteByte(current)
		case '/':
			switch next {
			case '/':
				inLineComment = true
				i++
			case '*':
				inBlockComment = true
				i++
			default:
				b.WriteByte(current)
			}
		default:
			b.WriteByte(current)
		}
	}
	return b.String()
}

func findFirstAutoDelegationBalancedJSON(s string) (string, bool) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return "", false
	}
	if trimmed[0] == '{' || trimmed[0] == '[' {
		if out, ok := cutAutoDelegationBalancedJSON(trimmed); ok {
			return out, true
		}
	}
	inString := false
	escaped := false
	for i := 0; i < len(s); i++ {
		current := s[i]
		if inString {
			if escaped {
				escaped = false
			} else if current == '\\' {
				escaped = true
			} else if current == '"' {
				inString = false
			}
			continue
		}
		if current == '"' {
			inString = true
			continue
		}
		if current == '{' || current == '[' {
			if out, ok := cutAutoDelegationBalancedJSON(s[i:]); ok {
				return out, true
			}
		}
	}
	return "", false
}

func cutAutoDelegationBalancedJSON(s string) (string, bool) {
	if len(s) == 0 {
		return "", false
	}
	open := s[0]
	var close byte
	switch open {
	case '{':
		close = '}'
	case '[':
		close = ']'
	default:
		return "", false
	}
	depth := 0
	inString := false
	escaped := false
	for i := 0; i < len(s); i++ {
		current := s[i]
		if inString {
			if escaped {
				escaped = false
			} else if current == '\\' {
				escaped = true
			} else if current == '"' {
				inString = false
			}
			continue
		}
		if current == '"' {
			inString = true
			continue
		}
		if current == open {
			depth++
		} else if current == close {
			depth--
			if depth == 0 {
				return s[:i+1], true
			}
		}
	}
	return "", false
}
