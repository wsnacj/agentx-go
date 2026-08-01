package productshell

import (
	"context"

	agentxcases "github.com/wsnacj/agentx-go/runtime/cases"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"

	"github.com/wsnacj/agentx-go/extensions/pack"
)

// PreparationRuntimeFuncs adapts functions to PreparationRuntime. Nil host
// hooks are side-effect-free and do not invent product policy.
type PreparationRuntimeFuncs struct {
	ApplyInputCaseFn              func(Input) Input
	ResolveShellBindingFn         func(context.Context, string, Input) (*PreparedShellBinding, error)
	ApplyShellBindingFn           func(Input, *PreparedShellBinding) (Input, error)
	ResolveCommandDispatchFn      func(context.Context, Input) (*PreparedCommandDispatch, error)
	ApplyCommandDispatchFn        func(Input, *PreparedCommandDispatch) Input
	ParseRequestedSkillsFn        func(Input) ([]string, string)
	ShouldAttemptPackSelectionFn  func(Input) bool
	ResolvePackSelectionFn        func(context.Context, Input, string) (*PreparedPackSelection, error)
	ApplyPackSelectionFn          func(Input, *PreparedPackSelection) Input
	ShouldAttemptCaseBindingFn    func(Input) bool
	ResolveCandidateCaseBindingFn func(Input, *PreparedPackSelection) (pack.Binding, bool, error)
	ResolveCaseBindingDraftFn     func(context.Context, Input, string, pack.Binding, int) (*PreparedCaseBinding, error)
	MergeCaseBindingMetricsFn     func(CaseBindingMetrics, *PreparedCaseBinding) CaseBindingMetrics
	ApplyCaseBindingFn            func(Input, *PreparedCaseBinding) Input
	ResolveWorkflowFn             func(Input) (ResolvedWorkflow, error)
	ResolveEffectiveCaseFn        func(string, Input, string, *agentxworkflow.Spec, *pack.Binding) (*agentxcases.Case, error)
	ValidateEffectiveCaseFn       func(*pack.Binding, *agentxcases.Case) error
	ApplyEffectiveCaseFn          func(Input, agentxcases.Case) Input
	FinalizeShellBindingMetricsFn func(ShellBindingMetrics, Input, *pack.Binding, *PreparedShellBinding) ShellBindingMetrics
	PackSelectionMetricsFromFn    func(*PreparedPackSelection) PackSelectionMetrics
}

func (rt PreparationRuntimeFuncs) ApplyInputCase(input Input) Input {
	if rt.ApplyInputCaseFn != nil {
		return rt.ApplyInputCaseFn(input)
	}
	return ApplyInputCase(input)
}

func (rt PreparationRuntimeFuncs) ResolveShellBinding(ctx context.Context, sessionID string, input Input) (*PreparedShellBinding, error) {
	if rt.ResolveShellBindingFn == nil {
		return nil, nil
	}
	return rt.ResolveShellBindingFn(ctx, sessionID, input)
}

func (rt PreparationRuntimeFuncs) ApplyShellBinding(input Input, prepared *PreparedShellBinding) (Input, error) {
	if rt.ApplyShellBindingFn != nil {
		return rt.ApplyShellBindingFn(input, prepared)
	}
	return ApplyShellBindingToInput(input, prepared)
}

func (rt PreparationRuntimeFuncs) ResolveCommandDispatch(ctx context.Context, input Input) (*PreparedCommandDispatch, error) {
	if rt.ResolveCommandDispatchFn == nil {
		return nil, nil
	}
	return rt.ResolveCommandDispatchFn(ctx, input)
}

func (rt PreparationRuntimeFuncs) ApplyCommandDispatch(input Input, prepared *PreparedCommandDispatch) Input {
	if rt.ApplyCommandDispatchFn != nil {
		return rt.ApplyCommandDispatchFn(input, prepared)
	}
	return ApplyCommandDispatch(input, prepared)
}

func (rt PreparationRuntimeFuncs) ParseRequestedSkills(input Input) ([]string, string) {
	if rt.ParseRequestedSkillsFn != nil {
		return rt.ParseRequestedSkillsFn(input)
	}
	return ParseRequestedSkills(input)
}

func (rt PreparationRuntimeFuncs) ShouldAttemptPackSelection(input Input) bool {
	return rt.ShouldAttemptPackSelectionFn != nil && rt.ShouldAttemptPackSelectionFn(input)
}

