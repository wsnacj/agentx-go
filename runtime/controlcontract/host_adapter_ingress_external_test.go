package controlcontract_test

import (
	"testing"

	controlcontract "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

func TestHostAdapterAndManagedIngressFailClosedWithoutHostCapabilities(t *testing.T) {
	registry := controlcontract.BuildHostAdapterRegistry(controlcontract.HostAdapterRegistryInput{})
	if !registry.Projected || registry.ReadyForRuntimeRequest || registry.RunnerEffect != "none" {
		t.Fatalf("unexpected empty Host adapter registry: %#v", registry)
	}

	request := controlcontract.BuildRuntimeAdapterExecutionRequest(controlcontract.RuntimeAdapterExecutionRequestInput{
		Registry: registry,
	})
	if !request.Projected || request.ReadyForHostExecution || request.RunnerEffect != "none" {
		t.Fatalf("unexpected hostless runtime adapter request: %#v", request)
	}

	ingress := controlcontract.BuildManagedObjectiveIngress(controlcontract.ManagedObjectiveIngressInput{})
	if !ingress.Projected || ingress.ReadyForRuntimeAdapter || ingress.RunnerEffect != "none" {
		t.Fatalf("unexpected hostless managed objective ingress: %#v", ingress)
	}
}

func TestStoreMutationAndCompensationContractsDoNotExecuteCoreEffects(t *testing.T) {
	storeRequest := controlcontract.BuildProductionAdapterStoreMutationRequest(controlcontract.ProductionAdapterStoreMutationRequestInput{})
	if storeRequest.ReadyForHostStoreMutation || storeRequest.CoreInvocationExecuted || storeRequest.DurableWriteByCore || storeRequest.RunnerEffect != "none" {
		t.Fatalf("unexpected empty store mutation request: %#v", storeRequest)
	}

	compensation := controlcontract.BuildObjectiveCompensationExecutionRequest(controlcontract.ObjectiveCompensationExecutionRequestInput{})
	if compensation.ReadyForHostCompensation || compensation.CoreExecutionExecuted || compensation.CompensationExecutedByCore || compensation.RunnerEffect != "none" {
		t.Fatalf("unexpected empty compensation request: %#v", compensation)
	}
}
