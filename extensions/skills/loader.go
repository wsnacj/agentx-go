package skills

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const skillDocName = "SKILL.md"

type skillDocTooLargeError struct {
	size     int64
	maxBytes int
}

func (e skillDocTooLargeError) Error() string {
	return fmt.Sprintf("skill doc exceeds max size (%d bytes > %d)", e.size, e.maxBytes)
}

type loadSource struct {
	Kind        Source
	Dir         string
	FS          fs.FS
	ID          string
	Fingerprint string
	FSBacked    bool
}

func Load(opts LoadOptions) ([]Skill, error) {
	items, _, err := LoadWithReport(opts)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func LoadWithReport(opts LoadOptions) ([]Skill, LoadReport, error) {
	return loadFromSourcesCached(buildLoadSources(opts), opts)
}

func LoadFromDirs(dirs []string) ([]Skill, error) {
	items, _, err := LoadFromDirsWithReport(dirs, LoadOptions{})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func LoadFromDirsWithReport(dirs []string, opts LoadOptions) ([]Skill, LoadReport, error) {
	sources := make([]loadSource, 0, len(dirs))
	for _, dir := range dirs {
		if resolved := normalizeLoadSourceDir(dir); resolved != "" {
			sources = append(sources, loadSource{Kind: SourceCustom, Dir: resolved})
		}
	}
	return loadFromSourcesCached(sources, opts)
}

func buildLoadSources(opts LoadOptions) []loadSource {
	sources := make([]loadSource, 0)
	for _, dir := range opts.ExtraDirs {
		if resolved := normalizeLoadSourceDir(dir); resolved != "" {
			sources = append(sources, loadSource{Kind: SourceExtra, Dir: resolved})
		}
	}
	for _, source := range opts.ExtraFS {
		if isZeroFSSource(source) {
			continue
		}
		sources = append(sources, newFSLoadSource(SourceExtra, source))
	}
	if bundled := normalizeLoadSourceDir(opts.BundledDir); bundled != "" {
		sources = append(sources, loadSource{Kind: SourceBundled, Dir: bundled})
	}
	if !isZeroFSSource(opts.BundledFS) {
		sources = append(sources, newFSLoadSource(SourceBundled, opts.BundledFS))
	}
	if managed := normalizeLoadSourceDir(opts.ManagedDir); managed != "" {
		sources = append(sources, loadSource{Kind: SourceManaged, Dir: managed})
	}
	for _, dir := range opts.CustomDirs {
		if resolved := normalizeLoadSourceDir(dir); resolved != "" {
			sources = append(sources, loadSource{Kind: SourceCustom, Dir: resolved})
		}
	}
	if workspace := normalizeLoadSourceDir(opts.WorkspaceDir); workspace != "" {
		sources = append(sources, loadSource{Kind: SourceWorkspace, Dir: workspace})
	}
	return sources
}

func loadFromSources(sources []loadSource, opts LoadOptions) ([]Skill, LoadReport, error) {
	report := LoadReport{}
	merged := make(map[string]Skill)
	for _, source := range sources {
		if source.FSBacked && !source.fsSource().Valid() {
			err := fmt.Errorf("immutable skill source requires id, filesystem, and fingerprint")
			if opts.FailFast {
				return nil, report, err
			}
			report.Issues = append(report.Issues, LoadIssue{
				Code:    "invalid_fs_source",
				Source:  source.Kind,
				Path:    loadSourceDisplayPath(source, "."),
				Stage:   "read_dir",
				Message: err.Error(),
			})
			continue
		}
		entries, err := readLoadSourceDir(source)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			if opts.FailFast {
				return nil, report, err
			}
			report.Issues = append(report.Issues, LoadIssue{
				Code:    "read_dir_failed",
				Source:  source.Kind,
				Path:    loadSourceDisplayPath(source, "."),
				Stage:   "read_dir",
				Message: err.Error(),
			})
			continue
		}
		loadedForSource := 0
		candidatesSeen := 0
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if !isLoadableSkillDir(entry.Name()) {
				continue
			}
			if opts.MaxCandidatesPerRoot > 0 && candidatesSeen >= opts.MaxCandidatesPerRoot {
				report.Skipped++
				report.Issues = append(report.Issues, LoadIssue{
					Code:    "candidate_limit_reached",
					Source:  source.Kind,
					Path:    loadSourceDisplayPath(source, entry.Name()),
					Stage:   "scan_dir",
					Message: fmt.Sprintf("skill candidate limit reached for source (%d)", opts.MaxCandidatesPerRoot),
				})
				continue
			}
			candidatesSeen++
			if opts.MaxSkillsLoadedPerSource > 0 && loadedForSource >= opts.MaxSkillsLoadedPerSource {
				report.Skipped++
				report.Issues = append(report.Issues, LoadIssue{
					Code:    "loaded_limit_reached",
					Source:  source.Kind,
					Path:    loadSourceDisplayPath(source, entry.Name()),
					Stage:   "load_skill",
					Message: fmt.Sprintf("skill load limit reached for source (%d)", opts.MaxSkillsLoadedPerSource),
				})
				continue
			}
			skillPath := loadSourceDisplayPath(source, path.Join(entry.Name(), skillDocName))
			skill, ok, err := loadSkillFromSource(source, entry.Name(), opts.StrictFrontmatter, opts.MaxSkillFileBytes)
			if err != nil {
				code := "skill_load_failed"
				if _, ok := err.(skillDocTooLargeError); ok {
					code = "skill_doc_too_large"
				} else {
					report.ParseFailed++
				}
				report.Skipped++
				report.Issues = append(report.Issues, LoadIssue{
					Code:    code,
					Source:  source.Kind,
					Path:    skillPath,
					Stage:   "load_skill",
					Message: err.Error(),
				})
				if opts.FailFast {
					return nil, report, err
				}
				continue
			}
			if !ok {
				report.Skipped++
				continue
			}
			if skill.Name == "" {
				skill.Name = entry.Name()
			}
			merged[strings.ToLower(skill.Name)] = skill
			loadedForSource++
		}
	}

	out := make([]Skill, 0, len(merged))
	for _, s := range merged {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	report.Loaded = len(out)
	return out, report, nil
}

func newFSLoadSource(kind Source, source FSSource) loadSource {
	return loadSource{
		Kind:        kind,
		FS:          source.FS,
		ID:          strings.TrimSpace(source.ID),
		Fingerprint: strings.TrimSpace(source.Fingerprint),
		FSBacked:    true,
	}
}

func isZeroFSSource(source FSSource) bool {
	return strings.TrimSpace(source.ID) == "" && source.FS == nil && strings.TrimSpace(source.Fingerprint) == ""
}

func (s loadSource) fsSource() FSSource {
	return FSSource{ID: s.ID, FS: s.FS, Fingerprint: s.Fingerprint}
}

func readLoadSourceDir(source loadSource) ([]fs.DirEntry, error) {
	if source.FSBacked {
		return fs.ReadDir(source.FS, ".")
	}
	return os.ReadDir(source.Dir)
}

func loadSourceDisplayPath(source loadSource, name string) string {
	if !source.FSBacked {
		if name == "." {
			return source.Dir
		}
		return filepath.Join(source.Dir, filepath.FromSlash(name))
	}
	root := "assetfs://" + strings.TrimSpace(source.ID)
	cleaned := path.Clean(strings.TrimSpace(name))
	if cleaned == "." {
		return root
	}
	return root + "/" + cleaned
}

func loadSkillFromSource(source loadSource, skillName string, strictFrontmatter bool, maxBytes int) (Skill, bool, error) {
	if !source.FSBacked {
		return loadSkillFile(filepath.Join(source.Dir, skillName, skillDocName), source.Kind, strictFrontmatter, maxBytes)
	}
	skillDir := path.Clean(skillName)
	skillPath := path.Join(skillDir, skillDocName)
	if maxBytes > 0 {
		info, err := fs.Stat(source.FS, skillPath)
		if err != nil {
			if os.IsNotExist(err) {
				return Skill{}, false, nil
			}
			return Skill{}, false, err
		}
		if info.Size() > int64(maxBytes) {
			return Skill{}, false, skillDocTooLargeError{size: info.Size(), maxBytes: maxBytes}
		}
	}
	raw, err := readFSFileBounded(source.FS, skillPath, maxBytes)
	if err != nil {
		if os.IsNotExist(err) {
			return Skill{}, false, nil
		}
		return Skill{}, false, err
	}
	return parseSkillFile(
		raw,
		loadSourceDisplayPath(source, skillPath),
		loadSourceDisplayPath(source, skillDir),
		source.Kind,
		strictFrontmatter,
		func() (Resources, error) {
			return scanSkillResourcesFS(source.FS, skillDir)
		},
	)
}

func readFSFileBounded(source fs.FS, name string, maxBytes int) ([]byte, error) {
	if source == nil {
		return nil, fs.ErrInvalid
	}
	file, err := source.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if maxBytes <= 0 {
		return io.ReadAll(file)
	}
	content, err := io.ReadAll(io.LimitReader(file, int64(maxBytes)+1))
	if err != nil {
		return nil, err
	}
	if len(content) > maxBytes {
		return nil, skillDocTooLargeError{size: int64(len(content)), maxBytes: maxBytes}
	}
	return content, nil
}

func loadSkillFile(path string, source Source, strictFrontmatter bool, maxBytes int) (Skill, bool, error) {
	if maxBytes > 0 {
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return Skill{}, false, nil
			}
			return Skill{}, false, err
		}
		if info.Size() > int64(maxBytes) {
			return Skill{}, false, skillDocTooLargeError{size: info.Size(), maxBytes: maxBytes}
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Skill{}, false, nil
		}
		return Skill{}, false, err
	}
	return parseSkillFile(
		raw,
		path,
		filepath.Dir(path),
		source,
		strictFrontmatter,
		func() (Resources, error) {
			return scanSkillResources(filepath.Dir(path))
		},
	)
}

