package hostkit

import (
	"context"
	"strings"
	"sync"
	"time"

	agentxcontrolplane "github.com/wsnacj/agentx-go/runtime/session"
)

const BackendKind = "host_delegation_worker_runtime"

type WorkerRuntime interface {
	InvokeDelegationWorker(ctx context.Context, request WorkerRequest) (WorkerResult, error)
	ReadDelegationWorkerResult(ctx context.Context, request WorkerReadbackRequest) (WorkerReadback, error)
}

type StateStore interface {
	WriteWorkerRun(ctx context.Context, record WorkerRunRecord) error
	ReadWorkerRun(ctx context.Context, workerRunRef agentxcontrolplane.DisplaySafeRef) (WorkerRunRecord, bool, error)
}

type Backend struct {
	Runtime     WorkerRuntime
	Store       StateStore
	BackendRef  agentxcontrolplane.DisplaySafeRef
	BackendKind string
	Durable     bool
	Now         func() time.Time
}

type BackendInput struct {
	Enabled                 bool
	Readiness               agentxcontrolplane.HostOwnedDelegationWorkerRuntimeReadiness
	InvocationReportRef     agentxcontrolplane.DisplaySafeRef
	ObservedInvocationRef   agentxcontrolplane.DisplaySafeRef
	HostWorkerRuntimeRunRef agentxcontrolplane.DisplaySafeRef
	VisibleToolRefs         []agentxcontrolplane.DisplaySafeRef
	ContextRefs             []agentxcontrolplane.DisplaySafeRef
	BudgetRef               agentxcontrolplane.DisplaySafeRef
	TimeoutRef              agentxcontrolplane.DisplaySafeRef
	ParallelismRef          agentxcontrolplane.DisplaySafeRef
	FailureRef              agentxcontrolplane.DisplaySafeRef
	CompensationRef         agentxcontrolplane.DisplaySafeRef
	EvidenceRefs            []agentxcontrolplane.EvidenceRef
	Boundaries              []agentxcontrolplane.Boundary
	RawOutputLoaded         bool
}

type BackendReport struct {
	Available                        bool
	Enabled                          bool
	Status                           agentxcontrolplane.HostActionStatus
	BackendKind                      string
	BackendRef                       agentxcontrolplane.DisplaySafeRef
	Readiness                        agentxcontrolplane.HostOwnedDelegationWorkerRuntimeReadiness
	Invocation                       agentxcontrolplane.HostOwnedDelegationWorkerRuntimeInvocation
	InvocationReportRef              agentxcontrolplane.DisplaySafeRef
	ObservedInvocationRef            agentxcontrolplane.DisplaySafeRef
	HostWorkerRuntimeRunRef          agentxcontrolplane.DisplaySafeRef
	VisibleToolRefs                  []agentxcontrolplane.DisplaySafeRef
	ContextRefs                      []agentxcontrolplane.DisplaySafeRef
	BudgetRef                        agentxcontrolplane.DisplaySafeRef
	TimeoutRef                       agentxcontrolplane.DisplaySafeRef
	ParallelismRef                   agentxcontrolplane.DisplaySafeRef
	WorkerRunAttempted               bool
	WorkerReadbackAttempted          bool
	WorkerResultRecorded             bool
	WorkerResultReadbackReady        bool
	ReadyForWorkerResultReview       bool
	ReadyForFailureReview            bool
	WorkerResultRequiresVerification bool
	WorkerOutputAcceptedAsFact       bool
	EvidenceRefs                     []agentxcontrolplane.EvidenceRef
	MissingInputs                    []agentxcontrolplane.MissingInput
	BlockedReasons                   []string
	FailureClass                     agentxcontrolplane.FailureClass
	Boundaries                       []agentxcontrolplane.Boundary
	NextHostAction                   agentxcontrolplane.NextHostAction
	RawOutputLoaded                  bool
}

type WorkerRequest struct {
	Readiness               agentxcontrolplane.HostOwnedDelegationWorkerRuntimeReadiness
	WorkerRunRef            agentxcontrolplane.DisplaySafeRef
	WorkerRequestRef        agentxcontrolplane.DisplaySafeRef
	InvocationRef           agentxcontrolplane.DisplaySafeRef
	HostWorkerRuntimeRunRef agentxcontrolplane.DisplaySafeRef
	VisibleToolRefs         []agentxcontrolplane.DisplaySafeRef
	ContextRefs             []agentxcontrolplane.DisplaySafeRef
	BudgetRef               agentxcontrolplane.DisplaySafeRef
	TimeoutRef              agentxcontrolplane.DisplaySafeRef
	ParallelismRef          agentxcontrolplane.DisplaySafeRef
	EvidenceRequirements    []agentxcontrolplane.EvidenceRef
	Boundaries              []agentxcontrolplane.Boundary
}

type WorkerResult struct {
	Completed            bool
	Failed               bool
	ObservedWorkerRunRef agentxcontrolplane.DisplaySafeRef
	WorkerResultRef      agentxcontrolplane.DisplaySafeRef
	WorkerReadbackRef    agentxcontrolplane.DisplaySafeRef
	ObservationRef       agentxcontrolplane.DisplaySafeRef
	FailureRef           agentxcontrolplane.DisplaySafeRef
	CompensationRef      agentxcontrolplane.DisplaySafeRef
	EvidenceRefs         []agentxcontrolplane.EvidenceRef
	MissingInputs        []agentxcontrolplane.MissingInput
	BlockedReasons       []string
	Boundaries           []agentxcontrolplane.Boundary
	NextHostAction       agentxcontrolplane.NextHostAction
	RawOutputLoaded      bool
}

