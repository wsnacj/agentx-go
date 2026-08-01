package skills

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/fsnotify/fsnotify"

	agentxassetfs "github.com/wsnacj/agentx-go/runtime/assetfs"
)

type lyingSizeFS struct {
	fs.FS
}

func (source lyingSizeFS) Open(name string) (fs.File, error) {
	file, err := source.FS.Open(name)
	if err != nil {
		return nil, err
	}
	return lyingSizeFile{File: file}, nil
}

func (source lyingSizeFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(source.FS, name)
}

type lyingSizeFile struct {
	fs.File
}

func (file lyingSizeFile) Stat() (fs.FileInfo, error) {
	info, err := file.File.Stat()
	if err != nil {
		return nil, err
	}
	return lyingSizeInfo{FileInfo: info}, nil
}

type lyingSizeInfo struct {
	fs.FileInfo
}

func (lyingSizeInfo) Size() int64 {
	return 0
}

func TestLoadFSRejectsContentBeyondLimitWhenStatUnderreportsSize(t *testing.T) {
	content := "---\nname: oversized\n---\n" + strings.Repeat("x", 256)
	source := lyingSizeFS{FS: fstest.MapFS{
		"oversized/SKILL.md": {Data: []byte(content)},
	}}
	_, report, err := LoadWithReport(LoadOptions{
		ExtraFS: []FSSource{{
			ID:          "test/lying-size",
			FS:          source,
			Fingerprint: "caller-untrusted",
		}},
		MaxSkillFileBytes: 64,
	})
	if err != nil {
		t.Fatalf("non-fail-fast load: %v", err)
	}
	if report.Skipped != 1 || report.ParseFailed != 0 || len(report.Issues) != 1 ||
		report.Issues[0].Code != "skill_doc_too_large" {
		t.Fatalf("expected bounded read rejection, got %+v", report)
	}
}

func TestLoadFromDirs(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "weather")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: weather\ndescription: query weather\nkeywords:\n  - weather\n  - forecast\ntags:\n  - meteo\nwhen_to_use:\n  - user asks weather\nnegative_examples:\n  - user only asks for weather definition\nsteps:\n  - resolve location\n  - fetch forecast\ntool_hints:\n  - web_search\neval_assertions:\n  - should include location in answer\n---\n# Weather\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	skills, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "weather" {
		t.Fatalf("unexpected skill name: %s", skills[0].Name)
	}
	if len(skills[0].WhenToUse) != 1 || skills[0].WhenToUse[0] != "user asks weather" {
		t.Fatalf("unexpected when_to_use: %#v", skills[0].WhenToUse)
	}
	if len(skills[0].Steps) != 2 {
		t.Fatalf("unexpected steps: %#v", skills[0].Steps)
	}
	if len(skills[0].ToolHints) != 1 || skills[0].ToolHints[0] != "web_search" {
		t.Fatalf("unexpected tool hints: %#v", skills[0].ToolHints)
	}
	if len(skills[0].Keywords) != 2 || skills[0].Keywords[0] != "weather" || skills[0].Keywords[1] != "forecast" {
		t.Fatalf("unexpected keywords: %#v", skills[0].Keywords)
	}
	if len(skills[0].Tags) != 1 || skills[0].Tags[0] != "meteo" {
		t.Fatalf("unexpected tags: %#v", skills[0].Tags)
	}
	if len(skills[0].NegativeExamples) != 1 {
		t.Fatalf("unexpected negative examples: %#v", skills[0].NegativeExamples)
	}
	if len(skills[0].EvalAssertions) != 1 {
		t.Fatalf("unexpected eval assertions: %#v", skills[0].EvalAssertions)
	}
}

func TestLoadFromDirs_LooseFrontmatterPreservesListFields(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "news")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\n" +
		"name: news\n" +
		"description: query news\n" +
		"steps:\n" +
		"  - `parse value` intentionally trips strict yaml\n" +
		"  - open source\n" +
		"tool_hints:\n" +
		"  - search\n" +
		"  - open_page\n" +
		"allowed_tools:\n" +
		"  - search\n" +
		"  - open_page\n" +
		"---\n" +
		"# News\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	skills, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if len(skills[0].Steps) != 2 || skills[0].Steps[0] != "`parse value` intentionally trips strict yaml" {
		t.Fatalf("expected loose frontmatter steps to remain a list, got %#v", skills[0].Steps)
	}
	if len(skills[0].ToolHints) != 2 || skills[0].ToolHints[0] != "search" || skills[0].ToolHints[1] != "open_page" {
		t.Fatalf("expected loose frontmatter tool_hints to remain a list, got %#v", skills[0].ToolHints)
	}
	if len(skills[0].AllowedTools) != 2 || skills[0].AllowedTools[0] != "search" || skills[0].AllowedTools[1] != "open_page" {
		t.Fatalf("expected loose frontmatter allowed_tools to remain a list, got %#v", skills[0].AllowedTools)
	}
}

func TestLoadFromDirs_ParseKeywordSections(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "weather")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: weather\n---\n# Keywords\n- weather\n- forecast\n\n# Tags\n- meteo\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	if len(items[0].Keywords) != 2 || items[0].Keywords[0] != "weather" || items[0].Keywords[1] != "forecast" {
		t.Fatalf("unexpected keyword section parse: %#v", items[0].Keywords)
	}
	if len(items[0].Tags) != 1 || items[0].Tags[0] != "meteo" {
		t.Fatalf("unexpected tags section parse: %#v", items[0].Tags)
	}
}

func TestLoadFromDirs_ParseMarkdownSections(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "incident")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: incident\n---\n# When to use\n- production failures\n\n# Steps\n1. gather logs\n2. identify root cause\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	if len(items[0].WhenToUse) != 1 || items[0].WhenToUse[0] != "production failures" {
		t.Fatalf("unexpected when_to_use from markdown: %#v", items[0].WhenToUse)
	}
	if len(items[0].Steps) != 2 {
		t.Fatalf("unexpected parsed steps: %#v", items[0].Steps)
	}
}

