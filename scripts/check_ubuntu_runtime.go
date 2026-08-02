//go:build ignore

// check_ubuntu_runtime runs the approved M6C Ubuntu runtime and distribution
// matrix. A cross-compiled binary is intentionally insufficient: this program
// exits unless it is itself running on Linux amd64.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var sourceModules = []string{".", "components", "runtime", "extensions"}

const expectedGoVersion = "go1.25.12"

func main() {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		fail(fmt.Errorf("requires a real linux/amd64 host, got %s/%s", runtime.GOOS, runtime.GOARCH))
	}
	if runtime.Version() != expectedGoVersion {
		fail(fmt.Errorf("requires %s, got %s", expectedGoVersion, runtime.Version()))
	}
	root, err := os.Getwd()
	check(err)
	revision := strings.TrimSpace(run(root, nil, "git", "rev-parse", "HEAD"))
	if revision == "" {
		fail(fmt.Errorf("empty source revision"))
	}

	env := append(os.Environ(), "GOWORK=off", "CGO_ENABLED=1")
	fmt.Printf("agentx-ubuntu-runtime-host:goos=%s:goarch=%s:revision=%s\n", runtime.GOOS, runtime.GOARCH, revision)
	fmt.Print(run(root, env, "uname", "-a"))
	fmt.Print(run(root, env, "go", "version"))
	fmt.Print(run(root, env, "go", "env", "GOOS", "GOARCH", "CGO_ENABLED", "GOTOOLCHAIN"))

	// The distribution lane includes API/platform signatures, docs, an empty
	// private-VCS cache, a frozen module-cache consumer, artifact provenance,
	// and normal test/vet/tidy/list checks for all four modules.
	fmt.Print(run(root, env, "go", "run", "./scripts/check_developer_preview_distribution.go", "-fresh-cache", "-read-only-cache", "-full"))

	for _, relative := range sourceModules {
		dir := filepath.Join(root, relative)
		fmt.Print(run(dir, env, "go", "test", "-race", "./..."))
	}

	fmt.Printf("agentx-ubuntu-runtime-ok:goos=%s:goarch=%s:cgo=1:modules=%d:revision=%s:ubuntu_runtime_ready=true:public_beta_ready=false\n", runtime.GOOS, runtime.GOARCH, len(sourceModules), revision)
}

func run(dir string, env []string, name string, args ...string) string {
	command := exec.Command(name, args...)
	command.Dir = dir
	if env != nil {
		command.Env = env
	}
	output, err := command.CombinedOutput()
	if err != nil {
		fail(fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, bytes.TrimSpace(output)))
	}
	return string(output)
}

func check(err error) {
	if err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "Ubuntu runtime gate:", err)
	os.Exit(1)
}
