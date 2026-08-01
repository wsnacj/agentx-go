package controlcontract

import "testing"

func TestAutoDelegationAsyncCompletionBlocksUnsupportedBackend(t *testing.T) {
	projection := BuildAutoDelegationAsyncCompletionProjection(AutoDelegationAsyncCompletionInput{
		HostBridge:  BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput()),
		BackendKind: AutoDelegationAsyncBackendUnsupported,
	})

	if projection.Status != VerificationBlocked ||
		projection.Ready ||
		projection.FailureClass != FailureUnsupportedOperation ||
		!autoDelegationMissingInputContains(projection.MissingInputs, "host:auto_delegation_async_backend") ||
		!autoDelegationBoundaryContains(projection.Boundaries, "auto_delegation_async_backend_unsupported") {
		t.Fatalf("expected unsupported async backend block: %+v", projection)
	}
}

func TestAutoDelegationAsyncCompletionProjectsProcessLocalActiveReadback(t *testing.T) {
	projection := BuildAutoDelegationAsyncCompletionProjection(AutoDelegationAsyncCompletionInput{
		HostBridge:  BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput()),
		BackendKind: AutoDelegationAsyncBackendProcessLocal,
		BackendRef:  "backend:process_local_child_runtime",
		Children: []AutoDelegationAsyncChildReadback{{
			ChildRef:             "child:collect_public_sources",
			Role:                 AutoDelegationChildRoleLeaf,
			Status:               AutoDelegationAsyncChildStatusActive,
			AgeSeconds:           17,
			CurrentAction:        "collect_public_source",
			CapabilityRefs:       []DisplaySafeRef{"capability:public_source"},
			AllowedToolRefs:      []DisplaySafeRef{"tool:public_source_read"},
			BoundCapabilityRefs:  []DisplaySafeRef{"capability:public_source"},
			BoundAllowedToolRefs: []DisplaySafeRef{"tool:public_source_read"},
			WorkerRunRef:         "run:child_collect_public_sources",
			WorkerReadbackRef:    "readback:child_collect_public_sources",
			CancellationRef:      "cancel:child_collect_public_sources",
			InterruptionRef:      "interrupt:child_collect_public_sources",
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:child_collect_public_sources_progress",
				Kind:     "progress",
				Strength: EvidenceAdequate,
				Source:   "host:auto_delegation_async_readback",
			}},
			CancelAvailable:    true,
			InterruptAvailable: true,
		}},
	})

	if projection.Status != VerificationPartial ||
		!projection.Ready ||
		!projection.ReadyForReadback ||
		projection.ReadyForResume ||
		!projection.ProcessLocal ||
		projection.Durable ||
		len(projection.ActiveChildRefs) != 1 ||
		projection.ActiveChildRefs[0] != "child:collect_public_sources" ||
		len(projection.Children) != 1 ||
		!projection.Children[0].CancelAvailable ||
		!projection.Children[0].InterruptAvailable ||
		projection.NextHostAction != "monitor_auto_delegation_async_children" {
		t.Fatalf("expected process-local active projection: %+v", projection)
	}
}

func TestAutoDelegationAsyncCompletionBlocksWhenDurableRequiredButProcessLocal(t *testing.T) {
	projection := BuildAutoDelegationAsyncCompletionProjection(AutoDelegationAsyncCompletionInput{
		HostBridge:     BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput()),
		BackendKind:    AutoDelegationAsyncBackendProcessLocal,
		RequireDurable: true,
		Children: []AutoDelegationAsyncChildReadback{{
			ChildRef:          "child:collect_public_sources",
			Status:            AutoDelegationAsyncChildStatusActive,
			WorkerReadbackRef: "readback:child_collect_public_sources",
		}},
	})

	if projection.Status != VerificationBlocked ||
		projection.Ready ||
		projection.FailureClass != FailureUnsupportedOperation ||
		!autoDelegationMissingInputContains(projection.MissingInputs, "host:durable_auto_delegation_child_readback") ||
		!autoDelegationBoundaryContains(projection.Boundaries, "durable_child_readback_required") {
		t.Fatalf("expected durable-required block: %+v", projection)
	}
}

