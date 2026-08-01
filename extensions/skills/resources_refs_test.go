package skills

import "testing"

func TestExtractReferencedResourcePaths(t *testing.T) {
	content := `
# Guide
Read references/runbook.md and scripts/fix.sh.
Also use assets/template.md.
Ignore ../references/escape.md.
`
	out := ExtractReferencedResourcePaths(content)
	if len(out) != 3 {
		t.Fatalf("expected 3 referenced paths, got %#v", out)
	}
	if out[0] != "assets/template.md" || out[1] != "references/runbook.md" || out[2] != "scripts/fix.sh" {
		t.Fatalf("unexpected referenced paths order: %#v", out)
	}
}

func TestMissingReferencedResourcePaths(t *testing.T) {
	skill := Skill{
		Name:    "resource-check",
		Content: "use references/runbook.md and scripts/fix.sh",
		Resources: Resources{
			References: []string{"references/runbook.md"},
		},
	}
	missing := MissingReferencedResourcePaths(skill)
	if len(missing) != 1 || missing[0] != "scripts/fix.sh" {
		t.Fatalf("unexpected missing resources: %#v", missing)
	}
}