type WorkerReadbackRequest struct {
	WorkerRunRef      agentxcontrolplane.DisplaySafeRef
	WorkerResultRef   agentxcontrolplane.DisplaySafeRef
	WorkerReadbackRef agentxcontrolplane.DisplaySafeRef
	Record            WorkerRunRecord
	Boundaries        []agentxcontrolplane.Boundary
}

type WorkerReadback struct {
	Ready                bool
	ResultVisible        bool
	ObservedWorkerRunRef agentxcontrolplane.DisplaySafeRef
	WorkerResultRef      agentxcontrolplane.DisplaySafeRef
	WorkerReadbackRef    agentxcontrolplane.DisplaySafeRef
	ObservationRef       agentxcontrolplane.DisplaySafeRef
	EvidenceRefs         []agentxcontrolplane.EvidenceRef
	MissingInputs        []agentxcontrolplane.MissingInput
	BlockedReasons       []string
	Boundaries           []agentxcontrolplane.Boundary
	NextHostAction       agentxcontrolplane.NextHostAction
	RawOutputLoaded      bool
}

type WorkerRunRecord struct {
	BackendKind             string
	BackendRef              agentxcontrolplane.DisplaySafeRef
	WorkerRunRef            agentxcontrolplane.DisplaySafeRef
	WorkerRequestRef        agentxcontrolplane.DisplaySafeRef
	InvocationRef           agentxcontrolplane.DisplaySafeRef
	InvocationReportRef     agentxcontrolplane.DisplaySafeRef
	HostWorkerRuntimeRunRef agentxcontrolplane.DisplaySafeRef
	WorkerResultRef         agentxcontrolplane.DisplaySafeRef
	WorkerReadbackRef       agentxcontrolplane.DisplaySafeRef
	ObservationRef          agentxcontrolplane.DisplaySafeRef
	Status                  string
	VisibleToolRefs         []agentxcontrolplane.DisplaySafeRef
	ContextRefs             []agentxcontrolplane.DisplaySafeRef
	BudgetRef               agentxcontrolplane.DisplaySafeRef
	TimeoutRef              agentxcontrolplane.DisplaySafeRef
	ParallelismRef          agentxcontrolplane.DisplaySafeRef
	EvidenceRefs            []agentxcontrolplane.EvidenceRef
	Boundaries              []agentxcontrolplane.Boundary
	UpdatedAt               time.Time
}

type InMemoryStateStore struct {
	mu      sync.Mutex
	records map[string]WorkerRunRecord
}

func NewInMemoryStateStore() *InMemoryStateStore {
	return &InMemoryStateStore{records: map[string]WorkerRunRecord{}}
}

func (s *InMemoryStateStore) WriteWorkerRun(_ context.Context, record WorkerRunRecord) error {
	if s == nil {
		return nil
	}
	record = record.Normalize()
	key := strings.TrimSpace(string(record.WorkerRunRef))
	if key == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.records == nil {
		s.records = map[string]WorkerRunRecord{}
	}
	s.records[key] = record
	return nil
}

