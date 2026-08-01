package controlcontract

type AutoDelegationAsyncBackendKind string

const (
	AutoDelegationAsyncBackendUnsupported  AutoDelegationAsyncBackendKind = "unsupported"
	AutoDelegationAsyncBackendProcessLocal AutoDelegationAsyncBackendKind = "process_local"
	AutoDelegationAsyncBackendDurable      AutoDelegationAsyncBackendKind = "durable"
)

func NormalizeAutoDelegationAsyncBackendKind(raw string) AutoDelegationAsyncBackendKind {
	switch normalizeEnumToken(raw) {
	case "process_local", "process", "local", "in_process", "memory":
		return AutoDelegationAsyncBackendProcessLocal
	case "durable", "persistent", "store", "stored":
		return AutoDelegationAsyncBackendDurable
	case "", "unsupported", "none", "unknown":
		return AutoDelegationAsyncBackendUnsupported
	default:
		return AutoDelegationAsyncBackendUnsupported
	}
}

type AutoDelegationAsyncChildStatus string

const (
	AutoDelegationAsyncChildStatusUnknown     AutoDelegationAsyncChildStatus = "unknown"
	AutoDelegationAsyncChildStatusQueued      AutoDelegationAsyncChildStatus = "queued"
	AutoDelegationAsyncChildStatusActive      AutoDelegationAsyncChildStatus = "active"
	AutoDelegationAsyncChildStatusCompleted   AutoDelegationAsyncChildStatus = "completed"
	AutoDelegationAsyncChildStatusFailed      AutoDelegationAsyncChildStatus = "failed"
	AutoDelegationAsyncChildStatusCancelled   AutoDelegationAsyncChildStatus = "cancelled"
	AutoDelegationAsyncChildStatusInterrupted AutoDelegationAsyncChildStatus = "interrupted"
)

func NormalizeAutoDelegationAsyncChildStatus(raw string) AutoDelegationAsyncChildStatus {
	switch normalizeEnumToken(raw) {
	case "queued", "pending", "waiting":
		return AutoDelegationAsyncChildStatusQueued
	case "active", "running", "in_progress", "started":
		return AutoDelegationAsyncChildStatusActive
	case "completed", "complete", "done", "succeeded", "success":
		return AutoDelegationAsyncChildStatusCompleted
	case "failed", "failure", "error":
		return AutoDelegationAsyncChildStatusFailed
	case "cancelled", "canceled", "cancel":
		return AutoDelegationAsyncChildStatusCancelled
	case "interrupted", "interrupt":
		return AutoDelegationAsyncChildStatusInterrupted
	default:
		return AutoDelegationAsyncChildStatusUnknown
	}
}

type AutoDelegationAsyncCompletionInput struct {
	HostBridge            AutoDelegationHostBridge           `json:"host_bridge,omitempty"`
	BackendKind           AutoDelegationAsyncBackendKind     `json:"backend_kind,omitempty"`
	BackendRef            DisplaySafeRef                     `json:"backend_ref,omitempty"`
	ParentObjectiveRef    DisplaySafeRef                     `json:"parent_objective_ref,omitempty"`
	ParentObjectiveRunRef DisplaySafeRef                     `json:"parent_objective_run_ref,omitempty"`
	ParentLedgerRef       DisplaySafeRef                     `json:"parent_ledger_ref,omitempty"`
	ParentResumeRef       DisplaySafeRef                     `json:"parent_resume_ref,omitempty"`
	RequireDurable        bool                               `json:"require_durable"`
	Children              []AutoDelegationAsyncChildReadback `json:"children,omitempty"`
	DecisionBasis         []DisplaySafeRef                   `json:"decision_basis,omitempty"`
	Boundaries            []Boundary                         `json:"boundaries,omitempty"`
	RawOutputLoaded       bool                               `json:"raw_output_loaded"`
}

type AutoDelegationAsyncChildReadback struct {
	ChildRef                         DisplaySafeRef                 `json:"child_ref,omitempty"`
	Role                             AutoDelegationChildRole        `json:"role,omitempty"`
	Status                           AutoDelegationAsyncChildStatus `json:"status,omitempty"`
	AgeSeconds                       int                            `json:"age_seconds,omitempty"`
	CurrentAction                    string                         `json:"current_action,omitempty"`
	CapabilityRefs                   []DisplaySafeRef               `json:"capability_refs,omitempty"`
	AllowedToolRefs                  []DisplaySafeRef               `json:"allowed_tool_refs,omitempty"`
	BoundCapabilityRefs              []DisplaySafeRef               `json:"bound_capability_refs,omitempty"`
	BoundAllowedToolRefs             []DisplaySafeRef               `json:"bound_allowed_tool_refs,omitempty"`
	WorkerRunRef                     DisplaySafeRef                 `json:"worker_run_ref,omitempty"`
	WorkerResultRef                  DisplaySafeRef                 `json:"worker_result_ref,omitempty"`
	WorkerReadbackRef                DisplaySafeRef                 `json:"worker_readback_ref,omitempty"`
	ObservationRef                   DisplaySafeRef                 `json:"observation_ref,omitempty"`
	FailureRef                       DisplaySafeRef                 `json:"failure_ref,omitempty"`
	FailureReviewRef                 DisplaySafeRef                 `json:"failure_review_ref,omitempty"`
	CancellationRef                  DisplaySafeRef                 `json:"cancellation_ref,omitempty"`
	InterruptionRef                  DisplaySafeRef                 `json:"interruption_ref,omitempty"`
	CompletionEnvelopeRef            DisplaySafeRef                 `json:"completion_envelope_ref,omitempty"`
	EvidenceRefs                     []EvidenceRef                  `json:"evidence_refs,omitempty"`
	ReadyForWorkerResultReview       bool                           `json:"ready_for_worker_result_review"`
	ReadyForFailureReview            bool                           `json:"ready_for_failure_review"`
	CancelAvailable                  bool                           `json:"cancel_available"`
	InterruptAvailable               bool                           `json:"interrupt_available"`
	WorkerResultRequiresVerification bool                           `json:"worker_result_requires_verification"`
	WorkerOutputAcceptedAsFact       bool                           `json:"worker_output_accepted_as_fact"`
	MissingInputs                    []MissingInput                 `json:"missing_inputs,omitempty"`
	BlockedReasons                   []string                       `json:"blocked_reasons,omitempty"`
	FailureClass                     FailureClass                   `json:"failure_class,omitempty"`
	Boundaries                       []Boundary                     `json:"boundaries,omitempty"`
	RawOutputLoaded                  bool                           `json:"raw_output_loaded"`
}

