package contextwindow

import (
	"context"
	"errors"
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/runtime/transcript"
)

const summaryMarker = "[agentx context summary]"

// ErrorCode is a stable context-window failure category.
type ErrorCode string

const (
	// ErrorCodeInvalidPolicy marks an unusable context-window policy.
	ErrorCodeInvalidPolicy ErrorCode = "invalid_policy"
	// ErrorCodeCanceled marks caller cancellation.
	ErrorCodeCanceled ErrorCode = "canceled"
	// ErrorCodeDeadlineExceeded marks a caller-owned deadline.
	ErrorCodeDeadlineExceeded ErrorCode = "deadline_exceeded"
	// ErrorCodeSummarizerUnavailable marks an overflow that requires a semantic summarizer.
	ErrorCodeSummarizerUnavailable ErrorCode = "summarizer_unavailable"
	// ErrorCodeSummarizationFailed marks a Host summarizer failure.
	ErrorCodeSummarizationFailed ErrorCode = "summarization_failed"
	// ErrorCodeInvalidSummary marks an empty or over-budget summary.
	ErrorCodeInvalidSummary ErrorCode = "invalid_summary"
	// ErrorCodeLimitUnresolved marks a result that still exceeds the configured limit.
	ErrorCodeLimitUnresolved ErrorCode = "limit_unresolved"
	// ErrorCodeTokenCountFailed marks a Host token counter failure.
	ErrorCodeTokenCountFailed ErrorCode = "token_count_failed"
)

// Error is a display-safe typed context-window error. Cause is available via
// errors.Unwrap but is never included in Error's display text.
type Error struct {
	Code  ErrorCode
	Cause error
}

// Error returns a stable display-safe message.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	switch e.Code {
	case ErrorCodeInvalidPolicy:
		return "context window policy is invalid"
	case ErrorCodeCanceled:
		return "context window preparation was canceled"
	case ErrorCodeDeadlineExceeded:
		return "context window preparation deadline was exceeded"
	case ErrorCodeSummarizerUnavailable:
		return "context window requires a host summarizer"
	case ErrorCodeSummarizationFailed:
		return "context window summarization failed"
	case ErrorCodeInvalidSummary:
		return "context window summarizer returned an invalid summary"
	case ErrorCodeLimitUnresolved:
		return "context window limit could not be satisfied"
	case ErrorCodeTokenCountFailed:
		return "context window token count failed"
	default:
		return "context window preparation failed"
	}
}

// Unwrap exposes the underlying cause for errors.Is/As without displaying it.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is compares context-window errors by stable code.
func (e *Error) Is(target error) bool {
	other, ok := target.(*Error)
	return ok && e != nil && other != nil && e.Code == other.Code
}

// AsError returns the typed context-window error when present.
func AsError(err error) (*Error, bool) {
	var typed *Error
	if !errors.As(err, &typed) {
		return nil, false
	}
	return typed, true
}

// Policy controls provider-neutral context preparation. MaxChars and
// SummaryTargetChars must be positive. ProtectedTailSegments must be at least
// one so the latest user intent cannot be summarized away.
type Policy struct {
	WarnChars              int
	MaxChars               int
	MaxEvents              int
	StrictToolProtocol     bool
	StripInternalReasoning bool
	ToolOutputAnchor       transcript.AnchorSelector
	ProtectedHeadSegments  int
	ProtectedTailSegments  int
	SummaryTargetChars     int
	WarnInputTokens        int64
	MaxInputTokens         int64
}

// Request is one immutable context preparation input. PreviousSummary is
// explicit state supplied by a Host; this package does not retain it.
type Request struct {
	Model           string
	SystemPrompt    string
	Messages        llm.Conversation
	Tools           []llm.Tool
	PreviousSummary string
}

// SummaryRequest contains only the compactable middle window plus the
// previous Host-supplied summary. Implementations choose model and prompt.
type SummaryRequest struct {
	Messages        llm.Conversation
	PreviousSummary string
	TargetChars     int
}

