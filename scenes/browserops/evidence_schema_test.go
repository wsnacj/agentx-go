package browserops

import "testing"

func TestBuildBrowserEvidenceBundleFromStateCapturesCanonicalEvidence(t *testing.T) {
	bundle := BuildBrowserEvidenceBundleFromState(map[string]any{
		"case": map[string]any{
			"input": map[string]any{
				"target_url": "https://example.com/forms",
			},
		},
		"review": map[string]any{
			"snapshot":           "heading Submitted\nstatus Success",
			"snapshot_path":      "data/tmp/browserops/snapshot.json",
			"final_url":          "https://example.com/forms/submitted",
			"evidence_path":      "data/tmp/browserops/success.png",
			"evidence_final_url": "https://example.com/forms/submitted",
			"trace_path":         "data/tmp/browserops/action-trace.zip",
			"task_plan": map[string]any{
				"task_id":   "browser-task-1",
				"goal":      "verify form submission",
				"case_type": "browser.verify_page_state",
				"ready":     true,
				"subtasks": []any{
					map[string]any{"id": "open", "kind": "open_target", "status": "observed", "required": true, "evidence_kind": BrowserEvidenceKindFinalURL},
					map[string]any{"id": "snapshot", "kind": "capture_page_snapshot", "status": "observed", "required": true, "evidence_kind": BrowserEvidenceKindPageSnapshot},
				},
			},
		},
		"download": map[string]any{
			"path":               "data/tmp/browserops/report.csv",
			"bytes":              128,
			"content_type":       "text/csv",
			"suggested_filename": "report.csv",
			"mode":               "download",
		},
	})

	if bundle.TargetURL != "https://example.com/forms" {
		t.Fatalf("expected target url from nested case input, got %#v", bundle.TargetURL)
	}
	if bundle.FinalURL == nil || bundle.FinalURL.URL != "https://example.com/forms/submitted" {
		t.Fatalf("expected final url evidence, got %#v", bundle.FinalURL)
	}
	if bundle.PageSnapshot == nil || bundle.PageSnapshot.Format != "aria" || !bundle.PageSnapshot.Ready {
		t.Fatalf("expected ready aria snapshot evidence, got %#v", bundle.PageSnapshot)
	}
	if bundle.Screenshot == nil || bundle.Screenshot.Artifact.Type != BrowserArtifactTypePageScreenshot {
		t.Fatalf("expected screenshot artifact evidence, got %#v", bundle.Screenshot)
	}
	if bundle.ActionTrace == nil || bundle.ActionTrace.Artifact.Type != BrowserArtifactTypeActionTrace {
		t.Fatalf("expected action trace artifact evidence, got %#v", bundle.ActionTrace)
	}
	if bundle.DownloadedFile == nil || bundle.DownloadedFile.Artifact.Type != BrowserArtifactTypeDownloadedFile || bundle.DownloadedFile.ByteSize != 128 {
		t.Fatalf("expected downloaded file artifact evidence, got %#v", bundle.DownloadedFile)
	}
	if bundle.TaskPlan == nil || len(bundle.TaskPlan.Subtasks) != 2 || bundle.TaskPlan.Subtasks[0].Kind != "open_target" {
		t.Fatalf("expected task plan evidence, got %#v", bundle.TaskPlan)
	}
	for _, artifactType := range []string{BrowserArtifactTypePageSnapshot, BrowserArtifactTypePageScreenshot, BrowserArtifactTypeActionTrace, BrowserArtifactTypeDownloadedFile} {
		if !browserEvidenceTestHasArtifactType(bundle.ArtifactRefs, artifactType) {
			t.Fatalf("expected artifact type %q in %#v", artifactType, bundle.ArtifactRefs)
		}
	}
}