func parseSkillFile(
	raw []byte,
	location string,
	baseDir string,
	source Source,
	strictFrontmatter bool,
	scanResources func() (Resources, error),
) (Skill, bool, error) {
	fm, body, err := parseFrontMatter(string(raw), strictFrontmatter)
	if err != nil {
		return Skill{}, false, err
	}
	s := Skill{
		Location:   location,
		BaseDir:    baseDir,
		Source:     source,
		Metadata:   map[string]string{},
		Content:    strings.TrimSpace(body),
		Invocation: InvocationPolicy{UserInvocable: true},
	}
	if fm != nil {
		compat, compatErr := parseCompatMetadata(fm["metadata"])
		if compatErr != nil {
			return Skill{}, false, fmt.Errorf("parse metadata for %s: %w", location, compatErr)
		}
		s.Name = readStringAny(fm["name"])
		s.Description = readStringAny(fm["description"])
		s.Keywords = readStringListWithAliases(fm, "keywords", "keyword")
		s.Tags = readStringListWithAliases(fm, "tags", "tag")
		s.WhenToUse = readStringListWithAliases(fm, "when_to_use", "when-to-use", "whenToUse")
		s.WhenNotToUse = readStringListWithAliases(fm, "when_not_to_use", "when-not-to-use", "whenNotToUse")
		s.NegativeExamples = readStringListWithAliases(fm, "negative_examples", "negative-examples", "negativeExamples")
		s.Steps = readStringListWithAliases(fm, "steps", "workflow")
		s.Paths = NormalizeSkillPathScopes(readStringListWithAliases(fm, "paths", "path", "path_scopes", "path-scopes", "pathScopes"))
		s.ToolHints = readStringListWithAliases(fm, "tool_hints", "tool-hints", "toolHints")
		s.ToolHintsMatch = NormalizeToolHintsMatch(readStringAny(firstNonNil(
			fm["tool_hints_match"],
			fm["tool-hints-match"],
			fm["toolHintsMatch"],
		)))
		context, allowedTools, effort, topLevelExecution := parseSkillExecutionSemantics(fm)
		s.ExecutionContext = context
		s.AllowedTools = allowedTools
		s.Effort = effort
		s.Examples = readStringListWithAliases(fm, "examples", "example_prompts", "examplePrompts")
		s.EvalAssertions = readStringListWithAliases(fm, "eval_assertions", "eval-assertions", "evalAssertions")
		invocation, topLevelInvocation := parseInvocationPolicy(fm)
		dispatch, topLevelDispatch := parseDispatchSpec(fm)
		s.Invocation = invocation
		s.Dispatch = dispatch
		s.OS = readStringListWithAliases(fm, "os", "platforms")
		s.Requires = parseRequires(firstNonNil(fm["requires"], fm["require"]))
		install, installErr := parseInstallSpecs(fm["install"])
		if installErr != nil {
			return Skill{}, false, fmt.Errorf("parse install specs for %s: %w", location, installErr)
		}
		s.Install = install
		for key, value := range fm {
			switch strings.ToLower(strings.TrimSpace(key)) {
			case "name", "description", "keywords", "keyword", "tags", "tag",
				"when_to_use", "when_not_to_use", "negative_examples",
				"steps", "paths", "path", "path_scopes", "tool_hints", "tool_hints_match", "examples", "eval_assertions",
				"context", "skill_context", "allowed_tools", "effort", "reasoning_effort",
				"when-to-use", "whentouse", "when-not-to-use", "whennottouse",
				"negative-examples", "negativeexamples", "path-scopes", "pathscopes", "tool-hints", "toolhints",
				"tool-hints-match", "toolhintsmatch",
				"skillcontext", "allowed-tools", "allowedtools", "reasoningeffort",
				"example_prompts", "exampleprompts", "eval-assertions", "evalassertions",
				"user-invocable", "user_invocable", "disable-model-invocation", "disable_model_invocation",
				"userinvocable", "disablemodelinvocation",
				"command-dispatch", "command_dispatch",
				"command-tool", "command_tool",
				"command-arg-mode", "command_arg_mode",
				"command-name", "command_name",
				"commanddispatch", "commandtool", "commandargmode", "commandname",
				"os", "platforms", "requires", "require", "install", "metadata":
				continue
			default:
				if text := readStringAny(value); text != "" {
					s.Metadata[key] = text
				}
			}
		}
		mergeCompatMetadata(&s, compat, mergeCompatOptions{
			TopLevelInvocation: topLevelInvocation,
			TopLevelDispatch:   topLevelDispatch,
			TopLevelExecution:  topLevelExecution,
		})
	}
	fillFromMarkdownSections(&s)
	if s.Name == "" && s.Description == "" {
		return Skill{}, false, nil
	}
	resources, resourceErr := scanResources()
	if resourceErr != nil {
		return Skill{}, false, fmt.Errorf("scan resources for %s: %w", location, resourceErr)
	}
	s.Resources = resources
	return s, true, nil
}

