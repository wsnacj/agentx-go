package team

import "errors"

type ErrorCode string

const (
	ErrorCodeInvalidSpec       ErrorCode = "invalid_spec"
	ErrorCodeUnsupportedSchema ErrorCode = "unsupported_schema"
	ErrorCodeForbiddenField    ErrorCode = "forbidden_field"
	ErrorCodeDependencyCycle   ErrorCode = "dependency_cycle"
)

type Error struct {
	Code  ErrorCode
	Cause error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	switch e.Code {
	case ErrorCodeInvalidSpec:
		return "team specification is invalid"
	case ErrorCodeUnsupportedSchema:
		return "team specification schema is unsupported"
	case ErrorCodeForbiddenField:
		return "team specification contains a host-owned field"
	case ErrorCodeDependencyCycle:
		return "team member dependency graph contains a cycle"
	default:
		return "team specification operation failed"
	}
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
func AsError(err error) (*Error, bool) {
	var typed *Error
	if !errors.As(err, &typed) {
		return nil, false
	}
	return typed, true
}
