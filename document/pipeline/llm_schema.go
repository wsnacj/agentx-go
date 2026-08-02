package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/wsnacj/agentx-go/document/pipeline/configs"
	"sort"
	"strings"
)

const (
	llmReasonInvalidJSON    = "llm_invalid_json"
	llmReasonSchemaMismatch = "schema_mismatch"
	llmReasonRepairFailed   = "repair_failed"
	llmReasonExtraFields    = "llm_extra_fields"
)

type llmSchemaValidation struct {
	ValidValues    map[string]any
	MissingFields  []configs.FieldSpec
	MismatchFields []configs.FieldSpec
	ExtraFields    []string
}

type llmRepairPlan struct {
	Reason string
	Fields []configs.FieldSpec
}

type llmRepairOutcome struct {
	MissingReason map[string]string
	FieldWarnings map[string][]string
	DocWarnings   []string
}

func validateLLMFieldSchema(values map[string]any, fields []configs.FieldSpec) llmSchemaValidation {
	out := llmSchemaValidation{ValidValues: map[string]any{}}
	expected := map[string]configs.FieldSpec{}
	for _, field := range fields {
		expected[field.Key] = field
	}
	for key := range values {
		if _, ok := expected[key]; !ok {
			out.ExtraFields = append(out.ExtraFields, key)
		}
	}
	sort.Strings(out.ExtraFields)
	for _, field := range fields {
		value, ok := values[field.Key]
		if !ok || value == nil {
			out.MissingFields = append(out.MissingFields, field)
			continue
		}
		if !llmValueMatchesFieldType(field, value) {
			out.MismatchFields = append(out.MismatchFields, field)
			continue
		}
		out.ValidValues[field.Key] = value
	}
	return out
}

func llmValueMatchesFieldType(field configs.FieldSpec, value any) bool {
	switch strings.ToLower(strings.TrimSpace(field.Type)) {
	case "", "string", "date", "enum":
		switch value.(type) {
		case []any, []string, map[string]any:
			return false
		default:
			return true
		}
	case "number":
		switch v := value.(type) {
		case float64, float32, int, int64, json.Number:
			return true
		case string:
			return true
		default:
			_ = v
			return false
		}
	case "array":
		switch value.(type) {
		case []any, []string:
			return true
		default:
			return false
		}
	default:
		return true
	}
}

func buildLLMRepairPlan(values map[string]any, fields []configs.FieldSpec, initialErr error) (map[string]any, llmRepairPlan, llmRepairOutcome, bool) {
	outcome := llmRepairOutcome{
		MissingReason: map[string]string{},
		FieldWarnings: map[string][]string{},
	}
	if initialErr != nil {
		if isLLMInvalidJSONError(initialErr) {
			outcome.DocWarnings = append(outcome.DocWarnings, llmReasonInvalidJSON)
			return map[string]any{}, llmRepairPlan{Reason: llmReasonInvalidJSON, Fields: fields}, outcome, true
		}
		return values, llmRepairPlan{}, outcome, false
	}
	validation := validateLLMFieldSchema(values, fields)
	valid := validation.ValidValues
	if len(validation.ExtraFields) > 0 {
		outcome.DocWarnings = append(outcome.DocWarnings, llmReasonExtraFields)
	}
	repairFields := append([]configs.FieldSpec{}, validation.MissingFields...)
	repairFields = append(repairFields, validation.MismatchFields...)
	if len(repairFields) == 0 {
		return valid, llmRepairPlan{}, outcome, false
	}
	for _, field := range validation.MissingFields {
		outcome.FieldWarnings[field.Key] = append(outcome.FieldWarnings[field.Key], "llm_field_missing")
	}
	for _, field := range validation.MismatchFields {
		outcome.FieldWarnings[field.Key] = append(outcome.FieldWarnings[field.Key], llmReasonSchemaMismatch)
	}
	outcome.DocWarnings = append(outcome.DocWarnings, llmReasonSchemaMismatch)
	return valid, llmRepairPlan{Reason: llmReasonSchemaMismatch, Fields: repairFields}, outcome, true
}

func applyLLMRepairValues(values map[string]any, repairMap map[string]any, plan llmRepairPlan, outcome *llmRepairOutcome) map[string]any {
	if values == nil {
		values = map[string]any{}
	}
	validation := validateLLMFieldSchema(repairMap, plan.Fields)
	for key, value := range validation.ValidValues {
		values[key] = value
	}
	failed := append([]configs.FieldSpec{}, validation.MissingFields...)
	failed = append(failed, validation.MismatchFields...)
	for _, field := range failed {
		outcome.MissingReason[field.Key] = llmReasonRepairFailed
		outcome.FieldWarnings[field.Key] = append(outcome.FieldWarnings[field.Key], plan.Reason)
	}
	if len(failed) > 0 {
		outcome.DocWarnings = append(outcome.DocWarnings, llmReasonRepairFailed)
	}
	return values
}

func markLLMRepairFailure(plan llmRepairPlan, outcome *llmRepairOutcome) {
	for _, field := range plan.Fields {
		outcome.MissingReason[field.Key] = llmReasonRepairFailed
		outcome.FieldWarnings[field.Key] = append(outcome.FieldWarnings[field.Key], plan.Reason)
	}
	outcome.DocWarnings = append(outcome.DocWarnings, llmReasonRepairFailed)
}

func buildLLMRepairPrompt(basePrompt string, rawResponse string, reason string, fields []configs.FieldSpec) string {
	var sb strings.Builder
	sb.WriteString("你是信息抽取修复助手。上一轮抽取结果未通过校验，请只修复下面列出的字段。\n")
	sb.WriteString("要求：\n- 仅输出 JSON 对象\n- 只输出需要修复的字段\n- 缺失字段输出 null\n- 不要输出解释文字\n")
	if strings.TrimSpace(reason) != "" {
		sb.WriteString("- 修复原因: " + strings.TrimSpace(reason) + "\n")
	}
	if strings.TrimSpace(rawResponse) != "" {
		sb.WriteString("\n上一轮原始输出：\n")
		sb.WriteString(strings.TrimSpace(rawResponse))
		sb.WriteString("\n")
	}
	if strings.TrimSpace(basePrompt) != "" {
		sb.WriteString("\n原始抽取约束摘要：\n")
		sb.WriteString(strings.TrimSpace(basePrompt))
		sb.WriteString("\n")
	}
	sb.WriteString("\n需要修复的 JSON 字段：\n{\n")
	for i, field := range fields {
		comma := ","
		if i == len(fields)-1 {
			comma = ""
		}
		parts := []string{}
		if field.Description != "" {
			parts = append(parts, field.Description)
		}
		if field.Type != "" {
			parts = append(parts, "类型:"+field.Type)
		}
		if field.Unit != "" {
			parts = append(parts, "单位:"+field.Unit)
		}
		comment := strings.Join(parts, " | ")
		sb.WriteString(fmt.Sprintf("  \"%s\": null%s  // %s\n", field.Key, comma, comment))
	}
	sb.WriteString("}\n")
	return sb.String()
}

func llmRepairFieldKeys(fields []configs.FieldSpec) []string {
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		out = append(out, field.Key)
	}
	return out
}

func isLLMInvalidJSONError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "invalid json") || strings.Contains(msg, "invalid json returned by llm")
}
