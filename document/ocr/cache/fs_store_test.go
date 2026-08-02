package cache

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wsnacj/agentx-go/document/ocr/config"
)

func TestFSStoreRespectsMaxSize(t *testing.T) {
	dir := t.TempDir()
	cfg := config.CacheConfig{
		Enabled:   true,
		Kind:      "fs",
		BaseDir:   dir,
		MaxSizeMB: 1,
	}
	store, err := newFSStore(cfg)
	if err != nil {
		t.Fatalf("new fs store: %v", err)
	}

	ctx := context.Background()
	entryOne := Entry{Data: bytes.Repeat([]byte("a"), 600*1024)}
	entryTwo := Entry{Data: bytes.Repeat([]byte("b"), 600*1024)}

	if err := store.Set(ctx, "k1", entryOne); err != nil {
		t.Fatalf("set first entry: %v", err)
	}
	// Ensure mod time differs for eviction ordering.
	time.Sleep(10 * time.Millisecond)
	if err := store.Set(ctx, "k2", entryTwo); err != nil {
		t.Fatalf("set second entry: %v", err)
	}

	if _, ok, err := store.Get(ctx, "k1"); err != nil {
		t.Fatalf("get first entry: %v", err)
	} else if ok {
		t.Fatalf("expected first entry to be evicted")
	}

	if _, ok, err := store.Get(ctx, "k2"); err != nil {
		t.Fatalf("get second entry: %v", err)
	} else if !ok {
		t.Fatalf("expected second entry to be present")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	var total int64
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			t.Fatalf("entry info: %v", err)
		}
		total += info.Size()
	}
	limit := int64(cfg.MaxSizeMB) * 1024 * 1024
	if total > limit {
		t.Fatalf("cache size exceeds limit: %d > %d", total, limit)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		t.Fatalf("glob cache files: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected only one cached file after eviction, got %d", len(files))
	}
}
