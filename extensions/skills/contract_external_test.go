package skills_test

import (
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/wsnacj/agentx-go/extensions/skills"
)

func TestPortableSkillLoadActivationAndSemantics(t *testing.T) {
	source := fstest.MapFS{
		"review/SKILL.md":                {Data: []byte("---\nname: Repo-Review\npaths: [src/**]\ncontext: fork\nallowed_tools: [Read, edit, read]\neffort: high\n---\n# Review\nRead references/checklist.md first.\n")},
		"review/references/checklist.md": {Data: []byte("check boundaries\n")},
	}
	loaded, report, err := skills.LoadWithReport(skills.LoadOptions{
		ExtraFS:           []skills.FSSource{{ID: "contract/review", FS: source, Fingerprint: "fixture-v1"}},
		StrictFrontmatter: true,
	})
	if err != nil {
		t.Fatalf("LoadWithReport(): %v", err)
	}
	if report.HasIssues() || len(loaded) != 1 || loaded[0].Name != "Repo-Review" {
		t.Fatalf("unexpected load result: skills=%#v report=%+v", loaded, report)
	}
	if active, reason, _ := skills.EvaluateSkillActivationPaths(loaded[0], []string{"src/main.go"}); !active || reason != "" {
		t.Fatalf("activation: active=%v reason=%q", active, reason)
	}
	semantics := skills.ResolveRequestedSkillSemantics(loaded, []string{"repo-review"})
	want := []skills.RequestedSkillSemantic{{
		Name:             "repo-review",
		ExecutionContext: skills.SkillExecutionContextFork,
		AllowedTools:     []string{"read", "edit"},
		Effort:           "high",
	}}
	if !reflect.DeepEqual(semantics, want) {
		t.Fatalf("semantics=%#v want=%#v", semantics, want)
	}
	if missing := skills.MissingReferencedResourcePaths(loaded[0]); len(missing) != 0 {
		t.Fatalf("missing resources: %v", missing)
	}
	cloned := skills.Clone(loaded)
	cloned[0].AllowedTools[0] = "changed"
	if loaded[0].AllowedTools[0] != "read" {
		t.Fatalf("Clone shared caller mutation: %#v", loaded[0].AllowedTools)
	}
}