// Summary is the provider-neutral output of a Host summarizer.
type Summary struct {
	Content string
}

// Summarizer is the only semantic compaction side-effect port. It must honor
// ctx cancellation and must not mutate request.Messages.
type Summarizer interface {
	Summarize(context.Context, SummaryRequest) (Summary, error)
}

// SummarizerFunc adapts a Host function to Summarizer.
type SummarizerFunc func(context.Context, SummaryRequest) (Summary, error)

// Summarize delegates to f.
func (f SummarizerFunc) Summarize(ctx context.Context, request SummaryRequest) (Summary, error) {
	if f == nil {
		return Summary{}, &Error{Code: ErrorCodeSummarizerUnavailable}
	}
	return f(ctx, request)
}

// Orchestrator performs stateless context preparation.
type Orchestrator struct {
	Policy       Policy
	Summarizer   Summarizer
	TokenCounter llm.TokenCounter
}

// Report contains bounded transformation facts and never includes message or
// summary content.
type Report struct {
	BeforeChars            int
	AfterChars             int
	Warned                 bool
	Sanitized              bool
	ProtocolAwareDrops     int
	CompactedToolOutputs   int
	CompactedHistoryBodies int
	SemanticSummaryUsed    bool
	SummarizedMessages     int
	ProtectedHeadSegments  int
	ProtectedTailSegments  int
	BeforeInputTokens      int64
	AfterInputTokens       int64
	InputTokenCountExact   bool
	InputTokenCountSource  string
}

// Result is the prepared conversation and the current semantic summary.
type Result struct {
	Messages llm.Conversation
	Summary  string
	Report   Report
}

