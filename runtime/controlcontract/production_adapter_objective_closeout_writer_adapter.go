package controlcontract

type ProductionAdapterObjectiveCloseoutWriterHostAdapter interface {
	ObjectiveCloseoutWriterDescriptor() ProductionAdapterObjectiveCloseoutWriterDescriptor
	DryRunObjectiveCloseoutWriter(ProductionAdapterObjectiveCloseoutWriterDryRunRequest) ProductionAdapterObjectiveCloseoutWriterDryRunResult
	ExecuteObjectiveCloseoutDurableWriter(ProductionAdapterObjectiveCloseoutWriterDurableRequest) ProductionAdapterObjectiveCloseoutWriterDurableResult
}

type ProductionAdapterObjectiveCloseoutWriterDryRunRequestInput struct {
	DryRunRequestRef          DisplaySafeRef                                          `json:"dry_run_request_ref,omitempty"`
	WriterFixture             ProductionAdapterObjectiveCloseoutWriterBlackboxFixture `json:"writer_fixture,omitempty"`
	HostDryRunConfirmationRef DisplaySafeRef                                          `json:"host_dry_run_confirmation_ref,omitempty"`
	ExpectedDryRunResultRef   DisplaySafeRef                                          `json:"expected_dry_run_result_ref,omitempty"`
	ExpectedReadbackRef       DisplaySafeRef                                          `json:"expected_readback_ref,omitempty"`
	RawOutputLoaded           bool                                                    `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterDryRunRequest struct {
	ContractVersion             string                                       `json:"contract_version,omitempty"`
	Projected                   bool                                         `json:"projected"`
	Available                   bool                                         `json:"available"`
	Status                      HostActionStatus                             `json:"status,omitempty"`
	Mode                        string                                       `json:"mode,omitempty"`
	ReadyForHostDryRun          bool                                         `json:"ready_for_host_dry_run"`
	HostDryRunAuthorized        bool                                         `json:"host_dry_run_authorized"`
	CoreInvocationExecuted      bool                                         `json:"core_invocation_executed"`
	DryRunByCore                bool                                         `json:"dry_run_by_core"`
	DurableWriteByCore          bool                                         `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore   bool                                         `json:"objective_store_write_by_core"`
	RunstoreWriteByCore         bool                                         `json:"runstore_write_by_core"`
	RequestedMode               ProductionAdapterObjectiveCloseoutWriterMode `json:"requested_mode,omitempty"`
	DryRunRequestRef            DisplaySafeRef                               `json:"dry_run_request_ref,omitempty"`
	WriterFixtureRef            DisplaySafeRef                               `json:"writer_fixture_ref,omitempty"`
	WriterOptInRef              DisplaySafeRef                               `json:"writer_opt_in_ref,omitempty"`
	WriterRef                   DisplaySafeRef                               `json:"writer_ref,omitempty"`
	OwnerRef                    DisplaySafeRef                               `json:"owner_ref,omitempty"`
	HostWriterBindingRef        DisplaySafeRef                               `json:"host_writer_binding_ref,omitempty"`
	HostDryRunConfirmationRef   DisplaySafeRef                               `json:"host_dry_run_confirmation_ref,omitempty"`
	ObjectiveCloseoutHandoffRef DisplaySafeRef                               `json:"objective_closeout_handoff_ref,omitempty"`
	HostUIHandoffRef            DisplaySafeRef                               `json:"host_ui_handoff_ref,omitempty"`
	ObjectiveCloseoutPacketRef  DisplaySafeRef                               `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                DisplaySafeRef                               `json:"objective_ref,omitempty"`
	HostObjectiveLifecycleRef   DisplaySafeRef                               `json:"host_objective_lifecycle_ref,omitempty"`
	HostRunstoreRef             DisplaySafeRef                               `json:"host_runstore_ref,omitempty"`
	DryRunPlanRef               DisplaySafeRef                               `json:"dry_run_plan_ref,omitempty"`
	ExpectedDryRunResultRef     DisplaySafeRef                               `json:"expected_dry_run_result_ref,omitempty"`
	ExpectedReadbackRef         DisplaySafeRef                               `json:"expected_readback_ref,omitempty"`
	DryRunContractRef           DisplaySafeRef                               `json:"dry_run_contract_ref,omitempty"`
	ReadbackContractRef         DisplaySafeRef                               `json:"readback_contract_ref,omitempty"`
	IdempotencyRef              DisplaySafeRef                               `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef      DisplaySafeRef                               `json:"idempotency_contract_ref,omitempty"`
	RedactionPolicyRef          DisplaySafeRef                               `json:"redaction_policy_ref,omitempty"`
	TimeoutPolicyRef            DisplaySafeRef                               `json:"timeout_policy_ref,omitempty"`
	AvailableCapabilityRefs     []DisplaySafeRef                             `json:"available_capability_refs,omitempty"`
	RequiredCapabilityRefs      []DisplaySafeRef                             `json:"required_capability_refs,omitempty"`
	PolicyRefs                  []DisplaySafeRef                             `json:"policy_refs,omitempty"`
	RequiredPolicyRefs          []DisplaySafeRef                             `json:"required_policy_refs,omitempty"`
	ApprovalRefs                []DisplaySafeRef                             `json:"approval_refs,omitempty"`
	RequiredApprovalRefs        []DisplaySafeRef                             `json:"required_approval_refs,omitempty"`
	BudgetRef                   DisplaySafeRef                               `json:"budget_ref,omitempty"`
	RequiredBudgetRef           DisplaySafeRef                               `json:"required_budget_ref,omitempty"`
	RollbackReviewRef           DisplaySafeRef                               `json:"rollback_review_ref,omitempty"`
	CompensationReviewRef       DisplaySafeRef                               `json:"compensation_review_ref,omitempty"`
	MissingInputs               []MissingInput                               `json:"missing_inputs,omitempty"`
	BlockedReasons              []string                                     `json:"blocked_reasons,omitempty"`
	FailureClass                FailureClass                                 `json:"failure_class,omitempty"`
	Boundaries                  []Boundary                                   `json:"boundaries,omitempty"`
	NextHostAction              NextHostAction                               `json:"next_host_action,omitempty"`
	RunnerEffect                string                                       `json:"runner_effect,omitempty"`
	PromptEffect                string                                       `json:"prompt_effect,omitempty"`
	RawOutputLoaded             bool                                         `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterDryRunResultInput struct {
	DryRunResultRef     DisplaySafeRef                                        `json:"dry_run_result_ref,omitempty"`
	DryRunRequest       ProductionAdapterObjectiveCloseoutWriterDryRunRequest `json:"dry_run_request,omitempty"`
	HostAdapterRunRef   DisplaySafeRef                                        `json:"host_adapter_run_ref,omitempty"`
	HostDryRunReported  bool                                                  `json:"host_dry_run_reported"`
	HostDryRunSucceeded bool                                                  `json:"host_dry_run_succeeded"`
	ExpectedReadbackRef DisplaySafeRef                                        `json:"expected_readback_ref,omitempty"`
	DryRunEvidenceRefs  []DisplaySafeRef                                      `json:"dry_run_evidence_refs,omitempty"`
	FailureRef          DisplaySafeRef                                        `json:"failure_ref,omitempty"`
	RawOutputLoaded     bool                                                  `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterDryRunResult struct {
	ContractVersion             string           `json:"contract_version,omitempty"`
	Projected                   bool             `json:"projected"`
	Available                   bool             `json:"available"`
	Status                      HostActionStatus `json:"status,omitempty"`
	Mode                        string           `json:"mode,omitempty"`
	ReadyForDurableWriteOptIn   bool             `json:"ready_for_durable_write_opt_in"`
	HostDryRunReported          bool             `json:"host_dry_run_reported"`
	HostDryRunSucceeded         bool             `json:"host_dry_run_succeeded"`
	HostDryRunRecorded          bool             `json:"host_dry_run_recorded"`
	CoreInvocationExecuted      bool             `json:"core_invocation_executed"`
	DryRunByCore                bool             `json:"dry_run_by_core"`
	DurableWriteByCore          bool             `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore   bool             `json:"objective_store_write_by_core"`
	RunstoreWriteByCore         bool             `json:"runstore_write_by_core"`
	DryRunResultRef             DisplaySafeRef   `json:"dry_run_result_ref,omitempty"`
	ExpectedDryRunResultRef     DisplaySafeRef   `json:"expected_dry_run_result_ref,omitempty"`
	DryRunRequestRef            DisplaySafeRef   `json:"dry_run_request_ref,omitempty"`
	WriterFixtureRef            DisplaySafeRef   `json:"writer_fixture_ref,omitempty"`
	WriterOptInRef              DisplaySafeRef   `json:"writer_opt_in_ref,omitempty"`
	WriterRef                   DisplaySafeRef   `json:"writer_ref,omitempty"`
	HostWriterBindingRef        DisplaySafeRef   `json:"host_writer_binding_ref,omitempty"`
	HostAdapterRunRef           DisplaySafeRef   `json:"host_adapter_run_ref,omitempty"`
	ObjectiveCloseoutHandoffRef DisplaySafeRef   `json:"objective_closeout_handoff_ref,omitempty"`
	HostUIHandoffRef            DisplaySafeRef   `json:"host_ui_handoff_ref,omitempty"`
	ObjectiveCloseoutPacketRef  DisplaySafeRef   `json:"objective_closeout_packet_ref,omitempty"`
	ObjectiveRef                DisplaySafeRef   `json:"objective_ref,omitempty"`
	HostRunstoreRef             DisplaySafeRef   `json:"host_runstore_ref,omitempty"`
	DryRunPlanRef               DisplaySafeRef   `json:"dry_run_plan_ref,omitempty"`
	ExpectedReadbackRef         DisplaySafeRef   `json:"expected_readback_ref,omitempty"`
	DryRunContractRef           DisplaySafeRef   `json:"dry_run_contract_ref,omitempty"`
	ReadbackContractRef         DisplaySafeRef   `json:"readback_contract_ref,omitempty"`
	IdempotencyRef              DisplaySafeRef   `json:"idempotency_ref,omitempty"`
	IdempotencyContractRef      DisplaySafeRef   `json:"idempotency_contract_ref,omitempty"`
	RedactionPolicyRef          DisplaySafeRef   `json:"redaction_policy_ref,omitempty"`
	TimeoutPolicyRef            DisplaySafeRef   `json:"timeout_policy_ref,omitempty"`
	DryRunEvidenceRefs          []DisplaySafeRef `json:"dry_run_evidence_refs,omitempty"`
	FailureRef                  DisplaySafeRef   `json:"failure_ref,omitempty"`
	MissingInputs               []MissingInput   `json:"missing_inputs,omitempty"`
	BlockedReasons              []string         `json:"blocked_reasons,omitempty"`
	FailureClass                FailureClass     `json:"failure_class,omitempty"`
	Boundaries                  []Boundary       `json:"boundaries,omitempty"`
	NextHostAction              NextHostAction   `json:"next_host_action,omitempty"`
	RunnerEffect                string           `json:"runner_effect,omitempty"`
	PromptEffect                string           `json:"prompt_effect,omitempty"`
	RawOutputLoaded             bool             `json:"raw_output_loaded"`
}

type ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessInput struct {
	SmokeRef        DisplaySafeRef                                        `json:"smoke_ref,omitempty"`
	DryRunRequest   ProductionAdapterObjectiveCloseoutWriterDryRunRequest `json:"dry_run_request,omitempty"`
	DryRunResult    ProductionAdapterObjectiveCloseoutWriterDryRunResult  `json:"dry_run_result,omitempty"`
	RawOutputLoaded bool                                                  `json:"raw_output_loaded"`
}

// agentx-api: internal_candidate
type ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness struct {
	ContractVersion           string         `json:"contract_version,omitempty"`
	Projected                 bool           `json:"projected"`
	Available                 bool           `json:"available"`
	Status                    string         `json:"status,omitempty"`
	Mode                      string         `json:"mode,omitempty"`
	ReadyForHostDisplay       bool           `json:"ready_for_host_display"`
	SmokePassed               bool           `json:"smoke_passed"`
	ReadyForDurableWriteOptIn bool           `json:"ready_for_durable_write_opt_in"`
	CoreInvocationExecuted    bool           `json:"core_invocation_executed"`
	DryRunByCore              bool           `json:"dry_run_by_core"`
	DurableWriteByCore        bool           `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore bool           `json:"objective_store_write_by_core"`
	RunstoreWriteByCore       bool           `json:"runstore_write_by_core"`
	SmokeRef                  DisplaySafeRef `json:"smoke_ref,omitempty"`
	DryRunRequestRef          DisplaySafeRef `json:"dry_run_request_ref,omitempty"`
	DryRunResultRef           DisplaySafeRef `json:"dry_run_result_ref,omitempty"`
	ExpectedDryRunResultRef   DisplaySafeRef `json:"expected_dry_run_result_ref,omitempty"`
	ExpectedReadbackRef       DisplaySafeRef `json:"expected_readback_ref,omitempty"`
	WriterFixtureRef          DisplaySafeRef `json:"writer_fixture_ref,omitempty"`
	WriterOptInRef            DisplaySafeRef `json:"writer_opt_in_ref,omitempty"`
	WriterRef                 DisplaySafeRef `json:"writer_ref,omitempty"`
	HostAdapterRunRef         DisplaySafeRef `json:"host_adapter_run_ref,omitempty"`
	MissingInputs             []MissingInput `json:"missing_inputs,omitempty"`
	BlockedReasons            []string       `json:"blocked_reasons,omitempty"`
	FailureClass              FailureClass   `json:"failure_class,omitempty"`
	Boundaries                []Boundary     `json:"boundaries,omitempty"`
	NextHostAction            NextHostAction `json:"next_host_action,omitempty"`
	RunnerEffect              string         `json:"runner_effect,omitempty"`
	PromptEffect              string         `json:"prompt_effect,omitempty"`
	RawOutputLoaded           bool           `json:"raw_output_loaded"`
}

func BuildProductionAdapterObjectiveCloseoutWriterDryRunRequest(input ProductionAdapterObjectiveCloseoutWriterDryRunRequestInput) ProductionAdapterObjectiveCloseoutWriterDryRunRequest {
	if productionAdapterObjectiveCloseoutWriterBlackboxFixtureEmpty(input.WriterFixture) {
		return unavailableProductionAdapterObjectiveCloseoutWriterDryRunRequest()
	}
	fixture := input.WriterFixture.Normalize()
	result := ProductionAdapterObjectiveCloseoutWriterDryRunRequest{
		ContractVersion:             ContractVersion,
		Projected:                   true,
		Available:                   fixture.Available,
		Status:                      HostActionBlocked,
		Mode:                        "production_adapter_objective_closeout_writer_dry_run_request",
		RequestedMode:               fixture.RequestedMode,
		DryRunRequestRef:            normalizeOneDisplaySafeRef(input.DryRunRequestRef),
		WriterFixtureRef:            fixture.FixtureRef,
		WriterOptInRef:              fixture.WriterOptInRef,
		WriterRef:                   fixture.WriterRef,
		OwnerRef:                    fixture.OwnerRef,
		HostWriterBindingRef:        fixture.HostWriterBindingRef,
		HostDryRunConfirmationRef:   normalizeOneDisplaySafeRef(input.HostDryRunConfirmationRef),
		ObjectiveCloseoutHandoffRef: fixture.ObjectiveCloseoutHandoffRef,
		HostUIHandoffRef:            fixture.HostUIHandoffRef,
		ObjectiveCloseoutPacketRef:  fixture.ObjectiveCloseoutPacketRef,
		ObjectiveRef:                fixture.ObjectiveRef,
		HostObjectiveLifecycleRef:   fixture.HostObjectiveLifecycleRef,
		HostRunstoreRef:             fixture.HostRunstoreRef,
		DryRunPlanRef:               fixture.DryRunPlanRef,
		ExpectedDryRunResultRef:     normalizeOneDisplaySafeRef(input.ExpectedDryRunResultRef),
		ExpectedReadbackRef:         normalizeOneDisplaySafeRef(input.ExpectedReadbackRef),
		DryRunContractRef:           fixture.DryRunContractRef,
		ReadbackContractRef:         fixture.ReadbackContractRef,
		IdempotencyRef:              fixture.IdempotencyRef,
		IdempotencyContractRef:      fixture.IdempotencyContractRef,
		RedactionPolicyRef:          fixture.RedactionPolicyRef,
		TimeoutPolicyRef:            fixture.TimeoutPolicyRef,
		AvailableCapabilityRefs:     cloneDisplaySafeRefs(fixture.AvailableCapabilityRefs),
		RequiredCapabilityRefs:      cloneDisplaySafeRefs(fixture.RequiredCapabilityRefs),
		PolicyRefs:                  cloneDisplaySafeRefs(fixture.PolicyRefs),
		RequiredPolicyRefs:          cloneDisplaySafeRefs(fixture.RequiredPolicyRefs),
		ApprovalRefs:                cloneDisplaySafeRefs(fixture.ApprovalRefs),
		RequiredApprovalRefs:        cloneDisplaySafeRefs(fixture.RequiredApprovalRefs),
		BudgetRef:                   fixture.BudgetRef,
		RequiredBudgetRef:           fixture.RequiredBudgetRef,
		RollbackReviewRef:           fixture.RollbackReviewRef,
		CompensationReviewRef:       fixture.CompensationReviewRef,
		FailureClass:                FailureNone,
		Boundaries:                  productionAdapterObjectiveCloseoutWriterDryRunRequestBoundaries(fixture.Boundaries),
		RunnerEffect:                "none",
		PromptEffect:                "none",
		RawOutputLoaded:             input.RawOutputLoaded || fixture.RawOutputLoaded,
	}
	if productionAdapterObjectiveCloseoutWriterDryRunRequestUnsafe(input, fixture) {
		result = productionAdapterObjectiveCloseoutWriterDryRunRequestBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !fixture.ReadyForWriterDryRunDisplay {
		result = productionAdapterObjectiveCloseoutWriterDryRunRequestBlock(result, firstFailureClass(fixture.FailureClass, FailureConfigMissing), "objective_closeout_writer_dry_run_display_not_ready", "host:objective_closeout_writer_dry_run_display", firstNextHostAction(fixture.NextHostAction, "review_objective_closeout_writer_dry_run_display"))
	}
	if result.DryRunRequestRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDryRunRequestBlock(result, FailureEvidenceMissing, "writer_dry_run_request_ref_missing", "host:objective_closeout_writer_dry_run_request_ref", "provide_objective_closeout_writer_dry_run_request")
	}
	if result.HostDryRunConfirmationRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDryRunRequestBlock(result, FailureAuthorizationMissing, "host_dry_run_confirmation_ref_missing", "host:objective_closeout_writer_dry_run_confirmation_ref", "request_objective_closeout_writer_dry_run_confirmation")
	}
	if result.ExpectedDryRunResultRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDryRunRequestBlock(result, FailureEvidenceMissing, "expected_dry_run_result_ref_missing", "host:objective_closeout_writer_expected_dry_run_result_ref", "provide_objective_closeout_writer_expected_dry_run_result")
	}
	if result.ExpectedReadbackRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDryRunRequestBlock(result, FailureEvidenceMissing, "expected_readback_ref_missing", "host:objective_closeout_writer_expected_readback_ref", "provide_objective_closeout_writer_expected_readback")
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostDryRun = true
		result.HostDryRunAuthorized = true
		result.NextHostAction = "host_may_run_objective_closeout_writer_dry_run_adapter"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_objective_closeout_writer_dry_run", "objective_closeout_writer_dry_run_only", "durable_write_not_enabled")
	}
	return result.Normalize()
}

