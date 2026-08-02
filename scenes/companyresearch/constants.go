package companyresearch

const (
	ToolCompanyResearchLookup = "company_research_lookup"
	ToolCompanyCompareLookup  = "company_compare_lookup"
	ToolCompanyResearchGuard  = "company_research_guard"

	SkillCompanyResearch = "company-research"

	PackID     = "company-research-pack"
	WorkflowID = "company_research_lookup_v1"
)

func ToolNames() []string {
	return []string{
		ToolCompanyResearchLookup,
		ToolCompanyCompareLookup,
		ToolCompanyResearchGuard,
	}
}

func SkillNames() []string {
	return []string{SkillCompanyResearch}
}
