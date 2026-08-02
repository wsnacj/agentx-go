// Package fault classifies provider and transport errors into stable buckets.
package fault

import (
	"context"
	"errors"
	"net"
	"sort"
	"strings"

	"github.com/wsnacj/agentx-go/providers"
)

// Kind identifies one stable provider error taxonomy bucket.
type Kind string

const (
	KindUnknown           Kind = "unknown"
	KindCanceled          Kind = "canceled"
	KindTimeout           Kind = "timeout"
	KindNetwork           Kind = "network"
	KindOverflow          Kind = "overflow"
	KindRoleOrdering      Kind = "role_ordering"
	KindSessionCorruption Kind = "session_corruption"
	KindRateLimit         Kind = "rate_limit"
	KindAuth              Kind = "auth"
	KindPermission        Kind = "permission"
	KindSafety            Kind = "safety"
	KindCapability        Kind = "capability"
	KindInvalidRequest    Kind = "invalid_request"
	KindRetryableUpstream Kind = "retryable_upstream"
	KindUpstream          Kind = "upstream"
)

type Classification struct {
	Kind       Kind
	Retryable  bool
	StatusCode int
	Detail     string
}
type KindCount struct {
	Kind  Kind
	Count int
}
type StatusCodeCount struct {
	StatusCode int
	Count      int
}
type Summary struct {
	TotalCount        int
	RetryableCount    int
	NonRetryableCount int
	DominantKind      Kind
	Kinds             []KindCount
	StatusCodes       []StatusCodeCount
}

type statusError interface {
	error
	HTTPStatusCode() int
	HTTPResponseBody() string
}
type kindError interface {
	error
	FaultKind() Kind
}
type classifiedError struct {
	kind  Kind
	cause error
}

func (e *classifiedError) Error() string {
	if e == nil || e.cause == nil {
		return "llmx fault"
	}
	return e.cause.Error()
}
func (e *classifiedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}
func (e *classifiedError) FaultKind() Kind {
	if e == nil {
		return KindUnknown
	}
	return e.kind
}

// Wrap attaches a fault kind without changing the visible error.
func Wrap(kind Kind, err error) error {
	if err == nil {
		return nil
	}
	if strings.TrimSpace(string(kind)) == "" {
		kind = KindUnknown
	}
	return &classifiedError{kind: kind, cause: err}
}

func KindOf(err error) Kind      { return Classify(err).Kind }
func IsRetryable(err error) bool { return Classify(err).Retryable }

func Summarize(items []Classification) Summary {
	if len(items) == 0 {
		return Summary{}
	}
	kindCounts := make(map[Kind]int)
	statusCounts := make(map[int]int)
	summary := Summary{TotalCount: len(items)}
	for _, item := range items {
		kind := item.Kind
		if kind == "" {
			kind = KindUnknown
		}
		kindCounts[kind]++
		if item.Retryable {
			summary.RetryableCount++
		} else {
			summary.NonRetryableCount++
		}
		if item.StatusCode > 0 {
			statusCounts[item.StatusCode]++
		}
	}
	summary.Kinds = summarizeKinds(kindCounts)
	summary.StatusCodes = summarizeStatusCodes(statusCounts)
	if len(summary.Kinds) > 0 {
		summary.DominantKind = summary.Kinds[0].Kind
	}
	return summary
}

func SummarizeErrors(errs []error) Summary {
	items := make([]Classification, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			items = append(items, Classify(err))
		}
	}
	return Summarize(items)
}

func Classify(err error) Classification {
	if err == nil {
		return Classification{}
	}
	if errors.Is(err, context.Canceled) {
		return Classification{Kind: KindCanceled, Detail: "context canceled"}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return Classification{Kind: KindTimeout, Retryable: true, Detail: "context deadline exceeded"}
	}
	if errors.Is(err, providers.ErrUnsupported) {
		return Classification{Kind: KindCapability, Detail: "provider capability unsupported"}
	}
	var typed kindError
	if errors.As(err, &typed) {
		kind := typed.FaultKind()
		if kind == "" {
			kind = KindUnknown
		}
		return Classification{Kind: kind, Retryable: retryableKind(kind), Detail: err.Error()}
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return Classification{Kind: KindTimeout, Retryable: true, Detail: netErr.Error()}
		}
		return Classification{Kind: KindNetwork, Retryable: true, Detail: netErr.Error()}
	}
	var httpErr statusError
	if errors.As(err, &httpErr) {
		return classifyHTTPError(httpErr.HTTPStatusCode(), httpErr.HTTPResponseBody())
	}
	return Classification{Kind: KindUnknown, Detail: err.Error()}
}

func classifyHTTPError(status int, body string) Classification {
	normalized := strings.ToLower(strings.TrimSpace(body))
	contains := func(terms ...string) bool {
		for _, term := range terms {
			if strings.Contains(normalized, term) {
				return true
			}
		}
		return false
	}
	switch {
	case contains("context length", "context window", "too many tokens", "maximum context", "prompt is too long", "input is too long", "token limit", "request too large", "maximum output tokens", "max tokens exceeded"):
		return Classification{Kind: KindOverflow, StatusCode: status, Detail: body}
	case contains("roles must alternate", "incorrect role", "role information", "role ordering"):
		return Classification{Kind: KindRoleOrdering, StatusCode: status, Detail: body}
	case contains("function call turn comes immediately after", "session history corrupted"):
		return Classification{Kind: KindSessionCorruption, StatusCode: status, Detail: body}
	case contains("content filter", "content_filter", "safety", "moderation", "prohibited content", "unsafe", "blocked", "recitation", "blocklist"):
		return Classification{Kind: KindSafety, StatusCode: status, Detail: body}
	}
	switch status {
	case 400, 404, 405, 406, 411, 412, 414, 415, 422:
		return Classification{Kind: KindInvalidRequest, StatusCode: status, Detail: body}
	case 401:
		return Classification{Kind: KindAuth, StatusCode: status, Detail: body}
	case 403:
		return Classification{Kind: KindPermission, StatusCode: status, Detail: body}
	case 408:
		return Classification{Kind: KindTimeout, Retryable: true, StatusCode: status, Detail: body}
	case 409:
		return Classification{Kind: KindRetryableUpstream, Retryable: true, StatusCode: status, Detail: body}
	case 413:
		return Classification{Kind: KindOverflow, StatusCode: status, Detail: body}
	case 429:
		return Classification{Kind: KindRateLimit, Retryable: true, StatusCode: status, Detail: body}
	}
	if status >= 500 && status <= 599 {
		return Classification{Kind: KindRetryableUpstream, Retryable: true, StatusCode: status, Detail: body}
	}
	if status >= 400 && status <= 499 {
		return Classification{Kind: KindInvalidRequest, StatusCode: status, Detail: body}
	}
	return Classification{Kind: KindUpstream, StatusCode: status, Detail: body}
}

func retryableKind(kind Kind) bool {
	return kind == KindTimeout || kind == KindNetwork || kind == KindRateLimit || kind == KindRetryableUpstream
}
func summarizeKinds(counts map[Kind]int) []KindCount {
	out := make([]KindCount, 0, len(counts))
	for kind, count := range counts {
		out = append(out, KindCount{Kind: kind, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Kind < out[j].Kind
	})
	return out
}
func summarizeStatusCodes(counts map[int]int) []StatusCodeCount {
	out := make([]StatusCodeCount, 0, len(counts))
	for code, count := range counts {
		out = append(out, StatusCodeCount{StatusCode: code, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].StatusCode < out[j].StatusCode
	})
	return out
}
