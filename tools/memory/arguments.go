package memory

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
)

var jsonStringConcatPattern = regexp.MustCompile(`"((?:\\.|[^"\\])*)"\s*\+\s*"((?:\\.|[^"\\])*)"`)

func decodeArgs(raw string) (map[string]any, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return map[string]any{}, nil
	}
	candidates := []string{trimmed}
	if stripped := stripCodeFence(trimmed); stripped != "" && stripped != trimmed {
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
	var out map[string]any
	if err := json.Unmarshal([]byte(candidate), &out); err == nil {
		if out == nil {
			return nil, fmt.Errorf("decode tool args: top-level JSON object is required")
		}
		return out, nil
	}
	if repaired := collapseJSONStringConcats(candidate); repaired != candidate {
		if err := json.Unmarshal([]byte(repaired), &out); err == nil && out != nil {
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
	case string:
		if depth < 1 {
			return decodeArgsCandidate(stripCodeFence(typed), depth+1)
		}
	}
	return nil, fmt.Errorf("decode tool args: top-level JSON object is required")
}

func stripCodeFence(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "```") {
		return trimmed
	}
	firstNewline := strings.IndexByte(trimmed, '\n')
	lastFence := strings.LastIndex(trimmed, "```")
	if firstNewline < 0 || lastFence <= firstNewline {
		return trimmed
	}
	return strings.TrimSpace(trimmed[firstNewline+1 : lastFence])
}

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
	value, _ := params[key].(string)
	return strings.TrimSpace(value)
}

func readInt(params map[string]any, key string) int {
	switch value := params[key].(type) {
	case float64:
		return int(value)
	case json.Number:
		parsed, _ := strconv.Atoi(value.String())
		return parsed
	case int:
		return value
	case int64:
		return int(value)
	case string:
		parsed, _ := strconv.Atoi(strings.TrimSpace(value))
		return parsed
	default:
		return 0
	}
}

func readBool(params map[string]any, key string) bool {
	switch value := params[key].(type) {
	case bool:
		return value
	case string:
		normalized := strings.ToLower(strings.TrimSpace(value))
		return normalized == "true" || normalized == "1" || normalized == "yes" || normalized == "on"
	case float64:
		return value != 0
	case json.Number:
		return value.String() != "0"
	case int:
		return value != 0
	case int64:
		return value != 0
	default:
		return false
	}
}

func readOptionalBool(params map[string]any, key string) *bool {
	if _, exists := params[key]; !exists {
		return nil
	}
	value := readBool(params, key)
	return &value
}

func readStringSlice(params map[string]any, key string) []string {
	raw := params[key]
	var values []string
	switch typed := raw.(type) {
	case string:
		values = []string{typed}
	case []any:
		values = make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				values = append(values, text)
			}
		}
	}
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		key := strings.ToLower(value)
		if value == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
