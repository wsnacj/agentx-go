package agent

func closedInputSchema(properties map[string]any, required []string) map[string]any {
	out := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties":           properties,
	}
	if len(required) > 0 {
		out["required"] = append([]string(nil), required...)
	}
	return out
}

func numberSchema(description string) map[string]any {
	return map[string]any{
		"type":        "number",
		"description": description,
	}
}

func jsonSchemaInputSchema(description string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"description":          description,
		"additionalProperties": true,
	}
}

func looseObjectArraySchema(description string) map[string]any {
	return map[string]any{
		"type":        "array",
		"description": description,
		"items": map[string]any{
			"type":                 "object",
			"additionalProperties": true,
		},
	}
}

func queueModeSchema(description string) map[string]any {
	return stringEnumSchema(description, "followup", "collect", "steer")
}

func taskStatusSchema(description string) map[string]any {
	return stringEnumSchema(description, "queued", "running", "completed", "failed", "canceled")
}

func subagentActionSchema() map[string]any {
	return stringEnumSchema("Subagent lifecycle action to perform. Prefer fanout for parallel new children, list/status before steering existing children, steer for new direction, replay for rerun, and cancel for obsolete work.", "list", "status", "run", "fanout", "cancel", "replay", "steer")
}

func childContractInputProperties() map[string]any {
	return map[string]any{
		"model":                        stringSchema("Optional child model config. Runtime policy may override or reject unsupported models."),
		"task_kind":                    stringSchema("Optional child task kind for product-shell or scheduler routing."),
		"product_shell":                stringSchema("Optional product shell profile for the child."),
		"requested_skill":              stringSchema("Single builtin or project skill requested for the child run. This is a prompt/selection request, not permission to access hidden tools."),
		"requested_skills":             stringArraySchema("Builtin or project skills requested for the child run. Runtime skill selection and tool policy still apply."),
		"skill_activation_paths":       stringArraySchema("Optional file or resource focus hints used by skill activation for the child run."),
		"skill_activation_path_source": stringSchema("Optional source label for child skill activation path hints."),
		"allowed_tools":                stringArraySchema("Tool allowlist visible to the child task."),
		"approval_mode":                stringSchema("Approval mode for child tool use."),
		"approval_allow_tools":         stringArraySchema("Tools explicitly allowed by child approval policy."),
		"approval_deny_tools":          stringArraySchema("Tools explicitly denied by child approval policy."),
		"approval_max_risk":            stringSchema("Maximum tool risk accepted by child approval policy."),
		"side_effect_class":            stringSchema("Expected side-effect class for the child, such as read_only or workspace_write."),
		"done_condition":               stringSchema("Natural-language condition used to judge whether the child output is complete."),
		"role_hint":                    stringEnumSchema("Child orchestration role hint. leaf children should complete directly; orchestrator children may coordinate child work.", "leaf", "orchestrator"),
		"evidence_required":            boolSchema("Require the child final answer to include concrete evidence handles, such as files, artifacts, task/session ids, command outputs, or source snippets."),
		"verification_required":        boolSchema("Require the child final answer to include verification steps, outcomes, or explicit blockers when verification cannot run."),
		"output_schema":                jsonSchemaInputSchema("Optional JSON Schema expected from the child assistant's final output."),
		"failure_policy":               stringEnumSchema("Parent-facing policy for child failure handling.", "fail_fast", "collect_partial", "retry", "cancel_cascade"),
		"failure_retry_max":            intSchema("Maximum child retries recommended by failure policy. Zero means unbounded by this child contract.", 0),
		"failure_retry_backoff_ms":     intSchema("Suggested retry backoff in milliseconds for failure_policy=retry.", 0),
		"failure_cancel_cascade":       boolSchema("Whether failed child handling should cancel active descendants."),
	}
}

