//go:build ignore

// check_cleanroom_consumer copies the fixed-version consumer outside the source
// checkout and verifies it only from the local immutable module cache.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var expectedModules = []string{
	"github.com/wsnacj/agentx-go",
	"github.com/wsnacj/agentx-go/components",
	"github.com/wsnacj/agentx-go/runtime",
	"github.com/wsnacj/agentx-go/extensions",
}

func main() {
	freshCache := flag.Bool("fresh-cache", false, "download fixed modules into an empty temporary module and build cache")
	flag.Parse()

	root, err := os.Getwd()
	check(err)
	versionBytes, err := os.ReadFile(filepath.Join(root, "docs/reference/developer-preview-version.txt"))
	check(err)
	version := strings.TrimSpace(string(versionBytes))

	consumer := filepath.Join(root, "conformance/consumer")
	goMod, err := os.ReadFile(filepath.Join(consumer, "go.mod"))
	check(err)
	if bytes.Contains(goMod, []byte("replace ")) || bytes.Contains(goMod, []byte("replace(")) || bytes.Contains(goMod, []byte("replace (")) {
		check(fmt.Errorf("conformance consumer must not contain replace directives"))
	}

	temporary, err := os.MkdirTemp("", "agentx-go-cleanroom-consumer-")
	check(err)
	defer os.RemoveAll(temporary)
	consumerCopy := filepath.Join(temporary, "consumer")
	check(copyConsumer(consumer, consumerCopy))

	mode := "module-cache"
	env := append(os.Environ(),
		"GOWORK=off",
		"GOPROXY=off",
		"GONOSUMDB=github.com/wsnacj/agentx-go",
		"GOFLAGS=-mod=readonly",
	)
	if *freshCache {
		mode = "fresh-vcs-cache"
		env = append(env,
			"GOPROXY=direct",
			"GOPRIVATE=github.com/wsnacj/agentx-go",
			"GOMODCACHE="+filepath.Join(temporary, "gomodcache"),
			"GOCACHE="+filepath.Join(temporary, "gocache"),
		)
		run(consumerCopy, env, "go", "mod", "download")
	}
	run(consumerCopy, env, "go", "mod", "verify")
	modules := run(consumerCopy, env, "go", "list", "-m", "-f", "{{.Path}} {{.Version}}", "all")
	for _, module := range expectedModules {
		want := module + " " + version
		if !containsLine(modules, want) {
			check(fmt.Errorf("clean-room module list missing %q", want))
		}
	}
	run(consumerCopy, env, "go", "test", "./...")
	output := strings.TrimSpace(run(consumerCopy, env, "go", "run", "."))
	if !strings.HasPrefix(output, "agentx-core-developer-preview-ok:") {
		check(fmt.Errorf("unexpected clean-room output %q", output))
	}
	fmt.Printf("agentx-cleanroom-consumer-ok:version=%s:modules=%d:source=%s\n", version, len(expectedModules), mode)
}

func copyConsumer(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		defer input.Close()
		output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(output, input)
		closeErr := output.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}

func run(dir string, env []string, name string, args ...string) string {
	command := exec.Command(name, args...)
	command.Dir = dir
	command.Env = env
	output, err := command.CombinedOutput()
	if err != nil {
		check(fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, bytes.TrimSpace(output)))
	}
	return string(output)
}

func containsLine(output, want string) bool {
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "clean-room consumer gate:", err)
	os.Exit(1)
}
