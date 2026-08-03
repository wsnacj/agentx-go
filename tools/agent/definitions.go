// Package agent provides the model-facing Task, Session child and Subagent
// tool contract over a Host-owned durable lifecycle.
package agent

import toolcontract "github.com/wsnacj/agentx-go/components/tool"

const (
	TasksSpawnName          = "tasks_spawn"
	TasksWaitName           = "tasks_wait"
	TasksRunName            = "tasks_run"
	TasksCancelName         = "tasks_cancel"
	TasksReplayName         = "tasks_replay"
	TasksCollectName        = "tasks_collect"
	TasksDeadletterListName = "tasks_deadletter_list"
	SubagentsName           = "subagents"
	AgentStepName           = "agent_step"
)

// Definitions returns the complete model-facing catalog in stable order.
func Definitions() []toolcontract.Definition {
	return []toolcontract.Definition{
		TasksSpawnDefinition(), TasksWaitDefinition(), TasksRunDefinition(),
		TasksCancelDefinition(), TasksReplayDefinition(), TasksCollectDefinition(),
		TasksDeadletterListDefinition(), SubagentsDefinition(), AgentStepDefinition(),
	}
}

func TasksSpawnDefinition() toolcontract.Definition {
	return definition(TasksSpawnName,
		"Create or reuse a task-backed child session with parent linkage. Use seed_message for the first child instruction; this only creates/queues the child and does not aggregate final output.",
		taskSpawnParametersSchema(), tasksSpawnOutputSchema())
}

func TasksWaitDefinition() toolcontract.Definition {
	return definition(TasksWaitName,
		"Wait for a child task status update and return the latest task/session state. Provide task_id or session_id; use wait_ms/timeout_ms to bound polling.",
		tasksWaitParametersSchema(), tasksWaitOutputSchema())
}

func TasksRunDefinition() toolcontract.Definition {
	return definition(TasksRunName,
		"Create or reuse a child task, then wait for its current/final state in one call. Provide seed_message as the child instruction. Use non_blocking=true for fire-and-collect-later; requeue is applied only when requeue=true.",
		tasksRunParametersSchema(), tasksRunOutputSchema())
}

func TasksCancelDefinition() toolcontract.Definition {
	return definition(TasksCancelName, "Cancel a child task and mark the task/session state as canceled.", objectSchema(map[string]any{
		"task_id":    stringSchema("Task identifier to cancel. Provide task_id or session_id."),
		"session_id": stringSchema("Session identifier backing the task to cancel. Provide task_id or session_id."),
		"reason":     stringSchema("Human-readable cancellation reason recorded in task state."),
	}), nil)
}

func TasksReplayDefinition() toolcontract.Definition {
	return definition(TasksReplayName, "Manually replay one or more tasks by re-enqueueing scheduler jobs.", objectSchema(map[string]any{
		"task_id":    stringSchema("Single task identifier to replay."),
		"task_ids":   stringArraySchema("Multiple task identifiers to replay in one operation."),
		"session_id": stringSchema("Session identifier backing the task to replay when task_id is not known."),
		"job_id":     stringSchema("Scheduler job id to replay or use for replay identity."),
		"queue_mode": queueModeSchema("Queue mode used for replayed scheduler jobs."),
		"lane":       stringSchema("Scheduler lane for replayed jobs."),
		"reason":     stringSchema("Human-readable replay reason recorded for operations audit."),
	}), nil)
}

func TasksDeadletterListDefinition() toolcontract.Definition {
	return definition(TasksDeadletterListName, "List tasks currently in dead_letter queue status for operations replay.", objectSchema(map[string]any{
		"parent_task_id":  stringSchema("Parent task id used to filter dead-lettered child tasks."),
		"task_ids":        stringArraySchema("Explicit task ids to inspect for dead-letter status."),
		"limit":           intSchema("Maximum number of dead-lettered tasks to return.", 1),
		"include_summary": boolSchema("Include latest child summary text when available."),
	}), nil)
}

func objectSchema(properties map[string]any) map[string]any {
	return map[string]any{"type": "object", "properties": properties}
}

func TasksCollectDefinition() toolcontract.Definition {
	return definition(TasksCollectName,
		"Collect child task states by parent_task_id or explicit task_ids, with optional queue/lifecycle diagnostics, tree aggregation, and terminal event inbox.",
		tasksCollectParametersSchema(), tasksCollectOutputSchema())
}

func SubagentsDefinition() toolcontract.Definition {
	return definition(SubagentsName,
		"Manage task-backed subagents through the existing scheduler/runstore lifecycle: list, status, run, fanout, cancel, replay, or steer. For run and fanout, provide message or seed_message. For multiple independent children, use action=fanout with one items entry per child; when the user specifies an exact count, set expected_count to that number. With queue_mode=collect, the fanout result already includes child handles/results, so avoid extra list/status calls unless the user asks to monitor or cancel. Use non_blocking=true to start now and collect later.",
		subagentsParametersSchema(), subagentsOutputSchema())
}

func AgentStepDefinition() toolcontract.Definition {
	return definition(AgentStepName,
		"Run one bounded child-agent LLM step with stable session/task provenance and a persisted child transcript. Use subagents fanout for multiple independent child tasks.",
		agentStepParametersSchema(), agentStepOutputSchema())
}

func definition(name, description string, parameters, output map[string]any) toolcontract.Definition {
	return toolcontract.Definition{Type: "function", Function: toolcontract.Function{
		Name: name, Description: description, Parameters: parameters, OutputSchema: output,
	}}
}
