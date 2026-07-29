package llm

import "context"

const (
	legacyTemperatureKey    = "temperature"
	legacyTopPKey           = "top_p"
	legacyTopKKey           = "top_k"
	legacyMaxTokensKey      = "max_tokens"
	legacyResponseFormatKey = "response_format"
	legacyDetailKey         = "detail"
	legacyThinkingKey       = "thinking"
	legacyEnableThinkingKey = "enable_thinking"
	legacyExtraBodyKey      = "extra_body"
	legacyMaxRetryDelayKey  = "max_retry_delay_ms"
	legacyParallelToolCalls = "parallel_tool_calls"
	internalOnPayloadKey    = "__llmx_on_payload"
	internalOnResponseKey   = "__llmx_on_response"
)

// ThinkingOptions captures typed thinking controls for provider adapters.
type ThinkingOptions struct {
	Enabled bool
	Mode    string
}

// ReasoningOptions captures typed reasoning controls for provider adapters.
type ReasoningOptions struct {
	Effort string
}

// PayloadHook can inspect or replace a provider-bound payload before marshal/send.
type PayloadHook func(ctx context.Context, payload any) (any, error)

// ResponseMetadata captures transport-level response details for typed hooks.
type ResponseMetadata struct {
	Method     string
	URL        string
	StatusCode int
	Headers    map[string][]string
}

// ResponseHook observes provider responses before body decoding/stream consumption.
type ResponseHook func(ctx context.Context, meta ResponseMetadata) error

// RequestOptions provides a typed façade over legacy option maps.
// Unknown provider-specific keys are preserved in ProviderFields.
type RequestOptions struct {
	Temperature       *float64
	TopP              *float64
	TopK              *int
	MaxTokens         *int
	ResponseFormat    any
	Detail            string
	Thinking          *ThinkingOptions
	Reasoning         *ReasoningOptions
	Transport         string
	SessionID         string
	ParallelToolCalls *bool
	CacheControl      any
	MaxRetryDelayMs   *int
	OnPayload         PayloadHook
	OnResponse        ResponseHook
	Headers           map[string]string
	Metadata          map[string]any
	ExtraBody         map[string]any
	ProviderFields    map[string]any
}

// Clone returns a defensive copy of the request options.
func (o RequestOptions) Clone() RequestOptions {
	cloned := o
	if o.Temperature != nil {
		temperature := *o.Temperature
		cloned.Temperature = &temperature
	}
	if o.TopP != nil {
		topP := *o.TopP
		cloned.TopP = &topP
	}
	if o.TopK != nil {
		topK := *o.TopK
		cloned.TopK = &topK
	}
	if o.MaxTokens != nil {
		maxTokens := *o.MaxTokens
		cloned.MaxTokens = &maxTokens
	}
	if o.Thinking != nil {
		thinking := *o.Thinking
		cloned.Thinking = &thinking
	}
	if o.Reasoning != nil {
		reasoning := *o.Reasoning
		cloned.Reasoning = &reasoning
	}
	if o.MaxRetryDelayMs != nil {
		delay := *o.MaxRetryDelayMs
		cloned.MaxRetryDelayMs = &delay
	}
	if o.ParallelToolCalls != nil {
		parallel := *o.ParallelToolCalls
		cloned.ParallelToolCalls = &parallel
	}
	cloned.OnPayload = o.OnPayload
	cloned.OnResponse = o.OnResponse
	cloned.Headers = cloneStringMap(o.Headers)
	cloned.Metadata = cloneAnyMap(o.Metadata)
	cloned.CacheControl = cloneAnyValue(o.CacheControl)
	cloned.ExtraBody = cloneAnyMap(o.ExtraBody)
	cloned.ProviderFields = cloneAnyMap(o.ProviderFields)
	return cloned
}

