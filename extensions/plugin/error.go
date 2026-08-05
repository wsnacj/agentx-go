package plugin

import "errors"

// ErrorCode is a stable plugin manifest failure category.
type ErrorCode string

const (
	ErrorCodeInvalidManifest   ErrorCode = "invalid_manifest"
	ErrorCodeUnsupportedSchema ErrorCode = "unsupported_schema"
	ErrorCodeInvalidPath       ErrorCode = "invalid_path"
	ErrorCodeForbiddenField    ErrorCode = "forbidden_field"
)

// Error is display-safe. Cause is available through errors.Is/As but is never
// included in the display text.
type Error struct {
	Code  ErrorCode
	Cause error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	switch e.Code {
	case ErrorCodeInvalidManifest:
		return "plugin manifest is invalid"
	case ErrorCodeUnsupportedSchema:
		return "plugin manifest schema is unsupported"
	case ErrorCodeInvalidPath:
		return "plugin manifest path is invalid"
	case ErrorCodeForbiddenField:
		return "plugin manifest contains a host-owned field"
	default:
		return "plugin manifest operation failed"
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

// AsError returns the typed plugin error when present.
func AsError(err error) (*Error, bool) {
	var typed *Error
	if !errors.As(err, &typed) {
		return nil, false
	}
	return typed, true
}
