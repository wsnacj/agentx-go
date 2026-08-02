package tools

import (
	"errors"
	"fmt"
	"strings"
)

type browserRouteGateKind string

const (
	browserRouteGateTarget         browserRouteGateKind = "target_requires_explicit_runtime"
	browserRouteGateCurrentPage    browserRouteGateKind = "current_page_requires_explicit_runtime"
	browserRouteGateDefaultBrowser browserRouteGateKind = "default_browser_requires_explicit_runtime"
	browserRouteGateURLRoute       browserRouteGateKind = "url_route_requires_explicit_runtime"
	browserRouteGateRuntimeAction  browserRouteGateKind = "runtime_action_requires_explicit_runtime"
	browserRouteGateManagedAction  browserRouteGateKind = "managed_action_requires_explicit_runtime"
	browserRouteGateDefaultRequest browserRouteGateKind = "default_request_requires_explicit_runtime"
)

type browserRouteGateError struct {
	kind    browserRouteGateKind
	message string
}

func (e *browserRouteGateError) Error() string {
	if e == nil {
		return "browser route requires an explicit runtime target"
	}
	return strings.TrimSpace(e.message)
}

func newBrowserRouteGateError(kind browserRouteGateKind, message string) error {
	return &browserRouteGateError{kind: kind, message: strings.TrimSpace(message)}
}

func browserRouteErrorHasGateKind(err error, kind browserRouteGateKind) bool {
	var gateErr *browserRouteGateError
	return errors.As(err, &gateErr) && gateErr.kind == kind
}

type browserManagedRouteUnavailableError struct {
	target   string
	endpoint string
	cause    error
}

func (e *browserManagedRouteUnavailableError) Error() string {
	if e == nil {
		return "browser proxy managed_route_unavailable"
	}
	return fmt.Sprintf(
		"browser proxy managed_route_unavailable target=%s endpoint=%s: %v",
		strings.TrimSpace(e.target),
		strings.TrimSpace(e.endpoint),
		e.cause,
	)
}

func (e *browserManagedRouteUnavailableError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}