func TestLoadFromDirs_ParsesPathScopesFrontmatter(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "router")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: router\npaths:\n  - web/**\n  - ./api/routes\n  - web/**\n---\n# Router\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	if len(items[0].Paths) != 2 || items[0].Paths[0] != "web/**" || items[0].Paths[1] != "api/routes" {
		t.Fatalf("unexpected path scopes: %#v", items[0].Paths)
	}
}

func TestLoadFromDirs_SkipsOversizedSkillDoc(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "huge")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: huge\n---\n" + strings.Repeat("x", 256)
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, report, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{MaxSkillFileBytes: 64})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected oversized skill to be skipped, got %d items", len(items))
	}
	if report.ParseFailed != 0 {
		t.Fatalf("expected oversized skill to skip without parse failure, got %d", report.ParseFailed)
	}
	if len(report.Issues) != 1 || report.Issues[0].Code != "skill_doc_too_large" {
		t.Fatalf("expected oversized issue, got %#v", report.Issues)
	}
}

func TestLoadFromDirs_RespectsCandidateAndLoadedCaps(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"alpha", "beta", "gamma"} {
		skillDir := filepath.Join(dir, name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := "---\nname: " + name + "\n---\n# Skill\n"
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	items, report, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{
		MaxCandidatesPerRoot:     2,
		MaxSkillsLoadedPerSource: 1,
	})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 || items[0].Name != "alpha" {
		t.Fatalf("expected only first skill to load, got %#v", items)
	}
	if len(report.Issues) != 2 {
		t.Fatalf("expected 2 issues, got %#v", report.Issues)
	}
	if report.Issues[0].Code != "loaded_limit_reached" || report.Issues[1].Code != "candidate_limit_reached" {
		t.Fatalf("unexpected issues: %#v", report.Issues)
	}
}

func TestLoadFromDirs_SkipsHiddenAndNodeModulesDirs(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{".hidden", "node_modules", "weather"} {
		skillDir := filepath.Join(dir, name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := "---\nname: " + strings.TrimPrefix(name, ".") + "\n---\n# Skill\n"
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 || items[0].Name != "weather" {
		t.Fatalf("expected only visible skill to load, got %#v", items)
	}
}

func TestLoadFromDirs_UsesNestedSkillsRoot(t *testing.T) {
	root := t.TempDir()
	skillsRoot := filepath.Join(root, "skills")
	skillDir := filepath.Join(skillsRoot, "weather")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: weather\n---\n# Weather\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{root})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 || items[0].Name != "weather" {
		t.Fatalf("expected nested skills root to load weather, got %#v", items)
	}
}

func TestLoadFromDirs_PrefersDirectSkillRootOverNestedSkillsDir(t *testing.T) {
	root := t.TempDir()
	directDir := filepath.Join(root, "weather")
	if err := os.MkdirAll(directDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directDir, "SKILL.md"), []byte("---\nname: weather\n---\n# Weather\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nestedDir := filepath.Join(root, "skills", "ignored")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "SKILL.md"), []byte("---\nname: ignored\n---\n# Ignored\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{root})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 || items[0].Name != "weather" {
		t.Fatalf("expected direct skill root to win, got %#v", items)
	}
}

func TestLoadFromDirs_ParseInvocationAndRuntimeMetadata(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "ops")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: ops
description: ops helper
disable-model-invocation: true
user-invocable: false
os:
  - darwin
requires:
  bins: [git]
  any_bins: [jq, yq]
  env: [GITHUB_TOKEN]
  config: [skills.enable]
install:
  - id: install-jq
    kind: brew
    formula: jq
    bins: [jq]
    rollback: [brew, uninstall, jq]
  - kind: custom
    depends_on: [install-jq]
    command: [echo, install]
    rollback_command: [echo, rollback]
---
# Ops
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	skill := items[0]
	if !skill.Invocation.DisableModelInvocation || skill.Invocation.UserInvocable {
		t.Fatalf("unexpected invocation policy: %#v", skill.Invocation)
	}
	if len(skill.OS) != 1 || skill.OS[0] != "darwin" {
		t.Fatalf("unexpected os metadata: %#v", skill.OS)
	}
	if len(skill.Requires.Bins) != 1 || skill.Requires.Bins[0] != "git" {
		t.Fatalf("unexpected requires bins: %#v", skill.Requires)
	}
	if len(skill.Install) != 2 || skill.Install[0].Kind != "brew" || skill.Install[0].Formula != "jq" {
		t.Fatalf("unexpected install specs: %#v", skill.Install)
	}
	if len(skill.Install[0].Rollback) != 3 || skill.Install[0].Rollback[0] != "brew" {
		t.Fatalf("expected rollback command to be parsed: %#v", skill.Install[0])
	}
	if skill.Install[1].DependsOn[0] != "install-jq" {
		t.Fatalf("expected depends_on to be parsed: %#v", skill.Install[1])
	}
	if len(skill.Install[1].Command) != 2 || skill.Install[1].Command[0] != "echo" {
		t.Fatalf("expected install command to be parsed: %#v", skill.Install[1])
	}
}

func TestLoadFromDirs_ParseCommandDispatchAndInstallExtras(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "dispatch")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: dispatch
description: command dispatch demo
command-dispatch: tool
command-tool: ticket_lookup
command-arg-mode: raw
command-name: ticket
command-aliases:
  - tk
  - ticketing
install:
  - kind: brew
    tap: homebrew/cask
    cask: slack
  - kind: apt
    package: jq
---
# Dispatch
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	skill := items[0]
	if skill.Dispatch == nil {
		t.Fatalf("expected command dispatch metadata, got nil")
	}
	if skill.Dispatch.Kind != "tool" || skill.Dispatch.Tool != "ticket_lookup" || skill.Dispatch.ArgMode != "raw" || skill.Dispatch.Command != "ticket" {
		t.Fatalf("unexpected dispatch metadata: %#v", skill.Dispatch)
	}
	if len(skill.Dispatch.Aliases) != 2 || skill.Dispatch.Aliases[0] != "ticketing" || skill.Dispatch.Aliases[1] != "tk" {
		t.Fatalf("unexpected dispatch aliases: %#v", skill.Dispatch)
	}
	if len(skill.Install) != 2 {
		t.Fatalf("expected 2 install specs, got %#v", skill.Install)
	}
	if skill.Install[0].Tap != "homebrew/cask" || skill.Install[0].Cask != "slack" {
		t.Fatalf("expected brew tap/cask parsed, got %#v", skill.Install[0])
	}
	if skill.Install[1].Kind != "apt" || skill.Install[1].Package != "jq" {
		t.Fatalf("expected apt install parsed, got %#v", skill.Install[1])
	}
}

func TestLoadFromDirs_ParseCommandDispatchMissingToolForLint(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "dispatch-invalid")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: dispatch-invalid
description: invalid dispatch demo
command-dispatch: tool
---
# Dispatch Invalid
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	if items[0].Dispatch == nil || items[0].Dispatch.Kind != "tool" {
		t.Fatalf("expected dispatch metadata to be preserved for lint, got %#v", items[0].Dispatch)
	}
	if strings.TrimSpace(items[0].Dispatch.Tool) != "" {
		t.Fatalf("expected empty tool for invalid dispatch metadata, got %#v", items[0].Dispatch)
	}
}

func TestLoadFromDirs_ParseCamelCaseFrontmatterAliases(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "camel")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: camel
description: camel case aliases
whenToUse: ["when alias works"]
toolHints: ["web_search"]
toolHintsMatch: all
skillContext: fork
allowedTools: ["web_search", "read", "web_search"]
reasoningEffort: high
userInvocable: false
disableModelInvocation: true
commandDispatch: tool
commandTool: web_search
commandArgMode: raw
commandName: lookup
commandAliases: ["search", "lookup-fast"]
---
# Camel
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	skill := items[0]
	if len(skill.WhenToUse) != 1 || skill.WhenToUse[0] != "when alias works" {
		t.Fatalf("unexpected whenToUse alias parsing: %#v", skill.WhenToUse)
	}
	if len(skill.ToolHints) != 1 || skill.ToolHints[0] != "web_search" {
		t.Fatalf("unexpected toolHints alias parsing: %#v", skill.ToolHints)
	}
	if skill.ToolHintsMatch != ToolHintsMatchAll {
		t.Fatalf("unexpected toolHintsMatch alias parsing: %q", skill.ToolHintsMatch)
	}
	if skill.ExecutionContext != SkillExecutionContextFork {
		t.Fatalf("unexpected skillContext alias parsing: %q", skill.ExecutionContext)
	}
	if len(skill.AllowedTools) != 2 || skill.AllowedTools[0] != "web_search" || skill.AllowedTools[1] != "read" {
		t.Fatalf("unexpected allowedTools alias parsing: %#v", skill.AllowedTools)
	}
	if skill.Effort != "high" {
		t.Fatalf("unexpected reasoningEffort alias parsing: %q", skill.Effort)
	}
	if skill.Invocation.UserInvocable || !skill.Invocation.DisableModelInvocation {
		t.Fatalf("unexpected invocation camelCase parsing: %#v", skill.Invocation)
	}
	if skill.Dispatch == nil {
		t.Fatalf("expected command dispatch from camelCase aliases")
	}
	if skill.Dispatch.Tool != "web_search" || skill.Dispatch.Command != "lookup" || skill.Dispatch.ArgMode != "raw" {
		t.Fatalf("unexpected dispatch alias parsing: %#v", skill.Dispatch)
	}
	if len(skill.Dispatch.Aliases) != 2 || skill.Dispatch.Aliases[0] != "lookup-fast" || skill.Dispatch.Aliases[1] != "search" {
		t.Fatalf("unexpected dispatch command aliases parsing: %#v", skill.Dispatch)
	}
}

func TestLoadFromDirs_ParseMetadataCompat(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "compat")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: compat
description: compat metadata
metadata: |
  {
    moltbot: {
      os: [linux],
      requires: {
        bins: [gh],
        anyBins: [jq, yq],
        env: [GITHUB_TOKEN],
        config: [skills.enable]
      },
      install: [
        { kind: "brew", formula: "gh", bins: [gh] }
      ],
      primaryEnv: "GITHUB_TOKEN"
    }
  }
---
# Compat
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	skill := items[0]
	if len(skill.OS) != 1 || skill.OS[0] != "linux" {
		t.Fatalf("unexpected os from metadata: %#v", skill.OS)
	}
	if len(skill.Requires.Bins) != 1 || skill.Requires.Bins[0] != "gh" {
		t.Fatalf("unexpected requires bins from metadata: %#v", skill.Requires)
	}
	if len(skill.Requires.AnyBins) != 2 || skill.Requires.AnyBins[0] != "jq" {
		t.Fatalf("unexpected requires anyBins from metadata: %#v", skill.Requires)
	}
	if len(skill.Install) != 1 || skill.Install[0].Kind != "brew" || skill.Install[0].Formula != "gh" {
		t.Fatalf("unexpected install from metadata: %#v", skill.Install)
	}
	if skill.Metadata["primary_env"] != "GITHUB_TOKEN" {
		t.Fatalf("expected metadata primary_env fallback, got %#v", skill.Metadata)
	}
}

func TestLoadFromDirs_ParseMetadataOpenClawCompat(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "compat-openclaw")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: compat-openclaw
description: compat metadata from openclaw envelope
metadata: |
  {
    openclaw: {
      os: [darwin],
      requires: {
        bins: [tmux],
        anyBins: [jq, yq],
        env: [OPENAI_API_KEY],
        config: [skills.enable]
      },
      install: [
        { kind: "brew", formula: "tmux", bins: [tmux] }
      ],
      primaryEnv: "OPENAI_API_KEY",
      skillKey: "weather-pro",
      homepage: "https://example.com"
    }
  }
---
# Compat OpenClaw
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	skill := items[0]
	if len(skill.OS) != 1 || skill.OS[0] != "darwin" {
		t.Fatalf("unexpected os from metadata.openclaw: %#v", skill.OS)
	}
	if len(skill.Requires.Bins) != 1 || skill.Requires.Bins[0] != "tmux" {
		t.Fatalf("unexpected requires bins from metadata.openclaw: %#v", skill.Requires)
	}
	if len(skill.Requires.AnyBins) != 2 || skill.Requires.AnyBins[0] != "jq" {
		t.Fatalf("unexpected requires anyBins from metadata.openclaw: %#v", skill.Requires)
	}
	if len(skill.Install) != 1 || skill.Install[0].Kind != "brew" || skill.Install[0].Formula != "tmux" {
		t.Fatalf("unexpected install from metadata.openclaw: %#v", skill.Install)
	}
	if skill.Metadata["primary_env"] != "OPENAI_API_KEY" {
		t.Fatalf("expected metadata primary_env fallback, got %#v", skill.Metadata)
	}
	if skill.Metadata["skill_key"] != "weather-pro" {
		t.Fatalf("expected metadata skill_key fallback, got %#v", skill.Metadata)
	}
	if skill.Metadata["homepage"] != "https://example.com" {
		t.Fatalf("expected metadata homepage fallback, got %#v", skill.Metadata)
	}
}

func TestLoadFromDirs_ParseMetadataCompat_OpenClawPreferred(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "compat-openclaw-priority")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: compat-openclaw-priority
description: openclaw envelope should win when multiple compat envelopes exist
metadata: |
  {
    openclaw: {
      requires: { bins: [openclaw-bin] },
      primaryEnv: "OPENCLAW_TOKEN"
    },
    moltbot: {
      requires: { bins: [moltbot-bin] },
      primaryEnv: "MOLTBOT_TOKEN"
    }
  }
---
# Compat OpenClaw Priority
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	skill := items[0]
	if len(skill.Requires.Bins) != 1 || skill.Requires.Bins[0] != "openclaw-bin" {
		t.Fatalf("expected metadata.openclaw to be preferred, got %#v", skill.Requires)
	}
	if skill.Metadata["primary_env"] != "OPENCLAW_TOKEN" {
		t.Fatalf("expected metadata.openclaw primary_env to be preferred, got %#v", skill.Metadata)
	}
}

func TestLoadFromDirs_MetadataCompatTopLevelOverrides(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "override")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: override
description: override metadata
os: [darwin]
requires:
  bins: [git]
install:
  - kind: custom
    command: [echo, install]
metadata: |
  {
    moltbot: {
      os: [linux],
      requires: {
        bins: [gh],
        anyBins: [jq, yq]
      },
      install: [
        { kind: "brew", formula: "gh", bins: [gh] }
      ]
    }
  }
---
# Override
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	skill := items[0]
	if len(skill.OS) != 1 || skill.OS[0] != "darwin" {
		t.Fatalf("expected top-level os override, got %#v", skill.OS)
	}
	if len(skill.Requires.Bins) != 1 || skill.Requires.Bins[0] != "git" {
		t.Fatalf("expected top-level requires bins override, got %#v", skill.Requires)
	}
	// Missing anyBins at top-level should fallback from metadata.moltbot.
	if len(skill.Requires.AnyBins) != 2 || skill.Requires.AnyBins[0] != "jq" {
		t.Fatalf("expected metadata anyBins fallback, got %#v", skill.Requires)
	}
	if len(skill.Install) != 1 || skill.Install[0].Kind != "custom" {
		t.Fatalf("expected top-level install override, got %#v", skill.Install)
	}
}

func TestLoadFromDirs_MetadataCompatInvocationDispatchFallback(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "compat-fallback")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: compat-fallback
description: metadata invocation and dispatch fallback
metadata: |
  {
    moltbot: {
      userInvocable: false,
      disableModelInvocation: true,
      commandDispatch: "tool",
      commandTool: "shell_execute",
      commandArgMode: "raw",
      commandName: "run",
      context: "fork",
      allowedTools: ["shell_execute", "read"],
      effort: "medium",
      whenToUse: ["run shell commands"],
      toolHints: ["shell_execute"],
      toolHintsMatch: "all",
      steps: ["collect args", "run command"],
      evalAssertions: ["must call shell_execute"]
    }
  }
---
# Compat Fallback
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	skill := items[0]
	if skill.Invocation.UserInvocable {
		t.Fatalf("expected metadata userInvocable=false fallback, got %#v", skill.Invocation)
	}
	if !skill.Invocation.DisableModelInvocation {
		t.Fatalf("expected metadata disableModelInvocation=true fallback, got %#v", skill.Invocation)
	}
	if skill.Dispatch == nil {
		t.Fatalf("expected metadata dispatch fallback, got nil")
	}
	if skill.Dispatch.Kind != "tool" || skill.Dispatch.Tool != "shell_execute" || skill.Dispatch.Command != "run" {
		t.Fatalf("unexpected dispatch fallback: %#v", skill.Dispatch)
	}
	if skill.ExecutionContext != SkillExecutionContextFork {
		t.Fatalf("expected metadata execution context fallback, got %q", skill.ExecutionContext)
	}
	if len(skill.AllowedTools) != 2 || skill.AllowedTools[0] != "shell_execute" || skill.AllowedTools[1] != "read" {
		t.Fatalf("unexpected allowed_tools fallback: %#v", skill.AllowedTools)
	}
	if skill.Effort != "medium" {
		t.Fatalf("unexpected effort fallback: %q", skill.Effort)
	}
	if len(skill.WhenToUse) != 1 || skill.WhenToUse[0] != "run shell commands" {
		t.Fatalf("unexpected when_to_use fallback: %#v", skill.WhenToUse)
	}
	if len(skill.ToolHints) != 1 || skill.ToolHints[0] != "shell_execute" {
		t.Fatalf("unexpected tool_hints fallback: %#v", skill.ToolHints)
	}
	if skill.ToolHintsMatch != ToolHintsMatchAll {
		t.Fatalf("unexpected tool_hints_match fallback: %q", skill.ToolHintsMatch)
	}
	if len(skill.EvalAssertions) != 1 || skill.EvalAssertions[0] != "must call shell_execute" {
		t.Fatalf("unexpected eval_assertions fallback: %#v", skill.EvalAssertions)
	}
}

func TestLoadFromDirs_MetadataCompatTopLevelInvocationDispatchOverride(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "compat-override")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: compat-override
description: top-level invocation and dispatch override metadata
user-invocable: true
context: inline
allowed_tools: [top_tool]
effort: high
command-dispatch: tool
command-tool: top_tool
command-name: topcmd
metadata: |
  {
    moltbot: {
      userInvocable: false,
      disableModelInvocation: true,
      context: "fork",
      allowedTools: ["metadata_tool"],
      effort: "low",
      commandDispatch: "tool",
      commandTool: "metadata_tool",
      commandName: "metacmd"
    }
  }
---
# Compat Override
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	skill := items[0]
	if !skill.Invocation.UserInvocable {
		t.Fatalf("expected top-level user-invocable override, got %#v", skill.Invocation)
	}
	if !skill.Invocation.DisableModelInvocation {
		t.Fatalf("expected metadata disable-model fallback when top-level absent, got %#v", skill.Invocation)
	}
	if skill.ExecutionContext != SkillExecutionContextInline {
		t.Fatalf("expected top-level execution context override, got %q", skill.ExecutionContext)
	}
	if len(skill.AllowedTools) != 1 || skill.AllowedTools[0] != "top_tool" {
		t.Fatalf("expected top-level allowed_tools override, got %#v", skill.AllowedTools)
	}
	if skill.Effort != "high" {
		t.Fatalf("expected top-level effort override, got %q", skill.Effort)
	}
	if skill.Dispatch == nil {
		t.Fatalf("expected top-level dispatch")
	}
	if skill.Dispatch.Tool != "top_tool" || skill.Dispatch.Command != "topcmd" {
		t.Fatalf("expected top-level dispatch override, got %#v", skill.Dispatch)
	}
}

func TestLoadFromDirs_MetadataCompatTopLevelDispatchPresenceBlocksFallback(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "compat-dispatch-presence")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: compat-dispatch-presence
description: explicit empty top-level dispatch disables metadata dispatch fallback
command-dispatch:
metadata: |
  {
    moltbot: {
      commandDispatch: "tool",
      commandTool: "metadata_tool",
      commandName: "metacmd"
    }
  }
---
# Compat Dispatch Presence
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	if items[0].Dispatch != nil {
		t.Fatalf("expected top-level dispatch presence to block metadata fallback, got %#v", items[0].Dispatch)
	}
}

func TestLoad_WithSourcePrecedence(t *testing.T) {
	makeSkill := func(root, name, desc string) {
		t.Helper()
		skillDir := filepath.Join(root, name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := "---\nname: " + name + "\ndescription: " + desc + "\n---\n# " + name + "\n"
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	extra := t.TempDir()
	bundled := t.TempDir()
	managed := t.TempDir()
	custom := t.TempDir()
	workspace := t.TempDir()
	makeSkill(extra, "planner", "from-extra")
	makeSkill(bundled, "planner", "from-bundled")
	makeSkill(managed, "planner", "from-managed")
	makeSkill(custom, "planner", "from-custom")
	makeSkill(workspace, "planner", "from-workspace")

	items, err := Load(LoadOptions{
		CustomDirs:   []string{custom},
		ExtraDirs:    []string{extra},
		BundledDir:   bundled,
		ManagedDir:   managed,
		WorkspaceDir: workspace,
	})
	if err != nil {
		t.Fatalf("load skills with precedence: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 merged skill, got %d", len(items))
	}
	if items[0].Description != "from-workspace" {
		t.Fatalf("expected workspace to override, got: %#v", items[0])
	}
	if items[0].Source != SourceWorkspace {
		t.Fatalf("expected workspace source, got: %s", items[0].Source)
	}
}

func TestBuildLoadSources_UsesCanonicalSourceOrder(t *testing.T) {
	sources := buildLoadSources(LoadOptions{
		ExtraDirs:    []string{"/extra-root"},
		BundledDir:   "/bundled-root",
		ManagedDir:   "/managed-root",
		CustomDirs:   []string{"/custom-root"},
		WorkspaceDir: "/workspace-root",
	})
	if len(sources) != 5 {
		t.Fatalf("expected exactly 5 canonical sources, got %#v", sources)
	}
	wantKinds := []Source{
		SourceExtra,
		SourceBundled,
		SourceManaged,
		SourceCustom,
		SourceWorkspace,
	}
	wantDirs := []string{
		"/extra-root",
		"/bundled-root",
		"/managed-root",
		"/custom-root",
		"/workspace-root",
	}
	for i, want := range wantKinds {
		if sources[i].Kind != want {
			t.Fatalf("unexpected source kind at %d: got %q want %q", i, sources[i].Kind, want)
		}
		if sources[i].Dir != wantDirs[i] {
			t.Fatalf("unexpected source dir at %d: got %q want %q", i, sources[i].Dir, wantDirs[i])
		}
	}
}

func TestBuildLoadSources_OrdersImmutableSourcesWithTheirPrecedenceClass(t *testing.T) {
	extraFS := FSSource{ID: "extra-fs", FS: fstest.MapFS{}, Fingerprint: "extra-v1"}
	bundledFS := FSSource{ID: "bundled-fs", FS: fstest.MapFS{}, Fingerprint: "bundled-v1"}
	sources := buildLoadSources(LoadOptions{
		ExtraDirs:    []string{"/extra-dir"},
		ExtraFS:      []FSSource{extraFS},
		BundledDir:   "/bundled-dir",
		BundledFS:    bundledFS,
		ManagedDir:   "/managed-dir",
		CustomDirs:   []string{"/custom-dir"},
		WorkspaceDir: "/workspace-dir",
	})
	if len(sources) != 7 {
		t.Fatalf("expected seven sources, got %#v", sources)
	}
	wantKinds := []Source{
		SourceExtra,
		SourceExtra,
		SourceBundled,
		SourceBundled,
		SourceManaged,
		SourceCustom,
		SourceWorkspace,
	}
	for idx, want := range wantKinds {
		if sources[idx].Kind != want {
			t.Fatalf("unexpected source kind at %d: got %q want %q", idx, sources[idx].Kind, want)
		}
	}
	if sources[1].ID != extraFS.ID || !sources[1].FSBacked {
		t.Fatalf("unexpected immutable extra source: %#v", sources[1])
	}
	if sources[3].ID != bundledFS.ID || !sources[3].FSBacked {
		t.Fatalf("unexpected immutable bundled source: %#v", sources[3])
	}
}

func TestLoadWithReport_ImmutableFSSourceResourcesAndCacheSnapshot(t *testing.T) {
	source := fstest.MapFS{
		"ops/SKILL.md":              &fstest.MapFile{Data: []byte("---\nname: ops\ndescription: embedded v1\n---\n# Ops\n")},
		"ops/scripts/run.sh":        &fstest.MapFile{Data: []byte("#!/bin/sh\n")},
		"ops/references/guide.md":   &fstest.MapFile{Data: []byte("# Guide\n")},
		"ops/assets/template.json":  &fstest.MapFile{Data: []byte("{}\n")},
		"ops/assets/nested/data.md": &fstest.MapFile{Data: []byte("data\n")},
	}
	provider, err := agentxassetfs.New("agentx.test.immutable."+t.Name(), source)
	if err != nil {
		t.Fatalf("new immutable provider: %v", err)
	}
	opts := LoadOptions{
		BundledFS: FSSource{
			ID:          provider.ID(),
			FS:          provider.FS(),
			Fingerprint: provider.Fingerprint(),
		},
	}
	if !opts.BundledFS.Valid() {
		t.Fatalf("expected immutable source to be valid: %#v", opts.BundledFS)
	}
	first, firstReport, err := LoadWithReport(opts)
	if err != nil {
		t.Fatalf("first immutable load: %v", err)
	}
	if firstReport.CacheHit || firstReport.Generation == 0 {
		t.Fatalf("first immutable load should establish a generation: %+v", firstReport)
	}
	if len(first) != 1 || first[0].Description != "embedded v1" {
		t.Fatalf("unexpected first immutable load: %#v", first)
	}
	if first[0].Location != "assetfs://"+opts.BundledFS.ID+"/ops/SKILL.md" ||
		first[0].BaseDir != "assetfs://"+opts.BundledFS.ID+"/ops" {
		t.Fatalf("unexpected immutable source refs: location=%q base=%q", first[0].Location, first[0].BaseDir)
	}
	if got := first[0].Resources.Scripts; len(got) != 1 || got[0] != "scripts/run.sh" {
		t.Fatalf("unexpected immutable scripts: %#v", got)
	}
	if got := first[0].Resources.References; len(got) != 1 || got[0] != "references/guide.md" {
		t.Fatalf("unexpected immutable references: %#v", got)
	}
	if got := first[0].Resources.Assets; len(got) != 2 || got[0] != "assets/nested/data.md" || got[1] != "assets/template.json" {
		t.Fatalf("unexpected immutable assets: %#v", got)
	}

	source["ops/SKILL.md"] = &fstest.MapFile{Data: []byte("---\nname: ops\ndescription: mutated without identity change\n---\n# Ops\n")}
	second, secondReport, err := LoadWithReport(opts)
	if err != nil {
		t.Fatalf("second immutable load: %v", err)
	}
	if !secondReport.CacheHit || secondReport.Generation != firstReport.Generation {
		t.Fatalf("stable immutable identity should reuse cache snapshot: first=%+v second=%+v", firstReport, secondReport)
	}
	if len(second) != 1 || second[0].Description != "embedded v1" {
		t.Fatalf("cache snapshot should remain detached from source mutation, got %#v", second)
	}

	changedProvider, err := agentxassetfs.New(provider.ID(), source)
	if err != nil {
		t.Fatalf("new changed provider: %v", err)
	}
	opts.BundledFS = FSSource{
		ID:          changedProvider.ID(),
		FS:          changedProvider.FS(),
		Fingerprint: changedProvider.Fingerprint(),
	}
	third, thirdReport, err := LoadWithReport(opts)
	if err != nil {
		t.Fatalf("third immutable load: %v", err)
	}
	if thirdReport.CacheHit || len(third) != 1 || third[0].Description != "mutated without identity change" {
		t.Fatalf("fingerprint change should select a fresh immutable snapshot: items=%#v report=%+v", third, thirdReport)
	}
}

func TestLoadWithReport_UnattestedFSSourceCannotPolluteCache(t *testing.T) {
	source := fstest.MapFS{
		"ops/SKILL.md": &fstest.MapFile{Data: []byte("---\nname: ops\ndescription: v1\n---\n# Ops\n")},
	}
	opts := LoadOptions{
		ExtraFS: []FSSource{{
			ID:          "agentx.test.unattested." + t.Name(),
			FS:          source,
			Fingerprint: "sha256:caller-assertion",
		}},
	}
	first, firstReport, err := LoadWithReport(opts)
	if err != nil {
		t.Fatalf("first unattested load: %v", err)
	}
	if firstReport.CacheHit || firstReport.Generation != 0 || len(first) != 1 || first[0].Description != "v1" {
		t.Fatalf("unattested source must load without cache: items=%#v report=%+v", first, firstReport)
	}

	source["ops/SKILL.md"] = &fstest.MapFile{Data: []byte("---\nname: ops\ndescription: v2\n---\n# Ops\n")}
	second, secondReport, err := LoadWithReport(opts)
	if err != nil {
		t.Fatalf("second unattested load: %v", err)
	}
	if secondReport.CacheHit || secondReport.Generation != 0 || len(second) != 1 || second[0].Description != "v2" {
		t.Fatalf("unattested source must not return stale cached content: items=%#v report=%+v", second, secondReport)
	}
	if _, ok := LoadGeneration(opts); ok {
		t.Fatal("unattested source must not claim an immutable loader generation")
	}
}

func TestLoadWithReport_InvalidImmutableFSSource(t *testing.T) {
	opts := LoadOptions{
		ExtraFS: []FSSource{{
			ID: "missing-fingerprint",
			FS: fstest.MapFS{},
		}},
	}
	if opts.ExtraFS[0].Valid() {
		t.Fatal("source without fingerprint must be invalid")
	}
	items, report, err := LoadWithReport(opts)
	if err != nil {
		t.Fatalf("tolerant invalid immutable load: %v", err)
	}
	if len(items) != 0 || len(report.Issues) != 1 || report.Issues[0].Code != "invalid_fs_source" {
		t.Fatalf("expected invalid source diagnostic, items=%#v report=%+v", items, report)
	}
	opts.FailFast = true
	if _, _, err := LoadWithReport(opts); err == nil {
		t.Fatal("fail-fast invalid immutable load should fail")
	}
}

func TestLoadFromDirs_ScanSkillResources(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "ops")
	if err := os.MkdirAll(filepath.Join(skillDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(skillDir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(skillDir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: ops\ndescription: ops skill\n---\n# Ops\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "scripts", "run.sh"), []byte("#!/bin/sh\necho ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "references", "guide.md"), []byte("# guide\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "assets", "template.txt"), []byte("template\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := LoadFromDirs([]string{dir})
	if err != nil {
		t.Fatalf("load skills: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(items))
	}
	skill := items[0]
	if skill.BaseDir == "" {
		t.Fatalf("expected base dir")
	}
	if len(skill.Resources.Scripts) != 1 || skill.Resources.Scripts[0] != "scripts/run.sh" {
		t.Fatalf("unexpected scripts: %#v", skill.Resources.Scripts)
	}
	if len(skill.Resources.References) != 1 || skill.Resources.References[0] != "references/guide.md" {
		t.Fatalf("unexpected references: %#v", skill.Resources.References)
	}
	if len(skill.Resources.Assets) != 1 || skill.Resources.Assets[0] != "assets/template.txt" {
		t.Fatalf("unexpected assets: %#v", skill.Resources.Assets)
	}
}

func TestLoadFromDirsWithReport_CacheHitWhenUnchanged(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "weather")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: weather\ndescription: cached skill\n---\n# Weather\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	first, firstReport, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
	if err != nil {
		t.Fatalf("first load skills: %v", err)
	}
	if firstReport.CacheHit {
		t.Fatalf("expected first load to miss cache, got %+v", firstReport)
	}
	if firstReport.Generation == 0 {
		t.Fatalf("expected first load to record a generation, got %+v", firstReport)
	}
	second, secondReport, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
	if err != nil {
		t.Fatalf("second load skills: %v", err)
	}
	if !secondReport.CacheHit {
		t.Fatalf("expected second load to hit cache, got %+v", secondReport)
	}
	if secondReport.Generation != firstReport.Generation {
		t.Fatalf("expected unchanged load generation to remain stable, got first=%d second=%d", firstReport.Generation, secondReport.Generation)
	}
	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("expected one skill from both loads, got %d and %d", len(first), len(second))
	}
	if second[0].Description != "cached skill" {
		t.Fatalf("unexpected cached skill description: %#v", second[0])
	}
}

func TestLoadFromDirsWithReport_CacheInvalidatesOnSkillDocChange(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "weather")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skillPath := filepath.Join(skillDir, "SKILL.md")
	writeSkill := func(description string, modTime time.Time) {
		t.Helper()
		content := "---\nname: weather\ndescription: " + description + "\n---\n# Weather\n"
		if err := os.WriteFile(skillPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(skillPath, modTime, modTime); err != nil {
			t.Fatal(err)
		}
	}

	firstTime := time.Now().Add(-2 * time.Second)
	writeSkill("first version", firstTime)
	first, firstReport, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
	if err != nil {
		t.Fatalf("first load skills: %v", err)
	}
	if firstReport.CacheHit {
		t.Fatalf("expected first load cache miss, got %+v", firstReport)
	}
	if firstReport.Generation == 0 {
		t.Fatalf("expected first load generation, got %+v", firstReport)
	}
	if first[0].Description != "first version" {
		t.Fatalf("unexpected first description: %#v", first[0])
	}

	writeSkill("second version", firstTime.Add(2*time.Second))
	second, secondReport, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
	if err != nil {
		t.Fatalf("second load skills: %v", err)
	}
	if secondReport.CacheHit {
		t.Fatalf("expected cache invalidation after skill doc change, got %+v", secondReport)
	}
	if secondReport.Generation <= firstReport.Generation {
		t.Fatalf("expected generation to advance after skill change, got first=%d second=%d", firstReport.Generation, secondReport.Generation)
	}
	if second[0].Description != "second version" {
		t.Fatalf("expected updated description after invalidation, got %#v", second[0])
	}

	third, thirdReport, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
	if err != nil {
		t.Fatalf("third load skills: %v", err)
	}
	if thirdReport.Generation < secondReport.Generation {
		t.Fatalf("expected generation to remain monotonic after refresh, got second=%d third=%d", secondReport.Generation, thirdReport.Generation)
	}
	if third[0].Description != "second version" {
		t.Fatalf("expected refreshed cached description, got %#v", third[0])
	}
	if !thirdReport.CacheHit {
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			fourth, fourthReport, loadErr := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
			if loadErr != nil {
				t.Fatalf("fourth load skills: %v", loadErr)
			}
			if fourth[0].Description != "second version" {
				t.Fatalf("expected refreshed cached description after conservative invalidation, got %#v", fourth[0])
			}
			if fourthReport.CacheHit {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Fatalf("expected cache to settle after conservative invalidation, got third=%+v", thirdReport)
	}
}

func TestLoadFromDirsWithReport_KeepsCacheOnUnrelatedSourceRootChanges(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "weather")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte("---\nname: weather\ndescription: stable skill\n---\n# Weather\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	first, firstReport, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
	if err != nil {
		t.Fatalf("first load skills: %v", err)
	}
	if firstReport.CacheHit {
		t.Fatalf("expected first load cache miss, got %+v", firstReport)
	}
	if firstReport.Generation == 0 {
		t.Fatalf("expected first load generation, got %+v", firstReport)
	}
	if len(first) != 1 || first[0].Description != "stable skill" {
		t.Fatalf("unexpected first load: %#v", first)
	}

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, secondReport, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
	if err != nil {
		t.Fatalf("second load skills: %v", err)
	}
	if !secondReport.CacheHit {
		t.Fatalf("expected cache hit after unrelated root file change, got %+v", secondReport)
	}
	if secondReport.Generation != firstReport.Generation {
		t.Fatalf("expected generation to remain stable after unrelated root file change, got first=%d second=%d", firstReport.Generation, secondReport.Generation)
	}
	if len(second) != 1 || second[0].Description != "stable skill" {
		t.Fatalf("unexpected second load: %#v", second)
	}

	if err := os.MkdirAll(filepath.Join(dir, "notes"), 0o755); err != nil {
		t.Fatal(err)
	}
	third, thirdReport, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
	if err != nil {
		t.Fatalf("third load skills: %v", err)
	}
	if !thirdReport.CacheHit {
		t.Fatalf("expected cache hit after unrelated root dir change, got %+v", thirdReport)
	}
	if thirdReport.Generation != firstReport.Generation {
		t.Fatalf("expected generation to remain stable after unrelated root dir change, got first=%d third=%d", firstReport.Generation, thirdReport.Generation)
	}
	if len(third) != 1 || third[0].Description != "stable skill" {
		t.Fatalf("unexpected third load: %#v", third)
	}
}

func TestLoadFromDirsWithReport_CacheInvalidatesOnResourceListChange(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "ops")
	if err := os.MkdirAll(filepath.Join(skillDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: ops\ndescription: ops skill\n---\n# Ops\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	first, firstReport, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
	if err != nil {
		t.Fatalf("first load skills: %v", err)
	}
	if firstReport.CacheHit {
		t.Fatalf("expected first load cache miss, got %+v", firstReport)
	}
	if len(first) != 1 || len(first[0].Resources.Scripts) != 0 {
		t.Fatalf("expected no scripts initially, got %#v", first)
	}

	scriptPath := filepath.Join(skillDir, "scripts", "run.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\necho ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, secondReport, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
	if err != nil {
		t.Fatalf("second load skills: %v", err)
	}
	if secondReport.CacheHit {
		t.Fatalf("expected cache invalidation after resource list change, got %+v", secondReport)
	}
	if len(second) != 1 || len(second[0].Resources.Scripts) != 1 || second[0].Resources.Scripts[0] != "scripts/run.sh" {
		t.Fatalf("expected updated script resources after invalidation, got %#v", second)
	}
}

func TestLoadFromDirsWithReport_CacheInvalidatesOnNestedResourceDirectoryChange(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "ops")
	if err := os.MkdirAll(filepath.Join(skillDir, "scripts", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: ops\ndescription: ops skill\n---\n# Ops\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	first, firstReport, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
	if err != nil {
		t.Fatalf("first load skills: %v", err)
	}
	if firstReport.CacheHit {
		t.Fatalf("expected first load cache miss, got %+v", firstReport)
	}
	if len(first) != 1 || len(first[0].Resources.Scripts) != 0 {
		t.Fatalf("expected no scripts initially, got %#v", first)
	}

	scriptPath := filepath.Join(skillDir, "scripts", "nested", "run.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\necho ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, secondReport, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
	if err != nil {
		t.Fatalf("second load skills: %v", err)
	}
	if secondReport.CacheHit {
		t.Fatalf("expected cache invalidation after nested resource change, got %+v", secondReport)
	}
	if len(second) != 1 || len(second[0].Resources.Scripts) != 1 || second[0].Resources.Scripts[0] != "scripts/nested/run.sh" {
		t.Fatalf("expected updated nested script resources after invalidation, got %#v", second)
	}
}

func TestLoadFromDirsWithReport_KeepsCacheOnResourceDirMetadataChange(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "ops")
	scriptsDir := filepath.Join(skillDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: ops\ndescription: ops skill\n---\n# Ops\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	origFactory := loadSourceWatcherFactory
	t.Cleanup(func() {
		loadSourceWatcherFactory = origFactory
	})
	loadSourceWatcherFactory = func(_ []loadSource, _ *loadGenerationEntry) (*loadSourceWatcher, error) {
		return &loadSourceWatcher{dirs: map[string]struct{}{}}, nil
	}

	first, firstReport, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
	if err != nil {
		t.Fatalf("first load skills: %v", err)
	}
	if firstReport.CacheHit {
		t.Fatalf("expected first load cache miss, got %+v", firstReport)
	}
	if len(first) != 1 || len(first[0].Resources.Scripts) != 0 {
		t.Fatalf("expected no scripts initially, got %#v", first)
	}

	updated := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(scriptsDir, updated, updated); err != nil {
		t.Fatalf("touch scripts dir: %v", err)
	}

	second, secondReport, err := LoadFromDirsWithReport([]string{dir}, LoadOptions{})
	if err != nil {
		t.Fatalf("second load skills: %v", err)
	}
	if !secondReport.CacheHit {
		t.Fatalf("expected cache hit after resource dir metadata change, got %+v", secondReport)
	}
	if secondReport.Generation != firstReport.Generation {
		t.Fatalf("expected generation to remain stable after resource dir metadata change, got first=%d second=%d", firstReport.Generation, secondReport.Generation)
	}
	if len(second) != 1 || len(second[0].Resources.Scripts) != 0 {
		t.Fatalf("expected resources to remain unchanged, got %#v", second)
	}
}

func TestLoadGeneration_WatcherAdvancesAfterSkillChange(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "weather")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte("---\nname: weather\ndescription: watch skill\n---\n# Weather\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	opts := LoadOptions{CustomDirs: []string{dir}}
	initial, ok := LoadGeneration(opts)
	if !ok || initial == 0 {
		t.Fatalf("expected initial generation, got %d ok=%v", initial, ok)
	}
	updated := time.Now().Add(2 * time.Second)
	if err := os.WriteFile(skillPath, []byte("---\nname: weather\ndescription: watch skill updated\n---\n# Weather\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(skillPath, updated, updated); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		current, ok := LoadGeneration(opts)
		if ok && current > initial {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	current, ok := LoadGeneration(opts)
	t.Fatalf("expected watcher-backed generation to advance after skill change, initial=%d current=%d ok=%v", initial, current, ok)
}

func TestLoadGeneration_RetriesWatcherCreationAfterFailure(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "weather")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: weather\ndescription: retry watcher\n---\n# Weather\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	origFactory := loadSourceWatcherFactory
	origBackoff := loadWatcherRetryBackoff
	t.Cleanup(func() {
		loadSourceWatcherFactory = origFactory
		loadWatcherRetryBackoff = origBackoff
	})

	attempts := 0
	loadSourceWatcherFactory = func(_ []loadSource, _ *loadGenerationEntry) (*loadSourceWatcher, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("boom")
		}
		return &loadSourceWatcher{dirs: map[string]struct{}{}}, nil
	}
	loadWatcherRetryBackoff = 0

	opts := LoadOptions{CustomDirs: []string{dir}}
	if generation, ok := LoadGeneration(opts); !ok || generation == 0 {
		t.Fatalf("expected initial generation, got %d ok=%v", generation, ok)
	}
	if attempts != 1 {
		t.Fatalf("expected first watcher attempt to run once, got %d", attempts)
	}

	sources := buildLoadSources(opts)
	key, ok := buildLoadCacheLookupKey(sources, opts)
	if !ok {
		t.Fatalf("expected load cache lookup key")
	}
	entry, found := getLoadGenerationEntry(key)
	if !found || entry == nil {
		t.Fatalf("expected generation entry after first load")
	}
	if entry.watcher != nil {
		t.Fatalf("expected watcher to be absent after first failure")
	}

	if generation, ok := LoadGeneration(opts); !ok || generation == 0 {
		t.Fatalf("expected second generation read, got %d ok=%v", generation, ok)
	}
	if attempts < 2 {
		t.Fatalf("expected watcher creation to retry, got %d attempts", attempts)
	}
	if entry.watcher == nil {
		t.Fatalf("expected watcher to be installed after retry")
	}
}

func TestLoadSourceWatcherSkipsBundledSource(t *testing.T) {
	bundled := t.TempDir()
	bundledSkillDir := filepath.Join(bundled, "weather")
	if err := os.MkdirAll(filepath.Join(bundledSkillDir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundledSkillDir, "SKILL.md"), []byte("---\nname: weather\ndescription: bundled source\n---\n# Weather\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundledSkillDir, "references", "provider.md"), []byte("provider"), 0o644); err != nil {
		t.Fatal(err)
	}

	watcher, err := newLoadSourceWatcher([]loadSource{{Kind: SourceBundled, Dir: bundled}}, &loadGenerationEntry{})
	if err != nil {
		t.Fatalf("new bundled watcher: %v", err)
	}
	defer watcher.Close()
	if watcher.watcher != nil {
		t.Fatalf("bundled-only source should use validation without opening an fsnotify watcher")
	}
	if len(watcher.dirs) != 0 || len(watcher.roots) != 0 {
		t.Fatalf("expected bundled source to be absent from watcher roots, got dirs=%#v roots=%#v", watcher.dirs, watcher.roots)
	}
}

func TestLoadSourceWatcherKeepsMutableSources(t *testing.T) {
	bundled := t.TempDir()
	if err := os.MkdirAll(filepath.Join(bundled, "weather"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundled, "weather", "SKILL.md"), []byte("---\nname: weather\ndescription: bundled source\n---\n# Weather\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	custom := t.TempDir()
	if err := os.MkdirAll(filepath.Join(custom, "ops"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(custom, "ops", "SKILL.md"), []byte("---\nname: ops\ndescription: custom source\n---\n# Ops\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	watcher, err := newLoadSourceWatcher([]loadSource{
		{Kind: SourceBundled, Dir: bundled},
		{Kind: SourceCustom, Dir: custom},
	}, &loadGenerationEntry{})
	if err != nil {
		t.Fatalf("new mixed watcher: %v", err)
	}
	defer watcher.Close()
	if watcher.watcher == nil {
		t.Fatalf("custom source should still install an fsnotify watcher")
	}
	if _, ok := watcher.dirs[filepath.Clean(bundled)]; ok {
		t.Fatalf("bundled source should not be watched, dirs=%#v", watcher.dirs)
	}
	if _, ok := watcher.dirs[filepath.Clean(custom)]; !ok {
		t.Fatalf("custom source root should be watched, dirs=%#v", watcher.dirs)
	}
}

func TestLoadGeneration_RecreatesWatcherAfterWatcherExit(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "weather")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: weather\ndescription: revive watcher\n---\n# Weather\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	origFactory := loadSourceWatcherFactory
	t.Cleanup(func() {
		loadSourceWatcherFactory = origFactory
	})

	attempts := 0
	loadSourceWatcherFactory = func(_ []loadSource, entry *loadGenerationEntry) (*loadSourceWatcher, error) {
		attempts++
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return nil, err
		}
		loadWatcher := &loadSourceWatcher{watcher: watcher, dirs: map[string]struct{}{}}
		go loadWatcher.run(entry)
		if attempts == 1 {
			go func() {
				time.Sleep(20 * time.Millisecond)
				_ = watcher.Close()
			}()
		}
		return loadWatcher, nil
	}

	opts := LoadOptions{CustomDirs: []string{dir}}
	if generation, ok := LoadGeneration(opts); !ok || generation == 0 {
		t.Fatalf("expected initial generation, got %d ok=%v", generation, ok)
	}
	sources := buildLoadSources(opts)
	key, ok := buildLoadCacheLookupKey(sources, opts)
	if !ok {
		t.Fatalf("expected load cache lookup key")
	}
	entry, found := getLoadGenerationEntry(key)
	if !found || entry == nil {
		t.Fatalf("expected generation entry")
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		entry.mu.RLock()
		watcherMissing := entry.watcher == nil
		entry.mu.RUnlock()
		if watcherMissing {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	entry.mu.RLock()
	watcherMissing := entry.watcher == nil
	entry.mu.RUnlock()
	if !watcherMissing {
		t.Fatalf("expected first watcher to clear after exit")
	}

	if generation, ok := LoadGeneration(opts); !ok || generation == 0 {
		t.Fatalf("expected second generation read, got %d ok=%v", generation, ok)
	}
	if attempts < 2 {
		t.Fatalf("expected watcher recreation after exit, got %d attempts", attempts)
	}
	entry.mu.RLock()
	defer entry.mu.RUnlock()
	if entry.watcher == nil {
		t.Fatalf("expected watcher to be recreated after exit")
	}
}

func TestLoadGeneration_FastValidationSkipsFullFingerprintWhenWatcherHealthy(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "weather")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: weather\ndescription: fast validation\n---\n# Weather\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	origFactory := loadSourceWatcherFactory
	origBuilder := loadFingerprintBuilder
	t.Cleanup(func() {
		loadSourceWatcherFactory = origFactory
		loadFingerprintBuilder = origBuilder
	})

	loadSourceWatcherFactory = func(_ []loadSource, _ *loadGenerationEntry) (*loadSourceWatcher, error) {
		return &loadSourceWatcher{dirs: map[string]struct{}{}}, nil
	}

	builderCalls := 0
	loadFingerprintBuilder = func(sources []loadSource, opts LoadOptions) (string, []loadSourceValidationState, bool) {
		builderCalls++
		return buildLoadFingerprint(sources, opts)
	}

	opts := LoadOptions{CustomDirs: []string{dir}}
	first, ok := LoadGeneration(opts)
	if !ok || first == 0 {
		t.Fatalf("expected initial generation, got %d ok=%v", first, ok)
	}
	second, ok := LoadGeneration(opts)
	if !ok || second != first {
		t.Fatalf("expected stable generation on second load, first=%d second=%d ok=%v", first, second, ok)
	}
	if builderCalls != 1 {
		t.Fatalf("expected healthy watcher fast validation to skip second full fingerprint, got %d builder calls", builderCalls)
	}
}
