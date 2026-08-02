package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	ToolResultMiddlewareModeObserveOnly = "observe_only"

	ToolResultMiddlewareReasonLargeResult           = "large_result"
	ToolResultMiddlewareReasonRuntimeGuardTruncated = "runtime_guard_truncated"
	ToolResultMiddlewareReasonSensitiveKeyPresent   = "sensitive_key_present"
	ToolResultMiddlewareReasonEmbeddedBinaryPayload = "embedded_binary_payload"

	ToolResultMiddlewareStrategySummarizeWithRawRef = "summarize_with_raw_ref"
	ToolResultMiddlewareStrategyPreserveRawArtifact = "preserve_raw_artifact"
	ToolResultMiddlewareStrategyRedactSensitiveKeys = "redact_sensitive_keys"
	ToolResultMiddlewareStrategyExtractArtifactRef  = "extract_artifact_ref"
)

const (
	defaultToolResultMiddlewareLargeBytes = 16 * 1024
	defaultToolResultMiddlewareLargeLines = 200
	toolResultMiddlewareMaxJSONInspect    = 1024 * 1024
	toolResultMiddlewareMaxKeys           = 16
)

type ToolResultMiddlewareInput struct {
	ToolName            string
	SessionID           string
	RunID               string
	Arguments           string
	Output              string
	OutputSchema        map[string]any
	RawResultRef        string
	ArtifactRefs        []string
	IsError             bool
	OutputOriginalBytes int
	OutputKeptBytes     int
	OutputTruncated     bool
	OutputGuardStrategy string
	LargeResultBytes    int
	LargeResultLines    int
	// ClassifyContent lets a Host preserve product-specific trust classification
	// without moving browser, document, process or network policy into this module.
	ClassifyContent ToolContentClassifier
}

// ToolContentClassifier classifies the trust boundary of one tool result.
// Implementations must be deterministic and must not perform side effects.
type ToolContentClassifier func(toolName string) (externalContent, untrustedContent bool)

type ToolResultMiddlewareEvent struct {
	ToolName              string                                `json:"tool_name"`
	Mode                  string                                `json:"mode"`
	SessionID             string                                `json:"session_id,omitempty"`
	RunID                 string                                `json:"run_id,omitempty"`
	ObservationOnly       bool                                  `json:"observation_only"`
	IsError               bool                                  `json:"is_error,omitempty"`
	ExternalContent       bool                                  `json:"external_content,omitempty"`
	UntrustedContent      bool                                  `json:"untrusted_content,omitempty"`
	Arguments             ToolResultArgumentsSummary            `json:"arguments,omitempty"`
	RawResultRef          string                                `json:"raw_result_ref,omitempty"`
	ArtifactRefs          []string                              `json:"artifact_refs,omitempty"`
	Content               ToolResultContentSummary              `json:"content"`
	OutputSchema          *ToolResultOutputSchema               `json:"output_schema,omitempty"`
	TerminalObservation   *ToolResultTerminalObservationSummary `json:"terminal_observation,omitempty"`
	Details               []string                              `json:"details,omitempty"`
	WouldTransform        bool                                  `json:"would_transform"`
	WouldTransformReasons []string                              `json:"would_transform_reasons,omitempty"`
	SuggestedStrategies   []string                              `json:"suggested_strategies,omitempty"`
}

type ToolResultArgumentsSummary struct {
	Bytes            int      `json:"bytes,omitempty"`
	Chars            int      `json:"chars,omitempty"`
	Fingerprint      string   `json:"fingerprint,omitempty"`
	JSONValid        bool     `json:"json_valid,omitempty"`
	JSONTopLevelKeys []string `json:"json_top_level_keys,omitempty"`
}

