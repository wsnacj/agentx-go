package controlcontract

type ObjectiveRuntimeProductizationInput struct {
	Activation                 Activation               `json:"activation,omitempty"`
	RuntimeLoop                ObjectiveRuntimeLoopStep `json:"runtime_loop,omitempty"`
	ObjectiveRunRef            DisplaySafeRef           `json:"objective_run_ref,omitempty"`
	TaskLedgerRef              DisplaySafeRef           `json:"task_ledger_ref,omitempty"`
	TrajectoryRef              DisplaySafeRef           `json:"trajectory_ref,omitempty"`
	WatchdogRef                DisplaySafeRef           `json:"watchdog_ref,omitempty"`
	HostRuntimeQueueRef        DisplaySafeRef           `json:"host_runtime_queue_ref,omitempty"`
	NoProgressAttemptThreshold int                      `json:"no_progress_attempt_threshold,omitempty"`
	EvidenceRefs               []EvidenceRef            `json:"evidence_refs,omitempty"`
	Boundaries                 []Boundary               `json:"boundaries,omitempty"`
	RawOutputLoaded            bool                     `json:"raw_output_loaded"`
}

type ObjectiveRuntimeProductizationReport struct {
	ContractVersion             string                    `json:"contract_version,omitempty"`
	Projected                   bool                      `json:"projected"`
	Available                   bool                      `json:"available"`
	Status                      string                    `json:"status,omitempty"`
	Activation                  Activation                `json:"activation,omitempty"`
	RuntimeLoopStatus           string                    `json:"runtime_loop_status,omitempty"`
	ObjectiveRunState           ObjectiveControllerState  `json:"objective_run_state,omitempty"`
	ControllerAction            ObjectiveControllerAction `json:"controller_action,omitempty"`
	RuntimeLoopReadyForPersist  bool                      `json:"runtime_loop_ready_for_persist"`
	RuntimeLoopReadyForContinue bool                      `json:"runtime_loop_ready_for_continue"`
	TaskLedgerReady             bool                      `json:"task_ledger_ready"`
	TrajectoryReady             bool                      `json:"trajectory_ready"`
	WatchdogReady               bool                      `json:"watchdog_ready"`
	ReadyForHostProductization  bool                      `json:"ready_for_host_productization"`
	ReadyForRuntimeContinuation bool                      `json:"ready_for_runtime_continuation"`
	ReadyForWatchdogStop        bool                      `json:"ready_for_watchdog_stop"`
	NoProgressDetected          bool                      `json:"no_progress_detected"`
	NoProgressAttemptCount      int                       `json:"no_progress_attempt_count,omitempty"`
	NoProgressAttemptThreshold  int                       `json:"no_progress_attempt_threshold,omitempty"`
	AttemptCount                int                       `json:"attempt_count,omitempty"`
	ObjectiveID                 string                    `json:"objective_id,omitempty"`
	CurrentStrategyRef          DisplaySafeRef            `json:"current_strategy_ref,omitempty"`
	ObjectiveRunRef             DisplaySafeRef            `json:"objective_run_ref,omitempty"`
	TaskLedgerRef               DisplaySafeRef            `json:"task_ledger_ref,omitempty"`
	TrajectoryRef               DisplaySafeRef            `json:"trajectory_ref,omitempty"`
	WatchdogRef                 DisplaySafeRef            `json:"watchdog_ref,omitempty"`
	HostRuntimeQueueRef         DisplaySafeRef            `json:"host_runtime_queue_ref,omitempty"`
	EvidenceRefs                []EvidenceRef             `json:"evidence_refs,omitempty"`
	FailureClass                FailureClass              `json:"failure_class,omitempty"`
	MissingInputs               []MissingInput            `json:"missing_inputs,omitempty"`
	BlockedReasons              []string                  `json:"blocked_reasons,omitempty"`
	Boundaries                  []Boundary                `json:"boundaries,omitempty"`
	NextHostAction              NextHostAction            `json:"next_host_action,omitempty"`
	RunnerEffect                string                    `json:"runner_effect,omitempty"`
	PromptEffect                string                    `json:"prompt_effect,omitempty"`
	RuntimeEffect               string                    `json:"runtime_effect,omitempty"`
	RawOutputLoaded             bool                      `json:"raw_output_loaded"`
}

