package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

func browserMissingActKindError(toolName string, action string) error {
	detail := strings.TrimSpace(toolName) + ": kind is required"
	if strings.TrimSpace(action) != "" {
		detail = fmt.Sprintf("%s: kind is required when action=%s", strings.TrimSpace(toolName), strings.TrimSpace(action))
	}
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:          "missing_kind",
		Detail:        detail,
		MissingFields: []string{"kind"},
	})
}

func browserMissingURLForKindError(toolName string, kind string) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:          "missing_url",
		Detail:        fmt.Sprintf("%s: url is required for kind %s", strings.TrimSpace(toolName), strings.TrimSpace(kind)),
		MissingFields: []string{"url"},
	})
}

func browserMissingPressKeyErrorWithRepair(toolName string, repairable bool, safeAutorepair bool, repairs []ToolArgumentRepair) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:           "missing_key",
		Detail:         fmt.Sprintf("%s: key is required for kind press", strings.TrimSpace(toolName)),
		Repairable:     repairable,
		SafeAutorepair: safeAutorepair,
		MissingFields:  []string{"key"},
		AllowedRepairs: repairs,
	})
}

func browserMissingScriptErrorWithRepair(toolName string, repairable bool, safeAutorepair bool, repairs []ToolArgumentRepair) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:           "missing_script",
		Detail:         fmt.Sprintf("%s: script is required for kind evaluate", strings.TrimSpace(toolName)),
		Repairable:     repairable,
		SafeAutorepair: safeAutorepair,
		MissingFields:  []string{"script"},
		AllowedRepairs: repairs,
	})
}

func browserMissingLocatorError(toolName string, detail string, repairable bool, safeAutorepair bool, repairs []ToolArgumentRepair) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:           "missing_locator",
		Detail:         strings.TrimSpace(detail),
		Repairable:     repairable,
		SafeAutorepair: safeAutorepair,
		MissingFields:  []string{"selector_or_ref"},
		AllowedRepairs: repairs,
	})
}

func browserMissingValueError(toolName string, kind string) error {
	return browserMissingValueErrorWithRepair(toolName, kind, false, false, nil)
}

func browserMissingValueErrorWithRepair(toolName string, kind string, repairable bool, safeAutorepair bool, repairs []ToolArgumentRepair) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:           "missing_value",
		Detail:         fmt.Sprintf("%s: value or values is required for kind %s", strings.TrimSpace(toolName), strings.TrimSpace(kind)),
		Repairable:     repairable,
		SafeAutorepair: safeAutorepair,
		MissingFields:  []string{"value_or_values"},
		AllowedRepairs: repairs,
	})
}

func browserMissingTypeTextError(toolName string) error {
	return browserMissingTypeTextErrorWithRepair(toolName, false, false, nil)
}

func browserMissingTypeTextErrorWithRepair(toolName string, repairable bool, safeAutorepair bool, repairs []ToolArgumentRepair) error {
	detail := fmt.Sprintf("%s: text or value is required", strings.TrimSpace(toolName))
	if strings.TrimSpace(toolName) == "browser_act" {
		detail = "browser_act: text or value is required for kind type"
	}
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:           "missing_value",
		Detail:         detail,
		Repairable:     repairable,
		SafeAutorepair: safeAutorepair,
		MissingFields:  []string{"text_or_value"},
		AllowedRepairs: repairs,
	})
}

func browserMissingFillInputError(toolName string) error {
	return browserMissingFillInputErrorWithRepair(toolName, false, false, nil)
}

func browserMissingFillInputErrorWithRepair(toolName string, repairable bool, safeAutorepair bool, repairs []ToolArgumentRepair) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:           "missing_fill_input",
		Detail:         fmt.Sprintf("%s: fields or selector/ref plus value is required for kind fill", strings.TrimSpace(toolName)),
		Repairable:     repairable,
		SafeAutorepair: safeAutorepair,
		MissingFields:  []string{"fields_or_locator_plus_value"},
		AllowedRepairs: repairs,
	})
}

