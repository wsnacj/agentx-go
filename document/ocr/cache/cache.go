package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/wsnacj/agentx-go/document/ocr/config"
)

// Entry encapsulates cached payload.
type Entry struct {
	Data       []byte
	StoredAt   time.Time
	Attributes map[string]any
}

// Store defines minimal cache functionality.
type Store interface {
	Get(ctx context.Context, key string) (Entry, bool, error)
	Set(ctx context.Context, key string, entry Entry) error
}

// Builder creates a cache Store from configuration.
type Builder func(cfg config.CacheConfig) (Store, error)

// DefaultBuilder returns a builder that constructs a noop cache.
func DefaultBuilder() Builder {
	return func(cfg config.CacheConfig) (Store, error) {
		if !cfg.Enabled {
			return noopStore{}, nil
		}
		switch cfg.Kind {
		case "", "fs":
			return newFSStore(cfg)
		default:
			return nil, fmt.Errorf("cache kind %s not supported", cfg.Kind)
		}
	}
}

type noopStore struct{}

func (noopStore) Get(ctx context.Context, key string) (Entry, bool, error) {
	return Entry{}, false, nil
}
func (noopStore) Set(ctx context.Context, key string, entry Entry) error { return nil }
