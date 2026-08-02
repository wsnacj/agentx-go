package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashPathNormalizesInput(t *testing.T) {
	hashA := HashPath("/tmp/demo/../file.txt", "suffix")
	hashB := HashPath("/tmp/file.txt", "suffix")
	hashC := HashPath("/tmp/file.txt", "other")

	if hashA != hashB {
		t.Fatalf("expected cleaned paths to hash equally: %s vs %s", hashA, hashB)
	}
	if hashA == hashC {
		t.Fatalf("expected different suffixes to change hash")
	}
}

func TestHashFileContentsDependsOnBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	if err := os.WriteFile(path, []byte("alpha"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	hashA, err := HashFileContents(path)
	if err != nil {
		t.Fatalf("hash file: %v", err)
	}

	if err := os.WriteFile(path, []byte("beta"), 0o644); err != nil {
		t.Fatalf("rewrite file: %v", err)
	}
	hashB, err := HashFileContents(path)
	if err != nil {
		t.Fatalf("hash rewritten file: %v", err)
	}

	if hashA == hashB {
		t.Fatalf("expected hash to change when content changes")
	}
}

func TestHashFileContentsReturnsErrorForMissingFile(t *testing.T) {
	if _, err := HashFileContents("/path/that/does/not/exist"); err == nil {
		t.Fatal("expected missing file error")
	}
}
