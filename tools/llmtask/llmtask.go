// Package llmtask provides one bounded, model-only JSON subtask tool.
package llmtask

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	llm "github.com/wsnacj/agentx-go/components/llm"
	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
)

const (
	defaultLLMTaskTimeoutMs  = 45000
	defaultLLMTaskMaxContent = 120000
	llmTaskSchemaName        = "agentx_llm_task_result"
)

// Name is the catalog name of the LLM task tool.
const Name = "llm_task"

// ChatWithInputFunc is the narrow Host-owned model invocation port.
type ChatWithInputFunc func(ctx context.Context, input llm.ChatInput) (*llm.ChatResponse, error)

// Options configures the portable LLM task coordinator.
//
// ChatWithInput is required. The package never discovers a provider, model,
// credential, endpoint, or global model registry.
type Options struct {
	ModelConfig        string
	AllowModelOverride bool
	DefaultTimeoutMs   int
	MaxContentChars    int
	ChatWithInput      ChatWithInputFunc
}

type llmTaskRequest struct {
	ModelConfig    string
	Instruction    string
	Input          string
	Schema         map[string]any
	Strict         bool
	Temperature    float64
	HasTemperature bool
	MaxTokens      int
	TimeoutMs      int
}

// Register adds llm_task when both a registrar and explicit model port exist.
func Register(reg toolcontract.Registrar, opts Options) {
	if reg == nil || opts.ChatWithInput == nil {
		return
	}
	reg.Register(Definition(), NewHandler(opts))
}

// NewHandler returns the portable model-facing handler.
func NewHandler(opts Options) toolcontract.Handler {
	defaultTimeout := opts.DefaultTimeoutMs
	if defaultTimeout <= 0 {
		defaultTimeout = defaultLLMTaskTimeoutMs
	}
	maxContent := opts.MaxContentChars
	if maxContent <= 0 {
		maxContent = defaultLLMTaskMaxContent
	}
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		params, err := decodeArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		req, err := buildLLMTaskRequest(params, opts, defaultTimeout)
		if err != nil {
			return "", err
		}
		runCtx := ctx
		if req.TimeoutMs > 0 {
			var cancel context.CancelFunc
			runCtx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutMs)*time.Millisecond)
			defer cancel()
		}
		resp, err := invokeLLMTaskChat(runCtx, opts, req)
		if err != nil {
			return "", fmt.Errorf("llm_task: chat failed: %w", err)
		}
		content := strings.TrimSpace(resp.Content)
		if len(content) == 0 {
			return "", fmt.Errorf("llm_task: empty model content")
		}
		value, rawJSON, err := decodeLLMTaskContent(content, maxContent)
		if err != nil {
			return "", fmt.Errorf("llm_task: %w", err)
		}
		if req.Schema != nil {
			if err := validateJSONSchemaSubset(value, req.Schema, "$"); err != nil {
				return "", fmt.Errorf("llm_task: schema validation failed: %w", err)
			}
		}
		payload := map[string]any{
			"tool":           "llm_task",
			"model":          req.ModelConfig,
			"schema_applied": req.Schema != nil,
			"result":         value,
			"raw_json":       rawJSON,
		}
		blob, err := json.Marshal(payload)
		if err != nil {
			return "", err
		}
		return string(blob), nil
	}
}

func invokeLLMTaskChat(ctx context.Context, opts Options, req llmTaskRequest) (*llm.ChatResponse, error) {
	if opts.ChatWithInput == nil {
		return nil, fmt.Errorf("%s: model adapter is unavailable", Name)
	}
	return opts.ChatWithInput(ctx, buildLLMTaskChatInput(req))
}

func buildLLMTaskChatInput(req llmTaskRequest) llm.ChatInput {
	request := llm.RequestOptions{}
	request.ResponseFormat = llmTaskResponseFormat(req)
	if req.HasTemperature {
		request.Temperature = &req.Temperature
	}
	if req.MaxTokens > 0 {
		request.MaxTokens = &req.MaxTokens
	}
	return llm.ChatInput{
		ConfigName:   req.ModelConfig,
		SystemPrompt: llmTaskSystemPrompt(req.Schema != nil),
		Messages: llm.Conversation{
			{Role: "user", Content: buildLLMTaskInput(req)},
		},
		Request: request,
		ToolChoice: &llm.ToolChoice{
			Type: "none",
		},
	}
}

func llmTaskResponseFormat(req llmTaskRequest) any {
	if req.Schema != nil {
		return map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   llmTaskSchemaName,
				"strict": req.Strict,
				"schema": req.Schema,
			},
		}
	}
	return map[string]any{"type": "json_object"}
}

