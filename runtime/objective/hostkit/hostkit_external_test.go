package hostkit_test

import (
	"context"
	"testing"

	objective "github.com/wsnacj/agentx-go/runtime/objective"
	objectivehostkit "github.com/wsnacj/agentx-go/runtime/objective/hostkit"
)

func TestPublicConstructionRejectsMissingHandler(t *testing.T) {
	_, err := objectivehostkit.New(objectivehostkit.Config{})
	if err == nil {
		t.Fatal("expected missing handler error")
	}
}

func TestPublicHandlerTypeUsesObjectivePackage(t *testing.T) {
	var handler objectivehostkit.Handler = func(context.Context, objective.RuntimeAdapterRequest) objective.RuntimeAdapterResult {
		return objective.RuntimeAdapterResult{}
	}
	if handler == nil {
		t.Fatal("handler should be assignable")
	}
}