func taskIdentityInputProperties() map[string]any {
	return map[string]any{
		"task_id":           stringSchema("Stable child task identifier. Omit to let the runtime generate one."),
		"session_id":        stringSchema("Stable child session identifier. Omit to let the runtime generate one."),
		"parent_session_id": stringSchema("Parent session that owns the child session."),
		"parent_task_id":    stringSchema("Parent task that owns the child task."),
		"job_id":            stringSchema("Optional scheduler job identifier."),
		"queue_mode":        queueModeSchema("Scheduler queue mode: followup for normal child work, collect for fanout aggregation, steer for revisions."),
		"lane":              stringSchema("Optional scheduler lane."),
	}
}

func taskInstructionInputProperties() map[string]any {
	return map[string]any{
		"title":         stringSchema("Short child task title used for task identity and dedupe."),
		"goal":          stringSchema("Child task goal or success criterion."),
		"status":        taskStatusSchema("Initial task status. Omit for runtime default."),
		"budget_tokens": intSchema("Optional child token budget. Runtime can inherit, clamp, or ignore according to parent constraints.", 0),
		"timeout_sec":   intSchema("Optional child hard timeout in seconds. Runtime clamps this to configured and parent limits.", 1),
		"attempt":       intSchema("Optional scheduler attempt number.", 0),
		"seed_message":  stringSchema("Initial user message written to the child session. Prefer this as the child instruction for tasks_run/tasks_spawn."),
		"seed_role":     stringSchema("Role for seed_message. Defaults to user."),
	}
}

func taskSpawnControlInputProperties() map[string]any {
	return map[string]any{
		"allow_existing": boolSchema("Allow reusing an existing task/session when identifiers or dedupe match. Defaults vary by tool."),
		"requeue":        boolSchema("Explicitly requeue an existing child task. Defaults to false."),
	}
}

func taskMetaInputProperties() map[string]any {
	return map[string]any{
		"meta_json": stringSchema("Optional raw JSON metadata string for the child session."),
		"meta":      jsonSchemaInputSchema("Optional metadata object for the child session."),
	}
}

func taskSpawnInputProperties() map[string]any {
	props := map[string]any{}
	mergeSchemaProperties(props, taskIdentityInputProperties())
	mergeSchemaProperties(props, taskInstructionInputProperties())
	mergeSchemaProperties(props, taskSpawnControlInputProperties())
	mergeSchemaProperties(props, childContractInputProperties())
	mergeSchemaProperties(props, taskMetaInputProperties())
	return props
}

func tasksRunInputProperties() map[string]any {
	props := taskSpawnInputProperties()
	props["dedupe_inflight"] = boolSchema("Allow runtime dedupe of in-flight child tasks with the same parent intent.")
	props["non_blocking"] = boolSchema("Start or reuse the child now and return after a short wait budget so the parent can collect later.")
	props["wait_ms"] = intSchema("Initial poll delay in milliseconds before checking child state.", 0)
	props["timeout_ms"] = intSchema("Maximum wait budget in milliseconds for this tasks_run call.", 0)
	return props
}

func tasksCollectInputProperties() map[string]any {
	return map[string]any{
		"parent_task_id":         stringSchema("Parent task whose child tasks should be collected."),
		"task_ids":               stringArraySchema("Explicit child task IDs to collect."),
		"queue_status":           stringSchema("Single scheduler queue status filter."),
		"queue_statuses":         stringArraySchema("Multiple scheduler queue status filters."),
		"queue_mode":             queueModeSchema("Queue mode filter for collected tasks."),
		"include_summary":        boolSchema("Include latest child summary text when available."),
		"include_tree":           boolSchema("Include a task tree rooted at the selected parent or tasks."),
		"tree_max_depth":         intSchema("Maximum task tree depth to include.", 0),
		"include_diagnostics":    boolSchema("Include stalled, blocked, timeout, bottleneck, and tree alert diagnostics."),
		"include_inbox":          boolSchema("Include derived terminal child event inbox entries."),
		"include_child_registry": boolSchema("Include a compact active child registry with tools, budget, approval, lease, and lifecycle evidence."),
		"inbox_after_ms":         intSchema("Only include event inbox entries after this Unix millisecond cursor.", 0),
		"stalled_after_ms":       intSchema("Mark running tasks as stalled after this many milliseconds.", 1),
		"limit":                  intSchema("Maximum task records to return.", 1),
	}
}

