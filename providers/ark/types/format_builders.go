package types

// NewTextFormatText requests plain text output.
func NewTextFormatText() *TextFormat {
	return &TextFormat{Format: map[string]any{"type": "text"}}
}

// NewTextFormatJSON requests JSON object output.
func NewTextFormatJSON() *TextFormat {
	return &TextFormat{Format: map[string]any{"type": "json_object"}}
}

// NewTextFormatJSONSchema requests JSON schema output.
func NewTextFormatJSONSchema(schema any) *TextFormat {
	return &TextFormat{Format: map[string]any{"type": "json_schema", "json_schema": schema}}
}