type ToolResultContentSummary struct {
	Bytes              int      `json:"bytes"`
	Chars              int      `json:"chars"`
	Lines              int      `json:"lines"`
	LooksBinary        bool     `json:"looks_binary,omitempty"`
	JSONValid          bool     `json:"json_valid,omitempty"`
	JSONTopLevelKeys   []string `json:"json_top_level_keys,omitempty"`
	SensitiveKeyHits   []string `json:"sensitive_key_hits,omitempty"`
	EmbeddedBinaryKeys []string `json:"embedded_binary_keys,omitempty"`
	OriginalBytes      int      `json:"original_bytes,omitempty"`
	KeptBytes          int      `json:"kept_bytes,omitempty"`
	Truncated          bool     `json:"truncated,omitempty"`
	GuardStrategy      string   `json:"guard_strategy,omitempty"`
}

type ToolResultOutputSchema struct {
	Present                bool     `json:"present"`
	Closed                 bool     `json:"closed"`
	PropertyCount          int      `json:"property_count,omitempty"`
	Required               []string `json:"required,omitempty"`
	MatchedTopLevelKeys    []string `json:"matched_top_level_keys,omitempty"`
	MissingRequired        []string `json:"missing_required,omitempty"`
	UnexpectedTopLevelKeys []string `json:"unexpected_top_level_keys,omitempty"`
	Drift                  bool     `json:"drift"`
}

