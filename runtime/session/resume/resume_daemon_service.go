package resume

import (
	"context"
	"strings"
	"time"

	scheduler "github.com/wsnacj/agentx-go/runtime/scheduler"
	session "github.com/wsnacj/agentx-go/runtime/session"
)

const ObjectiveRuntimeSchedulerResumeDaemonServiceKind = "host_objective_runtime_scheduler_resume_daemon_service"

type ObjectiveRuntimeSchedulerResumeDaemonService struct {
	Daemon ObjectiveRuntimeSchedulerResumeDaemon
	Wait   ObjectiveRuntimeSchedulerResumeDaemonServiceWaitFunc
}

type ObjectiveRuntimeSchedulerResumeDaemonServiceWaitFunc func(context.Context, ObjectiveRuntimeSchedulerResumeDaemonServiceWaitInput) error

type ObjectiveRuntimeSchedulerResumeDaemonServiceWaitInput struct {
	CycleIndex    int
	CycleInterval time.Duration
	ServiceRef    session.DisplaySafeRef
	ConfigRef     session.DisplaySafeRef
	DeploymentRef session.DisplaySafeRef
}

type ObjectiveRuntimeSchedulerResumeDaemonServiceInput struct {
	Enabled                                bool
	Lane                                   scheduler.Lane
	MaxCycles                              int
	MaxTicksPerCycle                       int
	ContinueOnIdle                         bool
	ContinueOnFailure                      bool
	CycleInterval                          time.Duration
	ServiceRef                             session.DisplaySafeRef
	ConfigRef                              session.DisplaySafeRef
	DeploymentRef                          session.DisplaySafeRef
	OperatorApprovalRef                    session.DisplaySafeRef
	InactivityWatchdogEnabled              bool
	InactivityWatchdogRef                  session.DisplaySafeRef
	InactivityWatchdogReviewRef            session.DisplaySafeRef
	InactivityWatchdogHumanInterventionRef session.DisplaySafeRef
	MaxConsecutiveIdleCycles               int
	MaxConsecutiveBlockedCycles            int
	MaxConsecutiveFailureCycles            int
	MaxConsecutiveNoProgressCycles         int
	Boundaries                             []string
}

