package agentx

import "context"

// ExecutionAdapter 是 Client 使用的最窄执行接缝。
//
// Client 会串行调用 Run；Shutdown 可以与 Run 并发，以便实现取消和收敛正在
// 执行的工作。实现必须让 Shutdown 支持重复调用，并自行持有底层 Runtime 的构造、
// 状态、结果提取和错误分类语义。
type ExecutionAdapter interface {
	Run(context.Context, AdapterRunRequest) (*AdapterRunResult, error)
	Shutdown(context.Context) error
	ClassifyError(error) ErrorCode
}

// AdapterRunRequest 是传递给 ExecutionAdapter 的底层无关输入。
type AdapterRunRequest struct {
	Input     string
	SessionID string
}

// AdapterRunResult 是 ExecutionAdapter 返回的底层无关结果。
//
// Status 是 adapter 对底层执行状态的观察；Client 会将其映射为稳定的公共状态、
// blocker 和 next action。
type AdapterRunResult struct {
	RunID     string
	SessionID string
	Status    string
	Reply     string
}