type AutoDelegationAsyncChildStatusProjection struct {
	ChildRef                         DisplaySafeRef                 `json:"child_ref,omitempty"`
	Role                             AutoDelegationChildRole        `json:"role,omitempty"`
	Status                           AutoDelegationAsyncChildStatus `json:"status,omitempty"`
	AgeSeconds                       int                            `json:"age_seconds,omitempty"`
	CurrentAction                    string                         `json:"current_action,omitempty"`
	CapabilityRefs                   []DisplaySafeRef               `json:"capability_refs,omitempty"`
	AllowedToolRefs                  []DisplaySafeRef               `json:"allowed_tool_refs,omitempty"`
	BoundCapabilityRefs              []DisplaySafeRef               `json:"bound_capability_refs,omitempty"`
	BoundAllowedToolRefs             []DisplaySafeRef               `json:"bound_allowed_tool_refs,omitempty"`
	WorkerRunRef                     DisplaySafeRef                 `json:"worker_run_ref,omitempty"`
	WorkerResultRef                  DisplaySafeRef                 `json:"worker_result_ref,omitempty"`
	WorkerReadbackRef                DisplaySafeRef                 `json:"worker_readback_ref,omitempty"`
	ObservationRef                   DisplaySafeRef                 `json:"observation_ref,omitempty"`
	FailureRef                       DisplaySafeRef                 `json:"failure_ref,omitempty"`
	FailureReviewRef                 DisplaySafeRef                 `json:"failure_review_ref,omitempty"`
	CancellationRef                  DisplaySafeRef                 `json:"cancellation_ref,omitempty"`
	InterruptionRef                  DisplaySafeRef                 `json:"interruption_ref,omitempty"`
	CompletionEnvelopeRef            DisplaySafeRef                 `json:"completion_envelope_ref,omitempty"`
	EvidenceRefs                     []EvidenceRef                  `json:"evidence_refs,omitempty"`
	ReadbackReady                    bool                           `json:"readback_ready"`
	ReadyForWorkerResultReview       bool                           `json:"ready_for_worker_result_review"`
	ReadyForFailureReview            bool                           `json:"ready_for_failure_review"`
	CancelAvailable                  bool                           `json:"cancel_available"`
	InterruptAvailable               bool                           `json:"interrupt_available"`
	WorkerResultRequiresVerification bool                           `json:"worker_result_requires_verification"`
	WorkerOutputAcceptedAsFact       bool                           `json:"worker_output_accepted_as_fact"`
	MissingInputs                    []MissingInput                 `json:"missing_inputs,omitempty"`
	BlockedReasons                   []string                       `json:"blocked_reasons,omitempty"`
	FailureClass                     FailureClass                   `json:"failure_class,omitempty"`
	Boundaries                       []Boundary                     `json:"boundaries,omitempty"`
	RawOutputLoaded                  bool                           `json:"raw_output_loaded"`
}

type AutoDelegationChildCompletionEnvelope struct {
	EnvelopeRef                DisplaySafeRef          `json:"envelope_ref,omitempty"`
	ChildRef                   DisplaySafeRef          `json:"child_ref,omitempty"`
	Role                       AutoDelegationChildRole `json:"role,omitempty"`
	WorkerRunRef               DisplaySafeRef          `json:"worker_run_ref,omitempty"`
	WorkerResultRef            DisplaySafeRef          `json:"worker_result_ref,omitempty"`
	WorkerReadbackRef          DisplaySafeRef          `json:"worker_readback_ref,omitempty"`
	ObservationRef             DisplaySafeRef          `json:"observation_ref,omitempty"`
	EvidenceRefs               []EvidenceRef           `json:"evidence_refs,omitempty"`
	BoundCapabilityRefs        []DisplaySafeRef        `json:"bound_capability_refs,omitempty"`
	BoundAllowedToolRefs       []DisplaySafeRef        `json:"bound_allowed_tool_refs,omitempty"`
	RequiresParentVerification bool                    `json:"requires_parent_verification"`
	WorkerOutputAcceptedAsFact bool                    `json:"worker_output_accepted_as_fact"`
}