type ObjectiveRuntimeSchedulerResumeDaemonServiceReport struct {
	Available                              bool                                             `json:"available"`
	Enabled                                bool                                             `json:"enabled"`
	Status                                 string                                           `json:"status,omitempty"`
	ServiceKind                            string                                           `json:"service_kind,omitempty"`
	RunnerKind                             string                                           `json:"runner_kind,omitempty"`
	DaemonKind                             string                                           `json:"daemon_kind,omitempty"`
	ServiceConfigured                      bool                                             `json:"service_configured"`
	ServiceStartRequested                  bool                                             `json:"service_start_requested"`
	ServiceStartedByHost                   bool                                             `json:"service_started_by_host"`
	ServiceStopRequested                   bool                                             `json:"service_stop_requested"`
	ServiceStoppedByHost                   bool                                             `json:"service_stopped_by_host"`
	ServiceLoopCompleted                   bool                                             `json:"service_loop_completed"`
	ServiceMutationByCore                  bool                                             `json:"service_mutation_by_core"`
	InactivityWatchdogEnabled              bool                                             `json:"inactivity_watchdog_enabled"`
	InactivityWatchdogReady                bool                                             `json:"inactivity_watchdog_ready"`
	InactivityWatchdogTriggered            bool                                             `json:"inactivity_watchdog_triggered"`
	ReadyForWatchdogStop                   bool                                             `json:"ready_for_watchdog_stop"`
	ReadyForHumanIntervention              bool                                             `json:"ready_for_human_intervention"`
	QueueAvailable                         bool                                             `json:"queue_available"`
	QueueRuntimeVisible                    bool                                             `json:"queue_runtime_visible"`
	KindAwareQueue                         bool                                             `json:"kind_aware_queue"`
	WorkerAvailable                        bool                                             `json:"worker_available"`
	MaxCycles                              int                                              `json:"max_cycles,omitempty"`
	MaxTicksPerCycle                       int                                              `json:"max_ticks_per_cycle,omitempty"`
	CyclesStarted                          int                                              `json:"cycles_started,omitempty"`
	CyclesCompleted                        int                                              `json:"cycles_completed,omitempty"`
	CyclesIdle                             int                                              `json:"cycles_idle,omitempty"`
	CyclesFailed                           int                                              `json:"cycles_failed,omitempty"`
	CyclesBlocked                          int                                              `json:"cycles_blocked,omitempty"`
	CyclesContextDone                      int                                              `json:"cycles_context_done,omitempty"`
	ConsecutiveIdleCycles                  int                                              `json:"consecutive_idle_cycles,omitempty"`
	ConsecutiveBlockedCycles               int                                              `json:"consecutive_blocked_cycles,omitempty"`
	ConsecutiveFailureCycles               int                                              `json:"consecutive_failure_cycles,omitempty"`
	ConsecutiveNoProgressCycles            int                                              `json:"consecutive_no_progress_cycles,omitempty"`
	MaxConsecutiveIdleCycles               int                                              `json:"max_consecutive_idle_cycles,omitempty"`
	MaxConsecutiveBlockedCycles            int                                              `json:"max_consecutive_blocked_cycles,omitempty"`
	MaxConsecutiveFailureCycles            int                                              `json:"max_consecutive_failure_cycles,omitempty"`
	MaxConsecutiveNoProgressCycles         int                                              `json:"max_consecutive_no_progress_cycles,omitempty"`
	WaitsRequested                         int                                              `json:"waits_requested,omitempty"`
	WaitsCompleted                         int                                              `json:"waits_completed,omitempty"`
	TicksAttempted                         int                                              `json:"ticks_attempted,omitempty"`
	TicksLeased                            int                                              `json:"ticks_leased,omitempty"`
	TicksAcked                             int                                              `json:"ticks_acked,omitempty"`
	TicksFailed                            int                                              `json:"ticks_failed,omitempty"`
	TicksIdle                              int                                              `json:"ticks_idle,omitempty"`
	TicksBlocked                           int                                              `json:"ticks_blocked,omitempty"`
	WorkerCalls                            int                                              `json:"worker_calls,omitempty"`
	QueueMutationByHost                    bool                                             `json:"queue_mutation_by_host"`
	QueueMutationByCore                    bool                                             `json:"queue_mutation_by_core"`
	WorkerMutationByHost                   bool                                             `json:"worker_mutation_by_host"`
	HostRuntimeDispatchByHost              bool                                             `json:"host_runtime_dispatch_by_host"`
	LLMWakeDispatched                      bool                                             `json:"llm_wake_dispatched"`
	RunnerDispatched                       bool                                             `json:"runner_dispatched"`
	RuntimeAdapterExecuted                 bool                                             `json:"runtime_adapter_executed"`
	ToolExecuted                           bool                                             `json:"tool_executed"`
	WorkflowDispatched                     bool                                             `json:"workflow_dispatched"`
	SchedulerApplied                       bool                                             `json:"scheduler_applied"`
	InstallerExecuted                      bool                                             `json:"installer_executed"`
	StoreMutationByCore                    bool                                             `json:"store_mutation_by_core"`
	ContextDone                            bool                                             `json:"context_done"`
	CycleIntervalMillis                    int64                                            `json:"cycle_interval_millis,omitempty"`
	Lane                                   string                                           `json:"lane,omitempty"`
	ServiceRef                             string                                           `json:"service_ref,omitempty"`
	ConfigRef                              string                                           `json:"config_ref,omitempty"`
	DeploymentRef                          string                                           `json:"deployment_ref,omitempty"`
	OperatorApprovalRef                    string                                           `json:"operator_approval_ref,omitempty"`
	InactivityWatchdogRef                  string                                           `json:"inactivity_watchdog_ref,omitempty"`
	InactivityWatchdogReviewRef            string                                           `json:"inactivity_watchdog_review_ref,omitempty"`
	InactivityWatchdogHumanInterventionRef string                                           `json:"inactivity_watchdog_human_intervention_ref,omitempty"`
	InactivityWatchdogReason               string                                           `json:"inactivity_watchdog_reason,omitempty"`
	LastRunStatus                          string                                           `json:"last_run_status,omitempty"`
	LastRunNextHostAction                  string                                           `json:"last_run_next_host_action,omitempty"`
	LastJobID                              string                                           `json:"last_job_id,omitempty"`
	LastJobKind                            string                                           `json:"last_job_kind,omitempty"`
	LastObjectiveRunRef                    string                                           `json:"last_objective_run_ref,omitempty"`
	LastGraphSnapshotRef                   string                                           `json:"last_objective_graph_snapshot_ref,omitempty"`
	LastGraphRef                           string                                           `json:"last_objective_graph_ref,omitempty"`
	LastGraphReadbackRef                   string                                           `json:"last_objective_graph_readback_ref,omitempty"`
	LastGraphState                         string                                           `json:"last_objective_graph_state,omitempty"`
	LastGraphRevision                      int                                              `json:"last_objective_graph_revision,omitempty"`
	LastReadyNodeRefs                      []string                                         `json:"last_ready_node_refs,omitempty"`
	MissingInputs                          []string                                         `json:"missing_inputs,omitempty"`
	BlockedReasons                         []string                                         `json:"blocked_reasons,omitempty"`
	Boundaries                             []string                                         `json:"boundaries,omitempty"`
	NextHostAction                         string                                           `json:"next_host_action,omitempty"`
	RunReports                             []ObjectiveRuntimeSchedulerResumeDaemonRunReport `json:"run_reports,omitempty"`
}