func TestAutoDelegationAsyncCompletionBuildsDurableResumeRequestForCompletedChild(t *testing.T) {
	projection := BuildAutoDelegationAsyncCompletionProjection(AutoDelegationAsyncCompletionInput{
		HostBridge:            BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput()),
		BackendKind:           AutoDelegationAsyncBackendDurable,
		BackendRef:            "backend:durable_child_runtime",
		RequireDurable:        true,
		ParentObjectiveRef:    "objective:root",
		ParentObjectiveRunRef: "objective_run:root",
		ParentLedgerRef:       "ledger:root",
		ParentResumeRef:       "resume:root_after_child",
		Children: []AutoDelegationAsyncChildReadback{{
			ChildRef:                         "child:collect_public_sources",
			Role:                             AutoDelegationChildRoleLeaf,
			Status:                           AutoDelegationAsyncChildStatusCompleted,
			CapabilityRefs:                   []DisplaySafeRef{"capability:public_source"},
			AllowedToolRefs:                  []DisplaySafeRef{"tool:public_source_read"},
			BoundCapabilityRefs:              []DisplaySafeRef{"capability:public_source"},
			BoundAllowedToolRefs:             []DisplaySafeRef{"tool:public_source_read"},
			WorkerRunRef:                     "run:child_collect_public_sources",
			WorkerResultRef:                  "result:child_collect_public_sources",
			WorkerReadbackRef:                "readback:child_collect_public_sources",
			ObservationRef:                   "observation:child_collect_public_sources",
			CompletionEnvelopeRef:            "envelope:child_collect_public_sources",
			ReadyForWorkerResultReview:       true,
			WorkerResultRequiresVerification: true,
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:child_collect_public_sources_result",
				Kind:     "worker_result",
				Strength: EvidenceAdequate,
				Source:   "host:auto_delegation_async_readback",
			}},
		}},
	})

	if projection.Status != VerificationSatisfied ||
		!projection.Ready ||
		!projection.ReadyForResume ||
		!projection.Durable ||
		len(projection.CompletedChildRefs) != 1 ||
		projection.CompletedChildRefs[0] != "child:collect_public_sources" ||
		len(projection.CompletionEnvelopes) != 1 ||
		projection.CompletionEnvelopes[0].WorkerOutputAcceptedAsFact ||
		!projection.CompletionEnvelopes[0].RequiresParentVerification ||
		len(projection.ResumeRequest.CompletionEnvelopeRefs) != 1 ||
		projection.ResumeRequest.CompletionEnvelopeRefs[0] != "envelope:child_collect_public_sources" ||
		projection.NextHostAction != "resume_parent_objective_for_delegation_merge" {
		t.Fatalf("expected durable resume-ready projection: %+v", projection)
	}
	for _, boundary := range []Boundary{
		"completion_envelope_only",
		"raw_child_tool_logs_not_consumed",
		"auto_delegation_parent_resume_ready",
	} {
		if !autoDelegationBoundaryContains(projection.Boundaries, boundary) {
			t.Fatalf("missing boundary %q in %+v", boundary, projection.Boundaries)
		}
	}
}

func TestAutoDelegationAsyncCompletionRejectsChildOutputAcceptedAsFact(t *testing.T) {
	projection := BuildAutoDelegationAsyncCompletionProjection(AutoDelegationAsyncCompletionInput{
		HostBridge:            BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput()),
		BackendKind:           AutoDelegationAsyncBackendDurable,
		ParentObjectiveRunRef: "objective_run:root",
		ParentResumeRef:       "resume:root_after_child",
		Children: []AutoDelegationAsyncChildReadback{{
			ChildRef:                         "child:collect_public_sources",
			Status:                           AutoDelegationAsyncChildStatusCompleted,
			WorkerRunRef:                     "run:child_collect_public_sources",
			WorkerResultRef:                  "result:child_collect_public_sources",
			WorkerReadbackRef:                "readback:child_collect_public_sources",
			CompletionEnvelopeRef:            "envelope:child_collect_public_sources",
			ReadyForWorkerResultReview:       true,
			WorkerResultRequiresVerification: true,
			WorkerOutputAcceptedAsFact:       true,
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:child_collect_public_sources_result",
				Kind:     "worker_result",
				Strength: EvidenceAdequate,
				Source:   "host:auto_delegation_async_readback",
			}},
		}},
	})

	if projection.Status != VerificationBlocked ||
		projection.Ready ||
		!autoDelegationMissingInputContains(projection.MissingInputs, "contract:child_output_not_fact") ||
		!autoDelegationHostBridgeStringContains(projection.BlockedReasons, "worker_output_accepted_as_fact_rejected") {
		t.Fatalf("expected child output-as-fact rejection: %+v", projection)
	}
}