// Prepare sanitizes and compacts one conversation. Error paths return a
// defensive copy of the original messages so callers never persist a partial
// replacement history.
func (o Orchestrator) Prepare(ctx context.Context, request Request) (Result, error) {
	original := cloneConversation(request.Messages)
	failure := func(report Report, err error) (Result, error) {
		report.AfterChars = estimate(request.SystemPrompt, original)
		if report.AfterInputTokens == 0 {
			report.AfterInputTokens = report.BeforeInputTokens
		}
		return Result{Messages: original, Summary: strings.TrimSpace(request.PreviousSummary), Report: report}, err
	}

	if ctx == nil {
		return failure(Report{}, &Error{Code: ErrorCodeInvalidPolicy})
	}
	if err := contextError(ctx.Err()); err != nil {
		return failure(Report{}, err)
	}
	if err := validatePolicy(o.Policy, o.TokenCounter); err != nil {
		return failure(Report{}, err)
	}

	report := Report{BeforeChars: estimate(request.SystemPrompt, original)}
	if tokenCount, err := o.countInput(ctx, request, original); err != nil {
		return failure(report, err)
	} else {
		report.BeforeInputTokens = tokenCount.Tokens
		report.AfterInputTokens = tokenCount.Tokens
		report.InputTokenCountExact = tokenCount.Exact
		report.InputTokenCountSource = displaySafeSource(tokenCount.Source)
	}
	working, diagnostic := transcript.Sanitize(original, transcript.SanitizePolicy{
		StrictToolProtocol:     o.Policy.StrictToolProtocol,
		StripInternalReasoning: o.Policy.StripInternalReasoning,
	})
	report.Sanitized = diagnostic.SynthesizedToolCallIDs > 0 ||
		diagnostic.RecoveredToolResults > 0 ||
		diagnostic.DowngradedToolResults > 0 ||
		diagnostic.StrippedReasoningMsgs > 0 ||
		diagnostic.MergedMessages > 0

	if o.Policy.MaxEvents > 0 {
		var drops int
		working, drops = transcript.PruneTailPreservingSystemPrefix(working, transcript.HistoryPolicy{
			MaxEvents:          o.Policy.MaxEvents,
			StrictToolProtocol: o.Policy.StrictToolProtocol,
		})
		report.ProtocolAwareDrops = drops
	}

	decision, err := o.evaluate(ctx, request, working)
	if err != nil {
		return failure(report, err)
	}
	applyAssessment(&report, decision)
	if !decision.Overflow {
		return Result{Messages: cloneConversation(working), Summary: effectivePreviousSummary(request.PreviousSummary, working), Report: report}, nil
	}

	working, report.CompactedToolOutputs = transcript.CompactToolOutputs(working, transcript.CompactionPolicy{
		MaxChars:         o.Policy.MaxChars,
		ToolOutputAnchor: o.Policy.ToolOutputAnchor,
	})
	decision, err = o.evaluate(ctx, request, working)
	if err != nil {
		return failure(report, err)
	}
	applyAssessment(&report, decision)
	if !decision.Overflow {
		return Result{Messages: cloneConversation(working), Summary: effectivePreviousSummary(request.PreviousSummary, working), Report: report}, nil
	}

	previousSummary, withoutOldSummaries := removeContextSummaries(effectivePreviousSummary(request.PreviousSummary, working), working)
	if o.Summarizer != nil {
		candidate, summary, summaryReport, err := o.semanticCompact(ctx, withoutOldSummaries, previousSummary)
		report.SummarizedMessages = summaryReport.SummarizedMessages
		report.ProtectedHeadSegments = summaryReport.ProtectedHeadSegments
		report.ProtectedTailSegments = summaryReport.ProtectedTailSegments
		if err != nil {
			return failure(report, err)
		}
		report.SemanticSummaryUsed = true
		decision, err = o.evaluate(ctx, request, candidate)
		if err != nil {
			return failure(report, err)
		}
		applyAssessment(&report, decision)
		if decision.Overflow {
			return failure(report, &Error{Code: ErrorCodeLimitUnresolved})
		}
		return Result{Messages: candidate, Summary: summary, Report: report}, nil
	}

	working, report.CompactedHistoryBodies = transcript.CompactHistoryBodies(working, o.Policy.MaxChars)
	decision, err = o.evaluate(ctx, request, working)
	if err != nil {
		return failure(report, err)
	}
	applyAssessment(&report, decision)
	if !decision.Overflow {
		return Result{Messages: working, Summary: previousSummary, Report: report}, nil
	}
	return failure(report, &Error{Code: ErrorCodeSummarizerUnavailable})
}

func (o Orchestrator) semanticCompact(ctx context.Context, messages llm.Conversation, previousSummary string) (llm.Conversation, string, Report, error) {
	prefixEnd := 0
	for prefixEnd < len(messages) && normalizeRole(messages[prefixEnd].Role) == "system" {
		prefixEnd++
	}
	prefix := cloneConversation(messages[:prefixEnd])
	segments := protocolSegments(messages[prefixEnd:])
	if len(segments) == 0 {
		return nil, "", Report{}, &Error{Code: ErrorCodeLimitUnresolved}
	}

	headCount := min(o.Policy.ProtectedHeadSegments, len(segments))
	tailStart := max(headCount, len(segments)-o.Policy.ProtectedTailSegments)
	if lastUser := latestUserSegment(segments); lastUser >= 0 && lastUser < tailStart {
		tailStart = max(headCount, lastUser)
	}
	if headCount >= tailStart {
		return nil, "", Report{}, &Error{Code: ErrorCodeLimitUnresolved}
	}

	middle := flattenSegments(segments[headCount:tailStart])
	if len(middle) == 0 {
		return nil, "", Report{}, &Error{Code: ErrorCodeLimitUnresolved}
	}
	summary, err := o.Summarizer.Summarize(ctx, SummaryRequest{
		Messages:        cloneConversation(middle),
		PreviousSummary: previousSummary,
		TargetChars:     o.Policy.SummaryTargetChars,
	})
	if err != nil {
		if ctxErr := contextError(ctx.Err()); ctxErr != nil {
			return nil, "", Report{}, ctxErr
		}
		return nil, "", Report{}, &Error{Code: ErrorCodeSummarizationFailed, Cause: err}
	}
	if ctxErr := contextError(ctx.Err()); ctxErr != nil {
		return nil, "", Report{}, ctxErr
	}
	content := strings.TrimSpace(summary.Content)
	if content == "" || len(content) > o.Policy.SummaryTargetChars {
		return nil, "", Report{}, &Error{Code: ErrorCodeInvalidSummary}
	}

	out := make(llm.Conversation, 0, len(prefix)+len(messages)-len(middle)+1)
	out = append(out, prefix...)
	out = append(out, flattenSegments(segments[:headCount])...)
	out = append(out, llm.Message{Role: "system", Content: summaryMarker + "\n" + content})
	out = append(out, flattenSegments(segments[tailStart:])...)
	return out, content, Report{
		SummarizedMessages:    len(middle),
		ProtectedHeadSegments: headCount,
		ProtectedTailSegments: len(segments) - tailStart,
	}, nil
}

