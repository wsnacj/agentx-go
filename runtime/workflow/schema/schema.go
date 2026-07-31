// Package schema 提供 portable Workflow JSON Schema normalization 和
// definition validation mechanism。
//
// 该 package 只验证受支持的 schema 子集，不选择 Workflow config key、alias、
// 默认值或业务 policy，也不执行 runtime value validation。
package schema

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
)

// Normalize 把 JSON object 或无外围空白的 JSON object string 规范化为 map。
// 空 object 和空 string 返回 nil；label 原样用于稳定错误路径。
func Normalize(raw any, label string) (map[string]any, error) {
	switch value := raw.(type) {
	case map[string]any:
		if len(value) == 0 {
			return nil, nil
		}
		return value, nil
	case string:
		if err := validateOptionalField(value, label); err != nil {
			return nil, err
		}
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return nil, nil
		}
		var out map[string]any
		if err := json.Unmarshal([]byte(trimmed), &out); err != nil {
			return nil, fmt.Errorf("workflow: %s must be a valid JSON object: %w", label, err)
		}
		if len(out) == 0 {
			return nil, nil
		}
		return out, nil
	default:
		return nil, fmt.Errorf("workflow: %s must be a JSON object or JSON object string", label)
	}
}

// ValidateDefinition 递归验证受支持的 Workflow JSON Schema definition。
// path 原样作为错误路径前缀；函数不修改 schema。
func ValidateDefinition(schema map[string]any, path string) error {
	types, err := readWorkflowSchemaTypes(schema["type"])
	if err != nil {
		return fmt.Errorf("workflow: %s.type: %w", path, err)
	}
	for _, item := range types {
		if !isSupportedWorkflowSchemaType(item) {
			return fmt.Errorf("workflow: %s.type: unsupported type %q", path, item)
		}
	}
	if err := validateWorkflowSchemaKeywordTypeApplicability(schema, types, path); err != nil {
		return err
	}
	if rawConst, exists := schema["const"]; exists {
		if err := validateWorkflowSchemaConstCompatibility(rawConst, schema, types); err != nil {
			return fmt.Errorf("workflow: %s.const: %w", path, err)
		}
	}
	if rawEnum, exists := schema["enum"]; exists {
		if err := validateWorkflowSchemaEnum(rawEnum, schema, types); err != nil {
			return fmt.Errorf("workflow: %s.enum: %w", path, err)
		}
	}
	if required, exists := schema["required"]; exists {
		if _, err := readWorkflowSchemaRequired(required); err != nil {
			return fmt.Errorf("workflow: %s.required: %w", path, err)
		}
	}
	if rawProps, exists := schema["properties"]; exists {
		props, ok := rawProps.(map[string]any)
		if !ok {
			return fmt.Errorf("workflow: %s.properties: must be an object", path)
		}
		for key, rawChild := range props {
			child, ok := rawChild.(map[string]any)
			if !ok {
				return fmt.Errorf("workflow: %s.properties.%s: must be an object", path, key)
			}
			if err := ValidateDefinition(child, path+".properties."+key); err != nil {
				return err
			}
		}
	}
	if rawItems, exists := schema["items"]; exists {
		child, ok := rawItems.(map[string]any)
		if !ok {
			return fmt.Errorf("workflow: %s.items: must be an object", path)
		}
		if err := ValidateDefinition(child, path+".items"); err != nil {
			return err
		}
	}
	if rawAdditional, exists := schema["additionalProperties"]; exists {
		switch value := rawAdditional.(type) {
		case bool:
			_ = value
		case map[string]any:
			if err := ValidateDefinition(value, path+".additionalProperties"); err != nil {
				return err
			}
		default:
			return fmt.Errorf("workflow: %s.additionalProperties: must be boolean or object", path)
		}
	}
	if rawPattern, exists := schema["pattern"]; exists {
		pattern, ok := rawPattern.(string)
		if !ok {
			return fmt.Errorf("workflow: %s.pattern: must be a string", path)
		}
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("workflow: %s.pattern: %w", path, err)
		}
	}
	for _, keyword := range []string{"minProperties", "maxProperties", "minItems", "maxItems", "minLength", "maxLength"} {
		if rawValue, exists := schema[keyword]; exists {
			if err := validateWorkflowSchemaNonNegativeIntegerKeyword(rawValue); err != nil {
				return fmt.Errorf("workflow: %s.%s: %w", path, keyword, err)
			}
		}
	}
	for _, keyword := range []string{"minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum"} {
		if rawValue, exists := schema[keyword]; exists {
			if err := validateWorkflowSchemaFiniteNumberKeyword(rawValue); err != nil {
				return fmt.Errorf("workflow: %s.%s: %w", path, keyword, err)
			}
		}
	}
	if err := validateWorkflowSchemaKeywordRanges(schema, path); err != nil {
		return err
	}
	return nil
}