func BuildObjectiveRuntimeProductization(input ObjectiveRuntimeProductizationInput) ObjectiveRuntimeProductizationReport {
	runtimeLoop := input.RuntimeLoop.Normalize()
	run := runtimeLoop.Run.Normalize()
	activation := NormalizeActivation(string(input.Activation))
	if activation == ActivationOff && run.Activation != ActivationOff {
		activation = run.Activation
	}
	threshold := input.NoProgressAttemptThreshold
	if threshold <= 0 {
		threshold = 2
	}
	report := ObjectiveRuntimeProductizationReport{
		ContractVersion:             ContractVersion,
		Projected:                   true,
		Available:                   true,
		Status:                      "blocked",
		Activation:                  activation,
		RuntimeLoopStatus:           runtimeLoop.Status,
		ObjectiveRunState:           run.State,
		ControllerAction:            runtimeLoop.ControllerDecision.Action,
		RuntimeLoopReadyForPersist:  runtimeLoop.ReadyForHostPersist,
		RuntimeLoopReadyForContinue: runtimeLoop.ReadyForNextRuntimeAction,
		NoProgressAttemptThreshold:  threshold,
		AttemptCount:                len(run.Ledger.Attempts),
		ObjectiveID:                 run.Frame.ID,
		CurrentStrategyRef:          run.CurrentStrategyRef,
		ObjectiveRunRef:             normalizeOneDisplaySafeRef(input.ObjectiveRunRef),
		TaskLedgerRef:               normalizeOneDisplaySafeRef(input.TaskLedgerRef),
		TrajectoryRef:               normalizeOneDisplaySafeRef(input.TrajectoryRef),
		WatchdogRef:                 normalizeOneDisplaySafeRef(input.WatchdogRef),
		HostRuntimeQueueRef:         normalizeOneDisplaySafeRef(input.HostRuntimeQueueRef),
		EvidenceRefs:                MergeEvidenceRefs(input.EvidenceRefs, runtimeLoop.EvidenceRefs, run.EvidenceRefs),
		FailureClass:                firstFailureClass(runtimeLoop.FailureClass, run.FailureClass),
		Boundaries: MergeBoundaries([]Boundary{
			"objective_runtime_productization",
			"host_owned_task_ledger_projection",
			"host_owned_trajectory_projection",
			"host_owned_no_progress_watchdog",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_runtime_adapter_execution",
			"no_tool_execution",
			"no_workflow_dispatch",
			"no_scheduler_apply",
			"no_install_apply",
			"no_store_mutation_by_core",
			"projection_only",
		}, input.Boundaries, runtimeLoop.Boundaries),
		NextHostAction:  "review_objective_runtime_productization",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RuntimeEffect:   "none",
		RawOutputLoaded: input.RawOutputLoaded || runtimeLoop.RawOutputLoaded || run.RawOutputLoaded,
	}
	if objectiveRuntimeProductizationUnsafe(input, runtimeLoop, run) {
		report.RawOutputLoaded = true
		report.Status = "review_required"
		report.FailureClass = FailureEvidenceWeak
		report.MissingInputs = AppendMissingInputs(report.MissingInputs, "host:display_safe_refs")
		report.BlockedReasons = append(report.BlockedReasons, "unsafe_input_ref")
		report.Boundaries = AppendBoundaries(report.Boundaries, "raw_output_not_allowed")
		report.NextHostAction = "provide_display_safe_refs"
		return report.Normalize()
	}
	if activation != ActivationManaged {
		report.Available = false
		report.Status = "inactive"
		report.FailureClass = FailurePolicyBlocked
		report.MissingInputs = AppendMissingInputs(report.MissingInputs, "host:managed_control_plane_activation")
		report.BlockedReasons = append(report.BlockedReasons, "objective_runtime_productization_requires_managed_activation")
		report.Boundaries = AppendBoundaries(report.Boundaries, "objective_runtime_productization_default_off")
		report.NextHostAction = "enable_managed_objective"
		return report.Normalize()
	}
	for _, check := range []struct {
		ok      bool
		missing MissingInput
		reason  string
	}{
		{!objectiveRuntimeProductizationLoopEmpty(runtimeLoop), "host:objective_runtime_loop", "objective_runtime_loop_missing"},
		{runtimeLoop.ReadyForHostPersist, "host:objective_runtime_loop_host_persist", "objective_runtime_loop_not_ready_for_host_persist"},
		{report.ObjectiveRunRef != "", "host:objective_run_ref", "objective_run_ref_missing"},
		{report.TaskLedgerRef != "", "host:task_ledger_ref", "task_ledger_ref_missing"},
		{report.TrajectoryRef != "", "host:trajectory_ref", "trajectory_ref_missing"},
		{report.WatchdogRef != "", "host:watchdog_ref", "watchdog_ref_missing"},
	} {
		if !check.ok {
			report.MissingInputs = AppendMissingInputs(report.MissingInputs, check.missing)
			report.BlockedReasons = append(report.BlockedReasons, check.reason)
		}
	}
	report.TaskLedgerReady = report.ObjectiveRunRef != "" && report.TaskLedgerRef != "" && report.AttemptCount > 0
	report.TrajectoryReady = report.TrajectoryRef != "" && len(report.EvidenceRefs) > 0
	noProgress, count := objectiveRuntimeProductizationNoProgress(run.Ledger.Attempts, run.CurrentStrategyRef, runtimeLoop.Verification.Normalize(), threshold)
	report.NoProgressDetected = noProgress
	report.NoProgressAttemptCount = count
	report.WatchdogReady = report.WatchdogRef != "" && report.AttemptCount > 0
	if report.NoProgressDetected {
		report.ReadyForWatchdogStop = report.WatchdogReady
		report.Status = "watchdog_blocked_repeated_no_progress"
		report.FailureClass = FailureRepeatedNoProgress
		report.BlockedReasons = append(report.BlockedReasons, "repeated_no_progress_detected")
		report.Boundaries = AppendBoundaries(report.Boundaries, "objective_runtime_productization_no_progress_watchdog_stop")
		report.NextHostAction = "return_blocked"
		return report.Normalize()
	}
	if len(report.MissingInputs) == 0 && len(report.BlockedReasons) == 0 && report.TaskLedgerReady && report.TrajectoryReady && report.WatchdogReady {
		report.ReadyForHostProductization = true
		report.ReadyForRuntimeContinuation = runtimeLoop.ReadyForNextRuntimeAction
		report.Status = "ready_for_host_runtime_productization"
		report.NextHostAction = "continue_objective_runtime_loop"
		report.Boundaries = AppendBoundaries(report.Boundaries,
			"objective_runtime_productization_task_ledger_ready",
			"objective_runtime_productization_trajectory_ready",
			"objective_runtime_productization_watchdog_ready",
		)
		return report.Normalize()
	}
	report.Status = "blocked"
	if report.FailureClass == FailureNone {
		report.FailureClass = FailureEvidenceMissing
	}
	return report.Normalize()
}