// Definition returns the stable llm_task schema.
func Definition() toolcontract.Definition {
	return toolcontract.Definition{
		Type: "function",
		Function: toolcontract.Function{
			Name:        "llm_task",
			Description: "Run one focused LLM-only subtask with a JSON result and optional schema validation. Use subagents action=fanout for multiple independent subtasks or task-backed lifecycle management.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"instruction":   stringSchema("Focused instruction for the LLM-only subtask. Runtime still accepts legacy aliases such as task/prompt/goal/request/query for compatibility."),
					"input":         stringSchema("Optional source text or context for the instruction. Runtime still accepts context as a compatibility alias."),
					"schema":        jsonSchemaInputSchema("Optional JSON Schema subset for validating the result."),
					"output_schema": jsonSchemaInputSchema("Compatibility alias for schema. Prefer schema for new calls."),
					"strict":        boolSchema("Whether to request strict JSON schema adherence when schema is provided. Defaults to true."),
					"model":         stringSchema("Optional model config override when the runtime allows model overrides."),
					"temperature":   numberSchema("Optional model temperature."),
					"max_tokens":    intSchema("Optional maximum model output tokens.", 1),
					"timeout_ms":    intSchema("Maximum subtask runtime in milliseconds. Omit to use the runtime default.", 1),
				},
				"required": []string{"instruction"},
			},
			OutputSchema: llmTaskOutputSchema(),
		},
	}
}

func buildLLMTaskRequest(params map[string]any, opts Options, defaultTimeout int) (llmTaskRequest, error) {
	instruction := firstString(params, "instruction", "task", "prompt", "goal", "request", "query")
	if instruction == "" {
		return llmTaskRequest{}, agentxtoolerrors.NewMissingRequiredToolArgumentError(Name, []string{"instruction"}, Name+": instruction is required")
	}
	modelConfig := strings.TrimSpace(opts.ModelConfig)
	if override := readString(params, "model"); override != "" {
		if !opts.AllowModelOverride && override != modelConfig {
			return llmTaskRequest{}, fmt.Errorf("llm_task: model override is not allowed")
		}
		modelConfig = override
	}
	if modelConfig == "" {
		return llmTaskRequest{}, agentxtoolerrors.NewMissingRequiredToolArgumentError(Name, []string{"model"}, Name+": model config is required")
	}
	schema, err := normalizeJSONSchema(firstNonNil(params["schema"], params["output_schema"]))
	if err != nil {
		return llmTaskRequest{}, fmt.Errorf("llm_task: %w", err)
	}
	if schema != nil {
		if err := validateSchemaDefinition(schema, "$"); err != nil {
			return llmTaskRequest{}, fmt.Errorf("llm_task: invalid schema: %w", err)
		}
	}
	strict := true
	if _, ok := params["strict"]; ok {
		strict = readBool(params, "strict")
	}
	timeoutMs := firstInt(params, "timeout_ms")
	if timeoutMs <= 0 {
		timeoutMs = defaultTimeout
	}
	temperature, hasTemperature := readFloat(params, "temperature")
	return llmTaskRequest{
		ModelConfig:    modelConfig,
		Instruction:    instruction,
		Input:          firstString(params, "input", "context"),
		Schema:         schema,
		Strict:         strict,
		Temperature:    temperature,
		HasTemperature: hasTemperature,
		MaxTokens:      firstInt(params, "max_tokens"),
		TimeoutMs:      timeoutMs,
	}, nil
}

func llmTaskSystemPrompt(hasSchema bool) string {
	if hasSchema {
		return "You are Agentx llm_task worker. Return only valid JSON that strictly follows the provided schema. The exact schema is included in the prompt. Do not include markdown."
	}
	return "You are Agentx llm_task worker. Return only valid JSON object output. Do not include markdown or extra commentary."
}

func buildLLMTaskInput(req llmTaskRequest) string {
	builder := strings.Builder{}
	builder.WriteString("Instruction:\n")
	builder.WriteString(strings.TrimSpace(req.Instruction))
	if strings.TrimSpace(req.Input) != "" {
		builder.WriteString("\n\nInput:\n")
		builder.WriteString(strings.TrimSpace(req.Input))
	}
	builder.WriteString("\n\nOutput requirements:\n- JSON only\n- No markdown fences")
	if req.Schema != nil {
		builder.WriteString("\n- Must satisfy schema constraints exactly")
		if schemaJSON, err := json.Marshal(req.Schema); err == nil {
			builder.WriteString("\n\nSchema:\n")
			builder.WriteString(string(schemaJSON))
		}
	}
	return builder.String()
}

