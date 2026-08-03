package process_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	stdruntime "runtime"
	"strings"
	"testing"
	"time"

	process "github.com/wsnacj/agentx-go/tools/process"
)

func TestLocalAdapterRunAndNonZeroExit(t *testing.T) {
	root := t.TempDir()
	adapter := process.NewLocalAdapter(process.LocalOptions{Root: root, MaxOutputBytes: 256})
	result, err := adapter.Run(context.Background(), process.Command{Command: "printf ready; printf warning >&2; exit 3"})
	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 3 || result.Stdout != "ready" || result.Stderr != "warning" || result.Workdir != root {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestLocalAdapterBoundsOutput(t *testing.T) {
	if stdruntime.GOOS == "windows" {
		t.Skip("shell loop fixture is unix-oriented")
	}
	adapter := process.NewLocalAdapter(process.LocalOptions{Root: t.TempDir(), MaxOutputBytes: 64})
	result, err := adapter.Run(context.Background(), process.Command{
		Command: "i=0; while [ $i -lt 200 ]; do printf x; i=$((i+1)); done",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Stdout) != 64 || !result.StdoutTruncated || result.StderrTruncated {
		t.Fatalf("unexpected bounded output: %#v", result)
	}
}

func TestLocalAdapterTimeoutCleansProcessGroup(t *testing.T) {
	if stdruntime.GOOS == "windows" {
		t.Skip("process-group fixture is unix-oriented")
	}
	root := t.TempDir()
	marker := filepath.Join(root, "child-survived")
	adapter := process.NewLocalAdapter(process.LocalOptions{Root: root})
	result, err := adapter.Run(context.Background(), process.Command{
		Command: "sleep 0.4; touch " + shellQuote(marker) + " & wait",
		Timeout: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("adapter-owned timeout should be a result: %v", err)
	}
	if !result.TimedOut || result.Status != "timed_out" || result.ExitCode != -1 || result.Termination == nil || !result.Termination.ProcessGroup {
		t.Fatalf("unexpected timeout result: %#v", result)
	}
	time.Sleep(650 * time.Millisecond)
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("child process survived timeout: %v", err)
	}
}

func TestLocalAdapterCallerCancellationIsTyped(t *testing.T) {
	adapter := process.NewLocalAdapter(process.LocalOptions{Root: t.TempDir()})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := adapter.Run(ctx, process.Command{Command: "printf should-not-run"})
	if !errors.Is(err, context.Canceled) || result.Status != "cancelled" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	typed, ok := process.AsError(err)
	if !ok || typed.Code != process.ErrorCodeContextCanceled {
		t.Fatalf("typed error=%#v ok=%t", typed, ok)
	}
}

func TestLocalAdapterRejectsSymlinkEscape(t *testing.T) {
	if stdruntime.GOOS == "windows" {
		t.Skip("symlink fixture is platform-sensitive")
	}
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "outside")); err != nil {
		t.Skip(err)
	}
	adapter := process.NewLocalAdapter(process.LocalOptions{Root: root})
	_, err := adapter.Run(context.Background(), process.Command{Command: "pwd", Workdir: "outside"})
	typed, ok := process.AsError(err)
	if !ok || typed.Code != process.ErrorCodeWorkdirResolution {
		t.Fatalf("typed error=%#v ok=%t err=%v", typed, ok, err)
	}
}

func TestLocalAdapterPreservesCommandFailureForMissingWorkdir(t *testing.T) {
	adapter := process.NewLocalAdapter(process.LocalOptions{Root: t.TempDir()})
	_, err := adapter.Run(context.Background(), process.Command{Command: "pwd", Workdir: "missing"})
	typed, ok := process.AsError(err)
	if !ok || typed.Code != process.ErrorCodeCommandFailed {
		t.Fatalf("typed error=%#v ok=%t err=%v", typed, ok, err)
	}
}

func TestLocalAdapterListIsReadOnlyAndBoundedByRows(t *testing.T) {
	adapter := process.NewLocalAdapter(process.LocalOptions{Root: t.TempDir()})
	result, err := adapter.List(context.Background(), process.ListRequest{Limit: 2})
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			t.Skipf("local process inspection is blocked by the test sandbox: %v", err)
		}
		t.Fatal(err)
	}
	if len(result.Lines) == 0 || len(result.Lines) > 2 {
		t.Fatalf("unexpected listing: %#v", result)
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
