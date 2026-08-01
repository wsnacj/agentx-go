package controlcontract

import (
	"context"
	"strings"
)

type ObjectiveFinalAnswerTraceStep struct {
	ContractVersion string             `json:"contract_version,omitempty"`
	StepRef         DisplaySafeRef     `json:"step_ref,omitempty"`
	NodeRef         DisplaySafeRef     `json:"node_ref,omitempty"`
	CapabilityRef   DisplaySafeRef     `json:"capability_ref,omitempty"`
	StrategyRef     DisplaySafeRef     `json:"strategy_ref,omitempty"`
	DescriptorRef   DisplaySafeRef     `json:"descriptor_ref,omitempty"`
	Status          VerificationStatus `json:"status,omitempty"`
	Action          string             `json:"action,omitempty"`
	EvidenceRefs    []EvidenceRef      `json:"evidence_refs,omitempty"`
	MissingInputs   []MissingInput     `json:"missing_inputs,omitempty"`
	Limitations     []string           `json:"limitations,omitempty"`
	FailureClass    FailureClass       `json:"failure_class,omitempty"`
	Boundaries      []Boundary         `json:"boundaries,omitempty"`
	NextHostAction  NextHostAction     `json:"next_host_action,omitempty"`
	RawOutputLoaded bool               `json:"raw_output_loaded"`
}

type ObjectiveFinalAnswerTrace struct {
	ContractVersion    string                          `json:"contract_version,omitempty"`
	Projected          bool                            `json:"projected"`
	TraceRef           DisplaySafeRef                  `json:"trace_ref,omitempty"`
	ObjectiveID        string                          `json:"objective_id,omitempty"`
	GoalSummary        string                          `json:"goal_summary,omitempty"`
	Status             VerificationStatus              `json:"status,omitempty"`
	Steps              []ObjectiveFinalAnswerTraceStep `json:"steps,omitempty"`
	UsedCapabilityRefs []DisplaySafeRef                `json:"used_capability_refs,omitempty"`
	EvidenceRefs       []EvidenceRef                   `json:"evidence_refs,omitempty"`
	Limitations        []string                        `json:"limitations,omitempty"`
	MissingInputs      []MissingInput                  `json:"missing_inputs,omitempty"`
	FailureClass       FailureClass                    `json:"failure_class,omitempty"`
	Boundaries         []Boundary                      `json:"boundaries,omitempty"`
	NextHostAction     NextHostAction                  `json:"next_host_action,omitempty"`
	RunnerEffect       string                          `json:"runner_effect,omitempty"`
	PromptEffect       string                          `json:"prompt_effect,omitempty"`
	RawOutputLoaded    bool                            `json:"raw_output_loaded"`
}

type ObjectiveFinalAnswerTraceInput struct {
	TraceRef           DisplaySafeRef                  `json:"trace_ref,omitempty"`
	Spec               ObjectiveSpec                   `json:"spec,omitempty"`
	Graph              ObjectiveGraph                  `json:"graph,omitempty"`
	Verification       ObjectiveVerificationGateResult `json:"verification,omitempty"`
	Replan             ObjectiveReplanProposal         `json:"replan,omitempty"`
	ReplanGraphPatch   ObjectiveReplanGraphPatch       `json:"replan_graph_patch,omitempty"`
	Status             VerificationStatus              `json:"status,omitempty"`
	Steps              []ObjectiveFinalAnswerTraceStep `json:"steps,omitempty"`
	UsedCapabilityRefs []DisplaySafeRef                `json:"used_capability_refs,omitempty"`
	EvidenceRefs       []EvidenceRef                   `json:"evidence_refs,omitempty"`
	Limitations        []string                        `json:"limitations,omitempty"`
	MissingInputs      []MissingInput                  `json:"missing_inputs,omitempty"`
	FailureClass       FailureClass                    `json:"failure_class,omitempty"`
	Boundaries         []Boundary                      `json:"boundaries,omitempty"`
	NextHostAction     NextHostAction                  `json:"next_host_action,omitempty"`
	RawOutputLoaded    bool                            `json:"raw_output_loaded"`
}

type ObjectiveFinalAnswerDraft struct {
	ContractVersion string             `json:"contract_version,omitempty"`
	DraftRef        DisplaySafeRef     `json:"draft_ref,omitempty"`
	Status          VerificationStatus `json:"status,omitempty"`
	AnswerText      string             `json:"answer_text,omitempty"`
	Conclusion      string             `json:"conclusion,omitempty"`
	Steps           []string           `json:"steps,omitempty"`
	Evidence        []string           `json:"evidence,omitempty"`
	EvidenceRefs    []EvidenceRef      `json:"evidence_refs,omitempty"`
	Limitations     []string           `json:"limitations,omitempty"`
	NextStep        string             `json:"next_step,omitempty"`
	Boundaries      []Boundary         `json:"boundaries,omitempty"`
	RawOutputLoaded bool               `json:"raw_output_loaded"`
}

type ObjectiveFinalAnswerSynthesisRequest struct {
	ContractVersion string                    `json:"contract_version,omitempty"`
	RequestRef      DisplaySafeRef            `json:"request_ref,omitempty"`
	Trace           ObjectiveFinalAnswerTrace `json:"trace,omitempty"`
	Boundaries      []Boundary                `json:"boundaries,omitempty"`
	RawOutputLoaded bool                      `json:"raw_output_loaded"`
}

type ObjectiveFinalAnswerSynthesisResponse struct {
	ContractVersion string                    `json:"contract_version,omitempty"`
	ResponseRef     DisplaySafeRef            `json:"response_ref,omitempty"`
	Draft           ObjectiveFinalAnswerDraft `json:"draft,omitempty"`
	Boundaries      []Boundary                `json:"boundaries,omitempty"`
	RawOutputLoaded bool                      `json:"raw_output_loaded"`
}

