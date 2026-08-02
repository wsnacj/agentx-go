package resume

import (
	"context"
	"encoding/json"
	"strings"

	scheduler "github.com/wsnacj/agentx-go/runtime/scheduler"
	session "github.com/wsnacj/agentx-go/runtime/session"
)

const ObjectiveRuntimeSchedulerResumeWorkerKind = "host_objective_runtime_scheduler_resume_worker"
const ObjectiveRuntimeSchedulerResumeWorkerPayloadVersion = "agentx.objective_runtime.scheduler_resume_tick.v1"

type ObjectiveRuntimeSchedulerResumeWorker struct {
	ContinuationReadback ObjectiveRuntimeSchedulerResumeContinuationReadbackFunc
	WakeDispatch         ObjectiveRuntimeSchedulerResumeWakeDispatchFunc
	Config               ObjectiveRuntimeSchedulerResumeWorkerConfig
}

type ObjectiveRuntimeSchedulerResumeWorkerConfig struct {
	ContinuationStoreRef session.DisplaySafeRef
	DispatchRef          session.DisplaySafeRef
	RuntimeDispatchRef   session.DisplaySafeRef
	HostRunnerRef        session.DisplaySafeRef
	HostRunnerVersionRef session.DisplaySafeRef
	OperatorApprovalRef  session.DisplaySafeRef
}

type ObjectiveRuntimeSchedulerResumeTickPayload struct {
	Version                      string                   `json:"version,omitempty"`
	TickRef                      session.DisplaySafeRef   `json:"tick_ref,omitempty"`
	SchedulerJobRef              session.DisplaySafeRef   `json:"scheduler_job_ref,omitempty"`
	SchedulerRuntimeQueueRef     session.DisplaySafeRef   `json:"scheduler_runtime_queue_ref,omitempty"`
	SchedulerWakeContinuationRef session.DisplaySafeRef   `json:"scheduler_wake_continuation_ref,omitempty"`
	ContinuationStoreRef         session.DisplaySafeRef   `json:"continuation_store_ref,omitempty"`
	ContinuationApplyRef         session.DisplaySafeRef   `json:"continuation_apply_ref,omitempty"`
	ContinuationReadbackRef      session.DisplaySafeRef   `json:"continuation_readback_ref,omitempty"`
	WakeCursorRef                session.DisplaySafeRef   `json:"wake_cursor_ref,omitempty"`
	ObjectiveRunRef              session.DisplaySafeRef   `json:"objective_run_ref,omitempty"`
	ObjectiveGraphSnapshotRef    session.DisplaySafeRef   `json:"objective_graph_snapshot_ref,omitempty"`
	ObjectiveGraphRef            session.DisplaySafeRef   `json:"objective_graph_ref,omitempty"`
	ObjectiveGraphReadbackRef    session.DisplaySafeRef   `json:"objective_graph_readback_ref,omitempty"`
	RuntimeDispatchRef           session.DisplaySafeRef   `json:"runtime_dispatch_ref,omitempty"`
	DispatchRef                  session.DisplaySafeRef   `json:"dispatch_ref,omitempty"`
	HostRunnerRef                session.DisplaySafeRef   `json:"host_runner_ref,omitempty"`
	HostRunnerVersionRef         session.DisplaySafeRef   `json:"host_runner_version_ref,omitempty"`
	OperatorApprovalRef          session.DisplaySafeRef   `json:"operator_approval_ref,omitempty"`
	SchedulerTickEvidenceRefs    []session.DisplaySafeRef `json:"scheduler_tick_evidence_refs,omitempty"`
	Boundaries                   []session.Boundary       `json:"boundaries,omitempty"`
}

type ObjectiveRuntimeSchedulerResumeContinuationReadbackInput struct {
	ContinuationStoreRef      session.DisplaySafeRef
	ContinuationApplyRef      session.DisplaySafeRef
	ContinuationReadbackRef   session.DisplaySafeRef
	ExpectedWakeCursorRef     session.DisplaySafeRef
	ExpectedObjectiveRunRef   session.DisplaySafeRef
	ExpectedGraphSnapshotRef  session.DisplaySafeRef
	ExpectedObjectiveGraphRef session.DisplaySafeRef
	ExpectedRuntimeQueueRef   session.DisplaySafeRef
	ExpectedContinuationRef   session.DisplaySafeRef
}