func validateWorkflowSchemaKeywordTypeApplicability(schema map[string]any, types []string, path string) error {
	if len(types) == 0 {
		return nil
	}
	if err := validateWorkflowSchemaKeywordsRequireType(schema, types, path, []string{
		"properties", "required", "additionalProperties", "minProperties", "maxProperties",
	}, "object"); err != nil {
		return err
	}
	if err := validateWorkflowSchemaKeywordsRequireType(schema, types, path, []string{
		"items", "minItems", "maxItems",
	}, "array"); err != nil {
		return err
	}
	if err := validateWorkflowSchemaKeywordsRequireType(schema, types, path, []string{
		"pattern", "minLength", "maxLength",
	}, "string"); err != nil {
		return err
	}
	if err := validateWorkflowSchemaKeywordsRequireAnyType(schema, types, path, []string{
		"minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum",
	}, []string{"number", "integer"}); err != nil {
		return err
	}
	return nil
}

func validateWorkflowSchemaKeywordsRequireType(schema map[string]any, types []string, path string, keywords []string, requiredType string) error {
	return validateWorkflowSchemaKeywordsRequireAnyType(schema, types, path, keywords, []string{requiredType})
}

func validateWorkflowSchemaKeywordsRequireAnyType(schema map[string]any, types []string, path string, keywords []string, requiredTypes []string) error {
	for _, keyword := range keywords {
		if _, exists := schema[keyword]; !exists {
			continue
		}
		if workflowSchemaTypesIncludeAny(types, requiredTypes...) {
			continue
		}
		return fmt.Errorf("workflow: %s.%s requires declared type to include %s", path, keyword, strings.Join(requiredTypes, " or "))
	}
	return nil
}

func workflowSchemaTypesIncludeAny(types []string, candidates ...string) bool {
	for _, item := range types {
		for _, candidate := range candidates {
			if item == candidate {
				return true
			}
		}
	}
	return false
}

func readWorkflowSchemaTypes(raw any) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	switch typed := raw.(type) {
	case string:
		value, err := normalizeWorkflowSchemaTypeEntry(typed)
		if err != nil {
			return nil, err
		}
		return []string{value}, nil
	case []string:
		seen := map[string]bool{}
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			value, err := normalizeWorkflowSchemaTypeEntry(item)
			if err != nil {
				return nil, err
			}
			if seen[value] {
				return nil, fmt.Errorf("must not contain duplicate entries")
			}
			seen[value] = true
			out = append(out, value)
		}
		return out, nil
	case []any:
		seen := map[string]bool{}
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("must contain strings")
			}
			value, err := normalizeWorkflowSchemaTypeEntry(text)
			if err != nil {
				return nil, err
			}
			if seen[value] {
				return nil, fmt.Errorf("must not contain duplicate entries")
			}
			seen[value] = true
			out = append(out, value)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("must be string or array")
	}
}

