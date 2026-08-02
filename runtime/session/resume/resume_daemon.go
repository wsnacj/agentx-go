package resume

import (
	"context"
	"errors"
	"strings"
	"time"

	scheduler "github.com/wsnacj/agentx-go/runtime/scheduler"
	session "github.com/wsnacj/agentx-go/runtime/session"
)

const ObjectiveRuntimeSchedulerResumeDaemonKind = "host_objective_runtime_scheduler_resume_daemon"
const ObjectiveRuntimeSchedulerResumeTickJobKind = "objective_runtime_scheduler_resume_tick"

type ObjectiveRuntimeSchedulerResumeDaemon struct {
	Queue  scheduler.Queue
	Worker ObjectiveRuntimeSchedulerResumeWorker
	Config ObjectiveRuntimeSchedulerResumeDaemonConfig
}

type ObjectiveRuntimeSchedulerResumeDaemonConfig struct {
	Lane        scheduler.Lane
	WorkerRef   session.DisplaySafeRef
	ProducerRef session.DisplaySafeRef
}

type ObjectiveRuntimeSchedulerResumeTickEnqueueInput struct {
	Enabled        bool
	Payload        ObjectiveRuntimeSchedulerResumeTickPayload
	JobID          session.DisplaySafeRef
	Lane           scheduler.Lane
	IdempotencyKey string
	CoalescingKey  string
	TrustedCaller  bool
}

type ObjectiveRuntimeSchedulerResumeTickEnqueueReport struct {
	Available                    bool                                       `json:"available"`
	Enabled                      bool                                       `json:"enabled"`
	Status                       string                                     `json:"status,omitempty"`
	DaemonKind                   string                                     `json:"daemon_kind,omitempty"`
	QueueAvailable               bool                                       `json:"queue_available"`
	QueueRuntimeVisible          bool                                       `json:"queue_runtime_visible"`
	TickAccepted                 bool                                       `json:"tick_accepted"`
	DisplaySafePayload           bool                                       `json:"display_safe_payload"`
	TickEnqueued                 bool                                       `json:"tick_enqueued"`
	QueueMutationByHost          bool                                       `json:"queue_mutation_by_host"`
	QueueMutationByCore          bool                                       `json:"queue_mutation_by_core"`
	QueuePendingReadback         bool                                       `json:"queue_pending_readback"`
	QueueResultReadbackReady     bool                                       `json:"queue_result_readback_ready"`
	QueuePending                 bool                                       `json:"queue_pending"`
	QueueResultStatus            string                                     `json:"queue_result_status,omitempty"`
	QueueResultSucceeded         bool                                       `json:"queue_result_succeeded"`
	JobID                        string                                     `json:"job_id,omitempty"`
	JobKind                      string                                     `json:"job_kind,omitempty"`
	Lane                         string                                     `json:"lane,omitempty"`
	TickRef                      string                                     `json:"tick_ref,omitempty"`
	SchedulerRuntimeQueueRef     string                                     `json:"scheduler_runtime_queue_ref,omitempty"`
	SchedulerWakeContinuationRef string                                     `json:"scheduler_wake_continuation_ref,omitempty"`
	ObjectiveRunRef              string                                     `json:"objective_run_ref,omitempty"`
	ObjectiveGraphSnapshotRef    string                                     `json:"objective_graph_snapshot_ref,omitempty"`
	ObjectiveGraphRef            string                                     `json:"objective_graph_ref,omitempty"`
	ObjectiveGraphReadbackRef    string                                     `json:"objective_graph_readback_ref,omitempty"`
	WorkerRef                    string                                     `json:"worker_ref,omitempty"`
	ProducerRef                  string                                     `json:"producer_ref,omitempty"`
	MissingInputs                []string                                   `json:"missing_inputs,omitempty"`
	BlockedReasons               []string                                   `json:"blocked_reasons,omitempty"`
	Boundaries                   []string                                   `json:"boundaries,omitempty"`
	NextHostAction               string                                     `json:"next_host_action,omitempty"`
	Payload                      ObjectiveRuntimeSchedulerResumeTickPayload `json:"payload,omitempty"`
}