func TestEvaluateBrowserEvidenceBundleRequiresReplayableEvidence(t *testing.T) {
	bundle := BrowserEvidenceBundle{
		TargetURL: "https://example.com/forms",
		FinalURL:  &BrowserFinalURLEvidence{URL: "https://example.com/forms/submitted", TargetURL: "https://example.com/forms"},
		PageSnapshot: &BrowserPageSnapshotEvidence{
			Text:   "heading Submitted\nstatus Success",
			Format: "aria",
		},
		Screenshot: &BrowserScreenshotEvidence{
			Path: "data/tmp/browserops/success.png",
			Artifact: BrowserArtifactRef{
				Type: BrowserArtifactTypePageScreenshot,
				Path: "data/tmp/browserops/success.png",
				Role: BrowserEvidenceKindScreenshot,
			},
		},
		ActionTrace: &BrowserActionTraceEvidence{
			Path: "data/tmp/browserops/action-trace.zip",
			Artifact: BrowserArtifactRef{
				Type: BrowserArtifactTypeActionTrace,
				Path: "data/tmp/browserops/action-trace.zip",
				Role: BrowserEvidenceKindActionTrace,
			},
		},
		TaskPlan: &BrowserTaskPlanEvidence{
			TaskID: "browser-task-1",
			Ready:  true,
			Subtasks: []BrowserSubtaskObservationEvidence{
				{ID: "open", Kind: "open_target", Status: "observed", Required: true, EvidenceKind: BrowserEvidenceKindFinalURL},
				{ID: "snapshot", Kind: "capture_page_snapshot", Status: "observed", Required: true, EvidenceKind: BrowserEvidenceKindPageSnapshot},
				{ID: "screenshot", Kind: "capture_submission_evidence", Status: "observed", Required: true, EvidenceKind: BrowserEvidenceKindScreenshot},
			},
		},
		ArtifactRefs: []BrowserArtifactRef{
			{Type: BrowserArtifactTypePageScreenshot, Path: "data/tmp/browserops/success.png", Role: BrowserEvidenceKindScreenshot},
			{Type: BrowserArtifactTypeActionTrace, Path: "data/tmp/browserops/action-trace.zip", Role: BrowserEvidenceKindActionTrace},
		},
	}

	result := EvaluateBrowserEvidenceBundle(bundle, BrowserEvidenceRequirements{
		RequiredSnapshotTerms: []string{"Submitted", "Success"},
		MinSnapshotChars:      20,
		RequireSnapshot:       true,
		RequireScreenshot:     true,
		RequireFinalURL:       true,
		RequireActionTrace:    true,
		RequireArtifactRefs:   true,
		RequireTaskPlan:       true,
		MinSubtasks:           3,
		RequiredSubtaskKinds:  []string{"open_target", "capture_page_snapshot", "capture_submission_evidence"},
	})
	if !result.Passed ||
		!result.SnapshotReady ||
		!result.ScreenshotReady ||
		!result.FinalURLReady ||
		!result.ActionTraceReady ||
		!result.ArtifactRefsReady ||
		!result.TaskPlanReady ||
		!result.FailureReasonsReady ||
		result.SubtaskCount != 3 ||
		result.Score != 1 {
		t.Fatalf("expected browser evidence bundle to pass, got %#v", result)
	}
	for _, expected := range []string{"snapshot_ready", "screenshot_ready", "final_url_ready", "action_trace_ready", "artifact_refs_ready", "task_plan_ready"} {
		if !containsBrowserFailureReason(result.Evidence, expected) {
			t.Fatalf("expected evidence %q in %#v", expected, result.Evidence)
		}
	}
}

func TestEvaluateBrowserEvidenceBundleReportsMissingEvidence(t *testing.T) {
	result := EvaluateBrowserEvidenceBundle(BrowserEvidenceBundle{
		TargetURL:      "https://example.com/forms",
		FailureReasons: []BrowserFailureReasonEvidence{{Code: "action_timeout", EvidenceKind: BrowserEvidenceKindFailureReason}},
	}, BrowserEvidenceRequirements{
		RequireSnapshot:     true,
		RequireScreenshot:   true,
		RequireFinalURL:     true,
		RequireActionTrace:  true,
		RequireArtifactRefs: true,
		RequireTaskPlan:     true,
		MinSubtasks:         2,
	})
	if result.Passed || result.EvidenceReady {
		t.Fatalf("expected missing evidence to fail, got %#v", result)
	}
	for _, reason := range []string{
		"snapshot_missing",
		"screenshot_missing",
		"final_url_missing",
		"action_trace_missing",
		"artifact_refs_missing",
		"task_plan_missing",
		"failure_reasons_present",
		"failure_reason:action_timeout",
	} {
		if !containsBrowserFailureReason(result.FailureReasons, reason) {
			t.Fatalf("expected failure reason %q in %#v", reason, result.FailureReasons)
		}
	}
}

