package productshell_test

import (
	"context"
	"testing"

	"github.com/wsnacj/agentx-go/extensions/productshell"
)

func TestExperimentalPreparationConsumer(t *testing.T) {
	pipeline := productshell.NewPreparationPipeline(productshell.PreparationRuntimeFuncs{})
	result, err := pipeline.Prepare(context.Background(), "session-a", productshell.Input{
		UserMessage: "[skill:review] check this",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.UserMessage != "check this" || len(result.RequestedSkills) != 1 || result.RequestedSkills[0] != "review" {
		t.Fatalf("unexpected preparation result: %#v", result)
	}
}