type ObjectiveRuntimeSchedulerResumeProcessInput struct {
	Enabled bool
	Lane    scheduler.Lane
}

type ObjectiveRuntimeSchedulerResumeProcessReport struct {
	Available                    bool                                        `json:"available"`
	Enabled                      bool                                        `json:"enabled"`
	Status                       string                                      `json:"status,omitempty"`
	DaemonKind                   string                                      `json:"daemon_kind,omitempty"`
	QueueAvailable               bool                                        `json:"queue_available"`
	QueueRuntimeVisible          bool                                        `json:"queue_runtime_visible"`
	WorkerAvailable              bool                                        `json:"worker_available"`
	LeaseRequested               bool                                        `json:"lease_requested"`
	LeaseAcquired                bool                                        `json:"lease_acquired"`
	HeartbeatAttempted           bool                                        `json:"heartbeat_attempted"`
	HeartbeatSucceeded           bool                                        `json:"heartbeat_succeeded"`
	WorkerCalled                 bool                                        `json:"worker_called"`
	WorkerMutationByHost         bool                                        `json:"worker_mutation_by_host"`
	QueueAcked                   bool                                        `json:"queue_acked"`
	QueueFailed                  bool                                        `json:"queue_failed"`
	QueueResultReadbackReady     bool                                        `json:"queue_result_readback_ready"`
	QueueResultStatus            string                                      `json:"queue_result_status,omitempty"`
	QueueResultSucceeded         bool                                        `json:"queue_result_succeeded"`
	ReadyForRuntimeWakeDispatch  bool                                        `json:"ready_for_runtime_wake_dispatch"`
	DispatchRequestRecorded      bool                                        `json:"dispatch_request_recorded"`
	HostRuntimeDispatchByHost    bool                                        `json:"host_runtime_dispatch_by_host"`
	LLMWakeDispatched            bool                                        `json:"llm_wake_dispatched"`
	RunnerDispatched             bool                                        `json:"runner_dispatched"`
	RuntimeAdapterExecuted       bool                                        `json:"runtime_adapter_executed"`
	ToolExecuted                 bool                                        `json:"tool_executed"`
	WorkflowDispatched           bool                                        `json:"workflow_dispatched"`
	SchedulerApplied             bool                                        `json:"scheduler_applied"`
	InstallerExecuted            bool                                        `json:"installer_executed"`
	StoreMutationByCore          bool                                        `json:"store_mutation_by_core"`
	JobID                        string                                      `json:"job_id,omitempty"`
	JobKind                      string                                      `json:"job_kind,omitempty"`
	Lane                         string                                      `json:"lane,omitempty"`
	Attempt                      int                                         `json:"attempt,omitempty"`
	TickRef                      string                                      `json:"tick_ref,omitempty"`
	SchedulerRuntimeQueueRef     string                                      `json:"scheduler_runtime_queue_ref,omitempty"`
	SchedulerWakeContinuationRef string                                      `json:"scheduler_wake_continuation_ref,omitempty"`
	ObjectiveRunRef              string                                      `json:"objective_run_ref,omitempty"`
	ObjectiveGraphSnapshotRef    string                                      `json:"objective_graph_snapshot_ref,omitempty"`
	ObjectiveGraphRef            string                                      `json:"objective_graph_ref,omitempty"`
	ObjectiveGraphReadbackRef    string                                      `json:"objective_graph_readback_ref,omitempty"`
	ObjectiveGraphState          string                                      `json:"objective_graph_state,omitempty"`
	ObjectiveGraphRevision       int                                         `json:"objective_graph_revision,omitempty"`
	ReadyNodeRefs                []string                                    `json:"ready_node_refs,omitempty"`
	WorkerRef                    string                                      `json:"worker_ref,omitempty"`
	MissingInputs                []string                                    `json:"missing_inputs,omitempty"`
	BlockedReasons               []string                                    `json:"blocked_reasons,omitempty"`
	Boundaries                   []string                                    `json:"boundaries,omitempty"`
	NextHostAction               string                                      `json:"next_host_action,omitempty"`
	WorkerReport                 ObjectiveRuntimeSchedulerResumeWorkerReport `json:"worker_report,omitempty"`
}