func readWorkflowSchemaRequired(raw any) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	switch typed := raw.(type) {
	case []string:
		seen := map[string]bool{}
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			name, err := normalizeWorkflowSchemaRequiredEntry(item)
			if err != nil {
				return nil, err
			}
			if seen[name] {
				return nil, fmt.Errorf("must not contain duplicate entries")
			}
			seen[name] = true
			out = append(out, name)
		}
		return out, nil
	case []any:
		seen := map[string]bool{}
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("must contain strings")
			}
			name, err := normalizeWorkflowSchemaRequiredEntry(text)
			if err != nil {
				return nil, err
			}
			if seen[name] {
				return nil, fmt.Errorf("must not contain duplicate entries")
			}
			seen[name] = true
			out = append(out, name)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("must be an array of strings")
	}
}

func normalizeWorkflowSchemaTypeEntry(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("must not contain empty entries")
	}
	if raw != trimmed {
		return "", fmt.Errorf("must not include surrounding whitespace")
	}
	lower := strings.ToLower(trimmed)
	if trimmed != lower {
		return "", fmt.Errorf("must use canonical lowercase")
	}
	return lower, nil
}

func normalizeWorkflowSchemaRequiredEntry(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("must not contain empty entries")
	}
	if raw != trimmed {
		return "", fmt.Errorf("must not include surrounding whitespace")
	}
	return trimmed, nil
}

func validateWorkflowSchemaEnum(raw any, schema map[string]any, types []string) error {
	items, ok := workflowSchemaLiteralSlice(raw)
	if !ok || len(items) == 0 {
		return fmt.Errorf("must be a non-empty array")
	}
	for idx, item := range items {
		for prev := 0; prev < idx; prev++ {
			if workflowSchemaValuesEqual(item, items[prev]) {
				return fmt.Errorf("[%d]: must not contain duplicate entries", idx)
			}
		}
		if err := validateWorkflowSchemaLiteralMatchesTypes(item, types); err != nil {
			return fmt.Errorf("[%d]: %w", idx, err)
		}
		if err := validateWorkflowSchemaLiteralAgainstDefinition(item, schema); err != nil {
			return fmt.Errorf("[%d]: %w", idx, err)
		}
	}
	return nil
}

func validateWorkflowSchemaConstCompatibility(raw any, schema map[string]any, types []string) error {
	if err := validateWorkflowSchemaLiteralMatchesTypes(raw, types); err != nil {
		return err
	}
	return validateWorkflowSchemaLiteralAgainstDefinition(raw, schema)
}

func validateWorkflowSchemaLiteralMatchesTypes(raw any, types []string) error {
	if len(types) == 0 {
		return nil
	}
	for _, schemaType := range types {
		if workflowSchemaValueMatchesType(raw, schemaType) {
			return nil
		}
	}
	return fmt.Errorf("does not match declared type constraint")
}

func validateWorkflowSchemaLiteralAgainstDefinition(raw any, schema map[string]any) error {
	if err := validateWorkflowSchemaLiteralMatchesTypes(raw, readWorkflowSchemaTypesOrNil(schema["type"])); err != nil {
		return err
	}
	switch value := raw.(type) {
	case string:
		return validateWorkflowSchemaLiteralStringConstraints(value, schema)
	case int:
		return validateWorkflowSchemaLiteralNumberConstraints(float64(value), schema)
	case int8:
		return validateWorkflowSchemaLiteralNumberConstraints(float64(value), schema)
	case int16:
		return validateWorkflowSchemaLiteralNumberConstraints(float64(value), schema)
	case int32:
		return validateWorkflowSchemaLiteralNumberConstraints(float64(value), schema)
	case int64:
		return validateWorkflowSchemaLiteralNumberConstraints(float64(value), schema)
	case uint:
		return validateWorkflowSchemaLiteralNumberConstraints(float64(value), schema)
	case uint8:
		return validateWorkflowSchemaLiteralNumberConstraints(float64(value), schema)
	case uint16:
		return validateWorkflowSchemaLiteralNumberConstraints(float64(value), schema)
	case uint32:
		return validateWorkflowSchemaLiteralNumberConstraints(float64(value), schema)
	case uint64:
		return validateWorkflowSchemaLiteralNumberConstraints(float64(value), schema)
	case float32:
		return validateWorkflowSchemaLiteralNumberConstraints(float64(value), schema)
	case float64:
		return validateWorkflowSchemaLiteralNumberConstraints(value, schema)
	}
	if object, ok := workflowSchemaLiteralMap(raw); ok {
		return validateWorkflowSchemaLiteralObjectConstraints(object, schema)
	}
	if items, ok := workflowSchemaLiteralSlice(raw); ok {
		return validateWorkflowSchemaLiteralArrayConstraints(items, schema)
	}
	return nil
}