type AutoDelegationParentResumeRequest struct {
	ParentObjectiveRef     DisplaySafeRef   `json:"parent_objective_ref,omitempty"`
	ParentObjectiveRunRef  DisplaySafeRef   `json:"parent_objective_run_ref,omitempty"`
	ParentLedgerRef        DisplaySafeRef   `json:"parent_ledger_ref,omitempty"`
	ParentResumeRef        DisplaySafeRef   `json:"parent_resume_ref,omitempty"`
	ChildRefs              []DisplaySafeRef `json:"child_refs,omitempty"`
	CompletionEnvelopeRefs []DisplaySafeRef `json:"completion_envelope_refs,omitempty"`
	WorkerResultRefs       []DisplaySafeRef `json:"worker_result_refs,omitempty"`
	WorkerReadbackRefs     []DisplaySafeRef `json:"worker_readback_refs,omitempty"`
}

type AutoDelegationAsyncCompletionProjection struct {
	ContractVersion       string                                     `json:"contract_version,omitempty"`
	Projected             bool                                       `json:"projected"`
	Status                VerificationStatus                         `json:"status,omitempty"`
	Ready                 bool                                       `json:"ready"`
	ReadyForReadback      bool                                       `json:"ready_for_readback"`
	ReadyForResume        bool                                       `json:"ready_for_resume"`
	BackendKind           AutoDelegationAsyncBackendKind             `json:"backend_kind,omitempty"`
	BackendRef            DisplaySafeRef                             `json:"backend_ref,omitempty"`
	ProcessLocal          bool                                       `json:"process_local"`
	Durable               bool                                       `json:"durable"`
	UnsupportedBackend    bool                                       `json:"unsupported_backend"`
	RequireDurable        bool                                       `json:"require_durable"`
	ParentObjectiveRef    DisplaySafeRef                             `json:"parent_objective_ref,omitempty"`
	ParentObjectiveRunRef DisplaySafeRef                             `json:"parent_objective_run_ref,omitempty"`
	ParentLedgerRef       DisplaySafeRef                             `json:"parent_ledger_ref,omitempty"`
	ParentResumeRef       DisplaySafeRef                             `json:"parent_resume_ref,omitempty"`
	HostBridge            AutoDelegationHostBridge                   `json:"host_bridge,omitempty"`
	Children              []AutoDelegationAsyncChildStatusProjection `json:"children,omitempty"`
	QueuedChildRefs       []DisplaySafeRef                           `json:"queued_child_refs,omitempty"`
	ActiveChildRefs       []DisplaySafeRef                           `json:"active_child_refs,omitempty"`
	CompletedChildRefs    []DisplaySafeRef                           `json:"completed_child_refs,omitempty"`
	FailedChildRefs       []DisplaySafeRef                           `json:"failed_child_refs,omitempty"`
	CancelledChildRefs    []DisplaySafeRef                           `json:"cancelled_child_refs,omitempty"`
	InterruptedChildRefs  []DisplaySafeRef                           `json:"interrupted_child_refs,omitempty"`
	CompletionEnvelopes   []AutoDelegationChildCompletionEnvelope    `json:"completion_envelopes,omitempty"`
	ResumeRequest         AutoDelegationParentResumeRequest          `json:"resume_request,omitempty"`
	MissingInputs         []MissingInput                             `json:"missing_inputs,omitempty"`
	BlockedReasons        []string                                   `json:"blocked_reasons,omitempty"`
	FailureClass          FailureClass                               `json:"failure_class,omitempty"`
	DecisionBasis         []DisplaySafeRef                           `json:"decision_basis,omitempty"`
	Boundaries            []Boundary                                 `json:"boundaries,omitempty"`
	NextHostAction        NextHostAction                             `json:"next_host_action,omitempty"`
	RunnerEffect          string                                     `json:"runner_effect,omitempty"`
	PromptEffect          string                                     `json:"prompt_effect,omitempty"`
	RuntimeEffect         string                                     `json:"runtime_effect,omitempty"`
	RawOutputLoaded       bool                                       `json:"raw_output_loaded"`
}