func tasksWaitParametersSchema() map[string]any {
	return closedInputSchema(map[string]any{
		"task_id":    stringSchema("Child task identifier to wait for. Provide task_id or session_id."),
		"task_ids":   stringArraySchema("Child task identifiers to wait for as a batch. Use this after subagents action=fanout returns multiple task IDs."),
		"session_id": stringSchema("Child session identifier to wait for. Provide task_id or session_id."),
		"queue_mode": queueModeSchema("Expected scheduler queue mode for the task."),
		"wait_ms":    intSchema("Initial poll delay in milliseconds.", 0),
		"timeout_ms": intSchema("Maximum wait budget in milliseconds. Use 0 for a status check without waiting.", 0),
	}, nil)
}

func taskSpawnParametersSchema() map[string]any {
	return closedInputSchema(taskSpawnInputProperties(), nil)
}

func tasksRunParametersSchema() map[string]any {
	return closedInputSchema(tasksRunInputProperties(), nil)
}

func tasksCollectParametersSchema() map[string]any {
	return closedInputSchema(tasksCollectInputProperties(), nil)
}

func subagentsParametersSchema() map[string]any {
	props := tasksRunInputProperties()
	props["action"] = subagentActionSchema()
	props["task_ids"] = stringArraySchema("Task IDs for list, replay, or collect-style actions.")
	props["queue_status"] = stringSchema("Single scheduler queue status filter for list.")
	props["queue_statuses"] = stringArraySchema("Multiple scheduler queue status filters for list.")
	props["message"] = stringSchema("Concrete child instruction for action=run, action=fanout items, or action=steer. Runtime copies this to seed_message for new children.")
	props["role"] = stringSchema("Message role for action=steer. Defaults to user.")
	props["reason"] = stringSchema("Reason for cancel, replay, or steer.")
	props["background"] = boolSchema("Compatibility flag for read-only background child work. Do not default side-effectful children to background; prefer bounded wait/collect or explicit non_blocking=true when the user wants later collection.")
	props["expected_count"] = intSchema("Optional action=fanout contract. Set this when the user asked for an exact number of parallel children; items length must match it.", 1)
	props["items"] = subagentFanoutItemsSchema()
	mergeSchemaProperties(props, tasksCollectInputProperties())
	return closedInputSchema(props, []string{"action"})
}

func subagentFanoutItemsSchema() map[string]any {
	props := tasksRunInputProperties()
	props["message"] = stringSchema("Concrete instruction for this fanout item. Runtime copies this to seed_message when seed_message is omitted.")
	props["background"] = boolSchema("Compatibility flag for read-only background child work. Do not default side-effectful children to background; prefer bounded wait/collect or explicit non_blocking=true when the user wants later collection.")
	return map[string]any{
		"type":        "array",
		"description": "Fanout children to start. Use one array entry per independent child task or target; do not merge multiple requested children into one item. Each item should provide message or seed_message plus optional title/goal/budget/model/tool policy overrides.",
		"items": map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties":           props,
		},
	}
}

func agentStepParametersSchema() map[string]any {
	props := map[string]any{
		"session_id":        stringSchema("Child session identifier. Omit to create a transient generated session."),
		"task_id":           stringSchema("Optional child task identifier when using task-backed provenance."),
		"parent_session_id": stringSchema("Parent session for the child step."),
		"parent_task_id":    stringSchema("Parent task for task-backed provenance."),
		"queue_mode":        queueModeSchema("Queue mode used when task-backed provenance is requested."),
		"lane":              stringSchema("Optional scheduler lane for task-backed provenance."),
		"title":             stringSchema("Optional child task/session title metadata."),
		"goal":              stringSchema("Optional child task goal metadata. Use instruction for the actual step prompt."),
		"instruction":       stringSchema("Focused instruction for this one bounded child-agent step. Runtime still accepts task/prompt/request/query/goal aliases for compatibility."),
		"input":             stringSchema("Optional source text or context for the step. Runtime still accepts context as a compatibility alias."),
		"schema":            jsonSchemaInputSchema("Optional JSON Schema subset for validating the step result."),
		"output_schema":     jsonSchemaInputSchema("Compatibility alias for schema. Prefer schema for new calls."),
		"strict":            boolSchema("Whether to request strict JSON schema adherence when schema is provided. Defaults to true."),
		"model":             stringSchema("Optional model config override when allowed by runtime configuration."),
		"temperature":       numberSchema("Optional model temperature."),
		"max_tokens":        intSchema("Optional maximum model output tokens.", 1),
		"timeout_ms":        intSchema("Maximum step runtime in milliseconds. Omit to use the runtime default.", 1),
		"allow_existing":    boolSchema("Allow appending to an existing child session. Defaults to true."),
		"persist_result":    boolSchema("Persist the assistant JSON result back into the child transcript. Defaults to true."),
	}
	return closedInputSchema(props, []string{"instruction"})
}

