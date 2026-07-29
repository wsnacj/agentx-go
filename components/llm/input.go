package llm

// ChatInput is the typed public input for chat-like requests.
type ChatInput struct {
	ConfigName   string
	SystemPrompt string
	Messages     Conversation
	Request      RequestOptions
	Tools        []Tool
	ToolChoice   *ToolChoice
}

// Clone returns a defensive copy of the chat input.
func (in ChatInput) Clone() ChatInput {
	cloned := in
	if len(in.Messages) > 0 {
		cloned.Messages = append(Conversation(nil), in.Messages...)
	}
	if len(in.Tools) > 0 {
		cloned.Tools = append([]Tool(nil), in.Tools...)
	}
	if in.ToolChoice != nil {
		choice := *in.ToolChoice
		if in.ToolChoice.Function != nil {
			fn := *in.ToolChoice.Function
			choice.Function = &fn
		}
		cloned.ToolChoice = &choice
	}
	cloned.Request = in.Request.Clone()
	return cloned
}

// VisionInput is the typed public input for multimodal chat requests.
type VisionInput struct {
	ConfigName   string
	SystemPrompt string
	Messages     Conversation
	Visuals      []VisualContent
	Request      RequestOptions
	Tools        []Tool
	ToolChoice   *ToolChoice
}

// Clone returns a defensive copy of the vision input.
func (in VisionInput) Clone() VisionInput {
	cloned := in
	if len(in.Messages) > 0 {
		cloned.Messages = append(Conversation(nil), in.Messages...)
	}
	if len(in.Visuals) > 0 {
		cloned.Visuals = append([]VisualContent(nil), in.Visuals...)
	}
	if len(in.Tools) > 0 {
		cloned.Tools = append([]Tool(nil), in.Tools...)
	}
	if in.ToolChoice != nil {
		choice := *in.ToolChoice
		if in.ToolChoice.Function != nil {
			fn := *in.ToolChoice.Function
			choice.Function = &fn
		}
		cloned.ToolChoice = &choice
	}
	cloned.Request = in.Request.Clone()
	return cloned
}

// BotInput is the typed public input for bot-style requests.
type BotInput struct {
	ConfigName   string
	SystemPrompt string
	Messages     Conversation
	Request      RequestOptions
	Tools        []Tool
	ToolChoice   *ToolChoice
}

// Clone returns a defensive copy of the bot input.
func (in BotInput) Clone() BotInput {
	cloned := in
	if len(in.Messages) > 0 {
		cloned.Messages = append(Conversation(nil), in.Messages...)
	}
	if len(in.Tools) > 0 {
		cloned.Tools = append([]Tool(nil), in.Tools...)
	}
	if in.ToolChoice != nil {
		choice := *in.ToolChoice
		if in.ToolChoice.Function != nil {
			fn := *in.ToolChoice.Function
			choice.Function = &fn
		}
		cloned.ToolChoice = &choice
	}
	cloned.Request = in.Request.Clone()
	return cloned
}

// EmbedInput is the typed public input for embedding requests.
type EmbedInput struct {
	ConfigName string
	Texts      []string
	Images     []string
	Request    EmbeddingOptions
}

// Clone returns a defensive copy of the embedding input.
func (in EmbedInput) Clone() EmbedInput {
	cloned := in
	if len(in.Texts) > 0 {
		cloned.Texts = append([]string(nil), in.Texts...)
	}
	if len(in.Images) > 0 {
		cloned.Images = append([]string(nil), in.Images...)
	}
	cloned.Request = in.Request.Clone()
	return cloned
}