func (s ObjectiveRuntimeSchedulerResumeDaemonService) Run(ctx context.Context, input ObjectiveRuntimeSchedulerResumeDaemonServiceInput) ObjectiveRuntimeSchedulerResumeDaemonServiceReport {
	lane := objectiveRuntimeSchedulerResumeDaemonLane(input.Lane, s.Daemon.Config.Lane)
	report := ObjectiveRuntimeSchedulerResumeDaemonServiceReport{
		Status:                                 "blocked",
		ServiceKind:                            ObjectiveRuntimeSchedulerResumeDaemonServiceKind,
		RunnerKind:                             ObjectiveRuntimeSchedulerResumeDaemonRunnerKind,
		DaemonKind:                             ObjectiveRuntimeSchedulerResumeDaemonKind,
		QueueAvailable:                         s.Daemon.Queue != nil,
		QueueRuntimeVisible:                    objectiveRuntimeSchedulerResumeQueueRuntimeVisible(s.Daemon.Queue),
		WorkerAvailable:                        s.Daemon.Worker.ContinuationReadback != nil && s.Daemon.Worker.WakeDispatch != nil,
		Lane:                                   string(lane),
		ServiceRef:                             string(objectiveRuntimeSchedulerResumeSafeRef(input.ServiceRef)),
		ConfigRef:                              string(objectiveRuntimeSchedulerResumeSafeRef(input.ConfigRef)),
		DeploymentRef:                          string(objectiveRuntimeSchedulerResumeSafeRef(input.DeploymentRef)),
		OperatorApprovalRef:                    string(objectiveRuntimeSchedulerResumeSafeRef(input.OperatorApprovalRef)),
		InactivityWatchdogEnabled:              input.InactivityWatchdogEnabled,
		InactivityWatchdogRef:                  string(objectiveRuntimeSchedulerResumeSafeRef(input.InactivityWatchdogRef)),
		InactivityWatchdogReviewRef:            string(objectiveRuntimeSchedulerResumeSafeRef(input.InactivityWatchdogReviewRef)),
		InactivityWatchdogHumanInterventionRef: string(objectiveRuntimeSchedulerResumeSafeRef(input.InactivityWatchdogHumanInterventionRef)),
		Boundaries: appendUniqueResumeStrings([]string{
			"host_owned_objective_runtime_scheduler_resume_daemon_service",
			"bounded_service_lifecycle_by_host",
			"service_start_stop_by_host",
			"display_safe_refs_only",
			"no_background_process_started_by_core",
			"no_scheduler_apply_by_core",
			"no_runner_dispatch_by_core",
			"no_runtime_adapter_execution_by_core",
			"no_tool_execution_by_core",
			"no_workflow_dispatch_by_core",
			"no_installer_execution_by_core",
			"no_store_mutation_by_core",
		}, input.Boundaries...),
		NextHostAction: "review_objective_runtime_scheduler_resume_daemon_service",
	}
	if kindQueue, ok := s.Daemon.Queue.(scheduler.KindAwareQueue); ok && kindQueue != nil {
		report.KindAwareQueue = true
	}
	if !input.Enabled {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:objective_runtime_scheduler_resume_daemon_service_enabled")
		report.Boundaries = appendUniqueResumeStrings(report.Boundaries, "objective_runtime_scheduler_resume_daemon_service_default_off")
		report.NextHostAction = "enable_objective_runtime_scheduler_resume_daemon_service"
		return report.Normalize()
	}
	report.Enabled = true
	if input.CycleInterval < 0 {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:objective_runtime_scheduler_resume_daemon_service_cycle_interval")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "objective_runtime_scheduler_resume_daemon_service_negative_cycle_interval")
		report.NextHostAction = "review_objective_runtime_scheduler_resume_daemon_service_config"
		return report.Normalize()
	}
	if !objectiveRuntimeSchedulerResumeDaemonServiceValidateWatchdog(input, &report) {
		return report.Normalize()
	}
	report.ServiceConfigured = true
	report.Available = s.Daemon.Queue != nil && report.WorkerAvailable
	report.MaxCycles = objectiveRuntimeSchedulerResumeDaemonServiceMax(input.MaxCycles)
	report.MaxTicksPerCycle = objectiveRuntimeSchedulerResumeDaemonServiceMax(input.MaxTicksPerCycle)
	report.MaxConsecutiveIdleCycles = input.MaxConsecutiveIdleCycles
	report.MaxConsecutiveBlockedCycles = input.MaxConsecutiveBlockedCycles
	report.MaxConsecutiveFailureCycles = input.MaxConsecutiveFailureCycles
	report.MaxConsecutiveNoProgressCycles = input.MaxConsecutiveNoProgressCycles
	report.CycleIntervalMillis = input.CycleInterval.Milliseconds()
	if ctx == nil {
		ctx = context.Background()
	}

	report.ServiceStartRequested = true
	report.ServiceStartedByHost = true
	report.NextHostAction = "run_objective_runtime_scheduler_resume_daemon_service_cycle"
	for report.CyclesStarted < report.MaxCycles {
		if err := ctx.Err(); err != nil {
			report.ContextDone = true
			report.CyclesContextDone++
			report.Status = "objective_runtime_scheduler_resume_daemon_service_context_done"
			report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "context_done")
			report.NextHostAction = "review_objective_runtime_scheduler_resume_daemon_service_context"
			break
		}
		report.CyclesStarted++
		run := s.Daemon.RunSchedulerResumeDaemon(ctx, ObjectiveRuntimeSchedulerResumeDaemonRunInput{
			Enabled:           true,
			Lane:              lane,
			MaxTicks:          report.MaxTicksPerCycle,
			ContinueOnFailure: input.ContinueOnFailure,
			Boundaries:        []string{"objective_runtime_scheduler_resume_daemon_service_cycle"},
		}).Normalize()
		report.RunReports = append(report.RunReports, run)
		report = objectiveRuntimeSchedulerResumeDaemonServiceBindRun(report, run)
		shouldContinue := objectiveRuntimeSchedulerResumeDaemonServiceShouldContinue(&report, run, input)
		if objectiveRuntimeSchedulerResumeDaemonServiceWatchdogTriggered(&report, input) {
			break
		}
		if !shouldContinue {
			break
		}
		if report.CyclesStarted >= report.MaxCycles {
			break
		}
		if !s.waitForNextCycle(ctx, input, report.CyclesStarted, &report) {
			break
		}
	}
	report.ServiceStopRequested = true
	report.ServiceStoppedByHost = report.ServiceStartedByHost
	if report.Status == "objective_runtime_scheduler_resume_daemon_service_running" {
		report.Status = "objective_runtime_scheduler_resume_daemon_service_completed"
		report.NextHostAction = "review_objective_runtime_scheduler_resume_daemon_service_report"
	}
	report.ServiceLoopCompleted = !report.ContextDone
	return report.Normalize()
}