func decodeLLMTaskContent(content string, maxChars int) (any, string, error) {
	scriptStripped := strings.TrimSpace(stripJSONPayload(content))
	candidates := []string{
		strings.TrimSpace(content),
		scriptStripped,
		strings.TrimSpace(stripCodeFence(content)),
	}
	if extracted := strings.TrimSpace(extractFirstJSONObjectOrArray(content)); extracted != "" {
		candidates = append(candidates, extracted)
	}
	if extracted := strings.TrimSpace(extractFirstJSONObjectOrArray(scriptStripped)); extracted != "" {
		candidates = append(candidates, extracted)
	}
	seen := map[string]bool{}
	for _, candidate := range candidates {
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		trimmed := truncateRunes(candidate, maxChars)
		var parsed any
		if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
			continue
		}
		return parsed, trimmed, nil
	}
	return nil, "", fmt.Errorf("response is not valid json")
}

func stripCodeFence(text string) string {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "```") {
		return trimmed
	}
	lines := strings.Split(trimmed, "\n")
	if len(lines) < 2 {
		return trimmed
	}
	last := len(lines) - 1
	if !strings.HasPrefix(strings.TrimSpace(lines[last]), "```") {
		return trimmed
	}
	body := strings.Join(lines[1:last], "\n")
	return strings.TrimSpace(body)
}

func extractFirstJSONObjectOrArray(text string) string {
	start := -1
	var open byte
	var close byte
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '{':
			start = i
			open = '{'
			close = '}'
		case '[':
			start = i
			open = '['
			close = ']'
		default:
			continue
		}
		break
	}
	if start < 0 {
		return ""
	}
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(text); i++ {
		ch := text[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		if ch == '"' {
			inString = true
			continue
		}
		if ch == open {
			depth++
			continue
		}
		if ch == close {
			depth--
			if depth == 0 {
				return strings.TrimSpace(text[start : i+1])
			}
		}
	}
	return ""
}

func truncateRunes(text string, maxChars int) string {
	if maxChars <= 0 {
		return text
	}
	if utf8.ValidString(text) {
		runes := []rune(text)
		if len(runes) <= maxChars {
			return text
		}
		return string(runes[:maxChars])
	}
	if len(text) <= maxChars {
		return text
	}
	return text[:maxChars]
}

func firstNonNil(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func normalizeJSONSchema(raw any) (map[string]any, error) {
	if raw == nil {
		return nil, nil
	}
	switch value := raw.(type) {
	case map[string]any:
		if len(value) == 0 {
			return nil, nil
		}
		return value, nil
	case string:
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return nil, nil
		}
		var out map[string]any
		if err := json.Unmarshal([]byte(trimmed), &out); err != nil {
			return nil, fmt.Errorf("schema is not valid json object: %w", err)
		}
		if len(out) == 0 {
			return nil, nil
		}
		return out, nil
	default:
		return nil, fmt.Errorf("schema must be an object")
	}
}

func validateSchemaDefinition(schema map[string]any, path string) error {
	types, err := readSchemaTypes(schema["type"])
	if err != nil {
		return fmt.Errorf("%s.type: %w", path, err)
	}
	if len(types) > 0 {
		for _, item := range types {
			if !isSupportedSchemaType(item) {
				return fmt.Errorf("%s.type: unsupported type %q", path, item)
			}
		}
	}
	if err := validateSchemaKeywordTypeApplicability(schema, types, path); err != nil {
		return err
	}
	if rawConst, exists := schema["const"]; exists {
		if err := validateSchemaConstCompatibility(rawConst, schema, types); err != nil {
			return fmt.Errorf("%s.const: %w", path, err)
		}
	}
	if rawEnum, exists := schema["enum"]; exists {
		if err := validateSchemaEnumDefinition(rawEnum, schema, types); err != nil {
			return fmt.Errorf("%s.enum: %w", path, err)
		}
	}
	if required, exists := schema["required"]; exists {
		if _, err := readSchemaRequired(required); err != nil {
			return fmt.Errorf("%s.required: %w", path, err)
		}
	}
	if rawProps, exists := schema["properties"]; exists {
		props, ok := rawProps.(map[string]any)
		if !ok {
			return fmt.Errorf("%s.properties: must be an object", path)
		}
		keys := make([]string, 0, len(props))
		for key := range props {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			child, ok := props[key].(map[string]any)
			if !ok {
				return fmt.Errorf("%s.properties.%s: must be an object", path, key)
			}
			if err := validateSchemaDefinition(child, path+".properties."+key); err != nil {
				return err
			}
		}
	}
	if rawItems, exists := schema["items"]; exists {
		child, ok := rawItems.(map[string]any)
		if !ok {
			return fmt.Errorf("%s.items: must be an object", path)
		}
		if err := validateSchemaDefinition(child, path+".items"); err != nil {
			return err
		}
	}
	if rawAdditional, exists := schema["additionalProperties"]; exists {
		switch value := rawAdditional.(type) {
		case bool:
			_ = value
		case map[string]any:
			if err := validateSchemaDefinition(value, path+".additionalProperties"); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%s.additionalProperties: must be boolean or object", path)
		}
	}
	if rawPattern, exists := schema["pattern"]; exists {
		pattern, ok := rawPattern.(string)
		if !ok {
			return fmt.Errorf("%s.pattern: must be a string", path)
		}
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("%s.pattern: %w", path, err)
		}
	}
	for _, keyword := range []string{"minProperties", "maxProperties", "minItems", "maxItems", "minLength", "maxLength"} {
		if rawValue, exists := schema[keyword]; exists {
			if err := validateSchemaNonNegativeIntegerKeyword(rawValue); err != nil {
				return fmt.Errorf("%s.%s: %w", path, keyword, err)
			}
		}
	}
	for _, keyword := range []string{"minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum"} {
		if rawValue, exists := schema[keyword]; exists {
			if err := validateSchemaFiniteNumberKeyword(rawValue); err != nil {
				return fmt.Errorf("%s.%s: %w", path, keyword, err)
			}
		}
	}
	if err := validateSchemaKeywordRanges(schema, path); err != nil {
		return err
	}
	return nil
}

