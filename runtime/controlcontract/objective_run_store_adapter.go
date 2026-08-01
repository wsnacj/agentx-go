package controlcontract

type ObjectiveRunStoreDurableRequestInput struct {
	RuntimeLoop                      ObjectiveRuntimeLoopStep                 `json:"runtime_loop,omitempty"`
	StoreMutationDescriptor          ProductionAdapterStoreMutationDescriptor `json:"store_mutation_descriptor,omitempty"`
	ObjectiveRunStoreRequestRef      DisplaySafeRef                           `json:"objective_run_store_request_ref,omitempty"`
	ObjectiveRunRef                  DisplaySafeRef                           `json:"objective_run_ref,omitempty"`
	LedgerRef                        DisplaySafeRef                           `json:"ledger_ref,omitempty"`
	VerificationRef                  DisplaySafeRef                           `json:"verification_ref,omitempty"`
	ReplannerDecisionRef             DisplaySafeRef                           `json:"replanner_decision_ref,omitempty"`
	ExpectedDurableEventRef          DisplaySafeRef                           `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef        DisplaySafeRef                           `json:"expected_objective_state_ref,omitempty"`
	HostStoreMutationConfirmationRef DisplaySafeRef                           `json:"host_store_mutation_confirmation_ref,omitempty"`
	ExpectedStoreMutationResultRef   DisplaySafeRef                           `json:"expected_store_mutation_result_ref,omitempty"`
	ExpectedStoreTransactionRef      DisplaySafeRef                           `json:"expected_store_transaction_ref,omitempty"`
	ExpectedStoreReadbackRef         DisplaySafeRef                           `json:"expected_store_readback_ref,omitempty"`
	AdditionalEvidenceRefs           []EvidenceRef                            `json:"additional_evidence_refs,omitempty"`
	Boundaries                       []Boundary                               `json:"boundaries,omitempty"`
	RawOutputLoaded                  bool                                     `json:"raw_output_loaded"`
}

type ObjectiveRunStoreDurableRequest struct {
	ContractVersion                  string                                `json:"contract_version,omitempty"`
	Projected                        bool                                  `json:"projected"`
	Available                        bool                                  `json:"available"`
	Status                           HostActionStatus                      `json:"status,omitempty"`
	Mode                             string                                `json:"mode,omitempty"`
	ReadyForHostObjectiveRunStore    bool                                  `json:"ready_for_host_objective_run_store"`
	HostObjectiveRunStoreAuthorized  bool                                  `json:"host_objective_run_store_authorized"`
	HostMayPersistObjectiveRun       bool                                  `json:"host_may_persist_objective_run"`
	RuntimeLoop                      ObjectiveRuntimeLoopStep              `json:"runtime_loop,omitempty"`
	StoreMutationRequest             ProductionAdapterStoreMutationRequest `json:"store_mutation_request,omitempty"`
	ObjectiveRunStoreRequestRef      DisplaySafeRef                        `json:"objective_run_store_request_ref,omitempty"`
	ObjectiveRunRef                  DisplaySafeRef                        `json:"objective_run_ref,omitempty"`
	ObjectiveRef                     DisplaySafeRef                        `json:"objective_ref,omitempty"`
	LedgerRef                        DisplaySafeRef                        `json:"ledger_ref,omitempty"`
	VerificationRef                  DisplaySafeRef                        `json:"verification_ref,omitempty"`
	ReplannerDecisionRef             DisplaySafeRef                        `json:"replanner_decision_ref,omitempty"`
	StoreMutationRequestRef          DisplaySafeRef                        `json:"store_mutation_request_ref,omitempty"`
	StoreAdapterRef                  DisplaySafeRef                        `json:"store_adapter_ref,omitempty"`
	HostRunstoreRef                  DisplaySafeRef                        `json:"host_runstore_ref,omitempty"`
	HostObjectiveStoreRef            DisplaySafeRef                        `json:"host_objective_store_ref,omitempty"`
	ExpectedDurableEventRef          DisplaySafeRef                        `json:"expected_durable_event_ref,omitempty"`
	ExpectedObjectiveStateRef        DisplaySafeRef                        `json:"expected_objective_state_ref,omitempty"`
	HostStoreMutationConfirmationRef DisplaySafeRef                        `json:"host_store_mutation_confirmation_ref,omitempty"`
	ExpectedStoreMutationResultRef   DisplaySafeRef                        `json:"expected_store_mutation_result_ref,omitempty"`
	ExpectedStoreTransactionRef      DisplaySafeRef                        `json:"expected_store_transaction_ref,omitempty"`
	ExpectedStoreReadbackRef         DisplaySafeRef                        `json:"expected_store_readback_ref,omitempty"`
	EvidenceRefs                     []EvidenceRef                         `json:"evidence_refs,omitempty"`
	MissingInputs                    []MissingInput                        `json:"missing_inputs,omitempty"`
	BlockedReasons                   []string                              `json:"blocked_reasons,omitempty"`
	FailureClass                     FailureClass                          `json:"failure_class,omitempty"`
	Boundaries                       []Boundary                            `json:"boundaries,omitempty"`
	NextHostAction                   NextHostAction                        `json:"next_host_action,omitempty"`
	RunnerEffect                     string                                `json:"runner_effect,omitempty"`
	PromptEffect                     string                                `json:"prompt_effect,omitempty"`
	CoreInvocationExecuted           bool                                  `json:"core_invocation_executed"`
	DurableWriteByCore               bool                                  `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore        bool                                  `json:"objective_store_write_by_core"`
	RunstoreWriteByCore              bool                                  `json:"runstore_write_by_core"`
	RawOutputLoaded                  bool                                  `json:"raw_output_loaded"`
}

