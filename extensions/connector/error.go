package connector

import "errors"

// ErrorCode is a stable connector declaration failure category.
type ErrorCode string

const (
	ErrorCodeInvalidSpec ErrorCode = "invalid_spec"
)

// Error is display-safe and supports errors.Is/As by stable code.
type Error struct {
	Code  ErrorCode
	Cause error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return "connector specification is invalid"
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *Error) Is(target error) bool {
	other, ok := target.(*Error)
	return ok && e != nil && other != nil && e.Code == other.Code
}

// AsError returns the typed connector error when present.
func AsError(err error) (*Error, bool) {
	var typed *Error
	if !errors.As(err, &typed) {
		return nil, false
	}
	return typed, true
}
