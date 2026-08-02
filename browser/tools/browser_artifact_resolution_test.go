package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

type copyingBrowserArtifactBackend struct {
	fakeBrowserBackend
	resolveReqs []browserArtifactResolveRequest
}

func (b *copyingBrowserArtifactBackend) ResolveBrowserArtifact(_ context.Context, req browserArtifactResolveRequest) (string, error) {
	b.resolveReqs = append(b.resolveReqs, req)
	if err := os.MkdirAll(filepath.Dir(req.RequestedPath), 0o755); err != nil {
		return "", err
	}
	blob, err := os.ReadFile(req.BackendPath)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(req.RequestedPath, blob, 0o644); err != nil {
		return "", err
	}
	return req.RequestedPath, nil
}

func TestResolveBrowserArtifactOutputCopiesExistingBackendPathOutsideWorkspace(t *testing.T) {
	root := t.TempDir()
	backendRoot := t.TempDir()
	backendPath := filepath.Join(backendRoot, "browserd-state", "artifacts", "download.pdf")
	if err := os.MkdirAll(filepath.Dir(backendPath), 0o755); err != nil {
		t.Fatalf("mkdir backend artifact: %v", err)
	}
	const content = "downloaded-pdf"
	if err := os.WriteFile(backendPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write backend artifact: %v", err)
	}

	requestedPath := filepath.Join(root, ".agentx", "browser", "download.pdf")
	backend := &copyingBrowserArtifactBackend{}
	localPath, displayPath, size, err := resolveBrowserArtifactOutput(
		context.Background(),
		backend,
		root,
		"wait_download",
		requestedPath,
		".agentx/browser/download.pdf",
		backendPath,
	)
	if err != nil {
		t.Fatalf("resolve browser artifact output: %v", err)
	}
	if len(backend.resolveReqs) != 1 {
		t.Fatalf("expected escaped backend artifact to go through resolver once, got %#v", backend.resolveReqs)
	}
	if localPath != requestedPath || displayPath != ".agentx/browser/download.pdf" || size != int64(len(content)) {
		t.Fatalf("unexpected resolved artifact local=%q display=%q size=%d", localPath, displayPath, size)
	}
	got, err := os.ReadFile(requestedPath)
	if err != nil {
		t.Fatalf("read copied artifact: %v", err)
	}
	if string(got) != content {
		t.Fatalf("unexpected copied artifact content %q", string(got))
	}
}