type compatExecutionSemantics struct {
	Context         string
	AllowedTools    []string
	Effort          string
	HasContext      bool
	HasAllowedTools bool
	HasEffort       bool
}

func parseSkillExecutionSemantics(frontmatter map[string]any) (string, []string, string, compatExecutionSemantics) {
	spec := compatExecutionSemantics{}
	if len(frontmatter) == 0 {
		return "", nil, "", spec
	}
	if hasAnyMapKey(frontmatter, "context", "skill_context", "skillContext") {
		spec.HasContext = true
		spec.Context = NormalizeSkillExecutionContext(readStringAny(firstNonNil(
			frontmatter["context"],
			frontmatter["skill_context"],
			frontmatter["skillContext"],
		)))
	}
	if hasAnyMapKey(frontmatter, "allowed_tools", "allowed-tools", "allowedTools") {
		spec.HasAllowedTools = true
		spec.AllowedTools = NormalizeSkillAllowedTools(readStringListAny(firstNonNil(
			frontmatter["allowed_tools"],
			frontmatter["allowed-tools"],
			frontmatter["allowedTools"],
		)))
	}
	if hasAnyMapKey(frontmatter, "effort", "reasoning_effort", "reasoningEffort") {
		spec.HasEffort = true
		spec.Effort = NormalizeSkillExecutionEffort(readStringAny(firstNonNil(
			frontmatter["effort"],
			frontmatter["reasoning_effort"],
			frontmatter["reasoningEffort"],
		)))
	}
	return spec.Context, spec.AllowedTools, spec.Effort, spec
}

func isLoadableSkillDir(name string) bool {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return false
	}
	if strings.HasPrefix(trimmed, ".") || strings.EqualFold(trimmed, "node_modules") {
		return false
	}
	return true
}

func normalizeLoadSourceDir(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if nested := resolveNestedSkillsRoot(trimmed); nested != "" {
		return nested
	}
	return trimmed
}

func resolveNestedSkillsRoot(dir string) string {
	trimmed := strings.TrimSpace(dir)
	if trimmed == "" {
		return ""
	}
	if sourceLooksLikeSkillRoot(trimmed) {
		return trimmed
	}
	nested := filepath.Join(trimmed, "skills")
	if sourceLooksLikeSkillRoot(nested) {
		return nested
	}
	return trimmed
}

