package tools

import (
	"sort"
	"strconv"
	"strings"
)

func firstNonEmpty(parts ...string) string {
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			return trimmed
		}
	}
	return ""
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

func readString(params map[string]any, key string) string {
	if params == nil {
		return ""
	}
	value, ok := params[key].(string)
	if !ok {
		return ""
	}
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
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err == nil {
			return parsed
		}
	}
	return 0
}

func readStringList(params map[string]any, key string) []string {
	if params == nil {
		return nil
	}
	raw, ok := params[key]
	if !ok || raw == nil {
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
	switch value := raw.(type) {
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
	if len(out) == 0 {
		return nil
	}
	return out
}

func readStringMap(raw any) map[string]string {
	if raw == nil {
		return nil
	}
	entries, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]string{}
	for key, value := range entries {
		name := strings.TrimSpace(key)
		if name == "" {
			continue
		}
		switch casted := value.(type) {
		case string:
			out[name] = casted
		case float64:
			out[name] = strconv.FormatFloat(casted, 'f', -1, 64)
		case int:
			out[name] = strconv.Itoa(casted)
		case int64:
			out[name] = strconv.FormatInt(casted, 10)
		case bool:
			out[name] = strconv.FormatBool(casted)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func mergeToolMetadataStrings(current, next []string) []string {
	if len(next) == 0 {
		if len(current) == 0 {
			return nil
		}
		return append([]string(nil), current...)
	}
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
	if len(out) == 0 {
		return nil
	}
	return out
}

func containsString(items []string, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item), target) {
			return true
		}
	}
	return false
}

func buildEnabledToolSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	out := map[string]bool{}
	for _, item := range items {
		if name := NormalizeToolName(item); name != "" {
			out[name] = true
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func toolEnabled(allowed map[string]bool, name string) bool {
	if len(allowed) == 0 {
		return true
	}
	normalized := NormalizeToolName(name)
	return normalized != "" && allowed[normalized]
}

func normalizeToolList(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		normalized := NormalizeToolName(item)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	if len(out) == 0 {
		return nil
	}
	sort.Strings(out)
	return out
}

func mergeToolMetadataItem(current, next ToolMetadata) ToolMetadata {
	if plugin := NormalizeToolName(next.Plugin); plugin != "" {
		current.Plugin = plugin
	}
	if len(next.Groups) > 0 {
		current.Groups = mergeToolMetadataStrings(current.Groups, next.Groups)
	}
	if toolType := strings.ToLower(strings.TrimSpace(next.Type)); toolType != "" {
		current.Type = toolType
	}
	if source := strings.ToLower(strings.TrimSpace(next.Source)); source != "" && source != "unknown" {
		current.Source = source
	}
	if len(next.Capabilities) > 0 {
		current.Capabilities = mergeToolMetadataStrings(current.Capabilities, next.Capabilities)
	}
	if len(next.AuditTags) > 0 {
		current.AuditTags = mergeToolMetadataStrings(current.AuditTags, next.AuditTags)
	}
	if riskProfile := strings.ToLower(strings.TrimSpace(next.RiskProfile)); riskProfile != "" {
		current.RiskProfile = riskProfile
	}
	if next.ReadOnly != nil {
		current.ReadOnly = builtinToolMetadataBoolPtr(*next.ReadOnly)
	}
	if next.ConcurrencySafe != nil {
		current.ConcurrencySafe = builtinToolMetadataBoolPtr(*next.ConcurrencySafe)
	}
	if next.Destructive != nil {
		current.Destructive = builtinToolMetadataBoolPtr(*next.Destructive)
	}
	return current
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
