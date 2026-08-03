package types

const (
	ThinkingEnabled  = "enabled"
	ThinkingDisabled = "disabled"
)

// NewThinkingConfig builds a thinking config.
func NewThinkingConfig(enabled bool) *ThinkingConfig {
	if enabled {
		return &ThinkingConfig{Type: ThinkingEnabled}
	}
	return &ThinkingConfig{Type: ThinkingDisabled}
}

// NewReasoningConfig builds a reasoning config with effort.
func NewReasoningConfig(effort string) *ReasoningConfig {
	if effort == "" {
		return nil
	}
	return &ReasoningConfig{Effort: effort}
}

// NewContextManagement builds a context management config.
func NewContextManagement(edits ...ContextEdit) *ContextManagement {
	if len(edits) == 0 {
		return nil
	}
	return &ContextManagement{Edits: edits}
}

// ClearThinkingKeepTurns keeps the latest N thinking turns.
func ClearThinkingKeepTurns(turns int) ContextEdit {
	return ContextEdit{
		Type: "clear_thinking",
		Keep: &ContextKeep{
			Type:  "thinking_turns",
			Value: turns,
		},
	}
}

// ClearThinkingKeepAll keeps all thinking turns.
func ClearThinkingKeepAll() ContextEdit {
	return ContextEdit{
		Type: "clear_thinking",
		Keep: &ContextKeep{
			Type: "all",
		},
	}
}

// ClearToolUses builds a clear_tool_uses edit.
func ClearToolUses(trigger int, keep int, exclude []string, clearInput *bool) ContextEdit {
	edit := ContextEdit{
		Type: "clear_tool_uses",
	}
	if keep > 0 {
		edit.Keep = &ContextKeep{Type: "tool_uses", Value: keep}
	}
	if trigger > 0 {
		edit.Trigger = &ContextTrigger{Type: "tool_uses", Value: trigger}
	}
	if len(exclude) > 0 {
		edit.ExcludeTools = exclude
	}
	if clearInput != nil {
		edit.ClearToolInput = clearInput
	}
	return edit
}