func browserMissingStorageEntriesErrorWithRepair(toolName string, repairable bool, safeAutorepair bool, repairs []ToolArgumentRepair) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:           "missing_storage_entries",
		Detail:         fmt.Sprintf("%s: kind storage_set requires key/value or entries[]", strings.TrimSpace(toolName)),
		Repairable:     repairable,
		SafeAutorepair: safeAutorepair,
		MissingFields:  []string{"key_value_or_entries"},
		AllowedRepairs: repairs,
	})
}

func browserMissingCookieEntriesErrorWithRepair(toolName string, repairable bool, safeAutorepair bool, repairs []ToolArgumentRepair) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:           "missing_cookie_entries",
		Detail:         fmt.Sprintf("%s: kind cookies_set requires cookie fields or cookies[]", strings.TrimSpace(toolName)),
		Repairable:     repairable,
		SafeAutorepair: safeAutorepair,
		MissingFields:  []string{"cookie_fields_or_cookies"},
		AllowedRepairs: repairs,
	})
}

func browserMissingHeadersErrorWithRepair(toolName string, repairable bool, safeAutorepair bool, repairs []ToolArgumentRepair) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:           "missing_headers",
		Detail:         fmt.Sprintf("%s: kind headers requires headers, headers_json, or clear=true", strings.TrimSpace(toolName)),
		Repairable:     repairable,
		SafeAutorepair: safeAutorepair,
		MissingFields:  []string{"headers_or_headers_json_or_clear"},
		AllowedRepairs: repairs,
	})
}

func browserMissingUploadPathsErrorWithRepair(toolName string, repairable bool, safeAutorepair bool, repairs []ToolArgumentRepair) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:           "missing_upload_paths",
		Detail:         fmt.Sprintf("%s: paths/files/file is required for kind upload", strings.TrimSpace(toolName)),
		Repairable:     repairable,
		SafeAutorepair: safeAutorepair,
		MissingFields:  []string{"file_or_paths"},
		AllowedRepairs: repairs,
	})
}

func browserInvalidCookieEntriesShapeError(toolName string, repairable bool, safeAutorepair bool, repairs []ToolArgumentRepair) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:           "invalid_cookie_entries_shape",
		Detail:         fmt.Sprintf("%s: cookies must be an array of objects", strings.TrimSpace(toolName)),
		Repairable:     repairable,
		SafeAutorepair: safeAutorepair,
		MissingFields:  []string{"cookies_array"},
		AllowedRepairs: repairs,
	})
}

func browserInvalidStorageEntriesShapeError(toolName string, repairable bool, safeAutorepair bool, repairs []ToolArgumentRepair) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:           "invalid_storage_entries_shape",
		Detail:         fmt.Sprintf("%s: entries must be an array of objects", strings.TrimSpace(toolName)),
		Repairable:     repairable,
		SafeAutorepair: safeAutorepair,
		MissingFields:  []string{"entries_array"},
		AllowedRepairs: repairs,
	})
}

func browserInvalidFillFieldsShapeError(toolName string, repairable bool, safeAutorepair bool, repairs []ToolArgumentRepair) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:           "invalid_fill_fields_shape",
		Detail:         fmt.Sprintf("%s: fields must be an array of objects for kind fill", strings.TrimSpace(toolName)),
		Repairable:     repairable,
		SafeAutorepair: safeAutorepair,
		MissingFields:  []string{"fields_array"},
		AllowedRepairs: repairs,
	})
}

func browserInvalidSnapshotModeError(toolName string) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:   "invalid_snapshot_mode",
		Detail: fmt.Sprintf("%s: mode must be efficient for kind snapshot", strings.TrimSpace(toolName)),
	})
}

func browserInvalidSnapshotRefsError(toolName string) error {
	return NewToolArgumentError(toolName, ToolArgumentErrorOptions{
		Code:   "invalid_snapshot_refs",
		Detail: fmt.Sprintf("%s: refs must be aria or role for kind snapshot", strings.TrimSpace(toolName)),
	})
}

func browserMissingRequiredArgumentError(toolName string, fields []string, detail string) error {
	return NewMissingRequiredToolArgumentError(toolName, fields, detail)
}

func browserInvalidArgumentError(toolName string, fields []string, detail string) error {
	return NewInvalidToolArgumentError(toolName, fields, detail)
}

