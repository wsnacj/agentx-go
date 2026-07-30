package workflow_test

import (
	"errors"
	"testing"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

func TestExternalConsumerImplementsValidatorAndPreservesErrorIdentity(t *testing.T) {
	want := errors.New("validation sentinel")
	var validator workflow.Validator = validatorFunc(
		func(workflow.Spec) error { return want },
	)
	if got := validator.ValidateSpec(workflow.Spec{ID: "external-consumer"}); !errors.Is(got, want) {
		t.Fatalf("ValidateSpec() error = %v, want sentinel identity", got)
	}
}

type validatorFunc func(workflow.Spec) error

func (f validatorFunc) ValidateSpec(spec workflow.Spec) error {
	return f(spec)
}
