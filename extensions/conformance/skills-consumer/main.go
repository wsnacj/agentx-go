package main

import (
	"fmt"
	"testing/fstest"

	"github.com/wsnacj/agentx-go/extensions/skills"
	"github.com/wsnacj/agentx-go/runtime/assetfs"
)

func run() (string, error) {
	provider, err := assetfs.New("agentx.example.skills", fstest.MapFS{
		"portable-review/SKILL.md":                {Data: []byte("---\nname: portable-review\ndescription: Review repository boundaries\npaths: [src/**]\ncontext: fork\nallowed_tools: [Read, edit, read]\neffort: high\n---\n# Repository review\nRead references/checklist.md before reviewing.\n")},
		"portable-review/references/checklist.md": {Data: []byte("- verify owner boundaries\n")},
	})
	if err != nil {
		return "", fmt.Errorf("create immutable skill source: %w", err)
	}

	opts := skills.LoadOptions{
		BundledFS: skills.FSSource{
			ID:          provider.ID(),
			FS:          provider.FS(),
			Fingerprint: provider.Fingerprint(),
		},
		StrictFrontmatter:        true,
		MaxCandidatesPerRoot:     4,
		MaxSkillsLoadedPerSource: 4,
		MaxSkillFileBytes:        64 * 1024,
	}
	loaded, firstReport, err := skills.LoadWithReport(opts)
	if err != nil {
		return "", fmt.Errorf("load skills: %w", err)
	}
	if len(loaded) != 1 || loaded[0].Source != skills.SourceBundled || firstReport.HasIssues() || firstReport.CacheHit || firstReport.Generation == 0 {
		return "", fmt.Errorf("unexpected first load: skills=%d report=%+v", len(loaded), firstReport)
	}

	second, secondReport, err := skills.LoadWithReport(opts)
	if err != nil {
		return "", fmt.Errorf("reload skills: %w", err)
	}
	if len(second) != 1 || !secondReport.CacheHit || secondReport.Generation != firstReport.Generation {
		return "", fmt.Errorf("immutable cache was not reused: first=%+v second=%+v", firstReport, secondReport)
	}

	active, reason, _ := skills.EvaluateSkillActivationPaths(second[0], []string{"src/main.go"})
	if !active {
		return "", fmt.Errorf("skill is not active: %s", reason)
	}
	semantics := skills.ResolveRequestedSkillSemantics(second, []string{"portable-review"})
	if len(semantics) != 1 || semantics[0].ExecutionContext != skills.SkillExecutionContextFork || len(semantics[0].AllowedTools) != 2 {
		return "", fmt.Errorf("unexpected requested semantics: %#v", semantics)
	}
	if missing := skills.MissingReferencedResourcePaths(second[0]); len(missing) != 0 {
		return "", fmt.Errorf("missing referenced resources: %v", missing)
	}
	cloned := skills.Clone(second)
	cloned[0].Name = "mutated"
	if second[0].Name != "portable-review" {
		return "", fmt.Errorf("clone mutated source skill: %#v", second[0])
	}

	return fmt.Sprintf(
		"agentx-skills-ok:%s:%s:%t:%t:%s:%d",
		second[0].Name,
		second[0].Source,
		secondReport.CacheHit,
		active,
		semantics[0].ExecutionContext,
		len(semantics[0].AllowedTools),
	), nil
}

func main() {
	result, err := run()
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
