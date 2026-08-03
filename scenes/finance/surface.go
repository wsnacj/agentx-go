package finance

import (
	"context"

	agentxpack "github.com/wsnacj/agentx-go/extensions/pack"
	financialreportbrief "github.com/wsnacj/agentx-go/scenes/finance/brief"
	financialreportmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
)

// ToolPayloadHandler is the provider-neutral host callback used by the kit.
type ToolPayloadHandler func(context.Context, map[string]any) (any, error)

// ToolNames returns the runtime tool identities required by the kit.
func ToolNames() []string {
	return []string{
		ToolFinanceReportLookup,
		ToolReportMetricsCandidates,
		ToolReportMetricsExtract,
		ToolReportMetricsGuard,
		ToolReportBriefExtract,
		ToolReportBriefGuard,
	}
}

// SkillNames returns the host-provided skill identities referenced by the kit.
func SkillNames() []string {
	return []string{
		SkillPublicReportMetrics,
		SkillPublicReportBrief,
		SkillOfficialReportDownload,
	}
}

// PackDefinitions returns the two portable financial-report Pack definitions.
func PackDefinitions() []agentxpack.Definition {
	return []agentxpack.Definition{
		financialreportmetrics.Definition(),
		financialreportbrief.Definition(),
	}
}

// RegisterPacksIntoRegistry installs both read-only Packs into reg.
func RegisterPacksIntoRegistry(reg agentxpack.Registry) error {
	if reg == nil {
		return nil
	}
	if err := financialreportmetrics.RegisterInto(reg); err != nil {
		return err
	}
	return financialreportbrief.RegisterInto(reg)
}
