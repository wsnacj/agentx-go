package browserops

import "testing"

func TestEvaluateBrowserSiteSearchEvidencePassesExpectedResults(t *testing.T) {
	result := EvaluateBrowserSiteSearchEvidence(BrowserSiteSearchEvaluationInput{
		SnapshotText:        "heading Search results for Acme Robotics\nAcme Robotics Account\n/accounts/acme\nRobotics customer account summary",
		ScreenshotPath:      "data/tmp/browserops/site-search.png",
		FinalURL:            "https://example.com/search?q=Acme+Robotics",
		TargetURL:           "https://example.com/search",
		Query:               "Acme Robotics",
		URLContains:         "q=Acme",
		RequireQueryVisible: true,
		RequireScreenshot:   true,
		RequireSearchAction: true,
		RequireSubmitted:    true,
		MinSnapshotChars:    20,
		FieldCount:          1,
		Submitted:           true,
		ExpectedResults: []BrowserSiteSearchResultExpectation{
			{
				Title:           "Acme Robotics Account",
				URLContains:     "/accounts/acme",
				SnippetContains: "customer account summary",
			},
		},
	})
	if !result.Passed ||
		!result.EvidenceReady ||
		!result.QueryReady ||
		!result.SearchActionReady ||
		!result.ResultsReady ||
		result.Score != 1 {
		t.Fatalf("expected site search evidence to pass, got %#v", result)
	}
	for _, expected := range []string{"query_visible:Acme Robotics", "search_action_ready", "result_ready:acme_robotics_account", "final_url_contains:q=Acme"} {
		if !containsBrowserFailureReason(result.Evidence, expected) {
			t.Fatalf("expected evidence %q in %#v", expected, result.Evidence)
		}
	}
}

func TestEvaluateBrowserSiteSearchEvidenceFailsMissingExpectedResult(t *testing.T) {
	result := EvaluateBrowserSiteSearchEvidence(BrowserSiteSearchEvaluationInput{
		SnapshotText:      "heading Search results for Acme Robotics\nNo matching accounts",
		ScreenshotPath:    "data/tmp/browserops/site-search.png",
		FinalURL:          "https://example.com/search?q=Acme+Robotics",
		TargetURL:         "https://example.com/search",
		Query:             "Acme Robotics",
		RequireScreenshot: true,
		FieldCount:        1,
		ExpectedResults: []BrowserSiteSearchResultExpectation{
			{Title: "Acme Robotics Account", URLContains: "/accounts/acme"},
		},
	})
	if result.Passed || result.ResultsReady {
		t.Fatalf("expected site search results to fail, got %#v", result)
	}
	for _, reason := range []string{
		"result_acme_robotics_account:title_missing",
		"result_acme_robotics_account:url_missing",
	} {
		if !containsBrowserFailureReason(result.FailureReasons, reason) {
			t.Fatalf("expected reason %q in %#v", reason, result.FailureReasons)
		}
	}
}

func TestEvaluateBrowserSiteSearchEvidenceUsesEvidenceBundle(t *testing.T) {
	bundle := BuildBrowserEvidenceBundleFromState(map[string]any{
		"target_url": "https://example.com/search",
		"review": map[string]any{
			"snapshot":      "heading Search results for Acme Robotics\nAcme Robotics Account\n/accounts/acme",
			"final_url":     "https://example.com/search?q=Acme+Robotics",
			"evidence_path": "data:image/png;base64,AAAA",
		},
	})
	result := EvaluateBrowserSiteSearchEvidence(BrowserSiteSearchEvaluationInput{
		Bundle:            bundle,
		Query:             "Acme Robotics",
		RequireScreenshot: true,
		ExpectedResults: []BrowserSiteSearchResultExpectation{
			{Title: "Acme Robotics Account", URLContains: "/accounts/acme"},
		},
	})
	if !result.Passed || result.EvidenceBundle.PageSnapshot == nil || result.EvidenceBundle.Screenshot == nil {
		t.Fatalf("expected evidence bundle-backed site search gate to pass, got %#v", result)
	}
}
