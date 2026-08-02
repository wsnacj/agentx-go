package resume

import (
	"context"
	"strings"

	scheduler "github.com/wsnacj/agentx-go/runtime/scheduler"
)

const ObjectiveRuntimeSchedulerResumeDaemonRunnerKind = "host_objective_runtime_scheduler_resume_daemon_runner"

type ObjectiveRuntimeSchedulerResumeDaemonRunInput struct {
	Enabled           bool
	Lane              scheduler.Lane
	MaxTicks          int
	ContinueOnIdle    bool
	ContinueOnFailure bool
	Boundaries        []string
}

type ObjectiveRuntimeSchedulerResumeDaemonRunReport struct {
	Available                 bool                                           `json:"available"`
	Enabled                   bool                                           `json:"enabled"`
	Status                    string                                         `json:"status,omitempty"`
	RunnerKind                string                                         `json:"runner_kind,omitempty"`
	DaemonKind                string                                         `json:"daemon_kind,omitempty"`
	QueueAvailable            bool                                           `json:"queue_available"`
	QueueRuntimeVisible       bool                                           `json:"queue_runtime_visible"`
	KindAwareQueue            bool                                           `json:"kind_aware_queue"`
	WorkerAvailable           bool                                           `json:"worker_available"`
	MaxTicks                  int                                            `json:"max_ticks,omitempty"`
	TicksAttempted            int                                            `json:"ticks_attempted,omitempty"`
	TicksLeased               int                                            `json:"ticks_leased,omitempty"`
	TicksAcked                int                                            `json:"ticks_acked,omitempty"`
	TicksFailed               int                                            `json:"ticks_failed,omitempty"`
	TicksIdle                 int                                            `json:"ticks_idle,omitempty"`
	TicksBlocked              int                                            `json:"ticks_blocked,omitempty"`
	WorkerCalls               int                                            `json:"worker_calls,omitempty"`
	QueueMutationByHost       bool                                           `json:"queue_mutation_by_host"`
	QueueMutationByCore       bool                                           `json:"queue_mutation_by_core"`
	WorkerMutationByHost      bool                                           `json:"worker_mutation_by_host"`
	HostRuntimeDispatchByHost bool                                           `json:"host_runtime_dispatch_by_host"`
	LLMWakeDispatched         bool                                           `json:"llm_wake_dispatched"`
	RunnerDispatched          bool                                           `json:"runner_dispatched"`
	RuntimeAdapterExecuted    bool                                           `json:"runtime_adapter_executed"`
	ToolExecuted              bool                                           `json:"tool_executed"`
	WorkflowDispatched        bool                                           `json:"workflow_dispatched"`
	SchedulerApplied          bool                                           `json:"scheduler_applied"`
	InstallerExecuted         bool                                           `json:"installer_executed"`
	StoreMutationByCore       bool                                           `json:"store_mutation_by_core"`
	ContextDone               bool                                           `json:"context_done"`
	Lane                      string                                         `json:"lane,omitempty"`
	LastJobID                 string                                         `json:"last_job_id,omitempty"`
	LastJobKind               string                                         `json:"last_job_kind,omitempty"`
	LastObjectiveRunRef       string                                         `json:"last_objective_run_ref,omitempty"`
	LastGraphSnapshotRef      string                                         `json:"last_objective_graph_snapshot_ref,omitempty"`
	LastGraphRef              string                                         `json:"last_objective_graph_ref,omitempty"`
	LastGraphReadbackRef      string                                         `json:"last_objective_graph_readback_ref,omitempty"`
	LastGraphState            string                                         `json:"last_objective_graph_state,omitempty"`
	LastGraphRevision         int                                            `json:"last_objective_graph_revision,omitempty"`
	LastReadyNodeRefs         []string                                       `json:"last_ready_node_refs,omitempty"`
	LastProcessStatus         string                                         `json:"last_process_status,omitempty"`
	LastProcessNextHostAction string                                         `json:"last_process_next_host_action,omitempty"`
	MissingInputs             []string                                       `json:"missing_inputs,omitempty"`
	BlockedReasons            []string                                       `json:"blocked_reasons,omitempty"`
	Boundaries                []string                                       `json:"boundaries,omitempty"`
	NextHostAction            string                                         `json:"next_host_action,omitempty"`
	ProcessReports            []ObjectiveRuntimeSchedulerResumeProcessReport `json:"process_reports,omitempty"`
}