func (s ObjectiveRuntimeSchedulerResumeDaemonService) waitForNextCycle(ctx context.Context, input ObjectiveRuntimeSchedulerResumeDaemonServiceInput, cycleIndex int, report *ObjectiveRuntimeSchedulerResumeDaemonServiceReport) bool {
	report.WaitsRequested++
	waitInput := ObjectiveRuntimeSchedulerResumeDaemonServiceWaitInput{
		CycleIndex:    cycleIndex,
		CycleInterval: input.CycleInterval,
		ServiceRef:    objectiveRuntimeSchedulerResumeSafeRef(input.ServiceRef),
		ConfigRef:     objectiveRuntimeSchedulerResumeSafeRef(input.ConfigRef),
		DeploymentRef: objectiveRuntimeSchedulerResumeSafeRef(input.DeploymentRef),
	}
	if s.Wait != nil {
		if err := s.Wait(ctx, waitInput); err != nil {
			report.Status = "objective_runtime_scheduler_resume_daemon_service_wait_failed"
			report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "objective_runtime_scheduler_resume_daemon_service_wait_failed")
			report.NextHostAction = "review_objective_runtime_scheduler_resume_daemon_service_wait"
			return false
		}
		report.WaitsCompleted++
		return true
	}
	if input.CycleInterval <= 0 {
		report.WaitsCompleted++
		return true
	}
	timer := time.NewTimer(input.CycleInterval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		report.ContextDone = true
		report.CyclesContextDone++
		report.Status = "objective_runtime_scheduler_resume_daemon_service_context_done"
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "context_done")
		report.NextHostAction = "review_objective_runtime_scheduler_resume_daemon_service_context"
		return false
	case <-timer.C:
		report.WaitsCompleted++
		return true
	}
}

