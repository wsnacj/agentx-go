package hostkit

import (
	"context"
	"strings"

	research "github.com/wsnacj/agentx-go/scenes/companyresearch"
)

type TaskExecutor func(context.Context, CompanyResearchTaskExecutionRequest) (CompanyResearchTaskExecutionResult, error)

type CompanyResearchTaskExecutionRequest struct {
	Plan              research.CompanyResearchTaskPlan     `json:"plan,omitempty"`
	Task              research.CompanyResearchTaskSpec     `json:"task,omitempty"`
	Intent            research.CompanyResearchIntent       `json:"intent,omitempty"`
	SubjectResolution *research.SubjectResolution          `json:"subject_resolution,omitempty"`
	Params            map[string]any                       `json:"params,omitempty"`
	Evidence          research.CompanyResearchEvidence     `json:"evidence,omitempty"`
	PreviousResults   []research.CompanyResearchTaskResult `json:"previous_results,omitempty"`
}

type CompanyResearchTaskExecutionResult struct {
	Handled    bool                                `json:"handled,omitempty"`
	Evidence   map[string]any                      `json:"evidence,omitempty"`
	TaskResult *research.CompanyResearchTaskResult `json:"task_result,omitempty"`
	Warnings   []string                            `json:"warnings,omitempty"`
}

func runTaskExecutor(ctx context.Context, cfg CompanyResearchConfig, payload research.CompanyResearchPayload, plan research.CompanyResearchTaskPlan, role research.CompanyResearchTaskRole, params map[string]any) (CompanyResearchTaskExecutionResult, bool) {
	if cfg.TaskExecutor == nil {
		return CompanyResearchTaskExecutionResult{}, false
	}
	task, ok := plan.TaskByRole(role)
	if !ok {
		return CompanyResearchTaskExecutionResult{}, false
	}
	result, err := cfg.TaskExecutor(ctx, CompanyResearchTaskExecutionRequest{
		Plan:              plan,
		Task:              task,
		Intent:            payload.Intent,
		SubjectResolution: payload.SubjectResolution,
		Params:            cloneTaskParams(params),
		Evidence:          payload.Evidence,
		PreviousResults:   append([]research.CompanyResearchTaskResult(nil), payload.TaskResults...),
	})
	if err != nil {
		result.Warnings = append(result.Warnings, "task_executor_error:"+string(role))
		result.TaskResult = &research.CompanyResearchTaskResult{
			TaskID:        task.ID,
			Role:          task.Role,
			Status:        research.CompanyResearchTaskStatusDegraded,
			Dimensions:    task.Dimensions,
			ExecutorID:    "host_task_executor",
			FailureCode:   "task_executor_error",
			Summary:       "host-owned task executor returned an error; default downstream handler may still run",
			Warnings:      []string{err.Error()},
			AdapterStatus: "degraded",
		}
		return result, true
	}
	if result.TaskResult != nil {
		normalized := normalizeExecutorTaskResult(task, *result.TaskResult)
		result.TaskResult = &normalized
	}
	return result, true
}

func applyHandledTaskResult(payload *research.CompanyResearchPayload, task research.CompanyResearchTaskSpec, execution CompanyResearchTaskExecutionResult) {
	if payload == nil {
		return
	}
	if execution.TaskResult != nil {
		appendTaskResult(payload, normalizeExecutorTaskResult(task, *execution.TaskResult))
		return
	}
	if taskResult, ok := research.TaskResultFromEvidence(research.CompanyResearchTaskPlan{Tasks: []research.CompanyResearchTaskSpec{task}}, task.Role, execution.Evidence); ok {
		taskResult.ExecutorID = "host_task_executor"
		appendTaskResult(payload, taskResult)
	}
}

func normalizeExecutorTaskResult(task research.CompanyResearchTaskSpec, result research.CompanyResearchTaskResult) research.CompanyResearchTaskResult {
	if strings.TrimSpace(result.TaskID) == "" {
		result.TaskID = task.ID
	}
	if result.Role == "" {
		result.Role = task.Role
	}
	if len(result.Dimensions) == 0 {
		result.Dimensions = append([]string(nil), task.Dimensions...)
	}
	if strings.TrimSpace(result.ExecutorID) == "" {
		result.ExecutorID = "host_task_executor"
	}
	result.Diagnostics = sanitizeTaskExecutorDiagnostics(result.Diagnostics)
	return result
}

func sanitizeTaskExecutorDiagnostics(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := map[string]string{}
	for key, value := range input {
		key = strings.TrimSpace(key)
		if key == "" || sensitiveTaskDiagnosticKey(key) {
			continue
		}
		out[key] = trimTaskDiagnosticValue(value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func sensitiveTaskDiagnosticKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, token := range []string{"model", "provider", "api_key", "apikey", "token", "secret", "credential", "password"} {
		if strings.Contains(key, token) {
			return true
		}
	}
	return false
}

func trimTaskDiagnosticValue(value string) string {
	value = strings.TrimSpace(value)
	const maxDiagnosticChars = 256
	if len(value) <= maxDiagnosticChars {
		return value
	}
	return value[:maxDiagnosticChars]
}

func cloneTaskParams(params map[string]any) map[string]any {
	if len(params) == 0 {
		return nil
	}
	out := make(map[string]any, len(params))
	for key, value := range params {
		out[key] = value
	}
	return out
}