type ObjectiveRunStoreDurableResultInput struct {
	ObjectiveRunStoreResultRef     DisplaySafeRef                  `json:"objective_run_store_result_ref,omitempty"`
	DurableRequest                 ObjectiveRunStoreDurableRequest `json:"durable_request,omitempty"`
	HostStoreAdapterRunRef         DisplaySafeRef                  `json:"host_store_adapter_run_ref,omitempty"`
	HostObjectiveRunStoreReported  bool                            `json:"host_objective_run_store_reported"`
	HostObjectiveRunStoreSucceeded bool                            `json:"host_objective_run_store_succeeded"`
	HostObjectiveRunStoreFailed    bool                            `json:"host_objective_run_store_failed"`
	AppliedTransactionRef          DisplaySafeRef                  `json:"applied_transaction_ref,omitempty"`
	AppliedRunstoreRef             DisplaySafeRef                  `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef       DisplaySafeRef                  `json:"applied_objective_state_ref,omitempty"`
	FailureRef                     DisplaySafeRef                  `json:"failure_ref,omitempty"`
	CompensationRef                DisplaySafeRef                  `json:"compensation_ref,omitempty"`
	StoreMutationEvidenceRefs      []DisplaySafeRef                `json:"store_mutation_evidence_refs,omitempty"`
	RawOutputLoaded                bool                            `json:"raw_output_loaded"`
}

type ObjectiveRunStoreDurableResult struct {
	ContractVersion                   string                               `json:"contract_version,omitempty"`
	Projected                         bool                                 `json:"projected"`
	Available                         bool                                 `json:"available"`
	Status                            HostActionStatus                     `json:"status,omitempty"`
	Mode                              string                               `json:"mode,omitempty"`
	ReadyForObjectiveRunStoreReadback bool                                 `json:"ready_for_objective_run_store_readback"`
	HostObjectiveRunStoreReported     bool                                 `json:"host_objective_run_store_reported"`
	HostObjectiveRunStoreSucceeded    bool                                 `json:"host_objective_run_store_succeeded"`
	HostObjectiveRunStoreFailed       bool                                 `json:"host_objective_run_store_failed"`
	HostObjectiveRunStoreRecorded     bool                                 `json:"host_objective_run_store_recorded"`
	DurableRequest                    ObjectiveRunStoreDurableRequest      `json:"durable_request,omitempty"`
	StoreMutationResult               ProductionAdapterStoreMutationResult `json:"store_mutation_result,omitempty"`
	ObjectiveRunStoreResultRef        DisplaySafeRef                       `json:"objective_run_store_result_ref,omitempty"`
	ObjectiveRunStoreRequestRef       DisplaySafeRef                       `json:"objective_run_store_request_ref,omitempty"`
	ObjectiveRunRef                   DisplaySafeRef                       `json:"objective_run_ref,omitempty"`
	ObjectiveRef                      DisplaySafeRef                       `json:"objective_ref,omitempty"`
	LedgerRef                         DisplaySafeRef                       `json:"ledger_ref,omitempty"`
	VerificationRef                   DisplaySafeRef                       `json:"verification_ref,omitempty"`
	ReplannerDecisionRef              DisplaySafeRef                       `json:"replanner_decision_ref,omitempty"`
	StoreMutationResultRef            DisplaySafeRef                       `json:"store_mutation_result_ref,omitempty"`
	StoreMutationRequestRef           DisplaySafeRef                       `json:"store_mutation_request_ref,omitempty"`
	StoreAdapterRef                   DisplaySafeRef                       `json:"store_adapter_ref,omitempty"`
	HostStoreAdapterRunRef            DisplaySafeRef                       `json:"host_store_adapter_run_ref,omitempty"`
	HostRunstoreRef                   DisplaySafeRef                       `json:"host_runstore_ref,omitempty"`
	HostObjectiveStoreRef             DisplaySafeRef                       `json:"host_objective_store_ref,omitempty"`
	ExpectedStoreTransactionRef       DisplaySafeRef                       `json:"expected_store_transaction_ref,omitempty"`
	ExpectedStoreReadbackRef          DisplaySafeRef                       `json:"expected_store_readback_ref,omitempty"`
	AppliedTransactionRef             DisplaySafeRef                       `json:"applied_transaction_ref,omitempty"`
	AppliedRunstoreRef                DisplaySafeRef                       `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef          DisplaySafeRef                       `json:"applied_objective_state_ref,omitempty"`
	FailureRef                        DisplaySafeRef                       `json:"failure_ref,omitempty"`
	CompensationRef                   DisplaySafeRef                       `json:"compensation_ref,omitempty"`
	MissingInputs                     []MissingInput                       `json:"missing_inputs,omitempty"`
	BlockedReasons                    []string                             `json:"blocked_reasons,omitempty"`
	FailureClass                      FailureClass                         `json:"failure_class,omitempty"`
	Boundaries                        []Boundary                           `json:"boundaries,omitempty"`
	NextHostAction                    NextHostAction                       `json:"next_host_action,omitempty"`
	RunnerEffect                      string                               `json:"runner_effect,omitempty"`
	PromptEffect                      string                               `json:"prompt_effect,omitempty"`
	CoreInvocationExecuted            bool                                 `json:"core_invocation_executed"`
	DurableWriteByCore                bool                                 `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore         bool                                 `json:"objective_store_write_by_core"`
	RunstoreWriteByCore               bool                                 `json:"runstore_write_by_core"`
	RawOutputLoaded                   bool                                 `json:"raw_output_loaded"`
}

type ObjectiveRunStoreDurableReadbackInput struct {
	ObjectiveRunStoreReadbackRef DisplaySafeRef                 `json:"objective_run_store_readback_ref,omitempty"`
	DurableResult                ObjectiveRunStoreDurableResult `json:"durable_result,omitempty"`
	ObservedTransactionRef       DisplaySafeRef                 `json:"observed_transaction_ref,omitempty"`
	ObservedRunstoreRef          DisplaySafeRef                 `json:"observed_runstore_ref,omitempty"`
	ObservedObjectiveStateRef    DisplaySafeRef                 `json:"observed_objective_state_ref,omitempty"`
	ReplayRef                    DisplaySafeRef                 `json:"replay_ref,omitempty"`
	RawOutputLoaded              bool                           `json:"raw_output_loaded"`
}

type ObjectiveRunStoreDurableReadback struct {
	ContractVersion                 string                                 `json:"contract_version,omitempty"`
	Projected                       bool                                   `json:"projected"`
	Available                       bool                                   `json:"available"`
	Status                          HostActionStatus                       `json:"status,omitempty"`
	Mode                            string                                 `json:"mode,omitempty"`
	ObjectiveRunStoreReadbackBound  bool                                   `json:"objective_run_store_readback_bound"`
	TransactionReplayVerified       bool                                   `json:"transaction_replay_verified"`
	ReadyForRuntimeLoopContinuation bool                                   `json:"ready_for_runtime_loop_continuation"`
	DurableResult                   ObjectiveRunStoreDurableResult         `json:"durable_result,omitempty"`
	StoreMutationReadback           ProductionAdapterStoreMutationReadback `json:"store_mutation_readback,omitempty"`
	ObjectiveRunStoreReadbackRef    DisplaySafeRef                         `json:"objective_run_store_readback_ref,omitempty"`
	ObjectiveRunStoreResultRef      DisplaySafeRef                         `json:"objective_run_store_result_ref,omitempty"`
	ObjectiveRunStoreRequestRef     DisplaySafeRef                         `json:"objective_run_store_request_ref,omitempty"`
	ObjectiveRunRef                 DisplaySafeRef                         `json:"objective_run_ref,omitempty"`
	ObjectiveRef                    DisplaySafeRef                         `json:"objective_ref,omitempty"`
	LedgerRef                       DisplaySafeRef                         `json:"ledger_ref,omitempty"`
	VerificationRef                 DisplaySafeRef                         `json:"verification_ref,omitempty"`
	ReplannerDecisionRef            DisplaySafeRef                         `json:"replanner_decision_ref,omitempty"`
	StoreMutationReadbackRef        DisplaySafeRef                         `json:"store_mutation_readback_ref,omitempty"`
	StoreMutationResultRef          DisplaySafeRef                         `json:"store_mutation_result_ref,omitempty"`
	StoreMutationRequestRef         DisplaySafeRef                         `json:"store_mutation_request_ref,omitempty"`
	StoreAdapterRef                 DisplaySafeRef                         `json:"store_adapter_ref,omitempty"`
	HostStoreAdapterRunRef          DisplaySafeRef                         `json:"host_store_adapter_run_ref,omitempty"`
	HostRunstoreRef                 DisplaySafeRef                         `json:"host_runstore_ref,omitempty"`
	HostObjectiveStoreRef           DisplaySafeRef                         `json:"host_objective_store_ref,omitempty"`
	ExpectedStoreTransactionRef     DisplaySafeRef                         `json:"expected_store_transaction_ref,omitempty"`
	ExpectedStoreReadbackRef        DisplaySafeRef                         `json:"expected_store_readback_ref,omitempty"`
	AppliedTransactionRef           DisplaySafeRef                         `json:"applied_transaction_ref,omitempty"`
	AppliedRunstoreRef              DisplaySafeRef                         `json:"applied_runstore_ref,omitempty"`
	AppliedObjectiveStateRef        DisplaySafeRef                         `json:"applied_objective_state_ref,omitempty"`
	ObservedTransactionRef          DisplaySafeRef                         `json:"observed_transaction_ref,omitempty"`
	ObservedRunstoreRef             DisplaySafeRef                         `json:"observed_runstore_ref,omitempty"`
	ObservedObjectiveStateRef       DisplaySafeRef                         `json:"observed_objective_state_ref,omitempty"`
	ReplayRef                       DisplaySafeRef                         `json:"replay_ref,omitempty"`
	MissingInputs                   []MissingInput                         `json:"missing_inputs,omitempty"`
	BlockedReasons                  []string                               `json:"blocked_reasons,omitempty"`
	FailureClass                    FailureClass                           `json:"failure_class,omitempty"`
	Boundaries                      []Boundary                             `json:"boundaries,omitempty"`
	NextHostAction                  NextHostAction                         `json:"next_host_action,omitempty"`
	RunnerEffect                    string                                 `json:"runner_effect,omitempty"`
	PromptEffect                    string                                 `json:"prompt_effect,omitempty"`
	CoreInvocationExecuted          bool                                   `json:"core_invocation_executed"`
	DurableWriteByCore              bool                                   `json:"durable_write_by_core"`
	ObjectiveStoreWriteByCore       bool                                   `json:"objective_store_write_by_core"`
	RunstoreWriteByCore             bool                                   `json:"runstore_write_by_core"`
	RawOutputLoaded                 bool                                   `json:"raw_output_loaded"`
}

func BuildObjectiveRunStoreDurableRequest(input ObjectiveRunStoreDurableRequestInput) ObjectiveRunStoreDurableRequest {
	if hostOwnedObjectiveExecutorRuntimeLoopEmpty(input.RuntimeLoop) || productionAdapterStoreMutationDescriptorEmpty(input.StoreMutationDescriptor) {
		return unavailableObjectiveRunStoreDurableRequest()
	}
	loop := input.RuntimeLoop.Normalize()
	descriptor := BuildProductionAdapterStoreMutationDescriptor(input.StoreMutationDescriptor)
	objectiveRunRef := normalizeOneDisplaySafeRef(input.ObjectiveRunRef)
	ledgerRef := firstDisplaySafeRef(input.LedgerRef, loop.Run.Ledger.LedgerRef)
	verificationRef := normalizeOneDisplaySafeRef(input.VerificationRef)
	replannerDecisionRef := normalizeOneDisplaySafeRef(input.ReplannerDecisionRef)
	expectedEventRef := normalizeOneDisplaySafeRef(input.ExpectedDurableEventRef)
	expectedObjectiveStateRef := normalizeOneDisplaySafeRef(input.ExpectedObjectiveStateRef)
	storeRequest := BuildProductionAdapterStoreMutationRequest(ProductionAdapterStoreMutationRequestInput{
		StoreMutationRequestRef:          input.ObjectiveRunStoreRequestRef,
		StoreMutationDescriptor:          descriptor,
		SourceDurableResultRef:           objectiveRunRef,
		SourceDurableEventRef:            expectedEventRef,
		SourceRunstoreRef:                descriptor.HostRunstoreRef,
		SourceObjectiveStateRef:          expectedObjectiveStateRef,
		HostStoreMutationConfirmationRef: input.HostStoreMutationConfirmationRef,
		ExpectedMutationResultRef:        input.ExpectedStoreMutationResultRef,
		ExpectedTransactionRef:           input.ExpectedStoreTransactionRef,
		ExpectedReadbackRef:              input.ExpectedStoreReadbackRef,
		RawOutputLoaded:                  input.RawOutputLoaded || loop.RawOutputLoaded,
	})
	result := ObjectiveRunStoreDurableRequest{
		ContractVersion:                  ContractVersion,
		Projected:                        true,
		Available:                        loop.Available && storeRequest.Available,
		Status:                           HostActionBlocked,
		Mode:                             "objective_run_store_durable_request",
		RuntimeLoop:                      loop,
		StoreMutationRequest:             storeRequest,
		ObjectiveRunStoreRequestRef:      storeRequest.StoreMutationRequestRef,
		ObjectiveRunRef:                  objectiveRunRef,
		ObjectiveRef:                     normalizeOneDisplaySafeRef(DisplaySafeRef(loop.Run.Frame.ID)),
		LedgerRef:                        ledgerRef,
		VerificationRef:                  verificationRef,
		ReplannerDecisionRef:             replannerDecisionRef,
		StoreMutationRequestRef:          storeRequest.StoreMutationRequestRef,
		StoreAdapterRef:                  storeRequest.StoreAdapterRef,
		HostRunstoreRef:                  storeRequest.HostRunstoreRef,
		HostObjectiveStoreRef:            storeRequest.HostObjectiveStoreRef,
		ExpectedDurableEventRef:          expectedEventRef,
		ExpectedObjectiveStateRef:        expectedObjectiveStateRef,
		HostStoreMutationConfirmationRef: storeRequest.HostStoreMutationConfirmationRef,
		ExpectedStoreMutationResultRef:   storeRequest.ExpectedMutationResultRef,
		ExpectedStoreTransactionRef:      storeRequest.ExpectedTransactionRef,
		ExpectedStoreReadbackRef:         storeRequest.ExpectedReadbackRef,
		EvidenceRefs:                     MergeEvidenceRefs(input.AdditionalEvidenceRefs, loop.EvidenceRefs, loop.Run.EvidenceRefs),
		FailureClass:                     FailureNone,
		Boundaries:                       objectiveRunStoreDurableRequestBoundaries(loop.Boundaries, storeRequest.Boundaries, input.Boundaries),
		NextHostAction:                   "prepare_objective_run_store_request",
		RunnerEffect:                     "none",
		PromptEffect:                     "none",
		RawOutputLoaded:                  input.RawOutputLoaded || loop.RawOutputLoaded || storeRequest.RawOutputLoaded,
	}
	if objectiveRunStoreDurableRequestUnsafe(input, loop, storeRequest) {
		result = objectiveRunStoreDurableRequestBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !loop.ReadyForHostPersist {
		result = objectiveRunStoreDurableRequestBlock(result, firstFailureClass(loop.FailureClass, FailureEvidenceMissing), "runtime_loop_not_ready_for_host_persist", "host:objective_runtime_loop_step", firstNextHostAction(loop.NextHostAction, "review_objective_runtime_loop"))
	}
	if !storeRequest.ReadyForHostStoreMutation || !storeRequest.HostMayExecuteStoreMutation {
		result = objectiveRunStoreDurableRequestBlock(result, firstFailureClass(storeRequest.FailureClass, FailureConfigMissing), "store_mutation_request_not_ready", "host:store_mutation_request", firstNextHostAction(storeRequest.NextHostAction, "review_store_mutation_request"))
	}
	for _, check := range []struct {
		ref     DisplaySafeRef
		reason  string
		missing MissingInput
		next    NextHostAction
	}{
		{result.ObjectiveRunStoreRequestRef, "objective_run_store_request_ref_missing", "host:objective_run_store_request_ref", "provide_objective_run_store_request_ref"},
		{result.ObjectiveRunRef, "objective_run_ref_missing", "host:objective_run_ref", "provide_objective_run_ref"},
		{result.LedgerRef, "objective_run_ledger_ref_missing", "host:objective_ledger_ref", "provide_objective_ledger_ref"},
		{result.VerificationRef, "objective_run_verification_ref_missing", "host:objective_verification_ref", "provide_objective_verification_ref"},
		{result.ReplannerDecisionRef, "objective_run_replanner_decision_ref_missing", "host:objective_replanner_decision_ref", "provide_objective_replanner_decision_ref"},
		{result.ExpectedDurableEventRef, "objective_run_durable_event_ref_missing", "host:objective_run_durable_event_ref", "provide_objective_run_durable_event_ref"},
		{result.ExpectedObjectiveStateRef, "objective_run_state_ref_missing", "host:objective_state_ref", "provide_objective_state_ref"},
	} {
		if check.ref == "" {
			result = objectiveRunStoreDurableRequestBlock(result, FailureEvidenceMissing, check.reason, check.missing, check.next)
		}
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionReady
		result.ReadyForHostObjectiveRunStore = true
		result.HostObjectiveRunStoreAuthorized = true
		result.HostMayPersistObjectiveRun = true
		result.NextHostAction = "host_may_persist_objective_run"
		result.Boundaries = AppendBoundaries(result.Boundaries, "ready_for_host_objective_run_store", "host_may_persist_objective_run", "core_objective_run_store_write_not_executed")
	}
	return result.Normalize()
}

func BuildObjectiveRunStoreDurableResult(input ObjectiveRunStoreDurableResultInput) ObjectiveRunStoreDurableResult {
	if objectiveRunStoreDurableRequestEmpty(input.DurableRequest) {
		return unavailableObjectiveRunStoreDurableResult()
	}
	request := input.DurableRequest.Normalize()
	storeResult := BuildProductionAdapterStoreMutationResult(ProductionAdapterStoreMutationResultInput{
		StoreMutationResultRef:     input.ObjectiveRunStoreResultRef,
		StoreMutationRequest:       request.StoreMutationRequest,
		HostStoreAdapterRunRef:     input.HostStoreAdapterRunRef,
		HostStoreMutationReported:  input.HostObjectiveRunStoreReported,
		HostStoreMutationSucceeded: input.HostObjectiveRunStoreSucceeded,
		HostStoreMutationFailed:    input.HostObjectiveRunStoreFailed,
		AppliedTransactionRef:      input.AppliedTransactionRef,
		AppliedRunstoreRef:         input.AppliedRunstoreRef,
		AppliedObjectiveStateRef:   input.AppliedObjectiveStateRef,
		FailureRef:                 input.FailureRef,
		CompensationRef:            input.CompensationRef,
		StoreMutationEvidenceRefs:  input.StoreMutationEvidenceRefs,
		RawOutputLoaded:            input.RawOutputLoaded || request.RawOutputLoaded,
	})
	result := ObjectiveRunStoreDurableResult{
		ContractVersion:                ContractVersion,
		Projected:                      true,
		Available:                      request.Available && storeResult.Available,
		Status:                         HostActionBlocked,
		Mode:                           "objective_run_store_durable_result",
		HostObjectiveRunStoreReported:  input.HostObjectiveRunStoreReported,
		HostObjectiveRunStoreSucceeded: input.HostObjectiveRunStoreSucceeded,
		HostObjectiveRunStoreFailed:    input.HostObjectiveRunStoreFailed,
		DurableRequest:                 request,
		StoreMutationResult:            storeResult,
		ObjectiveRunStoreResultRef:     storeResult.StoreMutationResultRef,
		ObjectiveRunStoreRequestRef:    request.ObjectiveRunStoreRequestRef,
		ObjectiveRunRef:                request.ObjectiveRunRef,
		ObjectiveRef:                   request.ObjectiveRef,
		LedgerRef:                      request.LedgerRef,
		VerificationRef:                request.VerificationRef,
		ReplannerDecisionRef:           request.ReplannerDecisionRef,
		StoreMutationResultRef:         storeResult.StoreMutationResultRef,
		StoreMutationRequestRef:        storeResult.StoreMutationRequestRef,
		StoreAdapterRef:                storeResult.StoreAdapterRef,
		HostStoreAdapterRunRef:         storeResult.HostStoreAdapterRunRef,
		HostRunstoreRef:                storeResult.HostRunstoreRef,
		HostObjectiveStoreRef:          storeResult.HostObjectiveStoreRef,
		ExpectedStoreTransactionRef:    storeResult.ExpectedTransactionRef,
		ExpectedStoreReadbackRef:       storeResult.ExpectedReadbackRef,
		AppliedTransactionRef:          storeResult.AppliedTransactionRef,
		AppliedRunstoreRef:             storeResult.AppliedRunstoreRef,
		AppliedObjectiveStateRef:       storeResult.AppliedObjectiveStateRef,
		FailureRef:                     storeResult.FailureRef,
		CompensationRef:                storeResult.CompensationRef,
		FailureClass:                   firstFailureClass(storeResult.FailureClass, FailureNone),
		MissingInputs:                  cloneMissingInputs(storeResult.MissingInputs),
		BlockedReasons:                 cloneStringSlice(storeResult.BlockedReasons),
		Boundaries:                     objectiveRunStoreDurableResultBoundaries(request.Boundaries, storeResult.Boundaries),
		NextHostAction:                 firstNextHostAction(storeResult.NextHostAction, "provide_objective_run_store_result"),
		RunnerEffect:                   "none",
		PromptEffect:                   "none",
		RawOutputLoaded:                input.RawOutputLoaded || request.RawOutputLoaded || storeResult.RawOutputLoaded,
	}
	if objectiveRunStoreDurableResultUnsafe(input, request, storeResult) {
		result = objectiveRunStoreDurableResultBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !request.ReadyForHostObjectiveRunStore || !request.HostMayPersistObjectiveRun {
		result = objectiveRunStoreDurableResultBlock(result, firstFailureClass(request.FailureClass, FailureConfigMissing), "objective_run_store_request_not_ready", "host:objective_run_store_request", firstNextHostAction(request.NextHostAction, "review_objective_run_store_request"))
	}
	if !storeResult.HostStoreMutationRecorded {
		result = objectiveRunStoreDurableResultBlock(result, firstFailureClass(storeResult.FailureClass, FailureEvidenceMissing), "store_mutation_result_not_recorded", "host:store_mutation_result", firstNextHostAction(storeResult.NextHostAction, "provide_store_mutation_result"))
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 && storeResult.ReadyForStoreMutationReadback {
		result.Status = HostActionRecorded
		result.HostObjectiveRunStoreRecorded = true
		result.ReadyForObjectiveRunStoreReadback = true
		result.NextHostAction = "bind_objective_run_store_readback"
		result.Boundaries = AppendBoundaries(result.Boundaries, "host_objective_run_store_recorded", "ready_for_objective_run_store_readback")
	}
	return result.Normalize()
}

func BuildObjectiveRunStoreDurableReadback(input ObjectiveRunStoreDurableReadbackInput) ObjectiveRunStoreDurableReadback {
	if objectiveRunStoreDurableResultEmpty(input.DurableResult) {
		return unavailableObjectiveRunStoreDurableReadback()
	}
	durableResult := input.DurableResult.Normalize()
	storeReadback := BuildProductionAdapterStoreMutationReadback(ProductionAdapterStoreMutationReadbackInput{
		StoreMutationReadbackRef:  input.ObjectiveRunStoreReadbackRef,
		StoreMutationResult:       durableResult.StoreMutationResult,
		ObservedTransactionRef:    input.ObservedTransactionRef,
		ObservedRunstoreRef:       input.ObservedRunstoreRef,
		ObservedObjectiveStateRef: input.ObservedObjectiveStateRef,
		ReplayRef:                 input.ReplayRef,
		RawOutputLoaded:           input.RawOutputLoaded || durableResult.RawOutputLoaded,
	})
	result := ObjectiveRunStoreDurableReadback{
		ContractVersion:              ContractVersion,
		Projected:                    true,
		Available:                    durableResult.Available && storeReadback.Available,
		Status:                       HostActionBlocked,
		Mode:                         "objective_run_store_durable_readback",
		DurableResult:                durableResult,
		StoreMutationReadback:        storeReadback,
		ObjectiveRunStoreReadbackRef: storeReadback.StoreMutationReadbackRef,
		ObjectiveRunStoreResultRef:   durableResult.ObjectiveRunStoreResultRef,
		ObjectiveRunStoreRequestRef:  durableResult.ObjectiveRunStoreRequestRef,
		ObjectiveRunRef:              durableResult.ObjectiveRunRef,
		ObjectiveRef:                 durableResult.ObjectiveRef,
		LedgerRef:                    durableResult.LedgerRef,
		VerificationRef:              durableResult.VerificationRef,
		ReplannerDecisionRef:         durableResult.ReplannerDecisionRef,
		StoreMutationReadbackRef:     storeReadback.StoreMutationReadbackRef,
		StoreMutationResultRef:       storeReadback.StoreMutationResultRef,
		StoreMutationRequestRef:      storeReadback.StoreMutationRequestRef,
		StoreAdapterRef:              storeReadback.StoreAdapterRef,
		HostStoreAdapterRunRef:       storeReadback.HostStoreAdapterRunRef,
		HostRunstoreRef:              storeReadback.HostRunstoreRef,
		HostObjectiveStoreRef:        storeReadback.HostObjectiveStoreRef,
		ExpectedStoreTransactionRef:  storeReadback.ExpectedTransactionRef,
		ExpectedStoreReadbackRef:     storeReadback.ExpectedReadbackRef,
		AppliedTransactionRef:        storeReadback.AppliedTransactionRef,
		AppliedRunstoreRef:           storeReadback.AppliedRunstoreRef,
		AppliedObjectiveStateRef:     storeReadback.AppliedObjectiveStateRef,
		ObservedTransactionRef:       storeReadback.ObservedTransactionRef,
		ObservedRunstoreRef:          storeReadback.ObservedRunstoreRef,
		ObservedObjectiveStateRef:    storeReadback.ObservedObjectiveStateRef,
		ReplayRef:                    storeReadback.ReplayRef,
		FailureClass:                 firstFailureClass(storeReadback.FailureClass, FailureNone),
		MissingInputs:                cloneMissingInputs(storeReadback.MissingInputs),
		BlockedReasons:               cloneStringSlice(storeReadback.BlockedReasons),
		Boundaries:                   objectiveRunStoreDurableReadbackBoundaries(durableResult.Boundaries, storeReadback.Boundaries),
		NextHostAction:               firstNextHostAction(storeReadback.NextHostAction, "provide_objective_run_store_readback"),
		RunnerEffect:                 "none",
		PromptEffect:                 "none",
		RawOutputLoaded:              input.RawOutputLoaded || durableResult.RawOutputLoaded || storeReadback.RawOutputLoaded,
	}
	if objectiveRunStoreDurableReadbackUnsafe(input, durableResult, storeReadback) {
		result = objectiveRunStoreDurableReadbackBlock(result, FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs")
		return result.Normalize()
	}
	if !durableResult.ReadyForObjectiveRunStoreReadback {
		result = objectiveRunStoreDurableReadbackBlock(result, firstFailureClass(durableResult.FailureClass, FailureEvidenceMissing), "objective_run_store_result_not_ready", "host:objective_run_store_result", firstNextHostAction(durableResult.NextHostAction, "review_objective_run_store_result"))
	}
	if !storeReadback.StoreMutationReadbackBound || !storeReadback.TransactionReplayVerified {
		result = objectiveRunStoreDurableReadbackBlock(result, firstFailureClass(storeReadback.FailureClass, FailureVerificationFailed), "store_mutation_readback_not_bound", "host:store_mutation_readback", firstNextHostAction(storeReadback.NextHostAction, "review_store_mutation_readback"))
	}
	if len(result.MissingInputs) == 0 && len(result.BlockedReasons) == 0 {
		result.Status = HostActionRecorded
		result.ObjectiveRunStoreReadbackBound = true
		result.TransactionReplayVerified = true
		result.ReadyForRuntimeLoopContinuation = true
		result.NextHostAction = "continue_objective_runtime_loop"
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_run_store_readback_bound", "ready_for_runtime_loop_continuation")
	}
	return result.Normalize()
}

func CloneObjectiveRunStoreDurableRequest(in ObjectiveRunStoreDurableRequest) ObjectiveRunStoreDurableRequest {
	out := in
	out.RuntimeLoop = in.RuntimeLoop.Clone()
	out.StoreMutationRequest = in.StoreMutationRequest.Clone()
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveRunStoreDurableRequest) Clone() ObjectiveRunStoreDurableRequest {
	return CloneObjectiveRunStoreDurableRequest(r)
}

func (r ObjectiveRunStoreDurableRequest) Normalize() ObjectiveRunStoreDurableRequest {
	out := CloneObjectiveRunStoreDurableRequest(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_run_store_durable_request"
	}
	out.RuntimeLoop = out.RuntimeLoop.Normalize()
	out.StoreMutationRequest = out.StoreMutationRequest.Normalize()
	out.ObjectiveRunStoreRequestRef = normalizeOneDisplaySafeRef(out.ObjectiveRunStoreRequestRef)
	out.ObjectiveRunRef = normalizeOneDisplaySafeRef(out.ObjectiveRunRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.LedgerRef = normalizeOneDisplaySafeRef(out.LedgerRef)
	out.VerificationRef = normalizeOneDisplaySafeRef(out.VerificationRef)
	out.ReplannerDecisionRef = normalizeOneDisplaySafeRef(out.ReplannerDecisionRef)
	out.StoreMutationRequestRef = normalizeOneDisplaySafeRef(out.StoreMutationRequestRef)
	out.StoreAdapterRef = normalizeOneDisplaySafeRef(out.StoreAdapterRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.HostObjectiveStoreRef = normalizeOneDisplaySafeRef(out.HostObjectiveStoreRef)
	out.ExpectedDurableEventRef = normalizeOneDisplaySafeRef(out.ExpectedDurableEventRef)
	out.ExpectedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ExpectedObjectiveStateRef)
	out.HostStoreMutationConfirmationRef = normalizeOneDisplaySafeRef(out.HostStoreMutationConfirmationRef)
	out.ExpectedStoreMutationResultRef = normalizeOneDisplaySafeRef(out.ExpectedStoreMutationResultRef)
	out.ExpectedStoreTransactionRef = normalizeOneDisplaySafeRef(out.ExpectedStoreTransactionRef)
	out.ExpectedStoreReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedStoreReadbackRef)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
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
		out.ReadyForHostObjectiveRunStore = false
		out.HostObjectiveRunStoreAuthorized = false
		out.HostMayPersistObjectiveRun = false
	}
	if out.RawOutputLoaded || objectiveRunStoreDurableRequestOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ReadyForHostObjectiveRunStore = false
		out.HostObjectiveRunStoreAuthorized = false
		out.HostMayPersistObjectiveRun = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.ReadyForHostObjectiveRunStore = out.ReadyForHostObjectiveRunStore &&
		out.Status == HostActionReady &&
		out.RuntimeLoop.ReadyForHostPersist &&
		out.StoreMutationRequest.ReadyForHostStoreMutation &&
		out.StoreMutationRequest.HostMayExecuteStoreMutation &&
		out.ObjectiveRunStoreRequestRef != "" &&
		out.ObjectiveRunRef != "" &&
		out.LedgerRef != "" &&
		out.VerificationRef != "" &&
		out.ReplannerDecisionRef != "" &&
		out.ExpectedDurableEventRef != "" &&
		out.ExpectedObjectiveStateRef != "" &&
		out.HostStoreMutationConfirmationRef != "" &&
		out.ExpectedStoreMutationResultRef != "" &&
		out.ExpectedStoreTransactionRef != "" &&
		out.ExpectedStoreReadbackRef != "" &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.HostObjectiveRunStoreAuthorized = out.HostObjectiveRunStoreAuthorized && out.ReadyForHostObjectiveRunStore
	out.HostMayPersistObjectiveRun = out.HostMayPersistObjectiveRun &&
		out.ReadyForHostObjectiveRunStore &&
		!out.CoreInvocationExecuted &&
		!out.DurableWriteByCore &&
		!out.ObjectiveStoreWriteByCore &&
		!out.RunstoreWriteByCore
	return out
}

func CloneObjectiveRunStoreDurableResult(in ObjectiveRunStoreDurableResult) ObjectiveRunStoreDurableResult {
	out := in
	out.DurableRequest = in.DurableRequest.Clone()
	out.StoreMutationResult = in.StoreMutationResult.Clone()
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveRunStoreDurableResult) Clone() ObjectiveRunStoreDurableResult {
	return CloneObjectiveRunStoreDurableResult(r)
}

func (r ObjectiveRunStoreDurableResult) Normalize() ObjectiveRunStoreDurableResult {
	out := CloneObjectiveRunStoreDurableResult(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_run_store_durable_result"
	}
	out.DurableRequest = out.DurableRequest.Normalize()
	out.StoreMutationResult = out.StoreMutationResult.Normalize()
	out.ObjectiveRunStoreResultRef = normalizeOneDisplaySafeRef(out.ObjectiveRunStoreResultRef)
	out.ObjectiveRunStoreRequestRef = normalizeOneDisplaySafeRef(out.ObjectiveRunStoreRequestRef)
	out.ObjectiveRunRef = normalizeOneDisplaySafeRef(out.ObjectiveRunRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.LedgerRef = normalizeOneDisplaySafeRef(out.LedgerRef)
	out.VerificationRef = normalizeOneDisplaySafeRef(out.VerificationRef)
	out.ReplannerDecisionRef = normalizeOneDisplaySafeRef(out.ReplannerDecisionRef)
	out.StoreMutationResultRef = normalizeOneDisplaySafeRef(out.StoreMutationResultRef)
	out.StoreMutationRequestRef = normalizeOneDisplaySafeRef(out.StoreMutationRequestRef)
	out.StoreAdapterRef = normalizeOneDisplaySafeRef(out.StoreAdapterRef)
	out.HostStoreAdapterRunRef = normalizeOneDisplaySafeRef(out.HostStoreAdapterRunRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.HostObjectiveStoreRef = normalizeOneDisplaySafeRef(out.HostObjectiveStoreRef)
	out.ExpectedStoreTransactionRef = normalizeOneDisplaySafeRef(out.ExpectedStoreTransactionRef)
	out.ExpectedStoreReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedStoreReadbackRef)
	out.AppliedTransactionRef = normalizeOneDisplaySafeRef(out.AppliedTransactionRef)
	out.AppliedRunstoreRef = normalizeOneDisplaySafeRef(out.AppliedRunstoreRef)
	out.AppliedObjectiveStateRef = normalizeOneDisplaySafeRef(out.AppliedObjectiveStateRef)
	out.FailureRef = normalizeOneDisplaySafeRef(out.FailureRef)
	out.CompensationRef = normalizeOneDisplaySafeRef(out.CompensationRef)
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
		out.HostObjectiveRunStoreRecorded = false
		out.ReadyForObjectiveRunStoreReadback = false
	}
	if out.RawOutputLoaded || objectiveRunStoreDurableResultOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.HostObjectiveRunStoreRecorded = false
		out.ReadyForObjectiveRunStoreReadback = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.HostObjectiveRunStoreRecorded = out.HostObjectiveRunStoreRecorded &&
		(out.Status == HostActionRecorded || out.Status == HostActionReviewRequired) &&
		out.HostObjectiveRunStoreReported &&
		out.ObjectiveRunStoreResultRef != "" &&
		out.HostStoreAdapterRunRef != "" &&
		out.StoreMutationResult.HostStoreMutationRecorded &&
		!out.RawOutputLoaded
	out.ReadyForObjectiveRunStoreReadback = out.ReadyForObjectiveRunStoreReadback &&
		out.Status == HostActionRecorded &&
		out.HostObjectiveRunStoreRecorded &&
		out.StoreMutationResult.ReadyForStoreMutationReadback &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	return out
}

func CloneObjectiveRunStoreDurableReadback(in ObjectiveRunStoreDurableReadback) ObjectiveRunStoreDurableReadback {
	out := in
	out.DurableResult = in.DurableResult.Clone()
	out.StoreMutationReadback = in.StoreMutationReadback.Clone()
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveRunStoreDurableReadback) Clone() ObjectiveRunStoreDurableReadback {
	return CloneObjectiveRunStoreDurableReadback(r)
}

func (r ObjectiveRunStoreDurableReadback) Normalize() ObjectiveRunStoreDurableReadback {
	out := CloneObjectiveRunStoreDurableReadback(r)
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Status = NormalizeHostActionStatus(string(out.Status))
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_run_store_durable_readback"
	}
	out.DurableResult = out.DurableResult.Normalize()
	out.StoreMutationReadback = out.StoreMutationReadback.Normalize()
	out.ObjectiveRunStoreReadbackRef = normalizeOneDisplaySafeRef(out.ObjectiveRunStoreReadbackRef)
	out.ObjectiveRunStoreResultRef = normalizeOneDisplaySafeRef(out.ObjectiveRunStoreResultRef)
	out.ObjectiveRunStoreRequestRef = normalizeOneDisplaySafeRef(out.ObjectiveRunStoreRequestRef)
	out.ObjectiveRunRef = normalizeOneDisplaySafeRef(out.ObjectiveRunRef)
	out.ObjectiveRef = normalizeOneDisplaySafeRef(out.ObjectiveRef)
	out.LedgerRef = normalizeOneDisplaySafeRef(out.LedgerRef)
	out.VerificationRef = normalizeOneDisplaySafeRef(out.VerificationRef)
	out.ReplannerDecisionRef = normalizeOneDisplaySafeRef(out.ReplannerDecisionRef)
	out.StoreMutationReadbackRef = normalizeOneDisplaySafeRef(out.StoreMutationReadbackRef)
	out.StoreMutationResultRef = normalizeOneDisplaySafeRef(out.StoreMutationResultRef)
	out.StoreMutationRequestRef = normalizeOneDisplaySafeRef(out.StoreMutationRequestRef)
	out.StoreAdapterRef = normalizeOneDisplaySafeRef(out.StoreAdapterRef)
	out.HostStoreAdapterRunRef = normalizeOneDisplaySafeRef(out.HostStoreAdapterRunRef)
	out.HostRunstoreRef = normalizeOneDisplaySafeRef(out.HostRunstoreRef)
	out.HostObjectiveStoreRef = normalizeOneDisplaySafeRef(out.HostObjectiveStoreRef)
	out.ExpectedStoreTransactionRef = normalizeOneDisplaySafeRef(out.ExpectedStoreTransactionRef)
	out.ExpectedStoreReadbackRef = normalizeOneDisplaySafeRef(out.ExpectedStoreReadbackRef)
	out.AppliedTransactionRef = normalizeOneDisplaySafeRef(out.AppliedTransactionRef)
	out.AppliedRunstoreRef = normalizeOneDisplaySafeRef(out.AppliedRunstoreRef)
	out.AppliedObjectiveStateRef = normalizeOneDisplaySafeRef(out.AppliedObjectiveStateRef)
	out.ObservedTransactionRef = normalizeOneDisplaySafeRef(out.ObservedTransactionRef)
	out.ObservedRunstoreRef = normalizeOneDisplaySafeRef(out.ObservedRunstoreRef)
	out.ObservedObjectiveStateRef = normalizeOneDisplaySafeRef(out.ObservedObjectiveStateRef)
	out.ReplayRef = normalizeOneDisplaySafeRef(out.ReplayRef)
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
		out.ObjectiveRunStoreReadbackBound = false
		out.TransactionReplayVerified = false
		out.ReadyForRuntimeLoopContinuation = false
	}
	if out.RawOutputLoaded || objectiveRunStoreDurableReadbackOutputUnsafe(out) {
		out.RawOutputLoaded = true
		out.Status = HostActionBlocked
		out.ObjectiveRunStoreReadbackBound = false
		out.TransactionReplayVerified = false
		out.ReadyForRuntimeLoopContinuation = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueControlToken(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		if out.NextHostAction == "" {
			out.NextHostAction = "provide_display_safe_refs"
		}
	}
	out.CoreInvocationExecuted = false
	out.DurableWriteByCore = false
	out.ObjectiveStoreWriteByCore = false
	out.RunstoreWriteByCore = false
	out.ObjectiveRunStoreReadbackBound = out.ObjectiveRunStoreReadbackBound &&
		out.Status == HostActionRecorded &&
		out.DurableResult.ReadyForObjectiveRunStoreReadback &&
		out.StoreMutationReadback.StoreMutationReadbackBound &&
		out.StoreMutationReadback.TransactionReplayVerified &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0 &&
		!out.RawOutputLoaded
	out.TransactionReplayVerified = out.TransactionReplayVerified && out.ObjectiveRunStoreReadbackBound
	out.ReadyForRuntimeLoopContinuation = out.ReadyForRuntimeLoopContinuation && out.ObjectiveRunStoreReadbackBound
	return out
}

func objectiveRunStoreDurableRequestBlock(result ObjectiveRunStoreDurableRequest, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ObjectiveRunStoreDurableRequest {
	result.Status = HostActionBlocked
	result.ReadyForHostObjectiveRunStore = false
	result.HostObjectiveRunStoreAuthorized = false
	result.HostMayPersistObjectiveRun = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_run_store_request_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func objectiveRunStoreDurableResultBlock(result ObjectiveRunStoreDurableResult, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ObjectiveRunStoreDurableResult {
	result.Status = HostActionBlocked
	result.HostObjectiveRunStoreRecorded = false
	result.ReadyForObjectiveRunStoreReadback = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_run_store_result_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func objectiveRunStoreDurableReadbackBlock(result ObjectiveRunStoreDurableReadback, failure FailureClass, reason string, missing MissingInput, next NextHostAction) ObjectiveRunStoreDurableReadback {
	result.Status = HostActionBlocked
	result.ObjectiveRunStoreReadbackBound = false
	result.TransactionReplayVerified = false
	result.ReadyForRuntimeLoopContinuation = false
	result.FailureClass = firstFailureClass(result.FailureClass, failure)
	result.BlockedReasons = appendUniqueControlToken(result.BlockedReasons, reason)
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_run_store_readback_blocked")
	result.NextHostAction = firstNextHostAction(result.NextHostAction, next)
	return result
}

func objectiveRunStoreDurableRequestBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"objective_run_store_durable_request",
			"durable_objective_run_store_adapter",
			"host_owned_objective_run_store",
			"objective_run_store_request_projection_only",
			"store_mutation_gate_reused",
			"transaction_replay_required",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func objectiveRunStoreDurableResultBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"objective_run_store_durable_result",
			"durable_objective_run_store_adapter",
			"host_owned_objective_run_store",
			"objective_run_store_result_projection_only",
			"store_mutation_gate_reused",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func objectiveRunStoreDurableReadbackBoundaries(groups ...[]Boundary) []Boundary {
	return MergeBoundaries(
		[]Boundary{
			"objective_run_store_durable_readback",
			"durable_objective_run_store_adapter",
			"host_owned_objective_run_store",
			"objective_run_store_readback_projection_only",
			"store_mutation_gate_reused",
			"transaction_replay_verified_by_host",
			"display_safe_refs_only",
			"no_runner_dispatch",
			"no_durable_write_by_core",
			"no_objective_store_write_by_core",
			"no_runstore_write_by_core",
		},
		MergeBoundaries(groups...),
	)
}

func objectiveRunStoreDurableRequestUnsafe(input ObjectiveRunStoreDurableRequestInput, loop ObjectiveRuntimeLoopStep, storeRequest ProductionAdapterStoreMutationRequest) bool {
	return input.RawOutputLoaded ||
		loop.RawOutputLoaded ||
		storeRequest.RawOutputLoaded ||
		displaySafeRefRejected(input.ObjectiveRunStoreRequestRef) ||
		displaySafeRefRejected(input.ObjectiveRunRef) ||
		displaySafeRefRejected(input.LedgerRef) ||
		displaySafeRefRejected(input.VerificationRef) ||
		displaySafeRefRejected(input.ReplannerDecisionRef) ||
		displaySafeRefRejected(input.ExpectedDurableEventRef) ||
		displaySafeRefRejected(input.ExpectedObjectiveStateRef) ||
		displaySafeRefRejected(input.HostStoreMutationConfirmationRef) ||
		displaySafeRefRejected(input.ExpectedStoreMutationResultRef) ||
		displaySafeRefRejected(input.ExpectedStoreTransactionRef) ||
		displaySafeRefRejected(input.ExpectedStoreReadbackRef) ||
		evidenceRefRejected(input.AdditionalEvidenceRefs) ||
		objectiveRunStoreDurableRequestLoopUnsafe(loop) ||
		productionAdapterStoreMutationRequestOutputUnsafe(storeRequest)
}

func objectiveRunStoreDurableRequestLoopUnsafe(loop ObjectiveRuntimeLoopStep) bool {
	return displaySafeRefRejected(loop.Run.CurrentStrategyRef) ||
		displaySafeRefRejected(loop.Run.Ledger.LedgerRef) ||
		evidenceRefRejected(loop.EvidenceRefs) ||
		evidenceRefRejected(loop.Run.EvidenceRefs) ||
		loop.RawOutputLoaded ||
		loop.Run.RawOutputLoaded ||
		loop.LedgerPatch.RawOutputLoaded ||
		loop.ControllerDecision.RawOutputLoaded ||
		loop.Verification.RawOutputLoaded
}

func objectiveRunStoreDurableResultUnsafe(input ObjectiveRunStoreDurableResultInput, request ObjectiveRunStoreDurableRequest, storeResult ProductionAdapterStoreMutationResult) bool {
	return input.RawOutputLoaded ||
		request.RawOutputLoaded ||
		storeResult.RawOutputLoaded ||
		displaySafeRefRejected(input.ObjectiveRunStoreResultRef) ||
		displaySafeRefRejected(input.HostStoreAdapterRunRef) ||
		displaySafeRefRejected(input.AppliedTransactionRef) ||
		displaySafeRefRejected(input.AppliedRunstoreRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		displaySafeRefSliceRejected(input.StoreMutationEvidenceRefs) ||
		objectiveRunStoreDurableRequestOutputUnsafe(request) ||
		productionAdapterStoreMutationResultOutputUnsafe(storeResult)
}

func objectiveRunStoreDurableReadbackUnsafe(input ObjectiveRunStoreDurableReadbackInput, result ObjectiveRunStoreDurableResult, storeReadback ProductionAdapterStoreMutationReadback) bool {
	return input.RawOutputLoaded ||
		result.RawOutputLoaded ||
		storeReadback.RawOutputLoaded ||
		displaySafeRefRejected(input.ObjectiveRunStoreReadbackRef) ||
		displaySafeRefRejected(input.ObservedTransactionRef) ||
		displaySafeRefRejected(input.ObservedRunstoreRef) ||
		displaySafeRefRejected(input.ObservedObjectiveStateRef) ||
		displaySafeRefRejected(input.ReplayRef) ||
		objectiveRunStoreDurableResultOutputUnsafe(result) ||
		productionAdapterStoreMutationReadbackOutputUnsafe(storeReadback)
}

func objectiveRunStoreDurableRequestOutputUnsafe(input ObjectiveRunStoreDurableRequest) bool {
	return displaySafeRefRejected(input.ObjectiveRunStoreRequestRef) ||
		displaySafeRefRejected(input.ObjectiveRunRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.LedgerRef) ||
		displaySafeRefRejected(input.VerificationRef) ||
		displaySafeRefRejected(input.ReplannerDecisionRef) ||
		displaySafeRefRejected(input.StoreMutationRequestRef) ||
		displaySafeRefRejected(input.StoreAdapterRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.HostObjectiveStoreRef) ||
		displaySafeRefRejected(input.ExpectedDurableEventRef) ||
		displaySafeRefRejected(input.ExpectedObjectiveStateRef) ||
		displaySafeRefRejected(input.HostStoreMutationConfirmationRef) ||
		displaySafeRefRejected(input.ExpectedStoreMutationResultRef) ||
		displaySafeRefRejected(input.ExpectedStoreTransactionRef) ||
		displaySafeRefRejected(input.ExpectedStoreReadbackRef) ||
		evidenceRefRejected(input.EvidenceRefs) ||
		productionAdapterStoreMutationRequestOutputUnsafe(input.StoreMutationRequest) ||
		input.RawOutputLoaded
}

func objectiveRunStoreDurableResultOutputUnsafe(input ObjectiveRunStoreDurableResult) bool {
	return displaySafeRefRejected(input.ObjectiveRunStoreResultRef) ||
		displaySafeRefRejected(input.ObjectiveRunStoreRequestRef) ||
		displaySafeRefRejected(input.ObjectiveRunRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.LedgerRef) ||
		displaySafeRefRejected(input.VerificationRef) ||
		displaySafeRefRejected(input.ReplannerDecisionRef) ||
		displaySafeRefRejected(input.StoreMutationResultRef) ||
		displaySafeRefRejected(input.StoreMutationRequestRef) ||
		displaySafeRefRejected(input.StoreAdapterRef) ||
		displaySafeRefRejected(input.HostStoreAdapterRunRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.HostObjectiveStoreRef) ||
		displaySafeRefRejected(input.ExpectedStoreTransactionRef) ||
		displaySafeRefRejected(input.ExpectedStoreReadbackRef) ||
		displaySafeRefRejected(input.AppliedTransactionRef) ||
		displaySafeRefRejected(input.AppliedRunstoreRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.FailureRef) ||
		displaySafeRefRejected(input.CompensationRef) ||
		objectiveRunStoreDurableRequestOutputUnsafe(input.DurableRequest) ||
		productionAdapterStoreMutationResultOutputUnsafe(input.StoreMutationResult) ||
		input.RawOutputLoaded
}

func objectiveRunStoreDurableReadbackOutputUnsafe(input ObjectiveRunStoreDurableReadback) bool {
	return displaySafeRefRejected(input.ObjectiveRunStoreReadbackRef) ||
		displaySafeRefRejected(input.ObjectiveRunStoreResultRef) ||
		displaySafeRefRejected(input.ObjectiveRunStoreRequestRef) ||
		displaySafeRefRejected(input.ObjectiveRunRef) ||
		displaySafeRefRejected(input.ObjectiveRef) ||
		displaySafeRefRejected(input.LedgerRef) ||
		displaySafeRefRejected(input.VerificationRef) ||
		displaySafeRefRejected(input.ReplannerDecisionRef) ||
		displaySafeRefRejected(input.StoreMutationReadbackRef) ||
		displaySafeRefRejected(input.StoreMutationResultRef) ||
		displaySafeRefRejected(input.StoreMutationRequestRef) ||
		displaySafeRefRejected(input.StoreAdapterRef) ||
		displaySafeRefRejected(input.HostStoreAdapterRunRef) ||
		displaySafeRefRejected(input.HostRunstoreRef) ||
		displaySafeRefRejected(input.HostObjectiveStoreRef) ||
		displaySafeRefRejected(input.ExpectedStoreTransactionRef) ||
		displaySafeRefRejected(input.ExpectedStoreReadbackRef) ||
		displaySafeRefRejected(input.AppliedTransactionRef) ||
		displaySafeRefRejected(input.AppliedRunstoreRef) ||
		displaySafeRefRejected(input.AppliedObjectiveStateRef) ||
		displaySafeRefRejected(input.ObservedTransactionRef) ||
		displaySafeRefRejected(input.ObservedRunstoreRef) ||
		displaySafeRefRejected(input.ObservedObjectiveStateRef) ||
		displaySafeRefRejected(input.ReplayRef) ||
		objectiveRunStoreDurableResultOutputUnsafe(input.DurableResult) ||
		productionAdapterStoreMutationReadbackOutputUnsafe(input.StoreMutationReadback) ||
		input.RawOutputLoaded
}

func objectiveRunStoreDurableRequestEmpty(request ObjectiveRunStoreDurableRequest) bool {
	return !request.Projected &&
		!request.Available &&
		request.Status == "" &&
		request.Mode == "" &&
		request.ObjectiveRunStoreRequestRef == "" &&
		len(request.MissingInputs) == 0 &&
		len(request.BlockedReasons) == 0 &&
		len(request.Boundaries) == 0 &&
		request.NextHostAction == "" &&
		!request.RawOutputLoaded
}

func objectiveRunStoreDurableResultEmpty(result ObjectiveRunStoreDurableResult) bool {
	return !result.Projected &&
		!result.Available &&
		result.Status == "" &&
		result.Mode == "" &&
		result.ObjectiveRunStoreResultRef == "" &&
		len(result.MissingInputs) == 0 &&
		len(result.BlockedReasons) == 0 &&
		len(result.Boundaries) == 0 &&
		result.NextHostAction == "" &&
		!result.RawOutputLoaded
}

func unavailableObjectiveRunStoreDurableRequest() ObjectiveRunStoreDurableRequest {
	return ObjectiveRunStoreDurableRequest{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "objective_run_store_durable_request",
		Boundaries:      objectiveRunStoreDurableRequestBoundaries(),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		NextHostAction:  "provide_objective_runtime_loop_and_store_descriptor",
	}
}

func unavailableObjectiveRunStoreDurableResult() ObjectiveRunStoreDurableResult {
	return ObjectiveRunStoreDurableResult{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "objective_run_store_durable_result",
		Boundaries:      objectiveRunStoreDurableResultBoundaries(),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		NextHostAction:  "provide_objective_run_store_request",
	}
}

func unavailableObjectiveRunStoreDurableReadback() ObjectiveRunStoreDurableReadback {
	return ObjectiveRunStoreDurableReadback{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       false,
		Status:          HostActionNotReady,
		Mode:            "objective_run_store_durable_readback",
		Boundaries:      objectiveRunStoreDurableReadbackBoundaries(),
		RunnerEffect:    "none",
		PromptEffect:    "none",
		NextHostAction:  "provide_objective_run_store_result",
	}
}
