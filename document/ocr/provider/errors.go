package provider

import "fmt"

// TextInError 表示 TextIn API 返回的错误。
type ErrorCategorizer interface {
	ErrorCategory() string
}

// TextInError 表示 TextIn API 返回的错误。
type TextInError struct {
	Status   int
	Code     int
	Message  string
	Raw      string
	Category string
}

func (e *TextInError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Code != 0 {
		return fmt.Sprintf("textin error status=%d code=%d msg=%s", e.Status, e.Code, e.Message)
	}
	return fmt.Sprintf("textin error status=%d msg=%s", e.Status, e.Message)
}

// ErrorCategory 实现 ErrorCategorizer，用于归类 provider 错误。
func (e *TextInError) ErrorCategory() string {
	if e == nil {
		return ""
	}
	if e.Category != "" {
		return e.Category
	}
	if e.Status >= 500 {
		return "server"
	}
	if e.Status >= 400 {
		return "client"
	}
	return "textin"
}