func TestAutoDelegationAsyncCompletionProjectsCanceledChildExactID(t *testing.T) {
	projection := BuildAutoDelegationAsyncCompletionProjection(AutoDelegationAsyncCompletionInput{
		HostBridge:  BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput()),
		BackendKind: AutoDelegationAsyncBackendDurable,
		Children: []AutoDelegationAsyncChildReadback{{
			ChildRef:        "child:collect_public_sources",
			Status:          AutoDelegationAsyncChildStatusCancelled,
			WorkerRunRef:    "run:child_collect_public_sources",
			CancellationRef: "cancel:child_collect_public_sources",
		}},
	})

	if projection.Status != VerificationPartial ||
		!projection.Ready ||
		len(projection.CancelledChildRefs) != 1 ||
		projection.CancelledChildRefs[0] != "child:collect_public_sources" ||
		projection.NextHostAction != "provide_auto_delegation_parent_merge" {
		t.Fatalf("expected canceled child exact-id projection: %+v", projection)
	}

	decision := BuildAutoDelegationControllerDecision(AutoDelegationControllerInput{
		HostBridge:      BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput()),
		AsyncCompletion: projection,
	})
	if decision.Action != AutoDelegationControllerActionCollectExisting ||
		!decision.HostMayCollect ||
		len(decision.CancelledChildRefs) != 1 ||
		decision.CancelledChildRefs[0] != "child:collect_public_sources" ||
		!autoDelegationMissingInputContains(decision.MissingInputs, "host:auto_delegation_parent_merge") {
		t.Fatalf("expected controller to collect canceled child exact id: %+v", decision)
	}
}

func TestAutoDelegationAsyncCompletionDoesNotResumeMixedCancelledChild(t *testing.T) {
	projection := BuildAutoDelegationAsyncCompletionProjection(AutoDelegationAsyncCompletionInput{
		HostBridge:            BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput()),
		BackendKind:           AutoDelegationAsyncBackendDurable,
		ParentObjectiveRunRef: "objective_run:root",
		ParentResumeRef:       "resume:root_after_child",
		Children: []AutoDelegationAsyncChildReadback{
			{
				ChildRef:                         "child:collect_public_sources",
				Status:                           AutoDelegationAsyncChildStatusCompleted,
				WorkerRunRef:                     "run:child_collect_public_sources",
				WorkerResultRef:                  "result:child_collect_public_sources",
				WorkerReadbackRef:                "readback:child_collect_public_sources",
				CompletionEnvelopeRef:            "envelope:child_collect_public_sources",
				ReadyForWorkerResultReview:       true,
				WorkerResultRequiresVerification: true,
				EvidenceRefs: []EvidenceRef{{
					Ref:      "evidence:child_collect_public_sources_result",
					Kind:     "worker_result",
					Strength: EvidenceAdequate,
					Source:   "host:auto_delegation_async_readback",
				}},
			},
			{
				ChildRef:        "child:inspect_transport_options",
				Status:          AutoDelegationAsyncChildStatusCancelled,
				WorkerRunRef:    "run:child_inspect_transport_options",
				CancellationRef: "cancel:child_inspect_transport_options",
			},
		},
	})

	if projection.Status != VerificationPartial ||
		!projection.Ready ||
		!projection.ReadyForReadback ||
		projection.ReadyForResume ||
		len(projection.CompletedChildRefs) != 1 ||
		len(projection.CancelledChildRefs) != 1 ||
		len(projection.CompletionEnvelopes) != 1 ||
		len(projection.ResumeRequest.ChildRefs) != 0 ||
		projection.NextHostAction != "provide_auto_delegation_parent_merge" {
		t.Fatalf("expected mixed completed/cancelled readback to require parent merge, got %+v", projection)
	}
}

