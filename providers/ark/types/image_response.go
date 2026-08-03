package types

// ImageGenerationResponse represents a non-stream image generation response.
type ImageGenerationResponse struct {
	Model        string                `json:"model,omitempty"`
	Created      int64                 `json:"created,omitempty"`
	Data         []ImageGenerationData `json:"data,omitempty"`
	Tools        []ImageGenerationTool `json:"tools,omitempty"`
	Usage        *ImageGenerationUsage `json:"usage,omitempty"`
	Error        *APIError             `json:"error,omitempty"`
	RequestModel string                `json:"-"`
}

// ImageGenerationData describes one generated image or one image-level failure.
type ImageGenerationData struct {
	URL     string    `json:"url,omitempty"`
	B64JSON string    `json:"b64_json,omitempty"`
	Size    string    `json:"size,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

// ImageGenerationUsage captures Images API usage fields.
type ImageGenerationUsage struct {
	GeneratedImages int            `json:"generated_images,omitempty"`
	OutputTokens    int            `json:"output_tokens,omitempty"`
	TotalTokens     int            `json:"total_tokens,omitempty"`
	ToolUsage       map[string]int `json:"tool_usage,omitempty"`
}
