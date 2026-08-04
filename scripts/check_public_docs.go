//go:build ignore

// check_public_docs verifies that first-user documentation does not expose
// internal migration narration or unstable validation versions.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var forbiddenPublicPatterns = []struct {
	name    string
	pattern *regexp.Regexp
}{
	{name: "internal milestone", pattern: regexp.MustCompile(`\b[MPW][0-9][A-Z0-9.-]*\b`)},
	{name: "HS abbreviation", pattern: regexp.MustCompile(`\bHS\b`)},
	{name: "checkpoint narration", pattern: regexp.MustCompile(`(?i)checkpoint`)},
	{name: "pseudo-version narration", pattern: regexp.MustCompile(`(?i)pseudo-version`)},
	{name: "source-authority narration", pattern: regexp.MustCompile(`(?i)source-authority`)},
	{name: "owner approval narration", pattern: regexp.MustCompile(`\bOwner\b`)},
	{name: "private validation narration", pattern: regexp.MustCompile(`(?i)private validation`)},
	{name: "raw pseudo version", pattern: regexp.MustCompile(`v0\.0\.0-[0-9]{14}-[0-9a-f]{12}`)},
	{name: "machine-local path", pattern: regexp.MustCompile(`/Users/[^/]+/`)},
}

func main() {
	root, err := os.Getwd()
	check(err)
	files := []string{
		"README.md",
		"CHANGELOG.md",
		"CONTRIBUTING.md",
		"SECURITY.md",
		"SUPPORT.md",
		"portal/README.md",
		"portal/content/index.md",
	}
	check(filepath.WalkDir(filepath.Join(root, "docs"), func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	}))

	for _, relative := range files {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		check(err)
		for _, forbidden := range forbiddenPublicPatterns {
			if location := forbidden.pattern.FindIndex(content); location != nil {
				check(fmt.Errorf("%s contains %s near %q", relative, forbidden.name, excerpt(content, location[0])))
			}
		}
	}

	readmeLines, err := lineCount(filepath.Join(root, "README.md"))
	check(err)
	if readmeLines > 230 {
		check(fmt.Errorf("README.md has %d lines, want at most 230", readmeLines))
	}

	for _, removed := range []string{
		"docs/guides/hs-migration.md",
		"docs/reference/pre-beta-admission.md",
		"docs/reference/pre-beta-owner-decisions.md",
	} {
		if _, err := os.Stat(filepath.Join(root, removed)); err == nil || !os.IsNotExist(err) {
			check(fmt.Errorf("internal-only public document still exists: %s", removed))
		}
	}

	checkInternalMaturity(filepath.Join(root, "docs/reference/developer-preview-packages.tsv"))
	fmt.Printf("agentx-public-docs-ok:files=%d:readme_lines=%d\n", len(files), readmeLines)
}

func checkInternalMaturity(path string) {
	file, err := os.Open(path)
	check(err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		fields := strings.Split(text, "\t")
		if len(fields) != 4 {
			check(fmt.Errorf("%s:%d malformed maturity row", path, line))
		}
		if fields[1] == "internalization_candidate" && !strings.Contains("/"+fields[0]+"/", "/internal/") {
			check(fmt.Errorf("%s:%d publicly importable internalization candidate %s", path, line, fields[0]))
		}
	}
	check(scanner.Err())
}

func lineCount(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	lines := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines++
	}
	return lines, scanner.Err()
}

func excerpt(content []byte, offset int) string {
	start := offset - 24
	if start < 0 {
		start = 0
	}
	end := offset + 64
	if end > len(content) {
		end = len(content)
	}
	return strings.ReplaceAll(string(content[start:end]), "\n", " ")
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "public docs gate:", err)
	os.Exit(1)
}