func validatePolicy(policy Policy, counter llm.TokenCounter) error {
	if policy.MaxChars <= 0 || policy.SummaryTargetChars <= 0 || policy.ProtectedHeadSegments < 0 || policy.ProtectedTailSegments < 1 {
		return &Error{Code: ErrorCodeInvalidPolicy}
	}
	if policy.WarnChars < 0 || policy.MaxEvents < 0 {
		return &Error{Code: ErrorCodeInvalidPolicy}
	}
	if policy.WarnInputTokens < 0 || policy.MaxInputTokens < 0 ||
		(policy.MaxInputTokens > 0 && policy.WarnInputTokens >= policy.MaxInputTokens) ||
		((policy.WarnInputTokens > 0 || policy.MaxInputTokens > 0) && counter == nil) {
		return &Error{Code: ErrorCodeInvalidPolicy}
	}
	return nil
}

type assessment struct {
	Chars    int
	Tokens   llm.TokenCount
	Warn     bool
	Overflow bool
}

func (o Orchestrator) evaluate(ctx context.Context, request Request, messages llm.Conversation) (assessment, error) {
	charDecision := transcript.Evaluate(transcript.EstimateInput{
		SystemPrompt: request.SystemPrompt,
		Messages:     messages,
		RoleAware:    true,
	}, transcript.GuardPolicy{WarnChars: o.Policy.WarnChars, MaxChars: o.Policy.MaxChars})
	result := assessment{Chars: charDecision.EstimatedChars, Warn: charDecision.Warn, Overflow: charDecision.Overflow}
	count, err := o.countInput(ctx, request, messages)
	if err != nil {
		return assessment{}, err
	}
	result.Tokens = count
	if o.Policy.WarnInputTokens > 0 && count.Tokens >= o.Policy.WarnInputTokens {
		result.Warn = true
	}
	if o.Policy.MaxInputTokens > 0 && count.Tokens > o.Policy.MaxInputTokens {
		result.Overflow = true
	}
	return result, nil
}

func (o Orchestrator) countInput(ctx context.Context, request Request, messages llm.Conversation) (llm.TokenCount, error) {
	if o.Policy.WarnInputTokens == 0 && o.Policy.MaxInputTokens == 0 {
		return llm.TokenCount{}, nil
	}
	count, err := o.TokenCounter.CountInput(ctx, llm.TokenCountRequest{
		Model: request.Model, System: request.SystemPrompt, Messages: cloneConversation(messages), Tools: append([]llm.Tool(nil), request.Tools...),
	})
	if err != nil {
		if ctxErr := contextError(ctx.Err()); ctxErr != nil {
			return llm.TokenCount{}, ctxErr
		}
		return llm.TokenCount{}, &Error{Code: ErrorCodeTokenCountFailed, Cause: err}
	}
	if count.Tokens < 0 {
		return llm.TokenCount{}, &Error{Code: ErrorCodeTokenCountFailed}
	}
	return count, nil
}