func readWorkflowSchemaTypesOrNil(raw any) []string {
	types, err := readWorkflowSchemaTypes(raw)
	if err != nil {
		return nil
	}
	return types
}

func validateWorkflowSchemaLiteralObjectConstraints(object map[string]any, schema map[string]any) error {
	if minProps, ok, err := readWorkflowSchemaNonNegativeIntegerKeyword(schema["minProperties"]); err == nil && ok && float64(len(object)) < minProps {
		return fmt.Errorf("violates minProperties")
	} else if err != nil {
		return err
	}
	if maxProps, ok, err := readWorkflowSchemaNonNegativeIntegerKeyword(schema["maxProperties"]); err == nil && ok && float64(len(object)) > maxProps {
		return fmt.Errorf("violates maxProperties")
	} else if err != nil {
		return err
	}
	required, err := readWorkflowSchemaRequired(schema["required"])
	if err != nil {
		return err
	}
	for _, key := range required {
		if _, exists := object[key]; !exists {
			return fmt.Errorf("violates required")
		}
	}
	properties := map[string]any{}
	if rawProps, exists := schema["properties"]; exists {
		typed, ok := rawProps.(map[string]any)
		if !ok {
			return fmt.Errorf("properties: must be an object")
		}
		properties = typed
	}
	additional := schema["additionalProperties"]
	for key, item := range object {
		if childSchemaRaw, exists := properties[key]; exists {
			childSchema, ok := childSchemaRaw.(map[string]any)
			if !ok {
				return fmt.Errorf("property schema must be object")
			}
			if err := validateWorkflowSchemaLiteralAgainstDefinition(item, childSchema); err != nil {
				return err
			}
			continue
		}
		switch typed := additional.(type) {
		case nil:
			continue
		case bool:
			if !typed {
				return fmt.Errorf("violates additionalProperties")
			}
		case map[string]any:
			if err := validateWorkflowSchemaLiteralAgainstDefinition(item, typed); err != nil {
				return err
			}
		default:
			return fmt.Errorf("additionalProperties: must be boolean or object")
		}
	}
	return nil
}

func validateWorkflowSchemaLiteralArrayConstraints(items []any, schema map[string]any) error {
	if minItems, ok, err := readWorkflowSchemaNonNegativeIntegerKeyword(schema["minItems"]); err == nil && ok && float64(len(items)) < minItems {
		return fmt.Errorf("violates minItems")
	} else if err != nil {
		return err
	}
	if maxItems, ok, err := readWorkflowSchemaNonNegativeIntegerKeyword(schema["maxItems"]); err == nil && ok && float64(len(items)) > maxItems {
		return fmt.Errorf("violates maxItems")
	} else if err != nil {
		return err
	}
	rawItems, exists := schema["items"]
	if !exists {
		return nil
	}
	itemSchema, ok := rawItems.(map[string]any)
	if !ok {
		return fmt.Errorf("items: must be an object")
	}
	for _, item := range items {
		if err := validateWorkflowSchemaLiteralAgainstDefinition(item, itemSchema); err != nil {
			return err
		}
	}
	return nil
}