func sourceLooksLikeSkillRoot(root string) bool {
	entries, err := os.ReadDir(root)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() || !isLoadableSkillDir(entry.Name()) {
			continue
		}
		info, err := os.Stat(filepath.Join(root, entry.Name(), skillDocName))
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}
		return true
	}
	return false
}

func scanSkillResources(skillDir string) (Resources, error) {
	out := Resources{
		Scripts:    []string{},
		References: []string{},
		Assets:     []string{},
	}
	if strings.TrimSpace(skillDir) == "" {
		return out, nil
	}
	absRoot, err := filepath.Abs(skillDir)
	if err != nil {
		return out, err
	}
	type resourceDir struct {
		name    string
		collect *[]string
	}
	dirs := []resourceDir{
		{name: "scripts", collect: &out.Scripts},
		{name: "references", collect: &out.References},
		{name: "assets", collect: &out.Assets},
	}
	for _, dir := range dirs {
		base := filepath.Join(absRoot, dir.name)
		info, statErr := os.Stat(base)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				continue
			}
			return out, statErr
		}
		if !info.IsDir() {
			continue
		}
		items := make([]string, 0)
		walkErr := filepath.WalkDir(base, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d == nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 {
				return nil
			}
			rel, relErr := filepath.Rel(absRoot, path)
			if relErr != nil {
				return relErr
			}
			normalized := filepath.ToSlash(filepath.Clean(rel))
			if normalized == "." || strings.HasPrefix(normalized, "../") {
				return nil
			}
			items = append(items, normalized)
			return nil
		})
		if walkErr != nil {
			return out, walkErr
		}
		sort.Strings(items)
		*dir.collect = items
	}
	return out, nil
}

func scanSkillResourcesFS(source fs.FS, skillDir string) (Resources, error) {
	out := Resources{
		Scripts:    []string{},
		References: []string{},
		Assets:     []string{},
	}
	if source == nil {
		return out, fs.ErrInvalid
	}
	skillDir = path.Clean(strings.TrimSpace(skillDir))
	if !fs.ValidPath(skillDir) {
		return out, fs.ErrInvalid
	}
	type resourceDir struct {
		name    string
		collect *[]string
	}
	dirs := []resourceDir{
		{name: "scripts", collect: &out.Scripts},
		{name: "references", collect: &out.References},
		{name: "assets", collect: &out.Assets},
	}
	for _, dir := range dirs {
		base := path.Join(skillDir, dir.name)
		info, statErr := fs.Stat(source, base)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				continue
			}
			return out, statErr
		}
		if !info.IsDir() {
			continue
		}
		items := make([]string, 0)
		walkErr := fs.WalkDir(source, base, func(name string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry == nil || entry.IsDir() || entry.Type()&fs.ModeSymlink != 0 {
				return nil
			}
			rel, ok := trimFSPathPrefix(name, skillDir)
			if !ok || rel == "." {
				return nil
			}
			items = append(items, rel)
			return nil
		})
		if walkErr != nil {
			return out, walkErr
		}
		sort.Strings(items)
		*dir.collect = items
	}
	return out, nil
}

func trimFSPathPrefix(name string, root string) (string, bool) {
	name = path.Clean(name)
	root = path.Clean(root)
	if name == root {
		return ".", true
	}
	prefix := root + "/"
	if root == "." {
		prefix = ""
	}
	if !strings.HasPrefix(name, prefix) {
		return "", false
	}
	rel := strings.TrimPrefix(name, prefix)
	if !fs.ValidPath(rel) {
		return "", false
	}
	return rel, true
}

func parseInvocationPolicy(frontmatter map[string]any) (InvocationPolicy, compatInvocation) {
	policy := InvocationPolicy{UserInvocable: true}
	flags := compatInvocation{}
	if value, ok, exists := readBoolByKeys(frontmatter, "user-invocable", "user_invocable", "userInvocable"); exists {
		flags.HasUserInvocable = true
		if ok {
			flags.UserInvocable = &value
			policy.UserInvocable = value
		}
	}
	if value, ok, exists := readBoolByKeys(frontmatter, "disable-model-invocation", "disable_model_invocation", "disableModelInvocation"); exists {
		flags.HasDisableModel = true
		if ok {
			flags.DisableModelInvocation = &value
			policy.DisableModelInvocation = value
		}
	}
	return policy, flags
}

func parseDispatchSpec(frontmatter map[string]any) (*DispatchSpec, bool) {
	if len(frontmatter) == 0 {
		return nil, false
	}
	hasDispatch := hasAnyMapKey(frontmatter,
		"command-dispatch", "command_dispatch", "commandDispatch",
		"command-tool", "command_tool", "commandTool",
		"command-arg-mode", "command_arg_mode", "commandArgMode",
		"command-name", "command_name", "commandName",
		"command-aliases", "command_aliases", "commandAliases",
		"command-alias", "command_alias", "commandAlias",
	)
	kind := strings.ToLower(strings.TrimSpace(firstNonEmptyString(
		readStringAny(frontmatter["command-dispatch"]),
		readStringAny(frontmatter["command_dispatch"]),
		readStringAny(frontmatter["commandDispatch"]),
	)))
	if kind == "" {
		return nil, hasDispatch
	}
	spec := &DispatchSpec{
		Kind: strings.ToLower(strings.TrimSpace(kind)),
		Tool: strings.TrimSpace(firstNonEmptyString(
			readStringAny(frontmatter["command-tool"]),
			readStringAny(frontmatter["command_tool"]),
			readStringAny(frontmatter["commandTool"]),
		)),
		ArgMode: strings.ToLower(strings.TrimSpace(firstNonEmptyString(
			readStringAny(frontmatter["command-arg-mode"]),
			readStringAny(frontmatter["command_arg_mode"]),
			readStringAny(frontmatter["commandArgMode"]),
		))),
		Command: strings.TrimSpace(firstNonEmptyString(
			readStringAny(frontmatter["command-name"]),
			readStringAny(frontmatter["command_name"]),
			readStringAny(frontmatter["commandName"]),
		)),
		Aliases: normalizeDispatchAliases("",
			readStringListAny(firstNonNil(
				frontmatter["command-aliases"],
				frontmatter["command_aliases"],
				frontmatter["commandAliases"],
				frontmatter["command-alias"],
				frontmatter["command_alias"],
				frontmatter["commandAlias"],
			))...,
		),
	}
	if spec.Command == "" {
		spec.Command = strings.TrimSpace(readStringAny(frontmatter["name"]))
	}
	spec.Command = strings.ToLower(strings.TrimSpace(spec.Command))
	spec.Aliases = normalizeDispatchAliases(spec.Command, spec.Aliases...)
	switch spec.ArgMode {
	case "", "raw":
	default:
		spec.ArgMode = "raw"
	}
	return spec, hasDispatch
}

