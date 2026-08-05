package mcp

import "errors"

type ErrorCode string

const (
	ErrorCodeInvalidConfig   ErrorCode = "invalid_config"
	ErrorCodeClosed          ErrorCode = "closed"
	ErrorCodeNotInitialized  ErrorCode = "not_initialized"
	ErrorCodeProtocol        ErrorCode = "protocol_error"
	ErrorCodeTransport       ErrorCode = "transport_error"
	ErrorCodeRemote          ErrorCode = "remote_error"
	ErrorCodeInvalidTool     ErrorCode = "invalid_tool"
	ErrorCodeToolUnavailable ErrorCode = "tool_unavailable"
)

// Error is display-safe. Remote messages and transport details remain in
// Cause and never appear in Error's display text.
type Error struct {
	Code       ErrorCode
	RemoteCode int
	Cause      error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	switch e.Code {
	case ErrorCodeInvalidConfig:
		return "MCP client configuration is invalid"
	case ErrorCodeClosed:
		return "MCP client is closed"
	case ErrorCodeNotInitialized:
		return "MCP client is not initialized"
	case ErrorCodeProtocol:
		return "MCP protocol response is invalid"
	case ErrorCodeTransport:
		return "MCP transport failed"
	case ErrorCodeRemote:
		return "MCP server returned an error"
	case ErrorCodeInvalidTool:
		return "MCP tool declaration is invalid"
	case ErrorCodeToolUnavailable:
		return "MCP tool is unavailable"
	default:
		return "MCP operation failed"
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
