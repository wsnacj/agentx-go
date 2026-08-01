package controlcontract_test

import (
	"context"
	"testing"

	controlcontract "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

func TestObjectiveDurableProjectionDoesNotWriteWithoutHostStore(t *testing.T) {
	request := controlcontract.BuildObjectiveRunStoreDurableRequest(controlcontract.ObjectiveRunStoreDurableRequestInput{})
	if request.ReadyForHostObjectiveRunStore || request.HostMayPersistObjectiveRun ||
		request.DurableWriteByCore || request.ObjectiveStoreWriteByCore || request.RunstoreWriteByCore ||
		request.RunnerEffect != "none" {
		t.Fatalf("unexpected hostless durable request: %#v", request)
	}
}

func TestObjectiveFinalAnswerRequiresInjectedSynthesizerWhenEnabled(t *testing.T) {
	answer := controlcontract.BuildObjectiveFinalAnswer(context.Background(), controlcontract.ObjectiveFinalAnswerInput{
		EnableSynthesizer: true,
	})
	if answer.ReadyForUser || answer.FailureClass != controlcontract.FailureConfigMissing ||
		answer.NextHostAction != "provide_objective_final_answer_synthesizer" ||
		answer.RunnerEffect != "none" || answer.PromptEffect != "none" {
		t.Fatalf("unexpected hostless final answer: %#v", answer)
	}
}

func TestAsyncDelegationAndControllerDoNotDispatchWithoutHostRuntime(t *testing.T) {
	completion := controlcontract.BuildAutoDelegationAsyncCompletionProjection(controlcontract.AutoDelegationAsyncCompletionInput{})
	if completion.Ready || completion.ReadyForResume || completion.RunnerEffect != "none" {
		t.Fatalf("unexpected hostless async completion: %#v", completion)
	}

	decision := controlcontract.BuildAutoDelegationControllerDecision(controlcontract.AutoDelegationControllerInput{})
	if decision.Ready || decision.HostMayDispatch || decision.RunnerEffect != "none" || decision.RuntimeEffect != "none" {
		t.Fatalf("unexpected hostless delegation controller: %#v", decision)
	}
}