func TestEvaluateBrowserEvidenceBundleTreatsBlockedTaskPlanAsNotReady(t *testing.T) {
	result := EvaluateBrowserEvidenceBundle(BrowserEvidenceBundle{
		TaskPlan: &BrowserTaskPlanEvidence{
			TaskID: "browser-task-blocked",
			Status: "blocked",
			Subtasks: []BrowserSubtaskObservationEvidence{
				{ID: "open", Kind: "open_target", Status: "blocked", Required: true, EvidenceKind: BrowserEvidenceKindFinalURL},
			},
		},
	}, BrowserEvidenceRequirements{
		RequiredSubtaskKinds: []string{"open_target"},
	})
	if result.Passed || result.TaskPlanReady {
		t.Fatalf("expected blocked task plan to fail, got %#v", result)
	}
	for _, reason := range []string{"task_plan_not_ready:blocked", "task_plan_subtask_not_observed:open"} {
		if !containsBrowserFailureReason(result.FailureReasons, reason) {
			t.Fatalf("expected task-plan failure reason %q in %#v", reason, result.FailureReasons)
		}
	}
}

func TestEvaluateBrowserEvidenceBundleRejectsTypeOnlyArtifactRefs(t *testing.T) {
	result := EvaluateBrowserEvidenceBundle(BrowserEvidenceBundle{
		ArtifactRefs: []BrowserArtifactRef{{Type: BrowserArtifactTypePageScreenshot, Role: BrowserEvidenceKindScreenshot}},
	}, BrowserEvidenceRequirements{
		RequireArtifactRefs: true,
	})
	if result.Passed || result.ArtifactRefsReady {
		t.Fatalf("expected type-only artifact ref to be insufficient, got %#v", result)
	}
	if !containsBrowserFailureReason(result.FailureReasons, "artifact_refs_missing") {
		t.Fatalf("expected artifact_refs_missing, got %#v", result.FailureReasons)
	}
}

func TestBrowserVisualEvidenceInputFromBundleMatchesVisualGate(t *testing.T) {
	bundle := BuildBrowserEvidenceBundleFromState(map[string]any{
		"target_url": "https://example.com/forms",
		"review": map[string]any{
			"snapshot":      "heading Submitted\nstatus Success",
			"final_url":     "https://example.com/forms/submitted",
			"evidence_path": "data:image/png;base64,AAAA",
		},
	})

	result := EvaluateBrowserVisualEvidenceGate(BrowserVisualEvidenceInputFromBundle(bundle, BrowserEvidenceRequirements{
		RequiredSnapshotTerms: []string{"Submitted"},
		RequireSnapshot:       true,
		RequireScreenshot:     true,
		RequireFinalURL:       true,
	}))
	if !result.Passed || !result.SnapshotReady || !result.ScreenshotReady || !result.FinalURLReady {
		t.Fatalf("expected visual gate to pass from evidence bundle, got %#v", result)
	}
}

func TestBrowserEvidenceBundleSchemaDocumentsCoreEvidenceKinds(t *testing.T) {
	schema := BrowserEvidenceBundleSchema()
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected schema properties, got %#v", schema)
	}
	for _, key := range []string{"final_url", "page_snapshot", "screenshot", "action_trace", "downloaded_file", "task_plan", "failure_reasons", "artifact_refs"} {
		if _, ok := props[key]; !ok {
			t.Fatalf("expected schema property %q in %#v", key, props)
		}
	}
}

func browserEvidenceTestHasArtifactType(values []BrowserArtifactRef, target string) bool {
	for _, value := range values {
		if value.Type == target {
			return true
		}
	}
	return false
}
