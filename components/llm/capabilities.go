package llm

// ModelCapabilities is the provider-neutral capability description for one
// configured text-generation adapter. A false field means callers must not
// assume the capability is available; it is not a global claim about every
// upstream model offered by that provider.
type ModelCapabilities struct {
	TextGeneration   bool
	ToolCalling      bool
	VisionInput      bool
	Streaming        bool
	LocalMediaInput  bool
	ReasoningControl bool
	ParallelTools    bool
	BotCompletion    bool
}