func validateWorkflowSchemaLiteralStringConstraints(value string, schema map[string]any) error {
	if minLength, ok, err := readWorkflowSchemaNonNegativeIntegerKeyword(schema["minLength"]); err == nil && ok && float64(len(value)) < minLength {
		return fmt.Errorf("violates minLength")
	} else if err != nil {
		return err
	}
	if maxLength, ok, err := readWorkflowSchemaNonNegativeIntegerKeyword(schema["maxLength"]); err == nil && ok && float64(len(value)) > maxLength {
		return fmt.Errorf("violates maxLength")
	} else if err != nil {
		return err
	}
	if rawPattern, exists := schema["pattern"]; exists {
		pattern, ok := rawPattern.(string)
		if !ok {
			return fmt.Errorf("pattern: must be a string")
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		if !re.MatchString(value) {
			return fmt.Errorf("violates pattern")
		}
	}
	return nil
}

func validateWorkflowSchemaLiteralNumberConstraints(value float64, schema map[string]any) error {
	if minimum, ok, err := readWorkflowSchemaFiniteNumberKeyword(schema["minimum"]); err == nil && ok && value < minimum {
		return fmt.Errorf("violates minimum")
	} else if err != nil {
		return err
	}
	if maximum, ok, err := readWorkflowSchemaFiniteNumberKeyword(schema["maximum"]); err == nil && ok && value > maximum {
		return fmt.Errorf("violates maximum")
	} else if err != nil {
		return err
	}
	if minimum, ok, err := readWorkflowSchemaFiniteNumberKeyword(schema["exclusiveMinimum"]); err == nil && ok && value <= minimum {
		return fmt.Errorf("violates exclusiveMinimum")
	} else if err != nil {
		return err
	}
	if maximum, ok, err := readWorkflowSchemaFiniteNumberKeyword(schema["exclusiveMaximum"]); err == nil && ok && value >= maximum {
		return fmt.Errorf("violates exclusiveMaximum")
	} else if err != nil {
		return err
	}
	return nil
}

func workflowSchemaLiteralMap(raw any) (map[string]any, bool) {
	if raw == nil {
		return nil, false
	}
	rv := reflect.ValueOf(raw)
	if rv.Kind() != reflect.Map || rv.Type().Key().Kind() != reflect.String {
		return nil, false
	}
	out := make(map[string]any, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		out[iter.Key().String()] = iter.Value().Interface()
	}
	return out, true
}

func workflowSchemaLiteralSlice(raw any) ([]any, bool) {
	if raw == nil {
		return nil, false
	}
	rv := reflect.ValueOf(raw)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}
	out := make([]any, rv.Len())
	for idx := 0; idx < rv.Len(); idx++ {
		out[idx] = rv.Index(idx).Interface()
	}
	return out, true
}

func workflowSchemaValueMatchesType(value any, schemaType string) bool {
	switch schemaType {
	case "object":
		if value == nil {
			return false
		}
		rv := reflect.ValueOf(value)
		return rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String
	case "array":
		if value == nil {
			return false
		}
		rv := reflect.ValueOf(value)
		return rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		return workflowSchemaValueAsFiniteNumber(value, false)
	case "integer":
		return workflowSchemaValueAsFiniteNumber(value, true)
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "null":
		return value == nil
	default:
		return false
	}
}

func workflowSchemaValueAsFiniteNumber(value any, integerOnly bool) bool {
	switch typed := value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32:
		num := float64(typed)
		if math.IsNaN(num) || math.IsInf(num, 0) {
			return false
		}
		return !integerOnly || math.Trunc(num) == num
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return false
		}
		return !integerOnly || math.Trunc(typed) == typed
	default:
		return false
	}
}

func workflowSchemaValuesEqual(left any, right any) bool {
	return reflect.DeepEqual(normalizeWorkflowSchemaComparable(left), normalizeWorkflowSchemaComparable(right))
}

