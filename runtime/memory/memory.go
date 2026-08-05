package memory

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"
)

// ErrorCode is a stable memory-lifecycle failure category.
type ErrorCode string

const (
	// ErrorCodeInvalidPolicy marks unusable coordinator limits.
	ErrorCodeInvalidPolicy ErrorCode = "invalid_policy"
	// ErrorCodeInvalidRequest marks an invalid caller-owned request.
	ErrorCodeInvalidRequest ErrorCode = "invalid_request"
	// ErrorCodeCanceled marks caller cancellation.
	ErrorCodeCanceled ErrorCode = "canceled"
	// ErrorCodeDeadlineExceeded marks a caller-owned deadline.
	ErrorCodeDeadlineExceeded ErrorCode = "deadline_exceeded"
	// ErrorCodeBackendUnavailable marks a missing backend operation.
	ErrorCodeBackendUnavailable ErrorCode = "backend_unavailable"
	// ErrorCodeRecallFailed marks a Host recall failure.
	ErrorCodeRecallFailed ErrorCode = "recall_failed"
	// ErrorCodeWriteFailed marks a Host write failure.
	ErrorCodeWriteFailed ErrorCode = "write_failed"
	// ErrorCodeArchiveFailed marks a Host archive failure.
	ErrorCodeArchiveFailed ErrorCode = "archive_failed"
	// ErrorCodeConflict marks a revision compare-and-swap conflict.
	ErrorCodeConflict ErrorCode = "conflict"
	// ErrorCodeInvalidBackendResult marks a result that violates the portable contract.
	ErrorCodeInvalidBackendResult ErrorCode = "invalid_backend_result"
)

// ErrConflict is returned by Backend implementations when ExpectedRevision no
// longer matches the durable record.
var ErrConflict = errors.New("memory lifecycle revision conflict")

// Error is a display-safe typed memory-lifecycle error. Cause is available via
// errors.Unwrap but is never included in Error's display text.
type Error struct {
	Code  ErrorCode
	Cause error
}

// Error returns a stable display-safe message.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	switch e.Code {
	case ErrorCodeInvalidPolicy:
		return "memory lifecycle policy is invalid"
	case ErrorCodeInvalidRequest:
		return "memory lifecycle request is invalid"
	case ErrorCodeCanceled:
		return "memory lifecycle operation was canceled"
	case ErrorCodeDeadlineExceeded:
		return "memory lifecycle operation deadline was exceeded"
	case ErrorCodeBackendUnavailable:
		return "memory lifecycle backend is unavailable"
	case ErrorCodeRecallFailed:
		return "memory recall failed"
	case ErrorCodeWriteFailed:
		return "memory write failed"
	case ErrorCodeArchiveFailed:
		return "memory archive failed"
	case ErrorCodeConflict:
		return "memory lifecycle revision conflict"
	case ErrorCodeInvalidBackendResult:
		return "memory lifecycle backend returned an invalid result"
	default:
		return "memory lifecycle operation failed"
	}
}

// Unwrap exposes the underlying cause for errors.Is/As without displaying it.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is compares memory-lifecycle errors by stable code.
func (e *Error) Is(target error) bool {
	other, ok := target.(*Error)
	return ok && e != nil && other != nil && e.Code == other.Code
}

// AsError returns the typed memory-lifecycle error when present.
func AsError(err error) (*Error, bool) {
	var typed *Error
	if !errors.As(err, &typed) {
		return nil, false
	}
	return typed, true
}

// RecordState is the portable lifecycle state of one durable memory record.
type RecordState string

const (
	// RecordStateActive is available to normal recall.
	RecordStateActive RecordState = "active"
	// RecordStateArchived is retained but excluded from normal recall.
	RecordStateArchived RecordState = "archived"
)

// Provenance identifies where a memory came from without prescribing a Host
// storage schema. SourceKind and SourceRef are required for every write.
type Provenance struct {
	SourceKind   string
	SourceRef    string
	SessionID    string
	RunID        string
	ArtifactRefs []string
	EvidenceRefs []string
	ObservedAt   time.Time
}

