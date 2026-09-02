package tools

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
)

const bindingSourcesKeyword = "x-agentx-binding-sources"

// BindingSource identifies one trusted, transient source for a Tool argument.
// It never carries the argument value and is safe to project into content-free
// evidence.
type BindingSource string

const (
	BindingSourceUserInput       BindingSource = "user_input"
	BindingSourceTrustedHost     BindingSource = "trusted_host"
	BindingSourcePriorToolResult BindingSource = "prior_tool_result"
	BindingSourceDefaultAllowed  BindingSource = "default_allowed"
)

// BindingContext supplies transient values used only while validating one
// model-produced Tool call. Callers must not persist or log these values.
type BindingContext struct {
	UserInput       string
	TrustedHost     map[string]string
	PriorToolResult map[string]string
	DefaultAllowed  map[string]bool
}

// BindingEvidence is a content-free proof that one argument path was matched
// to an allowed source.
type BindingEvidence struct {
	Path   string
	Source BindingSource
}

// ArgumentValidation reports structural validity and any source-bound fields
// that still require user clarification.
type ArgumentValidation struct {
	NeedsClarification []string
	Evidence           []BindingEvidence
}

// Ready reports whether the call can proceed to authorization and execution.
func (v ArgumentValidation) Ready() bool { return len(v.NeedsClarification) == 0 }

var (
	// ErrInvalidToolDefinition means the server-owned model-facing definition
	// is not valid for the supported portable schema subset.
	ErrInvalidToolDefinition = errors.New("tools: invalid tool definition")
	// ErrInvalidToolArguments means model-produced arguments are not one exact
	// JSON object satisfying the server-owned definition.
	ErrInvalidToolArguments = errors.New("tools: invalid tool arguments")
)

// ValidateArguments validates one exact JSON object against the supported
// model-facing JSON Schema subset. It does not perform binding checks.
func ValidateArguments(definition toolcontract.Definition, raw string) error {
	_, err := ValidateCallArguments(definition, raw, BindingContext{})
	return err
}

// ValidateCallArguments validates structure and the optional
// x-agentx-binding-sources annotations declared by the Tool owner. Missing
// bindings are returned as a clarification result and are not structural
// errors. The function never executes a Tool or performs I/O.
func ValidateCallArguments(definition toolcontract.Definition, raw string, bindings BindingContext) (ArgumentValidation, error) {
	if !strings.EqualFold(strings.TrimSpace(definition.Type), "function") || strings.TrimSpace(definition.Function.Name) == "" || definition.Function.Parameters == nil {
		return ArgumentValidation{}, fmt.Errorf("%w: function contract is incomplete", ErrInvalidToolDefinition)
	}
	if err := validateArgumentSchema(definition.Function.Parameters, "$", true); err != nil {
		return ArgumentValidation{}, fmt.Errorf("%w: %v", ErrInvalidToolDefinition, err)
	}
	value, err := decodeExactJSONObject(raw)
	if err != nil {
		return ArgumentValidation{}, fmt.Errorf("%w: %v", ErrInvalidToolArguments, err)
	}
	result := ArgumentValidation{}
	if err := validateArgumentValue(value, definition.Function.Parameters, "$", bindings, &result); err != nil {
		return ArgumentValidation{}, fmt.Errorf("%w: %v", ErrInvalidToolArguments, err)
	}
	result.NeedsClarification = uniqueSortedArgumentPaths(result.NeedsClarification)
	sort.Slice(result.Evidence, func(i, j int) bool {
		if result.Evidence[i].Path == result.Evidence[j].Path {
			return result.Evidence[i].Source < result.Evidence[j].Source
		}
		return result.Evidence[i].Path < result.Evidence[j].Path
	})
	return result, nil
}

func decodeExactJSONObject(raw string) (map[string]any, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("arguments are empty")
	}
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("arguments must be valid JSON: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err == nil {
		return nil, errors.New("arguments contain trailing JSON")
	} else if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("arguments contain invalid trailing data: %w", err)
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, errors.New("arguments must be one JSON object")
	}
	return object, nil
}

