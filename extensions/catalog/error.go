package catalog

import "errors"

// ErrorCode is a stable discovery failure category.
type ErrorCode string

const (
	ErrorCodeInvalidPolicy  ErrorCode = "invalid_policy"
	ErrorCodeInvalidAsset   ErrorCode = "invalid_asset"
	ErrorCodeDuplicateAsset ErrorCode = "duplicate_asset"
	ErrorCodeInvalidQuery   ErrorCode = "invalid_query"
)

// Error is display-safe. Cause is available through errors.Is/As but is never
// included in Error's display text.
type Error struct {
	Code  ErrorCode
	Cause error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	switch e.Code {
	case ErrorCodeInvalidPolicy:
		return "capability catalog policy is invalid"
	case ErrorCodeInvalidAsset:
		return "capability catalog asset is invalid"
	case ErrorCodeDuplicateAsset:
		return "capability catalog asset is duplicated"
	case ErrorCodeInvalidQuery:
		return "capability catalog query is invalid"
	default:
		return "capability catalog operation failed"
	}
}

// Unwrap exposes the underlying cause without displaying it.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is compares catalog errors by stable code.
func (e *Error) Is(target error) bool {
	other, ok := target.(*Error)
	return ok && e != nil && other != nil && e.Code == other.Code
}

// AsError returns the typed catalog error when present.
func AsError(err error) (*Error, bool) {
	var typed *Error
	if !errors.As(err, &typed) {
		return nil, false
	}
	return typed, true
}