func normalizeDispatchAliases(primary string, raw ...string) []string {
	primary = strings.ToLower(strings.TrimSpace(primary))
	if len(raw) == 0 {
		return nil
	}
	out := make([]string, 0, len(raw))
	seen := map[string]bool{}
	for _, item := range raw {
		alias := strings.ToLower(strings.TrimSpace(item))
		if alias == "" || alias == primary || seen[alias] {
			continue
		}
		seen[alias] = true
		out = append(out, alias)
	}
	if len(out) == 0 {
		return nil
	}
	sort.Strings(out)
	return out
}

func parseRequires(raw any) Requires {
	obj := toStringAnyMap(raw)
	if len(obj) == 0 {
		return Requires{}
	}
	return Requires{
		Bins:    readStringListAny(obj["bins"]),
		AnyBins: firstNonEmptyStringList(readStringListAny(obj["any_bins"]), readStringListAny(obj["anyBins"])),
		Env:     readStringListAny(obj["env"]),
		Config:  readStringListAny(obj["config"]),
	}
}

func parseInstallSpecs(raw any) ([]InstallSpec, error) {
	items, ok := raw.([]any)
	if !ok || len(items) == 0 {
		return nil, nil
	}
	out := make([]InstallSpec, 0, len(items))
	for _, item := range items {
		obj := toStringAnyMap(item)
		if len(obj) == 0 {
			continue
		}
		spec := InstallSpec{
			ID: readStringAny(obj["id"]),
			DependsOn: firstNonEmptyStringList(
				readStringListAny(obj["depends_on"]),
				readStringListAny(obj["dependsOn"]),
			),
			Kind:    strings.ToLower(readStringAny(firstNonNil(obj["kind"], obj["type"]))),
			Label:   readStringAny(obj["label"]),
			Bins:    readStringListAny(obj["bins"]),
			OS:      readStringListAny(obj["os"]),
			Tap:     readStringAny(obj["tap"]),
			Cask:    readStringAny(obj["cask"]),
			Formula: readStringAny(obj["formula"]),
			Package: readStringAny(obj["package"]),
			Module:  readStringAny(obj["module"]),
			URL:     readStringAny(obj["url"]),
			Command: firstNonEmptyStringList(
				readCommandAny(obj["command"]),
				readCommandAny(obj["cmd"]),
			),
			Rollback: firstNonEmptyStringList(
				readCommandAny(obj["rollback"]),
				readCommandAny(obj["rollback_command"]),
				readCommandAny(obj["rollbackCommand"]),
			),
			Archive: readStringAny(obj["archive"]),
			TargetDir: firstNonEmptyString(
				readStringAny(obj["targetDir"]),
				readStringAny(obj["target_dir"]),
			),
		}
		if extract, ok := readBoolAny(obj["extract"]); ok {
			spec.Extract = extract
		}
		if strip, ok := readIntAny(obj["stripComponents"]); ok {
			spec.StripComponents = strip
		} else if strip, ok := readIntAny(obj["strip_components"]); ok {
			spec.StripComponents = strip
		}
		if spec.Kind == "" {
			return nil, fmt.Errorf("install entry missing kind/type")
		}
		out = append(out, spec)
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

type compatMetadata struct {
	OS               []string
	Requires         Requires
	Install          []InstallSpec
	Keywords         []string
	Tags             []string
	WhenToUse        []string
	WhenNotToUse     []string
	NegativeExamples []string
	Steps            []string
	ToolHints        []string
	ToolHintsMatch   string
	Examples         []string
	EvalAssertions   []string
	Execution        compatExecutionSemantics
	Invocation       compatInvocation
	Dispatch         *DispatchSpec
	Extras           map[string]string
}

type compatInvocation struct {
	UserInvocable          *bool
	DisableModelInvocation *bool
	HasUserInvocable       bool
	HasDisableModel        bool
}

func parseCompatMetadata(raw any) (compatMetadata, error) {
	env, err := parseMetadataEnvelope(raw)
	if err != nil || len(env) == 0 {
		return compatMetadata{}, err
	}
	compatRaw := toStringAnyMap(firstNonNil(env["openclaw"], env["moltbot"], env["clawdbot"]))
	if len(compatRaw) == 0 {
		return compatMetadata{}, nil
	}
	install, installErr := parseInstallSpecs(compatRaw["install"])
	if installErr != nil {
		return compatMetadata{}, installErr
	}
	dispatch, _ := parseDispatchSpec(compatRaw)
	invocation := parseCompatInvocation(compatRaw)
	context, allowedTools, effort, execution := parseSkillExecutionSemantics(compatRaw)
	execution.Context = context
	execution.AllowedTools = allowedTools
	execution.Effort = effort
	extras := map[string]string{}
	setExtra := func(key string, value any) {
		if text := readStringAny(value); text != "" {
			extras[key] = text
		}
	}
	setExtra("skill_key", firstNonNil(compatRaw["skillKey"], compatRaw["skill_key"]))
	setExtra("primary_env", firstNonNil(compatRaw["primaryEnv"], compatRaw["primary_env"]))
	setExtra("homepage", compatRaw["homepage"])
	setExtra("emoji", compatRaw["emoji"])
	if flag, ok := readBoolAny(compatRaw["always"]); ok {
		extras["always"] = strconv.FormatBool(flag)
	}
	return compatMetadata{
		OS:               readStringListWithAliases(compatRaw, "os", "platforms"),
		Requires:         parseRequires(firstNonNil(compatRaw["requires"], compatRaw["require"])),
		Install:          install,
		Keywords:         readStringListWithAliases(compatRaw, "keywords", "keyword"),
		Tags:             readStringListWithAliases(compatRaw, "tags", "tag"),
		WhenToUse:        readStringListWithAliases(compatRaw, "when_to_use", "whenToUse"),
		WhenNotToUse:     readStringListWithAliases(compatRaw, "when_not_to_use", "whenNotToUse"),
		NegativeExamples: readStringListWithAliases(compatRaw, "negative_examples", "negativeExamples"),
		Steps:            readStringListWithAliases(compatRaw, "steps", "workflow"),
		ToolHints:        readStringListWithAliases(compatRaw, "tool_hints", "toolHints"),
		ToolHintsMatch: NormalizeToolHintsMatch(readStringAny(firstNonNil(
			compatRaw["tool_hints_match"],
			compatRaw["tool-hints-match"],
			compatRaw["toolHintsMatch"],
		))),
		Examples:       readStringListWithAliases(compatRaw, "examples", "example_prompts", "examplePrompts"),
		EvalAssertions: readStringListWithAliases(compatRaw, "eval_assertions", "evalAssertions"),
		Execution:      execution,
		Invocation:     invocation,
		Dispatch:       dispatch,
		Extras:         extras,
	}, nil
}

func parseMetadataEnvelope(raw any) (map[string]any, error) {
	if raw == nil {
		return nil, nil
	}
	if direct := toStringAnyMap(raw); len(direct) > 0 {
		return direct, nil
	}
	text := readStringAny(raw)
	if text == "" {
		return nil, nil
	}
	var out map[string]any
	if err := yaml.Unmarshal([]byte(text), &out); err == nil && len(out) > 0 {
		return out, nil
	}
	if err := json.Unmarshal([]byte(text), &out); err == nil && len(out) > 0 {
		return out, nil
	}
	return nil, fmt.Errorf("invalid metadata format")
}

type mergeCompatOptions struct {
	TopLevelInvocation compatInvocation
	TopLevelDispatch   bool
	TopLevelExecution  compatExecutionSemantics
}

func mergeCompatMetadata(skill *Skill, compat compatMetadata, opts mergeCompatOptions) {
	if skill == nil {
		return
	}
	if len(skill.OS) == 0 && len(compat.OS) > 0 {
		skill.OS = append([]string(nil), compat.OS...)
	}
	skill.Requires = mergeRequiresFallback(skill.Requires, compat.Requires)
	if len(skill.Install) == 0 && len(compat.Install) > 0 {
		skill.Install = append([]InstallSpec(nil), compat.Install...)
	}
	if len(skill.Keywords) == 0 && len(compat.Keywords) > 0 {
		skill.Keywords = append([]string(nil), compat.Keywords...)
	}
	if len(skill.Tags) == 0 && len(compat.Tags) > 0 {
		skill.Tags = append([]string(nil), compat.Tags...)
	}
	if len(skill.WhenToUse) == 0 && len(compat.WhenToUse) > 0 {
		skill.WhenToUse = append([]string(nil), compat.WhenToUse...)
	}
	if len(skill.WhenNotToUse) == 0 && len(compat.WhenNotToUse) > 0 {
		skill.WhenNotToUse = append([]string(nil), compat.WhenNotToUse...)
	}
	if len(skill.NegativeExamples) == 0 && len(compat.NegativeExamples) > 0 {
		skill.NegativeExamples = append([]string(nil), compat.NegativeExamples...)
	}
	if len(skill.Steps) == 0 && len(compat.Steps) > 0 {
		skill.Steps = append([]string(nil), compat.Steps...)
	}
	if len(skill.ToolHints) == 0 && len(compat.ToolHints) > 0 {
		skill.ToolHints = append([]string(nil), compat.ToolHints...)
	}
	if skill.ToolHintsMatch == "" && compat.ToolHintsMatch != "" {
		skill.ToolHintsMatch = compat.ToolHintsMatch
	}
	if len(skill.Examples) == 0 && len(compat.Examples) > 0 {
		skill.Examples = append([]string(nil), compat.Examples...)
	}
	if len(skill.EvalAssertions) == 0 && len(compat.EvalAssertions) > 0 {
		skill.EvalAssertions = append([]string(nil), compat.EvalAssertions...)
	}
	if !opts.TopLevelExecution.HasContext && skill.ExecutionContext == "" && compat.Execution.Context != "" {
		skill.ExecutionContext = compat.Execution.Context
	}
	if !opts.TopLevelExecution.HasAllowedTools && len(skill.AllowedTools) == 0 && len(compat.Execution.AllowedTools) > 0 {
		skill.AllowedTools = append([]string(nil), compat.Execution.AllowedTools...)
	}
	if !opts.TopLevelExecution.HasEffort && skill.Effort == "" && compat.Execution.Effort != "" {
		skill.Effort = compat.Execution.Effort
	}
	if !opts.TopLevelInvocation.HasUserInvocable && compat.Invocation.UserInvocable != nil {
		skill.Invocation.UserInvocable = *compat.Invocation.UserInvocable
	}
	if !opts.TopLevelInvocation.HasDisableModel && compat.Invocation.DisableModelInvocation != nil {
		skill.Invocation.DisableModelInvocation = *compat.Invocation.DisableModelInvocation
	}
	if !opts.TopLevelDispatch && skill.Dispatch == nil && compat.Dispatch != nil {
		dispatch := *compat.Dispatch
		skill.Dispatch = &dispatch
	}
	for key, value := range compat.Extras {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		if _, exists := skill.Metadata[key]; exists {
			continue
		}
		skill.Metadata[key] = value
	}
}

func mergeRequiresFallback(primary Requires, fallback Requires) Requires {
	out := primary
	if len(out.Bins) == 0 && len(fallback.Bins) > 0 {
		out.Bins = append([]string(nil), fallback.Bins...)
	}
	if len(out.AnyBins) == 0 && len(fallback.AnyBins) > 0 {
		out.AnyBins = append([]string(nil), fallback.AnyBins...)
	}
	if len(out.Env) == 0 && len(fallback.Env) > 0 {
		out.Env = append([]string(nil), fallback.Env...)
	}
	if len(out.Config) == 0 && len(fallback.Config) > 0 {
		out.Config = append([]string(nil), fallback.Config...)
	}
	return out
}

func parseCompatInvocation(raw map[string]any) compatInvocation {
	out := compatInvocation{}
	if value, ok, exists := readBoolByKeys(raw, "user-invocable", "user_invocable", "userInvocable"); exists {
		out.HasUserInvocable = true
		if ok {
			out.UserInvocable = &value
		}
	}
	if value, ok, exists := readBoolByKeys(raw, "disable-model-invocation", "disable_model_invocation", "disableModelInvocation"); exists {
		out.HasDisableModel = true
		if ok {
			out.DisableModelInvocation = &value
		}
	}
	return out
}

func readStringListWithAliases(raw map[string]any, keys ...string) []string {
	if len(raw) == 0 || len(keys) == 0 {
		return nil
	}
	for _, key := range keys {
		values := readStringListAny(raw[key])
		if len(values) > 0 {
			return values
		}
	}
	return nil
}

func hasAnyMapKey(raw map[string]any, keys ...string) bool {
	if len(raw) == 0 || len(keys) == 0 {
		return false
	}
	for _, key := range keys {
		if _, exists := raw[key]; exists {
			return true
		}
	}
	return false
}

func readBoolByKeys(raw map[string]any, keys ...string) (value bool, ok bool, exists bool) {
	if len(raw) == 0 {
		return false, false, false
	}
	for _, key := range keys {
		rawValue, present := raw[key]
		if !present {
			continue
		}
		exists = true
		if parsed, parsedOK := readBoolAny(rawValue); parsedOK {
			return parsed, true, true
		}
	}
	return false, false, exists
}

func parseFrontMatter(raw string, strict bool) (map[string]any, string, error) {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	normalized = strings.TrimPrefix(normalized, "\uFEFF")
	if !strings.HasPrefix(normalized, "---\n") {
		return nil, normalized, nil
	}
	end := strings.Index(normalized[4:], "\n---\n")
	if end < 0 {
		return nil, normalized, nil
	}
	fmRaw := normalized[4 : 4+end]
	body := normalized[4+end+5:]

	parsed, err := parseFrontMatterYAML(fmRaw)
	if err == nil {
		return parsed, body, nil
	}
	if strict {
		return nil, "", err
	}

	repaired := repairFrontMatterScalars(fmRaw)
	if repaired != fmRaw {
		repairedParsed, repairedErr := parseFrontMatterYAML(repaired)
		if repairedErr == nil {
			return repairedParsed, body, nil
		}
	}

	if fallback, ok := parseFrontMatterLoose(fmRaw); ok {
		return fallback, body, nil
	}
	return nil, "", err
}

func parseFrontMatterYAML(raw string) (map[string]any, error) {
	m := map[string]any{}
	if err := yaml.Unmarshal([]byte(raw), &m); err != nil {
		return nil, err
	}
	return m, nil
}

func repairFrontMatterScalars(raw string) string {
	lines := strings.Split(raw, "\n")
	changed := false
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		if key == "" {
			continue
		}
		value := strings.TrimSpace(line[idx+1:])
		if value == "" {
			continue
		}
		if isQuotedScalar(value) {
			continue
		}
		if strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[") || strings.HasPrefix(value, "|") || strings.HasPrefix(value, ">") {
			continue
		}
		if !strings.Contains(value, ": ") {
			continue
		}
		lines[i] = key + `: "` + strings.ReplaceAll(value, `"`, `\\"`) + `"`
		changed = true
	}
	if !changed {
		return raw
	}
	return strings.Join(lines, "\n")
}

func isQuotedScalar(value string) bool {
	if len(value) < 2 {
		return false
	}
	if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		return true
	}
	if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
		return true
	}
	return false
}

