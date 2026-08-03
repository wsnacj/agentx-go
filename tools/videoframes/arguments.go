package videoframes

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
)

func decodeArgs(raw string) (map[string]any, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return map[string]any{}, nil
	}
	candidates := []string{trimmed}
	if stripped := strings.TrimSpace(stripJSONPayload(trimmed)); stripped != "" && stripped != trimmed {
		candidates = append(candidates, stripped)
	}
	var firstErr error
	for _, candidate := range candidates {
		out, err := decodeArgsCandidate(candidate, 0)
		if err == nil {
			return out, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	if firstErr == nil {
		firstErr = fmt.Errorf("invalid json")
	}
	return nil, agentxtoolerrors.NewInvalidJSONToolArgumentError("", fmt.Errorf("decode tool args: %w", firstErr))
}

func decodeArgsCandidate(candidate string, depth int) (map[string]any, error) {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(candidate), &out); err == nil {
		if out == nil {
			return nil, fmt.Errorf("decode tool args: top-level JSON object is required")
		}
		return out, nil
	}
	if repaired := collapseJSONStringConcats(candidate); repaired != candidate {
		if err := json.Unmarshal([]byte(repaired), &out); err == nil {
			if out == nil {
				return nil, fmt.Errorf("decode tool args: top-level JSON object is required")
			}
			return out, nil
		}
	}
	decoder := json.NewDecoder(strings.NewReader(candidate))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	switch typed := value.(type) {
	case map[string]any:
		return typed, nil
	case []any:
		if len(typed) == 1 {
			if first, ok := typed[0].(map[string]any); ok && first != nil {
				return first, nil
			}
		}
		return nil, fmt.Errorf("decode tool args: top-level JSON object is required")
	case string:
		if depth < 1 {
			nested := strings.TrimSpace(stripJSONPayload(typed))
			if nested != "" {
				return decodeArgsCandidate(nested, depth+1)
			}
		}
		return nil, fmt.Errorf("decode tool args: top-level JSON object is required")
	default:
		return nil, fmt.Errorf("decode tool args: top-level JSON object is required")
	}
}

var jsonStringConcatPattern = regexp.MustCompile(`"((?:\\.|[^"\\])*)"\s*\+\s*"((?:\\.|[^"\\])*)"`)

func collapseJSONStringConcats(raw string) string {
	current := raw
	for i := 0; i < 16; i++ {
		next := jsonStringConcatPattern.ReplaceAllString(current, `"$1$2"`)
		if next == current {
			break
		}
		current = next
	}
	return current
}

func readString(params map[string]any, key string) string {
	if params == nil {
		return ""
	}
	value, _ := params[key].(string)
	return strings.TrimSpace(value)
}

func readInt(params map[string]any, key string) int {
	if params == nil {
		return 0
	}
	switch value := params[key].(type) {
	case float64:
		return int(value)
	case int:
		return value
	case int64:
		return int(value)
	case json.Number:
		parsed, _ := strconv.Atoi(string(value))
		return parsed
	case string:
		parsed, _ := strconv.Atoi(strings.TrimSpace(value))
		return parsed
	default:
		return 0
	}
}

func firstString(params map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := readString(params, key); value != "" {
			return value
		}
	}
	return ""
}

func firstInt(params map[string]any, keys ...string) int {
	for _, key := range keys {
		if value := readInt(params, key); value != 0 {
			return value
		}
	}
	return 0
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func intSchema(description string, minimum int) map[string]any {
	return map[string]any{"type": "integer", "description": description, "minimum": minimum}
}

func numberSchema(description string) map[string]any {
	return map[string]any{"type": "number", "description": description}
}

func stripJSONPayload(value string) string {
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
	startFence := strings.Index(value, "```")
	if startFence < 0 {
		return "", false
	}
	lineEnd := strings.Index(value[startFence+3:], "\n")
	if lineEnd < 0 {
		return "", false
	}
	start := startFence + 3 + lineEnd + 1
	end := strings.Index(value[start:], "```")
	if end < 0 {
		return "", false
	}
	return value[start : start+end], true
}

func stripJSONCommentsSafe(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))
	inString, inLine, inBlock, escaped := false, false, false, false
	for i := 0; i < len(value); i++ {
		current := value[i]
		var next byte
		if i+1 < len(value) {
			next = value[i+1]
		}
		if inLine {
			if current == '\n' || current == '\r' {
				inLine = false
				builder.WriteByte(current)
			}
			continue
		}
		if inBlock {
			if current == '*' && next == '/' {
				inBlock = false
				i++
			}
			continue
		}
		if inString {
			builder.WriteByte(current)
			if escaped {
				escaped = false
			} else if current == '\\' {
				escaped = true
			} else if current == '"' {
				inString = false
			}
			continue
		}
		switch current {
		case '"':
			inString = true
			builder.WriteByte(current)
		case '/':
			switch next {
			case '/':
				inLine = true
				i++
			case '*':
				inBlock = true
				i++
			default:
				builder.WriteByte(current)
			}
		default:
			builder.WriteByte(current)
		}
	}
	return builder.String()
}

func findFirstBalancedJSON(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed != "" && (trimmed[0] == '{' || trimmed[0] == '[') {
		if out, ok := cutBalancedJSON(trimmed); ok {
			return out, true
		}
	}
	inString, escaped := false, false
	for i := 0; i < len(value); i++ {
		current := value[i]
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
			if out, ok := cutBalancedJSON(value[i:]); ok {
				return out, true
			}
		}
	}
	return "", false
}

func cutBalancedJSON(value string) (string, bool) {
	if value == "" {
		return "", false
	}
	open := value[0]
	close := byte('}')
	if open == '[' {
		close = ']'
	} else if open != '{' {
		return "", false
	}
	depth := 0
	inString, escaped := false, false
	for i := 0; i < len(value); i++ {
		current := value[i]
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
				return value[:i+1], true
			}
		}
	}
	return "", false
}
