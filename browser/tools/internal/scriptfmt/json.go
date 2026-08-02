// Package scriptfmt contains the model-output JSON cleanup used by Browser
// tool argument decoding. It is deliberately local to the Browser module.
package scriptfmt

import "strings"

// StripCodeFence extracts the first balanced JSON object or array from a
// possibly fenced or annotated model response. Comments outside JSON strings
// are removed before extraction.
func StripCodeFence(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	if inner, ok := firstCodeFenceBlock(value); ok {
		value = inner
	}
	value = strings.TrimLeft(value, "\uFEFF")
	value = stripJSONCommentsSafe(value)
	if out, ok := findFirstBalancedJSON(value); ok {
		return strings.TrimSpace(out)
	}
	return strings.TrimSpace(value)
}

func firstCodeFenceBlock(value string) (string, bool) {
	startMarker := strings.Index(value, "```")
	if startMarker < 0 {
		return "", false
	}
	lineEnd := strings.Index(value[startMarker+3:], "\n")
	if lineEnd < 0 {
		return "", false
	}
	start := startMarker + 3 + lineEnd + 1
	end := strings.Index(value[start:], "```")
	if end < 0 {
		return "", false
	}
	return value[start : start+end], true
}

func stripJSONCommentsSafe(value string) string {
	var out strings.Builder
	out.Grow(len(value))
	inString := false
	inLineComment := false
	inBlockComment := false
	escaped := false
	for index := 0; index < len(value); index++ {
		current := value[index]
		var next byte
		if index+1 < len(value) {
			next = value[index+1]
		}
		if inLineComment {
			if current == '\n' || current == '\r' {
				inLineComment = false
				out.WriteByte(current)
			}
			continue
		}
		if inBlockComment {
			if current == '*' && next == '/' {
				inBlockComment = false
				index++
			}
			continue
		}
		if inString {
			out.WriteByte(current)
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
			out.WriteByte(current)
		case '/':
			switch next {
			case '/':
				inLineComment = true
				index++
			case '*':
				inBlockComment = true
				index++
			default:
				out.WriteByte(current)
			}
		default:
			out.WriteByte(current)
		}
	}
	return out.String()
}

func findFirstBalancedJSON(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", false
	}
	if trimmed[0] == '{' || trimmed[0] == '[' {
		if out, ok := cutBalancedJSON(trimmed); ok {
			return out, true
		}
	}
	inString := false
	escaped := false
	for index := 0; index < len(value); index++ {
		current := value[index]
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
			if out, ok := cutBalancedJSON(value[index:]); ok {
				return out, true
			}
		}
	}
	return "", false
}

func cutBalancedJSON(value string) (string, bool) {
	if len(value) == 0 {
		return "", false
	}
	open := value[0]
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
	for index := 0; index < len(value); index++ {
		current := value[index]
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
				return value[:index+1], true
			}
		}
	}
	return "", false
}
