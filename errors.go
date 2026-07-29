package agentx

import (
	"context"
	"errors"
	"strings"
)

// ErrorCode 是稳定、可供程序判断的错误类别。
type ErrorCode string

const (
	CodeInvalidArgument    ErrorCode = "invalid_argument"
	CodeCanceled           ErrorCode = "canceled"
	CodeDeadlineExceeded   ErrorCode = "deadline_exceeded"
	CodeClientClosed       ErrorCode = "client_closed"
	CodeUnsupportedProfile ErrorCode = "unsupported_profile"
	CodeExecutionFailed    ErrorCode = "execution_failed"
	CodeShutdownFailed     ErrorCode = "shutdown_failed"
)

// Error 是 Client 返回的 display-safe typed error。
//
// Message 不包含 adapter/backend 原始错误；Unwrap 仍保留 cause，供
// errors.Is/errors.As 进行程序化判断。
type Error struct {
	Code      ErrorCode
	Retryable bool
	Message   string

	cause error
}

// Error 返回可安全展示的错误文本。
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if message := strings.TrimSpace(e.Message); message != "" {
		return message
	}
	return string(e.Code)
}

// Unwrap 返回原始 cause，供程序化错误检查使用。
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// Is 按稳定 ErrorCode 比较两个 AgentX 错误。
func (e *Error) Is(target error) bool {
	other, ok := target.(*Error)
	return ok && e != nil && other != nil && e.Code != "" && e.Code == other.Code
}

func mapRunError(adapter ExecutionAdapter, err error) *Error {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return newError(CodeDeadlineExceeded, err)
	case errors.Is(err, context.Canceled):
		return newError(CodeCanceled, err)
	}
	code := CodeExecutionFailed
	if adapter != nil {
		code = normalizeAdapterErrorCode(adapter.ClassifyError(err))
	}
	return newError(code, err)
}

func normalizeAdapterErrorCode(code ErrorCode) ErrorCode {
	switch code {
	case CodeCanceled, CodeDeadlineExceeded, CodeClientClosed, CodeExecutionFailed:
		return code
	default:
		return CodeExecutionFailed
	}
}

func newError(code ErrorCode, cause error) *Error {
	return &Error{
		Code:      code,
		Retryable: false,
		Message:   messageForCode(code),
		cause:     cause,
	}
}

func messageForCode(code ErrorCode) string {
	switch code {
	case CodeInvalidArgument:
		return "invalid request argument"
	case CodeCanceled:
		return "execution canceled by caller"
	case CodeDeadlineExceeded:
		return "execution exceeded caller deadline"
	case CodeClientClosed:
		return "client is closing or closed"
	case CodeUnsupportedProfile:
		return "execution profile is not supported"
	case CodeShutdownFailed:
		return "client shutdown did not complete within this call"
	default:
		return "execution failed"
	}
}

func nextActionForCode(code ErrorCode) string {
	switch code {
	case CodeInvalidArgument:
		return "fix_request"
	case CodeCanceled:
		return "caller_decides_retry"
	case CodeDeadlineExceeded:
		return "increase_deadline_or_retry"
	case CodeClientClosed:
		return "create_new_client"
	case CodeUnsupportedProfile:
		return "use_supported_profile"
	case CodeShutdownFailed:
		return "continue_shutdown_wait"
	default:
		return "inspect_owner_diagnostics"
	}
}
