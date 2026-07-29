package agentx

import (
	"context"
	"strings"
	"sync/atomic"
)

// Config 绑定一个由 host 构造的底层无关 ExecutionAdapter。
//
// New 成功后，调用方将 Adapter 的 Run 和 Shutdown 独占访问权交给 Client，不应
// 再直接调用该 Adapter。
type Config struct {
	Adapter ExecutionAdapter
	Profile ExecutionProfile
}

// RunRequest 是最小同步执行输入。
type RunRequest struct {
	Input     string
	SessionID string
}

// RunResult 是一次 owner 执行的 display-safe 投影。
type RunResult struct {
	RunID      string
	SessionID  string
	Status     string
	Reply      string
	Evidence   []string
	Blockers   []string
	NextAction string
	Profile    ExecutionProfile
}

// Client 是 ExecutionAdapter 之上的薄公共合同。
//
// Client 方法可并发调用，但同一 Client 的重叠 Run 会被串行化。Shutdown 可与
// 当前 Run 并发；Shutdown 一旦开始，新的 Run 会返回 CodeClientClosed。
type Client struct {
	adapter      ExecutionAdapter
	profile      ExecutionProfile
	runGate      chan struct{}
	shutdownGate chan struct{}
	closing      atomic.Bool
}

// New 校验 Config 并创建 Client。
func New(config Config) (*Client, error) {
	if config.Adapter == nil {
		return nil, newError(CodeInvalidArgument, nil)
	}
	profile, err := resolveProfile(config.Profile)
	if err != nil {
		return nil, err
	}
	return &Client{
		adapter:      config.Adapter,
		profile:      profile,
		runGate:      make(chan struct{}, 1),
		shutdownGate: make(chan struct{}, 1),
	}, nil
}

// Run 执行已配置的同步 adapter 路径。
func (c *Client) Run(ctx context.Context, request RunRequest) (RunResult, error) {
	result := RunResult{}
	if c != nil {
		result.Profile = c.profile
	}
	if ctx == nil {
		err := newError(CodeInvalidArgument, nil)
		return resultForError(result, err), err
	}
	if c == nil || c.adapter == nil || c.closing.Load() {
		err := newError(CodeClientClosed, nil)
		return resultForError(result, err), err
	}
	if strings.TrimSpace(request.Input) == "" {
		err := newError(CodeInvalidArgument, nil)
		return resultForError(result, err), err
	}

	select {
	case c.runGate <- struct{}{}:
		defer func() { <-c.runGate }()
	case <-ctx.Done():
		err := mapRunError(c.adapter, ctx.Err())
		return resultForError(result, err), err
	}
	if c.closing.Load() {
		err := newError(CodeClientClosed, nil)
		return resultForError(result, err), err
	}

	output, runErr := c.adapter.Run(ctx, AdapterRunRequest{
		Input:     request.Input,
		SessionID: strings.TrimSpace(request.SessionID),
	})
	result = resultFromOutput(output, c.profile)
	if runErr != nil {
		err := mapRunError(c.adapter, runErr)
		return resultForError(result, err), err
	}
	result.Status, result.Blockers, result.NextAction = mapAdapterStatus(output)
	return result, nil
}

// Shutdown 开始或继续一次有界、幂等的 adapter 关闭过程。
//
// Adapter.Shutdown 必须支持在前一次调用因 context 到期返回后继续调用。
func (c *Client) Shutdown(ctx context.Context) error {
	if ctx == nil {
		return newError(CodeInvalidArgument, nil)
	}
	if c == nil || c.adapter == nil {
		return newError(CodeClientClosed, nil)
	}
	if err := ctx.Err(); err != nil {
		return newError(CodeShutdownFailed, err)
	}

	c.closing.Store(true)
	select {
	case c.shutdownGate <- struct{}{}:
		defer func() { <-c.shutdownGate }()
	case <-ctx.Done():
		return newError(CodeShutdownFailed, ctx.Err())
	}
	if err := c.adapter.Shutdown(ctx); err != nil {
		return newError(CodeShutdownFailed, err)
	}
	return nil
}

func resultFromOutput(output *AdapterRunResult, profile ExecutionProfile) RunResult {
	result := RunResult{Profile: profile}
	if output == nil {
		return result
	}
	result.RunID = strings.TrimSpace(output.RunID)
	result.Reply = output.Reply
	result.SessionID = strings.TrimSpace(output.SessionID)
	result.Evidence = identityEvidence(result.RunID, result.SessionID)
	return result
}

func mapAdapterStatus(output *AdapterRunResult) (status string, blockers []string, nextAction string) {
	if output == nil {
		return "failed", []string{string(CodeExecutionFailed)}, "inspect_owner_diagnostics"
	}
	switch strings.ToLower(strings.TrimSpace(output.Status)) {
	case "", "complete", "completed", "success", "succeeded":
		return "completed", nil, ""
	case "blocked", "incomplete", "partial", "review_required":
		return "blocked", []string{"execution_incomplete"}, "resolve_execution_blocker"
	case "canceled", "cancelled":
		return "canceled", []string{string(CodeCanceled)}, "caller_decides_retry"
	case "error", "failed", "failure":
		return "failed", []string{string(CodeExecutionFailed)}, "inspect_owner_diagnostics"
	default:
		return "failed", []string{string(CodeExecutionFailed)}, "inspect_owner_diagnostics"
	}
}

func resultForError(result RunResult, err *Error) RunResult {
	if err == nil {
		return result
	}
	switch err.Code {
	case CodeCanceled:
		result.Status = "canceled"
	default:
		result.Status = "failed"
	}
	result.Blockers = []string{string(err.Code)}
	result.NextAction = nextActionForCode(err.Code)
	return result
}

func identityEvidence(runID, sessionID string) []string {
	var evidence []string
	if runID != "" {
		evidence = append(evidence, "run:"+runID)
	}
	if sessionID != "" {
		evidence = append(evidence, "session:"+sessionID)
	}
	return evidence
}