func validateSchemaKeywordTypeApplicability(schema map[string]any, types []string, path string) error {
	if len(types) == 0 {
		return nil
	}
	if err := validateSchemaKeywordsRequireType(schema, types, path, []string{
		"properties", "required", "additionalProperties", "minProperties", "maxProperties",
	}, "object"); err != nil {
		return err
	}
	if err := validateSchemaKeywordsRequireType(schema, types, path, []string{
		"items", "minItems", "maxItems",
	}, "array"); err != nil {
		return err
	}
	if err := validateSchemaKeywordsRequireType(schema, types, path, []string{
		"pattern", "minLength", "maxLength",
	}, "string"); err != nil {
		return err
	}
	if err := validateSchemaKeywordsRequireAnyType(schema, types, path, []string{
		"minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum",
	}, []string{"number", "integer"}); err != nil {
		return err
	}
	return nil
}

func validateSchemaKeywordsRequireType(schema map[string]any, types []string, path string, keywords []string, requiredType string) error {
	return validateSchemaKeywordsRequireAnyType(schema, types, path, keywords, []string{requiredType})
}

func validateSchemaKeywordsRequireAnyType(schema map[string]any, types []string, path string, keywords []string, requiredTypes []string) error {
	for _, keyword := range keywords {
		if _, exists := schema[keyword]; !exists {
			continue
		}
		if schemaTypesIncludeAny(types, requiredTypes...) {
			continue
		}
		return fmt.Errorf("%s.%s: requires declared type to include %s", path, keyword, strings.Join(requiredTypes, " or "))
	}
	return nil
}

func schemaTypesIncludeAny(types []string, candidates ...string) bool {
	for _, item := range types {
		for _, candidate := range candidates {
			if item == candidate {
				return true
			}
		}
	}
	return false
}

func validateJSONSchemaSubset(value any, schema map[string]any, path string) error {
	if schema == nil {
		return nil
	}
	if err := validateSchemaTypeConstraint(value, schema, path); err != nil {
		return err
	}
	if err := validateSchemaEnumConstraint(value, schema, path); err != nil {
		return err
	}
	if err := validateSchemaConstConstraint(value, schema, path); err != nil {
		return err
	}
	switch typed := value.(type) {
	case map[string]any:
		if err := validateObjectConstraints(typed, schema, path); err != nil {
			return err
		}
	case []any:
		if err := validateArrayConstraints(typed, schema, path); err != nil {
			return err
		}
	case string:
		if err := validateStringConstraints(typed, schema, path); err != nil {
			return err
		}
	case float64:
		if err := validateNumberConstraints(typed, schema, path); err != nil {
			return err
		}
	case nil:
		// covered by type/enum/const checks.
	}
	return nil
}

func validateSchemaTypeConstraint(value any, schema map[string]any, path string) error {
	types, err := readSchemaTypes(schema["type"])
	if err != nil {
		return fmt.Errorf("%s.type: %w", path, err)
	}
	if len(types) == 0 {
		return nil
	}
	for _, schemaType := range types {
		if matchesSchemaType(value, schemaType) {
			return nil
		}
	}
	return fmt.Errorf("%s: type mismatch", path)
}

func validateSchemaEnumConstraint(value any, schema map[string]any, path string) error {
	rawEnum, ok := schema["enum"]
	if !ok {
		return nil
	}
	enumItems, ok := schemaLiteralSlice(rawEnum)
	if !ok || len(enumItems) == 0 {
		return fmt.Errorf("%s.enum: must be a non-empty array", path)
	}
	for _, item := range enumItems {
		if schemaValuesEqual(value, item) {
			return nil
		}
	}
	return fmt.Errorf("%s: value not in enum", path)
}

