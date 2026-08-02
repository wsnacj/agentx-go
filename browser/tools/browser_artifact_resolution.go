package tools

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type browserArtifactResolveRequest struct {
	Kind          string
	RequestedPath string
	BackendPath   string
}

type browserArtifactResolver interface {
	ResolveBrowserArtifact(context.Context, browserArtifactResolveRequest) (string, error)
}

var errBrowserArtifactPublicationSkipped = errors.New("browser artifact publication skipped")

type browserArtifactPublicationStageContextKey struct{}

type browserArtifactPublicationStage struct {
	path       string
	targetPath string
	file       *os.File
}

func publishBrowserArtifactOutput(
	ctx context.Context,
	publish BrowserArtifactPublisher,
	backend BrowserBackend,
	root string,
	kind string,
	targetPath string,
	produce func(string) (string, bool, error),
) (int64, bool, error) {
	if produce == nil {
		return 0, false, fmt.Errorf("browser artifact producer is required")
	}
	if publish == nil {
		return 0, false, fmt.Errorf("browser artifact publisher is required")
	}
	rootDir, err := resolveRootDir(root)
	if err != nil {
		return 0, false, err
	}
	size, err := publish(ctx, rootDir, targetPath, func(ctx context.Context, stage *os.File, stagePath string) error {
		stageCtx := context.WithValue(ctx, browserArtifactPublicationStageContextKey{}, browserArtifactPublicationStage{
			path:       filepath.Clean(stagePath),
			targetPath: filepath.Clean(targetPath),
			file:       stage,
		})
		backendPath, publish, err := produce(stagePath)
		if err != nil {
			return err
		}
		if !publish {
			return errBrowserArtifactPublicationSkipped
		}
		if browserArtifactProducerLeftEmptyStage(stagePath, backendPath) {
			return fmt.Errorf("browser artifact output missing: %s", stagePath)
		}
		localPath, _, _, err := resolveBrowserArtifactOutput(
			stageCtx,
			backend,
			rootDir,
			kind,
			stagePath,
			"",
			backendPath,
		)
		if err != nil {
			return err
		}
		if browserArtifactSamePath(localPath, stagePath) {
			return nil
		}
		return copyBrowserArtifactToStage(stageCtx, localPath, stagePath)
	})
	if errors.Is(err, errBrowserArtifactPublicationSkipped) {
		return 0, false, nil
	}
	return size, err == nil, err
}

func browserArtifactPublicationStageFile(ctx context.Context, targetPath string) (*os.File, bool) {
	stage, ok := browserArtifactPublicationStageFromContext(ctx, targetPath)
	if !ok {
		return nil, false
	}
	return stage.file, true
}

func browserArtifactPublicationRequestedPath(ctx context.Context, targetPath string) (string, bool) {
	stage, ok := browserArtifactPublicationStageFromContext(ctx, targetPath)
	if !ok || strings.TrimSpace(stage.targetPath) == "" {
		return "", false
	}
	return stage.targetPath, true
}

func browserArtifactPublicationStageFromContext(ctx context.Context, targetPath string) (browserArtifactPublicationStage, bool) {
	if ctx == nil {
		return browserArtifactPublicationStage{}, false
	}
	stage, ok := ctx.Value(browserArtifactPublicationStageContextKey{}).(browserArtifactPublicationStage)
	if !ok || stage.file == nil || !browserArtifactSamePath(stage.path, targetPath) {
		return browserArtifactPublicationStage{}, false
	}
	return stage, true
}

func browserArtifactProducerLeftEmptyStage(stagePath string, backendPath string) bool {
	backendPath = strings.TrimSpace(backendPath)
	if backendPath != "" && !browserArtifactSamePath(backendPath, stagePath) {
		return false
	}
	info, err := os.Lstat(stagePath)
	return err == nil && info.Mode().IsRegular() && info.Size() == 0
}

func copyBrowserArtifactToStage(ctx context.Context, sourcePath string, stagePath string) error {
	before, err := os.Lstat(sourcePath)
	if err != nil {
		return err
	}
	if before.Mode()&os.ModeSymlink != 0 || !before.Mode().IsRegular() {
		return fmt.Errorf("browser artifact source is unsafe: %s", sourcePath)
	}
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	opened, err := source.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(before, opened) {
		if err != nil {
			return err
		}
		return fmt.Errorf("browser artifact source changed before copy: %s", sourcePath)
	}
	target, ok := browserArtifactPublicationStageFile(ctx, stagePath)
	if !ok {
		return fmt.Errorf("browser artifact publication stage is unavailable: %s", stagePath)
	}
	if err := target.Truncate(0); err != nil {
		return err
	}
	if _, err := target.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := copyBrowserArtifactWithContext(ctx, target, source); err != nil {
		return err
	}
	if err := target.Sync(); err != nil {
		return err
	}
	after, err := os.Lstat(sourcePath)
	if err != nil || after.Mode()&os.ModeSymlink != 0 || !after.Mode().IsRegular() || !os.SameFile(opened, after) {
		if err != nil {
			return err
		}
		return fmt.Errorf("browser artifact source changed during copy: %s", sourcePath)
	}
	return nil
}

