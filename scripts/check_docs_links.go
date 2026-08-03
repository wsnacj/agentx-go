//go:build ignore

// check_docs_links verifies repository-local links in the AgentX Core Markdown sources.
package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	inlineLinkPattern    = regexp.MustCompile(`!?\[[^]]*\]\(([^)]+)\)`)
	referenceLinkPattern = regexp.MustCompile(`^\s*\[[^]]+\]:\s*(\S+)`)
)

func main() {
	root, err := os.Getwd()
	check(err)
	files, err := markdownFiles(root)
	check(err)

	checked := 0
	var failures []string
	for _, path := range files {
		count, current, err := checkFile(root, path)
		check(err)
		checked += count
		failures = append(failures, current...)
	}
	if len(failures) > 0 {
		sort.Strings(failures)
		for _, failure := range failures {
			fmt.Fprintln(os.Stderr, failure)
		}
		check(fmt.Errorf("%d local documentation links are invalid", len(failures)))
	}
	fmt.Printf("agentx-doc-links-ok:files=%d:local-links=%d\n", len(files), checked)
}

func markdownFiles(root string) ([]string, error) {
	var files []string
	for _, relative := range []string{
		"README.md",
		"docs",
		"examples",
		"components",
		"runtime",
		"extensions",
		"providers",
		"tools",
		"browser",
		"document",
		"scenes",
	} {
		path := filepath.Join(root, relative)
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			files = append(files, path)
			continue
		}
		err = filepath.WalkDir(path, func(current string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
				files = append(files, current)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(files)
	return files, nil
}

func checkFile(root, path string) (int, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, nil, err
	}
	defer file.Close()

	inFence := false
	checked := 0
	var failures []string
	scanner := bufio.NewScanner(file)
	for line := 1; scanner.Scan(); line++ {
		text := scanner.Text()
		trimmed := strings.TrimSpace(text)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		var targets []string
		for _, match := range inlineLinkPattern.FindAllStringSubmatch(text, -1) {
			targets = append(targets, match[1])
		}
		if match := referenceLinkPattern.FindStringSubmatch(text); len(match) == 2 {
			targets = append(targets, match[1])
		}
		for _, raw := range targets {
			target, local := localTarget(raw)
			if !local {
				continue
			}
			checked++
			if filepath.IsAbs(target) {
				failures = append(failures, fmt.Sprintf("%s:%d: absolute local link is not portable: %s", relative(root, path), line, raw))
				continue
			}
			resolved := filepath.Clean(filepath.Join(filepath.Dir(path), filepath.FromSlash(target)))
			if !withinRoot(root, resolved) {
				failures = append(failures, fmt.Sprintf("%s:%d: local link escapes repository: %s", relative(root, path), line, raw))
				continue
			}
			if _, err := os.Stat(resolved); err != nil {
				failures = append(failures, fmt.Sprintf("%s:%d: missing local link target %s", relative(root, path), line, raw))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, nil, err
	}
	return checked, failures, nil
}

func localTarget(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "<") {
		if end := strings.Index(raw, ">"); end >= 0 {
			raw = raw[1:end]
		}
	} else if fields := strings.Fields(raw); len(fields) > 0 {
		raw = fields[0]
	}
	if raw == "" || strings.HasPrefix(raw, "#") {
		return "", false
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" {
		return "", false
	}
	path, err := url.PathUnescape(parsed.Path)
	if err != nil || path == "" {
		return "", false
	}
	return path, true
}

func withinRoot(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func relative(root, path string) string {
	value, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(value)
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "documentation link gate:", err)
	os.Exit(1)
}
