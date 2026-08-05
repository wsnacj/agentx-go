package memory

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCoordinatorRecallBoundsPreservesOrderAndDefensiveCopies(t *testing.T) {
	now := time.Unix(10, 0).UTC()
	backendRecords := []Record{
		record("one", "scope", "first", 1, RecordStateActive, now),
		record("two", "scope", "second", 2, RecordStateActive, now),
	}
	var received RecallRequest
	coordinator := testCoordinator(BackendFuncs{RecallFunc: func(_ context.Context, request RecallRequest) (RecallResult, error) {
		received = request
		return RecallResult{Records: backendRecords}, nil
	}})
	result, err := coordinator.Recall(context.Background(), RecallRequest{ScopeRef: " scope ", Query: " query ", Limit: 99})
	if err != nil {
		t.Fatal(err)
	}
	if received.ScopeRef != "scope" || received.Query != "query" || received.Limit != 2 || len(received.States) != 1 || received.States[0] != RecordStateActive {
		t.Fatalf("received = %#v", received)
	}
	if result.Records[0].ID != "one" || result.Records[1].ID != "two" {
		t.Fatalf("result order = %#v", result.Records)
	}
	result.Records[0].Provenance.ArtifactRefs[0] = "changed"
	if backendRecords[0].Provenance.ArtifactRefs[0] != "artifact-1" {
		t.Fatal("result aliases backend provenance")
	}
}

func TestCoordinatorWriteAndArchiveEnforceReadback(t *testing.T) {
	now := time.Unix(10, 0).UTC()
	writes := 0
	archives := 0
	coordinator := testCoordinator(BackendFuncs{
		WriteFunc: func(_ context.Context, request WriteRequest) (WriteResult, error) {
			writes++
			return WriteResult{Record: Record{
				ID: "memory-1", ScopeRef: request.ScopeRef, Content: request.Content,
				State: RecordStateActive, Revision: request.ExpectedRevision + 1,
				Provenance: request.Provenance, CreatedAt: now, UpdatedAt: now,
			}, Created: request.RecordID == ""}, nil
		},
		ArchiveFunc: func(_ context.Context, request ArchiveRequest) (ArchiveResult, error) {
			archives++
			return ArchiveResult{Record: record(request.RecordID, request.ScopeRef, "remember", request.ExpectedRevision+1, RecordStateArchived, now)}, nil
		},
	})
	provenance := testProvenance(now)
	written, err := coordinator.Write(context.Background(), WriteRequest{
		ScopeRef: "scope", Content: "remember", Provenance: provenance, IdempotencyKey: "write-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if writes != 1 || written.Record.Revision != 1 || !written.Created {
		t.Fatalf("write result = %#v, calls=%d", written, writes)
	}
	archived, err := coordinator.Archive(context.Background(), ArchiveRequest{
		ScopeRef: "scope", RecordID: "memory-1", IdempotencyKey: "archive-1", ExpectedRevision: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if archives != 1 || archived.Record.State != RecordStateArchived || archived.Record.Revision != 2 {
		t.Fatalf("archive result = %#v, calls=%d", archived, archives)
	}
}

func TestCoordinatorRejectsInvalidInputBeforeBackend(t *testing.T) {
	calls := 0
	coordinator := testCoordinator(BackendFuncs{WriteFunc: func(context.Context, WriteRequest) (WriteResult, error) {
		calls++
		return WriteResult{}, nil
	}})
	_, err := coordinator.Write(context.Background(), WriteRequest{ScopeRef: "scope", Content: strings.Repeat("x", 65), IdempotencyKey: "key"})
	if !errors.Is(err, &Error{Code: ErrorCodeInvalidRequest}) || calls != 0 {
		t.Fatalf("err=%v, calls=%d", err, calls)
	}
}

func TestCoordinatorMapsCancellationConflictAndSafeDisplay(t *testing.T) {
	secret := errors.New("backend included private record content")
	coordinator := testCoordinator(BackendFuncs{
		RecallFunc: func(context.Context, RecallRequest) (RecallResult, error) { return RecallResult{}, secret },
		WriteFunc:  func(context.Context, WriteRequest) (WriteResult, error) { return WriteResult{}, ErrConflict },
	})
	_, err := coordinator.Recall(context.Background(), RecallRequest{ScopeRef: "scope", Query: "secret"})
	if !errors.Is(err, &Error{Code: ErrorCodeRecallFailed}) || strings.Contains(err.Error(), "private") {
		t.Fatalf("unsafe recall error = %v", err)
	}
	_, err = coordinator.Write(context.Background(), WriteRequest{
		ScopeRef: "scope", Content: "content", Provenance: testProvenance(time.Time{}), IdempotencyKey: "key",
	})
	if !errors.Is(err, &Error{Code: ErrorCodeConflict}) || !errors.Is(err, ErrConflict) {
		t.Fatalf("conflict error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = coordinator.Recall(ctx, RecallRequest{ScopeRef: "scope", Query: "query"})
	if !errors.Is(err, &Error{Code: ErrorCodeCanceled}) || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error = %v", err)
	}
}

func TestCoordinatorRejectsInvalidBackendResult(t *testing.T) {
	now := time.Unix(10, 0).UTC()
	coordinator := testCoordinator(BackendFuncs{RecallFunc: func(context.Context, RecallRequest) (RecallResult, error) {
		return RecallResult{Records: []Record{record("one", "another-scope", "content", 1, RecordStateActive, now)}}, nil
	}})
	_, err := coordinator.Recall(context.Background(), RecallRequest{ScopeRef: "scope", Query: "query"})
	if !errors.Is(err, &Error{Code: ErrorCodeInvalidBackendResult}) {
		t.Fatalf("error = %v", err)
	}
}

func testCoordinator(backend Backend) Coordinator {
	return Coordinator{Policy: Policy{MaxRecallLimit: 2, MaxContentBytes: 64, MaxReferenceCount: 4}, Backend: backend}
}

func record(id, scope, content string, revision uint64, state RecordState, now time.Time) Record {
	return Record{
		ID: id, ScopeRef: scope, Content: content, State: state, Revision: revision,
		Provenance: testProvenance(now), CreatedAt: now, UpdatedAt: now,
	}
}

func testProvenance(now time.Time) Provenance {
	return Provenance{SourceKind: "session", SourceRef: "session-1", ArtifactRefs: []string{"artifact-1"}, ObservedAt: now}
}