func validateSchemaConstConstraint(value any, schema map[string]any, path string) error {
	rawConst, ok := schema["const"]
	if !ok {
		return nil
	}
	if schemaValuesEqual(value, rawConst) {
		return nil
	}
	return fmt.Errorf("%s: value does not match const", path)
}

func validateObjectConstraints(object map[string]any, schema map[string]any, path string) error {
	if minProps, ok := readSchemaInt(schema["minProperties"]); ok && len(object) < minProps {
		return fmt.Errorf("%s: property count below minProperties", path)
	}
	if maxProps, ok := readSchemaInt(schema["maxProperties"]); ok && len(object) > maxProps {
		return fmt.Errorf("%s: property count above maxProperties", path)
	}
	required, err := readSchemaRequired(schema["required"])
	if err != nil {
		return fmt.Errorf("%s.required: %w", path, err)
	}
	for _, key := range required {
		if _, exists := object[key]; !exists {
			return fmt.Errorf("%s.%s: required property missing", path, key)
		}
	}
	properties := map[string]any{}
	if rawProps, exists := schema["properties"]; exists {
		typed, ok := rawProps.(map[string]any)
		if !ok {
			return fmt.Errorf("%s.properties: must be an object", path)
		}
		properties = typed
	}
	additional := schema["additionalProperties"]
	for key, item := range object {
		childPath := path + "." + key
		if childSchemaRaw, exists := properties[key]; exists {
			childSchema, ok := childSchemaRaw.(map[string]any)
			if !ok {
				return fmt.Errorf("%s: property schema must be object", childPath)
			}
			if err := validateJSONSchemaSubset(item, childSchema, childPath); err != nil {
				return err
			}
			continue
		}
		switch typed := additional.(type) {
		case nil:
			continue
		case bool:
			if !typed {
				return fmt.Errorf("%s: additional property is not allowed", childPath)
			}
		case map[string]any:
			if err := validateJSONSchemaSubset(item, typed, childPath); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%s.additionalProperties: must be boolean or object", path)
		}
	}
	return nil
}

func validateArrayConstraints(items []any, schema map[string]any, path string) error {
	if minItems, ok := readSchemaInt(schema["minItems"]); ok && len(items) < minItems {
		return fmt.Errorf("%s: item count below minItems", path)
	}
	if maxItems, ok := readSchemaInt(schema["maxItems"]); ok && len(items) > maxItems {
		return fmt.Errorf("%s: item count above maxItems", path)
	}
	rawItems, exists := schema["items"]
	if !exists {
		return nil
	}
	itemSchema, ok := rawItems.(map[string]any)
	if !ok {
		return fmt.Errorf("%s.items: must be an object", path)
	}
	for idx, item := range items {
		if err := validateJSONSchemaSubset(item, itemSchema, fmt.Sprintf("%s[%d]", path, idx)); err != nil {
			return err
		}
	}
	return nil
}

func validateStringConstraints(value string, schema map[string]any, path string) error {
	if minLength, ok := readSchemaInt(schema["minLength"]); ok && len(value) < minLength {
		return fmt.Errorf("%s: length below minLength", path)
	}
	if maxLength, ok := readSchemaInt(schema["maxLength"]); ok && len(value) > maxLength {
		return fmt.Errorf("%s: length above maxLength", path)
	}
	if rawPattern, exists := schema["pattern"]; exists {
		pattern, ok := rawPattern.(string)
		if !ok {
			return fmt.Errorf("%s.pattern: must be a string", path)
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("%s.pattern: %w", path, err)
		}
		if !re.MatchString(value) {
			return fmt.Errorf("%s: value does not match pattern", path)
		}
	}
	return nil
}

func validateNumberConstraints(value float64, schema map[string]any, path string) error {
	if minimum, ok := readSchemaNumber(schema["minimum"]); ok && value < minimum {
		return fmt.Errorf("%s: value below minimum", path)
	}
	if maximum, ok := readSchemaNumber(schema["maximum"]); ok && value > maximum {
		return fmt.Errorf("%s: value above maximum", path)
	}
	if minimum, ok := readSchemaNumber(schema["exclusiveMinimum"]); ok && value <= minimum {
		return fmt.Errorf("%s: value not above exclusiveMinimum", path)
	}
	if maximum, ok := readSchemaNumber(schema["exclusiveMaximum"]); ok && value >= maximum {
		return fmt.Errorf("%s: value not below exclusiveMaximum", path)
	}
	return nil
}

