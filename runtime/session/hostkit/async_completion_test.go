package hostkit

import (
	"testing"

	agentxcontrolplane "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

func TestDelegationWorkerAsyncCompletionProjectsDurableCompletedChild(t *testing.T) {
	projection := BuildDelegationWorkerAsyncCompletionProjection(DelegationWorkerAsyncCompletionInput{
		Run:            delegationWorkerObjectiveTestRun(),
		BackendKind:    agentxcontrolplane.AutoDelegationAsyncBackendDurable,
		BackendRef:     "backend:durable_delegation_worker",
		RequireDurable: true,
		Invocation:     delegationWorkerObjectiveTestInvocation(),
		Boundaries:     []agentxcontrolplane.Boundary{"test_delegation_worker_async_completion"},
	})

	if projection.Status != agentxcontrolplane.VerificationSatisfied ||
		!projection.Ready ||
		!projection.ReadyForReadback ||
		!projection.ReadyForResume ||
		!projection.Durable ||
		projection.ProcessLocal ||
		projection.BackendRef != "backend:durable_delegation_worker" ||
		len(projection.CompletedChildRefs) != 1 ||
		projection.CompletedChildRefs[0] != "subgoal:delegation_worker_closure" ||
		len(projection.CompletionEnvelopes) != 1 ||
		projection.CompletionEnvelopes[0].EnvelopeRef != "envelope:delegation_worker_closure" ||
		projection.CompletionEnvelopes[0].WorkerOutputAcceptedAsFact ||
		!projection.CompletionEnvelopes[0].RequiresParentVerification ||
		len(projection.ResumeRequest.ChildRefs) != 1 ||
		projection.ResumeRequest.ChildRefs[0] != "subgoal:delegation_worker_closure" ||
		projection.ResumeRequest.ParentLedgerRef != "ledger:delegation_worker_closure" ||
		projection.NextHostAction != "resume_parent_objective_for_delegation_merge" {
		t.Fatalf("expected durable completed child async completion projection: %+v", projection)
	}
	for _, boundary := range []agentxcontrolplane.Boundary{
		"hostruntime_delegation_worker_async_completion_projection",
		"completion_envelope_only",
		"raw_child_tool_logs_not_consumed",
		"no_child_task_spawn_by_core",
	} {
		if !hostruntimeBoundaryContains(projection.Boundaries, boundary) {
			t.Fatalf("expected boundary %q in %+v", boundary, projection.Boundaries)
		}
	}
}

func TestDelegationWorkerAsyncCompletionProjectsActiveChildCancelInterruptReadback(t *testing.T) {
	invocation := delegationWorkerObjectiveTestInvocation()
	invocation.HostInvocationCompleted = false
	invocation.ReadyForWorkerResultReview = false
	projection := BuildDelegationWorkerAsyncCompletionProjection(DelegationWorkerAsyncCompletionInput{
		Run:             delegationWorkerObjectiveTestRun(),
		BackendKind:     agentxcontrolplane.AutoDelegationAsyncBackendProcessLocal,
		BackendRef:      "backend:process_local_delegation_worker",
		Status:          agentxcontrolplane.AutoDelegationAsyncChildStatusActive,
		Invocation:      invocation,
		CancellationRef: "cancel:delegation_worker_closure",
		InterruptionRef: "interrupt:delegation_worker_closure",
		CurrentAction:   "collect_evidence",
	})

	if projection.Status != agentxcontrolplane.VerificationPartial ||
		!projection.Ready ||
		!projection.ReadyForReadback ||
		projection.ReadyForResume ||
		!projection.ProcessLocal ||
		projection.Durable ||
		len(projection.ActiveChildRefs) != 1 ||
		projection.ActiveChildRefs[0] != "subgoal:delegation_worker_closure" ||
		len(projection.Children) != 1 ||
		!projection.Children[0].CancelAvailable ||
		!projection.Children[0].InterruptAvailable ||
		projection.Children[0].CancellationRef != "cancel:delegation_worker_closure" ||
		projection.Children[0].InterruptionRef != "interrupt:delegation_worker_closure" ||
		projection.Children[0].CurrentAction != "collect_evidence" ||
		projection.NextHostAction != "monitor_auto_delegation_async_children" {
		t.Fatalf("expected active child cancel/interrupt readback projection: %+v", projection)
	}
}

func hostruntimeBoundaryContains(values []agentxcontrolplane.Boundary, want agentxcontrolplane.Boundary) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
