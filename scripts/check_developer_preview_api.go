//go:build ignore

// check_developer_preview_api verifies the focused Developer Preview package classification,
// Chinese Reference coverage, readable API snapshots, public type closure, and target-platform
// signatures of Developer Preview candidates.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const (
	manifestPath = "docs/reference/developer-preview-packages.tsv"
	snapshotDir  = "docs/reference/api-snapshots"
)

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

type target struct {
	goos   string
	goarch string
	cgo    int
}

func (t target) label() string {
	if t.goos == "" {
		return runtime.GOOS + "/" + runtime.GOARCH + "/host-cgo"
	}
	return fmt.Sprintf("%s/%s/cgo=%d", t.goos, t.goarch, t.cgo)
}

func main() {
	printSignatures := flag.Bool("print-signatures", false, "print candidate go doc hashes without checking the baseline")
	updateSnapshots := flag.Bool("update-snapshots", false, "write readable go doc snapshots for the selected target")
	checkPlatforms := flag.Bool("check-platforms", false, "verify darwin/arm64 and linux/amd64 CGO-disabled signatures")
	targetGOOS := flag.String("target-goos", "", "GOOS used by child go list/go doc commands")
	targetGOARCH := flag.String("target-goarch", "", "GOARCH used by child go list/go doc commands")
	targetCGO := flag.Int("target-cgo", -1, "CGO_ENABLED used by child commands; -1 inherits the host setting")
	flag.Parse()

	if (*targetGOOS == "") != (*targetGOARCH == "") {
		check(fmt.Errorf("target-goos and target-goarch must be set together"))
	}
	if *targetCGO < -1 || *targetCGO > 1 {
		check(fmt.Errorf("target-cgo must be -1, 0, or 1"))
	}
	if *checkPlatforms && (*targetGOOS != "" || *targetCGO != -1 || *updateSnapshots) {
		check(fmt.Errorf("check-platforms cannot be combined with explicit target or update-snapshots"))
	}

	root, err := os.Getwd()
	check(err)
	entries, err := readManifest(filepath.Join(root, manifestPath))
	check(err)

	targets := []target{{goos: *targetGOOS, goarch: *targetGOARCH, cgo: *targetCGO}}
	if *checkPlatforms {
		targets = []target{
			{goos: "darwin", goarch: "arm64", cgo: 0},
			{goos: "linux", goarch: "amd64", cgo: 0},
		}
	}

	candidates := 0
	var referenceHashes map[string]string
	for targetIndex, selected := range targets {
		discovered, err := discoverPackages(root, selected)
		check(err)
		check(packageSetDiff(entries, discovered))

		actualHashes := map[string]string{}
		for _, item := range entries {
			if !allowedMaturity[item.maturity] {
				check(fmt.Errorf("%s: unsupported maturity %q", item.dir, item.maturity))
			}
			if targetIndex == 0 {
				checkChineseReference(root, item)
			}
			if item.maturity != "developer_preview_candidate" {
				continue
			}
			if targetIndex == 0 {
				candidates++
			}
			checkPublicDependencies(root, item.dir, selected)
			output, err := docOutput(filepath.Join(root, item.dir), selected)
			check(err)
			actual := hashBytes(output)
			actualHashes[item.dir] = actual
			if *printSignatures {
				fmt.Printf("%s\t%s\t%s\n", item.dir, selected.label(), actual)
				continue
			}
			if item.hash == "" || item.hash == "-" || item.hash == "PENDING" {
				check(fmt.Errorf("%s: missing signature baseline", item.dir))
			}
			if actual != item.hash {
				check(fmt.Errorf("%s [%s]: signature drift: got %s, want %s", item.dir, selected.label(), actual, item.hash))
			}
			checkSnapshot(root, item.dir, output, *updateSnapshots)
		}
		if targetIndex == 0 {
			referenceHashes = actualHashes
			continue
		}
		for dir, actual := range actualHashes {
			if actual != referenceHashes[dir] {
				check(fmt.Errorf("%s: platform signature drift: %s=%s, %s=%s", dir, targets[0].label(), referenceHashes[dir], selected.label(), actual))
			}
		}
	}
	if !*printSignatures {
		fmt.Printf("agentx-developer-preview-api-gate-ok:packages=%d:candidates=%d:targets=%d\n", len(entries), candidates, len(targets))
	}
}