func parseFrontMatterLoose(raw string) (map[string]any, bool) {
	lines := strings.Split(raw, "\n")
	out := map[string]any{}
	var currentKey string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if currentKey == "" {
				continue
			}
			fragment := strings.TrimSpace(line)
			if fragment == "" {
				continue
			}
			prev := readStringAny(out[currentKey])
			if prev == "" {
				out[currentKey] = fragment
			} else {
				out[currentKey] = prev + "\n" + fragment
			}
			continue
		}
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		if key == "" {
			continue
		}
		value := strings.TrimSpace(line[idx+1:])
		currentKey = key
		if value == "" {
			out[key] = ""
			continue
		}
		if parsed, ok := parseLooseScalar(value); ok {
			out[key] = parsed
		} else {
			out[key] = value
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func parseLooseScalar(value string) (any, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", true
	}
	if isQuotedScalar(trimmed) {
		if len(trimmed) >= 2 {
			return strings.TrimSpace(trimmed[1 : len(trimmed)-1]), true
		}
		return "", true
	}
	lowered := strings.ToLower(trimmed)
	switch lowered {
	case "true", "yes", "on", "1":
		return true, true
	case "false", "no", "off", "0":
		return false, true
	}
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		inner := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
		if inner == "" {
			return []string{}, true
		}
		parts := strings.Split(inner, ",")
		items := make([]string, 0, len(parts))
		for _, part := range parts {
			item := strings.TrimSpace(part)
			if item == "" {
				continue
			}
			item = strings.Trim(item, `"`)
			item = strings.Trim(item, "'")
			if item != "" {
				items = append(items, item)
			}
		}
		if len(items) > 0 {
			return items, true
		}
	}
	return trimmed, true
}