func (rt PreparationRuntimeFuncs) ResolvePackSelection(ctx context.Context, input Input, message string) (*PreparedPackSelection, error) {
	if rt.ResolvePackSelectionFn == nil {
		return nil, nil
	}
	return rt.ResolvePackSelectionFn(ctx, input, message)
}

func (rt PreparationRuntimeFuncs) ApplyPackSelection(input Input, prepared *PreparedPackSelection) Input {
	if rt.ApplyPackSelectionFn != nil {
		return rt.ApplyPackSelectionFn(input, prepared)
	}
	return ApplyPackSelection(input, prepared)
}

func (rt PreparationRuntimeFuncs) ShouldAttemptCaseBinding(input Input) bool {
	return rt.ShouldAttemptCaseBindingFn != nil && rt.ShouldAttemptCaseBindingFn(input)
}

func (rt PreparationRuntimeFuncs) ResolveCandidateCaseBinding(input Input, selection *PreparedPackSelection) (pack.Binding, bool, error) {
	if rt.ResolveCandidateCaseBindingFn == nil {
		return pack.Binding{}, false, nil
	}
	return rt.ResolveCandidateCaseBindingFn(input, selection)
}

func (rt PreparationRuntimeFuncs) ResolveCaseBindingDraft(ctx context.Context, input Input, message string, binding pack.Binding, timeoutMs int) (*PreparedCaseBinding, error) {
	if rt.ResolveCaseBindingDraftFn == nil {
		return nil, nil
	}
	return rt.ResolveCaseBindingDraftFn(ctx, input, message, binding, timeoutMs)
}

func (rt PreparationRuntimeFuncs) MergeCaseBindingMetrics(metrics CaseBindingMetrics, prepared *PreparedCaseBinding) CaseBindingMetrics {
	if rt.MergeCaseBindingMetricsFn != nil {
		return rt.MergeCaseBindingMetricsFn(metrics, prepared)
	}
	return metrics
}

func (rt PreparationRuntimeFuncs) ApplyCaseBinding(input Input, prepared *PreparedCaseBinding) Input {
	if rt.ApplyCaseBindingFn != nil {
		return rt.ApplyCaseBindingFn(input, prepared)
	}
	return ApplyCaseBinding(input, prepared)
}

func (rt PreparationRuntimeFuncs) ResolveWorkflow(input Input) (ResolvedWorkflow, error) {
	if rt.ResolveWorkflowFn == nil {
		return ResolvedWorkflow{}, nil
	}
	return rt.ResolveWorkflowFn(input)
}

func (rt PreparationRuntimeFuncs) ResolveEffectiveCase(sessionID string, input Input, message string, spec *agentxworkflow.Spec, binding *pack.Binding) (*agentxcases.Case, error) {
	if rt.ResolveEffectiveCaseFn != nil {
		return rt.ResolveEffectiveCaseFn(sessionID, input, message, spec, binding)
	}
	return agentxcases.Clone(input.Case), nil
}

func (rt PreparationRuntimeFuncs) ValidateEffectiveCase(binding *pack.Binding, value *agentxcases.Case) error {
	if rt.ValidateEffectiveCaseFn == nil {
		return nil
	}
	return rt.ValidateEffectiveCaseFn(binding, value)
}

func (rt PreparationRuntimeFuncs) ApplyEffectiveCase(input Input, value agentxcases.Case) Input {
	if rt.ApplyEffectiveCaseFn != nil {
		return rt.ApplyEffectiveCaseFn(input, value)
	}
	return ApplyEffectiveCase(input, value)
}

func (rt PreparationRuntimeFuncs) FinalizeShellBindingMetrics(metrics ShellBindingMetrics, input Input, binding *pack.Binding, resolved *PreparedShellBinding) ShellBindingMetrics {
	if rt.FinalizeShellBindingMetricsFn != nil {
		return rt.FinalizeShellBindingMetricsFn(metrics, input, binding, resolved)
	}
	return FinalizeShellBindingMetrics(metrics, input, binding, resolved)
}

func (rt PreparationRuntimeFuncs) PackSelectionMetricsFromPrepared(prepared *PreparedPackSelection) PackSelectionMetrics {
	if rt.PackSelectionMetricsFromFn != nil {
		return rt.PackSelectionMetricsFromFn(prepared)
	}
	return PackSelectionMetricsFromPrepared(prepared)
}

var _ PreparationRuntime = PreparationRuntimeFuncs{}