func mergeSchemaProperties(dst map[string]any, src map[string]any) {
	for key, value := range src {
		dst[key] = value
	}
}

func taskStateOutputProperties() map[string]any {
	return map[string]any{
		"task_id":      stringSchema("Task identifier."),
		"session_id":   stringSchema("Session identifier backing the task."),
		"status":       stringSchema("Task status."),
		"done":         boolSchema("True when the task reached a terminal status."),
		"timed_out":    boolSchema("True when the wait budget expired before terminal status."),
		"waited_ms":    intSchema("Milliseconds spent waiting in this call.", 0),
		"queue_status": stringSchema("Scheduler queue status."),
		"queue_error":  stringSchema("Scheduler queue error when present."),
		"advice":       stringSchema("Operational advice when the task is still pending, timed out, or blocked."),
	}
}

func taskWaitOutputProperties() map[string]any {
	props := taskStateOutputProperties()
	mergeSchemaProperties(props, map[string]any{
		"job_id":                     stringSchema("Scheduler job identifier."),
		"queue_mode":                 queueModeSchema("Scheduler queue mode used by the task."),
		"wait_budget_ms":             intSchema("Effective wait budget in milliseconds for this status/wait call.", 0),
		"wait_budget_cap_ms":         intSchema("Maximum wait budget applied by the runtime.", 0),
		"wait_budget_capped":         boolSchema("True when the requested wait budget was capped by the runtime."),
		"requested_wait_budget_ms":   intSchema("Requested wait budget before runtime capping.", 0),
		"lane":                       stringSchema("Scheduler lane."),
		"attempt":                    intSchema("Current scheduler attempt.", 0),
		"retry_count":                intSchema("Scheduler retry count.", 0),
		"lifecycle_state":            stringSchema("Runtime lifecycle state."),
		"lifecycle_reason":           stringSchema("Runtime lifecycle reason."),
		"announce_state":             stringSchema("Announcement state for child result delivery."),
		"announce_error":             stringSchema("Announcement delivery error."),
		"announce_updated_at":        intSchema("Announcement update time in Unix milliseconds.", 0),
		"announce_next_at":           intSchema("Next announcement attempt time in Unix milliseconds.", 0),
		"descendant_cancel_async":    boolSchema("True when descendant cancellation is asynchronous."),
		"descendant_cancel_accepted": boolSchema("True when descendant cancellation was accepted."),
		"descendant_cancel_token":    stringSchema("Cancellation token for descendant cancellation."),
		"running_since":              intSchema("Task running start time in Unix milliseconds.", 0),
		"running_ms":                 intSchema("Milliseconds the task has been running.", 0),
		"session_status":             stringSchema("Backing session status."),
		"last_message":               stringSchema("Latest child message content preview."),
		"summary_text":               stringSchema("Latest child summary text."),
		"task_ids":                   stringArraySchema("Task identifiers waited for when using batch task_ids mode."),
		"count":                      intSchema("Number of tasks in the batch wait result.", 0),
		"done_count":                 intSchema("Number of terminal tasks in the batch wait result.", 0),
		"collect":                    looseObjectSchema("Latest tasks_collect-style batch state when using task_ids."),
	})
	return props
}