func (report ObjectiveRuntimeProductizationReport) Normalize() ObjectiveRuntimeProductizationReport {
	out := report
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	out.Activation = NormalizeActivation(string(out.Activation))
	out.RuntimeLoopStatus = normalizeControlToken(out.RuntimeLoopStatus)
	out.ObjectiveRunState = NormalizeObjectiveControllerState(string(out.ObjectiveRunState))
	out.ControllerAction = NormalizeObjectiveControllerAction(string(out.ControllerAction))
	out.ObjectiveID = normalizeControlToken(out.ObjectiveID)
	out.CurrentStrategyRef = normalizeOneDisplaySafeRef(out.CurrentStrategyRef)
	out.ObjectiveRunRef = normalizeOneDisplaySafeRef(out.ObjectiveRunRef)
	out.TaskLedgerRef = normalizeOneDisplaySafeRef(out.TaskLedgerRef)
	out.TrajectoryRef = normalizeOneDisplaySafeRef(out.TrajectoryRef)
	out.WatchdogRef = normalizeOneDisplaySafeRef(out.WatchdogRef)
	out.HostRuntimeQueueRef = normalizeOneDisplaySafeRef(out.HostRuntimeQueueRef)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeStringList(out.BlockedReasons)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	out.RuntimeEffect = normalizeControlToken(out.RuntimeEffect)
	if out.Status == "" {
		out.Status = "blocked"
	}
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
		out.Status = "review_required"
		out.ReadyForHostProductization = false
		out.ReadyForRuntimeContinuation = false
		out.ReadyForWatchdogStop = false
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = normalizeStringList(append(out.BlockedReasons, "unsafe_input_ref"))
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	out.TaskLedgerReady = out.TaskLedgerReady && out.ObjectiveRunRef != "" && out.TaskLedgerRef != "" && out.AttemptCount > 0 && !out.RawOutputLoaded
	out.TrajectoryReady = out.TrajectoryReady && out.TrajectoryRef != "" && len(out.EvidenceRefs) > 0 && !out.RawOutputLoaded
	out.WatchdogReady = out.WatchdogReady && out.WatchdogRef != "" && out.AttemptCount > 0 && !out.RawOutputLoaded
	out.ReadyForHostProductization = out.ReadyForHostProductization &&
		out.Available &&
		out.Activation == ActivationManaged &&
		out.RuntimeLoopReadyForPersist &&
		out.TaskLedgerReady &&
		out.TrajectoryReady &&
		out.WatchdogReady &&
		len(out.MissingInputs) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForRuntimeContinuation = out.ReadyForRuntimeContinuation && out.ReadyForHostProductization && out.RuntimeLoopReadyForContinue
	out.ReadyForWatchdogStop = out.ReadyForWatchdogStop &&
		out.NoProgressDetected &&
		out.WatchdogReady &&
		!out.RawOutputLoaded
	return out
}