func (d ObjectiveRuntimeSchedulerResumeDaemon) EnqueueSchedulerTick(ctx context.Context, input ObjectiveRuntimeSchedulerResumeTickEnqueueInput) ObjectiveRuntimeSchedulerResumeTickEnqueueReport {
	report := ObjectiveRuntimeSchedulerResumeTickEnqueueReport{
		Status:         "blocked",
		DaemonKind:     ObjectiveRuntimeSchedulerResumeDaemonKind,
		QueueAvailable: d.Queue != nil,
		WorkerRef:      string(objectiveRuntimeSchedulerResumeSafeRef(d.Config.WorkerRef)),
		ProducerRef:    string(objectiveRuntimeSchedulerResumeSafeRef(d.Config.ProducerRef)),
		Boundaries: []string{
			"host_owned_objective_runtime_scheduler_resume_daemon",
			"production_scheduler_tick_enqueue",
			"display_safe_refs_only",
			"no_scheduler_apply_by_core",
			"no_runner_dispatch_by_core",
			"no_store_mutation_by_core",
		},
		NextHostAction: "review_objective_runtime_scheduler_tick_enqueue",
	}
	if !input.Enabled {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:objective_runtime_scheduler_resume_daemon_enabled")
		report.Boundaries = appendUniqueResumeStrings(report.Boundaries, "objective_runtime_scheduler_resume_daemon_default_off")
		report.NextHostAction = "enable_objective_runtime_scheduler_resume_daemon"
		return report.Normalize()
	}
	report.Enabled = true
	report.Available = d.Queue != nil
	report.QueueRuntimeVisible = objectiveRuntimeSchedulerResumeQueueRuntimeVisible(d.Queue)
	if d.Queue == nil {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:scheduler_queue")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "scheduler_queue_missing")
		report.NextHostAction = "provide_scheduler_queue"
		return report.Normalize()
	}
	if !objectiveRuntimeSchedulerResumePayloadDisplaySafe(input.Payload) {
		report.DisplaySafePayload = false
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:display_safe_scheduler_tick_payload")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "scheduler_tick_payload_unsafe")
		report.Boundaries = appendUniqueResumeStrings(report.Boundaries, "raw_payload_not_allowed")
		return report.Normalize()
	}
	payload := BuildObjectiveRuntimeSchedulerResumeTickPayload(objectiveRuntimeSchedulerResumeDaemonPayload(input.Payload, d.Worker.Config))
	report.Payload = payload
	report.DisplaySafePayload = true
	report.TickAccepted = true
	report.TickRef = string(payload.TickRef)
	report.SchedulerRuntimeQueueRef = string(payload.SchedulerRuntimeQueueRef)
	report.SchedulerWakeContinuationRef = string(payload.SchedulerWakeContinuationRef)
	report.ObjectiveRunRef = string(payload.ObjectiveRunRef)
	report.ObjectiveGraphSnapshotRef = string(payload.ObjectiveGraphSnapshotRef)
	report.ObjectiveGraphRef = string(payload.ObjectiveGraphRef)
	report.ObjectiveGraphReadbackRef = string(payload.ObjectiveGraphReadbackRef)
	jobID := firstNonEmptyString(string(objectiveRuntimeSchedulerResumeSafeRef(input.JobID)), string(payload.SchedulerJobRef), string(payload.TickRef))
	if jobID == "" {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:scheduler_tick_job_ref")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "scheduler_tick_job_ref_missing")
		report.NextHostAction = "provide_scheduler_tick_job_ref"
		return report.Normalize()
	}
	payloadJSON, err := payload.JSON()
	if err != nil {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:scheduler_tick_payload")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "scheduler_tick_payload_marshal_failed")
		report.NextHostAction = "review_scheduler_tick_payload"
		return report.Normalize()
	}
	lane := objectiveRuntimeSchedulerResumeDaemonLane(input.Lane, d.Config.Lane)
	report.JobID = jobID
	report.JobKind = ObjectiveRuntimeSchedulerResumeTickJobKind
	report.Lane = string(lane)
	if ctx == nil {
		ctx = context.Background()
	}
	err = d.Queue.Enqueue(ctx, scheduler.Job{
		ID:             jobID,
		Lane:           lane,
		Payload:        payloadJSON,
		JobKind:        ObjectiveRuntimeSchedulerResumeTickJobKind,
		IdempotencyKey: strings.TrimSpace(input.IdempotencyKey),
		CoalescingKey:  strings.TrimSpace(input.CoalescingKey),
		TrustedCaller:  input.TrustedCaller,
		CreatedAt:      time.Now(),
	})
	if err != nil {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:scheduler_queue_enqueue")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "scheduler_tick_enqueue_failed")
		report.NextHostAction = "review_scheduler_tick_enqueue_failure"
		return report.Normalize()
	}
	report.TickEnqueued = true
	report.QueueMutationByHost = true
	report.Status = "objective_runtime_scheduler_tick_enqueued"
	report.NextHostAction = "run_objective_runtime_scheduler_resume_worker"
	report.Boundaries = appendUniqueResumeStrings(report.Boundaries,
		"scheduler_tick_queued_for_resume_worker",
		"queue_mutation_by_host",
	)
	if pending, err := d.Queue.Pending(ctx, jobID); err == nil {
		report.QueuePendingReadback = true
		report.QueuePending = pending
	}
	if result, ok, err := d.Queue.Result(ctx, jobID); err == nil && ok {
		report.QueueResultReadbackReady = true
		report.QueueResultStatus = strings.TrimSpace(result.Status)
		report.QueueResultSucceeded = result.Succeeded
	}
	return report.Normalize()
}

