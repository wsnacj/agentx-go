//go:build ignore

// prepare_docs_portal projects the canonical Markdown sources into an ignored
// VitePress source tree. It does not author or mutate API documentation.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type packageEntry struct {
	Package  string `json:"package"`
	Maturity string `json:"maturity"`
	Badge    string `json:"badge"`
	Source   string `json:"source"`
	Route    string `json:"route"`
	Hash     string `json:"hash"`
}

type navigation struct {
	Version      string         `json:"version"`
	SourceCommit string         `json:"source_commit"`
	Packages     []packageEntry `json:"packages"`
}

var inlineLinkPattern = regexp.MustCompile(`!?\[[^]]*\]\(([^)]+)\)`)

func main() {
	root, err := os.Getwd()
	check(err)
	checkFile(filepath.Join(root, "go.mod"))

	output := filepath.Join(root, "portal", ".generated")
	check(os.RemoveAll(output))
	check(os.MkdirAll(output, 0o755))
	commit := gitOutput(root, "rev-parse", "HEAD")

	check(copyFile(
		filepath.Join(root, "portal", "content", "index.md"),
		filepath.Join(output, "index.md"),
	))
	check(copyMarkdownTree(root, filepath.Join(root, "docs"), filepath.Join(output, "docs"), commit))
	for _, directory := range []string{"components", "runtime", "extensions", "conformance", "examples"} {
		check(copyMarkdownTree(root, filepath.Join(root, directory), filepath.Join(output, directory), commit))
	}

	for _, name := range []string{"SECURITY.md", "SUPPORT.md", "CONTRIBUTING.md", "CHANGELOG.md"} {
		check(copyFile(filepath.Join(root, name), filepath.Join(output, name)))
	}

	entries, err := readPackages(filepath.Join(root, "docs", "reference", "developer-preview-packages.tsv"))
	check(err)
	if len(entries) != 46 {
		check(fmt.Errorf("package maturity source has %d entries, want 46", len(entries)))
	}
	candidates := 0
	for _, entry := range entries {
		if entry.Maturity == "developer_preview_candidate" {
			candidates++
		}
	}
	if candidates != 9 {
		check(fmt.Errorf("package maturity source has %d Developer Preview candidates, want 9", candidates))
	}

	for index := range entries {
		entry := &entries[index]
		entry.Badge = maturityBadge(entry.Maturity)
		entry.Route = markdownRoute(entry.Source)
		source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(entry.Source)))
		check(err)
		projected := projectMarkdownLinks(root, filepath.Join(root, filepath.FromSlash(entry.Source)), commit, string(source))
		decorated := decorateAPI(*entry, commit, projected)
		target := filepath.Join(output, filepath.FromSlash(entry.Source))
		check(writeFile(target, []byte(decorated)))
	}

	versionBytes, err := os.ReadFile(filepath.Join(root, "docs", "reference", "developer-preview-version.txt"))
	check(err)
	nav := navigation{
		Version:      strings.TrimSpace(string(versionBytes)),
		SourceCommit: commit,
		Packages:     entries,
	}
	navigationJSON, err := json.MarshalIndent(nav, "", "  ")
	check(err)
	navigationJSON = append(navigationJSON, '\n')
	check(writeFile(filepath.Join(output, "navigation.json"), navigationJSON))
	check(writeFile(filepath.Join(output, "packages.md"), []byte(packagesPage(nav))))

	fmt.Printf("agentx-docs-portal-prepare-ok:packages=%d:candidates=%d:version=%s:source=%s\n", len(entries), candidates, nav.Version, short(commit))
}

func readPackages(path string) ([]packageEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []packageEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			return nil, fmt.Errorf("invalid package maturity row %q", line)
		}
		name := fields[0]
		if name == "." {
			name = "agentx"
		}
		entries = append(entries, packageEntry{
			Package:  name,
			Maturity: fields[1],
			Source:   fields[2],
			Hash:     fields[3],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Package < entries[j].Package })
	return entries, nil
}

func maturityBadge(value string) string {
	switch value {
	case "developer_preview_candidate":
		return "Developer Preview"
	case "experimental_extension":
		return "Experimental"
	case "internalization_candidate":
		return "Internalization candidate"
	default:
		check(fmt.Errorf("unknown package maturity %q", value))
		return ""
	}
}

func markdownRoute(source string) string {
	value := "/" + strings.TrimSuffix(filepath.ToSlash(source), ".md")
	if strings.HasSuffix(value, "/README") {
		value = strings.TrimSuffix(value, "README")
	}
	return value
}