type ObjectiveRuntimeSchedulerResumeContinuationReadbackResult struct {
	ReadyForWakeContinuationResume    bool                     `json:"ready_for_wake_continuation_resume"`
	Status                            string                   `json:"status,omitempty"`
	BackendKind                       string                   `json:"backend_kind,omitempty"`
	ContinuationStoreConfigured       bool                     `json:"continuation_store_configured"`
	ContinuationStoreRef              session.DisplaySafeRef   `json:"continuation_store_ref,omitempty"`
	ContinuationApplyRef              session.DisplaySafeRef   `json:"continuation_apply_ref,omitempty"`
	ContinuationReadbackRef           session.DisplaySafeRef   `json:"continuation_readback_ref,omitempty"`
	WakeCursorRef                     session.DisplaySafeRef   `json:"wake_cursor_ref,omitempty"`
	ObjectiveRunRef                   session.DisplaySafeRef   `json:"objective_run_ref,omitempty"`
	ObjectiveGraphSnapshotRef         session.DisplaySafeRef   `json:"objective_graph_snapshot_ref,omitempty"`
	ObjectiveGraphRef                 session.DisplaySafeRef   `json:"objective_graph_ref,omitempty"`
	ObjectiveGraphValidationRef       session.DisplaySafeRef   `json:"objective_graph_validation_ref,omitempty"`
	ObjectiveGraphReadbackRef         session.DisplaySafeRef   `json:"objective_graph_readback_ref,omitempty"`
	ObjectiveGraphState               string                   `json:"objective_graph_state,omitempty"`
	ObjectiveGraphRevision            int                      `json:"objective_graph_revision,omitempty"`
	ReadyNodeRefs                     []session.DisplaySafeRef `json:"ready_node_refs,omitempty"`
	TaskLedgerRef                     session.DisplaySafeRef   `json:"task_ledger_ref,omitempty"`
	HostRuntimeQueueRef               session.DisplaySafeRef   `json:"host_runtime_queue_ref,omitempty"`
	SchedulerScheduleRef              session.DisplaySafeRef   `json:"scheduler_schedule_ref,omitempty"`
	SchedulerRegistrationRef          session.DisplaySafeRef   `json:"scheduler_registration_ref,omitempty"`
	SchedulerRegistrationReadbackRef  session.DisplaySafeRef   `json:"scheduler_registration_readback_ref,omitempty"`
	SchedulerRuntimeQueueRef          session.DisplaySafeRef   `json:"scheduler_runtime_queue_ref,omitempty"`
	SchedulerWakeContinuationRef      session.DisplaySafeRef   `json:"scheduler_wake_continuation_ref,omitempty"`
	WakeGateRef                       session.DisplaySafeRef   `json:"wake_gate_ref,omitempty"`
	WakeSignalRef                     session.DisplaySafeRef   `json:"wake_signal_ref,omitempty"`
	WakeDecision                      string                   `json:"wake_decision,omitempty"`
	ArtifactStoreReadbackRef          session.DisplaySafeRef   `json:"artifact_store_readback_ref,omitempty"`
	ContinuationStoreRootRef          session.DisplaySafeRef   `json:"continuation_store_root_ref,omitempty"`
	ContinuationStoreDriverRef        session.DisplaySafeRef   `json:"continuation_store_driver_ref,omitempty"`
	ContinuationStoreReadbackInstance session.DisplaySafeRef   `json:"continuation_store_readback_instance_ref,omitempty"`
	ContinuationStoreAuditRetention   session.DisplaySafeRef   `json:"continuation_store_audit_retention_ref,omitempty"`
	ContinuationStoreOperatorApproval session.DisplaySafeRef   `json:"continuation_store_operator_approval_ref,omitempty"`
	WakeContinuationEvidenceRefs      []session.DisplaySafeRef `json:"wake_continuation_evidence_refs,omitempty"`
	WakeCursorVisible                 bool                     `json:"wake_cursor_visible"`
	ObjectiveGraphSnapshotBound       bool                     `json:"objective_graph_snapshot_bound"`
	ManifestVisible                   bool                     `json:"manifest_visible"`
	DurableReadback                   bool                     `json:"durable_readback"`
	CrossInstanceReadback             bool                     `json:"cross_instance_readback"`
	LLMWakeDispatched                 bool                     `json:"llm_wake_dispatched"`
	RunnerDispatched                  bool                     `json:"runner_dispatched"`
	RuntimeAdapterExecuted            bool                     `json:"runtime_adapter_executed"`
	ToolExecuted                      bool                     `json:"tool_executed"`
	WorkflowDispatched                bool                     `json:"workflow_dispatched"`
	SchedulerApplied                  bool                     `json:"scheduler_applied"`
	InstallerExecuted                 bool                     `json:"installer_executed"`
	StoreMutationByCore               bool                     `json:"store_mutation_by_core"`
	MissingInputs                     []string                 `json:"missing_inputs,omitempty"`
	BlockedReasons                    []string                 `json:"blocked_reasons,omitempty"`
	Boundaries                        []session.Boundary       `json:"boundaries,omitempty"`
}

type ObjectiveRuntimeSchedulerResumeWakeDispatchInput struct {
	Continuation         ObjectiveRuntimeSchedulerResumeContinuationReadbackResult
	DispatchRef          session.DisplaySafeRef
	RuntimeDispatchRef   session.DisplaySafeRef
	HostRunnerRef        session.DisplaySafeRef
	HostRunnerVersionRef session.DisplaySafeRef
	OperatorApprovalRef  session.DisplaySafeRef
	Boundaries           []session.Boundary
}

type ObjectiveRuntimeSchedulerResumeWakeDispatchResult struct {
	Status                      string                   `json:"status,omitempty"`
	DispatcherKind              string                   `json:"dispatcher_kind,omitempty"`
	ReadyForRuntimeWakeDispatch bool                     `json:"ready_for_runtime_wake_dispatch"`
	DispatchRequestRecorded     bool                     `json:"dispatch_request_recorded"`
	HostRuntimeDispatchByHost   bool                     `json:"host_runtime_dispatch_by_host"`
	LLMWakeDispatched           bool                     `json:"llm_wake_dispatched"`
	RunnerDispatched            bool                     `json:"runner_dispatched"`
	RuntimeAdapterExecuted      bool                     `json:"runtime_adapter_executed"`
	ToolExecuted                bool                     `json:"tool_executed"`
	WorkflowDispatched          bool                     `json:"workflow_dispatched"`
	SchedulerApplied            bool                     `json:"scheduler_applied"`
	InstallerExecuted           bool                     `json:"installer_executed"`
	StoreMutationByCore         bool                     `json:"store_mutation_by_core"`
	DispatchRef                 session.DisplaySafeRef   `json:"dispatch_ref,omitempty"`
	RuntimeDispatchRef          session.DisplaySafeRef   `json:"runtime_dispatch_ref,omitempty"`
	HostRunnerRef               session.DisplaySafeRef   `json:"host_runner_ref,omitempty"`
	HostRunnerVersionRef        session.DisplaySafeRef   `json:"host_runner_version_ref,omitempty"`
	OperatorApprovalRef         session.DisplaySafeRef   `json:"operator_approval_ref,omitempty"`
	ObjectiveGraphSnapshotRef   session.DisplaySafeRef   `json:"objective_graph_snapshot_ref,omitempty"`
	ObjectiveGraphRef           session.DisplaySafeRef   `json:"objective_graph_ref,omitempty"`
	ObjectiveGraphReadbackRef   session.DisplaySafeRef   `json:"objective_graph_readback_ref,omitempty"`
	ObjectiveGraphState         string                   `json:"objective_graph_state,omitempty"`
	ObjectiveGraphRevision      int                      `json:"objective_graph_revision,omitempty"`
	ReadyNodeRefs               []session.DisplaySafeRef `json:"ready_node_refs,omitempty"`
	EvidenceRefs                []session.DisplaySafeRef `json:"evidence_refs,omitempty"`
	MissingInputs               []string                 `json:"missing_inputs,omitempty"`
	BlockedReasons              []string                 `json:"blocked_reasons,omitempty"`
	Boundaries                  []string                 `json:"boundaries,omitempty"`
	NextHostAction              string                   `json:"next_host_action,omitempty"`
}