func (report ObjectiveRuntimeSchedulerResumeDaemonServiceReport) Normalize() ObjectiveRuntimeSchedulerResumeDaemonServiceReport {
	report.Status = strings.TrimSpace(report.Status)
	if report.Status == "" {
		report.Status = "blocked"
	}
	report.ServiceKind = strings.TrimSpace(report.ServiceKind)
	if report.ServiceKind == "" {
		report.ServiceKind = ObjectiveRuntimeSchedulerResumeDaemonServiceKind
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
	report.ServiceRef = strings.TrimSpace(report.ServiceRef)
	report.ConfigRef = strings.TrimSpace(report.ConfigRef)
	report.DeploymentRef = strings.TrimSpace(report.DeploymentRef)
	report.OperatorApprovalRef = strings.TrimSpace(report.OperatorApprovalRef)
	report.InactivityWatchdogRef = strings.TrimSpace(report.InactivityWatchdogRef)
	report.InactivityWatchdogReviewRef = strings.TrimSpace(report.InactivityWatchdogReviewRef)
	report.InactivityWatchdogHumanInterventionRef = strings.TrimSpace(report.InactivityWatchdogHumanInterventionRef)
	report.InactivityWatchdogReason = strings.TrimSpace(report.InactivityWatchdogReason)
	report.LastRunStatus = strings.TrimSpace(report.LastRunStatus)
	report.LastRunNextHostAction = strings.TrimSpace(report.LastRunNextHostAction)
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
	report.MissingInputs = appendUniqueResumeStrings(nil, report.MissingInputs...)
	report.BlockedReasons = appendUniqueResumeStrings(nil, report.BlockedReasons...)
	report.Boundaries = appendUniqueResumeStrings(nil, report.Boundaries...)
	report.NextHostAction = strings.TrimSpace(report.NextHostAction)
	if report.NextHostAction == "" {
		report.NextHostAction = "review_objective_runtime_scheduler_resume_daemon_service"
	}
	report.Available = report.Enabled && report.ServiceConfigured && report.QueueAvailable && report.WorkerAvailable
	report.ServiceMutationByCore = false
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
	if report.MaxCycles <= 0 && report.Enabled {
		report.MaxCycles = 1
	}
	if report.MaxTicksPerCycle <= 0 && report.Enabled {
		report.MaxTicksPerCycle = 1
	}
	if !report.InactivityWatchdogEnabled {
		report.InactivityWatchdogReady = false
	}
	if report.Status == "objective_runtime_scheduler_resume_daemon_service_completed" && report.CyclesCompleted == 0 {
		report.Status = "blocked"
	}
	return report
}

func objectiveRuntimeSchedulerResumeDaemonServiceBindRun(report ObjectiveRuntimeSchedulerResumeDaemonServiceReport, run ObjectiveRuntimeSchedulerResumeDaemonRunReport) ObjectiveRuntimeSchedulerResumeDaemonServiceReport {
	report.LastRunStatus = run.Status
	report.LastRunNextHostAction = run.NextHostAction
	report.LastJobID = run.LastJobID
	report.LastJobKind = run.LastJobKind
	report.LastObjectiveRunRef = run.LastObjectiveRunRef
	report.LastGraphSnapshotRef = run.LastGraphSnapshotRef
	report.LastGraphRef = run.LastGraphRef
	report.LastGraphReadbackRef = run.LastGraphReadbackRef
	report.LastGraphState = run.LastGraphState
	report.LastGraphRevision = run.LastGraphRevision
	report.LastReadyNodeRefs = appendUniqueResumeStrings(nil, run.LastReadyNodeRefs...)
	report.TicksAttempted += run.TicksAttempted
	report.TicksLeased += run.TicksLeased
	report.TicksAcked += run.TicksAcked
	report.TicksFailed += run.TicksFailed
	report.TicksIdle += run.TicksIdle
	report.TicksBlocked += run.TicksBlocked
	report.WorkerCalls += run.WorkerCalls
	report.WorkerMutationByHost = report.WorkerMutationByHost || run.WorkerMutationByHost
	report.HostRuntimeDispatchByHost = report.HostRuntimeDispatchByHost || run.HostRuntimeDispatchByHost
	report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, run.MissingInputs...)
	report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, run.BlockedReasons...)
	report.Boundaries = appendUniqueResumeStrings(report.Boundaries, run.Boundaries...)
	return report
}