func (s *InMemoryStateStore) ReadWorkerRun(_ context.Context, workerRunRef agentxcontrolplane.DisplaySafeRef) (WorkerRunRecord, bool, error) {
	if s == nil {
		return WorkerRunRecord{}, false, nil
	}
	ref, ok := normalizeDisplayRef(workerRunRef)
	if !ok {
		return WorkerRunRecord{}, false, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record, visible := s.records[string(ref)]
	return record.Normalize(), visible, nil
}

func (b Backend) RunDelegationWorkerRuntime(ctx context.Context, input BackendInput) (BackendReport, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	report := b.reportFromInput(input)
	if input.RawOutputLoaded || backendInputUnsafe(input, b.BackendRef) {
		report.RawOutputLoaded = true
		return report.block(agentxcontrolplane.FailureEvidenceWeak, "unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed").Normalize(), nil
	}
	if !input.Enabled {
		return report.block(agentxcontrolplane.FailureApprovalRequired, "delegation_worker_runtime_backend_not_enabled", "host:delegation_worker_runtime_enablement", "enable_delegation_worker_runtime_backend", "delegation_worker_runtime_backend_default_off").Normalize(), nil
	}
	if !report.Readiness.ReadyForHostWorkerRuntimeInvocation {
		report.MissingInputs = agentxcontrolplane.AppendMissingInputs(report.MissingInputs, report.Readiness.MissingInputs...)
		report.BlockedReasons = appendUniqueStrings(report.BlockedReasons, report.Readiness.BlockedReasons...)
		return report.block(firstFailureClass(report.Readiness.FailureClass, agentxcontrolplane.FailurePolicyBlocked), "delegation_worker_runtime_readiness_not_ready", "host:delegation_worker_runtime_readiness", "review_delegation_worker_runtime_readiness", "delegation_worker_runtime_readiness_not_ready").Normalize(), nil
	}
	if b.Runtime == nil {
		report = report.block(agentxcontrolplane.FailureHostAdapterMissing, "delegation_worker_runtime_worker_missing", "host:delegation_worker_runtime_worker", "provide_delegation_worker_runtime_worker", "delegation_worker_runtime_worker_missing")
	}
	if b.Store == nil {
		report = report.block(agentxcontrolplane.FailureHostAdapterMissing, "delegation_worker_runtime_state_store_missing", "host:delegation_worker_runtime_state_store", "provide_delegation_worker_runtime_state_store", "delegation_worker_runtime_state_store_missing")
	}
	report = requireWorkerRuntimeScope(report)
	if len(report.MissingInputs) > 0 || len(report.BlockedReasons) > 0 {
		return report.Normalize(), nil
	}

	workerRequest := WorkerRequest{
		Readiness:               report.Readiness,
		WorkerRunRef:            report.Readiness.WorkerRunRef,
		WorkerRequestRef:        report.Readiness.WorkerRequestRef,
		InvocationRef:           report.Readiness.InvocationRef,
		HostWorkerRuntimeRunRef: report.HostWorkerRuntimeRunRef,
		VisibleToolRefs:         cloneDisplayRefs(report.VisibleToolRefs),
		ContextRefs:             cloneDisplayRefs(report.ContextRefs),
		BudgetRef:               report.BudgetRef,
		TimeoutRef:              report.TimeoutRef,
		ParallelismRef:          report.ParallelismRef,
		EvidenceRequirements:    cloneEvidenceRefs(report.Readiness.EvidenceRefs),
		Boundaries: agentxcontrolplane.AppendBoundaries(report.Boundaries,
			"host_owned_delegation_worker_runtime_backend_request",
			"worker_scope_bound_by_host",
		),
	}
	report.WorkerRunAttempted = true
	workerResult, err := b.Runtime.InvokeDelegationWorker(ctx, workerRequest)
	if err != nil {
		workerResult = WorkerResult{
			Failed:               true,
			ObservedWorkerRunRef: report.Readiness.WorkerRunRef,
			FailureRef:           input.FailureRef,
			CompensationRef:      firstDisplayRef(input.CompensationRef, report.Readiness.CompensationRef),
			BlockedReasons:       []string{"delegation_worker_runtime_invocation_failed"},
			Boundaries:           []agentxcontrolplane.Boundary{"delegation_worker_runtime_invocation_failed"},
			NextHostAction:       "review_delegation_worker_runtime_failure",
		}
	}
	workerResult = workerResult.Normalize()
	if workerResult.RawOutputLoaded || workerResultUnsafe(workerResult) {
		report.RawOutputLoaded = true
		return report.block(agentxcontrolplane.FailureEvidenceWeak, "worker_result_unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed").Normalize(), nil
	}
	report.MissingInputs = agentxcontrolplane.AppendMissingInputs(report.MissingInputs, workerResult.MissingInputs...)
	report.BlockedReasons = appendUniqueStrings(report.BlockedReasons, workerResult.BlockedReasons...)
	report.Boundaries = agentxcontrolplane.AppendBoundaries(report.Boundaries, workerResult.Boundaries...)
	if workerResult.Failed || !workerResult.Completed {
		return b.recordFailureInvocation(report, input, workerResult).Normalize(), nil
	}
	if len(workerResult.EvidenceRefs) == 0 {
		return report.block(agentxcontrolplane.FailureEvidenceMissing, "delegation_worker_result_evidence_refs_missing", "host:delegation_worker_result_evidence_refs", "provide_delegation_worker_result_evidence_refs", "delegation_worker_result_evidence_required").Normalize(), nil
	}
	report = requireCompletedWorkerResultRefs(report, workerResult)
	if len(report.MissingInputs) > 0 || len(report.BlockedReasons) > 0 {
		return report.Normalize(), nil
	}

	record := workerRunRecordFromResult(report, b, workerResult)
	if err := b.Store.WriteWorkerRun(ctx, record); err != nil {
		return report.block(agentxcontrolplane.FailureInternalError, "delegation_worker_runtime_state_store_write_failed", "host:delegation_worker_runtime_state_store", "fix_delegation_worker_runtime_state_store", "delegation_worker_runtime_state_store_write_failed").Normalize(), nil
	}
	report.WorkerResultRecorded = true
	storedRecord, visible, err := b.Store.ReadWorkerRun(ctx, report.Readiness.WorkerRunRef)
	if err != nil {
		return report.block(agentxcontrolplane.FailureInternalError, "delegation_worker_runtime_state_store_read_failed", "host:delegation_worker_runtime_state_store", "fix_delegation_worker_runtime_state_store", "delegation_worker_runtime_state_store_read_failed").Normalize(), nil
	}
	if !visible {
		return report.block(agentxcontrolplane.FailureEvidenceMissing, "delegation_worker_runtime_state_record_not_visible", "host:delegation_worker_runtime_state_record", "readback_delegation_worker_runtime_state_record", "delegation_worker_runtime_state_record_not_visible").Normalize(), nil
	}
	if !workerRunRecordMatches(storedRecord, record) {
		return report.block(agentxcontrolplane.FailureVerificationFailed, "delegation_worker_runtime_state_record_mismatch", "host:delegation_worker_runtime_state_record", "repair_delegation_worker_runtime_state_record", "delegation_worker_runtime_state_record_mismatch").Normalize(), nil
	}

	report.WorkerReadbackAttempted = true
	readback, err := b.Runtime.ReadDelegationWorkerResult(ctx, WorkerReadbackRequest{
		WorkerRunRef:      report.Readiness.WorkerRunRef,
		WorkerResultRef:   workerResult.WorkerResultRef,
		WorkerReadbackRef: workerResult.WorkerReadbackRef,
		Record:            storedRecord,
		Boundaries: []agentxcontrolplane.Boundary{
			"host_owned_delegation_worker_runtime_backend_readback",
		},
	})
	if err != nil {
		return report.block(agentxcontrolplane.FailureInternalError, "delegation_worker_runtime_readback_failed", "host:delegation_worker_runtime_readback", "fix_delegation_worker_runtime_readback", "delegation_worker_runtime_readback_failed").Normalize(), nil
	}
	readback = readback.Normalize()
	if readback.RawOutputLoaded || workerReadbackUnsafe(readback) {
		report.RawOutputLoaded = true
		return report.block(agentxcontrolplane.FailureEvidenceWeak, "worker_readback_unsafe_input_ref", "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed").Normalize(), nil
	}
	report.MissingInputs = agentxcontrolplane.AppendMissingInputs(report.MissingInputs, readback.MissingInputs...)
	report.BlockedReasons = appendUniqueStrings(report.BlockedReasons, readback.BlockedReasons...)
	report.Boundaries = agentxcontrolplane.AppendBoundaries(report.Boundaries, readback.Boundaries...)
	if !readback.Ready || !readback.ResultVisible {
		return report.block(agentxcontrolplane.FailureEvidenceMissing, "delegation_worker_runtime_readback_not_ready", "host:delegation_worker_runtime_readback", "readback_delegation_worker_runtime_result", "delegation_worker_runtime_readback_not_ready").Normalize(), nil
	}
	if !workerReadbackMatchesResult(readback, workerResult, report.Readiness.WorkerRunRef) {
		return report.block(agentxcontrolplane.FailureVerificationFailed, "delegation_worker_runtime_readback_mismatch", "host:delegation_worker_runtime_readback", "repair_delegation_worker_runtime_readback", "delegation_worker_runtime_readback_mismatch").Normalize(), nil
	}
	report.WorkerResultReadbackReady = true
	report.EvidenceRefs = agentxcontrolplane.MergeEvidenceRefs(report.EvidenceRefs, workerResult.EvidenceRefs, readback.EvidenceRefs)
	report.Invocation = agentxcontrolplane.BuildHostOwnedDelegationWorkerRuntimeInvocation(agentxcontrolplane.HostOwnedDelegationWorkerRuntimeInvocationInput{
		Readiness:               report.Readiness,
		InvocationReportRef:     report.InvocationReportRef,
		ObservedInvocationRef:   report.ObservedInvocationRef,
		HostWorkerRuntimeRunRef: report.HostWorkerRuntimeRunRef,
		ObservedWorkerRunRef:    readback.ObservedWorkerRunRef,
		WorkerResultRef:         readback.WorkerResultRef,
		WorkerReadbackRef:       readback.WorkerReadbackRef,
		ObservationRef:          readback.ObservationRef,
		HostInvocationReported:  true,
		HostInvocationCompleted: true,
		EvidenceRefs:            report.EvidenceRefs,
		Boundaries: agentxcontrolplane.AppendBoundaries(report.Boundaries,
			"host_owned_delegation_worker_runtime_backend",
			"worker_result_readback_verified_by_host_backend",
		),
	})
	report = report.bindInvocation(report.Invocation)
	return report.Normalize(), nil
}

func (b Backend) reportFromInput(input BackendInput) BackendReport {
	readiness := input.Readiness.Normalize()
	observedInvocationRef := input.ObservedInvocationRef
	if observedInvocationRef == "" {
		observedInvocationRef = readiness.InvocationRef
	}
	return BackendReport{
		Available:                        b.Runtime != nil && b.Store != nil,
		Enabled:                          input.Enabled,
		Status:                           agentxcontrolplane.HostActionBlocked,
		BackendKind:                      firstNonEmpty(b.BackendKind, BackendKind),
		BackendRef:                       firstDisplayRef(b.BackendRef, agentxcontrolplane.DisplaySafeRef("backend:"+BackendKind)),
		Readiness:                        readiness,
		InvocationReportRef:              input.InvocationReportRef,
		ObservedInvocationRef:            observedInvocationRef,
		HostWorkerRuntimeRunRef:          input.HostWorkerRuntimeRunRef,
		VisibleToolRefs:                  normalizeDisplayRefs(input.VisibleToolRefs),
		ContextRefs:                      normalizeDisplayRefs(input.ContextRefs),
		BudgetRef:                        firstDisplayRef(input.BudgetRef, readiness.BudgetRef),
		TimeoutRef:                       input.TimeoutRef,
		ParallelismRef:                   input.ParallelismRef,
		WorkerResultRequiresVerification: true,
		WorkerOutputAcceptedAsFact:       false,
		EvidenceRefs:                     agentxcontrolplane.MergeEvidenceRefs(input.EvidenceRefs, readiness.EvidenceRefs),
		FailureClass:                     agentxcontrolplane.FailureNone,
		Boundaries: agentxcontrolplane.AppendBoundaries([]agentxcontrolplane.Boundary{
			"host_owned_delegation_worker_runtime_backend",
			"host_adapter_executes_worker_only_when_enabled",
			"worker_scope_bound_by_host",
			"display_safe_refs_only",
			"worker_output_not_fact",
			"worker_result_requires_verification",
			"parent_verification_required_before_merge",
		}, input.Boundaries...),
		NextHostAction:  "provide_delegation_worker_runtime_backend_inputs",
		RawOutputLoaded: input.RawOutputLoaded,
	}
}

func (b Backend) recordFailureInvocation(report BackendReport, input BackendInput, workerResult WorkerResult) BackendReport {
	failureRef := firstDisplayRef(workerResult.FailureRef, input.FailureRef)
	compensationRef := firstDisplayRef(workerResult.CompensationRef, input.CompensationRef, report.Readiness.CompensationRef)
	report.EvidenceRefs = agentxcontrolplane.MergeEvidenceRefs(report.EvidenceRefs, workerResult.EvidenceRefs)
	report.Invocation = agentxcontrolplane.BuildHostOwnedDelegationWorkerRuntimeInvocation(agentxcontrolplane.HostOwnedDelegationWorkerRuntimeInvocationInput{
		Readiness:               report.Readiness,
		InvocationReportRef:     report.InvocationReportRef,
		ObservedInvocationRef:   report.ObservedInvocationRef,
		HostWorkerRuntimeRunRef: report.HostWorkerRuntimeRunRef,
		ObservedWorkerRunRef:    firstDisplayRef(workerResult.ObservedWorkerRunRef, report.Readiness.WorkerRunRef),
		FailureRef:              failureRef,
		CompensationRef:         compensationRef,
		HostInvocationReported:  true,
		HostInvocationFailed:    true,
		EvidenceRefs:            report.EvidenceRefs,
		Boundaries: agentxcontrolplane.AppendBoundaries(report.Boundaries,
			"host_owned_delegation_worker_runtime_backend_failure",
		),
	})
	return report.bindInvocation(report.Invocation)
}

func workerRunRecordFromResult(report BackendReport, b Backend, result WorkerResult) WorkerRunRecord {
	return WorkerRunRecord{
		BackendKind:             firstNonEmpty(b.BackendKind, BackendKind),
		BackendRef:              report.BackendRef,
		WorkerRunRef:            report.Readiness.WorkerRunRef,
		WorkerRequestRef:        report.Readiness.WorkerRequestRef,
		InvocationRef:           report.Readiness.InvocationRef,
		InvocationReportRef:     report.InvocationReportRef,
		HostWorkerRuntimeRunRef: report.HostWorkerRuntimeRunRef,
		WorkerResultRef:         result.WorkerResultRef,
		WorkerReadbackRef:       result.WorkerReadbackRef,
		ObservationRef:          result.ObservationRef,
		Status:                  "completed",
		VisibleToolRefs:         cloneDisplayRefs(report.VisibleToolRefs),
		ContextRefs:             cloneDisplayRefs(report.ContextRefs),
		BudgetRef:               report.BudgetRef,
		TimeoutRef:              report.TimeoutRef,
		ParallelismRef:          report.ParallelismRef,
		EvidenceRefs:            cloneEvidenceRefs(result.EvidenceRefs),
		Boundaries: agentxcontrolplane.AppendBoundaries(report.Boundaries,
			"delegation_worker_runtime_state_recorded",
		),
		UpdatedAt: currentTime(b.Now),
	}.Normalize()
}

func (r BackendReport) block(failure agentxcontrolplane.FailureClass, reason string, missing agentxcontrolplane.MissingInput, next agentxcontrolplane.NextHostAction, boundary agentxcontrolplane.Boundary) BackendReport {
	r.Status = agentxcontrolplane.HostActionBlocked
	r.ReadyForWorkerResultReview = false
	r.ReadyForFailureReview = false
	r.WorkerResultReadbackReady = false
	r.FailureClass = firstFailureClass(r.FailureClass, failure)
	r.MissingInputs = agentxcontrolplane.AppendMissingInputs(r.MissingInputs, missing)
	r.BlockedReasons = appendUniqueStrings(r.BlockedReasons, reason)
	if next != "" {
		r.NextHostAction = next
	}
	r.Boundaries = agentxcontrolplane.AppendBoundaries(r.Boundaries, boundary)
	return r
}

func (r BackendReport) bindInvocation(invocation agentxcontrolplane.HostOwnedDelegationWorkerRuntimeInvocation) BackendReport {
	invocation = invocation.Normalize()
	r.Invocation = invocation
	r.Status = invocation.Status
	r.ReadyForWorkerResultReview = invocation.ReadyForWorkerResultReview
	r.ReadyForFailureReview = invocation.ReadyForFailureReview
	r.MissingInputs = agentxcontrolplane.AppendMissingInputs(r.MissingInputs, invocation.MissingInputs...)
	r.BlockedReasons = appendUniqueStrings(r.BlockedReasons, invocation.BlockedReasons...)
	r.Boundaries = agentxcontrolplane.AppendBoundaries(r.Boundaries, invocation.Boundaries...)
	r.NextHostAction = invocation.NextHostAction
	r.FailureClass = firstFailureClass(invocation.FailureClass, r.FailureClass)
	return r
}

func (r BackendReport) Normalize() BackendReport {
	out := r
	out.Readiness = out.Readiness.Normalize()
	out.Invocation = out.Invocation.Normalize()
	out.BackendKind = firstNonEmpty(out.BackendKind, BackendKind)
	out.BackendRef, _ = normalizeDisplayRef(out.BackendRef)
	out.InvocationReportRef, _ = normalizeDisplayRef(out.InvocationReportRef)
	out.ObservedInvocationRef, _ = normalizeDisplayRef(out.ObservedInvocationRef)
	out.HostWorkerRuntimeRunRef, _ = normalizeDisplayRef(out.HostWorkerRuntimeRunRef)
	out.VisibleToolRefs = normalizeDisplayRefs(out.VisibleToolRefs)
	out.ContextRefs = normalizeDisplayRefs(out.ContextRefs)
	out.BudgetRef, _ = normalizeDisplayRef(out.BudgetRef)
	out.TimeoutRef, _ = normalizeDisplayRef(out.TimeoutRef)
	out.ParallelismRef, _ = normalizeDisplayRef(out.ParallelismRef)
	out.EvidenceRefs = agentxcontrolplane.MergeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = agentxcontrolplane.AppendMissingInputs(nil, out.MissingInputs...)
	out.BlockedReasons = appendUniqueStrings(nil, out.BlockedReasons...)
	out.FailureClass = firstFailureClass(out.FailureClass, agentxcontrolplane.FailureNone)
	out.Boundaries = agentxcontrolplane.AppendBoundaries(nil, out.Boundaries...)
	out.WorkerResultRequiresVerification = true
	out.WorkerOutputAcceptedAsFact = false
	if out.Status == "" {
		out.Status = agentxcontrolplane.HostActionBlocked
	}
	if out.RawOutputLoaded && out.FailureClass == agentxcontrolplane.FailureNone {
		out.FailureClass = agentxcontrolplane.FailureEvidenceWeak
	}
	if len(out.MissingInputs) > 0 || len(out.BlockedReasons) > 0 || out.RawOutputLoaded {
		out.ReadyForWorkerResultReview = false
		out.ReadyForFailureReview = out.ReadyForFailureReview && !out.RawOutputLoaded
		if out.Status == agentxcontrolplane.HostActionRecorded {
			out.Status = agentxcontrolplane.HostActionBlocked
		}
	}
	return out
}

func (r WorkerResult) Normalize() WorkerResult {
	out := r
	out.ObservedWorkerRunRef, _ = normalizeDisplayRef(out.ObservedWorkerRunRef)
	out.WorkerResultRef, _ = normalizeDisplayRef(out.WorkerResultRef)
	out.WorkerReadbackRef, _ = normalizeDisplayRef(out.WorkerReadbackRef)
	out.ObservationRef, _ = normalizeDisplayRef(out.ObservationRef)
	out.FailureRef, _ = normalizeDisplayRef(out.FailureRef)
	out.CompensationRef, _ = normalizeDisplayRef(out.CompensationRef)
	out.EvidenceRefs = agentxcontrolplane.MergeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = agentxcontrolplane.AppendMissingInputs(nil, out.MissingInputs...)
	out.BlockedReasons = appendUniqueStrings(nil, out.BlockedReasons...)
	out.Boundaries = agentxcontrolplane.AppendBoundaries(nil, out.Boundaries...)
	return out
}

func (r WorkerReadback) Normalize() WorkerReadback {
	out := r
	out.ObservedWorkerRunRef, _ = normalizeDisplayRef(out.ObservedWorkerRunRef)
	out.WorkerResultRef, _ = normalizeDisplayRef(out.WorkerResultRef)
	out.WorkerReadbackRef, _ = normalizeDisplayRef(out.WorkerReadbackRef)
	out.ObservationRef, _ = normalizeDisplayRef(out.ObservationRef)
	out.EvidenceRefs = agentxcontrolplane.MergeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = agentxcontrolplane.AppendMissingInputs(nil, out.MissingInputs...)
	out.BlockedReasons = appendUniqueStrings(nil, out.BlockedReasons...)
	out.Boundaries = agentxcontrolplane.AppendBoundaries(nil, out.Boundaries...)
	return out
}

func (r WorkerRunRecord) Normalize() WorkerRunRecord {
	out := r
	out.BackendKind = firstNonEmpty(out.BackendKind, BackendKind)
	out.BackendRef, _ = normalizeDisplayRef(out.BackendRef)
	out.WorkerRunRef, _ = normalizeDisplayRef(out.WorkerRunRef)
	out.WorkerRequestRef, _ = normalizeDisplayRef(out.WorkerRequestRef)
	out.InvocationRef, _ = normalizeDisplayRef(out.InvocationRef)
	out.InvocationReportRef, _ = normalizeDisplayRef(out.InvocationReportRef)
	out.HostWorkerRuntimeRunRef, _ = normalizeDisplayRef(out.HostWorkerRuntimeRunRef)
	out.WorkerResultRef, _ = normalizeDisplayRef(out.WorkerResultRef)
	out.WorkerReadbackRef, _ = normalizeDisplayRef(out.WorkerReadbackRef)
	out.ObservationRef, _ = normalizeDisplayRef(out.ObservationRef)
	out.Status = strings.TrimSpace(out.Status)
	out.VisibleToolRefs = normalizeDisplayRefs(out.VisibleToolRefs)
	out.ContextRefs = normalizeDisplayRefs(out.ContextRefs)
	out.BudgetRef, _ = normalizeDisplayRef(out.BudgetRef)
	out.TimeoutRef, _ = normalizeDisplayRef(out.TimeoutRef)
	out.ParallelismRef, _ = normalizeDisplayRef(out.ParallelismRef)
	out.EvidenceRefs = agentxcontrolplane.MergeEvidenceRefs(out.EvidenceRefs)
	out.Boundaries = agentxcontrolplane.AppendBoundaries(nil, out.Boundaries...)
	return out
}

func requireWorkerRuntimeScope(report BackendReport) BackendReport {
	checks := []struct {
		ok       bool
		reason   string
		missing  agentxcontrolplane.MissingInput
		next     agentxcontrolplane.NextHostAction
		boundary agentxcontrolplane.Boundary
	}{
		{report.InvocationReportRef != "", "delegation_worker_runtime_invocation_report_ref_missing", "host:delegation_worker_runtime_invocation_report", "provide_delegation_worker_runtime_invocation_report", "delegation_worker_runtime_invocation_report_ref_missing"},
		{report.ObservedInvocationRef != "", "delegation_worker_runtime_observed_invocation_ref_missing", "host:delegation_worker_runtime_observed_invocation", "provide_delegation_worker_runtime_invocation_report", "delegation_worker_runtime_observed_invocation_ref_missing"},
		{report.HostWorkerRuntimeRunRef != "", "delegation_worker_runtime_run_ref_missing", "host:delegation_worker_runtime_run_ref", "provide_delegation_worker_runtime_run_ref", "delegation_worker_runtime_run_ref_missing"},
		{len(report.VisibleToolRefs) > 0, "delegation_worker_visible_tools_missing", "host:delegation_worker_visible_tools", "provide_delegation_worker_tool_boundary", "delegation_worker_visible_tools_required"},
		{len(report.ContextRefs) > 0, "delegation_worker_context_refs_missing", "host:delegation_worker_context_refs", "provide_delegation_worker_context_refs", "delegation_worker_context_refs_required"},
		{report.BudgetRef != "", "delegation_worker_budget_ref_missing", "host:delegation_worker_budget", "provide_delegation_worker_budget", "delegation_worker_budget_ref_required"},
		{report.TimeoutRef != "", "delegation_worker_timeout_ref_missing", "host:delegation_worker_timeout", "provide_delegation_worker_timeout", "delegation_worker_timeout_ref_required"},
		{report.ParallelismRef != "", "delegation_worker_parallelism_ref_missing", "host:delegation_worker_parallelism", "provide_delegation_worker_parallelism", "delegation_worker_parallelism_ref_required"},
	}
	for _, check := range checks {
		if !check.ok {
			report = report.block(agentxcontrolplane.FailureConfigMissing, check.reason, check.missing, check.next, check.boundary)
		}
	}
	return report
}

func requireCompletedWorkerResultRefs(report BackendReport, result WorkerResult) BackendReport {
	checks := []struct {
		ok       bool
		reason   string
		missing  agentxcontrolplane.MissingInput
		next     agentxcontrolplane.NextHostAction
		boundary agentxcontrolplane.Boundary
	}{
		{result.ObservedWorkerRunRef != "", "delegation_worker_result_observed_run_ref_missing", "host:delegation_worker_run_ref", "provide_delegation_worker_run_ref", "delegation_worker_result_observed_run_ref_required"},
		{result.WorkerResultRef != "", "delegation_worker_result_ref_missing", "host:delegation_worker_result_ref", "provide_delegation_worker_result_ref", "delegation_worker_result_ref_required"},
		{result.WorkerReadbackRef != "", "delegation_worker_readback_ref_missing", "host:delegation_worker_readback_ref", "provide_delegation_worker_readback_ref", "delegation_worker_readback_ref_required"},
		{result.ObservationRef != "", "delegation_worker_observation_ref_missing", "host:delegation_worker_observation_ref", "provide_delegation_worker_observation_ref", "delegation_worker_observation_ref_required"},
	}
	for _, check := range checks {
		if !check.ok {
			report = report.block(agentxcontrolplane.FailureEvidenceMissing, check.reason, check.missing, check.next, check.boundary)
		}
	}
	return report
}

func workerRunRecordMatches(a, b WorkerRunRecord) bool {
	a = a.Normalize()
	b = b.Normalize()
	return a.WorkerRunRef == b.WorkerRunRef &&
		a.WorkerRequestRef == b.WorkerRequestRef &&
		a.InvocationRef == b.InvocationRef &&
		a.InvocationReportRef == b.InvocationReportRef &&
		a.HostWorkerRuntimeRunRef == b.HostWorkerRuntimeRunRef &&
		a.WorkerResultRef == b.WorkerResultRef &&
		a.WorkerReadbackRef == b.WorkerReadbackRef &&
		a.ObservationRef == b.ObservationRef &&
		a.BudgetRef == b.BudgetRef &&
		a.TimeoutRef == b.TimeoutRef &&
		a.ParallelismRef == b.ParallelismRef &&
		displayRefsEqual(a.VisibleToolRefs, b.VisibleToolRefs) &&
		displayRefsEqual(a.ContextRefs, b.ContextRefs)
}

func workerReadbackMatchesResult(readback WorkerReadback, result WorkerResult, workerRunRef agentxcontrolplane.DisplaySafeRef) bool {
	readback = readback.Normalize()
	result = result.Normalize()
	workerRunRef, _ = normalizeDisplayRef(workerRunRef)
	return readback.ObservedWorkerRunRef == workerRunRef &&
		readback.WorkerResultRef == result.WorkerResultRef &&
		readback.WorkerReadbackRef == result.WorkerReadbackRef &&
		readback.ObservationRef == result.ObservationRef
}

func backendInputUnsafe(input BackendInput, backendRef agentxcontrolplane.DisplaySafeRef) bool {
	return displayRefUnsafe(backendRef) ||
		displayRefUnsafe(input.InvocationReportRef) ||
		displayRefUnsafe(input.ObservedInvocationRef) ||
		displayRefUnsafe(input.HostWorkerRuntimeRunRef) ||
		displayRefsUnsafe(input.VisibleToolRefs) ||
		displayRefsUnsafe(input.ContextRefs) ||
		displayRefUnsafe(input.BudgetRef) ||
		displayRefUnsafe(input.TimeoutRef) ||
		displayRefUnsafe(input.ParallelismRef) ||
		displayRefUnsafe(input.FailureRef) ||
		displayRefUnsafe(input.CompensationRef) ||
		evidenceRefsUnsafe(input.EvidenceRefs)
}

func workerResultUnsafe(result WorkerResult) bool {
	return displayRefUnsafe(result.ObservedWorkerRunRef) ||
		displayRefUnsafe(result.WorkerResultRef) ||
		displayRefUnsafe(result.WorkerReadbackRef) ||
		displayRefUnsafe(result.ObservationRef) ||
		displayRefUnsafe(result.FailureRef) ||
		displayRefUnsafe(result.CompensationRef) ||
		evidenceRefsUnsafe(result.EvidenceRefs)
}

func workerReadbackUnsafe(readback WorkerReadback) bool {
	return displayRefUnsafe(readback.ObservedWorkerRunRef) ||
		displayRefUnsafe(readback.WorkerResultRef) ||
		displayRefUnsafe(readback.WorkerReadbackRef) ||
		displayRefUnsafe(readback.ObservationRef) ||
		evidenceRefsUnsafe(readback.EvidenceRefs)
}

func displayRefUnsafe(ref agentxcontrolplane.DisplaySafeRef) bool {
	raw := strings.TrimSpace(string(ref))
	if raw == "" {
		return false
	}
	_, ok := agentxcontrolplane.NormalizeDisplaySafeRef(raw)
	return !ok
}

func displayRefsUnsafe(refs []agentxcontrolplane.DisplaySafeRef) bool {
	for _, ref := range refs {
		if displayRefUnsafe(ref) {
			return true
		}
	}
	return false
}

func evidenceRefsUnsafe(refs []agentxcontrolplane.EvidenceRef) bool {
	for _, ref := range refs {
		if displayRefUnsafe(ref.Ref) || displayRefUnsafe(ref.Source) {
			return true
		}
	}
	return false
}

func normalizeDisplayRef(ref agentxcontrolplane.DisplaySafeRef) (agentxcontrolplane.DisplaySafeRef, bool) {
	return agentxcontrolplane.NormalizeDisplaySafeRef(string(ref))
}

func normalizeDisplayRefs(refs []agentxcontrolplane.DisplaySafeRef) []agentxcontrolplane.DisplaySafeRef {
	out := make([]agentxcontrolplane.DisplaySafeRef, 0, len(refs))
	seen := map[agentxcontrolplane.DisplaySafeRef]struct{}{}
	for _, ref := range refs {
		normalized, ok := normalizeDisplayRef(ref)
		if !ok {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func cloneDisplayRefs(refs []agentxcontrolplane.DisplaySafeRef) []agentxcontrolplane.DisplaySafeRef {
	return append([]agentxcontrolplane.DisplaySafeRef(nil), refs...)
}

func cloneEvidenceRefs(refs []agentxcontrolplane.EvidenceRef) []agentxcontrolplane.EvidenceRef {
	return append([]agentxcontrolplane.EvidenceRef(nil), refs...)
}

func displayRefsEqual(a, b []agentxcontrolplane.DisplaySafeRef) bool {
	a = normalizeDisplayRefs(a)
	b = normalizeDisplayRefs(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func firstDisplayRef(values ...agentxcontrolplane.DisplaySafeRef) agentxcontrolplane.DisplaySafeRef {
	for _, value := range values {
		ref, ok := normalizeDisplayRef(value)
		if ok {
			return ref
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstFailureClass(values ...agentxcontrolplane.FailureClass) agentxcontrolplane.FailureClass {
	for _, value := range values {
		normalized := agentxcontrolplane.NormalizeFailureClass(string(value))
		if normalized != "" && normalized != agentxcontrolplane.FailureNone {
			return normalized
		}
	}
	return agentxcontrolplane.FailureNone
}

func appendUniqueStrings(base []string, values ...string) []string {
	out := append([]string(nil), base...)
	seen := map[string]struct{}{}
	for _, value := range out {
		seen[value] = struct{}{}
	}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func currentTime(now func() time.Time) time.Time {
	if now != nil {
		return now()
	}
	return time.Now().UTC()
}