func (d ObjectiveRuntimeSchedulerResumeDaemon) ProcessNextSchedulerTick(ctx context.Context, input ObjectiveRuntimeSchedulerResumeProcessInput) ObjectiveRuntimeSchedulerResumeProcessReport {
	report := ObjectiveRuntimeSchedulerResumeProcessReport{
		Status:          "blocked",
		DaemonKind:      ObjectiveRuntimeSchedulerResumeDaemonKind,
		QueueAvailable:  d.Queue != nil,
		WorkerAvailable: d.Worker.ContinuationReadback != nil && d.Worker.WakeDispatch != nil,
		WorkerRef:       string(objectiveRuntimeSchedulerResumeSafeRef(d.Config.WorkerRef)),
		Boundaries: []string{
			"host_owned_objective_runtime_scheduler_resume_daemon",
			"cross_process_scheduler_resume_worker",
			"scheduler_queue_lease_ack_readback",
			"display_safe_refs_only",
			"no_scheduler_apply_by_core",
			"no_runner_dispatch_by_core",
			"no_store_mutation_by_core",
		},
		NextHostAction: "review_objective_runtime_scheduler_resume_worker",
	}
	if !input.Enabled {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:objective_runtime_scheduler_resume_daemon_enabled")
		report.Boundaries = appendUniqueResumeStrings(report.Boundaries, "objective_runtime_scheduler_resume_daemon_default_off")
		report.NextHostAction = "enable_objective_runtime_scheduler_resume_daemon"
		return report.Normalize()
	}
	report.Enabled = true
	report.Available = d.Queue != nil && report.WorkerAvailable
	report.QueueRuntimeVisible = objectiveRuntimeSchedulerResumeQueueRuntimeVisible(d.Queue)
	if d.Queue == nil {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:scheduler_queue")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "scheduler_queue_missing")
		report.NextHostAction = "provide_scheduler_queue"
		return report.Normalize()
	}
	if !report.WorkerAvailable {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:objective_runtime_scheduler_resume_worker")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "objective_runtime_scheduler_resume_worker_missing")
		report.NextHostAction = "provide_objective_runtime_scheduler_resume_worker"
		return report.Normalize()
	}
	kindQueue, ok := d.Queue.(scheduler.KindAwareQueue)
	if !ok || kindQueue == nil {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:scheduler_kind_aware_queue")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "scheduler_kind_aware_queue_missing")
		report.Boundaries = appendUniqueResumeStrings(report.Boundaries, "scheduler_resume_daemon_requires_kind_aware_dequeue")
		report.NextHostAction = "provide_scheduler_kind_aware_queue"
		return report.Normalize()
	}
	lane := objectiveRuntimeSchedulerResumeDaemonLane(input.Lane, d.Config.Lane)
	report.Lane = string(lane)
	report.LeaseRequested = true
	if ctx == nil {
		ctx = context.Background()
	}
	job, err := kindQueue.DequeueByKind(ctx, lane, ObjectiveRuntimeSchedulerResumeTickJobKind)
	if errors.Is(err, scheduler.ErrQueueEmpty) {
		report.Status = "objective_runtime_scheduler_resume_worker_idle"
		report.NextHostAction = "wait_for_scheduler_tick"
		return report.Normalize()
	}
	if err != nil {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:scheduler_queue_dequeue")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "scheduler_queue_dequeue_failed")
		report.NextHostAction = "review_scheduler_queue_dequeue_failure"
		return report.Normalize()
	}
	report.LeaseAcquired = true
	report.JobID = string(objectiveRuntimeSchedulerResumeSafeRef(session.DisplaySafeRef(job.ID)))
	report.JobKind = strings.TrimSpace(job.JobKind)
	report.Attempt = job.Attempt
	report.Boundaries = appendUniqueResumeStrings(report.Boundaries, "scheduler_tick_job_leased_by_host")
	if strings.TrimSpace(job.JobKind) != ObjectiveRuntimeSchedulerResumeTickJobKind {
		report = d.objectiveRuntimeSchedulerResumeDaemonFailJob(ctx, report, job, "objective_runtime_scheduler_resume_job_kind_mismatch", "scheduler: unexpected job kind")
		return report.Normalize()
	}
	if heartbeatQueue, ok := d.Queue.(scheduler.HeartbeatCapableQueue); ok && heartbeatQueue != nil && heartbeatQueue.HeartbeatInterval() > 0 {
		report.HeartbeatAttempted = true
		if err := heartbeatQueue.Heartbeat(ctx, job); err == nil {
			report.HeartbeatSucceeded = true
			report.Boundaries = appendUniqueResumeStrings(report.Boundaries, "scheduler_tick_job_lease_heartbeat_recorded")
		} else {
			report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "scheduler_queue_heartbeat_failed")
		}
	}
	workerReport, workerErr := d.Worker.HandleSchedulerTick(ctx, job)
	report.WorkerCalled = true
	report.WorkerReport = workerReport.Normalize()
	report = objectiveRuntimeSchedulerResumeDaemonBindWorkerReport(report, report.WorkerReport)
	if workerErr != nil {
		report = d.objectiveRuntimeSchedulerResumeDaemonFailJob(ctx, report, job, "objective_runtime_scheduler_resume_worker_failed", workerErr.Error())
		return report.Normalize()
	}
	if report.WorkerReport.ReadyForRuntimeWakeDispatch {
		if err := d.Queue.Ack(ctx, scheduler.Result{JobID: job.ID, Lane: lane, Attempt: job.Attempt}); err != nil {
			report = d.objectiveRuntimeSchedulerResumeDaemonFailJob(ctx, report, job, "scheduler_queue_ack_failed", "scheduler: ack failed: "+err.Error())
			return report.Normalize()
		}
		report.QueueAcked = true
		report.Status = "objective_runtime_scheduler_resume_worker_ack_recorded"
		report.NextHostAction = firstNonEmptyString(report.WorkerReport.NextHostAction, "host_dispatch_runtime_wake_runner")
		report.Boundaries = appendUniqueResumeStrings(report.Boundaries, "scheduler_tick_job_ack_recorded")
	} else {
		report = d.objectiveRuntimeSchedulerResumeDaemonFailJob(ctx, report, job, "objective_runtime_scheduler_resume_worker_not_ready", objectiveRuntimeSchedulerResumeDaemonFailureText(report.WorkerReport))
		return report.Normalize()
	}
	report = d.objectiveRuntimeSchedulerResumeDaemonBindQueueResult(ctx, report, job.ID)
	return report.Normalize()
}

