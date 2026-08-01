package productshell

import "context"

// PreparationPipeline executes the canonical preparation stages in a fixed
// order while delegating product policy and side effects through
// PreparationRuntime.
type PreparationPipeline struct {
	runtime PreparationRuntime
}

// NewPreparationPipeline creates an experimental preparation pipeline. A nil
// runtime is accepted and produces a pass-through result.
func NewPreparationPipeline(runtime PreparationRuntime) *PreparationPipeline {
	return &PreparationPipeline{runtime: runtime}
}

// Prepare runs preparation for a session.
func (p *PreparationPipeline) Prepare(ctx context.Context, sessionID string, input Input) (PrepareResult, error) {
	return p.PrepareWithInput(ctx, PrepareInput{SessionID: sessionID, Input: input})
}

// PrepareWithInput runs the fixed preparation stage sequence. The first error
// is returned unchanged and no later stage is invoked.
func (p *PreparationPipeline) PrepareWithInput(ctx context.Context, request PrepareInput) (PrepareResult, error) {
	input := request.Input
	if p == nil || p.runtime == nil {
		return PrepareResult{Input: input, Workflow: &ResolvedWorkflow{}}, nil
	}
	rt := p.runtime
	input = rt.ApplyInputCase(input)
	preparedShellBinding, err := rt.ResolveShellBinding(ctx, request.SessionID, input)
	if err != nil {
		return PrepareResult{}, err
	}
	if preparedShellBinding != nil {
		input, err = rt.ApplyShellBinding(input, preparedShellBinding)
		if err != nil {
			return PrepareResult{}, err
		}
	}

	preparedCommandDispatch, err := rt.ResolveCommandDispatch(ctx, input)
	if err != nil {
		return PrepareResult{}, err
	}
	if preparedCommandDispatch != nil && preparedCommandDispatch.Matched {
		input = rt.ApplyCommandDispatch(input, preparedCommandDispatch)
	}

	requestedSkills, userMessage := rt.ParseRequestedSkills(input)
	input.UserMessage = userMessage
	input.ShellOptions.RequestedSkills = uniqueLowerOptionStrings(requestedSkills)
	input.Options = stripRequestedSkillOptions(input.Options)
	input.ShellOptions.SkillActivationPaths = ParseSkillActivationPaths(input)
	input.Options = stripSkillActivationPathOptions(input.Options)

	var preparedPackSelection *PreparedPackSelection
	if rt.ShouldAttemptPackSelection(input) {
		preparedPackSelection, err = rt.ResolvePackSelection(ctx, input, userMessage)
		if err != nil {
			return PrepareResult{}, err
		}
		if preparedPackSelection != nil && preparedPackSelection.Applied {
			input = rt.ApplyPackSelection(input, preparedPackSelection)
		}
	}

	caseBindingMetrics := CaseBindingMetrics{}
	if rt.ShouldAttemptCaseBinding(input) {
		candidateBinding, ok, bindErr := rt.ResolveCandidateCaseBinding(input, preparedPackSelection)
		if bindErr != nil {
			return PrepareResult{}, bindErr
		}
		if ok {
			preparedCaseBinding, bindErr := rt.ResolveCaseBindingDraft(
				ctx,
				input,
				userMessage,
				candidateBinding,
				request.LLMTaskTimeoutMs,
			)
			if bindErr != nil {
				return PrepareResult{}, bindErr
			}
			caseBindingMetrics = rt.MergeCaseBindingMetrics(caseBindingMetrics, preparedCaseBinding)
			if preparedCaseBinding != nil && preparedCaseBinding.Applied {
				input = rt.ApplyCaseBinding(input, preparedCaseBinding)
				if preparedPackSelection != nil && preparedPackSelection.Matched &&
					preparedPackSelection.Binding.PackID == candidateBinding.PackID &&
					preparedPackSelection.Binding.CaseType == candidateBinding.CaseType {
					preparedPackSelection.Applied = true
					preparedPackSelection.SkipReason = ""
				}
			}
		}
	}

	resolvedWorkflow, err := rt.ResolveWorkflow(input)
	if err != nil {
		return PrepareResult{}, err
	}
	if resolvedWorkflow.Spec != nil {
		input = ApplyResolvedWorkflow(input, *resolvedWorkflow.Spec)
	}
	effectiveCase, err := rt.ResolveEffectiveCase(
		request.SessionID,
		input,
		userMessage,
		resolvedWorkflow.Spec,
		resolvedWorkflow.PackBinding,
	)
	if err != nil {
		return PrepareResult{}, err
	}
	if err := rt.ValidateEffectiveCase(resolvedWorkflow.PackBinding, effectiveCase); err != nil {
		return PrepareResult{}, err
	}
	if effectiveCase != nil {
		input = rt.ApplyEffectiveCase(input, *effectiveCase)
	}

	shellBindingMetrics := ShellBindingMetrics{}
	if preparedShellBinding != nil {
		shellBindingMetrics = rt.FinalizeShellBindingMetrics(
			preparedShellBinding.Metrics,
			input,
			resolvedWorkflow.PackBinding,
			preparedShellBinding,
		)
	}

	return PrepareResult{
		Input:                  input,
		UserMessage:            userMessage,
		RequestedSkills:        requestedSkills,
		CommandDispatch:        preparedCommandDispatch,
		Workflow:               &resolvedWorkflow,
		EffectiveCase:          effectiveCase,
		ShellBinding:           preparedShellBinding,
		PackSelection:          preparedPackSelection,
		CommandDispatchMetrics: commandDispatchMetricsFromPrepared(preparedCommandDispatch),
		ShellBindingMetrics:    shellBindingMetrics,
		PackSelectionMetrics:   rt.PackSelectionMetricsFromPrepared(preparedPackSelection),
		CaseBindingMetrics:     caseBindingMetrics,
	}, nil
}

func commandDispatchMetricsFromPrepared(prepared *PreparedCommandDispatch) CommandDispatchMetrics {
	if prepared == nil {
		return CommandDispatchMetrics{}
	}
	return prepared.Metrics
}