func checkChineseReference(root string, item entry) {
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

func discoverPackages(root string, selected target) ([]string, error) {
	set := map[string]bool{}
	for _, module := range []string{".", "components", "runtime", "extensions", "scenes"} {
		command := exec.Command("go", "list", "-f", "{{.Dir}}", "./...")
		command.Dir = filepath.Join(root, module)
		command.Env = commandEnv(selected)
		output, diagnostics, err := splitOutput(command)
		if err != nil {
			return nil, fmt.Errorf("go list %s [%s]: %w: %s", module, selected.label(), err, diagnostics)
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

func docOutput(dir string, selected target) ([]byte, error) {
	command := exec.Command("go", "doc", "-all", ".")
	command.Dir = dir
	command.Env = commandEnv(selected)
	output, diagnostics, err := splitOutput(command)
	if err != nil {
		return nil, fmt.Errorf("go doc %s [%s]: %w: %s", dir, selected.label(), err, diagnostics)
	}
	return output, nil
}

func checkSnapshot(root, dir string, output []byte, update bool) {
	path := filepath.Join(root, snapshotDir, snapshotName(dir))
	if update {
		check(os.MkdirAll(filepath.Dir(path), 0o755))
		check(os.WriteFile(path, output, 0o644))
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		check(fmt.Errorf("%s: missing readable API snapshot %s", dir, path))
	}
	if !bytes.Equal(output, want) {
		check(fmt.Errorf("%s: readable API snapshot drift: %s", dir, path))
	}
}

func snapshotName(dir string) string {
	if dir == "." {
		return "root.txt"
	}
	return strings.ReplaceAll(filepath.ToSlash(dir), "/", "__") + ".txt"
}

func hashBytes(content []byte) string {
	sum := sha256.Sum256(content)
	return fmt.Sprintf("%x", sum[:])
}

type packageFiles struct {
	Dir      string
	GoFiles  []string
	CgoFiles []string
}

func checkPublicDependencies(root, dir string, selected target) {
	command := exec.Command("go", "list", "-json", ".")
	command.Dir = filepath.Join(root, dir)
	command.Env = commandEnv(selected)
	output, diagnostics, err := splitOutput(command)
	if err != nil {
		check(fmt.Errorf("go list public dependency closure %s [%s]: %w: %s", dir, selected.label(), err, diagnostics))
	}
	var listed packageFiles
	check(json.Unmarshal(output, &listed))
	files := append(append([]string{}, listed.GoFiles...), listed.CgoFiles...)
	for _, name := range files {
		checkPublicFile(filepath.Join(listed.Dir, name), dir)
	}
}

func splitOutput(command *exec.Cmd) ([]byte, string, error) {
	var diagnostics bytes.Buffer
	command.Stderr = &diagnostics
	output, err := command.Output()
	return output, strings.TrimSpace(diagnostics.String()), err
}

func checkPublicFile(path, packageDir string) {
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	check(err)
	imports := map[string]string{}
	for _, imported := range parsed.Imports {
		value, err := strconv.Unquote(imported.Path.Value)
		check(err)
		name := filepath.Base(value)
		if imported.Name != nil {
			name = imported.Name.Name
		}
		imports[name] = value
	}
	inspect := func(symbol string, node ast.Node) {
		if node == nil {
			return
		}
		ast.Inspect(node, func(current ast.Node) bool {
			selector, ok := current.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			qualifier, ok := selector.X.(*ast.Ident)
			if !ok {
				return true
			}
			importPath, ok := imports[qualifier.Name]
			if ok && forbiddenPublicDependency(importPath) {
				check(fmt.Errorf("%s: exported %s leaks forbidden dependency %s", packageDir, symbol, importPath))
			}
			return true
		})
	}
	for _, declaration := range parsed.Decls {
		switch value := declaration.(type) {
		case *ast.GenDecl:
			for _, spec := range value.Specs {
				switch item := spec.(type) {
				case *ast.TypeSpec:
					if item.Name.IsExported() {
						inspect(item.Name.Name, item.Type)
					}
				case *ast.ValueSpec:
					for _, name := range item.Names {
						if name.IsExported() {
							inspect(name.Name, item.Type)
						}
					}
				}
			}
		case *ast.FuncDecl:
			if !value.Name.IsExported() || (value.Recv != nil && !exportedReceiver(value.Recv)) {
				continue
			}
			inspect(value.Name.Name, value.Type)
		}
	}
}

func exportedReceiver(receivers *ast.FieldList) bool {
	if receivers == nil || len(receivers.List) == 0 {
		return true
	}
	var receiverName func(ast.Expr) string
	receiverName = func(expression ast.Expr) string {
		switch value := expression.(type) {
		case *ast.Ident:
			return value.Name
		case *ast.StarExpr:
			return receiverName(value.X)
		case *ast.IndexExpr:
			return receiverName(value.X)
		case *ast.IndexListExpr:
			return receiverName(value.X)
		default:
			return ""
		}
	}
	return ast.IsExported(receiverName(receivers.List[0].Type))
}

func forbiddenPublicDependency(path string) bool {
	return path == "github.com/wsnacj/agentx-go/runtime/controlcontract" ||
		strings.HasPrefix(path, "hs/") ||
		strings.Contains(path, "/internal/")
}

func commandEnv(selected target) []string {
	environment := make([]string, 0, len(os.Environ())+4)
	for _, value := range os.Environ() {
		if strings.HasPrefix(value, "GOWORK=") || strings.HasPrefix(value, "GOOS=") || strings.HasPrefix(value, "GOARCH=") || strings.HasPrefix(value, "CGO_ENABLED=") {
			continue
		}
		environment = append(environment, value)
	}
	environment = append(environment, "GOWORK=off")
	if selected.goos != "" {
		environment = append(environment, "GOOS="+selected.goos, "GOARCH="+selected.goarch)
	}
	if selected.cgo >= 0 {
		environment = append(environment, fmt.Sprintf("CGO_ENABLED=%d", selected.cgo))
	}
	return environment
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "developer preview API gate:", err)
	os.Exit(1)
}