func (d ObjectiveRuntimeSchedulerResumeDaemon) objectiveRuntimeSchedulerResumeDaemonFailJob(ctx context.Context, report ObjectiveRuntimeSchedulerResumeProcessReport, job scheduler.Job, reason string, text string) ObjectiveRuntimeSchedulerResumeProcessReport {
	report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, reason)
	if text == "" {
		text = reason
	}
	if d.Queue != nil {
		if err := d.Queue.Fail(ctx, scheduler.Result{
			JobID:   job.ID,
			Lane:    objectiveRuntimeSchedulerResumeDaemonLane(job.Lane, d.Config.Lane),
			Attempt: job.Attempt,
			Error:   text,
		}); err == nil {
			report.QueueFailed = true
			report.Boundaries = appendUniqueResumeStrings(report.Boundaries, "scheduler_tick_job_failure_recorded")
		} else {
			report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "scheduler_queue_fail_record_failed")
		}
		report = d.objectiveRuntimeSchedulerResumeDaemonBindQueueResult(ctx, report, job.ID)
	}
	report.Status = "objective_runtime_scheduler_resume_worker_failure_recorded"
	report.NextHostAction = "review_objective_runtime_scheduler_resume_worker_failure"
	return report
}

func (d ObjectiveRuntimeSchedulerResumeDaemon) objectiveRuntimeSchedulerResumeDaemonBindQueueResult(ctx context.Context, report ObjectiveRuntimeSchedulerResumeProcessReport, jobID string) ObjectiveRuntimeSchedulerResumeProcessReport {
	if d.Queue == nil || strings.TrimSpace(jobID) == "" {
		return report
	}
	result, ok, err := d.Queue.Result(ctx, jobID)
	if err != nil || !ok {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:scheduler_queue_result_readback")
		return report
	}
	report.QueueResultReadbackReady = true
	report.QueueResultStatus = strings.TrimSpace(result.Status)
	report.QueueResultSucceeded = result.Succeeded
	return report
}

