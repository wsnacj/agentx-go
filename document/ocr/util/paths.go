package util

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// HashPath returns a sha256-based identifier for the given path combined with suffix.
func HashPath(path string, suffix string) string {
	sum := sha256.Sum256([]byte(filepath.Clean(path) + "|" + suffix))
	return fmt.Sprintf("%x", sum[:])
}

// HashFileContents returns a sha256 digest of the file contents.
func HashFileContents(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
