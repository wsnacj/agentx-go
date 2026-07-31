// Package orchestration 提供与 lowering、executor 和存储 backend 无关的
// Workflow run orchestration mechanism。
//
// 当前 package 处于 Experimental/private validation。它组合 canonical
// bindingstate、transition、journal 和 nodeexec owner，但不拥有具体
// RunStore、tool/model mapping、provider、credential 或 Scene policy。
package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	workflowbindingstate "github.com/wsnacj/agentx-go/runtime/workflow/bindingstate"
	workflowjournal "github.com/wsnacj/agentx-go/runtime/workflow/journal"
	workflownodeexec "github.com/wsnacj/agentx-go/runtime/workflow/nodeexec"
	workflowtransition "github.com/wsnacj/agentx-go/runtime/workflow/transition"
)

// Plan is the portable projection produced by a host-owned lowerer.
type Plan struct {
	WorkflowID  string
	Version     string
	EntryNode   string
	NodeIDs     []string
	Nodes       map[string]PlannedNode
	Edges       []workflow.EdgeSpec
	StateSchema []workflow.StateSlotSpec
}

// PlannedNode combines a canonical node spec with its host-lowered call.
type PlannedNode struct {
	Spec           workflow.NodeSpec
	Call           workflownodeexec.Call
	Kind           workflow.NodeKind
	ExecutionMode  workflow.ExecutionMode
	OriginalConfig map[string]any
}

// Inputs contains portable run identity and binding roots.
type Inputs struct {
	RunID        string
	CaseID       string
	BranchID     string
	InitialState map[string]any
	CaseInput    map[string]any
	SessionInput map[string]any
}

// NodeExecution is the single substrate-neutral execution capability required
// by the run loop.
type NodeExecution interface {
	Execute(context.Context, workflownodeexec.Request) (workflownodeexec.Outcome, error)
}

// Dependencies contains host-owned capabilities. Journal and NodeExecution
// remain concrete canonical mechanisms; identity, clock and error display
// policy are injected by the host.
type Dependencies struct {
	Journal            *workflowjournal.Journal
	NodeExecution      NodeExecution
	NewNodeExecutionID func() string
	NowUnixMilli       func() int64
	ProjectError       func(error) string
}

// NodeResult is the portable observable result of one planned node.
type NodeResult struct {
	NodeID                string
	NodeExecutionID       string
	Call                  workflownodeexec.Call
	Output                string
	Error                 string
	FinalStatus           string
	StopReason            string
	ExecutionContractID   string
	ExecutionContractDiff []string
	Termination           *workflownodeexec.Termination
	DelegatedExecution    *workflownodeexec.DelegatedExecution
	ChildNodeExecutions   []workflownodeexec.NodeExecutionProjection
}

// Result contains the portable run result. Hosts may add compatibility-only
// projections such as lowered debug data after Run returns.
type Result struct {
	RunID       string
	FinalNode   string
	FinalStatus string
	StopReason  string
	NodeResults []NodeResult
	NodeOutput  map[string]string
	State       map[string]any
}

