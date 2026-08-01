package artifact_test

import (
	"context"
	"testing"

	artifact "github.com/wsnacj/agentx-go/runtime/artifact"
)

var (
	_ artifact.AuthoringRegistry = (*artifact.MemoryRegistry)(nil)
	_ artifact.QueryRegistry     = (*artifact.MemoryRegistry)(nil)
	_ artifact.Registry          = (*artifact.MemoryRegistry)(nil)
	_ artifact.BlobStore         = blobStoreFunc(nil)
)

func TestExternalPackageUsesRegistryAndBlobStoreContracts(t *testing.T) {
	registry := artifact.NewMemoryRegistry()
	if err := registry.Register(context.Background(), artifact.Record{
		ArtifactID: "artifact-1",
		RunID:      "run-1",
		SessionID:  "session-1",
		Kind:       "report",
		CreatedAt:  10,
	}); err != nil {
		t.Fatalf("Register(): %v", err)
	}
	if err := registry.Link(context.Background(), artifact.Link{
		SourceArtifactID: "manifest-1",
		TargetArtifactID: "artifact-1",
		Relation:         "contains",
		CreatedAt:        11,
	}); err != nil {
		t.Fatalf("Link(): %v", err)
	}
	records, err := registry.ListByRun(context.Background(), "run-1")
	if err != nil || len(records) != 1 || records[0].ArtifactID != "artifact-1" {
		t.Fatalf("ListByRun() = %#v, %v", records, err)
	}
	links, err := registry.ListLinks(context.Background(), artifact.LinkFilter{
		ArtifactID: "manifest-1",
		Direction:  "outbound",
	})
	if err != nil || len(links) != 1 || links[0].TargetArtifactID != "artifact-1" {
		t.Fatalf("ListLinks() = %#v, %v", links, err)
	}

	store := blobStoreFunc(func(_ context.Context, input artifact.BlobPutInput) (artifact.BlobRef, error) {
		return artifact.BlobRef{
			StorageRef: "host-owned",
			Bytes:      int64(len(input.Data)),
		}, nil
	})
	ref, err := store.Put(context.Background(), artifact.BlobPutInput{Data: []byte("payload")})
	if err != nil || ref.StorageRef != "host-owned" || ref.Bytes != 7 {
		t.Fatalf("BlobStore.Put() = %#v, %v", ref, err)
	}
}

type blobStoreFunc func(context.Context, artifact.BlobPutInput) (artifact.BlobRef, error)

func (f blobStoreFunc) Put(ctx context.Context, input artifact.BlobPutInput) (artifact.BlobRef, error) {
	return f(ctx, input)
}