func decorateAPI(entry packageEntry, commit, body string) string {
	sourceURL := fmt.Sprintf("https://github.com/wsnacj/agentx-go/blob/%s/%s", commit, entry.Source)
	return fmt.Sprintf(`---
title: %s
outline: [2, 3]
---

<div class="maturity-line">
  <span class="maturity-badge">%s</span>
  <span>package <code>%s</code></span>
  <a href="%s">查看正文源</a>
</div>

%s`, entry.Package, entry.Badge, entry.Package, sourceURL, body)
}

func packagesPage(nav navigation) string {
	var builder strings.Builder
	candidates := 0
	for _, entry := range nav.Packages {
		if entry.Maturity == "developer_preview_candidate" {
			candidates++
		}
	}
	builder.WriteString("---\ntitle: Package API\noutline: [2, 3]\n---\n\n")
	builder.WriteString("# Package API\n\n")
	builder.WriteString("当前索引由 `developer-preview-packages.tsv` 确定性生成，不是第二套API事实源。\n\n")
	fmt.Fprintf(&builder, "- 固定版本：`%s`\n", nav.Version)
	fmt.Fprintf(&builder, "- 构建源码：`%s`\n", short(nav.SourceCommit))
	fmt.Fprintf(&builder, "- Package：%d；Developer Preview candidate：%d\n\n", len(nav.Packages), candidates)

	groups := []struct {
		maturity string
		title    string
		detail   string
	}{
		{"developer_preview_candidate", "Developer Preview candidate", "进入focused API签名、类型闭包与中文Reference gate；尚无semver长期承诺。"},
		{"experimental_extension", "Experimental extension", "已有真实implementation与consumer，Beta前仍可能调整入口或owner。"},
		{"internalization_candidate", "Internalization candidate", "低层迁移owner或Go internal实现，新项目不应直接建立长期依赖。"},
	}
	for _, group := range groups {
		fmt.Fprintf(&builder, "## %s\n\n%s\n\n<div class=\"package-grid\">\n", group.title, group.detail)
		for _, entry := range nav.Packages {
			if entry.Maturity != group.maturity {
				continue
			}
			fmt.Fprintf(&builder, "  <a class=\"package-card\" href=\"%s\"><h3><code>%s</code></h3><p>%s</p></a>\n", entry.Route, entry.Package, entry.Badge)
		}
		builder.WriteString("</div>\n\n")
	}
	return builder.String()
}

func copyMarkdownTree(root, source, target, commit string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			return nil
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(target, relative)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		projected := []byte(projectMarkdownLinks(root, path, commit, string(content)))
		if err := writeFile(targetPath, projected); err != nil {
			return err
		}
		// Repository links commonly point to a consumer/example directory. Expose
		// the canonical README at the matching VitePress directory route.
		if strings.EqualFold(entry.Name(), "README.md") {
			return writeFile(filepath.Join(filepath.Dir(targetPath), "index.md"), projected)
		}
		return nil
	})
}

func projectMarkdownLinks(root, source, commit, content string) string {
	return inlineLinkPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := inlineLinkPattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		raw := strings.TrimSpace(parts[1])
		if raw == "" || strings.HasPrefix(raw, "#") || strings.ContainsAny(raw, " \t\"") {
			return match
		}
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Scheme != "" || parsed.Host != "" || parsed.Path == "" {
			return match
		}
		resolved := filepath.Clean(filepath.Join(filepath.Dir(source), filepath.FromSlash(parsed.Path)))
		relative, err := filepath.Rel(root, resolved)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return match
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return match
		}
		targetStart := strings.LastIndex(match, "](") + 2
		targetEnd := len(match) - 1
		if targetStart < 2 || targetEnd <= targetStart {
			return match
		}
		kind := "blob"
		if info.IsDir() {
			if _, err := os.Stat(filepath.Join(resolved, "README.md")); err == nil {
				target := strings.TrimSuffix(parsed.Path, "/") + "/README.md"
				if parsed.Fragment != "" {
					target += "#" + parsed.Fragment
				}
				return match[:targetStart] + target + match[targetEnd:]
			}
			kind = "tree"
		} else if strings.EqualFold(filepath.Ext(resolved), ".md") {
			return match
		}
		target := fmt.Sprintf("https://github.com/wsnacj/agentx-go/%s/%s/%s", kind, commit, filepath.ToSlash(relative))
		if parsed.Fragment != "" {
			target += "#" + parsed.Fragment
		}
		return match[:targetStart] + target + match[targetEnd:]
	})
}

func copyFile(source, target string) error {
	content, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return writeFile(target, content)
}

func writeFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func checkFile(path string) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() == 0 {
		check(fmt.Errorf("missing non-empty file %s", path))
	}
}

func gitOutput(root string, args ...string) string {
	command := exec.Command("git", args...)
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		check(fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output))))
	}
	return strings.TrimSpace(string(output))
}

func short(value string) string {
	if len(value) <= 12 {
		return value
	}
	return value[:12]
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "docs portal prepare:", err)
	os.Exit(1)
}
