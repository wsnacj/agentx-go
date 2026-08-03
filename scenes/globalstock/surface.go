package globalstock

import "context"

const (
	ToolGlobalStockInvestigation      = "global_stock_investigation"
	ToolGlobalStockQuoteLookup        = "global_stock_quote_lookup"
	ToolGlobalStockProfileLookup      = "global_stock_profile_lookup"
	ToolGlobalStockAnnouncementLookup = "global_stock_announcement_lookup"
	ToolGlobalStockResearchLookup     = "global_stock_research_lookup"
	ToolGlobalStockSignalLookup       = "global_stock_signal_lookup"
	ToolGlobalStockAnswerFormat       = "global_stock_answer_format"
	SkillGlobalStockData              = "global-stock-data"
)

// ToolPayloadHandler is the provider-neutral host callback used by the kit.
type ToolPayloadHandler func(context.Context, map[string]any) (any, error)

// ToolNames returns the runtime tool names required by the read-only kit.
func ToolNames() []string {
	return []string{
		ToolGlobalStockInvestigation,
		ToolGlobalStockQuoteLookup,
		ToolGlobalStockProfileLookup,
		ToolGlobalStockAnnouncementLookup,
		ToolGlobalStockResearchLookup,
		ToolGlobalStockSignalLookup,
		ToolGlobalStockAnswerFormat,
	}
}

// SkillNames returns the reusable skill identities exported by this kit.
func SkillNames() []string { return []string{SkillGlobalStockData} }