func (d ObjectiveRuntimeSchedulerResumeDaemon) RunSchedulerResumeDaemon(ctx context.Context, input ObjectiveRuntimeSchedulerResumeDaemonRunInput) ObjectiveRuntimeSchedulerResumeDaemonRunReport {
	lane := objectiveRuntimeSchedulerResumeDaemonLane(input.Lane, d.Config.Lane)
	report := ObjectiveRuntimeSchedulerResumeDaemonRunReport{
		Status:              "blocked",
		RunnerKind:          ObjectiveRuntimeSchedulerResumeDaemonRunnerKind,
		DaemonKind:          ObjectiveRuntimeSchedulerResumeDaemonKind,
		QueueAvailable:      d.Queue != nil,
		QueueRuntimeVisible: objectiveRuntimeSchedulerResumeQueueRuntimeVisible(d.Queue),
		WorkerAvailable:     d.Worker.ContinuationReadback != nil && d.Worker.WakeDispatch != nil,
		Lane:                string(lane),
		Boundaries: appendUniqueResumeStrings([]string{
			"host_owned_objective_runtime_scheduler_resume_daemon_runner",
			"bounded_scheduler_resume_worker_loop",
			"display_safe_refs_only",
			"no_scheduler_apply_by_core",
			"no_runner_dispatch_by_core",
			"no_runtime_adapter_execution_by_core",
			"no_tool_execution_by_core",
			"no_workflow_dispatch_by_core",
			"no_installer_execution_by_core",
			"no_store_mutation_by_core",
		}, input.Boundaries...),
		NextHostAction: "review_objective_runtime_scheduler_resume_daemon_runner",
	}
	if kindQueue, ok := d.Queue.(scheduler.KindAwareQueue); ok && kindQueue != nil {
		report.KindAwareQueue = true
	}
	if !input.Enabled {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:objective_runtime_scheduler_resume_daemon_runner_enabled")
		report.Boundaries = appendUniqueResumeStrings(report.Boundaries, "objective_runtime_scheduler_resume_daemon_runner_default_off")
		report.NextHostAction = "enable_objective_runtime_scheduler_resume_daemon_runner"
		return report.Normalize()
	}
	report.Enabled = true
	report.Available = d.Queue != nil && report.WorkerAvailable
	report.MaxTicks = objectiveRuntimeSchedulerResumeDaemonRunMaxTicks(input.MaxTicks)
	if ctx == nil {
		ctx = context.Background()
	}

	for report.TicksAttempted < report.MaxTicks {
		if err := ctx.Err(); err != nil {
			report.ContextDone = true
			report.Status = "objective_runtime_scheduler_resume_daemon_runner_context_done"
			report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "context_done")
			report.NextHostAction = "review_objective_runtime_scheduler_resume_daemon_runner_context"
			return report.Normalize()
		}
		report.TicksAttempted++
		process := d.ProcessNextSchedulerTick(ctx, ObjectiveRuntimeSchedulerResumeProcessInput{
			Enabled: true,
			Lane:    lane,
		}).Normalize()
		report.ProcessReports = append(report.ProcessReports, process)
		report = objectiveRuntimeSchedulerResumeDaemonRunBindProcess(report, process)

		switch process.Status {
		case "objective_runtime_scheduler_resume_worker_ack_recorded":
			report.TicksAcked++
			report.Status = "objective_runtime_scheduler_resume_daemon_runner_completed"
			report.NextHostAction = firstNonEmptyString(process.NextHostAction, "schedule_next_daemon_tick_or_continue_worker")
			continue
		case "objective_runtime_scheduler_resume_worker_idle":
			report.TicksIdle++
			report.Status = "objective_runtime_scheduler_resume_daemon_runner_idle"
			report.NextHostAction = firstNonEmptyString(process.NextHostAction, "wait_for_scheduler_tick")
			if !input.ContinueOnIdle {
				return report.Normalize()
			}
		case "objective_runtime_scheduler_resume_worker_failure_recorded":
			report.TicksFailed++
			report.Status = "objective_runtime_scheduler_resume_daemon_runner_worker_failure"
			report.NextHostAction = firstNonEmptyString(process.NextHostAction, "review_objective_runtime_scheduler_resume_worker_failure")
			if !input.ContinueOnFailure {
				return report.Normalize()
			}
		default:
			report.TicksBlocked++
			report.Status = "objective_runtime_scheduler_resume_daemon_runner_blocked"
			report.NextHostAction = firstNonEmptyString(process.NextHostAction, "review_objective_runtime_scheduler_resume_daemon_runner")
			if !input.ContinueOnFailure {
				return report.Normalize()
			}
		}
	}
	return report.Normalize()
}