func objectiveRuntimeSchedulerResumeDaemonServiceShouldContinue(report *ObjectiveRuntimeSchedulerResumeDaemonServiceReport, run ObjectiveRuntimeSchedulerResumeDaemonRunReport, input ObjectiveRuntimeSchedulerResumeDaemonServiceInput) bool {
	switch run.Status {
	case "objective_runtime_scheduler_resume_daemon_runner_completed":
		report.CyclesCompleted++
		report.ConsecutiveIdleCycles = 0
		report.ConsecutiveBlockedCycles = 0
		report.ConsecutiveFailureCycles = 0
		report.ConsecutiveNoProgressCycles = 0
		report.Status = "objective_runtime_scheduler_resume_daemon_service_running"
		report.NextHostAction = firstNonEmptyString(run.NextHostAction, "continue_objective_runtime_scheduler_resume_daemon_service")
		return true
	case "objective_runtime_scheduler_resume_daemon_runner_idle":
		report.CyclesIdle++
		report.ConsecutiveIdleCycles++
		report.ConsecutiveBlockedCycles = 0
		report.ConsecutiveFailureCycles = 0
		report.ConsecutiveNoProgressCycles++
		report.Status = "objective_runtime_scheduler_resume_daemon_service_idle"
		report.NextHostAction = firstNonEmptyString(run.NextHostAction, "wait_for_scheduler_tick")
		return input.ContinueOnIdle
	case "objective_runtime_scheduler_resume_daemon_runner_worker_failure":
		report.CyclesFailed++
		report.ConsecutiveIdleCycles = 0
		report.ConsecutiveBlockedCycles = 0
		report.ConsecutiveFailureCycles++
		report.ConsecutiveNoProgressCycles++
		report.Status = "objective_runtime_scheduler_resume_daemon_service_worker_failure"
		report.NextHostAction = firstNonEmptyString(run.NextHostAction, "review_objective_runtime_scheduler_resume_worker_failure")
		return input.ContinueOnFailure
	case "objective_runtime_scheduler_resume_daemon_runner_context_done":
		report.CyclesContextDone++
		report.ContextDone = true
		report.Status = "objective_runtime_scheduler_resume_daemon_service_context_done"
		report.NextHostAction = firstNonEmptyString(run.NextHostAction, "review_objective_runtime_scheduler_resume_daemon_service_context")
		return false
	default:
		report.CyclesBlocked++
		report.ConsecutiveIdleCycles = 0
		report.ConsecutiveBlockedCycles++
		report.ConsecutiveFailureCycles = 0
		report.ConsecutiveNoProgressCycles++
		report.Status = "objective_runtime_scheduler_resume_daemon_service_blocked"
		report.NextHostAction = firstNonEmptyString(run.NextHostAction, "review_objective_runtime_scheduler_resume_daemon_service")
		return input.ContinueOnFailure
	}
}

