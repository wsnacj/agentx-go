//go:build ignore

// check_developer_preview_api verifies the focused M3D package classification,
// Chinese Reference coverage, and signatures of Developer Preview candidates.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

const manifestPath = "docs/reference/developer-preview-packages.tsv"

var allowedMaturity = map[string]bool{
	"developer_preview_candidate": true,
	"experimental_extension":      true,
	"internalization_candidate":   true,
}

type entry struct {
	dir      string
	maturity string
	doc      string
	hash     string
}

func main() {
	printSignatures := flag.Bool("print-signatures", false, "print candidate go doc hashes without checking the baseline")
	flag.Parse()

	root, err := os.Getwd()
	check(err)
	entries, err := readManifest(filepath.Join(root, manifestPath))
	check(err)
	discovered, err := discoverPackages(root)
	check(err)
	check(packageSetDiff(entries, discovered))

	candidates := 0
	for _, item := range entries {
		if !allowedMaturity[item.maturity] {
			check(fmt.Errorf("%s: unsupported maturity %q", item.dir, item.maturity))
		}
		docPath := filepath.Join(root, item.doc)
		info, err := os.Stat(docPath)
		if err != nil || info.IsDir() || info.Size() == 0 {
			check(fmt.Errorf("%s: missing non-empty Chinese Reference %s", item.dir, item.doc))
		}
		content, err := os.ReadFile(docPath)
		check(err)
		if !containsHan(content) {
			check(fmt.Errorf("%s: Reference does not contain Chinese text: %s", item.dir, item.doc))
		}
		if item.maturity != "developer_preview_candidate" {
			continue
		}
		candidates++
		actual, err := signatureHash(filepath.Join(root, item.dir))
		check(err)
		if *printSignatures {
			fmt.Printf("%s\t%s\n", item.dir, actual)
			continue
		}
		if item.hash == "" || item.hash == "-" || item.hash == "PENDING" {
			check(fmt.Errorf("%s: missing signature baseline", item.dir))
		}
		if actual != item.hash {
			check(fmt.Errorf("%s: signature drift: got %s, want %s", item.dir, actual, item.hash))
		}
	}
	if !*printSignatures {
		fmt.Printf("agentx-developer-preview-api-gate-ok:packages=%d:candidates=%d\n", len(entries), candidates)
	}
}

func containsHan(content []byte) bool {
	for _, value := range string(content) {
		if unicode.Is(unicode.Han, value) {
			return true
		}
	}
	return false
}

func readManifest(path string) ([]entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	seen := map[string]bool{}
	var entries []entry
	scanner := bufio.NewScanner(file)
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		fields := strings.Split(text, "\t")
		if len(fields) != 4 {
			return nil, fmt.Errorf("%s:%d: want four tab-separated fields", path, line)
		}
		item := entry{dir: fields[0], maturity: fields[1], doc: fields[2], hash: fields[3]}
		clean := filepath.Clean(item.dir)
		if filepath.IsAbs(item.dir) || clean != item.dir || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("%s:%d: unsafe package directory %q", path, line, item.dir)
		}
		if seen[item.dir] {
			return nil, fmt.Errorf("%s:%d: duplicate package directory %q", path, line, item.dir)
		}
		seen[item.dir] = true
		entries = append(entries, item)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func discoverPackages(root string) ([]string, error) {
	set := map[string]bool{}
	for _, module := range []string{".", "components", "runtime"} {
		command := exec.Command("go", "list", "-f", "{{.Dir}}", "./...")
		command.Dir = filepath.Join(root, module)
		command.Env = append(os.Environ(), "GOWORK=off")
		output, err := command.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("go list %s: %w: %s", module, err, bytes.TrimSpace(output))
		}
		for _, path := range strings.Fields(string(output)) {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return nil, err
			}
			set[filepath.Clean(relative)] = true
		}
	}
	var paths []string
	for path := range set {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

func packageSetDiff(entries []entry, discovered []string) error {
	expected := make([]string, 0, len(entries))
	for _, item := range entries {
		expected = append(expected, item.dir)
	}
	sort.Strings(expected)
	if strings.Join(expected, "\n") == strings.Join(discovered, "\n") {
		return nil
	}
	return fmt.Errorf("classified package set does not match source\nclassified:\n%s\ndiscovered:\n%s", strings.Join(expected, "\n"), strings.Join(discovered, "\n"))
}

func signatureHash(dir string) (string, error) {
	command := exec.Command("go", "doc", "-all", ".")
	command.Dir = dir
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("go doc %s: %w: %s", dir, err, bytes.TrimSpace(output))
	}
	sum := sha256.Sum256(output)
	return fmt.Sprintf("%x", sum[:]), nil
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "developer preview API gate:", err)
	os.Exit(1)
}