func validateArgumentSchema(schema map[string]any, path string, root bool) error {
	if schema == nil {
		return fmt.Errorf("%s: schema is nil", path)
	}
	allowed := map[string]bool{
		"type": true, "description": true, "title": true, "properties": true, "required": true,
		"additionalProperties": true, "items": true, "enum": true, "const": true, "oneOf": true,
		"minItems": true, "maxItems": true, "minLength": true, "maxLength": true, "pattern": true,
		"minimum": true, "maximum": true, "format": true, "default": true, "examples": true,
		bindingSourcesKeyword: true,
	}
	for keyword := range schema {
		if !allowed[keyword] {
			return fmt.Errorf("%s.%s: unsupported schema keyword", path, keyword)
		}
	}
	types, err := schemaTypes(schema["type"])
	if err != nil {
		return fmt.Errorf("%s.type: %w", path, err)
	}
	if root && !containsString(types, "object") {
		return fmt.Errorf("%s.type: root schema must be object", path)
	}
	if raw, ok := schema["properties"]; ok {
		properties, ok := raw.(map[string]any)
		if !ok || !containsString(types, "object") {
			return fmt.Errorf("%s.properties: requires object type", path)
		}
		for name, childRaw := range properties {
			child, ok := childRaw.(map[string]any)
			if !ok || strings.TrimSpace(name) == "" {
				return fmt.Errorf("%s.properties: invalid property definition", path)
			}
			if err := validateArgumentSchema(child, path+"."+name, false); err != nil {
				return err
			}
		}
	}
	if raw, ok := schema["required"]; ok {
		if _, err := schemaStringList(raw, false); err != nil || !containsString(types, "object") {
			return fmt.Errorf("%s.required: invalid object requirement", path)
		}
	}
	if raw, ok := schema["additionalProperties"]; ok {
		switch typed := raw.(type) {
		case bool:
		case map[string]any:
			if err := validateArgumentSchema(typed, path+".additionalProperties", false); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%s.additionalProperties: must be boolean or schema", path)
		}
	}
	if raw, ok := schema["items"]; ok {
		child, ok := raw.(map[string]any)
		if !ok || !containsString(types, "array") {
			return fmt.Errorf("%s.items: requires array type", path)
		}
		if err := validateArgumentSchema(child, path+"[]", false); err != nil {
			return err
		}
	}
	if raw, ok := schema["oneOf"]; ok {
		alternatives, ok := raw.([]any)
		if !ok || len(alternatives) == 0 {
			return fmt.Errorf("%s.oneOf: must be a non-empty schema array", path)
		}
		for index, candidate := range alternatives {
			child, ok := candidate.(map[string]any)
			if !ok {
				return fmt.Errorf("%s.oneOf[%d]: must be a schema", path, index)
			}
			if err := validateArgumentSchema(child, fmt.Sprintf("%s.oneOf[%d]", path, index), false); err != nil {
				return err
			}
		}
	}
	if raw, ok := schema["enum"]; ok {
		values, valid := schemaLiteralList(raw)
		if !valid || len(values) == 0 {
			return fmt.Errorf("%s.enum: must be a non-empty JSON value array", path)
		}
	}
	if raw, ok := schema[bindingSourcesKeyword]; ok {
		sources, err := schemaStringList(raw, false)
		if err != nil || len(sources) == 0 || !containsString(types, "string") {
			return fmt.Errorf("%s.%s: requires a non-empty string source list", path, bindingSourcesKeyword)
		}
		for _, source := range sources {
			if !validBindingSource(BindingSource(source)) {
				return fmt.Errorf("%s.%s: unsupported source %q", path, bindingSourcesKeyword, source)
			}
		}
	}
	for _, keyword := range []string{"minItems", "maxItems", "minLength", "maxLength"} {
		if raw, ok := schema[keyword]; ok {
			if _, valid := nonNegativeInteger(raw); !valid {
				return fmt.Errorf("%s.%s: must be a non-negative integer", path, keyword)
			}
		}
	}
	for _, keyword := range []string{"minimum", "maximum"} {
		if raw, ok := schema[keyword]; ok {
			if _, valid := finiteNumber(raw); !valid {
				return fmt.Errorf("%s.%s: must be a finite number", path, keyword)
			}
		}
	}
	if raw, ok := schema["pattern"]; ok {
		pattern, ok := raw.(string)
		if !ok {
			return fmt.Errorf("%s.pattern: must be a string", path)
		}
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("%s.pattern: %w", path, err)
		}
	}
	if min, ok := nonNegativeInteger(schema["minItems"]); ok {
		if max, exists := nonNegativeInteger(schema["maxItems"]); exists && min > max {
			return fmt.Errorf("%s: minItems exceeds maxItems", path)
		}
	}
	if min, ok := nonNegativeInteger(schema["minLength"]); ok {
		if max, exists := nonNegativeInteger(schema["maxLength"]); exists && min > max {
			return fmt.Errorf("%s: minLength exceeds maxLength", path)
		}
	}
	return nil
}

