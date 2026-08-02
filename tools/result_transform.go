package tools

import (
	"encoding/json"
	"strings"
)

const (
	ToolResultMiddlewareModeControlledTransform = "controlled_transform"

	ToolResultTransformReasonMissingRawResultRef = "missing_raw_result_ref"
)

type ToolResultTransformInput struct {
	Event        ToolResultMiddlewareEvent
	Output       string
	RawResultRef string
	ArtifactRefs []string
}

type ToolResultTransformResult struct {
	Applied        bool                        `json:"applied"`
	Mode           string                      `json:"mode"`
	Output         string                      `json:"output"`
	RawResultRef   string                      `json:"raw_result_ref,omitempty"`
	ArtifactRefs   []string                    `json:"artifact_refs,omitempty"`
	Reasons        []string                    `json:"reasons,omitempty"`
	Strategies     []string                    `json:"strategies,omitempty"`
	Details        []string                    `json:"details,omitempty"`
	ErrorPreserved bool                        `json:"error_preserved,omitempty"`
	Summary        *ToolResultTransformSummary `json:"summary,omitempty"`
}

type ToolResultTransformSummary struct {
	ToolName           string                  `json:"tool_name,omitempty"`
	ContentBytes       int                     `json:"content_bytes,omitempty"`
	ContentChars       int                     `json:"content_chars,omitempty"`
	ContentLines       int                     `json:"content_lines,omitempty"`
	OriginalBytes      int                     `json:"original_bytes,omitempty"`
	KeptBytes          int                     `json:"kept_bytes,omitempty"`
	Truncated          bool                    `json:"truncated,omitempty"`
	GuardStrategy      string                  `json:"guard_strategy,omitempty"`
	JSONValid          bool                    `json:"json_valid,omitempty"`
	JSONTopLevelKeys   []string                `json:"json_top_level_keys,omitempty"`
	SensitiveKeyHits   []string                `json:"sensitive_key_hits,omitempty"`
	EmbeddedBinaryKeys []string                `json:"embedded_binary_keys,omitempty"`
	ExternalContent    bool                    `json:"external_content,omitempty"`
	UntrustedContent   bool                    `json:"untrusted_content,omitempty"`
	RawResultRef       string                  `json:"raw_result_ref,omitempty"`
	ArtifactRefs       []string                `json:"artifact_refs,omitempty"`
	OutputSchema       *ToolResultOutputSchema `json:"output_schema,omitempty"`
}

func BuildControlledToolResultTransform(input ToolResultTransformInput) ToolResultTransformResult {
	event := input.Event
	rawResultRef := firstNonBlank(input.RawResultRef, event.RawResultRef)
	artifactRefs := normalizeToolResultRefs(append(append([]string(nil), event.ArtifactRefs...), input.ArtifactRefs...))
	base := ToolResultTransformResult{
		Applied:      false,
		Mode:         ToolResultMiddlewareModeControlledTransform,
		Output:       input.Output,
		RawResultRef: rawResultRef,
		ArtifactRefs: artifactRefs,
		Reasons:      append([]string(nil), event.WouldTransformReasons...),
		Strategies:   append([]string(nil), event.SuggestedStrategies...),
	}
	if event.IsError {
		base.ErrorPreserved = true
		base.Details = append(base.Details, "error_result_preserved")
		return base
	}
	if !event.WouldTransform {
		return base
	}
	if toolResultTransformNeedsRawRef(event.WouldTransformReasons) && rawResultRef == "" && len(artifactRefs) == 0 {
		base.Details = append(base.Details, ToolResultTransformReasonMissingRawResultRef)
		return base
	}
	summary := buildToolResultTransformSummary(event, rawResultRef, artifactRefs)
	envelope := map[string]any{
		"tool_result_summary": summary,
	}
	blob, err := json.Marshal(envelope)
	if err != nil {
		return base
	}
	base.Applied = true
	base.Output = string(blob)
	base.Summary = &summary
	base.Strategies = uniqueSortedStrings(append(base.Strategies, ToolResultMiddlewareStrategySummarizeWithRawRef))
	return base
}

func AppendToolResultTransformTelemetryAttrs(attrs map[string]any, result ToolResultTransformResult) map[string]any {
	if attrs == nil {
		attrs = map[string]any{}
	}
	attrs["result_middleware_applied"] = result.Applied
	attrs["result_middleware_mode"] = strings.TrimSpace(result.Mode)
	if result.ErrorPreserved {
		attrs["result_middleware_error_preserved"] = true
	}
	if len(result.Reasons) > 0 {
		attrs["result_middleware_reasons"] = strings.Join(result.Reasons, ",")
	}
	if len(result.Strategies) > 0 {
		attrs["result_middleware_strategies"] = strings.Join(result.Strategies, ",")
	}
	if rawRef := strings.TrimSpace(result.RawResultRef); rawRef != "" {
		attrs["result_middleware_raw_result_ref"] = rawRef
	}
	if len(result.ArtifactRefs) > 0 {
		attrs["result_middleware_artifact_refs"] = strings.Join(result.ArtifactRefs, ",")
	}
	return attrs
}

func buildToolResultTransformSummary(event ToolResultMiddlewareEvent, rawResultRef string, artifactRefs []string) ToolResultTransformSummary {
	return ToolResultTransformSummary{
		ToolName:           event.ToolName,
		ContentBytes:       event.Content.Bytes,
		ContentChars:       event.Content.Chars,
		ContentLines:       event.Content.Lines,
		OriginalBytes:      event.Content.OriginalBytes,
		KeptBytes:          event.Content.KeptBytes,
		Truncated:          event.Content.Truncated,
		GuardStrategy:      event.Content.GuardStrategy,
		JSONValid:          event.Content.JSONValid,
		JSONTopLevelKeys:   append([]string(nil), event.Content.JSONTopLevelKeys...),
		SensitiveKeyHits:   append([]string(nil), event.Content.SensitiveKeyHits...),
		EmbeddedBinaryKeys: append([]string(nil), event.Content.EmbeddedBinaryKeys...),
		ExternalContent:    event.ExternalContent,
		UntrustedContent:   event.UntrustedContent,
		RawResultRef:       strings.TrimSpace(rawResultRef),
		ArtifactRefs:       append([]string(nil), artifactRefs...),
		OutputSchema:       cloneToolResultOutputSchema(event.OutputSchema),
	}
}

func cloneToolResultOutputSchema(schema *ToolResultOutputSchema) *ToolResultOutputSchema {
	if schema == nil {
		return nil
	}
	out := *schema
	out.Required = append([]string(nil), schema.Required...)
	out.MatchedTopLevelKeys = append([]string(nil), schema.MatchedTopLevelKeys...)
	out.MissingRequired = append([]string(nil), schema.MissingRequired...)
	out.UnexpectedTopLevelKeys = append([]string(nil), schema.UnexpectedTopLevelKeys...)
	return &out
}

func toolResultTransformNeedsRawRef(reasons []string) bool {
	for _, reason := range reasons {
		switch strings.TrimSpace(reason) {
		case ToolResultMiddlewareReasonLargeResult,
			ToolResultMiddlewareReasonRuntimeGuardTruncated,
			ToolResultMiddlewareReasonEmbeddedBinaryPayload:
			return true
		}
	}
	return false
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