func objectiveRuntimeSchedulerResumeDaemonServiceMax(value int) int {
	if value <= 0 {
		return 1
	}
	return value
}

func objectiveRuntimeSchedulerResumeDaemonServiceValidateWatchdog(input ObjectiveRuntimeSchedulerResumeDaemonServiceInput, report *ObjectiveRuntimeSchedulerResumeDaemonServiceReport) bool {
	if !input.InactivityWatchdogEnabled {
		return true
	}
	report.InactivityWatchdogEnabled = true
	report.Boundaries = appendUniqueResumeStrings(report.Boundaries,
		"objective_runtime_inactivity_watchdog_by_host",
		"watchdog_stop_requests_human_intervention",
	)
	if report.InactivityWatchdogRef == "" {
		objectiveRuntimeSchedulerResumeDaemonServiceMissingWatchdogRef(input.InactivityWatchdogRef, "host:objective_runtime_inactivity_watchdog_ref", "objective_runtime_inactivity_watchdog_ref_unsafe", report)
	}
	if report.InactivityWatchdogReviewRef == "" {
		objectiveRuntimeSchedulerResumeDaemonServiceMissingWatchdogRef(input.InactivityWatchdogReviewRef, "host:objective_runtime_inactivity_watchdog_review_ref", "objective_runtime_inactivity_watchdog_review_ref_unsafe", report)
	}
	if report.InactivityWatchdogHumanInterventionRef == "" {
		objectiveRuntimeSchedulerResumeDaemonServiceMissingWatchdogRef(input.InactivityWatchdogHumanInterventionRef, "host:objective_runtime_inactivity_watchdog_human_intervention_ref", "objective_runtime_inactivity_watchdog_human_intervention_ref_unsafe", report)
	}
	if input.MaxConsecutiveIdleCycles <= 0 &&
		input.MaxConsecutiveBlockedCycles <= 0 &&
		input.MaxConsecutiveFailureCycles <= 0 &&
		input.MaxConsecutiveNoProgressCycles <= 0 {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:objective_runtime_inactivity_watchdog_threshold")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "objective_runtime_inactivity_watchdog_threshold_missing")
	}
	if len(report.MissingInputs) > 0 || len(report.BlockedReasons) > 0 {
		report.NextHostAction = "provide_objective_runtime_inactivity_watchdog_config"
		return false
	}
	report.InactivityWatchdogReady = true
	return true
}

