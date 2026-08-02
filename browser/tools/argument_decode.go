package tools

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/wsnacj/agentx-go/browser/tools/internal/scriptfmt"
)

const defaultLLMTaskTimeoutMs = 45_000

func decodeArgs(raw string) (map[string]any, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return map[string]any{}, nil
	}
	candidates := []string{trimmed}
	if stripped := strings.TrimSpace(scriptfmt.StripCodeFence(trimmed)); stripped != "" && stripped != trimmed {
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
	return nil, NewInvalidJSONToolArgumentError("", fmt.Errorf("decode tool args: %w", firstErr))
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
			nested := strings.TrimSpace(scriptfmt.StripCodeFence(typed))
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
	for index := 0; index < 16; index++ {
		next := jsonStringConcatPattern.ReplaceAllString(current, `"$1$2"`)
		if next == current {
			break
		}
		current = next
	}
	return current
}