type ObjectiveRuntimeSchedulerResumeContinuationReadbackFunc func(context.Context, ObjectiveRuntimeSchedulerResumeContinuationReadbackInput) (ObjectiveRuntimeSchedulerResumeContinuationReadbackResult, error)
type ObjectiveRuntimeSchedulerResumeWakeDispatchFunc func(context.Context, ObjectiveRuntimeSchedulerResumeWakeDispatchInput) (ObjectiveRuntimeSchedulerResumeWakeDispatchResult, error)

type ObjectiveRuntimeSchedulerResumeWorkerReport struct {
	Available                     bool                                                      `json:"available"`
	Status                        string                                                    `json:"status,omitempty"`
	WorkerKind                    string                                                    `json:"worker_kind,omitempty"`
	SchedulerTickObserved         bool                                                      `json:"scheduler_tick_observed"`
	PayloadAccepted               bool                                                      `json:"payload_accepted"`
	DisplaySafePayload            bool                                                      `json:"display_safe_payload"`
	WakeContinuationReadbackReady bool                                                      `json:"wake_continuation_readback_ready"`
	ReadyForObjectiveRunResume    bool                                                      `json:"ready_for_objective_run_resume"`
	ReadyForRuntimeWakeDispatch   bool                                                      `json:"ready_for_runtime_wake_dispatch"`
	DispatchRequestRecorded       bool                                                      `json:"dispatch_request_recorded"`
	HostRuntimeDispatchByHost     bool                                                      `json:"host_runtime_dispatch_by_host"`
	LLMWakeDispatched             bool                                                      `json:"llm_wake_dispatched"`
	RunnerDispatched              bool                                                      `json:"runner_dispatched"`
	RuntimeAdapterExecuted        bool                                                      `json:"runtime_adapter_executed"`
	ToolExecuted                  bool                                                      `json:"tool_executed"`
	WorkflowDispatched            bool                                                      `json:"workflow_dispatched"`
	SchedulerApplied              bool                                                      `json:"scheduler_applied"`
	InstallerExecuted             bool                                                      `json:"installer_executed"`
	StoreMutationByCore           bool                                                      `json:"store_mutation_by_core"`
	WorkerMutationByHost          bool                                                      `json:"worker_mutation_by_host"`
	JobID                         string                                                    `json:"job_id,omitempty"`
	JobKind                       string                                                    `json:"job_kind,omitempty"`
	TickRef                       string                                                    `json:"tick_ref,omitempty"`
	SchedulerRuntimeQueueRef      string                                                    `json:"scheduler_runtime_queue_ref,omitempty"`
	SchedulerWakeContinuationRef  string                                                    `json:"scheduler_wake_continuation_ref,omitempty"`
	ContinuationStoreRef          string                                                    `json:"continuation_store_ref,omitempty"`
	ContinuationApplyRef          string                                                    `json:"continuation_apply_ref,omitempty"`
	ContinuationReadbackRef       string                                                    `json:"continuation_readback_ref,omitempty"`
	WakeCursorRef                 string                                                    `json:"wake_cursor_ref,omitempty"`
	ObjectiveRunRef               string                                                    `json:"objective_run_ref,omitempty"`
	ObjectiveGraphSnapshotRef     string                                                    `json:"objective_graph_snapshot_ref,omitempty"`
	ObjectiveGraphRef             string                                                    `json:"objective_graph_ref,omitempty"`
	ObjectiveGraphReadbackRef     string                                                    `json:"objective_graph_readback_ref,omitempty"`
	ObjectiveGraphState           string                                                    `json:"objective_graph_state,omitempty"`
	ObjectiveGraphRevision        int                                                       `json:"objective_graph_revision,omitempty"`
	ReadyNodeRefs                 []string                                                  `json:"ready_node_refs,omitempty"`
	DispatchRef                   string                                                    `json:"dispatch_ref,omitempty"`
	RuntimeDispatchRef            string                                                    `json:"runtime_dispatch_ref,omitempty"`
	HostRunnerRef                 string                                                    `json:"host_runner_ref,omitempty"`
	HostRunnerVersionRef          string                                                    `json:"host_runner_version_ref,omitempty"`
	OperatorApprovalRef           string                                                    `json:"operator_approval_ref,omitempty"`
	EvidenceRefs                  []string                                                  `json:"evidence_refs,omitempty"`
	MissingInputs                 []string                                                  `json:"missing_inputs,omitempty"`
	BlockedReasons                []string                                                  `json:"blocked_reasons,omitempty"`
	Boundaries                    []string                                                  `json:"boundaries,omitempty"`
	NextHostAction                string                                                    `json:"next_host_action,omitempty"`
	Payload                       ObjectiveRuntimeSchedulerResumeTickPayload                `json:"payload,omitempty"`
	WakeContinuationReadback      ObjectiveRuntimeSchedulerResumeContinuationReadbackResult `json:"wake_continuation_readback,omitempty"`
	WakeDispatch                  ObjectiveRuntimeSchedulerResumeWakeDispatchResult         `json:"wake_dispatch,omitempty"`
}