func browserRepairAdviceKinds(items []ToolArgumentRepair) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		kind := strings.TrimSpace(item.Kind)
		if kind == "" || seen[kind] {
			continue
		}
		seen[kind] = true
		out = append(out, kind)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func browserLocatorRepairAdviceFromParams(params map[string]any, fields ...string) (bool, bool, []ToolArgumentRepair) {
	for _, field := range fields {
		if strings.TrimSpace(firstString(params, field)) == "" {
			continue
		}
		repairs := []ToolArgumentRepair{{
			Kind: "use_alias_field",
			From: strings.TrimSpace(field),
			To:   "selector_or_ref",
		}}
		switch strings.TrimSpace(field) {
		case "element", "text", "label":
			repairs = append(repairs, ToolArgumentRepair{
				Kind: "use_declared_hint",
				From: strings.TrimSpace(field),
				To:   "label_hint",
			})
		}
		return true, true, repairs
	}
	return false, false, nil
}

func browserValueRepairAdviceFromParams(params map[string]any, fields ...string) (bool, bool, []ToolArgumentRepair) {
	for _, field := range fields {
		if strings.TrimSpace(firstString(params, field)) == "" && len(readStringList(params, field)) == 0 {
			continue
		}
		return true, true, []ToolArgumentRepair{{
			Kind: "use_alias_value",
			From: strings.TrimSpace(field),
			To:   "value_or_values",
		}}
	}
	return false, false, nil
}

func browserSelectValueRepairAdviceFromParams(params map[string]any) (bool, bool, []ToolArgumentRepair) {
	return browserValueRepairAdviceFromParams(params, "text", "content", "option", "options")
}

func browserUploadPathRepairAdviceFromParams(params map[string]any, fields ...string) (bool, bool, []ToolArgumentRepair) {
	for _, field := range fields {
		if strings.TrimSpace(firstString(params, field)) == "" {
			continue
		}
		return true, true, []ToolArgumentRepair{{
			Kind: "use_alias_upload_path",
			From: strings.TrimSpace(field),
			To:   "file_or_paths",
		}}
	}
	return false, false, nil
}

func browserTypeValueRepairAdviceFromParams(params map[string]any) (bool, bool, []ToolArgumentRepair) {
	return browserValueRepairAdviceFromParams(params, "content")
}

func browserFillValueRepairAdviceFromParams(params map[string]any) (bool, bool, []ToolArgumentRepair) {
	return browserValueRepairAdviceFromParams(params, "text", "content")
}

func browserPressKeyRepairAdviceFromParams(params map[string]any) (bool, bool, []ToolArgumentRepair) {
	repairs := make([]ToolArgumentRepair, 0, 2)
	for _, field := range []string{"key_name", "keyName"} {
		if strings.TrimSpace(firstString(params, field)) == "" {
			continue
		}
		repairs = append(repairs, ToolArgumentRepair{
			Kind: "use_alias_key",
			From: field,
			To:   "key",
		})
	}
	if len(repairs) == 0 {
		return false, false, nil
	}
	return true, true, repairs
}

func browserEvaluateScriptRepairAdviceFromParams(params map[string]any) (bool, bool, []ToolArgumentRepair) {
	repairs := make([]ToolArgumentRepair, 0, 2)
	for _, field := range []string{"code", "source"} {
		if strings.TrimSpace(firstString(params, field)) == "" {
			continue
		}
		repairs = append(repairs, ToolArgumentRepair{
			Kind: "use_alias_script",
			From: field,
			To:   "script",
		})
	}
	if len(repairs) == 0 {
		return false, false, nil
	}
	return true, true, repairs
}

