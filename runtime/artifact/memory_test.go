package artifact

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestMemoryRegistryIndexesAndOrdersRecords(t *testing.T) {
	reg := NewMemoryRegistry()
	for _, record := range []Record{
		{ArtifactID: "artifact-2", RunID: "run-1", SessionID: "session-1", Kind: "download", CreatedAt: 20},
		{ArtifactID: "artifact-1", RunID: "run-1", SessionID: "session-1", Kind: "screenshot", CreatedAt: 10},
		{ArtifactID: "artifact-0", RunID: "run-1", SessionID: "session-1", Kind: "trace", CreatedAt: 10},
	} {
		if err := reg.Register(context.Background(), record); err != nil {
			t.Fatalf("Register(%s): %v", record.ArtifactID, err)
		}
	}

	runRecords, err := reg.ListByRun(context.Background(), " run-1 ")
	if err != nil {
		t.Fatalf("ListByRun(): %v", err)
	}
	if got, want := recordIDs(runRecords), []string{"artifact-0", "artifact-1", "artifact-2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("run record order = %v, want %v", got, want)
	}

	sessionRecords, err := reg.ListBySession(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("ListBySession(): %v", err)
	}
	if got, want := recordIDs(sessionRecords), []string{"artifact-0", "artifact-1", "artifact-2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("session record order = %v, want %v", got, want)
	}
}

func TestMemoryRegistryMergesWithoutChangingFirstIndexes(t *testing.T) {
	reg := NewMemoryRegistry()
	if err := reg.Register(context.Background(), Record{
		ArtifactID:   " artifact-1 ",
		RunID:        " run-first ",
		SessionID:    " session-first ",
		ToolName:     " browser_screenshot ",
		Producer:     " tool:browser_screenshot ",
		Kind:         " screenshot ",
		Bytes:        10,
		Summary:      " evidence ",
		Labels:       []string{" kind:screenshot ", "tool:browser_screenshot", "kind:screenshot"},
		MetadataJSON: " {\"first\":true} ",
		CreatedAt:    20,
	}); err != nil {
		t.Fatalf("Register(first): %v", err)
	}
	if err := reg.Register(context.Background(), Record{
		ArtifactID:   "artifact-1",
		RunID:        "run-second",
		SessionID:    "session-second",
		MIMEType:     " image/png ",
		Format:       " png ",
		Bytes:        20,
		Summary:      "replacement",
		Labels:       []string{"format:png", "kind:screenshot"},
		MetadataJSON: "{\"second\":true}",
		CreatedAt:    10,
	}); err != nil {
		t.Fatalf("Register(second): %v", err)
	}

	records, err := reg.ListByRun(context.Background(), "run-first")
	if err != nil || len(records) != 1 {
		t.Fatalf("ListByRun(first) = %#v, %v", records, err)
	}
	record := records[0]
	if record.ArtifactID != "artifact-1" ||
		record.RunID != "run-first" ||
		record.SessionID != "session-first" ||
		record.ToolName != "browser_screenshot" ||
		record.Producer != "tool:browser_screenshot" ||
		record.Kind != "screenshot" ||
		record.MIMEType != "image/png" ||
		record.Format != "png" ||
		record.Bytes != 20 ||
		record.Summary != "evidence" ||
		record.MetadataJSON != `{"first":true}` ||
		record.CreatedAt != 10 {
		t.Fatalf("merged record = %#v", record)
	}
	if got, want := record.Labels, []string{"kind:screenshot", "tool:browser_screenshot", "format:png"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("merged labels = %v, want %v", got, want)
	}
	if records, err := reg.ListByRun(context.Background(), "run-second"); err != nil || records != nil {
		t.Fatalf("second run index = %#v, %v; want nil", records, err)
	}
	if records, err := reg.ListBySession(context.Background(), "session-second"); err != nil || records != nil {
		t.Fatalf("second session index = %#v, %v; want nil", records, err)
	}
}

func TestMemoryRegistryDoesNotBackfillIndexesForExistingRecord(t *testing.T) {
	reg := NewMemoryRegistry()
	if err := reg.Register(context.Background(), Record{ArtifactID: "artifact-1"}); err != nil {
		t.Fatalf("Register(first): %v", err)
	}
	if err := reg.Register(context.Background(), Record{
		ArtifactID: "artifact-1",
		RunID:      "late-run",
		SessionID:  "late-session",
	}); err != nil {
		t.Fatalf("Register(second): %v", err)
	}
	if records, err := reg.ListByRun(context.Background(), "late-run"); err != nil || records != nil {
		t.Fatalf("late run index = %#v, %v; want nil", records, err)
	}
	if records, err := reg.ListBySession(context.Background(), "late-session"); err != nil || records != nil {
		t.Fatalf("late session index = %#v, %v; want nil", records, err)
	}
}