func BuildObjectiveRuntimeSchedulerResumeTickPayload(input ObjectiveRuntimeSchedulerResumeTickPayload) ObjectiveRuntimeSchedulerResumeTickPayload {
	out := input.Normalize()
	if out.Version == "" {
		out.Version = ObjectiveRuntimeSchedulerResumeWorkerPayloadVersion
	}
	if out.SchedulerJobRef == "" && out.TickRef != "" {
		out.SchedulerJobRef = out.TickRef
	}
	out.SchedulerTickEvidenceRefs = append(out.SchedulerTickEvidenceRefs, out.TickRef, out.SchedulerJobRef, out.SchedulerRuntimeQueueRef, out.SchedulerWakeContinuationRef)
	out.SchedulerTickEvidenceRefs = objectiveRuntimeSchedulerResumeSafeRefs(out.SchedulerTickEvidenceRefs)
	out.Boundaries = session.MergeBoundaries(out.Boundaries, []session.Boundary{
		"host_owned_scheduler_tick_payload",
		"display_safe_refs_only",
	})
	return out.Normalize()
}

func (p ObjectiveRuntimeSchedulerResumeTickPayload) JSON() (string, error) {
	blob, err := json.Marshal(p.Normalize())
	if err != nil {
		return "", err
	}
	return string(blob), nil
}

func (p ObjectiveRuntimeSchedulerResumeTickPayload) Normalize() ObjectiveRuntimeSchedulerResumeTickPayload {
	out := p
	out.Version = strings.TrimSpace(out.Version)
	out.TickRef = objectiveRuntimeSchedulerResumeSafeRef(out.TickRef)
	out.SchedulerJobRef = objectiveRuntimeSchedulerResumeSafeRef(out.SchedulerJobRef)
	out.SchedulerRuntimeQueueRef = objectiveRuntimeSchedulerResumeSafeRef(out.SchedulerRuntimeQueueRef)
	out.SchedulerWakeContinuationRef = objectiveRuntimeSchedulerResumeSafeRef(out.SchedulerWakeContinuationRef)
	out.ContinuationStoreRef = objectiveRuntimeSchedulerResumeSafeRef(out.ContinuationStoreRef)
	out.ContinuationApplyRef = objectiveRuntimeSchedulerResumeSafeRef(out.ContinuationApplyRef)
	out.ContinuationReadbackRef = objectiveRuntimeSchedulerResumeSafeRef(out.ContinuationReadbackRef)
	out.WakeCursorRef = objectiveRuntimeSchedulerResumeSafeRef(out.WakeCursorRef)
	out.ObjectiveRunRef = objectiveRuntimeSchedulerResumeSafeRef(out.ObjectiveRunRef)
	out.ObjectiveGraphSnapshotRef = objectiveRuntimeSchedulerResumeSafeRef(out.ObjectiveGraphSnapshotRef)
	out.ObjectiveGraphRef = objectiveRuntimeSchedulerResumeSafeRef(out.ObjectiveGraphRef)
	out.ObjectiveGraphReadbackRef = objectiveRuntimeSchedulerResumeSafeRef(out.ObjectiveGraphReadbackRef)
	out.RuntimeDispatchRef = objectiveRuntimeSchedulerResumeSafeRef(out.RuntimeDispatchRef)
	out.DispatchRef = objectiveRuntimeSchedulerResumeSafeRef(out.DispatchRef)
	out.HostRunnerRef = objectiveRuntimeSchedulerResumeSafeRef(out.HostRunnerRef)
	out.HostRunnerVersionRef = objectiveRuntimeSchedulerResumeSafeRef(out.HostRunnerVersionRef)
	out.OperatorApprovalRef = objectiveRuntimeSchedulerResumeSafeRef(out.OperatorApprovalRef)
	out.SchedulerTickEvidenceRefs = objectiveRuntimeSchedulerResumeSafeRefs(out.SchedulerTickEvidenceRefs)
	out.Boundaries = session.MergeBoundaries(out.Boundaries)
	return out
}