func browserFillFieldsShapeRepairAdvice(raw any) (bool, bool, []ToolArgumentRepair) {
	text, ok := raw.(string)
	if ok && strings.TrimSpace(text) != "" {
		repairs := []ToolArgumentRepair{{
			Kind: "parse_stringified_fields",
			From: "fields",
			To:   "fields[]",
		}}
		if parsed, ok := browserParseStringifiedFillFields(text); ok {
			if repairable, safeAutorepair, valueRepairs := browserFillValueRepairAdviceFromFieldList(parsed); repairable && safeAutorepair {
				repairs = append(repairs, valueRepairs...)
			}
			return true, true, repairs
		}
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return false, false, nil
	}
	repairs := []ToolArgumentRepair{{
		Kind: "wrap_singleton_field",
		From: "fields",
		To:   "fields[]",
	}}
	if repairable, safeAutorepair, valueRepairs := browserFillValueRepairAdviceFromParams(obj); repairable && safeAutorepair {
		repairs = append(repairs, valueRepairs...)
	}
	return true, true, repairs
}

func browserFillFieldAliasRepairAdvice(raw any) (bool, bool, []ToolArgumentRepair) {
	obj, ok := raw.(map[string]any)
	if !ok {
		text, isText := raw.(string)
		if !isText {
			return false, false, nil
		}
		parsed, ok := browserParseSingleStringifiedObject(text)
		if !ok {
			return false, false, nil
		}
		obj = parsed
	}
	repairs := []ToolArgumentRepair{{
		Kind: "promote_singular_field",
		From: "field",
		To:   "fields[]",
	}}
	if repairable, safeAutorepair, valueRepairs := browserFillValueRepairAdviceFromParams(obj); repairable && safeAutorepair {
		repairs = append(repairs, valueRepairs...)
	}
	return true, true, repairs
}

func browserFillValueRepairAdviceFromFieldList(items []map[string]any) (bool, bool, []ToolArgumentRepair) {
	for _, item := range items {
		if repairable, safeAutorepair, repairs := browserFillValueRepairAdviceFromParams(item); repairable && safeAutorepair {
			return true, true, repairs
		}
	}
	return false, false, nil
}

func browserParseStringifiedFillFields(raw string) ([]map[string]any, bool) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil, false
	}
	var list []map[string]any
	if err := json.Unmarshal([]byte(text), &list); err == nil && len(list) > 0 {
		return list, true
	}
	var single map[string]any
	if err := json.Unmarshal([]byte(text), &single); err == nil && len(single) > 0 {
		return []map[string]any{single}, true
	}
	return nil, false
}

func browserParseStringifiedStorageEntries(raw string) ([]map[string]any, bool) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil, false
	}
	var list []map[string]any
	if err := json.Unmarshal([]byte(text), &list); err == nil && len(list) > 0 {
		return list, true
	}
	var single map[string]any
	if err := json.Unmarshal([]byte(text), &single); err == nil && len(single) > 0 {
		return []map[string]any{single}, true
	}
	return nil, false
}

func browserParseStringifiedCookieEntries(raw string) ([]map[string]any, bool) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil, false
	}
	var list []map[string]any
	if err := json.Unmarshal([]byte(text), &list); err == nil && len(list) > 0 {
		return list, true
	}
	var single map[string]any
	if err := json.Unmarshal([]byte(text), &single); err == nil && len(single) > 0 {
		return []map[string]any{single}, true
	}
	return nil, false
}

func browserParseSingleStringifiedObject(raw string) (map[string]any, bool) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil, false
	}
	var single map[string]any
	if err := json.Unmarshal([]byte(text), &single); err == nil && len(single) > 0 {
		return single, true
	}
	return nil, false
}

func browserStorageEntryAliasRepairAdvice(raw any) (bool, bool, []ToolArgumentRepair) {
	obj, ok := raw.(map[string]any)
	if !ok {
		text, isText := raw.(string)
		if !isText {
			return false, false, nil
		}
		parsed, ok := browserParseSingleStringifiedObject(text)
		if !ok {
			return false, false, nil
		}
		obj = parsed
	}
	if strings.TrimSpace(firstString(obj, "key")) == "" {
		return false, false, nil
	}
	return true, true, []ToolArgumentRepair{{
		Kind: "promote_singular_entry",
		From: "entry",
		To:   "entries[]",
	}}
}