func TestMemoryRegistryMergesFiltersAndOrdersLinks(t *testing.T) {
	reg := NewMemoryRegistry()
	links := []Link{
		{SourceArtifactID: " source-b ", TargetArtifactID: "target", Relation: " contains ", CreatedAt: 10},
		{SourceArtifactID: "source-a", TargetArtifactID: "target", Relation: "contains", CreatedAt: 10},
		{SourceArtifactID: "target", TargetArtifactID: "child", Relation: "derived", CreatedAt: 20},
		{SourceArtifactID: "source-a", TargetArtifactID: "target", Relation: "contains", MetadataJSON: ` {"origin":"first"} `, CreatedAt: 5},
		{SourceArtifactID: "source-a", TargetArtifactID: "target", Relation: "contains", MetadataJSON: `{"origin":"ignored"}`, CreatedAt: 8},
		{SourceArtifactID: " ", TargetArtifactID: "ignored", Relation: "contains", CreatedAt: 1},
	}
	for _, link := range links {
		if err := reg.Link(context.Background(), link); err != nil {
			t.Fatalf("Link(%#v): %v", link, err)
		}
	}

	all, err := reg.ListLinks(context.Background(), LinkFilter{})
	if err != nil {
		t.Fatalf("ListLinks(all): %v", err)
	}
	if len(all) != 3 ||
		all[0].SourceArtifactID != "source-a" || all[0].CreatedAt != 5 || all[0].MetadataJSON != `{"origin":"first"}` ||
		all[1].SourceArtifactID != "source-b" ||
		all[2].Relation != "derived" {
		t.Fatalf("ordered links = %#v", all)
	}

	outbound, err := reg.ListLinks(context.Background(), LinkFilter{
		ArtifactID: " source-a ",
		Relation:   " contains ",
		Direction:  " OUTBOUND ",
	})
	if err != nil || len(outbound) != 1 || outbound[0].TargetArtifactID != "target" {
		t.Fatalf("outbound links = %#v, %v", outbound, err)
	}
	inbound, err := reg.ListLinks(context.Background(), LinkFilter{ArtifactID: "target", Direction: "inbound"})
	if err != nil || len(inbound) != 2 {
		t.Fatalf("inbound links = %#v, %v", inbound, err)
	}
	either, err := reg.ListLinks(context.Background(), LinkFilter{ArtifactID: "target", Direction: "unexpected"})
	if err != nil || len(either) != 3 {
		t.Fatalf("bidirectional links = %#v, %v", either, err)
	}
}

func TestNilMemoryRegistryIsCompatibilityNoop(t *testing.T) {
	var reg *MemoryRegistry
	if err := reg.Register(nil, Record{ArtifactID: "artifact"}); err != nil {
		t.Fatalf("nil Register(): %v", err)
	}
	if err := reg.Link(nil, Link{SourceArtifactID: "a", TargetArtifactID: "b", Relation: "contains"}); err != nil {
		t.Fatalf("nil Link(): %v", err)
	}
	if records, err := reg.ListByRun(nil, "run"); err != nil || records != nil {
		t.Fatalf("nil ListByRun() = %#v, %v", records, err)
	}
	if records, err := reg.ListBySession(nil, "session"); err != nil || records != nil {
		t.Fatalf("nil ListBySession() = %#v, %v", records, err)
	}
	if links, err := reg.ListLinks(nil, LinkFilter{}); err != nil || links != nil {
		t.Fatalf("nil ListLinks() = %#v, %v", links, err)
	}
}

func TestMemoryRegistryConcurrentMethodCalls(t *testing.T) {
	reg := NewMemoryRegistry()
	const count = 64
	errs := make(chan error, count*2)
	var wg sync.WaitGroup
	for index := 0; index < count; index++ {
		index := index
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := fmt.Sprintf("artifact-%03d", index)
			if err := reg.Register(context.Background(), Record{
				ArtifactID: id,
				RunID:      "run",
				SessionID:  "session",
				CreatedAt:  int64(count - index),
			}); err != nil {
				errs <- err
			}
			if err := reg.Link(context.Background(), Link{
				SourceArtifactID: "root",
				TargetArtifactID: id,
				Relation:         "contains",
				CreatedAt:        int64(index),
			}); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent operation: %v", err)
	}

	records, err := reg.ListByRun(context.Background(), "run")
	if err != nil || len(records) != count {
		t.Fatalf("concurrent records = %d, %v", len(records), err)
	}
	for index := 1; index < len(records); index++ {
		if records[index-1].CreatedAt > records[index].CreatedAt {
			t.Fatalf("records are not ordered at %d: %#v", index, records)
		}
	}
	links, err := reg.ListLinks(context.Background(), LinkFilter{ArtifactID: "root", Direction: "outbound"})
	if err != nil || len(links) != count {
		t.Fatalf("concurrent links = %d, %v", len(links), err)
	}
}

func recordIDs(records []Record) []string {
	if len(records) == 0 {
		return nil
	}
	out := make([]string, 0, len(records))
	for _, record := range records {
		out = append(out, record.ArtifactID)
	}
	return out
}