func copyBrowserArtifactWithContext(ctx context.Context, dst io.Writer, src io.Reader) error {
	if ctx == nil {
		ctx = context.Background()
	}
	buffer := make([]byte, 64<<10)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		read, readErr := src.Read(buffer)
		if read > 0 {
			written, writeErr := dst.Write(buffer[:read])
			if writeErr != nil {
				return writeErr
			}
			if written != read {
				return io.ErrShortWrite
			}
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}

func resolveBrowserArtifactOutput(ctx context.Context, backend BrowserBackend, root string, kind string, requestedPath string, requestedDisplay string, backendPath string) (string, string, int64, error) {
	requestedPath = strings.TrimSpace(requestedPath)
	requestedDisplay = strings.TrimSpace(requestedDisplay)
	backendPath = strings.TrimSpace(backendPath)
	candidatePath := firstNonEmpty(backendPath, requestedPath)
	resolver, hasResolver := backend.(browserArtifactResolver)

	if localPath, size, ok, err := statBrowserArtifactPath(candidatePath); err != nil {
		return "", "", 0, err
	} else if ok {
		displayPath, err := browserArtifactDisplayPath(root, localPath, requestedPath, requestedDisplay)
		if err != nil {
			if !browserArtifactShouldResolveExternalBackendPath(hasResolver, root, localPath, requestedPath, backendPath) {
				return "", "", 0, err
			}
		} else {
			return localPath, displayPath, size, nil
		}
	}

	if !hasResolver {
		if backendPath != "" && !browserArtifactSamePath(candidatePath, requestedPath) {
			return "", "", 0, fmt.Errorf("browser artifact path %q requires a backend artifact resolver", backendPath)
		}
		return "", "", 0, fmt.Errorf("browser artifact output missing: %s", candidatePath)
	}

	localPath, err := resolver.ResolveBrowserArtifact(ctx, browserArtifactResolveRequest{
		Kind:          strings.TrimSpace(kind),
		RequestedPath: requestedPath,
		BackendPath:   backendPath,
	})
	if err != nil {
		return "", "", 0, err
	}
	localPath = firstNonEmpty(strings.TrimSpace(localPath), requestedPath)
	statPath, size, ok, err := statBrowserArtifactPath(localPath)
	if err != nil {
		return "", "", 0, err
	}
	if !ok {
		return "", "", 0, fmt.Errorf("browser artifact output missing after resolver: %s", localPath)
	}
	displayPath, err := browserArtifactDisplayPath(root, statPath, requestedPath, requestedDisplay)
	if err != nil {
		return "", "", 0, err
	}
	return statPath, displayPath, size, nil
}

func browserArtifactShouldResolveExternalBackendPath(hasResolver bool, root string, localPath string, requestedPath string, backendPath string) bool {
	if !hasResolver || strings.TrimSpace(root) == "" || strings.TrimSpace(requestedPath) == "" || strings.TrimSpace(backendPath) == "" {
		return false
	}
	if browserArtifactSamePath(localPath, requestedPath) {
		return false
	}
	return browserArtifactSamePath(localPath, backendPath)
}

func statBrowserArtifactPath(path string) (string, int64, bool, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", 0, false, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", 0, false, nil
		}
		return "", 0, false, err
	}
	if info.IsDir() {
		return "", 0, false, fmt.Errorf("browser artifact output path is a directory: %s", path)
	}
	return path, info.Size(), true, nil
}

func browserArtifactDisplayPath(root string, localPath string, requestedPath string, requestedDisplay string) (string, error) {
	localPath = strings.TrimSpace(localPath)
	if localPath == "" {
		return "", fmt.Errorf("browser artifact display path requires a local path")
	}
	if requestedDisplay != "" && browserArtifactSamePath(localPath, requestedPath) {
		return requestedDisplay, nil
	}
	if root == "" {
		return filepath.ToSlash(localPath), nil
	}
	resolved, display, err := resolvePathWithinRoot(root, localPath)
	if err != nil {
		return "", fmt.Errorf("browser artifact path escaped workspace: %w", err)
	}
	if !browserArtifactSamePath(resolved, localPath) {
		return "", fmt.Errorf("browser artifact path resolved unexpectedly: %s", localPath)
	}
	return display, nil
}

func browserArtifactSamePath(left string, right string) bool {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" || right == "" {
		return false
	}
	return filepath.Clean(left) == filepath.Clean(right)
}