func browserStorageEntriesShapeRepairAdvice(raw any) (bool, bool, []ToolArgumentRepair) {
	text, ok := raw.(string)
	if ok && strings.TrimSpace(text) != "" {
		if parsed, ok := browserParseStringifiedStorageEntries(text); ok && len(parsed) > 0 {
			if strings.TrimSpace(firstString(parsed[0], "key")) != "" {
				return true, true, []ToolArgumentRepair{{
					Kind: "parse_stringified_entries",
					From: "entries",
					To:   "entries[]",
				}}
			}
		}
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return false, false, nil
	}
	if strings.TrimSpace(firstString(obj, "key")) == "" {
		return false, false, nil
	}
	return true, true, []ToolArgumentRepair{{
		Kind: "wrap_singleton_entry",
		From: "entries",
		To:   "entries[]",
	}}
}

func browserCookieEntryAliasRepairAdvice(raw any) (bool, bool, []ToolArgumentRepair) {
	obj, ok := raw.(map[string]any)
	if !ok {
		text, isText := raw.(string)
		if !isText {
			return false, false, nil
		}
		parsed, ok := browserParseSingleStringifiedObject(text)
		if !ok {
			return false, false, nil
		}
		obj = parsed
	}
	if strings.TrimSpace(firstString(obj, "name")) == "" {
		return false, false, nil
	}
	return true, true, []ToolArgumentRepair{{
		Kind: "promote_singular_cookie",
		From: "cookie",
		To:   "cookies[]",
	}}
}

func browserCookieEntriesShapeRepairAdvice(raw any) (bool, bool, []ToolArgumentRepair) {
	text, ok := raw.(string)
	if ok && strings.TrimSpace(text) != "" {
		if parsed, ok := browserParseStringifiedCookieEntries(text); ok && len(parsed) > 0 {
			if strings.TrimSpace(firstString(parsed[0], "name")) != "" {
				return true, true, []ToolArgumentRepair{{
					Kind: "parse_stringified_cookies",
					From: "cookies",
					To:   "cookies[]",
				}}
			}
		}
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return false, false, nil
	}
	if strings.TrimSpace(firstString(obj, "name")) == "" {
		return false, false, nil
	}
	return true, true, []ToolArgumentRepair{{
		Kind: "wrap_singleton_cookie",
		From: "cookies",
		To:   "cookies[]",
	}}
}

func browserHeaderAliasRepairAdvice(raw any) (bool, bool, []ToolArgumentRepair) {
	if len(readStringMap(raw)) == 0 {
		return false, false, nil
	}
	return true, true, []ToolArgumentRepair{{
		Kind: "promote_singular_header",
		From: "header",
		To:   "headers",
	}}
}

func browserHeaderJSONMapRepairAdvice(raw any) (bool, bool, []ToolArgumentRepair) {
	if len(readStringMap(raw)) == 0 {
		return false, false, nil
	}
	return true, true, []ToolArgumentRepair{{
		Kind: "promote_header_json_map",
		From: "headers_json",
		To:   "headers",
	}}
}

func browserStringifiedHeadersRepairAdvice(raw any) (bool, bool, []ToolArgumentRepair) {
	text, ok := raw.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return false, false, nil
	}
	parsed := map[string]string{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(text)), &parsed); err != nil || len(parsed) == 0 {
		return false, false, nil
	}
	return true, true, []ToolArgumentRepair{{
		Kind: "parse_stringified_headers",
		From: "headers",
		To:   "headers",
	}}
}

func browserHeadersRepairAdviceFromParams(params map[string]any) (bool, bool, []ToolArgumentRepair) {
	repairable := false
	safe := false
	repairs := make([]ToolArgumentRepair, 0, 2)
	if ok, s, items := browserHeaderAliasRepairAdvice(params["header"]); ok {
		repairable = true
		safe = safe || s
		repairs = append(repairs, items...)
	}
	if ok, s, items := browserHeaderJSONMapRepairAdvice(params["headers_json"]); ok {
		repairable = true
		safe = safe || s
		repairs = append(repairs, items...)
	}
	if ok, s, items := browserStringifiedHeadersRepairAdvice(params["headers"]); ok {
		repairable = true
		safe = safe || s
		repairs = append(repairs, items...)
	}
	if !repairable {
		return false, false, nil
	}
	return true, safe, repairs
}
