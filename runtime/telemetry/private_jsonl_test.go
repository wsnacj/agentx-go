package telemetry

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPrivateJSONLSinksUsePrivateModes(t *testing.T) {
	tests := []struct {
		name  string
		new   func(string) (Sink, error)
		event Event
	}{
		{
			name: "raw",
			new: func(path string) (Sink, error) {
				return NewJSONLSink(path)
			},
			event: Event{Component: "runner", Name: "run.start"},
		},
		{
			name: "tool projection",
			new: func(path string) (Sink, error) {
				return NewToolEventJSONLSink(path)
			},
			event: Event{Component: "tool", Name: "tool.finish", Status: "ok"},
		},
		{
			name: "semantic projection",
			new: func(path string) (Sink, error) {
				return NewSemanticRunEventJSONLSink(path)
			},
			event: Event{Component: "runner", Name: "run.start", Attrs: map[string]any{"resume": true}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := filepath.Join(t.TempDir(), "private-observation")
			path := filepath.Join(directory, "events.jsonl")
			sink, err := test.new(path)
			if err != nil {
				t.Fatalf("new sink: %v", err)
			}
			if err := sink.Emit(context.Background(), test.event); err != nil {
				t.Fatalf("emit: %v", err)
			}
			assertPathMode(t, directory, privateObservationDirectoryMode)
			assertPathMode(t, path, privateObservationFileMode)
		})
	}
}

func TestPrivateJSONLSinkRepairsExistingFileMode(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "private-observation")
	if err := os.MkdirAll(directory, privateObservationDirectoryMode); err != nil {
		t.Fatalf("create directory: %v", err)
	}
	path := filepath.Join(directory, "events.jsonl")
	if err := os.WriteFile(path, []byte("existing\n"), 0o644); err != nil {
		t.Fatalf("write existing event file: %v", err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatalf("set existing event mode: %v", err)
	}
	sink, err := NewJSONLSink(path)
	if err != nil {
		t.Fatalf("new sink: %v", err)
	}
	assertPathMode(t, path, privateObservationFileMode)
	if err := sink.Emit(context.Background(), Event{Component: "runner", Name: "run.start"}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	assertPathMode(t, path, privateObservationFileMode)
}

func TestPrivateJSONLSinkRejectsSymlinkFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on Windows")
	}
	directory := filepath.Join(t.TempDir(), "private-observation")
	if err := os.MkdirAll(directory, privateObservationDirectoryMode); err != nil {
		t.Fatalf("create directory: %v", err)
	}
	target := filepath.Join(directory, "target.jsonl")
	if err := os.WriteFile(target, []byte("untouched\n"), 0o600); err != nil {
		t.Fatalf("write target: %v", err)
	}
	path := filepath.Join(directory, "events.jsonl")
	if err := os.Symlink(target, path); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	_, err := NewJSONLSink(path)
	if err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("expected symlink rejection, got %v", err)
	}
	raw, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("read target: %v", readErr)
	}
	if string(raw) != "untouched\n" {
		t.Fatalf("symlink target was modified: %q", raw)
	}
}

func assertPathMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != want.Perm() {
		t.Fatalf("mode %s: got %#o want %#o", path, got, want)
	}
}