func BuildAutoDelegationAsyncCompletionProjection(input AutoDelegationAsyncCompletionInput) AutoDelegationAsyncCompletionProjection {
	hostBridge := input.HostBridge.Normalize()
	backendKind := NormalizeAutoDelegationAsyncBackendKind(string(input.BackendKind))
	children := autoDelegationAsyncNormalizeChildReadbacks(input.Children)
	result := AutoDelegationAsyncCompletionProjection{
		ContractVersion:       ContractVersion,
		Projected:             true,
		Status:                VerificationBlocked,
		BackendKind:           backendKind,
		BackendRef:            normalizeOneDisplaySafeRef(input.BackendRef),
		ProcessLocal:          backendKind == AutoDelegationAsyncBackendProcessLocal,
		Durable:               backendKind == AutoDelegationAsyncBackendDurable,
		UnsupportedBackend:    backendKind == AutoDelegationAsyncBackendUnsupported,
		RequireDurable:        input.RequireDurable,
		ParentObjectiveRef:    normalizeOneDisplaySafeRef(input.ParentObjectiveRef),
		ParentObjectiveRunRef: normalizeOneDisplaySafeRef(input.ParentObjectiveRunRef),
		ParentLedgerRef:       normalizeOneDisplaySafeRef(input.ParentLedgerRef),
		ParentResumeRef:       normalizeOneDisplaySafeRef(input.ParentResumeRef),
		HostBridge:            hostBridge,
		FailureClass:          FailureNone,
		DecisionBasis: normalizeDisplaySafeRefs(append(
			[]DisplaySafeRef{
				"auto_delegation:async_completion",
				"auto_delegation:host_owned_async_readback",
			},
			input.DecisionBasis...,
		)),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"auto_delegation_async_completion",
				"host_owned_async_child_runtime",
				"completion_envelope_only",
				"raw_child_tool_logs_not_consumed",
				"parent_verification_required",
				"child_output_not_fact",
				"projection_only",
				"display_safe_refs_only",
				"no_background_process_by_core",
				"no_child_task_spawn_by_core",
				"no_runner_dispatch",
			},
			input.Boundaries,
			hostBridge.Boundaries,
		),
		NextHostAction:  "review_auto_delegation_async_completion",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RuntimeEffect:   "none",
		RawOutputLoaded: input.RawOutputLoaded || hostBridge.RawOutputLoaded,
	}
	if autoDelegationAsyncCompletionUnsafe(input, children) || result.RawOutputLoaded {
		result.RawOutputLoaded = true
		return autoDelegationAsyncCompletionBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed").Normalize()
	}
	if result.UnsupportedBackend {
		return autoDelegationAsyncCompletionBlock(result, VerificationBlocked, FailureUnsupportedOperation, "auto_delegation_async_backend_unsupported", "host:auto_delegation_async_backend", "provide_auto_delegation_async_backend", "auto_delegation_async_backend_unsupported").Normalize()
	}
	if result.RequireDurable && !result.Durable {
		return autoDelegationAsyncCompletionBlock(result, VerificationBlocked, FailureUnsupportedOperation, "durable_auto_delegation_child_readback_required", "host:durable_auto_delegation_child_readback", "provide_durable_auto_delegation_child_readback", "durable_child_readback_required").Normalize()
	}
	if len(children) == 0 {
		return autoDelegationAsyncCompletionBlock(result, VerificationBlocked, FailureInsufficientInformation, "auto_delegation_async_child_readback_missing", "host:auto_delegation_async_child_readback", "provide_auto_delegation_async_child_readback", "auto_delegation_async_child_readback_missing").Normalize()
	}
	for _, child := range children {
		status := autoDelegationAsyncChildStatusProjection(child)
		result.Children = append(result.Children, status)
		result.MissingInputs = AppendMissingInputs(result.MissingInputs, status.MissingInputs...)
		result.BlockedReasons = appendUniqueControlTokens(result.BlockedReasons, status.BlockedReasons)
		result.Boundaries = AppendBoundaries(result.Boundaries, status.Boundaries...)
		result.FailureClass = firstFailureClass(result.FailureClass, status.FailureClass)
		result.RawOutputLoaded = result.RawOutputLoaded || status.RawOutputLoaded
		switch status.Status {
		case AutoDelegationAsyncChildStatusQueued:
			result.QueuedChildRefs = appendDisplaySafeRefIfPresent(result.QueuedChildRefs, status.ChildRef)
		case AutoDelegationAsyncChildStatusActive:
			result.ActiveChildRefs = appendDisplaySafeRefIfPresent(result.ActiveChildRefs, status.ChildRef)
		case AutoDelegationAsyncChildStatusCompleted:
			result.CompletedChildRefs = appendDisplaySafeRefIfPresent(result.CompletedChildRefs, status.ChildRef)
			if envelope, ok := autoDelegationAsyncCompletionEnvelope(status); ok {
				result.CompletionEnvelopes = append(result.CompletionEnvelopes, envelope)
			}
		case AutoDelegationAsyncChildStatusFailed:
			result.FailedChildRefs = appendDisplaySafeRefIfPresent(result.FailedChildRefs, status.ChildRef)
		case AutoDelegationAsyncChildStatusCancelled:
			result.CancelledChildRefs = appendDisplaySafeRefIfPresent(result.CancelledChildRefs, status.ChildRef)
		case AutoDelegationAsyncChildStatusInterrupted:
			result.InterruptedChildRefs = appendDisplaySafeRefIfPresent(result.InterruptedChildRefs, status.ChildRef)
		default:
			result = autoDelegationAsyncCompletionBlock(result, VerificationBlocked, FailureInsufficientInformation, "auto_delegation_child_status_missing", "host:auto_delegation_child_status", "provide_auto_delegation_async_child_readback", "auto_delegation_child_status_missing")
		}
	}
	if result.RawOutputLoaded {
		return result.Normalize()
	}
	if len(result.MissingInputs) > 0 || len(result.BlockedReasons) > 0 {
		result.Status = VerificationBlocked
		result.Ready = false
		result.NextHostAction = firstNextHostAction(result.NextHostAction, "provide_auto_delegation_async_child_readback")
		return result.Normalize()
	}
	openOrTerminalIssueCount := len(result.QueuedChildRefs) + len(result.ActiveChildRefs) + len(result.FailedChildRefs) + len(result.CancelledChildRefs) + len(result.InterruptedChildRefs)
	if len(result.CompletionEnvelopes) > 0 && openOrTerminalIssueCount == 0 {
		for _, check := range []struct {
			ok      bool
			reason  string
			missing MissingInput
			next    NextHostAction
		}{
			{result.ParentObjectiveRunRef != "", "parent_objective_run_ref_missing", "host:parent_objective_run_ref", "provide_parent_objective_run_ref"},
			{result.ParentResumeRef != "", "parent_resume_ref_missing", "host:auto_delegation_parent_resume_ref", "provide_auto_delegation_parent_resume_ref"},
		} {
			if !check.ok {
				result = autoDelegationAsyncCompletionBlock(result, VerificationBlocked, FailureConfigMissing, check.reason, check.missing, check.next, "auto_delegation_parent_resume_ref_missing")
			}
		}
		if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
			result.Ready = true
			result.ReadyForReadback = true
			result.ReadyForResume = true
			result.Status = VerificationSatisfied
			result.NextHostAction = "resume_parent_objective_for_delegation_merge"
			result.ResumeRequest = autoDelegationAsyncCompletionResumeRequest(result)
			result.Boundaries = AppendBoundaries(result.Boundaries, "auto_delegation_parent_resume_ready", "child_completion_envelope_ready")
		}
		return result.Normalize()
	}
	if len(result.CompletionEnvelopes) > 0 || openOrTerminalIssueCount > 0 {
		result.Ready = true
		result.ReadyForReadback = true
		result.Status = VerificationPartial
		result.NextHostAction = "monitor_auto_delegation_async_children"
		if len(result.FailedChildRefs) > 0 || len(result.CancelledChildRefs) > 0 || len(result.InterruptedChildRefs) > 0 {
			result.NextHostAction = "provide_auto_delegation_parent_merge"
		}
		return result.Normalize()
	}
	return autoDelegationAsyncCompletionBlock(result, VerificationBlocked, FailureEvidenceMissing, "auto_delegation_async_no_progress_surface", "host:auto_delegation_async_child_readback", "provide_auto_delegation_runtime_progress", "auto_delegation_async_no_progress_surface").Normalize()
}

