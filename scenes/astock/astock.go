package astock

import (
	astockcontracts "github.com/wsnacj/agentx-go/scenes/astock/contracts"
	"github.com/wsnacj/agentx-go/scenes/astock/internal/packresearch"
	"github.com/wsnacj/agentx-go/scenes/astock/internal/packsignal"
	"github.com/wsnacj/agentx-go/scenes/astock/internal/packvaluation"
	"github.com/wsnacj/agentx-go/extensions/domainmodule"
	"github.com/wsnacj/agentx-go/extensions/pack"
	"github.com/wsnacj/agentx-go/runtime/workflow"
)

const (
	ModuleID             = astockcontracts.ModuleID
	AStockDataPluginName = astockcontracts.AStockDataPluginName
	SkillAStockData      = astockcontracts.SkillAStockData

	ToolAStockInvestigation      = astockcontracts.ToolAStockInvestigation
	ToolAStockQuoteLookup        = astockcontracts.ToolAStockQuoteLookup
	ToolAStockResearchLookup     = astockcontracts.ToolAStockResearchLookup
	ToolAStockSignalLookup       = astockcontracts.ToolAStockSignalLookup
	ToolAStockAnnouncementLookup = astockcontracts.ToolAStockAnnouncementLookup
	ToolAStockProfileLookup      = astockcontracts.ToolAStockProfileLookup
	ToolAStockAnswerFormat       = astockcontracts.ToolAStockAnswerFormat

	ResearchPackID           = packresearch.PackID
	ResearchCaseType         = packresearch.CaseTypeResearch
	ResearchDefaultWorkflow  = packresearch.DefaultWorkflow
	SignalPackID             = packsignal.PackID
	SignalCaseType           = packsignal.CaseTypeSignal
	SignalDefaultWorkflow    = packsignal.DefaultWorkflow
	ValuationPackID          = packvaluation.PackID
	QuoteCaseType            = packvaluation.CaseTypeQuote
	ValuationCaseType        = packvaluation.CaseTypeValuation
	ValuationDefaultWorkflow = packvaluation.DefaultWorkflow
)

type ResearchEvaluationInput = astockcontracts.ResearchEvaluationInput
type ResearchEvaluation = astockcontracts.ResearchEvaluation
type SignalEvaluationInput = astockcontracts.SignalEvaluationInput
type SignalEvaluation = astockcontracts.SignalEvaluation
type ValuationEvaluationInput = astockcontracts.ValuationEvaluationInput
type ValuationEvaluation = astockcontracts.ValuationEvaluation

// Manifest returns a fresh portable description of the A-share extension.
// It does not register a Runner, install executors or authorize network access.
func Manifest() domainmodule.Manifest {
	return domainmodule.Manifest{
		ID:          ModuleID,
		Name:        "AgentX A-Stock",
		Description: "A-share quote, valuation, research, announcement, profile, and signal workflows.",
		Skills:      SkillNames(),
		Tools:       ToolNames(),
		Packs:       []string{ValuationPackID, ResearchPackID, SignalPackID},
		Workflows:   []string{ValuationDefaultWorkflow, ResearchDefaultWorkflow, SignalDefaultWorkflow},
	}
}

func ToolNames() []string  { return astockcontracts.ToolNames() }
func SkillNames() []string { return astockcontracts.SkillNames() }

func ValuationDefinition() pack.Definition { return packvaluation.Definition() }
func ResearchDefinition() pack.Definition  { return packresearch.Definition() }
func SignalDefinition() pack.Definition    { return packsignal.Definition() }

// Definitions returns fresh, caller-owned Pack definitions in registration order.
func Definitions() []pack.Definition {
	return []pack.Definition{ValuationDefinition(), ResearchDefinition(), SignalDefinition()}
}

// PackRegistrar is the narrow capability required by RegisterPacks.
type PackRegistrar interface {
	Register(pack.Definition) error
}

// RegisterPacks installs all portable A-share Pack definitions in stable order.
func RegisterPacks(registrar PackRegistrar) error {
	if registrar == nil {
		return nil
	}
	for _, definition := range Definitions() {
		if err := registrar.Register(definition); err != nil {
			return err
		}
	}
	return nil
}

func MaterializeValuationWorkflow(coordinator *pack.Coordinator) (workflow.Spec, error) {
	return packvaluation.MaterializedDefaultWorkflow(coordinator)
}

func MaterializeResearchWorkflow(coordinator *pack.Coordinator) (workflow.Spec, error) {
	return packresearch.MaterializedDefaultWorkflow(coordinator)
}

func MaterializeSignalWorkflow(coordinator *pack.Coordinator) (workflow.Spec, error) {
	return packsignal.MaterializedDefaultWorkflow(coordinator)
}

func EvaluateResearchEvidence(input ResearchEvaluationInput) ResearchEvaluation {
	return packresearch.EvaluateResearchEvidence(input)
}

func EvaluateSignalEvidence(input SignalEvaluationInput) SignalEvaluation {
	return packsignal.EvaluateSignalEvidence(input)
}

func EvaluateValuationEvidence(input ValuationEvaluationInput) ValuationEvaluation {
	return packvaluation.EvaluateValuationEvidence(input)
}
