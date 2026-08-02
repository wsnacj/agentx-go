package hostkit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	types "github.com/wsnacj/agentx-go/components/llm"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

const DefaultSource = "agentx_docparse_livekit"

// ResultLoader reads a Host-approved local parse-result artifact. The
// canonical Host Kit never reads the filesystem by itself.
type ResultLoader func(path string) ([]byte, error)

type Config struct {
	Executor                 agentxtools.Executor
	ResultLoader             ResultLoader
	Source                   string
	RuntimeParseTool         string
	RuntimeSpecRecommendTool string
}

func DefaultConfig() Config {
	return Config{
		Source:                   DefaultSource,
		RuntimeParseTool:         RuntimeToolDocumentParse,
		RuntimeSpecRecommendTool: RuntimeToolDocumentSpecRecommend,
	}
}

type Kit struct {
	cfg Config
}

func New(cfg Config) *Kit {
	defaults := DefaultConfig()
	if strings.TrimSpace(cfg.Source) == "" {
		cfg.Source = defaults.Source
	}
	if strings.TrimSpace(cfg.RuntimeParseTool) == "" {
		cfg.RuntimeParseTool = defaults.RuntimeParseTool
	}
	if strings.TrimSpace(cfg.RuntimeSpecRecommendTool) == "" {
		cfg.RuntimeSpecRecommendTool = defaults.RuntimeSpecRecommendTool
	}
	return &Kit{cfg: cfg}
}

func BuildStandardToolHandlers(cfg Config) ToolHandlers {
	kit := New(cfg)
	return ToolHandlers{
		SpecSelect: func(ctx context.Context, params map[string]any) (any, error) {
			return kit.SpecSelect(ctx, params)
		},
		ProfileProbe: func(ctx context.Context, params map[string]any) (any, error) {
			return kit.ProfileProbe(ctx, params)
		},
		ExtractFields: func(ctx context.Context, params map[string]any) (any, error) {
			return kit.ExtractFields(ctx, params)
		},
		ExtractTable: func(ctx context.Context, params map[string]any) (any, error) {
			return kit.ExtractTable(ctx, params)
		},
		TraceEvidence: func(ctx context.Context, params map[string]any) (any, error) {
			return kit.TraceEvidence(ctx, params)
		},
		Validate: func(ctx context.Context, params map[string]any) (any, error) {
			return kit.Validate(ctx, params)
		},
		Guard: func(ctx context.Context, params map[string]any) (any, error) {
			return kit.Guard(ctx, params)
		},
	}
}

func (k *Kit) SpecSelect(ctx context.Context, params map[string]any) (any, error) {
	return k.executeRuntime(ctx, ToolDocparseSpecSelect, k.cfg.RuntimeSpecRecommendTool, cloneParams(params))
}

func (k *Kit) ProfileProbe(ctx context.Context, params map[string]any) (any, error) {
	args := cloneParams(params)
	args["task_kind"] = "document.profile_probe"
	args["profile_probe_only"] = true
	if hasLocalParseResult(args) {
		return k.buildProfileProbePayload(args), nil
	}
	runtimeOut, err := k.executeRuntime(ctx, ToolDocparseProfileProbe, k.cfg.RuntimeParseTool, args)
	if err != nil {
		return nil, err
	}
	if projected, ok := k.projectRuntimeParseOutput(ToolDocparseProfileProbe, args, runtimeOut); ok {
		return projected, nil
	}
	return runtimeOut, nil
}

func (k *Kit) ExtractFields(ctx context.Context, params map[string]any) (any, error) {
	args := cloneParams(params)
	args["task_kind"] = firstNonEmptyString(args["task_kind"], "document.extract_fields")
	if hasLocalParseResult(args) {
		return k.buildLocalFieldExtractionPayload(args), nil
	}
	runtimeOut, err := k.executeRuntime(ctx, ToolDocparseExtractFields, k.cfg.RuntimeParseTool, args)
	if err != nil {
		return nil, err
	}
	if projected, ok := k.projectRuntimeParseOutput(ToolDocparseExtractFields, args, runtimeOut); ok {
		return projected, nil
	}
	return runtimeOut, nil
}

func (k *Kit) ExtractTable(ctx context.Context, params map[string]any) (any, error) {
	args := cloneParams(params)
	args["task_kind"] = firstNonEmptyString(args["task_kind"], "document.extract_table")
	if hasLocalParseResult(args) {
		return k.buildLocalTableExtractionPayload(args), nil
	}
	runtimeOut, err := k.executeRuntime(ctx, ToolDocparseExtractTable, k.cfg.RuntimeParseTool, args)
	if err != nil {
		return nil, err
	}
	if projected, ok := k.projectRuntimeParseOutput(ToolDocparseExtractTable, args, runtimeOut); ok {
		return projected, nil
	}
	return runtimeOut, nil
}

