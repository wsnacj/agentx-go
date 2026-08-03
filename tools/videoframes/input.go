package videoframes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const defaultMaxInputBytes int64 = 512 << 20

var (
	// ErrUnsafeFile marks a symlink, special file or identity-changing input.
	ErrUnsafeFile = errors.New("unsafe local input file")
	// ErrFileTooLarge marks a local input that exceeds the configured byte limit.
	ErrFileTooLarge = errors.New("local input file byte limit exceeded")
)

type ownedOutput struct {
	path     string
	display  string
	identity os.FileInfo
	retained bool
}

func createOwnedOutput(root, sourcePath, strategy string, intervalSec float64) (*ownedOutput, error) {
	rootDir, err := resolveRootDir(root)
	if err != nil {
		return nil, err
	}
	artifactRoot, _, err := resolveDirWithinRoot(root, filepath.Join(".agentx", "artifacts", "video_frames"))
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(artifactRoot, 0o700); err != nil {
		return nil, err
	}
	rootReal, err := resolveExistingPath(rootDir)
	if err != nil {
		return nil, err
	}
	artifactReal, err := resolveExistingPath(artifactRoot)
	if err != nil {
		return nil, err
	}
	if err := ensurePathWithinBoundary(rootReal, artifactReal, artifactRoot); err != nil {
		return nil, err
	}
	artifactInfo, err := os.Lstat(artifactRoot)
	if err != nil {
		return nil, err
	}
	if artifactInfo.Mode()&os.ModeSymlink != 0 || !artifactInfo.IsDir() {
		return nil, fmt.Errorf("video_frames artifact root is not a regular directory")
	}
	path, err := os.MkdirTemp(artifactRoot, artifactStem(sourcePath, strategy, intervalSec)+"-")
	if err != nil {
		return nil, err
	}
	created := true
	defer func() {
		if created {
			_ = os.RemoveAll(path)
		}
	}()
	if err := os.Chmod(path, 0o700); err != nil {
		return nil, err
	}
	identity, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if identity.Mode()&os.ModeSymlink != 0 || !identity.IsDir() {
		return nil, fmt.Errorf("video_frames artifact directory is not owned")
	}
	pathReal, err := resolveExistingPath(path)
	if err != nil {
		return nil, err
	}
	if err := ensurePathWithinBoundary(artifactReal, pathReal, path); err != nil {
		return nil, err
	}
	display, err := filepath.Rel(rootDir, path)
	if err != nil {
		return nil, err
	}
	created = false
	return &ownedOutput{path: path, display: filepath.ToSlash(display), identity: identity}, nil
}

func (o *ownedOutput) retain() {
	if o != nil {
		o.retained = true
	}
}

func (o *ownedOutput) cleanup() {
	if o == nil || o.retained || strings.TrimSpace(o.path) == "" || o.identity == nil {
		return
	}
	current, err := os.Lstat(o.path)
	if err != nil || !current.IsDir() || !os.SameFile(o.identity, current) {
		return
	}
	_ = os.RemoveAll(o.path)
}

func snapshotInput(ctx context.Context, sourcePath, outputDir string, maxBytes int64) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if maxBytes <= 0 {
		return "", fmt.Errorf("video_frames max input bytes must be positive")
	}
	snapshotPath := filepath.Join(outputDir, "source"+inputExtension(sourcePath))
	target, err := os.OpenFile(snapshotPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", err
	}
	targetOpen, committed := true, false
	defer func() {
		if targetOpen {
			_ = target.Close()
		}
		if !committed {
			_ = os.Remove(snapshotPath)
		}
	}()
	err = withRegularFile(ctx, sourcePath, maxBytes, func(reader io.Reader) error {
		written, copyErr := copyInput(ctx, target, reader, maxBytes)
		if copyErr != nil {
			return copyErr
		}
		if written > maxBytes {
			return fmt.Errorf("%w: limit=%d observed_at_least=%d", ErrFileTooLarge, maxBytes, written)
		}
		return ctx.Err()
	})
	if err != nil {
		return "", err
	}
	if err := target.Sync(); err != nil {
		return "", err
	}
	if err := target.Close(); err != nil {
		return "", err
	}
	targetOpen = false
	info, err := os.Lstat(snapshotPath)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Size() > maxBytes {
		return "", ErrUnsafeFile
	}
	committed = true
	return snapshotPath, nil
}

func removeInputSnapshot(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return ErrUnsafeFile
	}
	return os.Remove(path)
}

func withRegularFile(ctx context.Context, path string, maxBytes int64, consume func(io.Reader) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if maxBytes <= 0 {
		return fmt.Errorf("local input max file bytes must be positive")
	}
	if consume == nil {
		return fmt.Errorf("local input consumer is required")
	}
	before, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if err := validateRegularFile(before); err != nil {
		return err
	}
	if before.Size() < 0 {
		return ErrUnsafeFile
	}
	if before.Size() > maxBytes {
		return fmt.Errorf("%w: limit=%d observed_at_least=%d", ErrFileTooLarge, maxBytes, before.Size())
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return err
	}
	if err := validateRegularFile(opened); err != nil {
		return err
	}
	if !os.SameFile(before, opened) {
		return fmt.Errorf("%w: local input changed before open", ErrUnsafeFile)
	}
	if opened.Size() < 0 {
		return ErrUnsafeFile
	}
	if opened.Size() > maxBytes {
		return fmt.Errorf("%w: limit=%d observed_at_least=%d", ErrFileTooLarge, maxBytes, opened.Size())
	}
	consumeErr := consume(contextReader{ctx: ctx, reader: file})
	after, err := file.Stat()
	if err != nil {
		return err
	}
	if err := validateRegularFile(after); err != nil {
		return err
	}
	if !os.SameFile(opened, after) {
		return fmt.Errorf("%w: opened local input identity changed", ErrUnsafeFile)
	}
	if after.Size() < 0 {
		return ErrUnsafeFile
	}
	if after.Size() > maxBytes {
		return fmt.Errorf("%w: limit=%d observed_at_least=%d", ErrFileTooLarge, maxBytes, after.Size())
	}
	pathAfter, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if err := validateRegularFile(pathAfter); err != nil {
		return err
	}
	if !os.SameFile(opened, pathAfter) {
		return fmt.Errorf("%w: local input path changed during read", ErrUnsafeFile)
	}
	return consumeErr
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(buffer []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(buffer)
}

func validateRegularFile(info os.FileInfo) error {
	if info == nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return ErrUnsafeFile
	}
	return nil
}

func copyInput(ctx context.Context, destination io.Writer, source io.Reader, maxBytes int64) (int64, error) {
	limit := maxBytes
	if limit < math.MaxInt64 {
		limit++
	}
	limited := io.LimitReader(source, limit)
	buffer := make([]byte, 64<<10)
	var written int64
	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		readBytes, readErr := limited.Read(buffer)
		if readBytes > 0 {
			if err := ctx.Err(); err != nil {
				return written, err
			}
			writeBytes, writeErr := destination.Write(buffer[:readBytes])
			written += int64(writeBytes)
			if writeErr != nil {
				return written, writeErr
			}
			if writeBytes != readBytes {
				return written, io.ErrShortWrite
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				return written, nil
			}
			return written, readErr
		}
		if readBytes == 0 {
			return written, io.ErrNoProgress
		}
	}
}

func inputExtension(path string) string {
	extension := strings.ToLower(filepath.Ext(strings.TrimSpace(path)))
	if len(extension) < 2 || len(extension) > 12 {
		return ".video"
	}
	for _, char := range extension[1:] {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') {
			return ".video"
		}
	}
	return extension
}