func BuildProductionAdapterObjectiveCloseoutWriterDryRunResult(input ProductionAdapterObjectiveCloseoutWriterDryRunResultInput) ProductionAdapterObjectiveCloseoutWriterDryRunResult {
	if productionAdapterObjectiveCloseoutWriterDryRunRequestEmpty(input.DryRunRequest) {
		return unavailableProductionAdapterObjectiveCloseoutWriterDryRunResult()
	}
	request := input.DryRunRequest.Normalize()
	result := ProductionAdapterObjectiveCloseoutWriterDryRunResult{
		ContractVersion:             ContractVersion,
		Projected:                   true,
		Available:                   request.Available,
		Status:                      HostActionBlocked,
		Mode:                        "production_adapter_objective_closeout_writer_dry_run_result",
		HostDryRunReported:          input.HostDryRunReported,
		HostDryRunSucceeded:         input.HostDryRunSucceeded,
		DryRunResultRef:             normalizeOneDisplaySafeRef(input.DryRunResultRef),
		ExpectedDryRunResultRef:     request.ExpectedDryRunResultRef,
		DryRunRequestRef:            request.DryRunRequestRef,
		WriterFixtureRef:            request.WriterFixtureRef,
		WriterOptInRef:              request.WriterOptInRef,
		WriterRef:                   request.WriterRef,
		HostWriterBindingRef:        request.HostWriterBindingRef,
		HostAdapterRunRef:           normalizeOneDisplaySafeRef(input.HostAdapterRunRef),
		ObjectiveCloseoutHandoffRef: request.ObjectiveCloseoutHandoffRef,
		HostUIHandoffRef:            request.HostUIHandoffRef,
		ObjectiveCloseoutPacketRef:  request.ObjectiveCloseoutPacketRef,
		ObjectiveRef:                request.ObjectiveRef,
		HostRunstoreRef:             request.HostRunstoreRef,
		DryRunPlanRef:               request.DryRunPlanRef,
		ExpectedReadbackRef:         normalizeOneDisplaySafeRef(input.ExpectedReadbackRef),
		DryRunContractRef:           request.DryRunContractRef,
		ReadbackContractRef:         request.ReadbackContractRef,
		IdempotencyRef:              request.IdempotencyRef,
		IdempotencyContractRef:      request.IdempotencyContractRef,
		RedactionPolicyRef:          request.RedactionPolicyRef,
		TimeoutPolicyRef:            request.TimeoutPolicyRef,
		DryRunEvidenceRefs:          normalizeDisplaySafeRefs(input.DryRunEvidenceRefs),
		FailureRef:                  normalizeOneDisplaySafeRef(input.FailureRef),
		FailureClass:                FailureNone,
		Boundaries:                  productionAdapterObjectiveCloseoutWriterDryRunResultBoundaries(request.Boundaries),
		RunnerEffect:                "none",
		PromptEffect:                "none",
		RawOutputLoaded:             input.RawOutputLoaded || request.RawOutputLoaded,
	}
	if productionAdapterObjectiveCloseoutWriterDryRunResultUnsafe(input, request) {
		result = productionAdapterObjectiveCloseoutWriterDryRunResultBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !request.ReadyForHostDryRun {
		result = productionAdapterObjectiveCloseoutWriterDryRunResultBlock(result, firstFailureClass(request.FailureClass, FailureConfigMissing), "objective_closeout_writer_dry_run_request_not_ready", "host:objective_closeout_writer_dry_run_request", firstNextHostAction(request.NextHostAction, "review_objective_closeout_writer_dry_run_request"))
	}
	if !input.HostDryRunReported {
		result = productionAdapterObjectiveCloseoutWriterDryRunResultBlock(result, FailureEvidenceMissing, "writer_dry_run_not_reported", "host:objective_closeout_writer_dry_run_report", "provide_objective_closeout_writer_dry_run_report")
	}
	if result.HostAdapterRunRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDryRunResultBlock(result, FailureEvidenceMissing, "host_adapter_run_ref_missing", "host:objective_closeout_writer_host_adapter_run_ref", "provide_objective_closeout_writer_dry_run_report")
	}
	if result.DryRunResultRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDryRunResultBlock(result, FailureEvidenceMissing, "dry_run_result_ref_missing", "host:objective_closeout_writer_dry_run_result_ref", "provide_objective_closeout_writer_dry_run_result")
	} else if request.ExpectedDryRunResultRef != "" && result.DryRunResultRef != request.ExpectedDryRunResultRef {
		result = productionAdapterObjectiveCloseoutWriterDryRunResultBlock(result, FailureVerificationFailed, "dry_run_result_ref_mismatch", "host:objective_closeout_writer_dry_run_result_ref", "review_objective_closeout_writer_dry_run_result")
	}
	if result.ExpectedReadbackRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDryRunResultBlock(result, FailureEvidenceMissing, "expected_readback_ref_missing", "host:objective_closeout_writer_expected_readback_ref", "provide_objective_closeout_writer_expected_readback")
	} else if request.ExpectedReadbackRef != "" && result.ExpectedReadbackRef != request.ExpectedReadbackRef {
		result = productionAdapterObjectiveCloseoutWriterDryRunResultBlock(result, FailureVerificationFailed, "expected_readback_ref_mismatch", "host:objective_closeout_writer_expected_readback_ref", "review_objective_closeout_writer_dry_run_result")
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.HostDryRunRecorded = true
		if input.HostDryRunSucceeded {
			result.Status = HostActionRecorded
			result.ReadyForDurableWriteOptIn = true
			result.NextHostAction = "review_objective_closeout_writer_durable_opt_in"
			result.Boundaries = AppendBoundaries(result.Boundaries, "host_objective_closeout_writer_dry_run_recorded", "ready_for_objective_closeout_writer_durable_opt_in", "durable_write_not_enabled")
		} else {
			result.Status = HostActionReviewRequired
			result.FailureClass = firstFailureClass(result.FailureClass, FailureVerificationFailed)
			result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, "writer_dry_run_failed")
			result.NextHostAction = "review_objective_closeout_writer_dry_run_failure"
			result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_dry_run_failed", "durable_write_not_enabled")
		}
	}
	return result.Normalize()
}

func BuildProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness(input ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessInput) ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness {
	if productionAdapterObjectiveCloseoutWriterDryRunRequestEmpty(input.DryRunRequest) || productionAdapterObjectiveCloseoutWriterDryRunResultEmpty(input.DryRunResult) {
		return unavailableProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness()
	}
	request := input.DryRunRequest.Normalize()
	dryRunResult := input.DryRunResult.Normalize()
	result := ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness{
		ContractVersion:         ContractVersion,
		Projected:               true,
		Available:               request.Available && dryRunResult.Available,
		Status:                  "blocked",
		Mode:                    "production_adapter_objective_closeout_writer_dry_run_smoke_harness",
		SmokeRef:                normalizeOneDisplaySafeRef(input.SmokeRef),
		DryRunRequestRef:        request.DryRunRequestRef,
		DryRunResultRef:         dryRunResult.DryRunResultRef,
		ExpectedDryRunResultRef: request.ExpectedDryRunResultRef,
		ExpectedReadbackRef:     request.ExpectedReadbackRef,
		WriterFixtureRef:        request.WriterFixtureRef,
		WriterOptInRef:          request.WriterOptInRef,
		WriterRef:               request.WriterRef,
		HostAdapterRunRef:       dryRunResult.HostAdapterRunRef,
		FailureClass:            FailureNone,
		Boundaries:              productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessBoundaries(request.Boundaries, dryRunResult.Boundaries),
		RunnerEffect:            "none",
		PromptEffect:            "none",
		RawOutputLoaded:         input.RawOutputLoaded || request.RawOutputLoaded || dryRunResult.RawOutputLoaded,
	}
	if productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessUnsafe(input, request, dryRunResult) {
		result = productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if result.SmokeRef == "" {
		result = productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessBlock(result, FailureEvidenceMissing, "dry_run_smoke_ref_missing", "host:objective_closeout_writer_dry_run_smoke_ref", "provide_objective_closeout_writer_dry_run_smoke")
	}
	if !request.ReadyForHostDryRun {
		result = productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessBlock(result, firstFailureClass(request.FailureClass, FailureConfigMissing), "dry_run_request_not_ready", "host:objective_closeout_writer_dry_run_request", "review_objective_closeout_writer_dry_run_request")
	}
	if !dryRunResult.ReadyForDurableWriteOptIn {
		result = productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessBlock(result, firstFailureClass(dryRunResult.FailureClass, FailureVerificationFailed), "dry_run_result_not_ready", "host:objective_closeout_writer_dry_run_result", "review_objective_closeout_writer_dry_run_result")
	}
	for _, mismatch := range productionAdapterObjectiveCloseoutWriterDryRunSmokeMismatches(request, dryRunResult) {
		result = productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessBlock(result, FailureVerificationFailed, mismatch.reason, mismatch.missing, "review_objective_closeout_writer_dry_run_smoke")
	}
	if len(result.BlockedReasons) == 0 && len(result.MissingInputs) == 0 {
		result.Status = "dry_run_smoke_passed"
		result.ReadyForHostDisplay = true
		result.SmokePassed = true
		result.ReadyForDurableWriteOptIn = true
		result.NextHostAction = "review_objective_closeout_writer_durable_opt_in"
		result.Boundaries = AppendBoundaries(result.Boundaries, "dry_run_smoke_passed", "ready_for_objective_closeout_writer_durable_opt_in", "durable_write_not_enabled")
	}
	return result.Normalize()
}

func CloneProductionAdapterObjectiveCloseoutWriterDryRunRequest(in ProductionAdapterObjectiveCloseoutWriterDryRunRequest) ProductionAdapterObjectiveCloseoutWriterDryRunRequest {
	out := in
	out.AvailableCapabilityRefs = cloneDisplaySafeRefs(in.AvailableCapabilityRefs)
	out.RequiredCapabilityRefs = cloneDisplaySafeRefs(in.RequiredCapabilityRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.RequiredPolicyRefs = cloneDisplaySafeRefs(in.RequiredPolicyRefs)
	out.ApprovalRefs = cloneDisplaySafeRefs(in.ApprovalRefs)
	out.RequiredApprovalRefs = cloneDisplaySafeRefs(in.RequiredApprovalRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ProductionAdapterObjectiveCloseoutWriterDryRunRequest) Clone() ProductionAdapterObjectiveCloseoutWriterDryRunRequest {
	return CloneProductionAdapterObjectiveCloseoutWriterDryRunRequest(r)
}

func (r ProductionAdapterObjectiveCloseoutWriterDryRunRequest) Normalize() ProductionAdapterObjectiveCloseoutWriterDryRunRequest {
	out := CloneProductionAdapterObjectiveCloseoutWriterDryRunRequest(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_dry_run_request"
	}
	out.RequestedMode = NormalizeProductionAdapterObjectiveCloseoutWriterMode(string(out.RequestedMode))
	out.DryRunRequestRef = normalizeOneDisplaySafeRef(out.DryRunRequestRef)
	out.WriterFixtureRef = normalizeOneDisplaySafeRef(out.WriterFixtureRef)
	out.WriterOptInRef = normalizeOneDisplaySafeRef(out.WriterOptInRef)
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.OwnerRef = normalizeOneDisplaySafeRef(out.OwnerRef)
	out.HostWriterBindingRef = normalizeOneDisplaySafeRef(out.HostWriterBindingRef)
	out.HostDryRunConfirmationRef = normalizeOneDisplaySafeRef(out.HostDryRunConfirmationRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.HostUIHandoffRef = normalizeOneDisplaySafeRef(out.HostUIHandoffRef)
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.HostObjectiveLifecycleRef = normalizeOneDisplaySafeRef(out.HostObjectiveLifecycleRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.DryRunPlanRef = normalizeOneDisplaySafeRef(out.DryRunPlanRef)
	out.ExpectedDryRunResultRef = normalizeOneDisplaySafeRef(out.ExpectedDryRunResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.DryRunContractRef = normalizeOneDisplaySafeRef(out.DryRunContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.RedactionPolicyRef = normalizeOneDisplaySafeRef(out.RedactionPolicyRef)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
	out.AvailableCapabilityRefs = normalizeDisplaySafeRefs(out.AvailableCapabilityRefs)
	out.RequiredCapabilityRefs = normalizeDisplaySafeRefs(out.RequiredCapabilityRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.RequiredPolicyRefs = normalizeDisplaySafeRefs(out.RequiredPolicyRefs)
	out.ApprovalRefs = normalizeDisplaySafeRefs(out.ApprovalRefs)
	out.RequiredApprovalRefs = normalizeDisplaySafeRefs(out.RequiredApprovalRefs)
	out.BudgetRef = normalizeOneDisplaySafeRef(out.BudgetRef)
	out.RequiredBudgetRef = normalizeOneDisplaySafeRef(out.RequiredBudgetRef)
	out.RollbackReviewRef = normalizeOneDisplaySafeRef(out.RollbackReviewRef)
	out.CompensationReviewRef = normalizeOneDisplaySafeRef(out.CompensationReviewRef)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if !out.Available {
		out.Status = HostActionNotReady
		out.ReadyForHostDryRun = false
		out.HostDryRunAuthorized = false
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterDryRunRequestOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForHostDryRun = false
		out.HostDryRunAuthorized = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DryRunByCore = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.ReadyForHostDryRun = out.ReadyForHostDryRun &&
		out.Status == HostActionReady &&
		out.RequestedMode == ProductionAdapterObjectiveCloseoutWriterDryRun &&
		out.DryRunRequestRef != "" &&
		out.WriterFixtureRef != "" &&
		out.WriterOptInRef != "" &&
		out.WriterRef != "" &&
		out.HostWriterBindingRef != "" &&
		out.HostDryRunConfirmationRef != "" &&
		out.DryRunPlanRef != "" &&
		out.ExpectedDryRunResultRef != "" &&
		out.ExpectedReadbackRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.HostDryRunAuthorized = out.HostDryRunAuthorized && out.ReadyForHostDryRun
	return out
}

func CloneProductionAdapterObjectiveCloseoutWriterDryRunResult(in ProductionAdapterObjectiveCloseoutWriterDryRunResult) ProductionAdapterObjectiveCloseoutWriterDryRunResult {
	out := in
	out.DryRunEvidenceRefs = cloneDisplaySafeRefs(in.DryRunEvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ProductionAdapterObjectiveCloseoutWriterDryRunResult) Clone() ProductionAdapterObjectiveCloseoutWriterDryRunResult {
	return CloneProductionAdapterObjectiveCloseoutWriterDryRunResult(r)
}

func (r ProductionAdapterObjectiveCloseoutWriterDryRunResult) Normalize() ProductionAdapterObjectiveCloseoutWriterDryRunResult {
	out := CloneProductionAdapterObjectiveCloseoutWriterDryRunResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_dry_run_result"
	}
	out.DryRunResultRef = normalizeOneDisplaySafeRef(out.DryRunResultRef)
	out.ExpectedDryRunResultRef = normalizeOneDisplaySafeRef(out.ExpectedDryRunResultRef)
	out.DryRunRequestRef = normalizeOneDisplaySafeRef(out.DryRunRequestRef)
	out.WriterFixtureRef = normalizeOneDisplaySafeRef(out.WriterFixtureRef)
	out.WriterOptInRef = normalizeOneDisplaySafeRef(out.WriterOptInRef)
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.HostWriterBindingRef = normalizeOneDisplaySafeRef(out.HostWriterBindingRef)
	out.HostAdapterRunRef = normalizeOneDisplaySafeRef(out.HostAdapterRunRef)
	out.ObjectiveCloseoutHandoffRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutHandoffRef)
	out.HostUIHandoffRef = normalizeOneDisplaySafeRef(out.HostUIHandoffRef)
	out.ObjectiveCloseoutPacketRef = normalizeOneDisplaySafeRef(out.ObjectiveCloseoutPacketRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.DryRunPlanRef = normalizeOneDisplaySafeRef(out.DryRunPlanRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.DryRunContractRef = normalizeOneDisplaySafeRef(out.DryRunContractRef)
	out.ReadbackContractRef = normalizeOneDisplaySafeRef(out.ReadbackContractRef)
	out.IdempotencyRef = normalizeOneDisplaySafeRef(out.IdempotencyRef)
	out.IdempotencyContractRef = normalizeOneDisplaySafeRef(out.IdempotencyContractRef)
	out.RedactionPolicyRef = normalizeOneDisplaySafeRef(out.RedactionPolicyRef)
	out.TimeoutPolicyRef = normalizeOneDisplaySafeRef(out.TimeoutPolicyRef)
	out.DryRunEvidenceRefs = normalizeDisplaySafeRefs(out.DryRunEvidenceRefs)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if !out.Available {
		out.Status = HostActionNotReady
		out.ReadyForDurableWriteOptIn = false
		out.HostDryRunRecorded = false
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterDryRunResultOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForDurableWriteOptIn = false
		out.HostDryRunRecorded = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DryRunByCore = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.HostDryRunRecorded = out.HostDryRunRecorded &&
		(out.Status == HostActionRecorded || out.Status == HostActionReviewRequired) &&
		out.HostDryRunReported &&
		out.DryRunRequestRef != "" &&
		out.DryRunResultRef != "" &&
		out.HostAdapterRunRef != "" &&
		!out.RawOutputLoaded
	out.ReadyForDurableWriteOptIn = out.ReadyForDurableWriteOptIn &&
		out.Status == HostActionRecorded &&
		out.HostDryRunRecorded &&
		out.HostDryRunSucceeded &&
		out.DryRunResultRef != "" &&
		out.ExpectedReadbackRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	return out
}

func CloneProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness(in ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness) ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness {
	out := in
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (h ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness) Clone() ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness {
	return CloneProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness(h)
}

func (h ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness) Normalize() ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness {
	out := CloneProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness(h)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = normalizeControlToken(out.Status)
	if out.Status == "" {
		out.Status = "blocked"
	}
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "production_adapter_objective_closeout_writer_dry_run_smoke_harness"
	}
	out.SmokeRef = normalizeOneDisplaySafeRef(out.SmokeRef)
	out.DryRunRequestRef = normalizeOneDisplaySafeRef(out.DryRunRequestRef)
	out.DryRunResultRef = normalizeOneDisplaySafeRef(out.DryRunResultRef)
	out.ExpectedDryRunResultRef = normalizeOneDisplaySafeRef(out.ExpectedDryRunResultRef)
	out.ExpectedReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedReadbackRef)
	out.WriterFixtureRef = normalizeOneDisplaySafeRef(out.WriterFixtureRef)
	out.WriterOptInRef = normalizeOneDisplaySafeRef(out.WriterOptInRef)
	out.WriterRef = normalizeOneDisplaySafeRef(out.WriterRef)
	out.HostAdapterRunRef = normalizeOneDisplaySafeRef(out.HostAdapterRunRef)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeControlTokenList(out.BlockedReasons)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if !out.Available {
		out.Status = "unavailable"
		out.ReadyForHostDisplay = false
		out.SmokePassed = false
		out.ReadyForDurableWriteOptIn = false
	}
	if out.RawOutputLoaded || productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = "blocked"
		out.ReadyForHostDisplay = false
		out.SmokePassed = false
		out.ReadyForDurableWriteOptIn = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DryRunByCore = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.SmokePassed = out.SmokePassed &&
		out.Available &&
		out.Status == "dry_run_smoke_passed" &&
		out.SmokeRef != "" &&
		out.DryRunRequestRef != "" &&
		out.DryRunResultRef != "" &&
		out.ExpectedDryRunResultRef != "" &&
		out.ExpectedReadbackRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.ReadyForHostDisplay = out.ReadyForHostDisplay && out.SmokePassed
	out.ReadyForDurableWriteOptIn = out.ReadyForDurableWriteOptIn && out.SmokePassed
	return out
}

func productionAdapterObjectiveCloseoutWriterDryRunRequestBlock(result ProductionAdapterObjectiveCloseoutWriterDryRunRequest, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterDryRunRequest {
	result.Status = HostActionBlocked
	result.ReadyForHostDryRun = false
	result.HostDryRunAuthorized = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_dry_run_request_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterObjectiveCloseoutWriterDryRunResultBlock(result ProductionAdapterObjectiveCloseoutWriterDryRunResult, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterDryRunResult {
	result.Status = HostActionBlocked
	result.ReadyForDurableWriteOptIn = false
	result.HostDryRunRecorded = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_dry_run_result_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessBlock(result ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness {
	result.Status = "blocked"
	result.ReadyForHostDisplay = false
	result.SmokePassed = false
	result.ReadyForDurableWriteOptIn = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_closeout_writer_dry_run_smoke_harness_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

type productionAdapterObjectiveCloseoutWriterDryRunMismatch struct {
	reason  string
	missing MissingInput
}

func productionAdapterObjectiveCloseoutWriterDryRunSmokeMismatches(request ProductionAdapterObjectiveCloseoutWriterDryRunRequest, result ProductionAdapterObjectiveCloseoutWriterDryRunResult) []productionAdapterObjectiveCloseoutWriterDryRunMismatch {
	var out []productionAdapterObjectiveCloseoutWriterDryRunMismatch
	out = append(out, productionAdapterObjectiveCloseoutWriterDryRunRefMismatch(request.DryRunRequestRef, result.DryRunRequestRef, "dry_run_request_ref_mismatch", "host:objective_closeout_writer_dry_run_request_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDryRunRefMismatch(request.WriterFixtureRef, result.WriterFixtureRef, "dry_run_writer_fixture_ref_mismatch", "host:objective_closeout_writer_fixture_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDryRunRefMismatch(request.WriterOptInRef, result.WriterOptInRef, "dry_run_writer_opt_in_ref_mismatch", "host:objective_closeout_writer_opt_in_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDryRunRefMismatch(request.WriterRef, result.WriterRef, "dry_run_writer_ref_mismatch", "host:objective_closeout_writer_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDryRunRefMismatch(request.ExpectedDryRunResultRef, result.DryRunResultRef, "dry_run_result_ref_mismatch", "host:objective_closeout_writer_dry_run_result_ref")...)
	out = append(out, productionAdapterObjectiveCloseoutWriterDryRunRefMismatch(request.ExpectedReadbackRef, result.ExpectedReadbackRef, "dry_run_expected_readback_ref_mismatch", "host:objective_closeout_writer_expected_readback_ref")...)
	return out
}

func productionAdapterObjectiveCloseoutWriterDryRunRefMismatch(left DisplaySafeRef, right DisplaySafeRef, reason string, missing MissingInput) []productionAdapterObjectiveCloseoutWriterDryRunMismatch {
	left = normalizeOneDisplaySafeRef(left)
	right = normalizeOneDisplaySafeRef(right)
	if left != "" && right != "" && left != right {
		return []productionAdapterObjectiveCloseoutWriterDryRunMismatch{{reason: reason, missing: missing}}
	}
	return nil
}

func productionAdapterObjectiveCloseoutWriterDryRunRequestBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_dry_run_request",
			"objective_closeout_writer_dry_run_request_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"objective_closeout_writer_dry_run_only",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterObjectiveCloseoutWriterDryRunResultBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_dry_run_result",
			"objective_closeout_writer_dry_run_result_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"objective_closeout_writer_dry_run_only",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"production_adapter_objective_closeout_writer_dry_run_smoke_harness",
			"objective_closeout_writer_dry_run_smoke_harness_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"objective_closeout_writer_dry_run_only",
			"display_safe_refs_only",
			"display_safe_result_refs_only",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func productionAdapterObjectiveCloseoutWriterDryRunRequestUnsafe(input ProductionAdapterObjectiveCloseoutWriterDryRunRequestInput, fixture ProductionAdapterObjectiveCloseoutWriterBlackboxFixture) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.DryRunRequestRef) ||
		displaySafeRefRejected(input.HostDryRunConfirmationRef) ||
		displaySafeRefRejected(input.ExpectedDryRunResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		productionAdapterObjectiveCloseoutWriterBlackboxFixtureUnsafeOutput(fixture)
}

func productionAdapterObjectiveCloseoutWriterDryRunRequestOutputUnsafe(input ProductionAdapterObjectiveCloseoutWriterDryRunRequest) bool {
	return displaySafeRefRejected(input.DryRunRequestRef) ||
		displaySafeRefRejected(input.WriterFixtureRef) ||
		displaySafeRefRejected(input.WriterOptInRef) ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.OwnerRef) ||
		displaySafeRefRejected(input.HostWriterBindingRef) ||
		displaySafeRefRejected(input.HostDryRunConfirmationRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.HostUIHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostObjectiveLifecycleRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.DryRunPlanRef) ||
		displaySafeRefRejected(input.ExpectedDryRunResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.DryRunContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.RedactionPolicyRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		displaySafeRefSliceRejected(input.AvailableCapabilityRefs) ||
		displaySafeRefSliceRejected(input.RequiredCapabilityRefs) ||
		displaySafeRefSliceRejected(input.PolicyRefs) ||
		displaySafeRefSliceRejected(input.RequiredPolicyRefs) ||
		displaySafeRefSliceRejected(input.ApprovalRefs) ||
		displaySafeRefSliceRejected(input.RequiredApprovalRefs) ||
		displaySafeRefRejected(input.BudgetRef) ||
		displaySafeRefRejected(input.RequiredBudgetRef) ||
		displaySafeRefRejected(input.RollbackReviewRef) ||
		displaySafeRefRejected(input.CompensationReviewRef) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterDryRunResultUnsafe(input ProductionAdapterObjectiveCloseoutWriterDryRunResultInput, request ProductionAdapterObjectiveCloseoutWriterDryRunRequest) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.DryRunResultRef) ||
		displaySafeRefRejected(input.HostAdapterRunRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefSliceRejected(input.DryRunEvidenceRefs) ||
		displaySafeRefRejected(input.FailureRef) ||
		productionAdapterObjectiveCloseoutWriterDryRunRequestOutputUnsafe(request)
}

func productionAdapterObjectiveCloseoutWriterDryRunResultOutputUnsafe(input ProductionAdapterObjectiveCloseoutWriterDryRunResult) bool {
	return displaySafeRefRejected(input.DryRunResultRef) ||
		displaySafeRefRejected(input.ExpectedDryRunResultRef) ||
		displaySafeRefRejected(input.DryRunRequestRef) ||
		displaySafeRefRejected(input.WriterFixtureRef) ||
		displaySafeRefRejected(input.WriterOptInRef) ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.HostWriterBindingRef) ||
		displaySafeRefRejected(input.HostAdapterRunRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutHandoffRef) ||
		displaySafeRefRejected(input.HostUIHandoffRef) ||
		displaySafeRefRejected(input.ObjectiveCloseoutPacketRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.DryRunPlanRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.DryRunContractRef) ||
		displaySafeRefRejected(input.ReadbackContractRef) ||
		displaySafeRefRejected(input.IdempotencyRef) ||
		displaySafeRefRejected(input.IdempotencyContractRef) ||
		displaySafeRefRejected(input.RedactionPolicyRef) ||
		displaySafeRefRejected(input.TimeoutPolicyRef) ||
		displaySafeRefSliceRejected(input.DryRunEvidenceRefs) ||
		displaySafeRefRejected(input.FailureRef) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessUnsafe(input ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessInput, request ProductionAdapterObjectiveCloseoutWriterDryRunRequest, result ProductionAdapterObjectiveCloseoutWriterDryRunResult) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.SmokeRef) ||
		productionAdapterObjectiveCloseoutWriterDryRunRequestOutputUnsafe(request) ||
		productionAdapterObjectiveCloseoutWriterDryRunResultOutputUnsafe(result)
}

func productionAdapterObjectiveCloseoutWriterDryRunSmokeHarnessOutputUnsafe(input ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness) bool {
	return displaySafeRefRejected(input.SmokeRef) ||
		displaySafeRefRejected(input.DryRunRequestRef) ||
		displaySafeRefRejected(input.DryRunResultRef) ||
		displaySafeRefRejected(input.ExpectedDryRunResultRef) ||
		displaySafeRefRejected(input.ExpectedReadbackRef) ||
		displaySafeRefRejected(input.WriterFixtureRef) ||
		displaySafeRefRejected(input.WriterOptInRef) ||
		displaySafeRefRejected(input.WriterRef) ||
		displaySafeRefRejected(input.HostAdapterRunRef) ||
		input.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterBlackboxFixtureEmpty(fixture ProductionAdapterObjectiveCloseoutWriterBlackboxFixture) bool {
	return !fixture.Projected &&
		!fixture.Available &&
		fixture.Status == "" &&
		fixture.Mode == "" &&
		fixture.FixtureRef == "" &&
		fixture.WriterOptInRef == "" &&
		fixture.WriterRef == "" &&
		len(fixture.MissingInputs) == 0 &&
		len(fixture.BlockedReasons) == 0 &&
		len(fixture.Boundaries) == 0 &&
		fixture.NextHostAction == "" &&
		!fixture.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterDryRunRequestEmpty(request ProductionAdapterObjectiveCloseoutWriterDryRunRequest) bool {
	return !request.Projected &&
		!request.Available &&
		request.Status == "" &&
		request.Mode == "" &&
		request.DryRunRequestRef == "" &&
		request.WriterFixtureRef == "" &&
		request.WriterOptInRef == "" &&
		len(request.MissingInputs) == 0 &&
		len(request.BlockedReasons) == 0 &&
		len(request.Boundaries) == 0 &&
		request.NextHostAction == "" &&
		!request.RawOutputLoaded
}

func productionAdapterObjectiveCloseoutWriterDryRunResultEmpty(result ProductionAdapterObjectiveCloseoutWriterDryRunResult) bool {
	return !result.Projected &&
		!result.Available &&
		result.Status == "" &&
		result.Mode == "" &&
		result.DryRunResultRef == "" &&
		result.DryRunRequestRef == "" &&
		result.WriterOptInRef == "" &&
		len(result.MissingInputs) == 0 &&
		len(result.BlockedReasons) == 0 &&
		len(result.Boundaries) == 0 &&
		result.NextHostAction == "" &&
		!result.RawOutputLoaded
}

func unavailableProductionAdapterObjectiveCloseoutWriterDryRunRequest() ProductionAdapterObjectiveCloseoutWriterDryRunRequest {
	return ProductionAdapterObjectiveCloseoutWriterDryRunRequest{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "production_adapter_objective_closeout_writer_dry_run_request",
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_dry_run_request",
			"objective_closeout_writer_dry_run_request_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"objective_closeout_writer_dry_run_only",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_objective_closeout_writer_fixture",
	}
}

func unavailableProductionAdapterObjectiveCloseoutWriterDryRunResult() ProductionAdapterObjectiveCloseoutWriterDryRunResult {
	return ProductionAdapterObjectiveCloseoutWriterDryRunResult{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "production_adapter_objective_closeout_writer_dry_run_result",
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_dry_run_result",
			"objective_closeout_writer_dry_run_result_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"objective_closeout_writer_dry_run_only",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_objective_closeout_writer_dry_run_request",
	}
}

func unavailableProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness() ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness {
	return ProductionAdapterObjectiveCloseoutWriterDryRunSmokeHarness{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          "unavailable",
		Mode:            "production_adapter_objective_closeout_writer_dry_run_smoke_harness",
		Boundaries: []Boundary{
			"production_adapter_objective_closeout_writer_dry_run_smoke_harness",
			"objective_closeout_writer_dry_run_smoke_harness_projection_only",
			"host_owned_objective_closeout_writer_adapter",
			"objective_closeout_writer_dry_run_only",
			"no_runner_dispatch",
			"no_dry_run_by_core",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		RunnerEffect:   "none",
		PromptEffect:   "none",
		NextHostAction: "provide_objective_closeout_writer_dry_run_result",
	}
}
