package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFromDirs_TolerantFrontmatterDescriptionColon(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "discord")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: discord
description: Use when you need to control Discord: send messages and manage channels.
metadata: {"moltbot":{"emoji":"🎮"}}
---
# Discord
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	items, err := LoadFromDirs([]string{root})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	if items[0].Name != "discord" {
		t.Fatalf("unexpected name: %#v", items[0])
	}
	if !strings.Contains(items[0].Description, "Discord") {
		t.Fatalf("expected description to be parsed, got: %q", items[0].Description)
	}
}

func TestLoadFromDirsWithReport_ContinueOnSkillError(t *testing.T) {
	root := t.TempDir()

	goodDir := filepath.Join(root, "good")
	if err := os.MkdirAll(goodDir, 0o755); err != nil {
		t.Fatal(err)
	}
	good := "---\nname: good\ndescription: ok\n---\n# Good\n"
	if err := os.WriteFile(filepath.Join(goodDir, "SKILL.md"), []byte(good), 0o644); err != nil {
		t.Fatal(err)
	}

	badDir := filepath.Join(root, "bad")
	if err := os.MkdirAll(badDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// metadata value is intentionally malformed to trigger parseCompatMetadata failure.
	bad := "---\nname: bad\ndescription: broken\nmetadata: \"{bad\"\n---\n# Bad\n"
	if err := os.WriteFile(filepath.Join(badDir, "SKILL.md"), []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}

	items, report, err := LoadFromDirsWithReport([]string{root}, LoadOptions{})
	if err != nil {
		t.Fatalf("expected tolerant load without fatal error, got %v", err)
	}
	if len(items) != 1 || items[0].Name != "good" {
		t.Fatalf("unexpected loaded items: %#v", items)
	}
	if report.Loaded != 1 {
		t.Fatalf("expected loaded=1, got %d", report.Loaded)
	}
	if report.ParseFailed != 1 {
		t.Fatalf("expected parse_failed=1, got %d", report.ParseFailed)
	}
	if len(report.Issues) == 0 {
		t.Fatalf("expected load issues, got %#v", report)
	}
	if !report.HasIssues() {
		t.Fatalf("expected HasIssues=true, got %#v", report)
	}
}

func TestLoadFromDirsWithReport_FailFast(t *testing.T) {
	root := t.TempDir()

	goodDir := filepath.Join(root, "good")
	if err := os.MkdirAll(goodDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(goodDir, "SKILL.md"), []byte("---\nname: good\ndescription: ok\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	badDir := filepath.Join(root, "bad")
	if err := os.MkdirAll(badDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(badDir, "SKILL.md"), []byte("---\nname: bad\nmetadata: \"{bad\"\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, report, err := LoadFromDirsWithReport([]string{root}, LoadOptions{FailFast: true})
	if err == nil {
		t.Fatalf("expected fail-fast error, report=%#v", report)
	}
	if report.ParseFailed != 1 {
		t.Fatalf("expected parse_failed=1 under fail-fast, got %d", report.ParseFailed)
	}
}

func TestLoadFromDirsWithReport_StrictFrontmatter(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "discord")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: discord\ndescription: bad: yaml\n---\n# Discord\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := LoadFromDirsWithReport([]string{root}, LoadOptions{StrictFrontmatter: true, FailFast: true})
	if err == nil {
		t.Fatal("expected strict frontmatter parse error")
	}
}