func tasksSpawnOutputSchema() map[string]any {
	return closedOutputSchema(map[string]any{
		"task_id":             stringSchema("Child task identifier."),
		"session_id":          stringSchema("Child session identifier."),
		"job_id":              stringSchema("Scheduler job identifier."),
		"queue_mode":          queueModeSchema("Scheduler queue mode used by the task."),
		"lane":                stringSchema("Scheduler lane."),
		"attempt":             intSchema("Scheduler attempt number.", 0),
		"queue_status":        stringSchema("Scheduler queue status."),
		"lifecycle_state":     stringSchema("Runtime lifecycle state."),
		"lifecycle_reason":    stringSchema("Runtime lifecycle reason."),
		"announce_state":      stringSchema("Announcement state for child result delivery."),
		"retry_count":         intSchema("Scheduler retry count.", 0),
		"announce_error":      stringSchema("Announcement delivery error."),
		"announce_updated_at": intSchema("Announcement update time in Unix milliseconds.", 0),
		"announce_next_at":    intSchema("Next announcement attempt time in Unix milliseconds.", 0),
		"enqueued":            boolSchema("True when a scheduler job was enqueued."),
		"requeue_applied":     boolSchema("True when explicit requeue was applied."),
		"created":             boolSchema("True when a new child session/task was created."),
		"seeded":              boolSchema("True when seed_message was written to the child transcript."),
		"status":              stringSchema("Task status after spawn."),
		"parent_task_id":      stringSchema("Parent task identifier."),
		"budget_tokens":       intSchema("Effective child token budget.", 0),
		"timeout_sec":         intSchema("Effective child timeout in seconds.", 0),
		"budget_inherited":    boolSchema("True when budget came from parent constraints."),
		"budget_clamped":      boolSchema("True when budget was clamped."),
		"timeout_inherited":   boolSchema("True when timeout came from parent constraints."),
		"timeout_clamped":     boolSchema("True when timeout was clamped."),
		"timeout_defaulted":   boolSchema("True when runtime defaulted the timeout."),
		"timeout_hard_capped": boolSchema("True when hard timeout cap was applied."),
		"parent_constrained":  boolSchema("True when parent task constraints affected the child."),
		"model":               stringSchema("Effective child model."),
		"model_route":         looseObjectSchema("Child model routing details."),
		"child_contract":      looseObjectSchema("Child task contract details."),
	}, []string{"task_id", "session_id", "created", "seeded", "status"})
}

func tasksWaitOutputSchema() map[string]any {
	return closedOutputSchema(taskWaitOutputProperties(), []string{"done", "running_ms", "session_status"})
}

func tasksRunOutputSchema() map[string]any {
	props := taskStateOutputProperties()
	mergeSchemaProperties(props, map[string]any{
		"queue_mode":               queueModeSchema("Scheduler queue mode used by the child task."),
		"wait_budget_ms":           intSchema("Wait budget used by tasks_run.", 0),
		"wait_budget_cap_ms":       intSchema("Maximum wait budget applied by the runtime.", 0),
		"wait_budget_capped":       boolSchema("True when the requested wait budget was capped by the runtime."),
		"requested_wait_budget_ms": intSchema("Requested wait budget before runtime capping.", 0),
		"reused":                   boolSchema("True when an existing child task/session was reused."),
		"reuse_reason":             stringSchema("Reason an existing child was reused."),
		"requeue_applied":          boolSchema("True when explicit requeue was applied."),
		"spawn":                    looseObjectSchema("Raw tasks_spawn result for this child."),
		"wait":                     looseObjectSchema("Raw tasks_wait result for this child."),
	})
	return closedOutputSchema(props, []string{"done", "spawn", "wait"})
}

