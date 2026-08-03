package videoframes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveRootDir(root string) (string, error) {
	trimmed := strings.TrimSpace(root)
	if trimmed == "" {
		trimmed = "."
	}
	absRoot, err := filepath.Abs(trimmed)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}
	return filepath.Clean(absRoot), nil
}

func resolvePathWithinRoot(root string, target string) (resolved string, display string, err error) {
	rootDir, err := resolveRootDir(root)
	if err != nil {
		return "", "", err
	}
	rootRealDir, err := resolveExistingPath(rootDir)
	if err != nil {
		return "", "", fmt.Errorf("resolve root symlinks: %w", err)
	}
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return "", "", fmt.Errorf("path is required")
	}
	candidate := trimmed
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(rootDir, candidate)
	}
	candidate, err = filepath.Abs(candidate)
	if err != nil {
		return "", "", fmt.Errorf("resolve path: %w", err)
	}
	candidate = filepath.Clean(candidate)

	rel, err := filepath.Rel(rootDir, candidate)
	if err != nil {
		return "", "", fmt.Errorf("resolve relative path: %w", err)
	}
	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", "", fmt.Errorf("path escapes workspace root: %s", target)
	}
	if rel == "." {
		rel = filepath.Base(candidate)
	}
	candidateRealPath, err := resolvePathForBoundaryCheck(candidate)
	if err != nil {
		return "", "", err
	}
	if err := ensurePathWithinBoundary(rootRealDir, candidateRealPath, target); err != nil {
		return "", "", err
	}
	return candidate, filepath.ToSlash(rel), nil
}

func resolveDirWithinRoot(root string, target string) (resolved string, display string, err error) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		resolved, err := resolveRootDir(root)
		if err != nil {
			return "", "", err
		}
		return resolved, filepath.ToSlash(filepath.Base(resolved)), nil
	}
	return resolvePathWithinRoot(root, trimmed)
}

func resolveExistingPath(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(resolved), nil
}

func resolvePathForBoundaryCheck(path string) (string, error) {
	ancestor := filepath.Clean(path)
	suffix := make([]string, 0, 4)
	for {
		if _, err := os.Lstat(ancestor); err == nil {
			resolvedAncestor, err := filepath.EvalSymlinks(ancestor)
			if err != nil {
				return "", fmt.Errorf("resolve symlinks: %w", err)
			}
			resolved := filepath.Clean(resolvedAncestor)
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return filepath.Clean(resolved), nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("stat path: %w", err)
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return "", fmt.Errorf("resolve symlinks: no existing ancestor for %s", path)
		}
		suffix = append(suffix, filepath.Base(ancestor))
		ancestor = parent
	}
}

func ensurePathWithinBoundary(rootRealDir string, candidateRealPath string, rawTarget string) error {
	rel, err := filepath.Rel(rootRealDir, candidateRealPath)
	if err != nil {
		return fmt.Errorf("resolve real relative path: %w", err)
	}
	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("path escapes workspace root via symlink: %s", rawTarget)
	}
	return nil
}