func (report ObjectiveRuntimeSchedulerResumeTickEnqueueReport) Normalize() ObjectiveRuntimeSchedulerResumeTickEnqueueReport {
	out := report
	out.Status = strings.TrimSpace(out.Status)
	if out.Status == "" {
		out.Status = "blocked"
	}
	out.DaemonKind = strings.TrimSpace(out.DaemonKind)
	if out.DaemonKind == "" {
		out.DaemonKind = ObjectiveRuntimeSchedulerResumeDaemonKind
	}
	out.JobID = strings.TrimSpace(out.JobID)
	out.JobKind = strings.TrimSpace(out.JobKind)
	out.Lane = strings.TrimSpace(out.Lane)
	out.TickRef = strings.TrimSpace(out.TickRef)
	out.SchedulerRuntimeQueueRef = strings.TrimSpace(out.SchedulerRuntimeQueueRef)
	out.SchedulerWakeContinuationRef = strings.TrimSpace(out.SchedulerWakeContinuationRef)
	out.ObjectiveRunRef = strings.TrimSpace(out.ObjectiveRunRef)
	out.ObjectiveGraphSnapshotRef = strings.TrimSpace(out.ObjectiveGraphSnapshotRef)
	out.ObjectiveGraphRef = strings.TrimSpace(out.ObjectiveGraphRef)
	out.ObjectiveGraphReadbackRef = strings.TrimSpace(out.ObjectiveGraphReadbackRef)
	out.WorkerRef = strings.TrimSpace(out.WorkerRef)
	out.ProducerRef = strings.TrimSpace(out.ProducerRef)
	out.QueueResultStatus = strings.TrimSpace(out.QueueResultStatus)
	out.MissingInputs = appendUniqueResumeStrings(nil, out.MissingInputs...)
	out.BlockedReasons = appendUniqueResumeStrings(nil, out.BlockedReasons...)
	out.Boundaries = appendUniqueResumeStrings(nil, out.Boundaries...)
	out.NextHostAction = strings.TrimSpace(out.NextHostAction)
	if out.NextHostAction == "" {
		out.NextHostAction = "review_objective_runtime_scheduler_tick_enqueue"
	}
	out.QueueMutationByCore = false
	out.QueueMutationByHost = out.QueueMutationByHost && out.TickEnqueued
	out.Available = out.Available && out.QueueAvailable
	if out.Status == "objective_runtime_scheduler_tick_enqueued" && !out.TickEnqueued {
		out.Status = "blocked"
	}
	return out
}

