package expert

import "errors"

type ErrorCode string

const (
	ErrorCodeInvalidSpec       ErrorCode = "invalid_spec"
	ErrorCodeUnsupportedSchema ErrorCode = "unsupported_schema"
	ErrorCodeForbiddenField    ErrorCode = "forbidden_field"
)

// Error is display-safe. Cause preserves diagnostic detail without including
// it in Error's user-facing text.
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
		return "expert specification is invalid"
	case ErrorCodeUnsupportedSchema:
		return "expert specification schema is unsupported"
	case ErrorCodeForbiddenField:
		return "expert specification contains a host-owned field"
	default:
		return "expert specification operation failed"
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
