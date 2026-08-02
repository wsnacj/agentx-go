//go:build ignore

// check_docs_portal verifies the generated source tree and VitePress build.
package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type packageEntry struct {
	Package  string `json:"package"`
	Maturity string `json:"maturity"`
	Source   string `json:"source"`
	Route    string `json:"route"`
}

type navigation struct {
	Version      string         `json:"version"`
	SourceCommit string         `json:"source_commit"`
	Packages     []packageEntry `json:"packages"`
}

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`sk-[A-Za-z0-9_-]{20,}`),
	regexp.MustCompile(`ghp_[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
}

func main() {
	root, err := os.Getwd()
	check(err)
	generated := filepath.Join(root, "portal", ".generated")
	dist := filepath.Join(root, "portal", ".vitepress", "dist")

	content, err := os.ReadFile(filepath.Join(generated, "navigation.json"))
	check(err)
	var nav navigation
	check(json.Unmarshal(content, &nav))
	if nav.Version == "" || nav.SourceCommit == "" {
		check(fmt.Errorf("portal navigation is missing version or source commit"))
	}
	if len(nav.Packages) != 46 {
		check(fmt.Errorf("portal has %d packages, want 46", len(nav.Packages)))
	}
	candidates := 0
	for _, entry := range nav.Packages {
		if entry.Maturity == "developer_preview_candidate" {
			candidates++
		}
		checkFile(routeFile(generated, entry.Route))
	}
	if candidates != 9 {
		check(fmt.Errorf("portal has %d Developer Preview candidates, want 9", candidates))
	}

	for _, relative := range []string{
		"index.html",
		"packages.html",
		"docs/quickstart.html",
		"docs/concepts/execution-model.html",
		"docs/guides/custom-adapter.html",
		"docs/guides/model-tool-hostkit.html",
		"docs/guides/workflow-hostkit.html",
		"docs/reference/package-maturity.html",
		"docs/reference/distribution-readiness.html",
		"SECURITY.html",
		"SUPPORT.html",
	} {
		checkFile(filepath.Join(dist, filepath.FromSlash(relative)))
	}

	htmlCount := 0
	searchAsset := false
	check(filepath.WalkDir(dist, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		name := strings.ToLower(entry.Name())
		if strings.HasSuffix(name, ".html") {
			htmlCount++
		}
		if strings.Contains(name, "minisearch") || strings.Contains(name, "local-search") || strings.Contains(name, "hashmap") {
			searchAsset = true
		}
		return nil
	}))
	if htmlCount < 60 {
		check(fmt.Errorf("portal built only %d HTML pages, want at least 60", htmlCount))
	}
	if !searchAsset {
		check(fmt.Errorf("portal build does not contain a local-search asset"))
	}

	checkNoSecrets(filepath.Join(root, "portal"))
	tracked := gitOutput(root, "ls-files", "--", "portal/.generated", "portal/.vitepress/dist", "node_modules")
	if tracked != "" {
		check(fmt.Errorf("generated portal output is tracked: %s", tracked))
	}
	ignored := gitOutput(root, "check-ignore", "portal/.generated/navigation.json", "portal/.vitepress/dist/index.html", "node_modules/.package-lock.json")
	if len(strings.Split(strings.TrimSpace(ignored), "\n")) != 3 {
		check(fmt.Errorf("portal generated/cache paths are not all ignored: %q", ignored))
	}

	fmt.Printf("agentx-docs-portal-ok:html=%d:packages=%d:candidates=%d:search=local:version=%s:source=%s\n", htmlCount, len(nav.Packages), candidates, nav.Version, short(nav.SourceCommit))
}

func routeFile(generated, route string) string {
	value := strings.TrimPrefix(route, "/")
	if value == "" {
		return filepath.Join(generated, "index.md")
	}
	return filepath.Join(generated, filepath.FromSlash(value)+".md")
}

func checkNoSecrets(root string) {
	check(filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == "dist" || entry.Name() == "cache" {
				return filepath.SkipDir
			}
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, pattern := range secretPatterns {
			if pattern.Match(content) {
				return fmt.Errorf("credential-like value in %s", path)
			}
		}
		return nil
	}))
}

func checkFile(path string) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() == 0 {
		check(fmt.Errorf("missing non-empty portal artifact %s", path))
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
	fmt.Fprintln(os.Stderr, "docs portal gate:", err)
	os.Exit(1)
}