func validateArgumentValue(value any, schema map[string]any, path string, bindings BindingContext, result *ArgumentValidation) error {
	if alternatives, ok := schema["oneOf"].([]any); ok {
		matched := 0
		var selected ArgumentValidation
		for _, raw := range alternatives {
			child := ArgumentValidation{}
			if validateArgumentValue(value, raw.(map[string]any), path, bindings, &child) == nil {
				matched++
				selected = child
			}
		}
		if matched != 1 {
			return fmt.Errorf("%s: must match exactly one oneOf schema", path)
		}
		result.NeedsClarification = append(result.NeedsClarification, selected.NeedsClarification...)
		result.Evidence = append(result.Evidence, selected.Evidence...)
		return nil
	}
	types, _ := schemaTypes(schema["type"])
	if len(types) > 0 && !matchesAnySchemaType(value, types) {
		return fmt.Errorf("%s: value has the wrong type", path)
	}
	if raw, ok := schema["enum"]; ok && !matchesLiteralList(value, raw) {
		return fmt.Errorf("%s: value is not in enum", path)
	}
	if expected, ok := schema["const"]; ok && !jsonValuesEqual(value, expected) {
		return fmt.Errorf("%s: value does not match const", path)
	}
	switch typed := value.(type) {
	case map[string]any:
		properties, _ := schema["properties"].(map[string]any)
		required, _ := schemaStringList(schema["required"], true)
		for _, name := range required {
			if _, exists := typed[name]; !exists {
				if child, ok := properties[name].(map[string]any); ok {
					if sources, err := schemaStringList(child[bindingSourcesKeyword], true); err == nil && len(sources) > 0 {
						result.NeedsClarification = append(result.NeedsClarification, path+"."+name)
						continue
					}
				}
				return fmt.Errorf("%s.%s: required property is missing", path, name)
			}
		}
		for name, childValue := range typed {
			childRaw, exists := properties[name]
			if !exists {
				switch additional := schema["additionalProperties"].(type) {
				case nil:
				case bool:
					if !additional {
						return fmt.Errorf("%s.%s: additional property is not allowed", path, name)
					}
				case map[string]any:
					if err := validateArgumentValue(childValue, additional, path+"."+name, bindings, result); err != nil {
						return err
					}
				}
				continue
			}
			if err := validateArgumentValue(childValue, childRaw.(map[string]any), path+"."+name, bindings, result); err != nil {
				return err
			}
		}
	case []any:
		if min, ok := nonNegativeInteger(schema["minItems"]); ok && len(typed) < min {
			return fmt.Errorf("%s: item count is below minItems", path)
		}
		if max, ok := nonNegativeInteger(schema["maxItems"]); ok && len(typed) > max {
			return fmt.Errorf("%s: item count exceeds maxItems", path)
		}
		if itemSchema, ok := schema["items"].(map[string]any); ok {
			for index, child := range typed {
				if err := validateArgumentValue(child, itemSchema, fmt.Sprintf("%s[%d]", path, index), bindings, result); err != nil {
					return err
				}
			}
		}
	case string:
		if min, ok := nonNegativeInteger(schema["minLength"]); ok && utf8.RuneCountInString(typed) < min {
			return fmt.Errorf("%s: length is below minLength", path)
		}
		if max, ok := nonNegativeInteger(schema["maxLength"]); ok && utf8.RuneCountInString(typed) > max {
			return fmt.Errorf("%s: length exceeds maxLength", path)
		}
		if pattern, ok := schema["pattern"].(string); ok && !regexp.MustCompile(pattern).MatchString(typed) {
			return fmt.Errorf("%s: value does not match pattern", path)
		}
		if sources, err := schemaStringList(schema[bindingSourcesKeyword], true); err == nil && len(sources) > 0 {
			matched := matchBinding(path, typed, sources, bindings)
			if matched == "" {
				result.NeedsClarification = append(result.NeedsClarification, path)
			} else {
				result.Evidence = append(result.Evidence, BindingEvidence{Path: path, Source: matched})
			}
		}
	case json.Number:
		number, err := typed.Float64()
		if err != nil || math.IsInf(number, 0) || math.IsNaN(number) {
			return fmt.Errorf("%s: number is not finite", path)
		}
		if minimum, ok := finiteNumber(schema["minimum"]); ok && number < minimum {
			return fmt.Errorf("%s: value is below minimum", path)
		}
		if maximum, ok := finiteNumber(schema["maximum"]); ok && number > maximum {
			return fmt.Errorf("%s: value exceeds maximum", path)
		}
	}
	return nil
}

