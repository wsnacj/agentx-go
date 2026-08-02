package contracts

const (
	// ModuleID is the portable A-share domain-module identity.
	ModuleID = "agentx_a_stock"

	AStockDataPluginName = "a-stock-data"
	SkillAStockData      = "a-stock-data"

	ToolAStockInvestigation      = "a_stock_investigation"
	ToolAStockQuoteLookup        = "a_stock_quote_lookup"
	ToolAStockResearchLookup     = "a_stock_research_lookup"
	ToolAStockSignalLookup       = "a_stock_signal_lookup"
	ToolAStockAnnouncementLookup = "a_stock_announcement_lookup"
	ToolAStockProfileLookup      = "a_stock_profile_lookup"
	ToolAStockAnswerFormat       = "a_stock_answer_format"
)

// ToolNames returns a fresh copy of the portable A-share tool identities.
func ToolNames() []string {
	return []string{
		ToolAStockInvestigation,
		ToolAStockQuoteLookup,
		ToolAStockResearchLookup,
		ToolAStockSignalLookup,
		ToolAStockAnnouncementLookup,
		ToolAStockProfileLookup,
		ToolAStockAnswerFormat,
	}
}

// SkillNames returns a fresh copy of the portable A-share skill identities.
func SkillNames() []string { return []string{SkillAStockData} }