func (p AutoDelegationAsyncCompletionProjection) Normalize() AutoDelegationAsyncCompletionProjection {
	out := p
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.BackendKind = NormalizeAutoDelegationAsyncBackendKind(string(out.BackendKind))
	out.BackendRef = normalizeOneDisplaySafeRef(out.BackendRef)
	out.ProcessLocal = out.BackendKind == AutoDelegationAsyncBackendProcessLocal
	out.Durable = out.BackendKind == AutoDelegationAsyncBackendDurable
	out.UnsupportedBackend = out.BackendKind == AutoDelegationAsyncBackendUnsupported
	out.ParentObjectiveRef = normalizeOneDisplaySafeRef(out.ParentObjectiveRef)
	out.ParentObjectiveRunRef = normalizeOneDisplaySafeRef(out.ParentObjectiveRunRef)
	out.ParentLedgerRef = normalizeOneDisplaySafeRef(out.ParentLedgerRef)
	out.ParentResumeRef = normalizeOneDisplaySafeRef(out.ParentResumeRef)
	out.HostBridge = out.HostBridge.Normalize()
	for i := range out.Children {
		out.Children[i] = out.Children[i].Normalize()
	}
	out.QueuedChildRefs = normalizeDisplaySafeRefs(out.QueuedChildRefs)
	out.ActiveChildRefs = normalizeDisplaySafeRefs(out.ActiveChildRefs)
	out.CompletedChildRefs = normalizeDisplaySafeRefs(out.CompletedChildRefs)
	out.FailedChildRefs = normalizeDisplaySafeRefs(out.FailedChildRefs)
	out.CancelledChildRefs = normalizeDisplaySafeRefs(out.CancelledChildRefs)
	out.InterruptedChildRefs = normalizeDisplaySafeRefs(out.InterruptedChildRefs)
	for i := range out.CompletionEnvelopes {
		out.CompletionEnvelopes[i] = out.CompletionEnvelopes[i].Normalize()
	}
	out.ResumeRequest = out.ResumeRequest.Normalize()
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.DecisionBasis = normalizeDisplaySafeRefs(out.DecisionBasis)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RuntimeEffect == "" {
		out.RuntimeEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.Ready = false
		out.ReadyForReadback = false
		out.ReadyForResume = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status != VerificationSatisfied && out.Status != VerificationPartial {
		out.Ready = false
		out.ReadyForReadback = false
		out.ReadyForResume = false
	}
	return out
}

func (r AutoDelegationAsyncChildReadback) Normalize() AutoDelegationAsyncChildReadback {
	out := r
	out.ChildRef = normalizeOneDisplaySafeRef(out.ChildRef)
	out.Role = NormalizeAutoDelegationChildRole(string(out.Role))
	out.Status = NormalizeAutoDelegationAsyncChildStatus(string(out.Status))
	if out.AgeSeconds < 0 {
		out.AgeSeconds = 0
	}
	out.CurrentAction = normalizeControlToken(out.CurrentAction)
	out.CapabilityRefs = normalizeDisplaySafeRefs(out.CapabilityRefs)
	out.AllowedToolRefs = normalizeDisplaySafeRefs(out.AllowedToolRefs)
	out.BoundCapabilityRefs = normalizeDisplaySafeRefs(out.BoundCapabilityRefs)
	out.BoundAllowedToolRefs = normalizeDisplaySafeRefs(out.BoundAllowedToolRefs)
	out.WorkerRunRef = normalizeOneDisplaySafeRef(out.WorkerRunRef)
	out.WorkerResultRef = normalizeOneDisplaySafeRef(out.WorkerResultRef)
	out.WorkerReadbackRef = normalizeOneDisplaySafeRef(out.WorkerReadbackRef)
	out.ObservationRef = normalizeOneDisplaySafeRef(out.ObservationRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.FailureReviewRef = normalizeOneDisplaySafeRef(out.FailureReviewRef)
	out.CancellationRef = normalizeOneDisplaySafeRef(out.CancellationRef)
	out.InterruptionRef = normalizeOneDisplaySafeRef(out.InterruptionRef)
	out.CompletionEnvelopeRef = normalizeOneDisplaySafeRef(out.CompletionEnvelopeRef)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	if !out.WorkerResultRequiresVerification {
		out.WorkerResultRequiresVerification = false
	}
	if out.RawOutputLoaded {
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
	}
	return out
}

func (p AutoDelegationAsyncChildStatusProjection) Normalize() AutoDelegationAsyncChildStatusProjection {
	out := p
	out.ChildRef = normalizeOneDisplaySafeRef(out.ChildRef)
	out.Role = NormalizeAutoDelegationChildRole(string(out.Role))
	out.Status = NormalizeAutoDelegationAsyncChildStatus(string(out.Status))
	if out.AgeSeconds < 0 {
		out.AgeSeconds = 0
	}
	out.CurrentAction = normalizeControlToken(out.CurrentAction)
	out.CapabilityRefs = normalizeDisplaySafeRefs(out.CapabilityRefs)
	out.AllowedToolRefs = normalizeDisplaySafeRefs(out.AllowedToolRefs)
	out.BoundCapabilityRefs = normalizeDisplaySafeRefs(out.BoundCapabilityRefs)
	out.BoundAllowedToolRefs = normalizeDisplaySafeRefs(out.BoundAllowedToolRefs)
	out.WorkerRunRef = normalizeOneDisplaySafeRef(out.WorkerRunRef)
	out.WorkerResultRef = normalizeOneDisplaySafeRef(out.WorkerResultRef)
	out.WorkerReadbackRef = normalizeOneDisplaySafeRef(out.WorkerReadbackRef)
	out.ObservationRef = normalizeOneDisplaySafeRef(out.ObservationRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.FailureReviewRef = normalizeOneDisplaySafeRef(out.FailureReviewRef)
	out.CancellationRef = normalizeOneDisplaySafeRef(out.CancellationRef)
	out.InterruptionRef = normalizeOneDisplaySafeRef(out.InterruptionRef)
	out.CompletionEnvelopeRef = normalizeOneDisplaySafeRef(out.CompletionEnvelopeRef)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.ReadbackReady = out.WorkerReadbackRef != "" || out.CompletionEnvelopeRef != "" || len(out.EvidenceRefs) > 0
	if out.RawOutputLoaded {
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
	}
	return out
}

func (e AutoDelegationChildCompletionEnvelope) Normalize() AutoDelegationChildCompletionEnvelope {
	out := e
	out.EnvelopeRef = normalizeOneDisplaySafeRef(out.EnvelopeRef)
	out.ChildRef = normalizeOneDisplaySafeRef(out.ChildRef)
	out.Role = NormalizeAutoDelegationChildRole(string(out.Role))
	out.WorkerRunRef = normalizeOneDisplaySafeRef(out.WorkerRunRef)
	out.WorkerResultRef = normalizeOneDisplaySafeRef(out.WorkerResultRef)
	out.WorkerReadbackRef = normalizeOneDisplaySafeRef(out.WorkerReadbackRef)
	out.ObservationRef = normalizeOneDisplaySafeRef(out.ObservationRef)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.BoundCapabilityRefs = normalizeDisplaySafeRefs(out.BoundCapabilityRefs)
	out.BoundAllowedToolRefs = normalizeDisplaySafeRefs(out.BoundAllowedToolRefs)
	out.RequiresParentVerification = true
	out.WorkerOutputAcceptedAsFact = false
	return out
}

func (r AutoDelegationParentResumeRequest) Normalize() AutoDelegationParentResumeRequest {
	out := r
	out.ParentObjectiveRef = normalizeOneDisplaySafeRef(out.ParentObjectiveRef)
	out.ParentObjectiveRunRef = normalizeOneDisplaySafeRef(out.ParentObjectiveRunRef)
	out.ParentLedgerRef = normalizeOneDisplaySafeRef(out.ParentLedgerRef)
	out.ParentResumeRef = normalizeOneDisplaySafeRef(out.ParentResumeRef)
	out.ChildRefs = normalizeDisplaySafeRefs(out.ChildRefs)
	out.CompletionEnvelopeRefs = normalizeDisplaySafeRefs(out.CompletionEnvelopeRefs)
	out.WorkerResultRefs = normalizeDisplaySafeRefs(out.WorkerResultRefs)
	out.WorkerReadbackRefs = normalizeDisplaySafeRefs(out.WorkerReadbackRefs)
	return out
}

func autoDelegationAsyncNormalizeChildReadbacks(in []AutoDelegationAsyncChildReadback) []AutoDelegationAsyncChildReadback {
	out := make([]AutoDelegationAsyncChildReadback, 0, len(in))
	for _, child := range in {
		normalized := child.Normalize()
		if normalized.ChildRef == "" && !child.RawOutputLoaded {
			continue
		}
		out = append(out, normalized)
	}
	return out
}

func autoDelegationAsyncChildStatusProjection(child AutoDelegationAsyncChildReadback) AutoDelegationAsyncChildStatusProjection {
	status := AutoDelegationAsyncChildStatusProjection{
		ChildRef:                         child.ChildRef,
		Role:                             child.Role,
		Status:                           child.Status,
		AgeSeconds:                       child.AgeSeconds,
		CurrentAction:                    child.CurrentAction,
		CapabilityRefs:                   cloneDisplaySafeRefs(child.CapabilityRefs),
		AllowedToolRefs:                  cloneDisplaySafeRefs(child.AllowedToolRefs),
		BoundCapabilityRefs:              cloneDisplaySafeRefs(child.BoundCapabilityRefs),
		BoundAllowedToolRefs:             cloneDisplaySafeRefs(child.BoundAllowedToolRefs),
		WorkerRunRef:                     child.WorkerRunRef,
		WorkerResultRef:                  child.WorkerResultRef,
		WorkerReadbackRef:                child.WorkerReadbackRef,
		ObservationRef:                   child.ObservationRef,
		FailureRef:                       child.FailureRef,
		FailureReviewRef:                 child.FailureReviewRef,
		CancellationRef:                  child.CancellationRef,
		InterruptionRef:                  child.InterruptionRef,
		CompletionEnvelopeRef:            child.CompletionEnvelopeRef,
		EvidenceRefs:                     cloneEvidenceRefs(child.EvidenceRefs),
		ReadyForWorkerResultReview:       child.ReadyForWorkerResultReview,
		ReadyForFailureReview:            child.ReadyForFailureReview,
		CancelAvailable:                  child.CancelAvailable,
		InterruptAvailable:               child.InterruptAvailable,
		WorkerResultRequiresVerification: child.WorkerResultRequiresVerification,
		WorkerOutputAcceptedAsFact:       child.WorkerOutputAcceptedAsFact,
		MissingInputs:                    cloneMissingInputs(child.MissingInputs),
		BlockedReasons:                   cloneStringSlice(child.BlockedReasons),
		FailureClass:                     child.FailureClass,
		Boundaries:                       cloneBoundaries(child.Boundaries),
		RawOutputLoaded:                  child.RawOutputLoaded,
	}.Normalize()
	if status.ChildRef == "" {
		status.MissingInputs = AppendMissingInputs(status.MissingInputs, "host:auto_delegation_child_ref")
		status.BlockedReasons = appendUniqueControlToken(status.BlockedReasons, "auto_delegation_child_ref_missing")
		status.FailureClass = firstFailureClass(status.FailureClass, FailureInsufficientInformation)
	}
	if status.Status == AutoDelegationAsyncChildStatusUnknown {
		status.MissingInputs = AppendMissingInputs(status.MissingInputs, "host:auto_delegation_child_status")
		status.BlockedReasons = appendUniqueControlToken(status.BlockedReasons, "auto_delegation_child_status_missing")
		status.FailureClass = firstFailureClass(status.FailureClass, FailureInsufficientInformation)
	}
	if status.Status == AutoDelegationAsyncChildStatusCompleted {
		for _, check := range []struct {
			ok      bool
			reason  string
			missing MissingInput
		}{
			{status.WorkerRunRef != "", "worker_run_ref_missing", "host:delegation_worker_run_ref"},
			{status.WorkerResultRef != "", "worker_result_ref_missing", "host:delegation_worker_result_ref"},
			{status.WorkerReadbackRef != "", "worker_readback_ref_missing", "host:delegation_worker_readback_ref"},
			{status.CompletionEnvelopeRef != "", "completion_envelope_ref_missing", "host:auto_delegation_child_completion_envelope"},
			{len(status.EvidenceRefs) > 0, "child_evidence_refs_missing", "host:auto_delegation_child_evidence_refs"},
			{status.ReadyForWorkerResultReview, "worker_result_review_not_ready", "host:delegation_worker_result_review"},
			{status.WorkerResultRequiresVerification, "worker_result_verification_contract_missing", "contract:worker_result_requires_parent_verification"},
			{!status.WorkerOutputAcceptedAsFact, "worker_output_accepted_as_fact_rejected", "contract:child_output_not_fact"},
		} {
			if !check.ok {
				status.MissingInputs = AppendMissingInputs(status.MissingInputs, check.missing)
				status.BlockedReasons = appendUniqueControlToken(status.BlockedReasons, check.reason)
				status.FailureClass = firstFailureClass(status.FailureClass, FailureEvidenceMissing)
			}
		}
	}
	if status.Status == AutoDelegationAsyncChildStatusFailed && !status.ReadyForFailureReview {
		status.MissingInputs = AppendMissingInputs(status.MissingInputs, "host:auto_delegation_child_failure_review")
		status.BlockedReasons = appendUniqueControlToken(status.BlockedReasons, "child_failure_review_not_ready")
		status.FailureClass = firstFailureClass(status.FailureClass, FailureEvidenceMissing)
	}
	if status.CancelAvailable && status.CancellationRef == "" {
		status.MissingInputs = AppendMissingInputs(status.MissingInputs, "host:auto_delegation_child_cancellation_ref")
		status.BlockedReasons = appendUniqueControlToken(status.BlockedReasons, "child_cancellation_ref_missing")
		status.FailureClass = firstFailureClass(status.FailureClass, FailureConfigMissing)
	}
	if status.InterruptAvailable && status.InterruptionRef == "" {
		status.MissingInputs = AppendMissingInputs(status.MissingInputs, "host:auto_delegation_child_interruption_ref")
		status.BlockedReasons = appendUniqueControlToken(status.BlockedReasons, "child_interruption_ref_missing")
		status.FailureClass = firstFailureClass(status.FailureClass, FailureConfigMissing)
	}
	return status.Normalize()
}

func autoDelegationAsyncCompletionEnvelope(status AutoDelegationAsyncChildStatusProjection) (AutoDelegationChildCompletionEnvelope, bool) {
	if status.Status != AutoDelegationAsyncChildStatusCompleted ||
		status.ChildRef == "" ||
		status.WorkerRunRef == "" ||
		status.WorkerResultRef == "" ||
		status.WorkerReadbackRef == "" ||
		status.CompletionEnvelopeRef == "" ||
		len(status.EvidenceRefs) == 0 ||
		status.WorkerOutputAcceptedAsFact ||
		!status.WorkerResultRequiresVerification ||
		!status.ReadyForWorkerResultReview {
		return AutoDelegationChildCompletionEnvelope{}, false
	}
	return AutoDelegationChildCompletionEnvelope{
		EnvelopeRef:                status.CompletionEnvelopeRef,
		ChildRef:                   status.ChildRef,
		Role:                       status.Role,
		WorkerRunRef:               status.WorkerRunRef,
		WorkerResultRef:            status.WorkerResultRef,
		WorkerReadbackRef:          status.WorkerReadbackRef,
		ObservationRef:             status.ObservationRef,
		EvidenceRefs:               cloneEvidenceRefs(status.EvidenceRefs),
		BoundCapabilityRefs:        cloneDisplaySafeRefs(status.BoundCapabilityRefs),
		BoundAllowedToolRefs:       cloneDisplaySafeRefs(status.BoundAllowedToolRefs),
		RequiresParentVerification: true,
		WorkerOutputAcceptedAsFact: false,
	}.Normalize(), true
}

func autoDelegationAsyncCompletionResumeRequest(result AutoDelegationAsyncCompletionProjection) AutoDelegationParentResumeRequest {
	request := AutoDelegationParentResumeRequest{
		ParentObjectiveRef:    result.ParentObjectiveRef,
		ParentObjectiveRunRef: result.ParentObjectiveRunRef,
		ParentLedgerRef:       result.ParentLedgerRef,
		ParentResumeRef:       result.ParentResumeRef,
	}
	for _, envelope := range result.CompletionEnvelopes {
		request.ChildRefs = appendDisplaySafeRefIfPresent(request.ChildRefs, envelope.ChildRef)
		request.CompletionEnvelopeRefs = appendDisplaySafeRefIfPresent(request.CompletionEnvelopeRefs, envelope.EnvelopeRef)
		request.WorkerResultRefs = appendDisplaySafeRefIfPresent(request.WorkerResultRefs, envelope.WorkerResultRef)
		request.WorkerReadbackRefs = appendDisplaySafeRefIfPresent(request.WorkerReadbackRefs, envelope.WorkerReadbackRef)
	}
	return request.Normalize()
}

func autoDelegationAsyncCompletionBlock(result AutoDelegationAsyncCompletionProjection, status VerificationStatus, failure FailureClass, reason string, missing MissingInput, next NextHostAction, boundary Boundary) AutoDelegationAsyncCompletionProjection {
	result.Status = status
	result.Ready = false
	result.ReadyForReadback = false
	result.ReadyForResume = false
	result.FailureClass = firstFailureClass(failure, result.FailureClass)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = firstNextHostAction(next, result.NextHostAction)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result
}

func autoDelegationAsyncCompletionUnsafe(input AutoDelegationAsyncCompletionInput, children []AutoDelegationAsyncChildReadback) bool {
	if input.RawOutputLoaded ||
		autoDelegationControllerHostBridgeUnsafe(input.HostBridge) ||
		displaySafeRefRejected(input.BackendRef) ||
		displaySafeRefRejected(input.ParentObjectiveRef) ||
		displaySafeRefRejected(input.ParentObjectiveRunRef) ||
		displaySafeRefRejected(input.ParentLedgerRef) ||
		displaySafeRefRejected(input.ParentResumeRef) ||
		displaySafeRefSliceRejected(input.DecisionBasis) {
		return true
	}
	for _, child := range children {
		if autoDelegationAsyncChildReadbackUnsafe(child) {
			return true
		}
	}
	return false
}

func autoDelegationAsyncChildReadbackUnsafe(child AutoDelegationAsyncChildReadback) bool {
	return child.RawOutputLoaded ||
		displaySafeRefRejected(child.ChildRef) ||
		displaySafeRefSliceRejected(child.CapabilityRefs) ||
		displaySafeRefSliceRejected(child.AllowedToolRefs) ||
		displaySafeRefSliceRejected(child.BoundCapabilityRefs) ||
		displaySafeRefSliceRejected(child.BoundAllowedToolRefs) ||
		displaySafeRefRejected(child.WorkerRunRef) ||
		displaySafeRefRejected(child.WorkerResultRef) ||
		displaySafeRefRejected(child.WorkerReadbackRef) ||
		displaySafeRefRejected(child.ObservationRef) ||
		displaySafeRefRejected(child.FailureRef) ||
		displaySafeRefRejected(child.FailureReviewRef) ||
		displaySafeRefRejected(child.CancellationRef) ||
		displaySafeRefRejected(child.InterruptionRef) ||
		displaySafeRefRejected(child.CompletionEnvelopeRef) ||
		evidenceRefRejected(child.EvidenceRefs)
}