func readSchemaTypes(raw any) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	switch typed := raw.(type) {
	case string:
		value, err := normalizeSchemaTypeEntry(typed)
		if err != nil {
			return nil, err
		}
		return []string{value}, nil
	case []string:
		seen := map[string]bool{}
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			value, err := normalizeSchemaTypeEntry(item)
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
			value, err := normalizeSchemaTypeEntry(text)
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

func readSchemaRequired(raw any) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	switch typed := raw.(type) {
	case []string:
		seen := map[string]bool{}
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			name, err := normalizeSchemaRequiredEntry(item)
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
			name, err := normalizeSchemaRequiredEntry(text)
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

func readSchemaInt(raw any) (int, bool) {
	number, ok := readSchemaNumber(raw)
	if !ok {
		return 0, false
	}
	if math.Trunc(number) != number {
		return 0, false
	}
	return int(number), true
}

func readSchemaNumber(raw any) (float64, bool) {
	if raw == nil {
		return 0, false
	}
	switch value := raw.(type) {
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return 0, false
		}
		return value, true
	case float32:
		num := float64(value)
		if math.IsNaN(num) || math.IsInf(num, 0) {
			return 0, false
		}
		return num, true
	case int:
		return float64(value), true
	case int8:
		return float64(value), true
	case int16:
		return float64(value), true
	case int32:
		return float64(value), true
	case int64:
		return float64(value), true
	case uint:
		return float64(value), true
	case uint8:
		return float64(value), true
	case uint16:
		return float64(value), true
	case uint32:
		return float64(value), true
	case uint64:
		return float64(value), true
	default:
		return 0, false
	}
}

func validateSchemaEnumDefinition(raw any, schema map[string]any, types []string) error {
	items, ok := schemaLiteralSlice(raw)
	if !ok || len(items) == 0 {
		return fmt.Errorf("must be a non-empty array")
	}
	for idx, item := range items {
		for prev := 0; prev < idx; prev++ {
			if schemaValuesEqual(item, items[prev]) {
				return fmt.Errorf("[%d]: must not contain duplicate entries", idx)
			}
		}
		if err := validateSchemaLiteralMatchesTypes(item, types); err != nil {
			return fmt.Errorf("[%d]: %w", idx, err)
		}
		if err := validateSchemaLiteralAgainstDefinition(item, schema); err != nil {
			return fmt.Errorf("[%d]: %w", idx, err)
		}
	}
	return nil
}

func validateSchemaConstCompatibility(raw any, schema map[string]any, types []string) error {
	if err := validateSchemaLiteralMatchesTypes(raw, types); err != nil {
		return err
	}
	return validateSchemaLiteralAgainstDefinition(raw, schema)
}

func validateSchemaLiteralMatchesTypes(raw any, types []string) error {
	if len(types) == 0 {
		return nil
	}
	for _, schemaType := range types {
		if schemaValueMatchesType(raw, schemaType) {
			return nil
		}
	}
	return fmt.Errorf("does not match declared type constraint")
}

func validateSchemaLiteralAgainstDefinition(raw any, schema map[string]any) error {
	if err := validateSchemaLiteralMatchesTypes(raw, readSchemaTypesOrNil(schema["type"])); err != nil {
		return err
	}
	switch value := raw.(type) {
	case string:
		return validateSchemaLiteralStringConstraints(value, schema)
	case int:
		return validateSchemaLiteralNumberConstraints(float64(value), schema)
	case int8:
		return validateSchemaLiteralNumberConstraints(float64(value), schema)
	case int16:
		return validateSchemaLiteralNumberConstraints(float64(value), schema)
	case int32:
		return validateSchemaLiteralNumberConstraints(float64(value), schema)
	case int64:
		return validateSchemaLiteralNumberConstraints(float64(value), schema)
	case uint:
		return validateSchemaLiteralNumberConstraints(float64(value), schema)
	case uint8:
		return validateSchemaLiteralNumberConstraints(float64(value), schema)
	case uint16:
		return validateSchemaLiteralNumberConstraints(float64(value), schema)
	case uint32:
		return validateSchemaLiteralNumberConstraints(float64(value), schema)
	case uint64:
		return validateSchemaLiteralNumberConstraints(float64(value), schema)
	case float32:
		return validateSchemaLiteralNumberConstraints(float64(value), schema)
	case float64:
		return validateSchemaLiteralNumberConstraints(value, schema)
	}
	if object, ok := schemaLiteralMap(raw); ok {
		return validateSchemaLiteralObjectConstraints(object, schema)
	}
	if items, ok := schemaLiteralSlice(raw); ok {
		return validateSchemaLiteralArrayConstraints(items, schema)
	}
	return nil
}

func readSchemaTypesOrNil(raw any) []string {
	types, err := readSchemaTypes(raw)
	if err != nil {
		return nil
	}
	return types
}

func validateSchemaLiteralObjectConstraints(object map[string]any, schema map[string]any) error {
	if minProps, ok, err := readSchemaNonNegativeIntegerKeyword(schema["minProperties"]); err == nil && ok && float64(len(object)) < minProps {
		return fmt.Errorf("violates minProperties")
	} else if err != nil {
		return err
	}
	if maxProps, ok, err := readSchemaNonNegativeIntegerKeyword(schema["maxProperties"]); err == nil && ok && float64(len(object)) > maxProps {
		return fmt.Errorf("violates maxProperties")
	} else if err != nil {
		return err
	}
	required, err := readSchemaRequired(schema["required"])
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
			if err := validateSchemaLiteralAgainstDefinition(item, childSchema); err != nil {
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
			if err := validateSchemaLiteralAgainstDefinition(item, typed); err != nil {
				return err
			}
		default:
			return fmt.Errorf("additionalProperties: must be boolean or object")
		}
	}
	return nil
}