func normalizeWorkflowSchemaComparable(value any) any {
	if number, ok, err := readWorkflowSchemaFiniteNumberKeyword(value); err == nil && ok {
		return number
	}
	if value == nil {
		return nil
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return value
		}
		out := make(map[string]any, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			out[iter.Key().String()] = normalizeWorkflowSchemaComparable(iter.Value().Interface())
		}
		return out
	case reflect.Slice, reflect.Array:
		out := make([]any, rv.Len())
		for idx := 0; idx < rv.Len(); idx++ {
			out[idx] = normalizeWorkflowSchemaComparable(rv.Index(idx).Interface())
		}
		return out
	default:
		return value
	}
}

func validateWorkflowSchemaNonNegativeIntegerKeyword(raw any) error {
	switch value := raw.(type) {
	case int:
		if value < 0 {
			return fmt.Errorf("must be >= 0")
		}
		return nil
	case int8:
		if value < 0 {
			return fmt.Errorf("must be >= 0")
		}
		return nil
	case int16:
		if value < 0 {
			return fmt.Errorf("must be >= 0")
		}
		return nil
	case int32:
		if value < 0 {
			return fmt.Errorf("must be >= 0")
		}
		return nil
	case int64:
		if value < 0 {
			return fmt.Errorf("must be >= 0")
		}
		return nil
	case uint, uint8, uint16, uint32, uint64:
		return nil
	case float32:
		num := float64(value)
		if math.IsNaN(num) || math.IsInf(num, 0) {
			return fmt.Errorf("must be a finite integer")
		}
		if math.Trunc(num) != num {
			return fmt.Errorf("must be an integer")
		}
		if num < 0 {
			return fmt.Errorf("must be >= 0")
		}
		return nil
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("must be a finite integer")
		}
		if math.Trunc(value) != value {
			return fmt.Errorf("must be an integer")
		}
		if value < 0 {
			return fmt.Errorf("must be >= 0")
		}
		return nil
	default:
		return fmt.Errorf("must be a non-negative integer")
	}
}

func validateWorkflowSchemaFiniteNumberKeyword(raw any) error {
	switch value := raw.(type) {
	case int, int8, int16, int32, int64:
		return nil
	case uint, uint8, uint16, uint32, uint64:
		return nil
	case float32:
		num := float64(value)
		if math.IsNaN(num) || math.IsInf(num, 0) {
			return fmt.Errorf("must be a finite number")
		}
		return nil
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("must be a finite number")
		}
		return nil
	default:
		return fmt.Errorf("must be a number")
	}
}

