package browserruntime

import "testing"

func TestBuildBrowserSnapshotFreshnessMatchedSnapshotRef(t *testing.T) {
	freshness := BuildBrowserSnapshotFreshness(BrowserSnapshotFreshnessRequest{
		Action:         "click",
		TargetKind:     "ref",
		Target:         "elem1:submit",
		SnapshotEpoch:  "snapshot-42",
		SnapshotSource: BrowserSnapshotFreshnessSourceSnapshotRef,
		ResolverOutcome: &BrowserElementResolverOutcome{
			Status:      "matched",
			MatchedKind: "native_ref",
		},
	})
	if freshness == nil {
		t.Fatal("expected snapshot freshness")
	}
	if freshness.State != BrowserSnapshotFreshnessStateFresh ||
		freshness.Source != BrowserSnapshotFreshnessSourceSnapshotRef ||
		freshness.SnapshotEpoch != "snapshot-42" ||
		freshness.TargetKind != "ref" ||
		freshness.Target != "elem1:submit" ||
		freshness.RefreshRecommended {
		t.Fatalf("unexpected freshness: %#v", freshness)
	}
}

func TestBuildBrowserSnapshotFreshnessRecommendsSnapshotForPageBinding(t *testing.T) {
	freshness := BuildBrowserSnapshotFreshness(BrowserSnapshotFreshnessRequest{
		Action:     "type",
		TargetKind: "ref",
		Target:     "elem1:email",
		PageURL:    "https://example.com/form",
		PageTitle:  "Example Form",
		ResolverOutcome: &BrowserElementResolverOutcome{
			Status:         "page_binding_blocked",
			BlockedBy:      "page_url",
			RecoveryAction: "browser action=snapshot",
		},
	})
	if freshness == nil {
		t.Fatal("expected snapshot freshness")
	}
	if freshness.State != BrowserSnapshotFreshnessStateStale ||
		freshness.Source != BrowserSnapshotFreshnessSourcePageBinding ||
		!freshness.RefreshRecommended ||
		freshness.RefreshReason != BrowserSnapshotRefreshReasonPageChanged ||
		freshness.RecoveryAction != "browser action=snapshot" ||
		freshness.NextStepAlias != "snapshot" {
		t.Fatalf("unexpected freshness: %#v", freshness)
	}
}

func TestBuildBrowserSnapshotFreshnessUsesRecoverySnapshot(t *testing.T) {
	freshness := BuildBrowserSnapshotFreshness(BrowserSnapshotFreshnessRequest{
		Action:                    "click",
		TargetKind:                "selector",
		Target:                    "button.primary",
		RecoverySnapshotAvailable: true,
		ResolverOutcome: &BrowserElementResolverOutcome{
			Status:         "unresolved",
			BlockedBy:      "multiple_candidates",
			RecoveryAction: "browser action=snapshot",
		},
	})
	if freshness == nil {
		t.Fatal("expected snapshot freshness")
	}
	if freshness.State != BrowserSnapshotFreshnessStateFresh ||
		freshness.Source != BrowserSnapshotFreshnessSourceRecoverySnapshot ||
		freshness.RefreshRecommended ||
		freshness.RefreshReason != BrowserSnapshotRefreshReasonAmbiguousTarget ||
		freshness.RecoveryAction != BrowserSnapshotRecoveryUseRecoverySnapshot ||
		freshness.NextStepAlias != "snapshot" ||
		len(freshness.Notes) != 1 ||
		freshness.Notes[0] != "recovery_snapshot_available" {
		t.Fatalf("unexpected freshness: %#v", freshness)
	}
}

func TestBuildBrowserSnapshotFreshnessMarksMissingSnapshot(t *testing.T) {
	freshness := BuildBrowserSnapshotFreshness(BrowserSnapshotFreshnessRequest{
		Action:     "screenshot",
		TargetKind: "selector",
		Target:     "#chart",
	})
	if freshness == nil {
		t.Fatal("expected snapshot freshness")
	}
	if freshness.State != BrowserSnapshotFreshnessStateMissing ||
		!freshness.RefreshRecommended ||
		freshness.RefreshReason != BrowserSnapshotRefreshReasonSnapshotUnavailable ||
		freshness.RecoveryAction != "browser action=snapshot" ||
		freshness.NextStepAlias != "snapshot" {
		t.Fatalf("unexpected freshness: %#v", freshness)
	}
}