func validateSchemaLiteralArrayConstraints(items []any, schema map[string]any) error {
	if minItems, ok, err := readSchemaNonNegativeIntegerKeyword(schema["minItems"]); err == nil && ok && float64(len(items)) < minItems {
		return fmt.Errorf("violates minItems")
	} else if err != nil {
		return err
	}
	if maxItems, ok, err := readSchemaNonNegativeIntegerKeyword(schema["maxItems"]); err == nil && ok && float64(len(items)) > maxItems {
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
		if err := validateSchemaLiteralAgainstDefinition(item, itemSchema); err != nil {
			return err
		}
	}
	return nil
}

func validateSchemaLiteralStringConstraints(value string, schema map[string]any) error {
	if minLength, ok, err := readSchemaNonNegativeIntegerKeyword(schema["minLength"]); err == nil && ok && float64(len(value)) < minLength {
		return fmt.Errorf("violates minLength")
	} else if err != nil {
		return err
	}
	if maxLength, ok, err := readSchemaNonNegativeIntegerKeyword(schema["maxLength"]); err == nil && ok && float64(len(value)) > maxLength {
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

func validateSchemaLiteralNumberConstraints(value float64, schema map[string]any) error {
	if minimum, ok, err := readSchemaFiniteNumberKeyword(schema["minimum"]); err == nil && ok && value < minimum {
		return fmt.Errorf("violates minimum")
	} else if err != nil {
		return err
	}
	if maximum, ok, err := readSchemaFiniteNumberKeyword(schema["maximum"]); err == nil && ok && value > maximum {
		return fmt.Errorf("violates maximum")
	} else if err != nil {
		return err
	}
	if minimum, ok, err := readSchemaFiniteNumberKeyword(schema["exclusiveMinimum"]); err == nil && ok && value <= minimum {
		return fmt.Errorf("violates exclusiveMinimum")
	} else if err != nil {
		return err
	}
	if maximum, ok, err := readSchemaFiniteNumberKeyword(schema["exclusiveMaximum"]); err == nil && ok && value >= maximum {
		return fmt.Errorf("violates exclusiveMaximum")
	} else if err != nil {
		return err
	}
	return nil
}

func validateSchemaNonNegativeIntegerKeyword(raw any) error {
	number, ok := readSchemaNumber(raw)
	if !ok {
		return fmt.Errorf("must be a non-negative integer")
	}
	if math.Trunc(number) != number {
		return fmt.Errorf("must be an integer")
	}
	if number < 0 {
		return fmt.Errorf("must be >= 0")
	}
	return nil
}

func validateSchemaFiniteNumberKeyword(raw any) error {
	if _, ok := readSchemaNumber(raw); !ok {
		return fmt.Errorf("must be a number")
	}
	return nil
}

func validateSchemaKeywordRanges(schema map[string]any, path string) error {
	if min, ok, err := readSchemaNonNegativeIntegerKeyword(schema["minProperties"]); err != nil {
		return fmt.Errorf("%s.minProperties: %w", path, err)
	} else if max, okMax, err := readSchemaNonNegativeIntegerKeyword(schema["maxProperties"]); err != nil {
		return fmt.Errorf("%s.maxProperties: %w", path, err)
	} else if ok && okMax && min > max {
		return fmt.Errorf("%s.minProperties: must be <= maxProperties", path)
	}
	if min, ok, err := readSchemaNonNegativeIntegerKeyword(schema["minItems"]); err != nil {
		return fmt.Errorf("%s.minItems: %w", path, err)
	} else if max, okMax, err := readSchemaNonNegativeIntegerKeyword(schema["maxItems"]); err != nil {
		return fmt.Errorf("%s.maxItems: %w", path, err)
	} else if ok && okMax && min > max {
		return fmt.Errorf("%s.minItems: must be <= maxItems", path)
	}
	if min, ok, err := readSchemaNonNegativeIntegerKeyword(schema["minLength"]); err != nil {
		return fmt.Errorf("%s.minLength: %w", path, err)
	} else if max, okMax, err := readSchemaNonNegativeIntegerKeyword(schema["maxLength"]); err != nil {
		return fmt.Errorf("%s.maxLength: %w", path, err)
	} else if ok && okMax && min > max {
		return fmt.Errorf("%s.minLength: must be <= maxLength", path)
	}
	if min, ok, err := readSchemaFiniteNumberKeyword(schema["minimum"]); err != nil {
		return fmt.Errorf("%s.minimum: %w", path, err)
	} else if max, okMax, err := readSchemaFiniteNumberKeyword(schema["maximum"]); err != nil {
		return fmt.Errorf("%s.maximum: %w", path, err)
	} else if ok && okMax && min > max {
		return fmt.Errorf("%s.minimum: must be <= maximum", path)
	}
	if min, ok, err := readSchemaFiniteNumberKeyword(schema["exclusiveMinimum"]); err != nil {
		return fmt.Errorf("%s.exclusiveMinimum: %w", path, err)
	} else if max, okMax, err := readSchemaFiniteNumberKeyword(schema["exclusiveMaximum"]); err != nil {
		return fmt.Errorf("%s.exclusiveMaximum: %w", path, err)
	} else if ok && okMax && min >= max {
		return fmt.Errorf("%s.exclusiveMinimum: must be < exclusiveMaximum", path)
	}
	return nil
}

func readSchemaNonNegativeIntegerKeyword(raw any) (float64, bool, error) {
	if raw == nil {
		return 0, false, nil
	}
	if err := validateSchemaNonNegativeIntegerKeyword(raw); err != nil {
		return 0, false, err
	}
	value, ok := readSchemaNumber(raw)
	return value, ok, nil
}

func readSchemaFiniteNumberKeyword(raw any) (float64, bool, error) {
	if raw == nil {
		return 0, false, nil
	}
	if err := validateSchemaFiniteNumberKeyword(raw); err != nil {
		return 0, false, err
	}
	value, ok := readSchemaNumber(raw)
	return value, ok, nil
}

func normalizeSchemaTypeEntry(raw string) (string, error) {
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

func normalizeSchemaRequiredEntry(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("must not contain empty entries")
	}
	if raw != trimmed {
		return "", fmt.Errorf("must not include surrounding whitespace")
	}
	return trimmed, nil
}

func schemaLiteralMap(raw any) (map[string]any, bool) {
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

func schemaLiteralSlice(raw any) ([]any, bool) {
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

func schemaValueMatchesType(value any, schemaType string) bool {
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
		return schemaValueAsFiniteNumber(value, false)
	case "integer":
		return schemaValueAsFiniteNumber(value, true)
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "null":
		return value == nil
	default:
		return false
	}
}

func schemaValueAsFiniteNumber(value any, integerOnly bool) bool {
	number, ok := readSchemaNumber(value)
	if !ok {
		return false
	}
	return !integerOnly || math.Trunc(number) == number
}

func schemaValuesEqual(left any, right any) bool {
	return reflect.DeepEqual(normalizeSchemaComparable(left), normalizeSchemaComparable(right))
}

func normalizeSchemaComparable(value any) any {
	if number, ok := readSchemaNumber(value); ok {
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
			out[iter.Key().String()] = normalizeSchemaComparable(iter.Value().Interface())
		}
		return out
	case reflect.Slice, reflect.Array:
		out := make([]any, rv.Len())
		for idx := 0; idx < rv.Len(); idx++ {
			out[idx] = normalizeSchemaComparable(rv.Index(idx).Interface())
		}
		return out
	default:
		return value
	}
}

func readFloat(params map[string]any, key string) (float64, bool) {
	if params == nil {
		return 0, false
	}
	raw, ok := params[key]
	if !ok || raw == nil {
		return 0, false
	}
	switch value := raw.(type) {
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return 0, false
		}
		return value, true
	case float32:
		num := float64(value)
		if math.IsNaN(num) || math.IsInf(num, 0) {
			return 0, false
		}
		return num, true
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	default:
		return 0, false
	}
}

func isSupportedSchemaType(schemaType string) bool {
	switch schemaType {
	case "object", "array", "string", "number", "integer", "boolean", "null":
		return true
	default:
		return false
	}
}

func matchesSchemaType(value any, schemaType string) bool {
	switch schemaType {
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok := value.(float64)
		return ok
	case "integer":
		number, ok := value.(float64)
		return ok && math.Trunc(number) == number
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "null":
		return value == nil
	default:
		return false
	}
}
