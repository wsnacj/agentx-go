package cache

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/wsnacj/agentx-go/document/ocr/config"
	"github.com/wsnacj/agentx-go/document/ocr/util"
)

// fsStore is a simple filesystem backed cache implementation.
type fsStore struct {
	baseDir string
	ttl     time.Duration
	maxSize int64
	mu      sync.RWMutex
}

func newFSStore(cfg config.CacheConfig) (Store, error) {
	if cfg.BaseDir == "" {
		cfg.BaseDir = filepath.Join(os.TempDir(), "ocrx-cache")
	}
	if err := os.MkdirAll(cfg.BaseDir, 0o755); err != nil {
		return nil, fmt.Errorf("fs cache: mkdir: %w", err)
	}
	maxBytes := int64(cfg.MaxSizeMB) * 1024 * 1024
	return &fsStore{
		baseDir: cfg.BaseDir,
		ttl:     cfg.TTL,
		maxSize: maxBytes,
	}, nil
}

func (s *fsStore) pathFor(key string) string {
	return filepath.Join(s.baseDir, util.HashPath(key, "cache"))
}

func (s *fsStore) Get(ctx context.Context, key string) (Entry, bool, error) {
	select {
	case <-ctx.Done():
		return Entry{}, false, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	path := s.pathFor(key)
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return Entry{}, false, nil
	}
	if err != nil {
		return Entry{}, false, fmt.Errorf("fs cache: stat: %w", err)
	}

	if s.ttl > 0 && time.Since(info.ModTime()) > s.ttl {
		_ = os.Remove(path)
		return Entry{}, false, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Entry{}, false, fmt.Errorf("fs cache: read: %w", err)
	}

	return Entry{
		Data:     data,
		StoredAt: info.ModTime(),
	}, true, nil
}

func (s *fsStore) Set(ctx context.Context, key string, entry Entry) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	path := s.pathFor(key)
	tmp := path + ".tmp"

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.WriteFile(tmp, entry.Data, 0o644); err != nil {
		return fmt.Errorf("fs cache: write tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("fs cache: rename: %w", err)
	}
	if s.maxSize > 0 {
		s.trimLocked()
	}
	return nil
}

func (s *fsStore) trimLocked() {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return
	}
	type fileStat struct {
		name string
		size int64
		mod  time.Time
	}
	files := make([]fileStat, 0, len(entries))
	var total int64
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		size := info.Size()
		total += size
		files = append(files, fileStat{name: entry.Name(), size: size, mod: info.ModTime()})
	}
	if total <= s.maxSize {
		return
	}
	sort.Slice(files, func(i, j int) bool { return files[i].mod.Before(files[j].mod) })
	for _, f := range files {
		if total <= s.maxSize {
			break
		}
		if err := os.Remove(filepath.Join(s.baseDir, f.name)); err != nil {
			continue
		}
		total -= f.size
	}
}
