package resourcepolicy_test

import (
	"testing"

	"github.com/wsnacj/agentx-go/runtime/hosthttp/resourcepolicy"
)

func TestExternalConsumerNarrowsHostBudget(t *testing.T) {
	got, err := resourcepolicy.NarrowPositiveInt(50, 10)
	if err != nil {
		t.Fatalf("NarrowPositiveInt: %v", err)
	}
	if got != 10 {
		t.Fatalf("budget = %d", got)
	}
}
