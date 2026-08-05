package team_test

import (
	"errors"
	"testing"

	capabilitycatalog "github.com/wsnacj/agentx-go/extensions/catalog"
	"github.com/wsnacj/agentx-go/extensions/team"
)

func TestBuildPlanIsDeterministicAndDetached(t *testing.T) {
	raw := team.Spec{ID: "review-team", Name: "Review Team", Coordinator: "lead", Members: []team.Member{
		{ID: "lead", ExpertID: "reviewer", DependsOn: []string{"research"}},
		{ID: "research", ExpertID: "researcher", Responsibility: "collect evidence"},
		{ID: "parallel", ExpertID: "researcher"},
	}, Tags: []string{"Review", "evidence"}}
	plan, err := team.BuildPlan(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Stages) != 2 || len(plan.Stages[0].Members) != 2 || plan.Stages[0].Members[0].ID != "parallel" || plan.Stages[1].Members[0].ID != "lead" {
		t.Fatalf("plan=%#v", plan)
	}
	raw.Members[0].DependsOn[0] = "mutated"
	if plan.Stages[0].Members[1].ID != "research" {
		t.Fatal("plan aliases caller input")
	}
	asset, err := team.Project("host:teams", rawTeam())
	if err != nil {
		t.Fatal(err)
	}
	if asset.Identity != (capabilitycatalog.Identity{Kind: capabilitycatalog.KindTeam, ID: "review-team"}) {
		t.Fatalf("asset=%#v", asset)
	}
}

func TestBuildPlanRejectsCycleAndHostOwnedFields(t *testing.T) {
	cycle := rawTeam()
	cycle.Members[0].DependsOn = []string{"review"}
	cycle.Members[1].DependsOn = []string{"research"}
	if _, err := team.BuildPlan(cycle); !errors.Is(err, &team.Error{Code: team.ErrorCodeDependencyCycle}) {
		t.Fatalf("err=%v", err)
	}
	_, err := team.Parse([]byte(`{"id":"t","name":"T","coordinator":"m","members":[{"id":"m","expert_id":"e"}],"scheduler":"auto"}`))
	if !errors.Is(err, &team.Error{Code: team.ErrorCodeForbiddenField}) {
		t.Fatalf("err=%v", err)
	}
}

func rawTeam() team.Spec {
	return team.Spec{ID: "review-team", Name: "Review Team", Coordinator: "review", Members: []team.Member{
		{ID: "research", ExpertID: "researcher"},
		{ID: "review", ExpertID: "reviewer", DependsOn: []string{"research"}},
	}}
}