func (w ObjectiveRuntimeSchedulerResumeWorker) HandleSchedulerTick(ctx context.Context, job scheduler.Job) (ObjectiveRuntimeSchedulerResumeWorkerReport, error) {
	report := ObjectiveRuntimeSchedulerResumeWorkerReport{
		Available:             w.ContinuationReadback != nil && w.WakeDispatch != nil,
		Status:                "blocked",
		WorkerKind:            ObjectiveRuntimeSchedulerResumeWorkerKind,
		SchedulerTickObserved: true,
		JobID:                 strings.TrimSpace(job.ID),
		JobKind:               strings.TrimSpace(job.JobKind),
		Boundaries: []string{
			"host_owned_objective_runtime_scheduler_resume_worker",
			"production_scheduler_tick_worker",
			"durable_objective_run_resume_worker",
			"display_safe_refs_only",
			"no_llm_wake_by_core",
			"no_runner_dispatch_by_core",
			"no_runtime_adapter_execution_by_core",
			"no_tool_execution",
			"no_workflow_dispatch_by_core",
			"no_scheduler_apply_by_core",
			"no_install_apply",
			"no_store_mutation_by_core",
		},
		NextHostAction: "review_objective_runtime_scheduler_resume_worker",
	}
	payload, err := decodeObjectiveRuntimeSchedulerResumePayload(job.Payload)
	if err != nil {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:scheduler_tick_payload")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "scheduler_tick_payload_invalid")
		return report.Normalize(), nil
	}
	report.Payload = payload
	report.PayloadAccepted = true
	report.DisplaySafePayload = objectiveRuntimeSchedulerResumePayloadDisplaySafe(payload)
	payload = payload.Normalize()
	report.Payload = payload
	report = objectiveRuntimeSchedulerResumeBindPayload(report, payload)
	if !report.DisplaySafePayload {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:display_safe_scheduler_tick_payload")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "scheduler_tick_payload_unsafe")
		report.Boundaries = appendUniqueResumeStrings(report.Boundaries, "raw_payload_not_allowed")
		return report.Normalize(), nil
	}
	if w.ContinuationReadback == nil {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:objective_runtime_wake_continuation_store")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "objective_runtime_wake_continuation_store_missing")
		return report.Normalize(), nil
	}
	if w.WakeDispatch == nil {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:objective_runtime_wake_dispatcher")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "objective_runtime_wake_dispatcher_missing")
		return report.Normalize(), nil
	}
	readbackInput := ObjectiveRuntimeSchedulerResumeContinuationReadbackInput{
		ContinuationStoreRef:      firstObjectiveRuntimeSchedulerResumeRef(payload.ContinuationStoreRef, w.Config.ContinuationStoreRef),
		ContinuationApplyRef:      payload.ContinuationApplyRef,
		ContinuationReadbackRef:   payload.ContinuationReadbackRef,
		ExpectedWakeCursorRef:     payload.WakeCursorRef,
		ExpectedObjectiveRunRef:   payload.ObjectiveRunRef,
		ExpectedGraphSnapshotRef:  payload.ObjectiveGraphSnapshotRef,
		ExpectedObjectiveGraphRef: payload.ObjectiveGraphRef,
		ExpectedRuntimeQueueRef:   payload.SchedulerRuntimeQueueRef,
		ExpectedContinuationRef:   payload.SchedulerWakeContinuationRef,
	}
	if missing := objectiveRuntimeSchedulerResumeReadbackMissingInputs(readbackInput); len(missing) > 0 {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, missing...)
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "scheduler_tick_payload_readback_refs_missing")
		report.NextHostAction = "provide_objective_runtime_wake_cursor_readback_refs"
		return report.Normalize(), nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	readback, err := w.ContinuationReadback(ctx, readbackInput)
	if err != nil {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:objective_runtime_wake_continuation_readback")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "objective_runtime_wake_continuation_readback_failed")
		return report.Normalize(), nil
	}
	readback = readback.Normalize()
	report.WakeContinuationReadback = readback
	report.WakeContinuationReadbackReady = readback.ReadyForWakeContinuationResume
	report.ReadyForObjectiveRunResume = readback.ReadyForWakeContinuationResume
	report.ObjectiveGraphSnapshotRef = string(readback.ObjectiveGraphSnapshotRef)
	report.ObjectiveGraphRef = string(readback.ObjectiveGraphRef)
	report.ObjectiveGraphReadbackRef = string(readback.ObjectiveGraphReadbackRef)
	report.ObjectiveGraphState = strings.TrimSpace(readback.ObjectiveGraphState)
	report.ObjectiveGraphRevision = readback.ObjectiveGraphRevision
	report.ReadyNodeRefs = displaySafeRefsToStrings(readback.ReadyNodeRefs)
	report.EvidenceRefs = appendUniqueResumeStrings(report.EvidenceRefs, displaySafeRefsToStrings(readback.WakeContinuationEvidenceRefs)...)
	report.Boundaries = appendUniqueResumeStrings(report.Boundaries, resumeBoundariesToStrings(readback.Boundaries)...)
	if !readback.ReadyForWakeContinuationResume {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:objective_runtime_wake_continuation_readback")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "objective_runtime_wake_continuation_readback_not_ready")
		report.NextHostAction = "review_objective_runtime_wake_continuation_readback"
		return report.Normalize(), nil
	}
	dispatchBoundaries := session.MergeBoundaries(payload.Boundaries, []session.Boundary{
		"host_owned_scheduler_tick_to_wake_dispatch",
		"durable_objective_run_resume_cursor_consumed",
	})
	report.Boundaries = appendUniqueResumeStrings(report.Boundaries, resumeBoundariesToStrings(dispatchBoundaries)...)
	dispatch, err := w.WakeDispatch(ctx, ObjectiveRuntimeSchedulerResumeWakeDispatchInput{
		Continuation:         readback,
		DispatchRef:          firstObjectiveRuntimeSchedulerResumeRef(payload.DispatchRef, w.Config.DispatchRef),
		RuntimeDispatchRef:   firstObjectiveRuntimeSchedulerResumeRef(payload.RuntimeDispatchRef, w.Config.RuntimeDispatchRef),
		HostRunnerRef:        firstObjectiveRuntimeSchedulerResumeRef(payload.HostRunnerRef, w.Config.HostRunnerRef),
		HostRunnerVersionRef: firstObjectiveRuntimeSchedulerResumeRef(payload.HostRunnerVersionRef, w.Config.HostRunnerVersionRef),
		OperatorApprovalRef:  firstObjectiveRuntimeSchedulerResumeRef(payload.OperatorApprovalRef, w.Config.OperatorApprovalRef),
		Boundaries:           dispatchBoundaries,
	})
	if err != nil {
		report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, "host:objective_runtime_wake_dispatcher")
		report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, "objective_runtime_wake_dispatcher_failed")
		return report.Normalize(), nil
	}
	report.WakeDispatch = dispatch
	report.ReadyForRuntimeWakeDispatch = dispatch.ReadyForRuntimeWakeDispatch
	report.DispatchRequestRecorded = dispatch.DispatchRequestRecorded
	report.HostRuntimeDispatchByHost = dispatch.HostRuntimeDispatchByHost
	report.LLMWakeDispatched = dispatch.LLMWakeDispatched
	report.RunnerDispatched = dispatch.RunnerDispatched
	report.RuntimeAdapterExecuted = dispatch.RuntimeAdapterExecuted
	report.ToolExecuted = dispatch.ToolExecuted
	report.WorkflowDispatched = dispatch.WorkflowDispatched
	report.SchedulerApplied = dispatch.SchedulerApplied
	report.InstallerExecuted = dispatch.InstallerExecuted
	report.StoreMutationByCore = dispatch.StoreMutationByCore
	report.DispatchRef = firstNonEmptyString(string(dispatch.DispatchRef), report.DispatchRef)
	report.RuntimeDispatchRef = firstNonEmptyString(string(dispatch.RuntimeDispatchRef), report.RuntimeDispatchRef)
	report.HostRunnerRef = firstNonEmptyString(string(dispatch.HostRunnerRef), report.HostRunnerRef)
	report.HostRunnerVersionRef = firstNonEmptyString(string(dispatch.HostRunnerVersionRef), report.HostRunnerVersionRef)
	report.OperatorApprovalRef = firstNonEmptyString(string(dispatch.OperatorApprovalRef), report.OperatorApprovalRef)
	report.ObjectiveGraphSnapshotRef = firstNonEmptyString(string(dispatch.ObjectiveGraphSnapshotRef), report.ObjectiveGraphSnapshotRef)
	report.ObjectiveGraphRef = firstNonEmptyString(string(dispatch.ObjectiveGraphRef), report.ObjectiveGraphRef)
	report.ObjectiveGraphReadbackRef = firstNonEmptyString(string(dispatch.ObjectiveGraphReadbackRef), report.ObjectiveGraphReadbackRef)
	report.ObjectiveGraphState = firstNonEmptyString(strings.TrimSpace(dispatch.ObjectiveGraphState), report.ObjectiveGraphState)
	if dispatch.ObjectiveGraphRevision > 0 {
		report.ObjectiveGraphRevision = dispatch.ObjectiveGraphRevision
	}
	report.ReadyNodeRefs = appendUniqueResumeStrings(report.ReadyNodeRefs, displaySafeRefsToStrings(dispatch.ReadyNodeRefs)...)
	report.EvidenceRefs = appendUniqueResumeStrings(report.EvidenceRefs, displaySafeRefsToStrings(dispatch.EvidenceRefs)...)
	report.MissingInputs = appendUniqueResumeStrings(report.MissingInputs, dispatch.MissingInputs...)
	report.BlockedReasons = appendUniqueResumeStrings(report.BlockedReasons, dispatch.BlockedReasons...)
	report.Boundaries = appendUniqueResumeStrings(report.Boundaries, dispatch.Boundaries...)
	report.NextHostAction = firstNonEmptyString(dispatch.NextHostAction, report.NextHostAction)
	if dispatch.ReadyForRuntimeWakeDispatch {
		report.Status = "objective_runtime_scheduler_resume_dispatch_ready"
		report.WorkerMutationByHost = true
		report.Boundaries = appendUniqueResumeStrings(report.Boundaries,
			"scheduler_tick_resumed_objective_run",
			"scheduler_tick_bound_to_runtime_wake_dispatch",
		)
	}
	return report.Normalize(), nil
}

