package types

// WebSearchOptions configures a web search tool.
type WebSearchOptions struct {
	Sources      []string
	Limit        int
	MaxKeyword   int
	UserLocation *UserLocation
}

// KnowledgeSearchOptions configures a knowledge search tool.
type KnowledgeSearchOptions struct {
	KnowledgeSearchID string
	Description       string
	Limit             int
	MaxKeyword        int
	UserLocation      *UserLocation
	DocFilters        map[string]any
	DenseWeight       *float64
	RankingOptions    *RankingOptions
}

// ImageProcessOptions configures image process tool toggles.
type ImageProcessOptions struct {
	Point     *bool
	Grounding *bool
	Zoom      *bool
	Rotate    *bool
}

// MCPOptions configures MCP tool settings.
type MCPOptions struct {
	ServerLabel     string
	ServerURL       string
	Headers         map[string]string
	RequireApproval any
	AllowedTools    any
}

// DoubaoAppOptions configures doubao_app tool settings.
type DoubaoAppOptions struct {
	Feature      *DoubaoAppFeature
	UserLocation *UserLocation
}

// NewFunctionTool builds a function tool definition.
func NewFunctionTool(name, description string, parameters map[string]any, strict *bool) Tool {
	return Tool{
		Type:        ToolTypeFunction,
		Name:        name,
		Description: description,
		Parameters:  parameters,
		Strict:      strict,
	}
}

// NewWebSearchTool builds a web search tool definition.
func NewWebSearchTool(opts WebSearchOptions) Tool {
	tool := Tool{
		Type:         ToolTypeWebSearch,
		Sources:      opts.Sources,
		UserLocation: opts.UserLocation,
	}
	if opts.Limit > 0 {
		tool.Limit = intPtr(opts.Limit)
	}
	if opts.MaxKeyword > 0 {
		tool.MaxKeyword = intPtr(opts.MaxKeyword)
	}
	return tool
}

// NewKnowledgeSearchTool builds a knowledge search tool definition.
func NewKnowledgeSearchTool(opts KnowledgeSearchOptions) Tool {
	tool := Tool{
		Type:              ToolTypeKnowledgeSearch,
		KnowledgeSearchID: opts.KnowledgeSearchID,
		Description:       opts.Description,
		UserLocation:      opts.UserLocation,
		DocFilters:        opts.DocFilters,
		DenseWeight:       opts.DenseWeight,
		RankingOptions:    opts.RankingOptions,
	}
	if opts.Limit > 0 {
		tool.Limit = intPtr(opts.Limit)
	}
	if opts.MaxKeyword > 0 {
		tool.MaxKeyword = intPtr(opts.MaxKeyword)
	}
	return tool
}

// NewImageProcessTool builds an image process tool definition.
func NewImageProcessTool(opts ImageProcessOptions) Tool {
	return Tool{
		Type:      ToolTypeImageProcess,
		Point:     toggleFromBool(opts.Point),
		Grounding: toggleFromBool(opts.Grounding),
		Zoom:      toggleFromBool(opts.Zoom),
		Rotate:    toggleFromBool(opts.Rotate),
	}
}

// NewMCPTool builds an MCP tool definition.
func NewMCPTool(opts MCPOptions) Tool {
	return Tool{
		Type:            ToolTypeMCP,
		ServerLabel:     opts.ServerLabel,
		ServerURL:       opts.ServerURL,
		Headers:         opts.Headers,
		RequireApproval: opts.RequireApproval,
		AllowedTools:    opts.AllowedTools,
	}
}

// NewDoubaoAppTool builds a doubao_app tool definition.
func NewDoubaoAppTool(opts DoubaoAppOptions) Tool {
	return Tool{
		Type:         ToolTypeDoubaoApp,
		Feature:      opts.Feature,
		UserLocation: opts.UserLocation,
	}
}

func toggleFromBool(v *bool) *ToggleConfig {
	if v == nil {
		return nil
	}
	if *v {
		return &ToggleConfig{Type: ToggleEnabled}
	}
	return &ToggleConfig{Type: ToggleDisabled}
}

func intPtr(v int) *int {
	return &v
}