func applyAssessment(report *Report, value assessment) {
	report.AfterChars = value.Chars
	report.Warned = report.Warned || value.Warn
	report.AfterInputTokens = value.Tokens.Tokens
	if value.Tokens.Source != "" {
		report.InputTokenCountExact = value.Tokens.Exact
		report.InputTokenCountSource = displaySafeSource(value.Tokens.Source)
	}
}

func displaySafeSource(source string) string {
	source = strings.TrimSpace(strings.NewReplacer("\n", " ", "\r", " ", "\t", " ").Replace(source))
	if len(source) > 128 {
		return source[:128]
	}
	return source
}

func contextError(err error) error {
	if err == nil {
		return nil
	}
	code := ErrorCodeCanceled
	if errors.Is(err, context.DeadlineExceeded) {
		code = ErrorCodeDeadlineExceeded
	}
	return &Error{Code: code, Cause: err}
}

func estimate(systemPrompt string, messages llm.Conversation) int {
	return transcript.EstimateChars(transcript.EstimateInput{
		SystemPrompt: systemPrompt,
		Messages:     messages,
		RoleAware:    true,
	})
}

func effectivePreviousSummary(explicit string, messages llm.Conversation) string {
	if trimmed := strings.TrimSpace(explicit); trimmed != "" {
		return trimmed
	}
	for index := len(messages) - 1; index >= 0; index-- {
		if summary, ok := contextSummaryContent(messages[index]); ok {
			return summary
		}
	}
	return ""
}

func removeContextSummaries(previous string, messages llm.Conversation) (string, llm.Conversation) {
	out := make(llm.Conversation, 0, len(messages))
	for _, message := range messages {
		if summary, ok := contextSummaryContent(message); ok {
			if strings.TrimSpace(previous) == "" {
				previous = summary
			}
			continue
		}
		out = append(out, cloneMessage(message))
	}
	return strings.TrimSpace(previous), out
}

func contextSummaryContent(message llm.Message) (string, bool) {
	content := strings.TrimSpace(message.Content)
	if normalizeRole(message.Role) != "system" || !strings.HasPrefix(content, summaryMarker) {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(content, summaryMarker)), true
}

func protocolSegments(messages llm.Conversation) []llm.Conversation {
	segments := make([]llm.Conversation, 0, len(messages))
	for index := 0; index < len(messages); {
		message := messages[index]
		segment := llm.Conversation{cloneMessage(message)}
		index++
		if normalizeRole(message.Role) == "assistant" && len(message.ToolCalls) > 0 {
			for index < len(messages) && normalizeRole(messages[index].Role) == "tool" {
				segment = append(segment, cloneMessage(messages[index]))
				index++
			}
		}
		segments = append(segments, segment)
	}
	return segments
}

func latestUserSegment(segments []llm.Conversation) int {
	for index := len(segments) - 1; index >= 0; index-- {
		for _, message := range segments[index] {
			if normalizeRole(message.Role) == "user" {
				return index
			}
		}
	}
	return -1
}

func flattenSegments(segments []llm.Conversation) llm.Conversation {
	var total int
	for _, segment := range segments {
		total += len(segment)
	}
	out := make(llm.Conversation, 0, total)
	for _, segment := range segments {
		out = append(out, cloneConversation(segment)...)
	}
	return out
}

func cloneConversation(messages llm.Conversation) llm.Conversation {
	if messages == nil {
		return nil
	}
	out := make(llm.Conversation, len(messages))
	for index, message := range messages {
		out[index] = cloneMessage(message)
	}
	return out
}

func cloneMessage(message llm.Message) llm.Message {
	cloned := message
	if message.ToolCalls != nil {
		cloned.ToolCalls = append([]llm.FunctionCall(nil), message.ToolCalls...)
	}
	return cloned
}

func normalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