// Run executes one already-lowered portable Workflow plan.
func Run(ctx context.Context, plan Plan, inputs Inputs, dependencies Dependencies) (Result, error) {
	if dependencies.NodeExecution == nil {
		return Result{}, fmt.Errorf("workflow orchestration: node execution is required")
	}
	if dependencies.NowUnixMilli == nil {
		return Result{}, fmt.Errorf("workflow orchestration: clock is required")
	}
	runID, err := dependencies.Journal.EnsureRun(ctx, workflowjournal.EnsureRunRequest{
		RunID:           inputs.RunID,
		CaseID:          inputs.CaseID,
		WorkflowID:      plan.WorkflowID,
		WorkflowVersion: plan.Version,
	})
	if err != nil {
		return Result{}, err
	}
	state := workflowbindingstate.New(workflowbindingstate.Inputs{
		InitialState: inputs.InitialState,
		SessionInput: inputs.SessionInput,
		CaseInput:    inputs.CaseInput,
	})
	result := Result{
		RunID:      runID,
		NodeOutput: map[string]string{},
		State:      state.State(),
	}
	if err := dependencies.Journal.AppendRunStart(ctx, workflowjournal.StartRunEventRequest{
		RunID:      runID,
		BranchID:   inputs.BranchID,
		WorkflowID: plan.WorkflowID,
		EntryNode:  plan.EntryNode,
	}); err != nil {
		return result, err
	}

	machine := workflowtransition.New(workflowtransition.Plan{
		EntryNode: plan.EntryNode,
		NodeIDs:   plannedNodeIDs(plan),
		Edges:     plan.Edges,
	})
	for {
		current, transitionErr := machine.Enter()
		if transitionErr != nil {
			if finalizeErr := finishRun(ctx, plan, runID, dependencies, "failed", now(dependencies), transitionErr.Error()); finalizeErr != nil {
				return result, finalizeErr
			}
			return result, transitionErr
		}
		if current == "" {
			break
		}
		node, ok := plan.Nodes[current]
		if !ok {
			missingErr := fmt.Errorf("workflow: node spec %q missing", current)
			if finalizeErr := finishRun(ctx, plan, runID, dependencies, "failed", now(dependencies), missingErr.Error()); finalizeErr != nil {
				return result, finalizeErr
			}
			return result, missingErr
		}

		nodeExecutionID, err := newNodeExecutionID(dependencies)
		if err != nil {
			return result, err
		}
		startedAt := now(dependencies)
		arguments, bindErr := state.MaterializeArguments(current, node.Call.Arguments, node.Spec.Inputs)
		if bindErr != nil {
			if finalizeErr := finishRun(ctx, plan, runID, dependencies, "failed", startedAt, bindErr.Error()); finalizeErr != nil {
				return result, finalizeErr
			}
			return result, bindErr
		}
		call := node.Call
		call.Arguments = arguments
		inputStateRef, err := dependencies.Journal.StartNode(ctx, workflowjournal.StartNodeRequest{
			Node: workflowjournal.NodeExecution{
				NodeExecutionID: nodeExecutionID,
				RunID:           runID,
				BranchID:        inputs.BranchID,
				NodeID:          current,
				Kind:            string(node.Kind),
				Status:          "running",
				Attempt:         1,
				StartedAt:       startedAt,
			},
			State: state.State(),
			EventPayload: map[string]any{
				"node_id":   current,
				"kind":      node.Kind,
				"tool_name": call.Name,
			},
		})
		if err != nil {
			return result, err
		}

		outcome, callErr := dependencies.NodeExecution.Execute(ctx, workflownodeexec.Request{
			NodeExecutionID: nodeExecutionID,
			NodeID:          current,
			Kind:            node.Kind,
			ExecutionMode:   node.ExecutionMode,
			OriginalConfig:  node.OriginalConfig,
			Spec:            node.Spec,
			Call:            call,
		})
		finishedAt := now(dependencies)
		nodeFinalStatus := workflowtransition.NormalizeFinalStatus(outcome.FinalStatus, callErr != nil)
		nodeResult := NodeResult{
			NodeID:                current,
			NodeExecutionID:       nodeExecutionID,
			Call:                  call,
			Output:                outcome.Output,
			FinalStatus:           nodeFinalStatus,
			StopReason:            outcome.StopReason,
			ExecutionContractID:   outcome.ExecutionContractID,
			ExecutionContractDiff: append([]string(nil), outcome.ExecutionContractDiff...),
			Termination:           cloneTermination(outcome.Termination),
			DelegatedExecution:    cloneDelegatedExecution(outcome.DelegatedExecution),
			ChildNodeExecutions:   cloneChildNodeExecutions(outcome.ChildNodeExecutions),
		}
		status := nodeFinalStatus
		if status == "" {
			status = "completed"
		}
		runtimeNode := workflowbindingstate.NewNodeResult(nodeFinalStatus, outcome.Output, "")
		if callErr != nil {
			nodeResult.Error = projectError(dependencies, callErr)
			nodeResult.FinalStatus = "failed"
			if nodeResult.StopReason == "" {
				nodeResult.StopReason = nodeResult.Error
			}
			status = "failed"
			runtimeNode = workflowbindingstate.NewNodeResult("failed", outcome.Output, nodeResult.Error)
		}
		if callErr == nil && status == "completed" {
			if err := state.ApplyNodeOutputs(current, node.Spec.Outputs, runtimeNode); err != nil {
				if finalizeErr := finishRun(ctx, plan, runID, dependencies, "failed", finishedAt, projectError(dependencies, err)); finalizeErr != nil {
					return result, finalizeErr
				}
				return result, err
			}
		}
		result.State = state.State()
		result.NodeResults = append(result.NodeResults, nodeResult)
		if outcome.Output != "" {
			result.NodeOutput[current] = outcome.Output
		}
		payload := nodeFinishPayload(node, nodeResult, status)
		if _, err := dependencies.Journal.FinishNode(ctx, workflowjournal.FinishNodeRequest{
			Node: workflowjournal.NodeExecution{
				NodeExecutionID:           nodeExecutionID,
				RunID:                     runID,
				BranchID:                  inputs.BranchID,
				NodeID:                    current,
				Kind:                      string(node.Kind),
				Status:                    status,
				Attempt:                   1,
				InputStateRef:             inputStateRef,
				ExecutionContractID:       nodeResult.ExecutionContractID,
				ExecutionContractDiffJSON: marshalContractDiff(nodeResult.ExecutionContractDiff),
				TerminationJSON:           marshalTermination(nodeResult.Termination),
				DelegatedExecutionJSON:    marshalDelegatedExecution(nodeResult.DelegatedExecution),
				StartedAt:                 startedAt,
				FinishedAt:                finishedAt,
			},
			State:        state.State(),
			EventPayload: payload,
		}); err != nil {
			return result, err
		}

		result.FinalNode = current
		if status == "incomplete" {
			result.FinalStatus = "incomplete"
			result.StopReason = firstNonEmpty(nodeResult.StopReason, "incomplete")
			if finalizeErr := finishRun(ctx, plan, runID, dependencies, "incomplete", finishedAt, result.StopReason); finalizeErr != nil {
				return result, finalizeErr
			}
			return result, nil
		}
		if status == "failed" {
			result.FinalStatus = "failed"
			result.StopReason = firstNonEmpty(nodeResult.StopReason, nodeResult.Error, "failed")
			next, nextErr := machine.Advance(workflowtransition.TriggerFailure)
			if nextErr != nil {
				return result, nextErr
			}
			if next == "" {
				if finalizeErr := finishRun(ctx, plan, runID, dependencies, "failed", finishedAt, result.StopReason); finalizeErr != nil {
					return result, finalizeErr
				}
				if callErr != nil {
					return result, fmt.Errorf("workflow: node %q failed: %w", current, callErr)
				}
				return result, fmt.Errorf("workflow: node %q failed: %s", current, result.StopReason)
			}
			continue
		}

		next, nextErr := machine.Advance(workflowtransition.TriggerSuccess)
		if nextErr != nil {
			return result, nextErr
		}
		if next == "" {
			if err := state.ValidateRequiredSlots(plan.StateSchema); err != nil {
				if finalizeErr := finishRun(ctx, plan, runID, dependencies, "failed", finishedAt, projectError(dependencies, err)); finalizeErr != nil {
					return result, finalizeErr
				}
				return result, err
			}
			result.FinalStatus = "completed"
			result.StopReason = "completed"
			if err := finishRun(ctx, plan, runID, dependencies, "completed", finishedAt, ""); err != nil {
				return result, err
			}
			return result, nil
		}
	}

	result.FinalStatus = "completed"
	result.StopReason = "completed"
	if err := finishRun(ctx, plan, runID, dependencies, "completed", now(dependencies), ""); err != nil {
		return result, err
	}
	return result, nil
}