func matchBinding(path, value string, sources []string, bindings BindingContext) BindingSource {
	for _, raw := range sources {
		source := BindingSource(raw)
		switch source {
		case BindingSourceUserInput:
			if normalizedBindingText(value) != "" && strings.Contains(normalizedBindingText(bindings.UserInput), normalizedBindingText(value)) {
				return source
			}
		case BindingSourceTrustedHost:
			if bindingValueMatches(bindings.TrustedHost[path], value) {
				return source
			}
		case BindingSourcePriorToolResult:
			if bindingValueMatches(bindings.PriorToolResult[path], value) {
				return source
			}
		case BindingSourceDefaultAllowed:
			if bindings.DefaultAllowed[path] {
				return source
			}
		}
	}
	return ""
}

func bindingValueMatches(expected, actual string) bool {
	return normalizedBindingText(expected) != "" && normalizedBindingText(expected) == normalizedBindingText(actual)
}

func normalizedBindingText(value string) string {
	return strings.Join(strings.Fields(strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		}
		if unicode.IsSpace(r) {
			return ' '
		}
		return -1
	}, value)), " ")
}

func schemaTypes(raw any) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	values, err := schemaStringList(raw, false)
	if text, ok := raw.(string); ok {
		values, err = []string{text}, nil
	}
	if err != nil || len(values) == 0 {
		return nil, errors.New("must be a string or non-empty string array")
	}
	seen := map[string]bool{}
	for index, value := range values {
		value = strings.TrimSpace(value)
		if !containsString([]string{"object", "array", "string", "number", "integer", "boolean", "null"}, value) || seen[value] {
			return nil, fmt.Errorf("unsupported or duplicate type %q", value)
		}
		values[index], seen[value] = value, true
	}
	return values, nil
}

func schemaStringList(raw any, allowNil bool) ([]string, error) {
	if raw == nil && allowNil {
		return nil, nil
	}
	switch typed := raw.(type) {
	case []string:
		return append([]string(nil), typed...), nil
	case []any:
		out := make([]string, len(typed))
		for index, item := range typed {
			value, ok := item.(string)
			if !ok || strings.TrimSpace(value) == "" {
				return nil, errors.New("must contain non-empty strings")
			}
			out[index] = strings.TrimSpace(value)
		}
		return out, nil
	default:
		return nil, errors.New("must be a string array")
	}
}

func matchesAnySchemaType(value any, types []string) bool {
	for _, schemaType := range types {
		switch schemaType {
		case "object":
			_, ok := value.(map[string]any)
			if ok {
				return true
			}
		case "array":
			_, ok := value.([]any)
			if ok {
				return true
			}
		case "string":
			_, ok := value.(string)
			if ok {
				return true
			}
		case "number":
			if number, ok := value.(json.Number); ok {
				if parsed, err := number.Float64(); err == nil && !math.IsInf(parsed, 0) && !math.IsNaN(parsed) {
					return true
				}
			}
		case "integer":
			if number, ok := value.(json.Number); ok {
				if parsed, err := number.Float64(); err == nil && math.Trunc(parsed) == parsed {
					return true
				}
			}
		case "boolean":
			_, ok := value.(bool)
			if ok {
				return true
			}
		case "null":
			if value == nil {
				return true
			}
		}
	}
	return false
}

func matchesLiteralList(value, raw any) bool {
	items, ok := schemaLiteralList(raw)
	if !ok || len(items) == 0 {
		return false
	}
	for _, item := range items {
		if jsonValuesEqual(value, item) {
			return true
		}
	}
	return false
}

func schemaLiteralList(raw any) ([]any, bool) {
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, false
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var values []any
	if err := decoder.Decode(&values); err != nil || values == nil {
		return nil, false
	}
	return values, true
}

func jsonValuesEqual(left, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftJSON, rightJSON)
}

func nonNegativeInteger(raw any) (int, bool) {
	number, ok := finiteNumber(raw)
	if !ok || number < 0 || math.Trunc(number) != number || number > float64(math.MaxInt) {
		return 0, false
	}
	return int(number), true
}

func finiteNumber(raw any) (float64, bool) {
	switch typed := raw.(type) {
	case int:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case float32:
		value := float64(typed)
		return value, !math.IsNaN(value) && !math.IsInf(value, 0)
	case float64:
		return typed, !math.IsNaN(typed) && !math.IsInf(typed, 0)
	case json.Number:
		value, err := typed.Float64()
		return value, err == nil && !math.IsNaN(value) && !math.IsInf(value, 0)
	default:
		return 0, false
	}
}

func validBindingSource(source BindingSource) bool {
	return source == BindingSourceUserInput || source == BindingSourceTrustedHost || source == BindingSourcePriorToolResult || source == BindingSourceDefaultAllowed
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func uniqueSortedArgumentPaths(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" && !seen[value] {
			seen[value], out = true, append(out, value)
		}
	}
	sort.Strings(out)
	return out
}