func taskCollectItemOutputSchema() map[string]any {
	props := taskWaitOutputProperties()
	mergeSchemaProperties(props, map[string]any{
		"finished_at":                intSchema("Task finish time in Unix milliseconds.", 0),
		"parent_task_id":             stringSchema("Parent task identifier."),
		"budget_tokens":              intSchema("Effective child token budget.", 0),
		"timeout_sec":                intSchema("Effective child hard timeout in seconds.", 0),
		"updated_at":                 intSchema("Task update time in Unix milliseconds.", 0),
		"age_ms":                     intSchema("Task age in milliseconds.", 0),
		"scheduler_status":           stringSchema("Raw scheduler queue row status, including coalesced aliases."),
		"coalesced":                  boolSchema("True when the scheduler row is a coalesced alias."),
		"coalesced_to_job_id":        stringSchema("Target scheduler job for a coalesced alias."),
		"coalesced_reason":           stringSchema("Reason the scheduler job was coalesced."),
		"scheduler_lease_owner":      stringSchema("Current scheduler lease owner for running jobs."),
		"scheduler_lease_expires_at": intSchema("Current scheduler lease expiry in Unix milliseconds.", 0),
		"scheduler_heartbeat_at":     intSchema("Current scheduler heartbeat time in Unix milliseconds.", 0),
		"lease_owner":                stringSchema("Backward-compatible alias for scheduler_lease_owner."),
		"lease_expires_at":           intSchema("Backward-compatible alias for scheduler_lease_expires_at.", 0),
		"heartbeat_at":               intSchema("Backward-compatible alias for scheduler_heartbeat_at.", 0),
		"scheduler_stale":            boolSchema("True when diagnostics classify the scheduler job as stale."),
		"scheduler_stale_reason":     stringSchema("Reason the scheduler job is classified as stale."),
		"is_blocked":                 boolSchema("True when diagnostics classify the task as blocked."),
		"is_timeout":                 boolSchema("True when diagnostics classify the task as timed out."),
		"is_bottleneck":              boolSchema("True when diagnostics classify the task as a bottleneck."),
		"cancel_propagating":         boolSchema("True when descendant cancellation is propagating."),
		"alert_reason":               stringSchema("Diagnostic alert reason."),
		"child_contract":             looseObjectSchema("Projected child contract metadata."),
		"child_result":               looseObjectSchema("Projected child result contract with summary, evidence handles, artifacts, changed files, blockers, uncertainty, and self-report warning."),
		"model_route":                looseObjectSchema("Child model route metadata."),
		"output_validation":          looseObjectSchema("Derived child output validation status."),
		"failure_policy":             looseObjectSchema("Projected child failure policy metadata."),
		"failure_policy_action":      stringSchema("Recommended policy action for failed, timed-out, stale, blocked, or invalid-output child tasks."),
		"failure_policy_reason":      stringSchema("Reason that triggered the failure policy action."),
	})
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties":           props,
	}
}