func TestAutoDelegationControllerConsumesAsyncActiveReadback(t *testing.T) {
	bridge := BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput())
	async := BuildAutoDelegationAsyncCompletionProjection(AutoDelegationAsyncCompletionInput{
		HostBridge:  bridge,
		BackendKind: AutoDelegationAsyncBackendProcessLocal,
		Children: []AutoDelegationAsyncChildReadback{{
			ChildRef:          "child:collect_public_sources",
			Status:            AutoDelegationAsyncChildStatusActive,
			WorkerRunRef:      "run:child_collect_public_sources",
			WorkerReadbackRef: "readback:child_collect_public_sources",
		}},
	})

	decision := BuildAutoDelegationControllerDecision(AutoDelegationControllerInput{
		HostBridge:        bridge,
		AsyncCompletion:   async,
		RequestedDispatch: true,
	})

	if decision.Action != AutoDelegationControllerActionCollectExisting ||
		!decision.HostMayCollect ||
		decision.HostMayDispatch ||
		len(decision.OpenChildRefs) != 1 ||
		decision.OpenChildRefs[0] != "child:collect_public_sources" ||
		len(decision.RejectedActions) != 1 ||
		decision.RejectedActions[0] != AutoDelegationControllerActionSpawnOnce ||
		!autoDelegationBoundaryContains(decision.Boundaries, "auto_delegation_controller_consumed_async_child_readback") {
		t.Fatalf("expected controller to collect async active child: %+v", decision)
	}
}

func TestAutoDelegationControllerConsumesAsyncCompletedReadback(t *testing.T) {
	bridge := BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput())
	async := BuildAutoDelegationAsyncCompletionProjection(AutoDelegationAsyncCompletionInput{
		HostBridge:            bridge,
		BackendKind:           AutoDelegationAsyncBackendDurable,
		ParentObjectiveRunRef: "objective_run:root",
		ParentResumeRef:       "resume:root_after_child",
		Children: []AutoDelegationAsyncChildReadback{{
			ChildRef:                         "child:collect_public_sources",
			Status:                           AutoDelegationAsyncChildStatusCompleted,
			WorkerRunRef:                     "run:child_collect_public_sources",
			WorkerResultRef:                  "result:child_collect_public_sources",
			WorkerReadbackRef:                "readback:child_collect_public_sources",
			CompletionEnvelopeRef:            "envelope:child_collect_public_sources",
			ReadyForWorkerResultReview:       true,
			WorkerResultRequiresVerification: true,
			EvidenceRefs: []EvidenceRef{{
				Ref:      "evidence:child_collect_public_sources_result",
				Kind:     "worker_result",
				Strength: EvidenceAdequate,
				Source:   "host:auto_delegation_async_readback",
			}},
		}},
	})

	decision := BuildAutoDelegationControllerDecision(AutoDelegationControllerInput{
		HostBridge:      bridge,
		AsyncCompletion: async,
	})

	if decision.Action != AutoDelegationControllerActionCollectExisting ||
		!decision.HostMayCollect ||
		len(decision.CompletedChildRefs) != 1 ||
		decision.CompletedChildRefs[0] != "child:collect_public_sources" ||
		!autoDelegationMissingInputContains(decision.MissingInputs, "host:auto_delegation_parent_merge") ||
		decision.NextHostAction != "provide_auto_delegation_parent_merge" {
		t.Fatalf("expected controller to collect async terminal child: %+v", decision)
	}
}

func TestAutoDelegationControllerTreatsAsyncInterruptedReadbackAsTerminal(t *testing.T) {
	bridge := BuildAutoDelegationHostBridge(validAutoDelegationHostBridgeInput())
	async := BuildAutoDelegationAsyncCompletionProjection(AutoDelegationAsyncCompletionInput{
		HostBridge:  bridge,
		BackendKind: AutoDelegationAsyncBackendDurable,
		Children: []AutoDelegationAsyncChildReadback{{
			ChildRef:        "child:collect_public_sources",
			Status:          AutoDelegationAsyncChildStatusInterrupted,
			WorkerRunRef:    "run:child_collect_public_sources",
			InterruptionRef: "interrupt:child_collect_public_sources",
		}},
	})

	decision := BuildAutoDelegationControllerDecision(AutoDelegationControllerInput{
		HostBridge:      bridge,
		AsyncCompletion: async,
	})

	if decision.Action != AutoDelegationControllerActionCollectExisting ||
		!decision.HostMayCollect ||
		len(decision.CancelledChildRefs) != 1 ||
		decision.CancelledChildRefs[0] != "child:collect_public_sources" ||
		!autoDelegationMissingInputContains(decision.MissingInputs, "host:auto_delegation_parent_merge") {
		t.Fatalf("expected controller to collect interrupted child as terminal: %+v", decision)
	}
}
