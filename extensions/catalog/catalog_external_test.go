package catalog_test

import (
	"errors"
	"strings"
	"testing"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/extensions/catalog"
	"github.com/wsnacj/agentx-go/extensions/skills"
)

func TestBuildSearchAndProjection(t *testing.T) {
	assets := catalog.ProjectTools("canonical:tools", []toolcontract.Definition{{
		Type: "function",
		Function: toolcontract.Function{
			Name:        "Web_Search",
			Description: "检索公开网页",
		},
	}})
	assets = append(assets, catalog.ProjectSkills("workspace:skills", []skills.Skill{{
		Name:        "Research Brief",
		Description: "汇总公开研究资料",
		Tags:        []string{"Research", "Public"},
		Keywords:    []string{"Evidence", "Research"},
		Metadata:    map[string]string{"skill_key": "research-brief"},
	}})...)
	assets = append(assets,
		asset(catalog.KindPlugin, "research-kit", "Research Kit", "host:plugins", "research"),
		asset(catalog.KindConnector, "public-web", "Public Web", "host:connectors", "network"),
		asset(catalog.KindExpert, "researcher", "Researcher", "host:experts", "research"),
		asset(catalog.KindTeam, "research-team", "Research Team", "host:teams", "research"),
	)

	index, err := catalog.Build(catalog.DefaultPolicy(), assets)
	if err != nil {
		t.Fatal(err)
	}
	if index.Len() != 6 {
		t.Fatalf("Len()=%d, want 6", index.Len())
	}
	snapshot := index.Snapshot()
	if snapshot.Fingerprint == "" || len(snapshot.Assets) != 6 {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	if got := snapshot.Assets[5].Identity.Kind; got != catalog.KindTool {
		t.Fatalf("stable kind order ended with %q, want tool", got)
	}

	result, err := index.Search(catalog.Query{Text: "research", AnyTags: []string{"research"}, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if result.Matched != 4 || len(result.Hits) != 2 || !result.Limited {
		t.Fatalf("unexpected bounded result: %#v", result)
	}
	if result.Hits[0].Score <= 0 {
		t.Fatalf("expected explainable score: %#v", result.Hits[0])
	}

	skillResult, err := index.Search(catalog.Query{Kinds: []catalog.Kind{catalog.KindSkill}, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(skillResult.Hits) != 1 || skillResult.Hits[0].Asset.Identity.ID != "research-brief" {
		t.Fatalf("unexpected skill projection: %#v", skillResult)
	}

	// Returned snapshots and hits are detached from the immutable catalog.
	snapshot.Assets[0].Tags = append(snapshot.Assets[0].Tags, "mutated")
	result.Hits[0].Asset.Name = "mutated"
	again := index.Snapshot()
	if again.Assets[0].Name == "mutated" || contains(again.Assets[0].Tags, "mutated") {
		t.Fatalf("catalog was mutated through detached output: %#v", again.Assets[0])
	}
}

func TestBuildRejectsDuplicateIdentityWithDisplaySafeError(t *testing.T) {
	secret := "do-not-display"
	_, err := catalog.Build(catalog.DefaultPolicy(), []catalog.Asset{
		asset(catalog.KindTool, "echo", secret, "source:a", "utility"),
		asset(catalog.KindTool, " ECHO ", secret, "source:b", "utility"),
	})
	if !errors.Is(err, &catalog.Error{Code: catalog.ErrorCodeDuplicateAsset}) {
		t.Fatalf("err=%v", err)
	}
	typed, ok := catalog.AsError(err)
	if !ok || typed.Code != catalog.ErrorCodeDuplicateAsset {
		t.Fatalf("typed=%#v ok=%v", typed, ok)
	}
	if err.Error() == "" || strings.Contains(err.Error(), secret) {
		t.Fatalf("error is not display-safe: %q", err)
	}
}

func TestSearchIsStableAndDiffTracksIdentityChanges(t *testing.T) {
	policy := catalog.DefaultPolicy()
	before, err := catalog.Build(policy, []catalog.Asset{
		asset(catalog.KindTool, "fetch", "Fetch", "canonical:tools", "network"),
		asset(catalog.KindSkill, "research", "Research", "canonical:skills", "research"),
	})
	if err != nil {
		t.Fatal(err)
	}
	after, err := catalog.Build(policy, []catalog.Asset{
		asset(catalog.KindTool, "fetch", "Fetch URL", "canonical:tools", "network"),
		asset(catalog.KindTeam, "research", "Research Team", "host:teams", "research"),
	})
	if err != nil {
		t.Fatal(err)
	}
	changes := catalog.Diff(before.Snapshot(), after.Snapshot())
	if len(changes.Added) != 1 || changes.Added[0] != (catalog.Identity{Kind: catalog.KindTeam, ID: "research"}) {
		t.Fatalf("added=%#v", changes.Added)
	}
	if len(changes.Removed) != 1 || changes.Removed[0] != (catalog.Identity{Kind: catalog.KindSkill, ID: "research"}) {
		t.Fatalf("removed=%#v", changes.Removed)
	}
	if len(changes.Changed) != 1 || changes.Changed[0] != (catalog.Identity{Kind: catalog.KindTool, ID: "fetch"}) {
		t.Fatalf("changed=%#v", changes.Changed)
	}
}

func asset(kind catalog.Kind, id, name, sourceRef, tag string) catalog.Asset {
	return catalog.Asset{
		Identity:  catalog.Identity{Kind: kind, ID: id},
		Name:      name,
		SourceRef: sourceRef,
		Tags:      []string{tag},
		Keywords:  []string{tag},
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