func fillFromMarkdownSections(s *Skill) {
	if s == nil || s.Content == "" {
		return
	}
	if len(s.Keywords) == 0 {
		s.Keywords = parseSectionItems(s.Content, "keywords")
	}
	if len(s.Tags) == 0 {
		s.Tags = parseSectionItems(s.Content, "tags")
	}
	if len(s.WhenToUse) == 0 {
		s.WhenToUse = parseSectionItems(s.Content, "when to use")
	}
	if len(s.WhenNotToUse) == 0 {
		s.WhenNotToUse = parseSectionItems(s.Content, "when not to use")
	}
	if len(s.Steps) == 0 {
		s.Steps = parseSectionItems(s.Content, "steps")
	}
	if len(s.NegativeExamples) == 0 {
		s.NegativeExamples = parseSectionItems(s.Content, "negative examples")
	}
	if len(s.ToolHints) == 0 {
		s.ToolHints = parseSectionItems(s.Content, "tool hints")
	}
	if len(s.Examples) == 0 {
		s.Examples = parseSectionItems(s.Content, "examples")
	}
	if len(s.EvalAssertions) == 0 {
		s.EvalAssertions = parseSectionItems(s.Content, "eval assertions")
	}
}

func parseSectionItems(markdown string, heading string) []string {
	lines := strings.Split(markdown, "\n")
	target := normalizeHeading(heading)
	inSection := false
	items := make([]string, 0)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			title := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			if normalizeHeading(title) == target {
				inSection = true
				continue
			}
			if inSection {
				break
			}
			continue
		}
		if !inSection {
			continue
		}
		if item, ok := parseListItem(trimmed); ok {
			items = append(items, item)
		}
	}
	return uniqueNonEmpty(items)
}

