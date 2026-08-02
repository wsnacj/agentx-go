package tools

import (
	"encoding/json"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

// AttemptConstrainedBrowserArgumentRepair applies only canonical, explicitly
// declared Browser repairs. Authorization and retry policy remain Host-owned.
func AttemptConstrainedBrowserArgumentRepair(toolName string, raw string, argErr *ToolArgumentError) (string, []string, bool, error) {
	if argErr == nil || !argErr.Repairable || !argErr.SafeAutorepair {
		return "", nil, false, nil
	}
	normalized := NormalizeToolName(toolName)
	if !agentxbrowserruntime.IsBrowserToolName(normalized) || normalized == "browser_runtime" {
		return "", nil, false, nil
	}
	params, err := decodeArgs(raw)
	if err != nil {
		return "", nil, false, err
	}
	if !toolSupportsConstrainedArgumentRepair(normalized, params, argErr) {
		return "", nil, false, nil
	}
	appliedKinds := constrainedBrowserArgumentRepairKinds(params, argErr)
	if len(appliedKinds) == 0 {
		return "", nil, false, nil
	}
	blob, err := json.Marshal(params)
	if err != nil {
		return "", nil, false, err
	}
	return string(blob), appliedKinds, true, nil
}

func toolSupportsConstrainedArgumentRepair(toolName string, params map[string]any, argErr *ToolArgumentError) bool {
	action := browserToolSharedActionName(toolName, params)
	if action == "" || argErr == nil {
		return false
	}
	switch strings.TrimSpace(argErr.Code) {
	case "missing_locator":
		return action == "click" || action == "type" || action == "fill" || action == "select" || action == "hover" || action == "highlight"
	case "missing_value":
		return action == "type" || action == "fill" || action == "select"
	case "invalid_fill_fields_shape", "missing_fill_input":
		return action == "fill"
	case "missing_upload_paths":
		return action == "upload"
	case "missing_key":
		return action == "press"
	case "missing_script":
		return action == "evaluate"
	case "missing_storage_entries", "invalid_storage_entries_shape":
		return action == "storage_set"
	case "missing_cookie_entries", "invalid_cookie_entries_shape":
		return action == "cookies_set"
	case "missing_headers":
		return action == "headers"
	default:
		return false
	}
}

func constrainedBrowserArgumentRepairKinds(params map[string]any, argErr *ToolArgumentError) []string {
	if params == nil || argErr == nil {
		return nil
	}
	code := strings.TrimSpace(argErr.Code)
	if code != "missing_locator" && code != "missing_value" && code != "invalid_fill_fields_shape" && code != "missing_fill_input" && code != "missing_upload_paths" && code != "missing_key" && code != "missing_script" && code != "missing_storage_entries" && code != "invalid_storage_entries_shape" && code != "missing_cookie_entries" && code != "invalid_cookie_entries_shape" && code != "missing_headers" {
		return nil
	}
	applied := make([]string, 0, len(argErr.AllowedRepairs))
	for _, repair := range argErr.AllowedRepairs {
		switch strings.TrimSpace(repair.Kind) {
		case "use_alias_field":
			if code == "missing_locator" && browserApplyAliasLocatorRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "use_declared_hint":
			if code == "missing_locator" && browserApplyDeclaredHintRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "use_alias_value":
			if (code == "missing_locator" || code == "missing_value" || code == "invalid_fill_fields_shape" || code == "missing_fill_input") && browserApplyAliasValueRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "wrap_singleton_field":
			if code == "invalid_fill_fields_shape" && browserApplyWrapSingletonFieldRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "parse_stringified_fields":
			if code == "invalid_fill_fields_shape" && browserApplyParseStringifiedFieldsRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "promote_singular_field":
			if code == "missing_fill_input" && browserApplyPromoteSingularFieldRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "use_alias_upload_path":
			if code == "missing_upload_paths" && browserApplyUploadPathAliasRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "use_alias_key":
			if code == "missing_key" && browserApplyPressKeyAliasRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "use_alias_script":
			if code == "missing_script" && browserApplyEvaluateScriptAliasRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "promote_singular_entry":
			if code == "missing_storage_entries" && browserApplyPromoteSingularEntryRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "wrap_singleton_entry":
			if code == "invalid_storage_entries_shape" && browserApplyWrapSingletonEntryRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "parse_stringified_entries":
			if code == "invalid_storage_entries_shape" && browserApplyParseStringifiedEntriesRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "promote_singular_cookie":
			if code == "missing_cookie_entries" && browserApplyPromoteSingularCookieRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "wrap_singleton_cookie":
			if code == "invalid_cookie_entries_shape" && browserApplyWrapSingletonCookieRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "parse_stringified_cookies":
			if code == "invalid_cookie_entries_shape" && browserApplyParseStringifiedCookiesRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "promote_singular_header":
			if code == "missing_headers" && browserApplyPromoteSingularHeaderRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "promote_header_json_map":
			if code == "missing_headers" && browserApplyPromoteHeaderJSONMapRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		case "parse_stringified_headers":
			if code == "missing_headers" && browserApplyParseStringifiedHeadersRepair(params, repair) {
				applied = appendUniqueRepairKind(applied, repair.Kind)
			}
		}
	}
	if len(applied) == 0 {
		return nil
	}
	return applied
}

func browserApplyAliasLocatorRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	if strings.TrimSpace(firstString(params, "selector")) != "" || strings.TrimSpace(firstString(params, "ref", "element_ref")) != "" {
		return false
	}
	value := browserSanitizeLooseArgumentString(firstString(params, strings.TrimSpace(repair.From)))
	if value == "" {
		return false
	}
	switch {
	case browserElementRefHasKnownPrefix(value):
		params["ref"] = value
		return true
	case browserLooksLikeSelector(value):
		params["selector"] = value
		return true
	default:
		return false
	}
}

func browserApplyDeclaredHintRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	if strings.TrimSpace(firstString(params, "element")) != "" {
		return false
	}
	value := browserSanitizeLooseArgumentString(firstString(params, strings.TrimSpace(repair.From)))
	if value == "" {
		return false
	}
	params["element"] = value
	return true
}

func browserApplyAliasValueRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	if strings.TrimSpace(firstString(params, "value")) == "" && len(readStringList(params, "values")) == 0 {
		value := browserSanitizeLooseArgumentString(firstString(params, strings.TrimSpace(repair.From)))
		if value != "" {
			params["value"] = value
			return true
		}
		values := readStringList(params, strings.TrimSpace(repair.From))
		if len(values) > 0 {
			params["values"] = append([]string(nil), values...)
			return true
		}
	}
	items, ok := params["fields"].([]any)
	if !ok {
		return false
	}
	changed := false
	for _, item := range items {
		field, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if strings.TrimSpace(firstString(field, "value")) != "" || len(readStringList(field, "values")) > 0 {
			continue
		}
		value := browserSanitizeLooseArgumentString(firstString(field, strings.TrimSpace(repair.From)))
		if value != "" {
			field["value"] = value
			changed = true
			continue
		}
		values := readStringList(field, strings.TrimSpace(repair.From))
		if len(values) > 0 {
			field["values"] = append([]string(nil), values...)
			changed = true
			continue
		}
	}
	return changed
}

func browserApplyWrapSingletonFieldRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	raw, ok := params[strings.TrimSpace(repair.From)]
	if !ok || raw == nil {
		return false
	}
	if _, ok := raw.([]any); ok {
		return false
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return false
	}
	params[strings.TrimSpace(repair.From)] = []any{obj}
	return true
}

func browserApplyParseStringifiedFieldsRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	items, ok := browserParseStringifiedFillFields(firstString(params, strings.TrimSpace(repair.From)))
	if !ok || len(items) == 0 {
		return false
	}
	out := make([]any, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	params[strings.TrimSpace(repair.From)] = out
	return true
}

func browserApplyPromoteSingularFieldRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	if raw := params["fields"]; raw != nil {
		if items, ok := raw.([]any); ok && len(items) > 0 {
			return false
		}
	}
	raw, ok := params[strings.TrimSpace(repair.From)]
	if !ok || raw == nil {
		return false
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		text, isText := raw.(string)
		if !isText {
			return false
		}
		parsed, ok := browserParseSingleStringifiedObject(text)
		if !ok {
			return false
		}
		obj = parsed
	}
	params["fields"] = []any{obj}
	delete(params, strings.TrimSpace(repair.From))
	return true
}

func browserApplyUploadPathAliasRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	if len(readStringList(params, "paths")) > 0 || len(readStringList(params, "files")) > 0 || strings.TrimSpace(firstString(params, "file")) != "" {
		return false
	}
	value := browserSanitizeLooseArgumentString(firstString(params, strings.TrimSpace(repair.From)))
	if value == "" {
		return false
	}
	params["file"] = value
	return true
}

func browserApplyPressKeyAliasRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	if strings.TrimSpace(firstString(params, "key")) != "" {
		return false
	}
	value := browserSanitizeLooseArgumentString(firstString(params, strings.TrimSpace(repair.From)))
	if value == "" {
		return false
	}
	params["key"] = value
	return true
}

func browserApplyEvaluateScriptAliasRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	if strings.TrimSpace(firstString(params, "script")) != "" {
		return false
	}
	value := browserSanitizeLooseArgumentString(firstString(params, strings.TrimSpace(repair.From)))
	if value == "" {
		return false
	}
	params["script"] = value
	return true
}

func browserApplyPromoteSingularEntryRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	if raw := params["entries"]; raw != nil {
		if items, ok := raw.([]any); ok && len(items) > 0 {
			return false
		}
	}
	raw, ok := params[strings.TrimSpace(repair.From)]
	if !ok || raw == nil {
		return false
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		text, isText := raw.(string)
		if !isText {
			return false
		}
		parsed, ok := browserParseSingleStringifiedObject(text)
		if !ok {
			return false
		}
		obj = parsed
	}
	params["entries"] = []any{obj}
	delete(params, strings.TrimSpace(repair.From))
	return true
}

func browserApplyWrapSingletonEntryRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	raw, ok := params[strings.TrimSpace(repair.From)]
	if !ok || raw == nil {
		return false
	}
	if _, ok := raw.([]any); ok {
		return false
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return false
	}
	params[strings.TrimSpace(repair.From)] = []any{obj}
	return true
}

func browserApplyParseStringifiedEntriesRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	parsed, ok := browserParseStringifiedStorageEntries(firstString(params, strings.TrimSpace(repair.From)))
	if !ok || len(parsed) == 0 {
		return false
	}
	items := make([]any, 0, len(parsed))
	for _, item := range parsed {
		if strings.TrimSpace(firstString(item, "key")) == "" {
			return false
		}
		items = append(items, item)
	}
	params[strings.TrimSpace(repair.From)] = items
	return true
}

func browserApplyPromoteSingularCookieRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	if raw := params["cookies"]; raw != nil {
		if items, ok := raw.([]any); ok && len(items) > 0 {
			return false
		}
	}
	raw, ok := params[strings.TrimSpace(repair.From)]
	if !ok || raw == nil {
		return false
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		text, isText := raw.(string)
		if !isText {
			return false
		}
		parsed, ok := browserParseSingleStringifiedObject(text)
		if !ok {
			return false
		}
		obj = parsed
	}
	params["cookies"] = []any{obj}
	delete(params, strings.TrimSpace(repair.From))
	return true
}

func browserApplyWrapSingletonCookieRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	raw, ok := params[strings.TrimSpace(repair.From)]
	if !ok || raw == nil {
		return false
	}
	if _, ok := raw.([]any); ok {
		return false
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return false
	}
	params[strings.TrimSpace(repair.From)] = []any{obj}
	return true
}

func browserApplyParseStringifiedCookiesRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	parsed, ok := browserParseStringifiedCookieEntries(firstString(params, strings.TrimSpace(repair.From)))
	if !ok || len(parsed) == 0 {
		return false
	}
	items := make([]any, 0, len(parsed))
	for _, item := range parsed {
		if strings.TrimSpace(firstString(item, "name")) == "" {
			return false
		}
		items = append(items, item)
	}
	params[strings.TrimSpace(repair.From)] = items
	return true
}

func browserApplyPromoteSingularHeaderRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	if len(readStringMap(params["headers"])) > 0 {
		return false
	}
	raw, ok := params[strings.TrimSpace(repair.From)]
	if !ok || raw == nil {
		return false
	}
	if len(readStringMap(raw)) == 0 {
		return false
	}
	params["headers"] = raw
	delete(params, strings.TrimSpace(repair.From))
	return true
}

func browserApplyPromoteHeaderJSONMapRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	if len(readStringMap(params["headers"])) > 0 {
		return false
	}
	raw, ok := params[strings.TrimSpace(repair.From)]
	if !ok || raw == nil {
		return false
	}
	if len(readStringMap(raw)) == 0 {
		return false
	}
	params["headers"] = raw
	delete(params, strings.TrimSpace(repair.From))
	return true
}

func browserApplyParseStringifiedHeadersRepair(params map[string]any, repair ToolArgumentRepair) bool {
	if params == nil {
		return false
	}
	if len(readStringMap(params["headers"])) > 0 {
		return false
	}
	raw := strings.TrimSpace(firstString(params, strings.TrimSpace(repair.From)))
	if raw == "" {
		return false
	}
	parsed := map[string]string{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil || len(parsed) == 0 {
		return false
	}
	params["headers"] = parsed
	return true
}

func appendUniqueRepairKind(base []string, kind string) []string {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return base
	}
	for _, item := range base {
		if strings.TrimSpace(item) == kind {
			return base
		}
	}
	return append(base, kind)
}
