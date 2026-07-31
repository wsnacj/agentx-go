// Package execution 提供根 AgentX Client 与 concrete execution host 之间的
// substrate-neutral Run dispatch、result assembly和 lifecycle adapter。
//
// 当前 package 处于 Experimental/private validation。它不构造模型、工具、
// Workflow、RunStore或 Scene；这些能力由 Host拥有。
package execution

import (
	"context"
	"errors"

	agentx "github.com/wsnacj/agentx-go"
)

// Request 是传递给 concrete Host 的底层无关 Run 输入。
type Request struct {
	Input     string
	SessionID string
}

// Result 是 concrete Host 返回的底层无关 Run 结果。
type Result struct {
	RunID     string
	SessionID string
	Status    string
	Reply     string
}

// Host 持有 concrete execution、资源关闭和错误分类语义。
//
// Run 必须原样遵守调用方 context。Shutdown 必须支持重复调用，并允许前一次调用
// 因 context到期返回后继续收敛。ClassifyError只返回根 agentx合同支持的 code。
type Host interface {
	Run(context.Context, Request) (*Result, error)
	Shutdown(context.Context) error
	ClassifyError(error) agentx.ErrorCode
}

// Runtime 把一个 Host组合为根 agentx.ExecutionAdapter。
//
// Runtime不增加并发 gate；并发与关闭后调用语义由拥有它的 agentx.Client负责。
// 一个 Runtime成功交给 Client后，调用方不应再直接调用 Runtime或 Host。
type Runtime struct {
	host Host
}

var _ agentx.ExecutionAdapter = (*Runtime)(nil)

// New 校验 Host并构造一个 execution Runtime。
func New(host Host) (*Runtime, error) {
	if host == nil {
		return nil, errors.New("agentx execution: host is required")
	}
	return &Runtime{host: host}, nil
}

// Run 将一个 adapter request确定性分派给 Host，并组装底层无关 adapter result。
//
// context、文本、partial result和 error identity均不被包装或改写。请求校验、
// SessionID规范化和稳定公共状态映射由根 agentx.Client负责。
func (rt *Runtime) Run(ctx context.Context, request agentx.AdapterRunRequest) (*agentx.AdapterRunResult, error) {
	if rt == nil || rt.host == nil {
		return nil, errors.New("agentx execution: runtime is required")
	}
	result, err := rt.host.Run(ctx, Request{
		Input:     request.Input,
		SessionID: request.SessionID,
	})
	if result == nil {
		return nil, err
	}
	return &agentx.AdapterRunResult{
		RunID:     result.RunID,
		SessionID: result.SessionID,
		Status:    result.Status,
		Reply:     result.Reply,
	}, err
}

// Shutdown 把有界关闭请求原样传递给 Host。
func (rt *Runtime) Shutdown(ctx context.Context) error {
	if rt == nil || rt.host == nil {
		return errors.New("agentx execution: runtime is required")
	}
	return rt.host.Shutdown(ctx)
}

// ClassifyError 委托 Host把 concrete error映射到根 agentx错误码。
func (rt *Runtime) ClassifyError(err error) agentx.ErrorCode {
	if rt == nil || rt.host == nil {
		return agentx.CodeExecutionFailed
	}
	return rt.host.ClassifyError(err)
}
