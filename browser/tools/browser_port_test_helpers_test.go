package tools

import (
	"context"
	"os"
	"path/filepath"
)

func testBrowserArtifactPublisher(
	ctx context.Context,
	root string,
	targetPath string,
	write func(context.Context, *os.File, string) error,
) (int64, error) {
	path := targetPath
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return 0, err
	}
	stage, err := os.CreateTemp(filepath.Dir(path), ".agentx-browser-stage-*")
	if err != nil {
		return 0, err
	}
	stagePath := stage.Name()
	committed := false
	defer func() {
		_ = stage.Close()
		if !committed {
			_ = os.Remove(stagePath)
		}
	}()
	if err := write(ctx, stage, stagePath); err != nil {
		return 0, err
	}
	if err := stage.Close(); err != nil {
		return 0, err
	}
	if err := os.Rename(stagePath, path); err != nil {
		return 0, err
	}
	committed = true
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