func nodeFinishPayload(node PlannedNode, result NodeResult, status string) map[string]any {
	payload := map[string]any{
		"node_id":     result.NodeID,
		"kind":        node.Kind,
		"tool_name":   result.Call.Name,
		"status":      status,
		"error":       result.Error,
		"stop_reason": result.StopReason,
	}
	if result.ExecutionContractID != "" {
		payload["execution_contract_id"] = result.ExecutionContractID
	}
	if len(result.ExecutionContractDiff) > 0 {
		payload["execution_contract_diff"] = append([]string(nil), result.ExecutionContractDiff...)
	}
	if result.Termination != nil {
		payload["termination"] = cloneTermination(result.Termination)
	}
	if result.DelegatedExecution != nil {
		payload["delegated_execution"] = cloneDelegatedExecution(result.DelegatedExecution)
	}
	if len(result.ChildNodeExecutions) > 0 {
		payload["child_node_executions"] = cloneChildNodeExecutions(result.ChildNodeExecutions)
	}
	return payload
}

func finishRun(ctx context.Context, plan Plan, runID string, dependencies Dependencies, status string, finishedAt int64, errorText string) error {
	return dependencies.Journal.FinishRun(ctx, workflowjournal.FinishRunRequest{
		RunID:           runID,
		WorkflowID:      plan.WorkflowID,
		WorkflowVersion: plan.Version,
		Status:          status,
		FinishedAt:      finishedAt,
		ErrorText:       errorText,
	})
}

func newNodeExecutionID(dependencies Dependencies) (string, error) {
	if dependencies.NewNodeExecutionID == nil {
		return "", fmt.Errorf("workflow orchestration: node execution id source is required")
	}
	value := dependencies.NewNodeExecutionID()
	if value == "" {
		return "", fmt.Errorf("workflow orchestration: node execution id source returned empty value")
	}
	return value, nil
}