func objectiveRuntimeSchedulerResumeDaemonServiceMissingWatchdogRef(ref session.DisplaySafeRef, missingInput string, unsafeReason string, report *ObjectiveRuntimeSchedulerResumeDaemonServiceReport) {
	if strings.TrimSpace(string(ref)) == "" {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, missingInput)
		return
	}
	report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:display_safe_"+strings.TrimPrefix(missingInput, "host:"))
	report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, unsafeReason)
	report.Boundaries = appendUniqueResumeStrings(report.Boundaries, "unsafe_inactivity_watchdog_ref_rejected")
}

func objectiveRuntimeSchedulerResumeDaemonServiceWatchdogTriggered(report *ObjectiveRuntimeSchedulerResumeDaemonServiceReport, input ObjectiveRuntimeSchedulerResumeDaemonServiceInput) bool {
	if !input.InactivityWatchdogEnabled || !report.InactivityWatchdogReady {
		return false
	}
	switch {
	case input.MaxConsecutiveIdleCycles > 0 && report.ConsecutiveIdleCycles >= input.MaxConsecutiveIdleCycles:
		return objectiveRuntimeSchedulerResumeDaemonServiceTriggerWatchdog(report, "consecutive_idle_cycles_exceeded")
	case input.MaxConsecutiveBlockedCycles > 0 && report.ConsecutiveBlockedCycles >= input.MaxConsecutiveBlockedCycles:
		return objectiveRuntimeSchedulerResumeDaemonServiceTriggerWatchdog(report, "consecutive_blocked_cycles_exceeded")
	case input.MaxConsecutiveFailureCycles > 0 && report.ConsecutiveFailureCycles >= input.MaxConsecutiveFailureCycles:
		return objectiveRuntimeSchedulerResumeDaemonServiceTriggerWatchdog(report, "consecutive_failure_cycles_exceeded")
	case input.MaxConsecutiveNoProgressCycles > 0 && report.ConsecutiveNoProgressCycles >= input.MaxConsecutiveNoProgressCycles:
		return objectiveRuntimeSchedulerResumeDaemonServiceTriggerWatchdog(report, "consecutive_no_progress_cycles_exceeded")
	default:
		return false
	}
}

func objectiveRuntimeSchedulerResumeDaemonServiceTriggerWatchdog(report *ObjectiveRuntimeSchedulerResumeDaemonServiceReport, reason string) bool {
	report.InactivityWatchdogTriggered = true
	report.ReadyForWatchdogStop = true
	report.ReadyForHumanIntervention = true
	report.InactivityWatchdogReason = reason
	report.Status = "objective_runtime_scheduler_resume_daemon_service_watchdog_triggered"
	report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "objective_runtime_inactivity_watchdog_triggered", reason)
	report.Boundaries = appendUniqueResumeStrings(report.Boundaries,
		"objective_runtime_inactivity_watchdog_triggered",
		"watchdog_stop_by_host",
		"human_intervention_requested_by_host",
	)
	report.NextHostAction = "review_objective_runtime_inactivity_watchdog"
	return true
}