func (report ObjectiveRuntimeSchedulerResumeDaemonRunReport) Normalize() ObjectiveRuntimeSchedulerResumeDaemonRunReport {
	report.Status = strings.TrimSpace(report.Status)
	if report.Status == "" {
		report.Status = "blocked"
	}
	report.RunnerKind = strings.TrimSpace(report.RunnerKind)
	if report.RunnerKind == "" {
		report.RunnerKind = ObjectiveRuntimeSchedulerResumeDaemonRunnerKind
	}
	report.DaemonKind = strings.TrimSpace(report.DaemonKind)
	if report.DaemonKind == "" {
		report.DaemonKind = ObjectiveRuntimeSchedulerResumeDaemonKind
	}
	report.Lane = strings.TrimSpace(report.Lane)
	report.LastJobID = strings.TrimSpace(report.LastJobID)
	report.LastJobKind = strings.TrimSpace(report.LastJobKind)
	report.LastObjectiveRunRef = strings.TrimSpace(report.LastObjectiveRunRef)
	report.LastGraphSnapshotRef = strings.TrimSpace(report.LastGraphSnapshotRef)
	report.LastGraphRef = strings.TrimSpace(report.LastGraphRef)
	report.LastGraphReadbackRef = strings.TrimSpace(report.LastGraphReadbackRef)
	report.LastGraphState = strings.TrimSpace(report.LastGraphState)
	if report.LastGraphRevision < 0 {
		report.LastGraphRevision = 0
	}
	report.LastReadyNodeRefs = appendUniqueResumeStrings(nil, report.LastReadyNodeRefs...)
	report.LastProcessStatus = strings.TrimSpace(report.LastProcessStatus)
	report.LastProcessNextHostAction = strings.TrimSpace(report.LastProcessNextHostAction)
	report.MissingInputs = appendUniqueResumeStrings(nil, report.MissingInputs...)
	report.BlockedReasons = appendUniqueResumeStrings(nil, report.BlockedReasons...)
	report.Boundaries = appendUniqueResumeStrings(nil, report.Boundaries...)
	report.NextHostAction = strings.TrimSpace(report.NextHostAction)
	if report.NextHostAction == "" {
		report.NextHostAction = "review_objective_runtime_scheduler_resume_daemon_runner"
	}
	report.Available = report.Enabled && report.QueueAvailable && report.WorkerAvailable
	report.QueueMutationByCore = false
	report.LLMWakeDispatched = false
	report.RunnerDispatched = false
	report.RuntimeAdapterExecuted = false
	report.ToolExecuted = false
	report.WorkflowDispatched = false
	report.SchedulerApplied = false
	report.InstallerExecuted = false
	report.StoreMutationByCore = false
	report.QueueMutationByHost = report.TicksAcked > 0 || report.TicksFailed > 0
	if report.Status == "objective_runtime_scheduler_resume_daemon_runner_completed" && report.TicksAcked == 0 {
		report.Status = "blocked"
	}
	if report.MaxTicks <= 0 && report.Enabled {
		report.MaxTicks = 1
	}
	return report
}

func objectiveRuntimeSchedulerResumeDaemonRunBindProcess(report ObjectiveRuntimeSchedulerResumeDaemonRunReport, process ObjectiveRuntimeSchedulerResumeProcessReport) ObjectiveRuntimeSchedulerResumeDaemonRunReport {
	report.LastProcessStatus = process.Status
	report.LastProcessNextHostAction = process.NextHostAction
	report.LastJobID = process.JobID
	report.LastJobKind = process.JobKind
	report.LastObjectiveRunRef = process.ObjectiveRunRef
	report.LastGraphSnapshotRef = process.ObjectiveGraphSnapshotRef
	report.LastGraphRef = process.ObjectiveGraphRef
	report.LastGraphReadbackRef = process.ObjectiveGraphReadbackRef
	report.LastGraphState = process.ObjectiveGraphState
	report.LastGraphRevision = process.ObjectiveGraphRevision
	report.LastReadyNodeRefs = appendUniqueResumeStrings(nil, process.ReadyNodeRefs...)
	if process.LeaseAcquired {
		report.TicksLeased++
	}
	if process.WorkerCalled {
		report.WorkerCalls++
	}
	report.WorkerMutationByHost = report.WorkerMutationByHost || process.WorkerMutationByHost
	report.HostRuntimeDispatchByHost = report.HostRuntimeDispatchByHost || process.HostRuntimeDispatchByHost
	report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, process.MissingInputs...)
	report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, process.BlockedReasons...)
	report.Boundaries = appendUniqueResumeStrings(report.Boundaries, process.Boundaries...)
	return report
}

func objectiveRuntimeSchedulerResumeDaemonRunMaxTicks(value int) int {
	if value <= 0 {
		return 1
	}
	return value
}