func (w ObjectiveRuntimeSchedulerResumeWorker) Handler(sink func(ObjectiveRuntimeSchedulerResumeWorkerReport)) scheduler.Handler {
	return func(ctx context.Context, job scheduler.Job) error {
		report, err := w.HandleSchedulerTick(ctx, job)
		if sink != nil {
			sink(report)
		}
		return err
	}
}

func (report ObjectiveRuntimeSchedulerResumeContinuationReadbackResult) Normalize() ObjectiveRuntimeSchedulerResumeContinuationReadbackResult {
	out := report
	out.Status = strings.TrimSpace(out.Status)
	if out.Status == "" {
		out.Status = "blocked"
	}
	out.BackendKind = strings.TrimSpace(out.BackendKind)
	out.ContinuationStoreRef = objectiveRuntimeSchedulerResumeSafeRef(out.ContinuationStoreRef)
	out.ContinuationApplyRef = objectiveRuntimeSchedulerResumeSafeRef(out.ContinuationApplyRef)
	out.ContinuationReadbackRef = objectiveRuntimeSchedulerResumeSafeRef(out.ContinuationReadbackRef)
	out.WakeCursorRef = objectiveRuntimeSchedulerResumeSafeRef(out.WakeCursorRef)
	out.ObjectiveRunRef = objectiveRuntimeSchedulerResumeSafeRef(out.ObjectiveRunRef)
	out.ObjectiveGraphSnapshotRef = objectiveRuntimeSchedulerResumeSafeRef(out.ObjectiveGraphSnapshotRef)
	out.ObjectiveGraphRef = objectiveRuntimeSchedulerResumeSafeRef(out.ObjectiveGraphRef)
	out.ObjectiveGraphValidationRef = objectiveRuntimeSchedulerResumeSafeRef(out.ObjectiveGraphValidationRef)
	out.ObjectiveGraphReadbackRef = objectiveRuntimeSchedulerResumeSafeRef(out.ObjectiveGraphReadbackRef)
	out.ObjectiveGraphState = strings.TrimSpace(out.ObjectiveGraphState)
	if out.ObjectiveGraphRevision < 0 {
		out.ObjectiveGraphRevision = 0
	}
	out.ReadyNodeRefs = objectiveRuntimeSchedulerResumeSafeRefs(out.ReadyNodeRefs)
	out.TaskLedgerRef = objectiveRuntimeSchedulerResumeSafeRef(out.TaskLedgerRef)
	out.HostRuntimeQueueRef = objectiveRuntimeSchedulerResumeSafeRef(out.HostRuntimeQueueRef)
	out.SchedulerScheduleRef = objectiveRuntimeSchedulerResumeSafeRef(out.SchedulerScheduleRef)
	out.SchedulerRegistrationRef = objectiveRuntimeSchedulerResumeSafeRef(out.SchedulerRegistrationRef)
	out.SchedulerRegistrationReadbackRef = objectiveRuntimeSchedulerResumeSafeRef(out.SchedulerRegistrationReadbackRef)
	out.SchedulerRuntimeQueueRef = objectiveRuntimeSchedulerResumeSafeRef(out.SchedulerRuntimeQueueRef)
	out.SchedulerWakeContinuationRef = objectiveRuntimeSchedulerResumeSafeRef(out.SchedulerWakeContinuationRef)
	out.WakeGateRef = objectiveRuntimeSchedulerResumeSafeRef(out.WakeGateRef)
	out.WakeSignalRef = objectiveRuntimeSchedulerResumeSafeRef(out.WakeSignalRef)
	out.WakeDecision = strings.TrimSpace(out.WakeDecision)
	out.ArtifactStoreReadbackRef = objectiveRuntimeSchedulerResumeSafeRef(out.ArtifactStoreReadbackRef)
	out.ContinuationStoreRootRef = objectiveRuntimeSchedulerResumeSafeRef(out.ContinuationStoreRootRef)
	out.ContinuationStoreDriverRef = objectiveRuntimeSchedulerResumeSafeRef(out.ContinuationStoreDriverRef)
	out.ContinuationStoreReadbackInstance = objectiveRuntimeSchedulerResumeSafeRef(out.ContinuationStoreReadbackInstance)
	out.ContinuationStoreAuditRetention = objectiveRuntimeSchedulerResumeSafeRef(out.ContinuationStoreAuditRetention)
	out.ContinuationStoreOperatorApproval = objectiveRuntimeSchedulerResumeSafeRef(out.ContinuationStoreOperatorApproval)
	out.WakeContinuationEvidenceRefs = objectiveRuntimeSchedulerResumeSafeRefs(out.WakeContinuationEvidenceRefs)
	out.ObjectiveGraphSnapshotBound = out.ObjectiveGraphSnapshotRef != "" && out.ObjectiveGraphRef != ""
	out.MissingInputs = appendUniqueResumeStrings(nil, out.MissingInputs...)
	out.BlockedReasons = appendUniqueResumeStrings(nil, out.BlockedReasons...)
	out.Boundaries = session.MergeBoundaries(out.Boundaries)
	out.LLMWakeDispatched = false
	out.RunnerDispatched = false
	out.RuntimeAdapterExecuted = false
	out.ToolExecuted = false
	out.WorkflowDispatched = false
	out.SchedulerApplied = false
	out.InstallerExecuted = false
	out.StoreMutationByCore = false
	out.ReadyForWakeContinuationResume = out.ReadyForWakeContinuationResume &&
		out.WakeCursorVisible &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0
	if out.Status == "wake_continuation_readback_recorded" && !out.ReadyForWakeContinuationResume {
		out.Status = "blocked"
	}
	return out
}