func tasksCollectOutputSchema() map[string]any {
	return closedOutputSchema(map[string]any{
		"parent_task_id":           stringSchema("Parent task identifier used for collection."),
		"queue_mode":               queueModeSchema("Queue mode filter used for collection."),
		"include_summary":          boolSchema("True when summaries were requested."),
		"count":                    intSchema("Number of task records returned.", 0),
		"done_count":               intSchema("Number of terminal task records returned.", 0),
		"dead_letter_count":        intSchema("Number of dead-letter task records returned.", 0),
		"status_digest":            looseObjectSchema("Derived child task status digest, including stale scheduler job counts and recommended next action."),
		"result_digest":            looseObjectSchema("Derived child result digest with summary, evidence, artifact, blocker, and missing-evidence counts."),
		"stale_job_count":          intSchema("Number of active scheduler jobs classified as stale.", 0),
		"include_child_registry":   boolSchema("True when compact active child registry output was requested or defaulted for parent collection."),
		"child_registry":           looseObjectSchema("Compact active child registry with current child runs, tools, budget, approval, lease, lifecycle, and diagnostics evidence."),
		"tasks":                    map[string]any{"type": "array", "description": "Collected task records.", "items": taskCollectItemOutputSchema()},
		"by_status":                looseObjectSchema("Counts by task status."),
		"by_queue_status":          looseObjectSchema("Counts by scheduler queue status."),
		"by_lifecycle_state":       looseObjectSchema("Counts by lifecycle state."),
		"by_announce_state":        looseObjectSchema("Counts by announcement state."),
		"missing_tasks":            stringArraySchema("Requested task IDs that were not found."),
		"queue_filters":            stringArraySchema("Queue status filters applied."),
		"include_tree":             boolSchema("True when task tree output was requested."),
		"tree_max_depth":           intSchema("Task tree depth used for output.", 0),
		"task_tree":                looseObjectArraySchema("Task tree nodes when include_tree=true."),
		"task_tree_levels":         looseObjectArraySchema("Aggregated task tree levels."),
		"include_diagnostics":      boolSchema("True when diagnostics were requested."),
		"include_inbox":            boolSchema("True when event inbox was requested."),
		"inbox_after_ms":           intSchema("Event inbox cursor used for filtering.", 0),
		"event_inbox":              looseObjectArraySchema("Derived terminal child event inbox."),
		"event_inbox_cursor":       intSchema("Cursor for the next event inbox poll.", 0),
		"stalled_after_ms":         intSchema("Stalled threshold used by diagnostics.", 0),
		"alert_count":              intSchema("Number of diagnostic alerts.", 0),
		"blocked_count":            intSchema("Number of blocked tasks.", 0),
		"timeout_count":            intSchema("Number of timed-out tasks.", 0),
		"cancel_propagating_count": intSchema("Number of tasks with cancellation propagating.", 0),
		"tree_alerts":              looseObjectArraySchema("Tree diagnostic alerts."),
		"bottleneck_task_id":       stringSchema("Task ID selected as the current bottleneck."),
		"bottleneck_reason":        stringSchema("Reason a bottleneck task was selected."),
		"bottleneck_path":          stringArraySchema("Task path to the bottleneck."),
	}, []string{"count", "done_count", "tasks"})
}

func subagentsOutputSchema() map[string]any {
	return closedOutputSchema(map[string]any{
		"action":         stringSchema("Subagent action that was executed."),
		"result":         looseObjectSchema("Wrapped result from tasks_run, tasks_wait, tasks_collect, tasks_cancel, or tasks_replay."),
		"task_id":        stringSchema("Task ID for direct steer responses."),
		"session_id":     stringSchema("Session ID for direct steer responses."),
		"role":           stringSchema("Message role for direct steer responses."),
		"message_length": intSchema("Length of steer message appended to the child transcript.", 0),
		"replay":         looseObjectSchema("Replay result emitted by steer when scheduler replay is enabled."),
	}, []string{"action"})
}

func agentStepOutputSchema() map[string]any {
	return closedOutputSchema(map[string]any{
		"session_id":     stringSchema("Child session identifier used by the step."),
		"task_id":        stringSchema("Child task identifier when task-backed provenance was used."),
		"created":        boolSchema("True when a new child session/task was created."),
		"seeded":         boolSchema("True when the user step prompt was written to the child transcript."),
		"persisted":      boolSchema("True when the assistant JSON result was persisted to the child transcript."),
		"model":          stringSchema("Effective model config used for the step."),
		"schema_applied": boolSchema("True when schema/output_schema validation was applied."),
		"queue_mode":     queueModeSchema("Queue mode used when task-backed provenance was requested."),
		"status":         stringSchema("Child task/session status after spawn."),
		"lifecycle":      stringSchema("Child lifecycle state after spawn."),
		"parent_task_id": stringSchema("Parent task identifier."),
		"result":         looseObjectSchema("Parsed JSON result returned by the child step."),
		"raw_json":       stringSchema("Raw JSON result string returned by the child step."),
	}, []string{"session_id", "created", "seeded", "persisted", "model", "schema_applied", "result", "raw_json"})
}

func llmTaskOutputSchema() map[string]any {
	return closedOutputSchema(map[string]any{
		"tool":           stringSchema("Tool name, always llm_task."),
		"model":          stringSchema("Effective model config used for the subtask."),
		"schema_applied": boolSchema("True when schema/output_schema validation was applied."),
		"result":         looseObjectSchema("Parsed JSON object returned by the model."),
		"raw_json":       stringSchema("Raw JSON result string returned by the model."),
	}, []string{"tool", "model", "schema_applied", "result", "raw_json"})
}