type ObjectiveFinalAnswerSynthesizer interface {
	SynthesizeObjectiveFinalAnswer(context.Context, ObjectiveFinalAnswerSynthesisRequest) (ObjectiveFinalAnswerSynthesisResponse, error)
}

type ObjectiveFinalAnswerSynthesizerFunc func(context.Context, ObjectiveFinalAnswerSynthesisRequest) (ObjectiveFinalAnswerSynthesisResponse, error)

func (fn ObjectiveFinalAnswerSynthesizerFunc) SynthesizeObjectiveFinalAnswer(ctx context.Context, request ObjectiveFinalAnswerSynthesisRequest) (ObjectiveFinalAnswerSynthesisResponse, error) {
	return fn(ctx, request)
}

type ObjectiveFinalAnswerInput struct {
	Trace               ObjectiveFinalAnswerTrace       `json:"trace,omitempty"`
	TraceInput          ObjectiveFinalAnswerTraceInput  `json:"trace_input,omitempty"`
	Draft               ObjectiveFinalAnswerDraft       `json:"draft,omitempty"`
	EnableSynthesizer   bool                            `json:"enable_synthesizer"`
	Synthesizer         ObjectiveFinalAnswerSynthesizer `json:"-"`
	SynthesisRequestRef DisplaySafeRef                  `json:"synthesis_request_ref,omitempty"`
	Boundaries          []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded     bool                            `json:"raw_output_loaded"`
}

type ObjectiveFinalAnswer struct {
	ContractVersion string                    `json:"contract_version,omitempty"`
	Projected       bool                      `json:"projected"`
	ReadyForUser    bool                      `json:"ready_for_user"`
	Status          VerificationStatus        `json:"status,omitempty"`
	Trace           ObjectiveFinalAnswerTrace `json:"trace,omitempty"`
	Draft           ObjectiveFinalAnswerDraft `json:"draft,omitempty"`
	SynthesisRef    DisplaySafeRef            `json:"synthesis_ref,omitempty"`
	FailureClass    FailureClass              `json:"failure_class,omitempty"`
	MissingInputs   []MissingInput            `json:"missing_inputs,omitempty"`
	Boundaries      []Boundary                `json:"boundaries,omitempty"`
	NextHostAction  NextHostAction            `json:"next_host_action,omitempty"`
	RunnerEffect    string                    `json:"runner_effect,omitempty"`
	PromptEffect    string                    `json:"prompt_effect,omitempty"`
	RawOutputLoaded bool                      `json:"raw_output_loaded"`
}

func BuildObjectiveFinalAnswerTrace(input ObjectiveFinalAnswerTraceInput) ObjectiveFinalAnswerTrace {
	rawInputUnsafe := objectiveFinalAnswerTraceInputUnsafe(input)
	spec := input.Spec.Normalize()
	graph := input.Graph.Normalize()
	verification := input.Verification.Normalize()
	replan := input.Replan.Normalize()
	replanGraphPatch := ObjectiveReplanGraphPatch{}
	if objectiveFinalAnswerReplanGraphPatchProvided(input.ReplanGraphPatch) {
		replanGraphPatch = input.ReplanGraphPatch.Normalize()
	}
	steps := normalizeObjectiveFinalAnswerTraceSteps(input.Steps)
	if len(steps) == 0 {
		steps = objectiveFinalAnswerTraceStepsFromRecoveryPath(graph, replan, replanGraphPatch)
	}
	result := ObjectiveFinalAnswerTrace{
		ContractVersion: defaultContractVersion(""),
		Projected:       true,
		TraceRef:        firstDisplaySafeRef(input.TraceRef, "trace:objective_final_answer"),
		ObjectiveID:     firstNonEmptyContractString(spec.ObjectiveID, graph.ObjectiveID, verification.Frame.ID, replan.ObjectiveID),
		GoalSummary:     objectiveFinalAnswerSafeText(spec.GoalSummary),
		Status:          firstObjectiveFinalAnswerStatus(input.Status, verification.Status, replan.Status, replanGraphPatch.Status, objectiveFinalAnswerTraceStepStatus(steps)),
		Steps:           steps,
		UsedCapabilityRefs: normalizeDisplaySafeRefs(append(
			append(cloneDisplaySafeRefs(input.UsedCapabilityRefs), objectiveFinalAnswerCapabilityRefsFromGraph(graph)...),
			objectiveFinalAnswerCapabilityRefsFromSteps(steps)...,
		)),
		EvidenceRefs: MergeEvidenceRefs(
			input.EvidenceRefs,
			verification.EvidenceRefs,
			replan.EvidenceRefs,
			replanGraphPatch.EvidenceRefs,
			objectiveFinalAnswerEvidenceRefsFromSteps(steps),
		),
		Limitations: normalizeStringList(append(
			append(cloneStringSlice(input.Limitations), objectiveFinalAnswerLimitationsFromSteps(steps)...),
			verification.FailureReason,
		)),
		MissingInputs: MergeMissingInputs(
			input.MissingInputs,
			spec.MissingInputs,
			graph.MissingInputs,
			verification.MissingInputs,
			replan.MissingInputs,
			replanGraphPatch.MissingInputs,
			objectiveFinalAnswerMissingInputsFromSteps(steps),
		),
		FailureClass: firstFailureClass(input.FailureClass, verification.FailureClass, objectiveFinalAnswerFailureFromReplan(replan), objectiveFinalAnswerFailureFromSteps(steps)),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_final_answer_trace",
				"trace_backed_final_answer",
				"final_answer_input_must_be_trace_or_evidence",
				"no_runner_dispatch",
				"no_runtime_adapter_execution",
				"no_transcript_fact_inference",
			},
			input.Boundaries,
			spec.Boundaries,
			graph.Boundaries,
			verification.Boundaries,
			replan.Boundaries,
			replanGraphPatch.Boundaries,
			objectiveFinalAnswerBoundariesFromSteps(steps),
		),
		NextHostAction:  firstNextHostAction(input.NextHostAction, firstNextHostAction(replan.NextHostAction, verification.NextHostAction)),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: rawInputUnsafe || input.RawOutputLoaded || spec.RawOutputLoaded || graph.RawOutputLoaded || verification.RawOutputLoaded || replan.RawOutputLoaded || replanGraphPatch.RawOutputLoaded || objectiveFinalAnswerStepsRawOutputLoaded(steps),
	}
	if result.Status == VerificationNotEvaluated {
		result.Status = VerificationBlocked
	}
	if result.RawOutputLoaded || objectiveFinalAnswerTraceUnsafe(result) {
		result.Status = VerificationReviewRequired
		result.FailureClass = firstFailureClass(result.FailureClass, FailureEvidenceWeak)
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:display_safe_refs")
		result.Boundaries = AppendBoundaries(result.Boundaries, "raw_output_not_allowed")
		result.NextHostAction = "provide_display_safe_refs"
	}
	return result.Normalize()
}