func now(dependencies Dependencies) int64 {
	return dependencies.NowUnixMilli()
}

func projectError(dependencies Dependencies, err error) string {
	if err == nil {
		return ""
	}
	if dependencies.ProjectError != nil {
		return dependencies.ProjectError(err)
	}
	return err.Error()
}

func plannedNodeIDs(plan Plan) []string {
	if plan.NodeIDs != nil {
		return append([]string(nil), plan.NodeIDs...)
	}
	nodeIDs := make([]string, 0, len(plan.Nodes))
	for nodeID := range plan.Nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	return nodeIDs
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func cloneTermination(in *workflownodeexec.Termination) *workflownodeexec.Termination {
	if in == nil {
		return nil
	}
	out := *in
	if out.Kind == "" &&
		out.CheckpointStage == "" &&
		out.CheckpointError == "" &&
		out.EventName == "" &&
		out.EventStatus == "" &&
		!out.ReplyPersisted {
		return nil
	}
	return &out
}

func cloneDelegatedExecution(in *workflownodeexec.DelegatedExecution) *workflownodeexec.DelegatedExecution {
	if in == nil {
		return nil
	}
	out := *in
	if len(in.Rounds) > 0 {
		out.Rounds = make([]workflownodeexec.DelegatedRound, 0, len(in.Rounds))
		for _, round := range in.Rounds {
			if round.NodeExecID == "" && round.Round <= 0 && round.OutcomeKind == "" &&
				round.StopReason == "" && round.ToolCalls <= 0 && round.ToolRuns <= 0 {
				continue
			}
			out.Rounds = append(out.Rounds, round)
		}
	}
	if out.Driver == "" && out.OutcomeKind == "" && out.RoundCount <= 0 &&
		out.ToolCalls <= 0 && len(out.Rounds) == 0 {
		return nil
	}
	return &out
}

func cloneChildNodeExecutions(in []workflownodeexec.NodeExecutionProjection) []workflownodeexec.NodeExecutionProjection {
	if len(in) == 0 {
		return nil
	}
	out := make([]workflownodeexec.NodeExecutionProjection, 0, len(in))
	for _, item := range in {
		cloned := item
		cloned.ExecutionContractDiff = append([]string(nil), item.ExecutionContractDiff...)
		cloned.Termination = cloneChildTermination(item.Termination)
		cloned.DelegatedExecution = cloneChildDelegatedExecution(item.DelegatedExecution)
		cloned.ChildNodeExecutions = cloneChildNodeExecutions(item.ChildNodeExecutions)
		out = append(out, cloned)
	}
	return out
}

func cloneChildTermination(in *workflownodeexec.Termination) *workflownodeexec.Termination {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneChildDelegatedExecution(in *workflownodeexec.DelegatedExecution) *workflownodeexec.DelegatedExecution {
	if in == nil {
		return nil
	}
	out := *in
	out.Rounds = nil
	for _, round := range in.Rounds {
		if strings.TrimSpace(round.NodeExecID) == "" && round.Round == 0 &&
			strings.TrimSpace(round.OutcomeKind) == "" &&
			strings.TrimSpace(round.StopReason) == "" &&
			round.ToolCalls == 0 && round.ToolRuns == 0 {
			continue
		}
		out.Rounds = append(out.Rounds, round)
	}
	if strings.TrimSpace(out.Driver) == "" && strings.TrimSpace(out.OutcomeKind) == "" &&
		out.RoundCount == 0 && out.ToolCalls == 0 && len(out.Rounds) == 0 {
		return nil
	}
	return &out
}

func marshalTermination(in *workflownodeexec.Termination) string {
	if in == nil {
		return ""
	}
	payload, err := json.Marshal(cloneTermination(in))
	if err != nil {
		return ""
	}
	return string(payload)
}

func marshalDelegatedExecution(in *workflownodeexec.DelegatedExecution) string {
	if in == nil {
		return ""
	}
	payload, err := json.Marshal(cloneDelegatedExecution(in))
	if err != nil {
		return ""
	}
	return string(payload)
}

func marshalContractDiff(in []string) string {
	if len(in) == 0 {
		return ""
	}
	changed := make([]string, 0, len(in))
	for _, item := range in {
		if item != "" {
			changed = append(changed, item)
		}
	}
	if len(changed) == 0 {
		return ""
	}
	payload, err := json.Marshal(map[string]any{"changed_fields": changed})
	if err != nil {
		return ""
	}
	return string(payload)
}
