// Package providers contains contracts shared by optional model-provider clients.
//
// This module is experimental. It never discovers credentials or selects a
// product model: hosts must inject those decisions explicitly.
package providers

import (
	"errors"
	"fmt"
)

// ErrUnsupported indicates that a provider does not implement a requested capability.
var ErrUnsupported = errors.New("llmx: capability not supported")

// APIError represents a non-successful HTTP response from a provider.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("llmx: api status %d: %s", e.StatusCode, e.Body)
}

// HTTPStatusCode returns the provider response status code.
func (e *APIError) HTTPStatusCode() int {
	if e == nil {
		return 0
	}
	return e.StatusCode
}

// HTTPResponseBody returns the provider response body used for classification.
func (e *APIError) HTTPResponseBody() string {
	if e == nil {
		return ""
	}
	return e.Body
}