func validateWorkflowSchemaKeywordRanges(schema map[string]any, path string) error {
	if min, ok, err := readWorkflowSchemaNonNegativeIntegerKeyword(schema["minProperties"]); err != nil {
		return fmt.Errorf("workflow: %s.minProperties: %w", path, err)
	} else if max, okMax, err := readWorkflowSchemaNonNegativeIntegerKeyword(schema["maxProperties"]); err != nil {
		return fmt.Errorf("workflow: %s.maxProperties: %w", path, err)
	} else if ok && okMax && min > max {
		return fmt.Errorf("workflow: %s.minProperties: must be <= maxProperties", path)
	}
	if min, ok, err := readWorkflowSchemaNonNegativeIntegerKeyword(schema["minItems"]); err != nil {
		return fmt.Errorf("workflow: %s.minItems: %w", path, err)
	} else if max, okMax, err := readWorkflowSchemaNonNegativeIntegerKeyword(schema["maxItems"]); err != nil {
		return fmt.Errorf("workflow: %s.maxItems: %w", path, err)
	} else if ok && okMax && min > max {
		return fmt.Errorf("workflow: %s.minItems: must be <= maxItems", path)
	}
	if min, ok, err := readWorkflowSchemaNonNegativeIntegerKeyword(schema["minLength"]); err != nil {
		return fmt.Errorf("workflow: %s.minLength: %w", path, err)
	} else if max, okMax, err := readWorkflowSchemaNonNegativeIntegerKeyword(schema["maxLength"]); err != nil {
		return fmt.Errorf("workflow: %s.maxLength: %w", path, err)
	} else if ok && okMax && min > max {
		return fmt.Errorf("workflow: %s.minLength: must be <= maxLength", path)
	}
	if min, ok, err := readWorkflowSchemaFiniteNumberKeyword(schema["minimum"]); err != nil {
		return fmt.Errorf("workflow: %s.minimum: %w", path, err)
	} else if max, okMax, err := readWorkflowSchemaFiniteNumberKeyword(schema["maximum"]); err != nil {
		return fmt.Errorf("workflow: %s.maximum: %w", path, err)
	} else if ok && okMax && min > max {
		return fmt.Errorf("workflow: %s.minimum: must be <= maximum", path)
	}
	if min, ok, err := readWorkflowSchemaFiniteNumberKeyword(schema["exclusiveMinimum"]); err != nil {
		return fmt.Errorf("workflow: %s.exclusiveMinimum: %w", path, err)
	} else if max, okMax, err := readWorkflowSchemaFiniteNumberKeyword(schema["exclusiveMaximum"]); err != nil {
		return fmt.Errorf("workflow: %s.exclusiveMaximum: %w", path, err)
	} else if ok && okMax && min >= max {
		return fmt.Errorf("workflow: %s.exclusiveMinimum: must be < exclusiveMaximum", path)
	}
	return nil
}

func readWorkflowSchemaNonNegativeIntegerKeyword(raw any) (float64, bool, error) {
	if raw == nil {
		return 0, false, nil
	}
	if err := validateWorkflowSchemaNonNegativeIntegerKeyword(raw); err != nil {
		return 0, false, err
	}
	switch value := raw.(type) {
	case int:
		return float64(value), true, nil
	case int8:
		return float64(value), true, nil
	case int16:
		return float64(value), true, nil
	case int32:
		return float64(value), true, nil
	case int64:
		return float64(value), true, nil
	case uint:
		return float64(value), true, nil
	case uint8:
		return float64(value), true, nil
	case uint16:
		return float64(value), true, nil
	case uint32:
		return float64(value), true, nil
	case uint64:
		return float64(value), true, nil
	case float32:
		return float64(value), true, nil
	case float64:
		return value, true, nil
	default:
		return 0, false, nil
	}
}

func readWorkflowSchemaFiniteNumberKeyword(raw any) (float64, bool, error) {
	if raw == nil {
		return 0, false, nil
	}
	if err := validateWorkflowSchemaFiniteNumberKeyword(raw); err != nil {
		return 0, false, err
	}
	switch value := raw.(type) {
	case int:
		return float64(value), true, nil
	case int8:
		return float64(value), true, nil
	case int16:
		return float64(value), true, nil
	case int32:
		return float64(value), true, nil
	case int64:
		return float64(value), true, nil
	case uint:
		return float64(value), true, nil
	case uint8:
		return float64(value), true, nil
	case uint16:
		return float64(value), true, nil
	case uint32:
		return float64(value), true, nil
	case uint64:
		return float64(value), true, nil
	case float32:
		return float64(value), true, nil
	case float64:
		return value, true, nil
	default:
		return 0, false, nil
	}
}

func isSupportedWorkflowSchemaType(schemaType string) bool {
	switch schemaType {
	case "object", "array", "string", "number", "integer", "boolean", "null":
		return true
	default:
		return false
	}
}

func validateOptionalField(value string, label string) error {
	if value == "" {
		return nil
	}
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("workflow: %s %q must not be whitespace-only", label, value)
	}
	if value != strings.TrimSpace(value) {
		return fmt.Errorf("workflow: %s %q must not include surrounding whitespace", label, value)
	}
	return nil
}