type ToolResultTerminalObservationSummary struct {
	Present        bool   `json:"present"`
	Kind           string `json:"kind,omitempty"`
	Visibility     string `json:"visibility,omitempty"`
	Surface        string `json:"surface,omitempty"`
	Action         string `json:"action,omitempty"`
	DisplayCommand string `json:"display_command,omitempty"`
	UserVisible    bool   `json:"user_visible,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

func BuildToolResultMiddlewareEvent(input ToolResultMiddlewareInput) ToolResultMiddlewareEvent {
	toolName := NormalizeToolName(input.ToolName)
	externalContent, untrustedContent := classifyToolContent(input.ClassifyContent, toolName)
	largeBytes := input.LargeResultBytes
	if largeBytes <= 0 {
		largeBytes = defaultToolResultMiddlewareLargeBytes
	}
	largeLines := input.LargeResultLines
	if largeLines <= 0 {
		largeLines = defaultToolResultMiddlewareLargeLines
	}
	content := summarizeToolResultContent(input.Output, input)
	event := ToolResultMiddlewareEvent{
		ToolName:            toolName,
		Mode:                ToolResultMiddlewareModeObserveOnly,
		SessionID:           strings.TrimSpace(input.SessionID),
		RunID:               strings.TrimSpace(input.RunID),
		ObservationOnly:     true,
		IsError:             input.IsError,
		ExternalContent:     externalContent,
		UntrustedContent:    untrustedContent,
		Arguments:           summarizeToolResultArguments(input.Arguments),
		RawResultRef:        strings.TrimSpace(input.RawResultRef),
		ArtifactRefs:        normalizeToolResultRefs(input.ArtifactRefs),
		Content:             content,
		OutputSchema:        summarizeToolResultOutputSchema(input.OutputSchema, content),
		TerminalObservation: summarizeToolResultTerminalObservation(input.Output),
	}
	if input.IsError {
		event.Details = append(event.Details, "error_result_preserved")
		return event
	}

	reasons := []string{}
	strategies := []string{}
	if content.OriginalBytes > largeBytes || content.Bytes > largeBytes || content.Lines > largeLines {
		reasons = append(reasons, ToolResultMiddlewareReasonLargeResult)
		strategies = append(strategies, ToolResultMiddlewareStrategySummarizeWithRawRef)
	}
	if content.Truncated {
		reasons = append(reasons, ToolResultMiddlewareReasonRuntimeGuardTruncated)
		strategies = append(strategies, ToolResultMiddlewareStrategyPreserveRawArtifact)
	}
	if len(content.SensitiveKeyHits) > 0 {
		reasons = append(reasons, ToolResultMiddlewareReasonSensitiveKeyPresent)
		strategies = append(strategies, ToolResultMiddlewareStrategyRedactSensitiveKeys)
	}
	if len(content.EmbeddedBinaryKeys) > 0 {
		reasons = append(reasons, ToolResultMiddlewareReasonEmbeddedBinaryPayload)
		strategies = append(strategies, ToolResultMiddlewareStrategyExtractArtifactRef)
	}
	event.WouldTransformReasons = uniqueSortedStrings(reasons)
	event.SuggestedStrategies = uniqueSortedStrings(strategies)
	event.WouldTransform = len(event.WouldTransformReasons) > 0
	return event
}

func AppendToolResultMiddlewareTelemetryAttrs(attrs map[string]any, event ToolResultMiddlewareEvent) map[string]any {
	if attrs == nil {
		attrs = map[string]any{}
	}
	attrs["result_middleware_observed"] = true
	attrs["result_middleware_mode"] = strings.TrimSpace(event.Mode)
	attrs["result_middleware_observation_only"] = event.ObservationOnly
	attrs["result_middleware_would_transform"] = event.WouldTransform
	attrs["result_middleware_content_bytes"] = event.Content.Bytes
	attrs["result_middleware_content_chars"] = event.Content.Chars
	attrs["result_middleware_content_lines"] = event.Content.Lines
	if event.Content.OriginalBytes > 0 {
		attrs["result_middleware_original_bytes"] = event.Content.OriginalBytes
	}
	if event.Content.KeptBytes > 0 {
		attrs["result_middleware_kept_bytes"] = event.Content.KeptBytes
	}
	if event.Content.Truncated {
		attrs["result_middleware_truncated"] = true
	}
	if strategy := strings.TrimSpace(event.Content.GuardStrategy); strategy != "" {
		attrs["result_middleware_guard_strategy"] = strategy
	}
	if event.ExternalContent {
		attrs["result_middleware_external_content"] = true
	}
	if event.UntrustedContent {
		attrs["result_middleware_untrusted_content"] = true
	}
	if event.IsError {
		attrs["result_middleware_error_preserved"] = true
	}
	if len(event.WouldTransformReasons) > 0 {
		attrs["result_middleware_reasons"] = strings.Join(event.WouldTransformReasons, ",")
	}
	if len(event.SuggestedStrategies) > 0 {
		attrs["result_middleware_strategies"] = strings.Join(event.SuggestedStrategies, ",")
	}
	if len(event.Content.JSONTopLevelKeys) > 0 {
		attrs["result_middleware_json_keys"] = strings.Join(event.Content.JSONTopLevelKeys, ",")
	}
	if event.Content.JSONValid {
		attrs["result_middleware_json_valid"] = true
	}
	if len(event.Content.SensitiveKeyHits) > 0 {
		attrs["result_middleware_sensitive_keys"] = strings.Join(event.Content.SensitiveKeyHits, ",")
	}
	if len(event.Content.EmbeddedBinaryKeys) > 0 {
		attrs["result_middleware_embedded_binary_keys"] = strings.Join(event.Content.EmbeddedBinaryKeys, ",")
	}
	if event.OutputSchema != nil && event.OutputSchema.Present {
		attrs["result_middleware_output_schema_present"] = true
		attrs["result_middleware_output_schema_closed"] = event.OutputSchema.Closed
		attrs["result_middleware_output_schema_property_count"] = event.OutputSchema.PropertyCount
		attrs["result_middleware_output_schema_drift"] = event.OutputSchema.Drift
		if len(event.OutputSchema.Required) > 0 {
			attrs["result_middleware_output_schema_required"] = strings.Join(event.OutputSchema.Required, ",")
		}
		if len(event.OutputSchema.MatchedTopLevelKeys) > 0 {
			attrs["result_middleware_output_schema_matched_keys"] = strings.Join(event.OutputSchema.MatchedTopLevelKeys, ",")
		}
		if len(event.OutputSchema.MissingRequired) > 0 {
			attrs["result_middleware_output_schema_missing_required"] = strings.Join(event.OutputSchema.MissingRequired, ",")
		}
		if len(event.OutputSchema.UnexpectedTopLevelKeys) > 0 {
			attrs["result_middleware_output_schema_unexpected_keys"] = strings.Join(event.OutputSchema.UnexpectedTopLevelKeys, ",")
		}
	}
	if event.TerminalObservation != nil && event.TerminalObservation.Present {
		attrs["terminal_observation_present"] = true
		if value := strings.TrimSpace(event.TerminalObservation.Kind); value != "" {
			attrs["terminal_observation_kind"] = value
		}
		if value := strings.TrimSpace(event.TerminalObservation.Visibility); value != "" {
			attrs["terminal_observation_visibility"] = value
		}
		if value := strings.TrimSpace(event.TerminalObservation.Surface); value != "" {
			attrs["terminal_observation_surface"] = value
		}
		if value := strings.TrimSpace(event.TerminalObservation.Action); value != "" {
			attrs["terminal_observation_action"] = value
		}
		if value := strings.TrimSpace(event.TerminalObservation.DisplayCommand); value != "" {
			attrs["terminal_observation_display_command"] = value
		}
		if value := strings.TrimSpace(event.TerminalObservation.Reason); value != "" {
			attrs["terminal_observation_reason"] = value
		}
		attrs["terminal_observation_user_visible"] = event.TerminalObservation.UserVisible
	}
	if event.Arguments.Bytes > 0 {
		attrs["result_middleware_argument_bytes"] = event.Arguments.Bytes
	}
	if fingerprint := strings.TrimSpace(event.Arguments.Fingerprint); fingerprint != "" {
		attrs["result_middleware_argument_fingerprint"] = fingerprint
	}
	if len(event.ArtifactRefs) > 0 {
		attrs["result_middleware_artifact_refs"] = strings.Join(event.ArtifactRefs, ",")
	}
	if rawRef := strings.TrimSpace(event.RawResultRef); rawRef != "" {
		attrs["result_middleware_raw_result_ref"] = rawRef
	}
	if runID := strings.TrimSpace(event.RunID); runID != "" {
		attrs["result_middleware_run_id"] = runID
	}
	return attrs
}

func summarizeToolResultArguments(raw string) ToolResultArgumentsSummary {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ToolResultArgumentsSummary{}
	}
	summary := ToolResultArgumentsSummary{
		Bytes:       len(raw),
		Chars:       utf8.RuneCountInString(raw),
		Fingerprint: shortSHA256(raw),
	}
	if len(raw) <= toolResultMiddlewareMaxJSONInspect {
		keys, ok, _, _ := inspectToolResultJSON(raw)
		summary.JSONValid = ok
		summary.JSONTopLevelKeys = keys
	}
	return summary
}

func summarizeToolResultContent(raw string, input ToolResultMiddlewareInput) ToolResultContentSummary {
	summary := ToolResultContentSummary{
		Bytes:         len(raw),
		Chars:         utf8.RuneCountInString(raw),
		Lines:         countResultLines(raw),
		LooksBinary:   !utf8.ValidString(raw) || strings.ContainsRune(raw, '\x00'),
		OriginalBytes: input.OutputOriginalBytes,
		KeptBytes:     input.OutputKeptBytes,
		Truncated:     input.OutputTruncated,
		GuardStrategy: strings.TrimSpace(input.OutputGuardStrategy),
	}
	if summary.OriginalBytes <= 0 && raw != "" {
		summary.OriginalBytes = len(raw)
	}
	if summary.KeptBytes <= 0 && raw != "" {
		summary.KeptBytes = len(raw)
	}
	if len(raw) <= toolResultMiddlewareMaxJSONInspect {
		keys, ok, sensitiveKeys, binaryKeys := inspectToolResultJSON(raw)
		summary.JSONValid = ok
		summary.JSONTopLevelKeys = keys
		summary.SensitiveKeyHits = sensitiveKeys
		summary.EmbeddedBinaryKeys = binaryKeys
	}
	return summary
}

func summarizeToolResultOutputSchema(schema map[string]any, content ToolResultContentSummary) *ToolResultOutputSchema {
	if len(schema) == 0 {
		return nil
	}
	properties := toolResultSchemaProperties(schema["properties"])
	required := toolResultSchemaStringList(schema["required"])
	summary := &ToolResultOutputSchema{
		Present:       true,
		Closed:        schema["additionalProperties"] == false,
		PropertyCount: len(properties),
		Required:      required,
	}
	if content.JSONValid {
		keys := normalizeToolResultStringList(content.JSONTopLevelKeys)
		if len(properties) > 0 && len(keys) > 0 {
			matched := make([]string, 0)
			unexpected := make([]string, 0)
			for _, key := range keys {
				if properties[key] {
					matched = append(matched, key)
					continue
				}
				if summary.Closed {
					unexpected = append(unexpected, key)
				}
			}
			summary.MatchedTopLevelKeys = limitToolResultStrings(matched)
			summary.UnexpectedTopLevelKeys = limitToolResultStrings(unexpected)
		}
		if len(required) > 0 {
			presentKeys := map[string]bool{}
			for _, key := range keys {
				presentKeys[key] = true
			}
			missing := make([]string, 0)
			for _, key := range required {
				if !presentKeys[key] {
					missing = append(missing, key)
				}
			}
			summary.MissingRequired = limitToolResultStrings(missing)
		}
	}
	summary.Drift = len(summary.MissingRequired) > 0 || len(summary.UnexpectedTopLevelKeys) > 0
	return summary
}

func summarizeToolResultTerminalObservation(raw string) *ToolResultTerminalObservationSummary {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > toolResultMiddlewareMaxJSONInspect {
		return nil
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || payload == nil {
		return nil
	}
	observation, ok := payload["terminal_observation"].(map[string]any)
	if !ok || len(observation) == 0 {
		return nil
	}
	summary := &ToolResultTerminalObservationSummary{
		Present:        true,
		Kind:           toolResultString(observation["kind"]),
		Visibility:     toolResultString(observation["visibility"]),
		Surface:        toolResultString(observation["surface"]),
		Action:         toolResultString(observation["action"]),
		DisplayCommand: toolResultString(observation["display_command"]),
		Reason:         toolResultString(observation["reason"]),
	}
	if visible, ok := observation["user_visible"].(bool); ok {
		summary.UserVisible = visible
	}
	if summary.Kind == "" && summary.Visibility == "" && summary.Surface == "" {
		return nil
	}
	return summary
}

func toolResultSchemaProperties(raw any) map[string]bool {
	props, ok := raw.(map[string]any)
	if !ok || len(props) == 0 {
		return nil
	}
	out := map[string]bool{}
	for key := range props {
		key = strings.TrimSpace(key)
		if key != "" {
			out[key] = true
		}
	}
	return out
}

func toolResultString(raw any) string {
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func toolResultSchemaStringList(raw any) []string {
	switch values := raw.(type) {
	case []string:
		return normalizeToolResultStringList(values)
	case []any:
		out := make([]string, 0, len(values))
		for _, value := range values {
			if text, ok := value.(string); ok {
				out = append(out, text)
			}
		}
		return normalizeToolResultStringList(out)
	default:
		return nil
	}
}

func normalizeToolResultStringList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	return limitToolResultStrings(out)
}

func inspectToolResultJSON(raw string) ([]string, bool, []string, []string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, false, nil, nil
	}
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		return nil, false, nil, nil
	}
	var parsed any
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nil, false, nil, nil
	}
	keys := toolResultTopLevelKeys(parsed)
	sensitive := map[string]bool{}
	embeddedBinary := map[string]bool{}
	collectToolResultJSONSignals(parsed, "", sensitive, embeddedBinary)
	return keys, true, limitedSortedMapKeys(sensitive), limitedSortedMapKeys(embeddedBinary)
}

func toolResultTopLevelKeys(parsed any) []string {
	obj, ok := parsed.(map[string]any)
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(obj))
	for key := range obj {
		key = strings.TrimSpace(key)
		if key != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if len(keys) > toolResultMiddlewareMaxKeys {
		keys = keys[:toolResultMiddlewareMaxKeys]
	}
	return keys
}

func collectToolResultJSONSignals(value any, currentKey string, sensitive map[string]bool, embeddedBinary map[string]bool) {
	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			trimmedKey := strings.TrimSpace(key)
			if toolResultSensitiveKey(trimmedKey) {
				sensitive[trimmedKey] = true
			}
			collectToolResultJSONSignals(item, trimmedKey, sensitive, embeddedBinary)
		}
	case []any:
		for _, item := range typed {
			collectToolResultJSONSignals(item, currentKey, sensitive, embeddedBinary)
		}
	case string:
		if len(typed) >= 512 && toolResultEmbeddedBinaryKey(currentKey) {
			embeddedBinary[strings.TrimSpace(currentKey)] = true
		}
	}
}

func toolResultSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	compact := strings.NewReplacer("_", "", "-", "", " ", "").Replace(normalized)
	switch compact {
	case "password", "passwd", "secret", "apikey", "authorization", "cookie", "setcookie",
		"credential", "credentials", "privatekey", "accesstoken", "refreshtoken", "sessiontoken":
		return true
	default:
		return strings.HasSuffix(compact, "password") ||
			strings.HasSuffix(compact, "secret") ||
			strings.HasSuffix(compact, "apikey") ||
			strings.HasSuffix(compact, "token")
	}
}

func toolResultEmbeddedBinaryKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	compact := strings.NewReplacer("_", "", "-", "", " ", "").Replace(normalized)
	return strings.Contains(compact, "base64") ||
		strings.HasSuffix(compact, "bytes") ||
		strings.HasSuffix(compact, "binary")
}

func classifyToolContent(classifier ToolContentClassifier, toolName string) (bool, bool) {
	if classifier != nil {
		return classifier(toolName)
	}
	external := defaultToolResultExternalContent(toolName)
	return external, external || defaultToolResultUntrustedContent(toolName)
}

func defaultToolResultExternalContent(toolName string) bool {
	normalized := NormalizeToolName(toolName)
	if normalized == "browser" || strings.HasPrefix(normalized, "browser_") ||
		normalized == "pdf" || strings.HasPrefix(normalized, "pdf_") {
		return true
	}
	switch normalized {
	case "search", "open_page", "find_in_page", "web_fetch", "web_search", "http_request", "weather_lookup",
		"image_analyze", "video_frames":
		return true
	default:
		return false
	}
}

func defaultToolResultUntrustedContent(toolName string) bool {
	normalized := NormalizeToolName(toolName)
	switch normalized {
	case "exec", "process", "gateway", "nodes", "extensions", "message", "llm_task":
		return true
	default:
		return false
	}
}

func normalizeToolResultRefs(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	set := map[string]bool{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			set[trimmed] = true
		}
	}
	return limitedSortedMapKeys(set)
}

func countResultLines(raw string) int {
	if raw == "" {
		return 0
	}
	return strings.Count(raw, "\n") + 1
}

func uniqueSortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	set := map[string]bool{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			set[trimmed] = true
		}
	}
	return limitedSortedMapKeys(set)
}

func limitToolResultStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := uniqueSortedStrings(values)
	if len(out) == 0 {
		return nil
	}
	return out
}

func limitedSortedMapKeys(values map[string]bool) []string {
	if len(values) == 0 {
		return nil
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		key = strings.TrimSpace(key)
		if key != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if len(keys) > toolResultMiddlewareMaxKeys {
		keys = keys[:toolResultMiddlewareMaxKeys]
	}
	return keys
}

func shortSHA256(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return "sha256:" + hex.EncodeToString(sum[:])[:16]
}