func (report ObjectiveRuntimeSchedulerResumeWorkerReport) Normalize() ObjectiveRuntimeSchedulerResumeWorkerReport {
	out := report
	out.Status = strings.TrimSpace(out.Status)
	if out.Status == "" {
		out.Status = "blocked"
	}
	out.WorkerKind = strings.TrimSpace(out.WorkerKind)
	if out.WorkerKind == "" {
		out.WorkerKind = ObjectiveRuntimeSchedulerResumeWorkerKind
	}
	out.JobID = strings.TrimSpace(out.JobID)
	out.JobKind = strings.TrimSpace(out.JobKind)
	out.TickRef = strings.TrimSpace(out.TickRef)
	out.SchedulerRuntimeQueueRef = strings.TrimSpace(out.SchedulerRuntimeQueueRef)
	out.SchedulerWakeContinuationRef = strings.TrimSpace(out.SchedulerWakeContinuationRef)
	out.ContinuationStoreRef = strings.TrimSpace(out.ContinuationStoreRef)
	out.ContinuationApplyRef = strings.TrimSpace(out.ContinuationApplyRef)
	out.ContinuationReadbackRef = strings.TrimSpace(out.ContinuationReadbackRef)
	out.WakeCursorRef = strings.TrimSpace(out.WakeCursorRef)
	out.ObjectiveRunRef = strings.TrimSpace(out.ObjectiveRunRef)
	out.ObjectiveGraphSnapshotRef = strings.TrimSpace(out.ObjectiveGraphSnapshotRef)
	out.ObjectiveGraphRef = strings.TrimSpace(out.ObjectiveGraphRef)
	out.ObjectiveGraphReadbackRef = strings.TrimSpace(out.ObjectiveGraphReadbackRef)
	out.ObjectiveGraphState = strings.TrimSpace(out.ObjectiveGraphState)
	if out.ObjectiveGraphRevision < 0 {
		out.ObjectiveGraphRevision = 0
	}
	out.ReadyNodeRefs = appendUniqueResumeStrings(nil, out.ReadyNodeRefs...)
	out.DispatchRef = strings.TrimSpace(out.DispatchRef)
	out.RuntimeDispatchRef = strings.TrimSpace(out.RuntimeDispatchRef)
	out.HostRunnerRef = strings.TrimSpace(out.HostRunnerRef)
	out.HostRunnerVersionRef = strings.TrimSpace(out.HostRunnerVersionRef)
	out.OperatorApprovalRef = strings.TrimSpace(out.OperatorApprovalRef)
	out.EvidenceRefs = appendUniqueResumeStrings(nil, out.EvidenceRefs...)
	out.MissingInputs = appendUniqueResumeStrings(nil, out.MissingInputs...)
	out.BlockedReasons = appendUniqueResumeStrings(nil, out.BlockedReasons...)
	out.Boundaries = appendUniqueResumeStrings(nil, out.Boundaries...)
	out.NextHostAction = strings.TrimSpace(out.NextHostAction)
	if out.NextHostAction == "" {
		out.NextHostAction = "review_objective_runtime_scheduler_resume_worker"
	}
	out.LLMWakeDispatched = out.LLMWakeDispatched && out.RunnerDispatched && out.HostRuntimeDispatchByHost
	out.RunnerDispatched = false
	out.RuntimeAdapterExecuted = false
	out.ToolExecuted = false
	out.WorkflowDispatched = false
	out.SchedulerApplied = false
	out.InstallerExecuted = false
	out.StoreMutationByCore = false
	out.WorkerMutationByHost = out.WorkerMutationByHost && out.DispatchRequestRecorded
	out.ReadyForObjectiveRunResume = out.ReadyForObjectiveRunResume && out.WakeContinuationReadbackReady && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0
	out.ReadyForRuntimeWakeDispatch = out.ReadyForRuntimeWakeDispatch && out.ReadyForObjectiveRunResume && out.DispatchRequestRecorded && len(out.MissingInputs) == 0 && len(out.BlockedReasons) == 0
	if out.Status == "objective_runtime_scheduler_resume_dispatch_ready" && !out.ReadyForRuntimeWakeDispatch {
		out.Status = "blocked"
	}
	return out
}

func decodeObjectiveRuntimeSchedulerResumePayload(raw string) (ObjectiveRuntimeSchedulerResumeTickPayload, error) {
	var payload ObjectiveRuntimeSchedulerResumeTickPayload
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &payload); err != nil {
		return ObjectiveRuntimeSchedulerResumeTickPayload{}, err
	}
	return payload, nil
}

