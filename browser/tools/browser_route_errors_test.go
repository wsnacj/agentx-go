package tools

import (
	"errors"
	"fmt"
	"testing"
)

func TestBrowserRouteGateMatchingUsesTypeNotMessage(t *testing.T) {
	requested := BrowserRuntimeInfo{}
	typed := browserImplicitLegacyHostDefaultRequestError(true, requested, "default")
	if typed == nil || !browserImplicitLegacyHostRouteErrMatchesDefaultRequestError(true, requested, "default", fmt.Errorf("wrapped: %w", typed)) {
		t.Fatalf("typed route gate was not recognized: %v", typed)
	}
	if browserImplicitLegacyHostRouteErrMatchesDefaultRequestError(true, requested, "default", errors.New(typed.Error())) {
		t.Fatal("matching error text must not impersonate a route gate")
	}
}

func TestBrowserManagedRouteFailureUsesTypeNotMessage(t *testing.T) {
	typed := &browserManagedRouteUnavailableError{
		target:   "node",
		endpoint: "http://127.0.0.1:1",
		cause:    errors.New("launch message changed"),
	}
	if !browserRuntimeRouteErrIsManagedLaunchFailure(fmt.Errorf("wrapped: %w", typed)) {
		t.Fatal("typed managed route failure was not recognized")
	}
	if browserRuntimeRouteErrIsManagedLaunchFailure(errors.New(typed.Error())) {
		t.Fatal("matching error text must not impersonate a managed route failure")
	}
}