func objectiveRuntimeProductizationUnsafe(input ObjectiveRuntimeProductizationInput, runtimeLoop ObjectiveRuntimeLoopStep, run ObjectiveRun) bool {
	return input.RawOutputLoaded ||
		runtimeLoop.RawOutputLoaded ||
		run.RawOutputLoaded ||
		displaySafeRefRejected(input.ObjectiveRunRef) ||
		displaySafeRefRejected(input.TaskLedgerRef) ||
		displaySafeRefRejected(input.TrajectoryRef) ||
		displaySafeRefRejected(input.WatchdogRef) ||
		displaySafeRefRejected(input.HostRuntimeQueueRef) ||
		evidenceRefRejected(input.EvidenceRefs)
}

func objectiveRuntimeProductizationLoopEmpty(loop ObjectiveRuntimeLoopStep) bool {
	return !loop.Projected &&
		!loop.Available &&
		loop.Status == "" &&
		loop.Run.Frame.ID == "" &&
		loop.Run.Ledger.LedgerRef == "" &&
		len(loop.Run.Ledger.Attempts) == 0
}

func objectiveRuntimeProductizationNoProgress(attempts []AttemptSummary, current DisplaySafeRef, verification ObjectiveVerificationGateResult, threshold int) (bool, int) {
	if threshold <= 0 {
		threshold = 2
	}
	verification = verification.Normalize()
	if verification.FailureClass == FailureRepeatedNoProgress {
		return true, threshold
	}
	current = normalizeOneDisplaySafeRef(current)
	if current == "" {
		return false, 0
	}
	count := 0
	for _, attempt := range normalizeAttemptSummaries(attempts) {
		if attempt.StrategyID != string(current) {
			continue
		}
		if attempt.Status == VerificationSatisfied {
			continue
		}
		if len(attempt.EvidenceRefs) > 0 || attempt.ObservationCount > 0 {
			continue
		}
		count++
		if attempt.FailureClass == FailureRepeatedNoProgress || count >= threshold {
			return true, count
		}
	}
	return false, count
}
