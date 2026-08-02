package fault_test

import (
	"context"
	"errors"
	"testing"

	"github.com/wsnacj/agentx-go/providers"
	"github.com/wsnacj/agentx-go/providers/fault"
)

func TestClassifyPreservesStableKinds(t *testing.T) {
	tests := []struct {
		err       error
		kind      fault.Kind
		retryable bool
	}{
		{context.Canceled, fault.KindCanceled, false},
		{context.DeadlineExceeded, fault.KindTimeout, true},
		{providers.ErrUnsupported, fault.KindCapability, false},
		{&providers.APIError{StatusCode: 429, Body: "rate limit"}, fault.KindRateLimit, true},
		{&providers.APIError{StatusCode: 400, Body: "context length exceeded"}, fault.KindOverflow, false},
	}
	for _, test := range tests {
		classification := fault.Classify(test.err)
		if classification.Kind != test.kind || classification.Retryable != test.retryable {
			t.Fatalf("%v => %#v", test.err, classification)
		}
	}
	cause := errors.New("role order")
	wrapped := fault.Wrap(fault.KindRoleOrdering, cause)
	if !errors.Is(wrapped, cause) || fault.KindOf(wrapped) != fault.KindRoleOrdering {
		t.Fatalf("wrapped = %v", wrapped)
	}
}