func BuildObjectiveFinalAnswer(ctx context.Context, input ObjectiveFinalAnswerInput) ObjectiveFinalAnswer {
	rawTraceUnsafe := objectiveFinalAnswerTraceUnsafe(input.Trace) || objectiveFinalAnswerTraceInputUnsafe(input.TraceInput)
	rawDraftUnsafe := objectiveFinalAnswerDraftUnsafe(input.Draft)
	trace := input.Trace.Normalize()
	if objectiveFinalAnswerTraceEmpty(trace) {
		trace = BuildObjectiveFinalAnswerTrace(input.TraceInput)
	}
	draft := input.Draft.Normalize()
	result := ObjectiveFinalAnswer{
		ContractVersion: defaultContractVersion(""),
		Projected:       true,
		Status:          firstObjectiveFinalAnswerStatus(trace.Status, draft.Status),
		Trace:           trace,
		Draft:           draft,
		FailureClass:    trace.FailureClass,
		MissingInputs:   cloneMissingInputs(trace.MissingInputs),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_final_answer",
				"trace_backed_synthesis_only",
				"answer_text_must_be_supported_by_trace",
				"display_safe_final_answer",
				"no_runner_dispatch",
				"no_runtime_adapter_execution",
			},
			input.Boundaries,
			trace.Boundaries,
			draft.Boundaries,
		),
		NextHostAction:  trace.NextHostAction,
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: rawTraceUnsafe || rawDraftUnsafe || input.RawOutputLoaded || trace.RawOutputLoaded || draft.RawOutputLoaded,
	}
	if input.EnableSynthesizer {
		if input.Synthesizer == nil {
			return objectiveFinalAnswerBlock(result, VerificationReviewRequired, FailureConfigMissing, "host:objective_final_answer_synthesizer", "provide_objective_final_answer_synthesizer", "objective_final_answer_synthesizer_missing")
		}
		if ctx == nil {
			ctx = context.Background()
		}
		request := ObjectiveFinalAnswerSynthesisRequest{
			ContractVersion: defaultContractVersion(""),
			RequestRef:      firstDisplaySafeRef(input.SynthesisRequestRef, "request:objective_final_answer_synthesis"),
			Trace:           trace,
			Boundaries: []Boundary{
				"llm_synthesis_phrase_only",
				"llm_must_not_introduce_unbacked_facts",
				"llm_input_trace_only",
			},
		}.Normalize()
		response, err := input.Synthesizer.SynthesizeObjectiveFinalAnswer(ctx, request)
		if err != nil {
			return objectiveFinalAnswerBlock(result, VerificationReviewRequired, FailureInternalError, "host:objective_final_answer_synthesizer", "provide_trace_backed_final_answer", "objective_final_answer_synthesizer_failed")
		}
		rawResponseUnsafe := objectiveFinalAnswerDraftUnsafe(response.Draft)
		response = response.Normalize()
		result.SynthesisRef = response.ResponseRef
		result.Boundaries = MergeBoundaries(result.Boundaries, response.Boundaries)
		result.RawOutputLoaded = result.RawOutputLoaded || rawResponseUnsafe || response.RawOutputLoaded
		if !objectiveFinalAnswerDraftEmpty(response.Draft) {
			result.Draft = response.Draft.Normalize()
		}
	}
	if objectiveFinalAnswerDraftEmpty(result.Draft) {
		result.Draft = objectiveFinalAnswerFallbackDraft(trace).Normalize()
	}
	result.Status = firstObjectiveFinalAnswerStatus(trace.Status, result.Draft.Status)
	result.RawOutputLoaded = result.RawOutputLoaded || result.Draft.RawOutputLoaded
	result.Boundaries = MergeBoundaries(result.Boundaries, result.Draft.Boundaries)
	if result.RawOutputLoaded || objectiveFinalAnswerTraceUnsafe(result.Trace) || objectiveFinalAnswerDraftUnsafe(result.Draft) {
		return objectiveFinalAnswerBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_final_answer_trace", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if missing, next, boundary, failure := objectiveFinalAnswerCompletenessIssue(result.Trace, result.Draft); missing != "" {
		return objectiveFinalAnswerBlock(result, VerificationReviewRequired, failure, missing, next, boundary)
	}
	result.ReadyForUser = true
	return result.Normalize()
}