func (k *Kit) TraceEvidence(_ context.Context, params map[string]any) (any, error) {
	return k.buildEvidencePayload(ToolDocparseTraceEvidence, params), nil
}

func (k *Kit) Validate(_ context.Context, params map[string]any) (any, error) {
	evidence := k.buildEvidencePayload(ToolDocparseTraceEvidence, params)
	return validateEvidencePayload(ToolDocparseValidate, k.cfg.Source, params, evidence), nil
}

func (k *Kit) Guard(_ context.Context, params map[string]any) (any, error) {
	evidence := k.buildEvidencePayload(ToolDocparseTraceEvidence, params)
	validation := validateEvidencePayload(ToolDocparseGuard, k.cfg.Source, params, evidence)
	if validation.Passed {
		validation.Status = "answer_ready"
		validation.Summary = "docparse guard passed"
		return validation, nil
	}
	if validation.ReviewRequired {
		validation.Status = "review_required"
		if validation.Summary == "" {
			validation.Summary = "docparse guard requires review"
		}
		return validation, nil
	}
	validation.Status = "failed"
	if validation.Summary == "" {
		validation.Summary = "docparse guard failed"
	}
	return validation, nil
}

func (k *Kit) executeRuntime(ctx context.Context, semanticTool string, runtimeTool string, args map[string]any) (any, error) {
	runtimeTool = strings.TrimSpace(runtimeTool)
	if runtimeTool == "" {
		runtimeTool = RuntimeToolDocumentParse
	}
	if k.cfg.Executor == nil {
		return k.unsupportedPayload(semanticTool, runtimeTool, args), nil
	}
	blob, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%s: encode runtime arguments: %w", semanticTool, err)
	}
	return k.cfg.Executor.Execute(ctx, types.FunctionCall{
		Name:      runtimeTool,
		Arguments: string(blob),
	})
}

func (k *Kit) projectRuntimeParseOutput(semanticTool string, args map[string]any, runtimeOut any) (any, bool) {
	raw, ok := runtimeParseOutputMap(runtimeOut)
	if !ok {
		return nil, false
	}
	if strings.TrimSpace(fmt.Sprint(raw["failure_code"])) == "docparse_parser_executor_not_configured" {
		return nil, false
	}
	next := cloneParams(args)
	next["parse_result"] = raw
	switch semanticTool {
	case ToolDocparseProfileProbe:
		return k.buildProfileProbePayload(next), true
	case ToolDocparseExtractFields:
		return k.buildLocalFieldExtractionPayload(next), true
	case ToolDocparseExtractTable:
		return k.buildLocalTableExtractionPayload(next), true
	default:
		return nil, false
	}
}

func runtimeParseOutputMap(runtimeOut any) (map[string]any, bool) {
	switch typed := runtimeOut.(type) {
	case map[string]any:
		return typed, true
	case string:
		out := map[string]any{}
		if err := json.Unmarshal([]byte(typed), &out); err == nil && len(out) > 0 {
			return out, true
		}
	case []byte:
		out := map[string]any{}
		if err := json.Unmarshal(typed, &out); err == nil && len(out) > 0 {
			return out, true
		}
	}
	return nil, false
}

func (k *Kit) unsupportedPayload(semanticTool string, runtimeTool string, args map[string]any) map[string]any {
	payload := map[string]any{
		"tool":            semanticTool,
		"source":          k.cfg.Source,
		"status":          "unsupported",
		"adapter_status":  "unsupported",
		"failure_code":    "docparse_parser_executor_not_configured",
		"failure_class":   "parser_capability_missing",
		"missing":         "livekit.executor",
		"runtime_tool":    runtimeTool,
		"passed":          false,
		"review_required": true,
		"summary":         "docparse host executor is not configured",
	}
	if value := firstStringArg(args, "document_path", "result_path"); value != "" {
		payload["document_path"] = value
	}
	if value := firstStringArg(args, "spec_path"); value != "" {
		payload["spec_path"] = value
	}
	return payload
}

func cloneParams(params map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range params {
		if strings.TrimSpace(key) == "" {
			continue
		}
		out[key] = value
	}
	return out
}

func firstNonEmptyString(value any, fallback string) string {
	if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
		return strings.TrimSpace(text)
	}
	return fallback
}

func firstStringArg(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