func objectiveRuntimeSchedulerResumeBindPayload(report ObjectiveRuntimeSchedulerResumeWorkerReport, payload ObjectiveRuntimeSchedulerResumeTickPayload) ObjectiveRuntimeSchedulerResumeWorkerReport {
	report.TickRef = string(payload.TickRef)
	report.SchedulerRuntimeQueueRef = string(payload.SchedulerRuntimeQueueRef)
	report.SchedulerWakeContinuationRef = string(payload.SchedulerWakeContinuationRef)
	report.ContinuationStoreRef = string(payload.ContinuationStoreRef)
	report.ContinuationApplyRef = string(payload.ContinuationApplyRef)
	report.ContinuationReadbackRef = string(payload.ContinuationReadbackRef)
	report.WakeCursorRef = string(payload.WakeCursorRef)
	report.ObjectiveRunRef = string(payload.ObjectiveRunRef)
	report.ObjectiveGraphSnapshotRef = string(payload.ObjectiveGraphSnapshotRef)
	report.ObjectiveGraphRef = string(payload.ObjectiveGraphRef)
	report.ObjectiveGraphReadbackRef = string(payload.ObjectiveGraphReadbackRef)
	report.DispatchRef = string(payload.DispatchRef)
	report.RuntimeDispatchRef = string(payload.RuntimeDispatchRef)
	report.HostRunnerRef = string(payload.HostRunnerRef)
	report.HostRunnerVersionRef = string(payload.HostRunnerVersionRef)
	report.OperatorApprovalRef = string(payload.OperatorApprovalRef)
	report.EvidenceRefs = appendUniqueResumeStrings(report.EvidenceRefs, displaySafeRefsToStrings(payload.SchedulerTickEvidenceRefs)...)
	report.Boundaries = appendUniqueResumeStrings(report.Boundaries, resumeBoundariesToStrings(payload.Boundaries)...)
	return report
}

func objectiveRuntimeSchedulerResumeReadbackMissingInputs(input ObjectiveRuntimeSchedulerResumeContinuationReadbackInput) []string {
	var missing []string
	for _, check := range []struct {
		ref   session.DisplaySafeRef
		input string
	}{
		{input.ContinuationStoreRef, "host:objective_runtime_wake_continuation_store_ref"},
		{input.ContinuationApplyRef, "host:objective_runtime_wake_continuation_apply_ref"},
		{input.ContinuationReadbackRef, "host:objective_runtime_wake_continuation_readback_ref"},
		{input.ExpectedWakeCursorRef, "host:objective_runtime_wake_cursor_ref"},
		{input.ExpectedObjectiveRunRef, "host:objective_run_ref"},
		{input.ExpectedRuntimeQueueRef, "host:scheduler_runtime_queue_ref"},
		{input.ExpectedContinuationRef, "host:scheduler_wake_continuation_ref"},
	} {
		if objectiveRuntimeSchedulerResumeSafeRef(check.ref) == "" {
			missing = appendUniqueResumeStrings(missing, check.input)
		}
	}
	return missing
}

func objectiveRuntimeSchedulerResumePayloadDisplaySafe(payload ObjectiveRuntimeSchedulerResumeTickPayload) bool {
	for _, ref := range []session.DisplaySafeRef{
		payload.TickRef,
		payload.SchedulerJobRef,
		payload.SchedulerRuntimeQueueRef,
		payload.SchedulerWakeContinuationRef,
		payload.ContinuationStoreRef,
		payload.ContinuationApplyRef,
		payload.ContinuationReadbackRef,
		payload.WakeCursorRef,
		payload.ObjectiveRunRef,
		payload.ObjectiveGraphSnapshotRef,
		payload.ObjectiveGraphRef,
		payload.ObjectiveGraphReadbackRef,
		payload.RuntimeDispatchRef,
		payload.DispatchRef,
		payload.HostRunnerRef,
		payload.HostRunnerVersionRef,
		payload.OperatorApprovalRef,
	} {
		if strings.TrimSpace(string(ref)) != "" && objectiveRuntimeSchedulerResumeSafeRef(ref) == "" {
			return false
		}
	}
	for _, ref := range payload.SchedulerTickEvidenceRefs {
		if strings.TrimSpace(string(ref)) != "" && objectiveRuntimeSchedulerResumeSafeRef(ref) == "" {
			return false
		}
	}
	return true
}

func objectiveRuntimeSchedulerResumeSafeRef(ref session.DisplaySafeRef) session.DisplaySafeRef {
	raw := strings.TrimSpace(string(ref))
	if raw == "" {
		return ""
	}
	normalized, ok := session.NormalizeDisplaySafeRef(raw)
	if !ok {
		return ""
	}
	return normalized
}

func objectiveRuntimeSchedulerResumeSafeRefs(refs []session.DisplaySafeRef) []session.DisplaySafeRef {
	out := make([]session.DisplaySafeRef, 0, len(refs))
	seen := map[session.DisplaySafeRef]bool{}
	for _, ref := range refs {
		safe := objectiveRuntimeSchedulerResumeSafeRef(ref)
		if safe == "" || seen[safe] {
			continue
		}
		seen[safe] = true
		out = append(out, safe)
	}
	return out
}

func firstObjectiveRuntimeSchedulerResumeRef(values ...session.DisplaySafeRef) session.DisplaySafeRef {
	for _, value := range values {
		if ref := objectiveRuntimeSchedulerResumeSafeRef(value); ref != "" {
			return ref
		}
	}
	return ""
}

func displaySafeRefsToStrings(refs []session.DisplaySafeRef) []string {
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		if safe := objectiveRuntimeSchedulerResumeSafeRef(ref); safe != "" {
			out = append(out, string(safe))
		}
	}
	return out
}

func resumeBoundariesToStrings(boundaries []session.Boundary) []string {
	out := make([]string, 0, len(boundaries))
	for _, boundary := range session.MergeBoundaries(boundaries) {
		if value := strings.TrimSpace(string(boundary)); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func appendUniqueResumeStrings(items []string, values ...string) []string {
	seen := make(map[string]bool, len(items)+len(values))
	out := make([]string, 0, len(items)+len(values))
	for _, value := range append(items, values...) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
