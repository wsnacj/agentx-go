package expert_test

import (
	"errors"
	"testing"

	capabilitycatalog "github.com/wsnacj/agentx-go/extensions/catalog"
	"github.com/wsnacj/agentx-go/extensions/expert"
)

func TestNormalizeAndProjectExpert(t *testing.T) {
	raw := expert.Spec{
		ID: " Researcher ", Name: " Research Expert ", Instructions: " verify sources ",
		Requirements: []expert.Requirement{
			{Kind: capabilitycatalog.KindSkill, ID: "research-guide"},
			{Kind: capabilitycatalog.KindConnector, ID: "local-search", Optional: true},
			{Kind: capabilitycatalog.KindTool, ID: "web-search"},
		},
		Tags: []string{"Research", "evidence"},
	}
	normalized, err := expert.Normalize(raw)
	if err != nil {
		t.Fatal(err)
	}
	if normalized.ID != "researcher" || normalized.SchemaVersion != expert.SchemaVersionV1 || normalized.Instructions != "verify sources" {
		t.Fatalf("normalized=%#v", normalized)
	}
	if normalized.Requirements[0].Kind != capabilitycatalog.KindConnector || normalized.Requirements[2].Kind != capabilitycatalog.KindTool {
		t.Fatalf("requirements=%#v", normalized.Requirements)
	}
	raw.Requirements[0].ID = "mutated"
	if normalized.Requirements[1].ID != "research-guide" {
		t.Fatal("normalized requirements alias caller input")
	}
	asset, err := expert.Project("host:experts", normalized)
	if err != nil {
		t.Fatal(err)
	}
	if asset.Identity.Kind != capabilitycatalog.KindExpert || asset.Identity.ID != "researcher" || asset.SourceRef != "host:experts" {
		t.Fatalf("asset=%#v", asset)
	}
}

func TestParseRejectsHostOwnedFieldsAndTypedErrors(t *testing.T) {
	_, err := expert.Parse([]byte(`{"id":"researcher","name":"Researcher","instructions":"work","model":"secret-model"}`))
	if !errors.Is(err, &expert.Error{Code: expert.ErrorCodeForbiddenField}) {
		t.Fatalf("err=%v", err)
	}
	if typed, ok := expert.AsError(err); !ok || typed.Code != expert.ErrorCodeForbiddenField || typed.Cause == nil {
		t.Fatalf("typed=%#v ok=%t", typed, ok)
	}
	if _, err := expert.Normalize(expert.Spec{ID: "bad/id", Name: "Bad", Instructions: "work"}); !errors.Is(err, &expert.Error{Code: expert.ErrorCodeInvalidSpec}) {
		t.Fatalf("err=%v", err)
	}
}
