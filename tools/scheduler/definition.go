package scheduler

import toolcontract "github.com/wsnacj/agentx-go/components/tool"

// Definition returns the stable scheduled-command tool schema.
func Definition() toolcontract.Definition {
	return toolcontract.Definition{Type: "function", Function: toolcontract.Function{
		Name:        Name,
		Description: "Manage persisted scheduled task jobs: add, list, inspect status, run immediately, or remove a scheduled task.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action":            stringEnumSchema("Cron operation to perform.", ActionAdd, ActionList, ActionStatus, ActionRun, ActionRemove),
				"task_id":           stringSchema("Scheduled task identifier. Omit on add to let the runtime generate one."),
				"session_id":        stringSchema("Session identifier backing the scheduled task."),
				"parent_session_id": stringSchema("Parent session that owns the scheduled task."),
				"parent_task_id":    stringSchema("Parent task that owns the scheduled task."),
				"job_id":            stringSchema("Scheduler job identifier for the scheduled task."),
				"title":             stringSchema("Short title recorded on the scheduled task."),
				"name":              stringSchema("Compatibility alias for title."),
				"goal":              stringSchema("Scheduled task goal or success criterion."),
				"message":           stringSchema("Scheduled task instruction. Used as a goal/prompt alias when goal is omitted."),
				"instruction":       stringSchema("Compatibility alias for the scheduled task instruction."),
				"task":              stringSchema("Compatibility alias for the scheduled task instruction."),
				"prompt":            stringSchema("Compatibility alias for the scheduled task instruction."),
				"request":           stringSchema("Compatibility alias for the scheduled task instruction."),
				"query":             stringSchema("Compatibility alias for the scheduled task instruction."),
				"input":             stringSchema("Compatibility alias for the scheduled task instruction."),
				"queue_mode":        stringEnumSchema("Scheduler queue mode for the scheduled task.", "followup", "collect", "steer"),
				"lane":              stringSchema("Scheduler lane for the scheduled task."),
				"run_at":            stringSchema("Absolute scheduled time accepted by the runtime parser."),
				"run_at_ms":         intSchema("Absolute scheduled time in Unix milliseconds.", 1),
				"delay_ms":          intSchema("Delay from now in milliseconds before the scheduled task should run.", 0),
				"timeout_sec":       intSchema("Task hard timeout in seconds. Runtime clamps this to configured limits.", 1),
				"allow_existing":    boolSchema("Allow action=add to update or reuse an existing scheduled task."),
				"include_summary":   boolSchema("Include latest task summary text for list/status responses when available."),
				"queue_status":      stringSchema("Single scheduler queue status filter for list."),
				"queue_statuses":    stringArraySchema("Multiple scheduler queue status filters for list."),
				"limit":             intSchema("Maximum number of scheduled tasks to return for list.", 1),
				"reason":            stringSchema("Human-readable reason for remove or run-now operations."),
				"model":             stringSchema("Model config recorded for the scheduled task session."),
				"meta_json":         stringSchema("Raw JSON metadata string for the scheduled task session."),
				"meta":              map[string]any{"type": "object", "description": "Metadata object for the scheduled task session.", "additionalProperties": true},
			},
			"required": []string{"action"},
		},
	}}
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}
func boolSchema(description string) map[string]any {
	return map[string]any{"type": "boolean", "description": description}
}
func intSchema(description string, minimum int) map[string]any {
	return map[string]any{"type": "integer", "description": description, "minimum": minimum}
}
func stringArraySchema(description string) map[string]any {
	return map[string]any{"type": "array", "description": description, "items": map[string]any{"type": "string"}}
}
func stringEnumSchema(description string, values ...string) map[string]any {
	return map[string]any{"type": "string", "description": description, "enum": append([]string(nil), values...)}
}
