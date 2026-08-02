package tools

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

type ToolMetadata = agentxtools.ToolMetadata

const ToolSourceBuiltin = agentxtools.ToolSourceBuiltin

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
	return nil, agentxtoolerrors.NewInvalidJSONToolArgumentError("", fmt.Errorf("decode tool args: %w", firstErr))
}

func decodeArgsCandidate(candidate string, depth int) (map[string]any, error) {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(candidate), &out); err == nil && out != nil {
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
	for index := 0; index < 16; index++ {
		next := jsonStringConcatPattern.ReplaceAllString(current, `"$1$2"`)
		if next == current {
			break
		}
		current = next
	}
	return current
}

func firstString(params map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := readString(params, key); value != "" {
			return value
		}
	}
	return ""
}

func readString(params map[string]any, key string) string {
	if params == nil {
		return ""
	}
	value, _ := params[key].(string)
	return strings.TrimSpace(value)
}

func firstInt(params map[string]any, keys ...string) int {
	for _, key := range keys {
		if value := readInt(params, key); value != 0 {
			return value
		}
	}
	return 0
}

func readInt(params map[string]any, key string) int {
	if params == nil {
		return 0
	}
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
	if params == nil {
		return false
	}
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

func readFloat(params map[string]any, key string) (float64, bool) {
	if params == nil {
		return 0, false
	}
	switch value := params[key].(type) {
	case float64:
		return value, !math.IsNaN(value) && !math.IsInf(value, 0)
	case float32:
		result := float64(value)
		return result, !math.IsNaN(result) && !math.IsInf(result, 0)
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	case json.Number:
		result, err := value.Float64()
		return result, err == nil
	default:
		return 0, false
	}
}

func readStringList(params map[string]any, key string) []string {
	if params == nil {
		return nil
	}
	out := make([]string, 0)
	seen := map[string]bool{}
	appendItem := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		out = append(out, value)
	}
	switch value := params[key].(type) {
	case string:
		for _, item := range strings.Split(value, ",") {
			appendItem(item)
		}
	case []string:
		for _, item := range value {
			appendItem(item)
		}
	case []any:
		for _, item := range value {
			if text, ok := item.(string); ok {
				appendItem(text)
			}
		}
	}
	return out
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func newMissingRequiredToolArgumentError(tool string, fields []string, detail string) error {
	return agentxtoolerrors.NewMissingRequiredToolArgumentError(tool, fields, detail)
}

func toolUserConversation(chunks ...string) llm.Conversation {
	out := make(llm.Conversation, 0, len(chunks))
	for _, chunk := range chunks {
		if strings.TrimSpace(chunk) != "" {
			out = append(out, llm.Message{Role: "user", Content: chunk})
		}
	}
	return out
}

func truncateToolText(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || runeLen(value) <= max {
		return value
	}
	if max <= 3 {
		result, _ := trimToMaxChars(value, max)
		return result
	}
	result, _ := trimToMaxChars(value, max-3)
	return strings.TrimSpace(result) + "..."
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}
func boolSchema(description string) map[string]any {
	return map[string]any{"type": "boolean", "description": description}
}
func intSchema(description string, minimum int) map[string]any {
	return map[string]any{"type": "integer", "description": description, "minimum": minimum}
}
func numberSchema(description string) map[string]any {
	return map[string]any{"type": "number", "description": description}
}
func stringArraySchema(description string) map[string]any {
	return map[string]any{"type": "array", "description": description, "items": map[string]any{"type": "string"}}
}
func stringEnumSchema(description string, values ...string) map[string]any {
	return map[string]any{"type": "string", "description": description, "enum": append([]string(nil), values...)}
}
func looseObjectSchema(description string) map[string]any {
	return map[string]any{"type": "object", "description": description, "additionalProperties": true}
}
func looseObjectArraySchema(description string) map[string]any {
	return map[string]any{"type": "array", "description": description, "items": map[string]any{"type": "object", "additionalProperties": true}}
}
func objectArraySchema(description string, properties map[string]any) map[string]any {
	return map[string]any{"type": "array", "description": description, "items": map[string]any{"type": "object", "additionalProperties": false, "properties": properties}}
}
func closedInputSchema(properties map[string]any, required []string) map[string]any {
	return closedObjectSchema(properties, required)
}
func closedOutputSchema(properties map[string]any, required []string) map[string]any {
	return closedObjectSchema(properties, required)
}
func closedObjectSchema(properties map[string]any, required []string) map[string]any {
	out := map[string]any{"type": "object", "additionalProperties": false, "properties": properties}
	if len(required) > 0 {
		out["required"] = append([]string(nil), required...)
	}
	return out
}

func builtinToolMetadataBoolPtr(value bool) *bool { return &value }

func mergeToolMetadataStrings(current []string, next []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(current)+len(next))
	for _, values := range [][]string{current, next} {
		for _, raw := range values {
			value := strings.ToLower(strings.TrimSpace(raw))
			if value == "" || seen[value] {
				continue
			}
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}

func IsBuiltinCoreTool(name string) bool { return isPDFToolName(name) }

func buildEnabledToolSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	out := map[string]bool{}
	for _, item := range items {
		if name := agentxtools.NormalizeToolName(item); name != "" {
			out[name] = true
		}
	}
	return out
}

func toolEnabled(allowed map[string]bool, name string) bool {
	if len(allowed) == 0 {
		return true
	}
	return allowed[agentxtools.NormalizeToolName(name)]
}

// OCR profile discovery belongs to Host. Canonical cache identity therefore
// uses the explicit opaque profile exactly as supplied by construction.
func resolveOCRXConfig(profile string) string { return strings.TrimSpace(profile) }

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func isValidImageToolSchemeToken(scheme string) bool {
	if scheme == "" {
		return false
	}
	for index, r := range scheme {
		switch {
		case index == 0 && ((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')):
		case index > 0 && ((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '+' || r == '-' || r == '.'):
		default:
			return false
		}
	}
	return true
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