// ToMap converts typed request options back into the legacy map form.
func (o RequestOptions) ToMap() map[string]any {
	out := cloneAnyMap(o.ProviderFields)
	if out == nil {
		out = map[string]any{}
	}
	if o.Temperature != nil {
		out[legacyTemperatureKey] = *o.Temperature
	}
	if o.TopP != nil {
		out[legacyTopPKey] = *o.TopP
	}
	if o.TopK != nil {
		out[legacyTopKKey] = *o.TopK
	}
	if o.MaxTokens != nil {
		out[legacyMaxTokensKey] = *o.MaxTokens
	}
	if o.ResponseFormat != nil {
		out[legacyResponseFormatKey] = o.ResponseFormat
	}
	if o.Detail != "" {
		out[legacyDetailKey] = o.Detail
	}
	if o.Thinking != nil {
		out[legacyThinkingKey] = o.Thinking.Enabled
		out[legacyEnableThinkingKey] = o.Thinking.Enabled
		if o.Thinking.Mode != "" {
			out["thinking_mode"] = o.Thinking.Mode
		}
	}
	if o.Reasoning != nil && o.Reasoning.Effort != "" {
		out["reasoning_effort"] = o.Reasoning.Effort
	}
	if o.Transport != "" {
		out["transport"] = o.Transport
	}
	if o.SessionID != "" {
		out["session_id"] = o.SessionID
	}
	if o.ParallelToolCalls != nil {
		out[legacyParallelToolCalls] = *o.ParallelToolCalls
	}
	if o.MaxRetryDelayMs != nil {
		out[legacyMaxRetryDelayKey] = *o.MaxRetryDelayMs
	}
	if o.OnPayload != nil {
		out[internalOnPayloadKey] = o.OnPayload
	}
	if o.OnResponse != nil {
		out[internalOnResponseKey] = o.OnResponse
	}
	if len(o.Headers) > 0 {
		out["headers"] = cloneStringMap(o.Headers)
	}
	if len(o.Metadata) > 0 {
		out["metadata"] = cloneAnyMap(o.Metadata)
	}
	extraBody := cloneAnyMap(o.ExtraBody)
	if o.CacheControl != nil {
		if extraBody == nil {
			extraBody = map[string]any{}
		}
		extraBody["cache_control"] = cloneAnyValue(o.CacheControl)
	}
	if len(extraBody) > 0 {
		out[legacyExtraBodyKey] = extraBody
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// RequestOptionsFromMap upgrades a legacy options map into the typed request form.
func RequestOptionsFromMap(in map[string]any) RequestOptions {
	if len(in) == 0 {
		return RequestOptions{}
	}
	out := RequestOptions{
		ProviderFields: map[string]any{},
	}
	for key, value := range in {
		switch key {
		case legacyTemperatureKey:
			if parsed, ok := parseFloat64(value); ok {
				out.Temperature = &parsed
				continue
			}
		case legacyTopPKey:
			if parsed, ok := parseFloat64(value); ok {
				out.TopP = &parsed
				continue
			}
		case legacyTopKKey:
			if parsed, ok := parseInt(value); ok {
				out.TopK = &parsed
				continue
			}
		case legacyMaxTokensKey:
			if parsed, ok := parseInt(value); ok {
				out.MaxTokens = &parsed
				continue
			}
		case legacyResponseFormatKey:
			out.ResponseFormat = value
			continue
		case legacyDetailKey:
			if parsed, ok := parseString(value); ok {
				out.Detail = parsed
				continue
			}
		case legacyThinkingKey, legacyEnableThinkingKey:
			if parsed, ok := parseBool(value); ok {
				if out.Thinking == nil {
					out.Thinking = &ThinkingOptions{}
				}
				out.Thinking.Enabled = parsed
				continue
			}
		case "thinking_mode":
			if parsed, ok := parseString(value); ok {
				if out.Thinking == nil {
					out.Thinking = &ThinkingOptions{}
				}
				out.Thinking.Mode = parsed
				continue
			}
		case "reasoning_effort":
			if parsed, ok := parseString(value); ok {
				out.Reasoning = &ReasoningOptions{Effort: parsed}
				continue
			}
		case "transport":
			if parsed, ok := parseString(value); ok {
				out.Transport = parsed
				continue
			}
		case "session_id":
			if parsed, ok := parseString(value); ok {
				out.SessionID = parsed
				continue
			}
		case legacyParallelToolCalls:
			if parsed, ok := parseBool(value); ok {
				out.ParallelToolCalls = &parsed
				continue
			}
		case legacyMaxRetryDelayKey:
			if parsed, ok := parseInt(value); ok {
				out.MaxRetryDelayMs = &parsed
				continue
			}
		case internalOnPayloadKey:
			if hook, ok := value.(PayloadHook); ok {
				out.OnPayload = hook
				continue
			}
		case internalOnResponseKey:
			if hook, ok := value.(ResponseHook); ok {
				out.OnResponse = hook
				continue
			}
		case "headers":
			if parsed, ok := parseStringMap(value); ok {
				out.Headers = parsed
				continue
			}
		case "metadata":
			if parsed, ok := parseAnyMap(value); ok {
				out.Metadata = parsed
				continue
			}
		case legacyExtraBodyKey:
			if parsed, ok := parseAnyMap(value); ok {
				if cacheControl, exists := parsed["cache_control"]; exists {
					out.CacheControl = cloneAnyValue(cacheControl)
					delete(parsed, "cache_control")
				}
				if len(parsed) > 0 {
					out.ExtraBody = parsed
				}
				continue
			}
		}
		out.ProviderFields[key] = value
	}
	if len(out.ProviderFields) == 0 {
		out.ProviderFields = nil
	}
	return out
}

// EmbeddingOptions provides a typed façade over embedding-specific options.
type EmbeddingOptions struct {
	Dimensions      *int
	Encoding        string
	Instructions    string
	SparseEmbedding *bool
	Path            string
	BatchSize       *int
	MaxRetryDelayMs *int
	OnPayload       PayloadHook
	OnResponse      ResponseHook
	ProviderFields  map[string]any
}

// Clone returns a defensive copy of embedding options.
func (o EmbeddingOptions) Clone() EmbeddingOptions {
	cloned := o
	if o.Dimensions != nil {
		dimensions := *o.Dimensions
		cloned.Dimensions = &dimensions
	}
	if o.SparseEmbedding != nil {
		sparseEmbedding := *o.SparseEmbedding
		cloned.SparseEmbedding = &sparseEmbedding
	}
	if o.BatchSize != nil {
		batchSize := *o.BatchSize
		cloned.BatchSize = &batchSize
	}
	if o.MaxRetryDelayMs != nil {
		delay := *o.MaxRetryDelayMs
		cloned.MaxRetryDelayMs = &delay
	}
	cloned.OnPayload = o.OnPayload
	cloned.OnResponse = o.OnResponse
	cloned.ProviderFields = cloneAnyMap(o.ProviderFields)
	return cloned
}

// ToMap converts typed embedding options back into the legacy map form.
func (o EmbeddingOptions) ToMap() map[string]any {
	out := cloneAnyMap(o.ProviderFields)
	if out == nil {
		out = map[string]any{}
	}
	if o.Dimensions != nil {
		out["dimensions"] = *o.Dimensions
	}
	if o.Encoding != "" {
		out["encoding"] = o.Encoding
	}
	if o.Instructions != "" {
		out["instructions"] = o.Instructions
	}
	if o.SparseEmbedding != nil {
		out["sparse_embedding"] = *o.SparseEmbedding
	}
	if o.MaxRetryDelayMs != nil {
		out[legacyMaxRetryDelayKey] = *o.MaxRetryDelayMs
	}
	if o.OnPayload != nil {
		out[internalOnPayloadKey] = o.OnPayload
	}
	if o.OnResponse != nil {
		out[internalOnResponseKey] = o.OnResponse
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// EmbeddingOptionsFromMap upgrades a legacy embedding option map into typed options.
func EmbeddingOptionsFromMap(in map[string]any) EmbeddingOptions {
	if len(in) == 0 {
		return EmbeddingOptions{}
	}
	out := EmbeddingOptions{
		ProviderFields: map[string]any{},
	}
	for key, value := range in {
		switch key {
		case "dimensions":
			if parsed, ok := parseInt(value); ok {
				out.Dimensions = &parsed
				continue
			}
		case "encoding", "encoding_format":
			if parsed, ok := parseString(value); ok {
				out.Encoding = parsed
				continue
			}
		case "instructions":
			if parsed, ok := parseString(value); ok {
				out.Instructions = parsed
				continue
			}
		case "sparse_embedding":
			if parsed, ok := parseBool(value); ok {
				out.SparseEmbedding = &parsed
				continue
			}
		case "path":
			if parsed, ok := parseString(value); ok {
				out.Path = parsed
				continue
			}
		case "batch_size":
			if parsed, ok := parseInt(value); ok {
				out.BatchSize = &parsed
				continue
			}
		case legacyMaxRetryDelayKey:
			if parsed, ok := parseInt(value); ok {
				out.MaxRetryDelayMs = &parsed
				continue
			}
		case internalOnPayloadKey:
			if hook, ok := value.(PayloadHook); ok {
				out.OnPayload = hook
				continue
			}
		case internalOnResponseKey:
			if hook, ok := value.(ResponseHook); ok {
				out.OnResponse = hook
				continue
			}
		}
		out.ProviderFields[key] = value
	}
	if len(out.ProviderFields) == 0 {
		out.ProviderFields = nil
	}
	return out
}

// SanitizeProviderOptionMap strips internal in-memory transport/runtime keys before
// provider payload compilation while preserving ordinary provider-specific fields.
func SanitizeProviderOptionMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := cloneAnyMap(in)
	delete(out, internalOnPayloadKey)
	delete(out, internalOnResponseKey)
	if len(out) == 0 {
		return nil
	}
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneAnyMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = cloneAnyValue(v)
	}
	return out
}

func cloneAnyValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return cloneAnyMap(v)
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = cloneAnyValue(item)
		}
		return out
	case []string:
		out := make([]string, len(v))
		copy(out, v)
		return out
	case map[string]string:
		return cloneStringMap(v)
	default:
		return value
	}
}

func parseFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func parseInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	default:
		return 0, false
	}
}

func parseBool(value any) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	default:
		return false, false
	}
}

func parseString(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	default:
		return "", false
	}
}

func parseAnyMap(value any) (map[string]any, bool) {
	switch v := value.(type) {
	case map[string]any:
		return cloneAnyMap(v), true
	case nil:
		return nil, true
	default:
		return nil, false
	}
}

func parseStringMap(value any) (map[string]string, bool) {
	switch v := value.(type) {
	case map[string]string:
		return cloneStringMap(v), true
	case map[string]any:
		out := make(map[string]string, len(v))
		for key, raw := range v {
			str, ok := raw.(string)
			if !ok {
				return nil, false
			}
			out[key] = str
		}
		return out, true
	case nil:
		return nil, true
	default:
		return nil, false
	}
}
