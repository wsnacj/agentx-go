package browserd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestNewManagerDoesNotResolveOrMaterializeBundledLaunch(t *testing.T) {
	emptyBin := filepath.Join(t.TempDir(), "empty-bin")
	if err := os.MkdirAll(emptyBin, 0o755); err != nil {
		t.Fatalf("mkdir empty bin: %v", err)
	}
	t.Setenv("PATH", emptyBin)
	stateRoot := filepath.Join(t.TempDir(), "not-created", "browserd")
	manager, err := NewManager(ManagerOptions{
		Plan: Plan{
			Enabled:       true,
			Command:       bundledBrowserdCommandAuto,
			Endpoint:      "http://127.0.0.1:43123",
			StateRoot:     stateRoot,
			ProfilesRoot:  filepath.Join(stateRoot, "profiles"),
			ArtifactsRoot: filepath.Join(stateRoot, "artifacts"),
			LogsRoot:      filepath.Join(stateRoot, "logs"),
		},
	})
	if err != nil {
		t.Fatalf("NewManager must not require node during construction: %v", err)
	}
	if manager.launchResolved || manager.launchCommand != "" || len(manager.launchArgs) != 0 {
		t.Fatalf("constructor resolved bundled launch: %#v", manager)
	}
	if _, err := os.Stat(stateRoot); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("constructor materialized state root, stat error = %v", err)
	}
	if err := manager.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestBootstrapBundledNodeModulesHonorsContextDeadline(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script npm stub is not portable to Windows")
	}
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	npmPath := filepath.Join(binDir, "npm")
	if err := os.WriteFile(npmPath, []byte("#!/bin/sh\nexec /bin/sleep 30\n"), 0o755); err != nil {
		t.Fatalf("write npm stub: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	targetRoot := filepath.Join(t.TempDir(), "node")
	stateRoot := t.TempDir()
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		t.Fatalf("mkdir target root: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Millisecond)
	defer cancel()
	started := time.Now()
	err := bootstrapBundledNodeModulesContext(ctx, targetRoot, stateRoot)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("bootstrap error = %v, want context deadline exceeded", err)
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("bootstrap ignored deadline: elapsed=%s", elapsed)
	}
}

func TestBootstrapBundledNodeModulesBoundsAndRedactsCommandOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script npm stub is not portable to Windows")
	}
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	const sentinel = "raw-bootstrap-secret-sentinel"
	script := "#!/bin/sh\n" +
		"i=0\n" +
		"while [ \"$i\" -lt 12000 ]; do\n" +
		"  printf '" + sentinel + "\\n'\n" +
		"  i=$((i + 1))\n" +
		"done\n" +
		"exit 7\n"
	npmPath := filepath.Join(binDir, "npm")
	if err := os.WriteFile(npmPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write npm stub: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	targetRoot := filepath.Join(t.TempDir(), "node")
	stateRoot := t.TempDir()
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		t.Fatalf("mkdir target root: %v", err)
	}

	err := bootstrapBundledNodeModulesContext(context.Background(), targetRoot, stateRoot)
	if !errors.Is(err, ErrProcessOutputLimitExceeded) {
		t.Fatalf("bootstrap error = %v, want bounded output error", err)
	}
	message := err.Error()
	if strings.Contains(message, sentinel) {
		t.Fatalf("bootstrap error leaked command output: %q", message)
	}
	if !strings.Contains(message, "stdout_limit_exceeded=true") {
		t.Fatalf("bootstrap error lacks safe capture summary: %q", message)
	}
}
