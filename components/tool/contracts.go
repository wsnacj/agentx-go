// Package tool contains provider-neutral tool definition and execution contracts.
package tool

import (
	"context"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

// Definition is the canonical function-tool declaration used on the model wire.
// It aliases the existing LLM contract so callers do not maintain two DTO graphs.
type Definition = llm.Tool

// Function is the callable function schema within a Definition.
type Function = llm.Function

// Choice is the model-facing tool selection contract.
type Choice = llm.ToolChoice

// Call is one normalized tool invocation emitted by a model or host.
type Call = llm.FunctionCall

// Result is the current text/JSON-compatible result returned to the model.
type Result = string

// Handler executes one registered tool call.
type Handler func(context.Context, Call) (Result, error)

// Executor executes tool calls without prescribing a concrete catalog.
type Executor interface {
	Execute(context.Context, Call) (Result, error)
}

// DefinitionProvider exposes a stable snapshot of available tool declarations.
type DefinitionProvider interface {
	Definitions() []Definition
}

// Registrar accepts one tool definition and its implementation.
type Registrar interface {
	Register(Definition, Handler)
}