// Record is one Host-stored memory snapshot. ScopeRef is an opaque Host-owned
// visibility boundary, not a model-controlled path or tenant selector.
type Record struct {
	ID         string
	ScopeRef   string
	Content    string
	State      RecordState
	Revision   uint64
	Provenance Provenance
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// RecallRequest asks a Host backend for an already-authorized scope. Empty
// States means active records only.
type RecallRequest struct {
	ScopeRef string
	Query    string
	Limit    int
	States   []RecordState
}

// RecallResult preserves backend order; Coordinator never reranks records.
type RecallResult struct {
	Records []Record
}

// WriteRequest creates a record when RecordID and ExpectedRevision are both
// empty/zero, or updates one active record through revision CAS otherwise.
type WriteRequest struct {
	ScopeRef         string
	RecordID         string
	Content          string
	Provenance       Provenance
	IdempotencyKey   string
	ExpectedRevision uint64
}

// WriteResult is the authoritative backend readback for one write.
type WriteResult struct {
	Record           Record
	Created          bool
	IdempotentReplay bool
}

// ArchiveRequest archives one active record through revision CAS.
type ArchiveRequest struct {
	ScopeRef         string
	RecordID         string
	IdempotencyKey   string
	ExpectedRevision uint64
}

// ArchiveResult is the authoritative backend readback for one archive.
type ArchiveResult struct {
	Record           Record
	IdempotentReplay bool
}

// Backend is the Host-owned memory data plane. Implementations own concrete
// stores, ranking, visibility, retention and atomic persistence.
type Backend interface {
	Recall(context.Context, RecallRequest) (RecallResult, error)
	Write(context.Context, WriteRequest) (WriteResult, error)
	Archive(context.Context, ArchiveRequest) (ArchiveResult, error)
}

// BackendFuncs adapts Host functions to Backend.
type BackendFuncs struct {
	RecallFunc  func(context.Context, RecallRequest) (RecallResult, error)
	WriteFunc   func(context.Context, WriteRequest) (WriteResult, error)
	ArchiveFunc func(context.Context, ArchiveRequest) (ArchiveResult, error)
}

// Recall delegates to RecallFunc and fails closed when unavailable.
func (b BackendFuncs) Recall(ctx context.Context, request RecallRequest) (RecallResult, error) {
	if b.RecallFunc == nil {
		return RecallResult{}, &Error{Code: ErrorCodeBackendUnavailable}
	}
	return b.RecallFunc(ctx, request)
}

// Write delegates to WriteFunc and fails closed when unavailable.
func (b BackendFuncs) Write(ctx context.Context, request WriteRequest) (WriteResult, error) {
	if b.WriteFunc == nil {
		return WriteResult{}, &Error{Code: ErrorCodeBackendUnavailable}
	}
	return b.WriteFunc(ctx, request)
}

// Archive delegates to ArchiveFunc and fails closed when unavailable.
func (b BackendFuncs) Archive(ctx context.Context, request ArchiveRequest) (ArchiveResult, error) {
	if b.ArchiveFunc == nil {
		return ArchiveResult{}, &Error{Code: ErrorCodeBackendUnavailable}
	}
	return b.ArchiveFunc(ctx, request)
}

// Policy supplies explicit portable safety bounds.
type Policy struct {
	MaxRecallLimit    int
	MaxContentBytes   int
	MaxReferenceCount int
}

// Coordinator performs stateless memory lifecycle coordination.
type Coordinator struct {
	Policy  Policy
	Backend Backend
}

// Recall validates and bounds a recall request, calls Backend exactly once and
// returns a defensive copy without changing backend order.
func (c Coordinator) Recall(ctx context.Context, request RecallRequest) (RecallResult, error) {
	if err := c.validateCall(ctx); err != nil {
		return RecallResult{}, err
	}
	request.ScopeRef = strings.TrimSpace(request.ScopeRef)
	request.Query = strings.TrimSpace(request.Query)
	if request.ScopeRef == "" || request.Query == "" || request.Limit < 0 {
		return RecallResult{}, &Error{Code: ErrorCodeInvalidRequest}
	}
	if request.Limit == 0 || request.Limit > c.Policy.MaxRecallLimit {
		request.Limit = c.Policy.MaxRecallLimit
	}
	if len(request.States) == 0 {
		request.States = []RecordState{RecordStateActive}
	} else {
		request.States = slices.Clone(request.States)
		for _, state := range request.States {
			if !validState(state) {
				return RecallResult{}, &Error{Code: ErrorCodeInvalidRequest}
			}
		}
	}
	result, err := c.Backend.Recall(ctx, request)
	if err != nil {
		return RecallResult{}, backendError(ctx, ErrorCodeRecallFailed, err)
	}
	if err := contextError(ctx.Err()); err != nil {
		return RecallResult{}, err
	}
	if len(result.Records) > request.Limit {
		return RecallResult{}, &Error{Code: ErrorCodeInvalidBackendResult}
	}
	for _, record := range result.Records {
		if err := validateRecord(record, request.ScopeRef, c.Policy); err != nil || !slices.Contains(request.States, record.State) {
			return RecallResult{}, &Error{Code: ErrorCodeInvalidBackendResult, Cause: err}
		}
	}
	return cloneRecallResult(result), nil
}

// Write creates or updates one active record. Coordinator calls Backend exactly
// once and never retries an uncertain side effect.
func (c Coordinator) Write(ctx context.Context, request WriteRequest) (WriteResult, error) {
	if err := c.validateCall(ctx); err != nil {
		return WriteResult{}, err
	}
	normalizeWriteRequest(&request)
	if err := validateWriteRequest(request, c.Policy); err != nil {
		return WriteResult{}, err
	}
	result, err := c.Backend.Write(ctx, cloneWriteRequest(request))
	if err != nil {
		return WriteResult{}, backendError(ctx, ErrorCodeWriteFailed, err)
	}
	if err := contextError(ctx.Err()); err != nil {
		return WriteResult{}, err
	}
	if err := validateWriteResult(request, result, c.Policy); err != nil {
		return WriteResult{}, &Error{Code: ErrorCodeInvalidBackendResult, Cause: err}
	}
	result.Record = cloneRecord(result.Record)
	return result, nil
}

// Archive moves one active record to archived through revision CAS. Coordinator
// calls Backend exactly once and never deletes content.
func (c Coordinator) Archive(ctx context.Context, request ArchiveRequest) (ArchiveResult, error) {
	if err := c.validateCall(ctx); err != nil {
		return ArchiveResult{}, err
	}
	request.ScopeRef = strings.TrimSpace(request.ScopeRef)
	request.RecordID = strings.TrimSpace(request.RecordID)
	request.IdempotencyKey = strings.TrimSpace(request.IdempotencyKey)
	if request.ScopeRef == "" || request.RecordID == "" || request.IdempotencyKey == "" || request.ExpectedRevision == 0 {
		return ArchiveResult{}, &Error{Code: ErrorCodeInvalidRequest}
	}
	result, err := c.Backend.Archive(ctx, request)
	if err != nil {
		return ArchiveResult{}, backendError(ctx, ErrorCodeArchiveFailed, err)
	}
	if err := contextError(ctx.Err()); err != nil {
		return ArchiveResult{}, err
	}
	if err := validateRecord(result.Record, request.ScopeRef, c.Policy); err != nil ||
		result.Record.ID != request.RecordID || result.Record.State != RecordStateArchived ||
		result.Record.Revision != request.ExpectedRevision+1 {
		return ArchiveResult{}, &Error{Code: ErrorCodeInvalidBackendResult, Cause: err}
	}
	result.Record = cloneRecord(result.Record)
	return result, nil
}

func (c Coordinator) validateCall(ctx context.Context) error {
	if ctx == nil {
		return &Error{Code: ErrorCodeInvalidRequest}
	}
	if err := contextError(ctx.Err()); err != nil {
		return err
	}
	if c.Policy.MaxRecallLimit <= 0 || c.Policy.MaxContentBytes <= 0 || c.Policy.MaxReferenceCount < 0 {
		return &Error{Code: ErrorCodeInvalidPolicy}
	}
	if c.Backend == nil {
		return &Error{Code: ErrorCodeBackendUnavailable}
	}
	return nil
}

func validateWriteRequest(request WriteRequest, policy Policy) error {
	if request.ScopeRef == "" || request.IdempotencyKey == "" || strings.TrimSpace(request.Content) == "" || len(request.Content) > policy.MaxContentBytes {
		return &Error{Code: ErrorCodeInvalidRequest}
	}
	if (request.RecordID == "") != (request.ExpectedRevision == 0) {
		return &Error{Code: ErrorCodeInvalidRequest}
	}
	if err := validateProvenance(request.Provenance, policy); err != nil {
		return &Error{Code: ErrorCodeInvalidRequest, Cause: err}
	}
	return nil
}

func validateWriteResult(request WriteRequest, result WriteResult, policy Policy) error {
	if err := validateRecord(result.Record, request.ScopeRef, policy); err != nil {
		return err
	}
	if result.Record.State != RecordStateActive || result.Record.Content != request.Content ||
		result.Record.Revision != request.ExpectedRevision+1 || !equalProvenance(result.Record.Provenance, request.Provenance) {
		return errors.New("write readback does not match request")
	}
	if request.RecordID != "" && result.Record.ID != request.RecordID {
		return errors.New("write readback changed record identity")
	}
	if request.RecordID == "" && !result.Created && !result.IdempotentReplay {
		return errors.New("create readback is not marked created")
	}
	if request.RecordID != "" && result.Created {
		return errors.New("update readback is marked created")
	}
	return nil
}

func validateRecord(record Record, scopeRef string, policy Policy) error {
	if strings.TrimSpace(record.ID) == "" || record.ScopeRef != scopeRef || strings.TrimSpace(record.Content) == "" ||
		len(record.Content) > policy.MaxContentBytes || !validState(record.State) || record.Revision == 0 {
		return errors.New("record invariant failed")
	}
	if record.CreatedAt.IsZero() || record.UpdatedAt.IsZero() || record.UpdatedAt.Before(record.CreatedAt) {
		return errors.New("record timestamps are invalid")
	}
	return validateProvenance(record.Provenance, policy)
}

func validateProvenance(provenance Provenance, policy Policy) error {
	if strings.TrimSpace(provenance.SourceKind) == "" || strings.TrimSpace(provenance.SourceRef) == "" {
		return errors.New("provenance source is required")
	}
	if len(provenance.ArtifactRefs)+len(provenance.EvidenceRefs) > policy.MaxReferenceCount {
		return errors.New("provenance reference limit exceeded")
	}
	for _, ref := range append(slices.Clone(provenance.ArtifactRefs), provenance.EvidenceRefs...) {
		if strings.TrimSpace(ref) == "" {
			return errors.New("provenance reference is empty")
		}
	}
	return nil
}

func validState(state RecordState) bool {
	return state == RecordStateActive || state == RecordStateArchived
}

func backendError(ctx context.Context, operation ErrorCode, err error) error {
	if contextErr := contextError(ctx.Err()); contextErr != nil {
		return contextErr
	}
	if errors.Is(err, ErrConflict) {
		return &Error{Code: ErrorCodeConflict, Cause: err}
	}
	if typed, ok := AsError(err); ok && typed.Code == ErrorCodeBackendUnavailable {
		return typed
	}
	return &Error{Code: operation, Cause: err}
}

func contextError(err error) error {
	if err == nil {
		return nil
	}
	code := ErrorCodeCanceled
	if errors.Is(err, context.DeadlineExceeded) {
		code = ErrorCodeDeadlineExceeded
	}
	return &Error{Code: code, Cause: err}
}

func normalizeWriteRequest(request *WriteRequest) {
	request.ScopeRef = strings.TrimSpace(request.ScopeRef)
	request.RecordID = strings.TrimSpace(request.RecordID)
	request.IdempotencyKey = strings.TrimSpace(request.IdempotencyKey)
	request.Provenance.SourceKind = strings.TrimSpace(request.Provenance.SourceKind)
	request.Provenance.SourceRef = strings.TrimSpace(request.Provenance.SourceRef)
	request.Provenance.SessionID = strings.TrimSpace(request.Provenance.SessionID)
	request.Provenance.RunID = strings.TrimSpace(request.Provenance.RunID)
}

func equalProvenance(left, right Provenance) bool {
	return left.SourceKind == right.SourceKind && left.SourceRef == right.SourceRef &&
		left.SessionID == right.SessionID && left.RunID == right.RunID && left.ObservedAt.Equal(right.ObservedAt) &&
		slices.Equal(left.ArtifactRefs, right.ArtifactRefs) && slices.Equal(left.EvidenceRefs, right.EvidenceRefs)
}

func cloneRecallResult(result RecallResult) RecallResult {
	out := RecallResult{Records: make([]Record, len(result.Records))}
	for index, record := range result.Records {
		out.Records[index] = cloneRecord(record)
	}
	return out
}

func cloneWriteRequest(request WriteRequest) WriteRequest {
	request.Provenance = cloneProvenance(request.Provenance)
	return request
}

func cloneRecord(record Record) Record {
	record.Provenance = cloneProvenance(record.Provenance)
	return record
}

func cloneProvenance(provenance Provenance) Provenance {
	provenance.ArtifactRefs = slices.Clone(provenance.ArtifactRefs)
	provenance.EvidenceRefs = slices.Clone(provenance.EvidenceRefs)
	return provenance
}