func parseListItem(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return "", false
	}
	for _, prefix := range []string{"- ", "* ", "+ "} {
		if strings.HasPrefix(trimmed, prefix) {
			out := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
			return out, out != ""
		}
	}
	parts := strings.SplitN(trimmed, ".", 2)
	if len(parts) == 2 {
		if _, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
			out := strings.TrimSpace(parts[1])
			return out, out != ""
		}
	}
	return "", false
}

func normalizeHeading(in string) string {
	in = strings.ToLower(strings.TrimSpace(in))
	in = strings.ReplaceAll(in, "_", " ")
	in = strings.ReplaceAll(in, "-", " ")
	in = strings.ReplaceAll(in, ":", "")
	in = strings.Join(strings.Fields(in), " ")
	return in
}

func readStringAny(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case fmtStringer:
		return strings.TrimSpace(v.String())
	default:
		return ""
	}
}

type fmtStringer interface {
	String() string
}

func readBoolAny(value any) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		text := strings.ToLower(strings.TrimSpace(v))
		switch text {
		case "true", "yes", "1", "on":
			return true, true
		case "false", "no", "0", "off":
			return false, true
		}
	}
	return false, false
}

func readIntAny(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func readStringListAny(value any) []string {
	switch v := value.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if text := readStringAny(item); text != "" {
				out = append(out, text)
			}
		}
		return uniqueNonEmpty(out)
	case []string:
		return uniqueNonEmpty(v)
	case string:
		if text := strings.TrimSpace(v); text != "" {
			if items, ok := readLooseListString(text); ok {
				return uniqueNonEmpty(items)
			}
			return []string{text}
		}
	}
	return nil
}

func readLooseListString(text string) ([]string, bool) {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "- ") {
			return nil, false
		}
		item := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		if item != "" {
			out = append(out, item)
		}
	}
	return out, len(out) > 0
}

func readCommandAny(value any) []string {
	list := readStringListAny(value)
	if len(list) > 1 {
		return list
	}
	if len(list) == 1 {
		text := strings.TrimSpace(list[0])
		if text == "" {
			return nil
		}
		if strings.ContainsAny(text, " \t") {
			return strings.Fields(text)
		}
		return []string{text}
	}
	return nil
}

func uniqueNonEmpty(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, item := range values {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func firstNonEmptyStringList(values ...[]string) []string {
	for _, item := range values {
		if len(item) > 0 {
			return item
		}
	}
	return nil
}

func firstNonEmptyString(values ...string) string {
	for _, item := range values {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstNonNil(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func toStringAnyMap(raw any) map[string]any {
	switch value := raw.(type) {
	case map[string]any:
		return value
	case map[any]any:
		if len(value) == 0 {
			return nil
		}
		out := make(map[string]any, len(value))
		for key, item := range value {
			text := readStringAny(key)
			if text == "" {
				continue
			}
			out[text] = item
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		return nil
	}
}
