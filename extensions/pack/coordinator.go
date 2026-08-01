package pack

import (
	"errors"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

// Coordinator组合 portable Pack机制与 Host拥有的 Workflow validation policy。
//
// Coordinator不持有 Runner、Scene、provider、credential、memory/eval backend或
// 真实副作用。构造后可安全供多个 goroutine并发使用；注入的 Validator也必须支持
// 与调用方相同的并发边界。
type Coordinator struct {
	validator workflow.Validator
	lowerer   ToolArgumentLowerer
}

// ToolArgumentLowerer把一个已经物化的 tool node投影为 arguments JSON。
// 具体 tool config mapping与默认值继续由 Host拥有。
type ToolArgumentLowerer interface {
	LowerToolArguments(workflow.NodeSpec) (string, error)
}

// NewCoordinator创建一个 Pack协调器。validator和 lowerer必须由 Host显式提供；
// nil会 fail closed，避免把具体 Runtime能力策略隐式固化在扩展包中。
func NewCoordinator(validator workflow.Validator, lowerer ToolArgumentLowerer) (*Coordinator, error) {
	if validator == nil {
		return nil, errors.New("agentx pack: workflow validator is required")
	}
	if lowerer == nil {
		return nil, errors.New("agentx pack: tool argument lowerer is required")
	}
	return &Coordinator{validator: validator, lowerer: lowerer}, nil
}

func (c *Coordinator) toolArgumentLowerer() (ToolArgumentLowerer, error) {
	if c == nil || c.lowerer == nil {
		return nil, errors.New("agentx pack: coordinator is required")
	}
	return c.lowerer, nil
}

func (c *Coordinator) workflowValidator() (workflow.Validator, error) {
	if c == nil || c.validator == nil {
		return nil, errors.New("agentx pack: coordinator is required")
	}
	return c.validator, nil
}
