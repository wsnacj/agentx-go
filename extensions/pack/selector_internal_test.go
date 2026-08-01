package pack

import (
	"testing"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

func TestWorkflowRouteCaseTypesRejectsWhitespacePaddedStoredValues(t *testing.T) {
	spec := agentxworkflow.Spec{
		CaseTypes: []string{" browser.form_submission", "browser.record_update"},
	}
	got := workflowRouteCaseTypes(Manifest{}, spec)
	if len(got) != 1 || got[0] != "browser.record_update" {
		t.Fatalf("expected selector case types to keep only canonical values, got %#v", got)
	}

	manifest := Manifest{
		SupportedCaseTypes: []string{" browser.form_submission", "browser.record_update"},
	}
	got = workflowRouteCaseTypes(manifest, agentxworkflow.Spec{})
	if len(got) != 1 || got[0] != "browser.record_update" {
		t.Fatalf("expected selector manifest case types to keep only canonical values, got %#v", got)
	}
}

func TestScoreRouteCandidateDoesNotCanonicalizeStoredIdentifiers(t *testing.T) {
	def := Definition{
		Manifest: Manifest{
			ID:              " browserops ",
			DefaultWorkflow: " browser_form_submit_v1 ",
		},
	}
	spec := agentxworkflow.Spec{
		ID: " browser_form_submit_v1 ",
	}
	candidate := scoreRouteCandidate(def, spec, "browser.form_submission", normalizeRouteText("browserops browser_form_submit_v1 browser.form_submission"))

	if candidate.PackID != " browserops " || candidate.WorkflowID != " browser_form_submit_v1 " {
		t.Fatalf("expected raw identifiers to be preserved, got %#v", candidate)
	}
	if candidate.Score != 28 {
		t.Fatalf("expected only canonical case type exact match to score, got %#v", candidate)
	}
}

func TestSelectBindingPreservesRawMessageAndMatchedHintOutput(t *testing.T) {
	reg := selectorTestRegistry{defs: []Definition{
		{
			Manifest: Manifest{
				ID:                 "browserops",
				SupportedCaseTypes: []string{"browser.form_submission"},
				RouteHints:         []string{" browser submit "},
				DefaultWorkflow:    "browser_form_submit_v1",
			},
			Workflows: []agentxworkflow.Spec{
				{
					ID:         "browser_form_submit_v1",
					CaseTypes:  []string{"browser.form_submission"},
					RouteHints: []string{"browser submit"},
				},
			},
		},
	}}

	selection, ok := SelectBinding(reg, " browser submit ", SelectOptions{})
	if !ok {
		t.Fatalf("expected raw hint to still fuzzy-match, got %#v", selection)
	}
	if selection.Message != " browser submit " {
		t.Fatalf("expected raw selection message to be preserved, got %#v", selection)
	}
	if len(selection.Selected.MatchedHints) != 2 || selection.Selected.MatchedHints[0] != " browser submit " || selection.Selected.MatchedHints[1] != "browser submit" {
		t.Fatalf("expected raw matched hint to be preserved, got %#v", selection.Selected)
	}
}

type selectorTestRegistry struct {
	defs []Definition
}

func (r selectorTestRegistry) Register(def Definition) error { return nil }

func (r selectorTestRegistry) Get(id string) (Definition, bool) {
	for _, def := range r.defs {
		if def.Manifest.ID == id {
			return def, true
		}
	}
	return Definition{}, false
}

func (r selectorTestRegistry) List() []Definition {
	return append([]Definition(nil), r.defs...)
}

func (r selectorTestRegistry) ResolveWorkflow(packID string, caseType string) (agentxworkflow.Spec, bool) {
	return agentxworkflow.Spec{}, false
}

func (r selectorTestRegistry) ResolveMaterializedWorkflow(packID string, caseType string) (agentxworkflow.Spec, bool, error) {
	return agentxworkflow.Spec{}, false, nil
}