func CloneObjectiveFinalAnswerTraceStep(in ObjectiveFinalAnswerTraceStep) ObjectiveFinalAnswerTraceStep {
	out := in
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Limitations = cloneStringSlice(in.Limitations)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (s ObjectiveFinalAnswerTraceStep) Clone() ObjectiveFinalAnswerTraceStep {
	return CloneObjectiveFinalAnswerTraceStep(s)
}

func (s ObjectiveFinalAnswerTraceStep) Normalize() ObjectiveFinalAnswerTraceStep {
	out := s.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.StepRef = normalizeOneDisplaySafeRef(out.StepRef)
	out.NodeRef = normalizeOneDisplaySafeRef(out.NodeRef)
	out.CapabilityRef = normalizeOneDisplaySafeRef(out.CapabilityRef)
	out.StrategyRef = normalizeOneDisplaySafeRef(out.StrategyRef)
	out.DescriptorRef = normalizeOneDisplaySafeRef(out.DescriptorRef)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Action = objectiveFinalAnswerSafeText(out.Action)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Limitations = normalizeStringList(out.Limitations)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func CloneObjectiveFinalAnswerTrace(in ObjectiveFinalAnswerTrace) ObjectiveFinalAnswerTrace {
	out := in
	out.Steps = cloneObjectiveFinalAnswerTraceSteps(in.Steps)
	out.UsedCapabilityRefs = cloneDisplaySafeRefs(in.UsedCapabilityRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Limitations = cloneStringSlice(in.Limitations)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (t ObjectiveFinalAnswerTrace) Clone() ObjectiveFinalAnswerTrace {
	return CloneObjectiveFinalAnswerTrace(t)
}

func (t ObjectiveFinalAnswerTrace) Normalize() ObjectiveFinalAnswerTrace {
	rawUnsafe := objectiveFinalAnswerTraceUnsafe(t)
	out := t.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.TraceRef = normalizeOneDisplaySafeRef(out.TraceRef)
	out.ObjectiveID = firstNonEmptyContractString(out.ObjectiveID)
	out.GoalSummary = objectiveFinalAnswerSafeText(out.GoalSummary)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = objectiveFinalAnswerTraceStepStatus(out.Steps)
	}
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Steps = normalizeObjectiveFinalAnswerTraceSteps(out.Steps)
	out.UsedCapabilityRefs = normalizeDisplaySafeRefs(append(out.UsedCapabilityRefs, objectiveFinalAnswerCapabilityRefsFromSteps(out.Steps)...))
	out.EvidenceRefs = MergeEvidenceRefs(out.EvidenceRefs, objectiveFinalAnswerEvidenceRefsFromSteps(out.Steps))
	out.Limitations = normalizeStringList(append(out.Limitations, objectiveFinalAnswerLimitationsFromSteps(out.Steps)...))
	out.MissingInputs = MergeMissingInputs(out.MissingInputs, objectiveFinalAnswerMissingInputsFromSteps(out.Steps))
	out.FailureClass = firstFailureClass(out.FailureClass, objectiveFinalAnswerFailureFromSteps(out.Steps))
	out.Boundaries = MergeBoundaries(out.Boundaries, objectiveFinalAnswerBoundariesFromSteps(out.Steps))
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if rawUnsafe || out.RawOutputLoaded || objectiveFinalAnswerStepsRawOutputLoaded(out.Steps) {
		out.RawOutputLoaded = true
		out.Status = VerificationReviewRequired
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func CloneObjectiveFinalAnswerDraft(in ObjectiveFinalAnswerDraft) ObjectiveFinalAnswerDraft {
	out := in
	out.Steps = cloneStringSlice(in.Steps)
	out.Evidence = cloneStringSlice(in.Evidence)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.Limitations = cloneStringSlice(in.Limitations)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (d ObjectiveFinalAnswerDraft) Clone() ObjectiveFinalAnswerDraft {
	return CloneObjectiveFinalAnswerDraft(d)
}

func (d ObjectiveFinalAnswerDraft) Normalize() ObjectiveFinalAnswerDraft {
	rawUnsafe := objectiveFinalAnswerDraftUnsafe(d)
	out := d.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.DraftRef = normalizeOneDisplaySafeRef(out.DraftRef)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	out.AnswerText = objectiveFinalAnswerSafeText(out.AnswerText)
	out.Conclusion = objectiveFinalAnswerSafeText(out.Conclusion)
	out.Steps = normalizeStringList(out.Steps)
	out.Evidence = normalizeStringList(out.Evidence)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.Limitations = normalizeStringList(out.Limitations)
	out.NextStep = objectiveFinalAnswerSafeText(out.NextStep)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	if rawUnsafe || out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.RawOutputLoaded = true
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
	}
	return out
}

func (r ObjectiveFinalAnswerSynthesisRequest) Normalize() ObjectiveFinalAnswerSynthesisRequest {
	out := r
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.RequestRef = normalizeOneDisplaySafeRef(out.RequestRef)
	out.Trace = out.Trace.Normalize()
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	if out.RawOutputLoaded {
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
	}
	return out
}

func (r ObjectiveFinalAnswerSynthesisResponse) Normalize() ObjectiveFinalAnswerSynthesisResponse {
	rawUnsafe := objectiveFinalAnswerDraftUnsafe(r.Draft)
	out := r
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.ResponseRef = normalizeOneDisplaySafeRef(out.ResponseRef)
	out.Draft = out.Draft.Normalize()
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	if rawUnsafe || out.RawOutputLoaded || out.Draft.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
	}
	return out
}

func CloneObjectiveFinalAnswer(in ObjectiveFinalAnswer) ObjectiveFinalAnswer {
	out := in
	out.Trace = in.Trace.Clone()
	out.Draft = in.Draft.Clone()
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (a ObjectiveFinalAnswer) Clone() ObjectiveFinalAnswer {
	return CloneObjectiveFinalAnswer(a)
}

func (a ObjectiveFinalAnswer) Normalize() ObjectiveFinalAnswer {
	out := a.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Trace = out.Trace.Normalize()
	out.Draft = out.Draft.Normalize()
	out.SynthesisRef = normalizeOneDisplaySafeRef(out.SynthesisRef)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded || out.Trace.RawOutputLoaded || out.Draft.RawOutputLoaded {
		out.RawOutputLoaded = true
		out.ReadyForUser = false
		out.Status = VerificationReviewRequired
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func normalizeObjectiveFinalAnswerTraceSteps(in []ObjectiveFinalAnswerTraceStep) []ObjectiveFinalAnswerTraceStep {
	out := make([]ObjectiveFinalAnswerTraceStep, 0, len(in))
	for _, value := range in {
		normalized := value.Normalize()
		if normalized.StepRef == "" && normalized.NodeRef == "" && normalized.Action == "" && len(normalized.EvidenceRefs) == 0 {
			continue
		}
		out = append(out, normalized)
	}
	return out
}

func cloneObjectiveFinalAnswerTraceSteps(in []ObjectiveFinalAnswerTraceStep) []ObjectiveFinalAnswerTraceStep {
	if len(in) == 0 {
		return nil
	}
	out := make([]ObjectiveFinalAnswerTraceStep, 0, len(in))
	for _, step := range in {
		out = append(out, step.Clone())
	}
	return out
}

func objectiveFinalAnswerTraceStepsFromGraph(graph ObjectiveGraph) []ObjectiveFinalAnswerTraceStep {
	graph = graph.Normalize()
	out := make([]ObjectiveFinalAnswerTraceStep, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		step := ObjectiveFinalAnswerTraceStep{
			StepRef:         firstDisplaySafeRef(node.NodeRef, "step:objective_node"),
			NodeRef:         node.NodeRef,
			CapabilityRef:   node.CapabilityRef,
			StrategyRef:     node.StrategyRef,
			DescriptorRef:   node.DescriptorRef,
			Status:          objectiveFinalAnswerStatusFromNodeState(node.State),
			Action:          node.Kind,
			EvidenceRefs:    node.RequiredEvidence,
			MissingInputs:   node.MissingInputs,
			FailureClass:    objectiveFinalAnswerFailureFromNodeState(node.State),
			Boundaries:      node.Boundaries,
			RawOutputLoaded: node.RawOutputLoaded,
		}.Normalize()
		out = append(out, step)
	}
	return out
}

func objectiveFinalAnswerTraceStepsFromReplan(replan ObjectiveReplanProposal) []ObjectiveFinalAnswerTraceStep {
	replan = replan.Normalize()
	out := make([]ObjectiveFinalAnswerTraceStep, 0, len(replan.Steps))
	for _, step := range replan.Steps {
		out = append(out, ObjectiveFinalAnswerTraceStep{
			StepRef:        step.StepRef,
			CapabilityRef:  firstDisplaySafeRef(step.CapabilityRefs...),
			StrategyRef:    firstDisplaySafeRef(step.NextStrategy, step.CurrentStrategy),
			Status:         replan.Status,
			Action:         string(step.Action),
			EvidenceRefs:   step.EvidenceRefs,
			MissingInputs:  step.MissingInputs,
			Boundaries:     step.Boundaries,
			NextHostAction: step.NextHostAction,
		}.Normalize())
	}
	return out
}

func objectiveFinalAnswerTraceStepsFromRecoveryPath(graph ObjectiveGraph, replan ObjectiveReplanProposal, patch ObjectiveReplanGraphPatch) []ObjectiveFinalAnswerTraceStep {
	graphSteps := objectiveFinalAnswerTraceStepsFromGraph(graph)
	patchSteps := objectiveFinalAnswerTraceStepsFromReplanGraphPatch(patch)
	replanSteps := objectiveFinalAnswerTraceStepsFromReplan(replan)
	if len(patchSteps) == 0 {
		if len(graphSteps) > 0 {
			return graphSteps
		}
		return replanSteps
	}
	return normalizeObjectiveFinalAnswerTraceSteps(append(append(replanSteps, patchSteps...), graphSteps...))
}

func objectiveFinalAnswerTraceStepsFromReplanGraphPatch(patch ObjectiveReplanGraphPatch) []ObjectiveFinalAnswerTraceStep {
	if !objectiveFinalAnswerReplanGraphPatchProvided(patch) {
		return nil
	}
	patch = patch.Normalize()
	if !patch.Available || patch.PatchRef == "" {
		return nil
	}
	out := []ObjectiveFinalAnswerTraceStep{ObjectiveFinalAnswerTraceStep{
		StepRef:        objectiveFinalAnswerPatchReviewStepRef(patch.PatchRef),
		Status:         patch.Status,
		Action:         "objective_replan_graph_patch_review",
		EvidenceRefs:   patch.EvidenceRefs,
		MissingInputs:  patch.MissingInputs,
		Limitations:    objectiveFinalAnswerLimitationsFromPatch(patch),
		Boundaries:     patch.Boundaries,
		NextHostAction: patch.NextHostAction,
	}.Normalize()}
	for _, node := range patch.PatchNodes {
		node = node.Normalize()
		out = append(out, ObjectiveFinalAnswerTraceStep{
			StepRef:         objectiveFinalAnswerPatchNodeStepRef(node.NodeRef),
			NodeRef:         node.NodeRef,
			CapabilityRef:   node.CapabilityRef,
			StrategyRef:     node.StrategyRef,
			DescriptorRef:   node.DescriptorRef,
			Status:          objectiveFinalAnswerStatusFromNodeState(node.State),
			Action:          "objective_replan_graph_patch_node",
			EvidenceRefs:    node.RequiredEvidence,
			MissingInputs:   node.MissingInputs,
			Limitations:     []string{"host must bind recovery node before runtime"},
			Boundaries:      node.Boundaries,
			NextHostAction:  patch.NextHostAction,
			RawOutputLoaded: node.RawOutputLoaded,
		}.Normalize())
	}
	return normalizeObjectiveFinalAnswerTraceSteps(out)
}

func objectiveFinalAnswerPatchReviewStepRef(ref DisplaySafeRef) DisplaySafeRef {
	token := strings.TrimPrefix(string(ref), "patch:")
	token = normalizeControlToken(token)
	if token == "" {
		token = "objective_replan_graph_patch"
	}
	if len(token) > 80 {
		token = strings.Trim(token[:80], "_.:-")
	}
	return DisplaySafeRef("step:patch_review_" + token)
}

func objectiveFinalAnswerPatchNodeStepRef(ref DisplaySafeRef) DisplaySafeRef {
	token := strings.TrimPrefix(string(ref), "node:")
	token = normalizeControlToken(token)
	if token == "" {
		token = "objective_replan_graph_patch_node"
	}
	if len(token) > 80 {
		token = strings.Trim(token[:80], "_.:-")
	}
	return DisplaySafeRef("step:patch_node_" + token)
}

func objectiveFinalAnswerLimitationsFromPatch(patch ObjectiveReplanGraphPatch) []string {
	patch = patch.Normalize()
	if patch.ReadyForGraphApply {
		return nil
	}
	if patch.ReadyForHostReview {
		return []string{"recovery graph patch requires host review before runtime"}
	}
	if patch.Status == VerificationNotApplicable {
		return nil
	}
	return []string{"recovery graph patch is not ready for runtime"}
}

func objectiveFinalAnswerFallbackDraft(trace ObjectiveFinalAnswerTrace) ObjectiveFinalAnswerDraft {
	trace = trace.Normalize()
	draft := ObjectiveFinalAnswerDraft{
		DraftRef:     "draft:objective_final_answer_fallback",
		Status:       trace.Status,
		Conclusion:   objectiveFinalAnswerFallbackConclusion(trace.Status),
		AnswerText:   objectiveFinalAnswerFallbackConclusion(trace.Status),
		Steps:        objectiveFinalAnswerDraftStepsFromTrace(trace),
		EvidenceRefs: trace.EvidenceRefs,
		Limitations:  trace.Limitations,
		NextStep:     string(trace.NextHostAction),
		Boundaries: []Boundary{
			"deterministic_final_answer_fallback",
			"host_may_replace_with_llm_phrase_only_synthesis",
		},
	}
	if draft.NextStep == "" && trace.Status != VerificationSatisfied {
		draft.NextStep = "review_objective_trace"
	}
	return draft
}

func objectiveFinalAnswerFallbackConclusion(status VerificationStatus) string {
	switch NormalizeVerificationStatus(string(status)) {
	case VerificationSatisfied:
		return "Objective satisfied with trace-backed evidence."
	case VerificationPartial:
		return "Objective partially satisfied; see limitations and next step."
	case VerificationBlocked, VerificationFailed, VerificationReviewRequired:
		return "Objective cannot be completed without the next host action."
	default:
		return "Objective result requires host review."
	}
}

func objectiveFinalAnswerDraftStepsFromTrace(trace ObjectiveFinalAnswerTrace) []string {
	out := make([]string, 0, len(trace.Steps))
	for _, step := range trace.Steps {
		label := firstNonEmptyContractString(step.Action, string(step.CapabilityRef), string(step.NodeRef), string(step.StepRef))
		if label != "" {
			out = append(out, label)
		}
	}
	return normalizeStringList(out)
}

func objectiveFinalAnswerCompletenessIssue(trace ObjectiveFinalAnswerTrace, draft ObjectiveFinalAnswerDraft) (MissingInput, NextHostAction, Boundary, FailureClass) {
	trace = trace.Normalize()
	draft = draft.Normalize()
	if objectiveFinalAnswerDraftEmpty(draft) || (draft.AnswerText == "" && draft.Conclusion == "") {
		return "host:objective_final_answer_text", "provide_trace_backed_final_answer", "final_answer_text_missing", FailureEvidenceMissing
	}
	if len(trace.Steps) == 0 {
		return "host:objective_final_answer_trace_steps", "provide_objective_execution_trace", "final_answer_trace_steps_missing", FailureEvidenceMissing
	}
	if len(trace.EvidenceRefs) == 0 {
		return "host:objective_final_answer_evidence_refs", "provide_objective_evidence_refs", "final_answer_evidence_missing", FailureEvidenceMissing
	}
	switch trace.Status {
	case VerificationSatisfied:
		if trace.FailureClass != FailureNone {
			return "host:objective_final_answer_success_state", "review_objective_trace", "satisfied_answer_has_failure_class", FailureVerificationFailed
		}
	case VerificationPartial:
		if len(trace.Limitations) == 0 && len(draft.Limitations) == 0 {
			return "host:objective_final_answer_limitations", "provide_partial_limitations", "partial_answer_limitations_missing", FailureInsufficientInformation
		}
		if trace.NextHostAction == "" && draft.NextStep == "" {
			return "host:objective_final_answer_next_step", "provide_partial_next_step", "partial_answer_next_step_missing", FailureInsufficientInformation
		}
	case VerificationBlocked, VerificationFailed, VerificationReviewRequired:
		if trace.FailureClass == FailureNone && len(trace.MissingInputs) == 0 && len(trace.Limitations) == 0 && len(draft.Limitations) == 0 {
			return "host:objective_final_answer_block_reason", "provide_blocked_reason", "blocked_answer_reason_missing", FailureInsufficientInformation
		}
		if trace.NextHostAction == "" && draft.NextStep == "" {
			return "host:objective_final_answer_next_host_action", "provide_blocked_next_action", "blocked_answer_next_action_missing", FailureInsufficientInformation
		}
	default:
		return "host:objective_final_answer_status", "review_objective_trace", "final_answer_status_not_ready", FailureInsufficientInformation
	}
	return "", "", "", FailureNone
}

func objectiveFinalAnswerBlock(result ObjectiveFinalAnswer, status VerificationStatus, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) ObjectiveFinalAnswer {
	result.ReadyForUser = false
	result.Status = status
	result.FailureClass = firstFailureClass(failure, result.FailureClass)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	return result.Normalize()
}

func objectiveFinalAnswerTraceEmpty(trace ObjectiveFinalAnswerTrace) bool {
	return trace.TraceRef == "" && trace.ObjectiveID == "" && trace.GoalSummary == "" && len(trace.Steps) == 0 && len(trace.EvidenceRefs) == 0
}

func objectiveFinalAnswerDraftEmpty(draft ObjectiveFinalAnswerDraft) bool {
	return draft.DraftRef == "" && draft.AnswerText == "" && draft.Conclusion == "" && len(draft.Steps) == 0 && len(draft.EvidenceRefs) == 0
}

func objectiveFinalAnswerSafeText(value string) string {
	return objectiveSpecSafeText(value)
}

func objectiveFinalAnswerTraceUnsafe(trace ObjectiveFinalAnswerTrace) bool {
	if ContainsUnsafeRawOutput(trace.ObjectiveID, trace.GoalSummary) {
		return true
	}
	if displaySafeRefRejected(trace.TraceRef) || displaySafeRefSliceRejected(trace.UsedCapabilityRefs) || evidenceRefRejected(trace.EvidenceRefs) {
		return true
	}
	for _, step := range trace.Steps {
		if objectiveFinalAnswerTraceStepUnsafe(step) {
			return true
		}
	}
	return objectiveFinalAnswerStringListUnsafe(trace.Limitations)
}

func objectiveFinalAnswerTraceInputUnsafe(input ObjectiveFinalAnswerTraceInput) bool {
	return displaySafeRefRejected(input.TraceRef) ||
		displaySafeRefSliceRejected(input.UsedCapabilityRefs) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		objectiveFinalAnswerStringListUnsafe(input.Limitations) ||
		objectiveFinalAnswerTraceStepsUnsafe(input.Steps) ||
		objectiveFinalAnswerReplanGraphPatchUnsafe(input.ReplanGraphPatch)
}

func objectiveFinalAnswerTraceStepsUnsafe(steps []ObjectiveFinalAnswerTraceStep) bool {
	for _, step := range steps {
		if objectiveFinalAnswerTraceStepUnsafe(step) {
			return true
		}
	}
	return false
}

func objectiveFinalAnswerTraceStepUnsafe(step ObjectiveFinalAnswerTraceStep) bool {
	return displaySafeRefRejected(step.StepRef) ||
		displaySafeRefRejected(step.NodeRef) ||
		displaySafeRefRejected(step.CapabilityRef) ||
		displaySafeRefRejected(step.StrategyRef) ||
		displaySafeRefRejected(step.DescriptorRef) ||
		evidenceRefRejected(step.EvidenceRefs) ||
		ContainsUnsafeRawOutput(step.Action) ||
		objectiveFinalAnswerStringListUnsafe(step.Limitations)
}

func objectiveFinalAnswerReplanGraphPatchUnsafe(patch ObjectiveReplanGraphPatch) bool {
	if !objectiveFinalAnswerReplanGraphPatchProvided(patch) {
		return false
	}
	if displaySafeRefRejected(patch.PatchRef) ||
		displaySafeRefRejected(patch.SourceGraphRef) ||
		displaySafeRefRejected(patch.SourceNodeRef) ||
		displaySafeRefRejected(patch.ProposalRef) ||
		evidenceRefRejected(patch.EvidenceRefs) {
		return true
	}
	for _, node := range patch.PatchNodes {
		if objectiveFinalAnswerTraceNodeUnsafe(node) {
			return true
		}
	}
	return false
}

func objectiveFinalAnswerReplanGraphPatchProvided(patch ObjectiveReplanGraphPatch) bool {
	return patch.PatchRef != "" ||
		patch.SourceGraphRef != "" ||
		patch.SourceNodeRef != "" ||
		patch.ProposalRef != "" ||
		patch.Status != "" ||
		patch.Action != "" ||
		len(patch.PatchNodes) > 0 ||
		len(patch.EvidenceRefs) > 0 ||
		len(patch.MissingInputs) > 0 ||
		len(patch.BlockedReasons) > 0 ||
		len(patch.Boundaries) > 0 ||
		patch.NextHostAction != "" ||
		patch.RawOutputLoaded
}

func objectiveFinalAnswerTraceNodeUnsafe(node ObjectiveNode) bool {
	return displaySafeRefRejected(node.NodeRef) ||
		displaySafeRefRejected(node.CapabilityRef) ||
		displaySafeRefRejected(node.StrategyRef) ||
		displaySafeRefRejected(node.DescriptorRef) ||
		displaySafeRefRejected(node.SourceRef) ||
		displaySafeRefRejected(node.InputSchemaRef) ||
		displaySafeRefRejected(node.OutputSchemaRef) ||
		displaySafeRefRejected(node.EvidenceContractRef) ||
		evidenceRefRejected(node.RequiredEvidence) ||
		displaySafeRefSliceRejected(node.ApprovalRefs) ||
		displaySafeRefSliceRejected(node.PolicyRefs)
}

func objectiveFinalAnswerDraftUnsafe(draft ObjectiveFinalAnswerDraft) bool {
	return ContainsUnsafeRawOutput(draft.AnswerText, draft.Conclusion, draft.NextStep) ||
		objectiveFinalAnswerStringListUnsafe(draft.Steps) ||
		objectiveFinalAnswerStringListUnsafe(draft.Evidence) ||
		evidenceRefRejected(draft.EvidenceRefs) ||
		displaySafeRefRejected(draft.DraftRef)
}

func objectiveFinalAnswerStringListUnsafe(values []string) bool {
	for _, value := range values {
		if ContainsUnsafeRawOutput(value) {
			return true
		}
	}
	return false
}

func firstObjectiveFinalAnswerStatus(values ...VerificationStatus) VerificationStatus {
	for _, value := range values {
		status := NormalizeVerificationStatus(string(value))
		if status != VerificationNotEvaluated && status != VerificationNotApplicable {
			return status
		}
	}
	return VerificationNotEvaluated
}

func objectiveFinalAnswerTraceStepStatus(steps []ObjectiveFinalAnswerTraceStep) VerificationStatus {
	hasPartial := false
	hasSatisfied := false
	for _, step := range normalizeObjectiveFinalAnswerTraceSteps(steps) {
		switch step.Status {
		case VerificationBlocked, VerificationFailed, VerificationReviewRequired:
			return step.Status
		case VerificationPartial:
			hasPartial = true
		case VerificationSatisfied:
			hasSatisfied = true
		}
	}
	if hasPartial {
		return VerificationPartial
	}
	if hasSatisfied {
		return VerificationSatisfied
	}
	return VerificationNotEvaluated
}

func objectiveFinalAnswerStatusFromNodeState(state ObjectiveNodeState) VerificationStatus {
	switch NormalizeObjectiveNodeState(string(state)) {
	case ObjectiveNodeStateSatisfied, ObjectiveNodeStateSkipped:
		return VerificationSatisfied
	case ObjectiveNodeStatePartial:
		return VerificationPartial
	case ObjectiveNodeStateBlocked:
		return VerificationBlocked
	case ObjectiveNodeStateFailedRetryable:
		return VerificationFailed
	case ObjectiveNodeStateRunning, ObjectiveNodeStateReady, ObjectiveNodeStatePending:
		return VerificationNotEvaluated
	default:
		return VerificationBlocked
	}
}

func objectiveFinalAnswerFailureFromNodeState(state ObjectiveNodeState) FailureClass {
	switch NormalizeObjectiveNodeState(string(state)) {
	case ObjectiveNodeStateBlocked:
		return FailureInsufficientInformation
	case ObjectiveNodeStateFailedRetryable:
		return FailureVerificationFailed
	default:
		return FailureNone
	}
}

func objectiveFinalAnswerFailureFromReplan(replan ObjectiveReplanProposal) FailureClass {
	replan = replan.Normalize()
	switch replan.Action {
	case ObjectiveReplanProposalActionCapabilityGap:
		return FailureCapabilityMissing
	case ObjectiveReplanProposalActionRequestApproval:
		return FailureApprovalRequired
	case ObjectiveReplanProposalActionAskUser:
		return FailureInsufficientInformation
	case ObjectiveReplanProposalActionReviewRefs:
		return FailureEvidenceWeak
	default:
		return FailureNone
	}
}

func objectiveFinalAnswerCapabilityRefsFromGraph(graph ObjectiveGraph) []DisplaySafeRef {
	out := []DisplaySafeRef{}
	for _, node := range graph.Normalize().Nodes {
		out = append(out, node.CapabilityRef)
	}
	return normalizeDisplaySafeRefs(out)
}

func objectiveFinalAnswerCapabilityRefsFromSteps(steps []ObjectiveFinalAnswerTraceStep) []DisplaySafeRef {
	out := []DisplaySafeRef{}
	for _, step := range steps {
		out = append(out, step.CapabilityRef)
	}
	return normalizeDisplaySafeRefs(out)
}

func objectiveFinalAnswerEvidenceRefsFromSteps(steps []ObjectiveFinalAnswerTraceStep) []EvidenceRef {
	out := []EvidenceRef{}
	for _, step := range normalizeObjectiveFinalAnswerTraceSteps(steps) {
		out = append(out, step.EvidenceRefs...)
	}
	return normalizeEvidenceRefs(out)
}

func objectiveFinalAnswerMissingInputsFromSteps(steps []ObjectiveFinalAnswerTraceStep) []MissingInput {
	out := []MissingInput{}
	for _, step := range normalizeObjectiveFinalAnswerTraceSteps(steps) {
		out = append(out, step.MissingInputs...)
	}
	return normalizeMissingInputs(out)
}

func objectiveFinalAnswerLimitationsFromSteps(steps []ObjectiveFinalAnswerTraceStep) []string {
	out := []string{}
	for _, step := range normalizeObjectiveFinalAnswerTraceSteps(steps) {
		out = append(out, step.Limitations...)
	}
	return normalizeStringList(out)
}

func objectiveFinalAnswerBoundariesFromSteps(steps []ObjectiveFinalAnswerTraceStep) []Boundary {
	out := []Boundary{}
	for _, step := range normalizeObjectiveFinalAnswerTraceSteps(steps) {
		out = append(out, step.Boundaries...)
	}
	return normalizeBoundaries(out)
}

func objectiveFinalAnswerFailureFromSteps(steps []ObjectiveFinalAnswerTraceStep) FailureClass {
	for _, step := range normalizeObjectiveFinalAnswerTraceSteps(steps) {
		if failure := NormalizeFailureClass(string(step.FailureClass)); failure != FailureNone {
			return failure
		}
	}
	return FailureNone
}

func objectiveFinalAnswerStepsRawOutputLoaded(steps []ObjectiveFinalAnswerTraceStep) bool {
	for _, step := range steps {
		if step.RawOutputLoaded {
			return true
		}
	}
	return false
}