func (report ObjectiveRuntimeSchedulerResumeProcessReport) Normalize() ObjectiveRuntimeSchedulerResumeProcessReport {
	out := report
	out.Status = strings.TrimSpace(out.Status)
	if out.Status == "" {
		out.Status = "blocked"
	}
	out.DaemonKind = strings.TrimSpace(out.DaemonKind)
	if out.DaemonKind == "" {
		out.DaemonKind = ObjectiveRuntimeSchedulerResumeDaemonKind
	}
	out.JobID = strings.TrimSpace(out.JobID)
	out.JobKind = strings.TrimSpace(out.JobKind)
	out.Lane = strings.TrimSpace(out.Lane)
	out.TickRef = strings.TrimSpace(out.TickRef)
	out.SchedulerRuntimeQueueRef = strings.TrimSpace(out.SchedulerRuntimeQueueRef)
	out.SchedulerWakeContinuationRef = strings.TrimSpace(out.SchedulerWakeContinuationRef)
	out.ObjectiveRunRef = strings.TrimSpace(out.ObjectiveRunRef)
	out.ObjectiveGraphSnapshotRef = strings.TrimSpace(out.ObjectiveGraphSnapshotRef)
	out.ObjectiveGraphRef = strings.TrimSpace(out.ObjectiveGraphRef)
	out.ObjectiveGraphReadbackRef = strings.TrimSpace(out.ObjectiveGraphReadbackRef)
	out.ObjectiveGraphState = strings.TrimSpace(out.ObjectiveGraphState)
	if out.ObjectiveGraphRevision < 0 {
		out.ObjectiveGraphRevision = 0
	}
	out.ReadyNodeRefs = appendUniqueResumeStrings(nil, out.ReadyNodeRefs...)
	out.WorkerRef = strings.TrimSpace(out.WorkerRef)
	out.QueueResultStatus = strings.TrimSpace(out.QueueResultStatus)
	out.MissingInputs = appendUniqueResumeStrings(nil, out.MissingInputs...)
	out.BlockedReasons = appendUniqueResumeStrings(nil, out.BlockedReasons...)
	out.Boundaries = appendUniqueResumeStrings(nil, out.Boundaries...)
	out.NextHostAction = strings.TrimSpace(out.NextHostAction)
	if out.NextHostAction == "" {
		out.NextHostAction = "review_objective_runtime_scheduler_resume_worker"
	}
	out.Available = out.Available && out.QueueAvailable && out.WorkerAvailable
	out.LLMWakeDispatched = out.LLMWakeDispatched && out.RunnerDispatched && out.HostRuntimeDispatchByHost
	out.RunnerDispatched = false
	out.RuntimeAdapterExecuted = false
	out.ToolExecuted = false
	out.WorkflowDispatched = false
	out.SchedulerApplied = false
	out.InstallerExecuted = false
	out.StoreMutationByCore = false
	out.WorkerMutationByHost = out.WorkerMutationByHost && out.DispatchRequestRecorded
	if out.Status == "objective_runtime_scheduler_resume_worker_ack_recorded" && (!out.QueueAcked || !out.QueueResultReadbackReady || !out.QueueResultSucceeded) {
		out.Status = "blocked"
	}
	return out
}

func objectiveRuntimeSchedulerResumeDaemonPayload(payload ObjectiveRuntimeSchedulerResumeTickPayload, config ObjectiveRuntimeSchedulerResumeWorkerConfig) ObjectiveRuntimeSchedulerResumeTickPayload {
	payload.ContinuationStoreRef = firstObjectiveRuntimeSchedulerResumeRef(payload.ContinuationStoreRef, config.ContinuationStoreRef)
	payload.DispatchRef = firstObjectiveRuntimeSchedulerResumeRef(payload.DispatchRef, config.DispatchRef)
	payload.RuntimeDispatchRef = firstObjectiveRuntimeSchedulerResumeRef(payload.RuntimeDispatchRef, config.RuntimeDispatchRef)
	payload.HostRunnerRef = firstObjectiveRuntimeSchedulerResumeRef(payload.HostRunnerRef, config.HostRunnerRef)
	payload.HostRunnerVersionRef = firstObjectiveRuntimeSchedulerResumeRef(payload.HostRunnerVersionRef, config.HostRunnerVersionRef)
	payload.OperatorApprovalRef = firstObjectiveRuntimeSchedulerResumeRef(payload.OperatorApprovalRef, config.OperatorApprovalRef)
	return payload
}

func objectiveRuntimeSchedulerResumeDaemonLane(values ...scheduler.Lane) scheduler.Lane {
	for _, value := range values {
		switch value {
		case scheduler.LaneMain, scheduler.LaneSubtask, scheduler.LaneBackground:
			return value
		}
	}
	return scheduler.LaneBackground
}

func objectiveRuntimeSchedulerResumeQueueRuntimeVisible(queue scheduler.Queue) bool {
	if visible, ok := queue.(scheduler.RuntimeVisibleQueue); ok && visible != nil {
		return visible.HasRuntimeVisibility()
	}
	return false
}

func objectiveRuntimeSchedulerResumeDaemonBindWorkerReport(report ObjectiveRuntimeSchedulerResumeProcessReport, worker ObjectiveRuntimeSchedulerResumeWorkerReport) ObjectiveRuntimeSchedulerResumeProcessReport {
	report.TickRef = worker.TickRef
	report.SchedulerRuntimeQueueRef = worker.SchedulerRuntimeQueueRef
	report.SchedulerWakeContinuationRef = worker.SchedulerWakeContinuationRef
	report.ObjectiveRunRef = worker.ObjectiveRunRef
	report.ObjectiveGraphSnapshotRef = worker.ObjectiveGraphSnapshotRef
	report.ObjectiveGraphRef = worker.ObjectiveGraphRef
	report.ObjectiveGraphReadbackRef = worker.ObjectiveGraphReadbackRef
	report.ObjectiveGraphState = worker.ObjectiveGraphState
	report.ObjectiveGraphRevision = worker.ObjectiveGraphRevision
	report.ReadyNodeRefs = appendUniqueResumeStrings(report.ReadyNodeRefs, worker.ReadyNodeRefs...)
	report.WorkerMutationByHost = worker.WorkerMutationByHost
	report.ReadyForRuntimeWakeDispatch = worker.ReadyForRuntimeWakeDispatch
	report.DispatchRequestRecorded = worker.DispatchRequestRecorded
	report.HostRuntimeDispatchByHost = worker.HostRuntimeDispatchByHost
	report.LLMWakeDispatched = worker.LLMWakeDispatched
	report.RunnerDispatched = worker.RunnerDispatched
	report.RuntimeAdapterExecuted = worker.RuntimeAdapterExecuted
	report.ToolExecuted = worker.ToolExecuted
	report.WorkflowDispatched = worker.WorkflowDispatched
	report.SchedulerApplied = worker.SchedulerApplied
	report.InstallerExecuted = worker.InstallerExecuted
	report.StoreMutationByCore = worker.StoreMutationByCore
	report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, worker.MissingInputs...)
	report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, worker.BlockedReasons...)
	report.Boundaries = appendUniqueResumeStrings(report.Boundaries, worker.Boundaries...)
	report.NextHostAction = firstNonEmptyString(worker.NextHostAction, report.NextHostAction)
	return report
}

func objectiveRuntimeSchedulerResumeDaemonFailureText(worker ObjectiveRuntimeSchedulerResumeWorkerReport) string {
	if value := firstNonEmptyString(worker.BlockedReasons...); value != "" {
		return value
	}
	if value := firstNonEmptyString(worker.MissingInputs...); value != "" {
		return value
	}
	return firstNonEmptyString(worker.Status, "objective_runtime_scheduler_resume_worker_not_ready")
}
