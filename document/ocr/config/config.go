package config

import (
	"fmt"
	"time"
)

// ServiceConfig is the top-level configuration structure for the OCRX service.
type ServiceConfig struct {
	Pipelines map[string]PipelineConfig `json:"pipelines" yaml:"pipelines"`
}

// Validate ensures pipline definitions are consistent.
func (c ServiceConfig) Validate() error {
	if len(c.Pipelines) == 0 {
		return fmt.Errorf("no pipelines configured")
	}
	for name, p := range c.Pipelines {
		if err := p.Validate(); err != nil {
			return fmt.Errorf("pipeline %s: %w", name, err)
		}
	}
	return nil
}

// PipelineConfig bundles all knobs for a single OCR pipeline.
type PipelineConfig struct {
	Provider ProviderConfig `json:"provider" yaml:"provider"`
	Splitter SplitterConfig `json:"splitter" yaml:"splitter"`
	Cache    CacheConfig    `json:"cache" yaml:"cache"`
	Worker   WorkerConfig   `json:"worker" yaml:"worker"`
	Limits   LimitConfig    `json:"limits" yaml:"limits"`
	Retry    RetryConfig    `json:"retry" yaml:"retry"`
	Diff     DiffConfig     `json:"diff" yaml:"diff"`
}

// Validate ensures all nested configuration segments are valid.
func (c PipelineConfig) Validate() error {
	if err := c.Provider.Validate(); err != nil {
		return fmt.Errorf("provider: %w", err)
	}
	if err := c.Splitter.Validate(); err != nil {
		return fmt.Errorf("splitter: %w", err)
	}
	if err := c.Cache.Validate(); err != nil {
		return fmt.Errorf("cache: %w", err)
	}
	if err := c.Worker.Validate(); err != nil {
		return fmt.Errorf("worker: %w", err)
	}
	if err := c.Limits.Validate(); err != nil {
		return fmt.Errorf("limits: %w", err)
	}
	if err := c.Retry.Validate(); err != nil {
		return fmt.Errorf("retry: %w", err)
	}
	if err := c.Diff.Validate(); err != nil {
		return fmt.Errorf("diff: %w", err)
	}
	return nil
}

// ProviderConfig contains fields specific to the remote service.
type ProviderConfig struct {
	Kind       string            `json:"kind" yaml:"kind"`
	BaseURL    string            `json:"base_url" yaml:"base_url"`
	Auth       map[string]string `json:"auth" yaml:"auth"`
	Timeout    time.Duration     `json:"timeout" yaml:"timeout"`
	Headers    map[string]string `json:"headers" yaml:"headers"`
	Additional map[string]any    `json:"additional" yaml:"additional"`
}

// Validate ensures provider configuration is consistent.
func (c ProviderConfig) Validate() error {
	if c.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if c.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}
	return nil
}

// SplitterConfig controls PDF splitting behaviour.
type SplitterConfig struct {
	Kind        string         `json:"kind" yaml:"kind"`
	DPI         int            `json:"dpi" yaml:"dpi"`
	BatchPages  int            `json:"batch_pages" yaml:"batch_pages"`
	Options     map[string]any `json:"options" yaml:"options"`
	MaxParallel int            `json:"max_parallel" yaml:"max_parallel"`
}

// Validate ensures splitter configuration is valid.
func (c SplitterConfig) Validate() error {
	if c.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if c.DPI < 0 {
		return fmt.Errorf("dpi cannot be negative")
	}
	if c.BatchPages < 0 {
		return fmt.Errorf("batch_pages cannot be negative")
	}
	if c.MaxParallel < 0 {
		return fmt.Errorf("max_parallel cannot be negative")
	}
	return nil
}

// CacheConfig defines the behaviour of response caching.
type CacheConfig struct {
	Kind       string         `json:"kind" yaml:"kind"`
	Enabled    bool           `json:"enabled" yaml:"enabled"`
	TTL        time.Duration  `json:"ttl" yaml:"ttl"`
	BaseDir    string         `json:"base_dir" yaml:"base_dir"`
	MaxSizeMB  int            `json:"max_size_mb" yaml:"max_size_mb"`
	Additional map[string]any `json:"additional" yaml:"additional"`
}

// Validate ensures cache configuration is consistent.
func (c CacheConfig) Validate() error {
	if c.Enabled && c.TTL < 0 {
		return fmt.Errorf("ttl cannot be negative")
	}
	if c.MaxSizeMB < 0 {
		return fmt.Errorf("max_size_mb cannot be negative")
	}
	return nil
}

// WorkerConfig sets worker-pool characteristics.
type WorkerConfig struct {
	MaxConcurrent int `json:"max_concurrent" yaml:"max_concurrent"`
	QueueSize     int `json:"queue_size" yaml:"queue_size"`
}

// Validate ensures worker configuration values make sense.
func (c WorkerConfig) Validate() error {
	if c.MaxConcurrent < 0 {
		return fmt.Errorf("max_concurrent cannot be negative")
	}
	if c.QueueSize < 0 {
		return fmt.Errorf("queue_size cannot be negative")
	}
	return nil
}

// LimitConfig handles rate and budget limits.
type LimitConfig struct {
	MaxPages       int `json:"max_pages" yaml:"max_pages"`
	MaxFiles       int `json:"max_files" yaml:"max_files"`
	MaxPerFilePage int `json:"max_per_file_page" yaml:"max_per_file_page"`
}

// Validate ensures limit values are consistent.
func (c LimitConfig) Validate() error {
	if c.MaxPages < 0 {
		return fmt.Errorf("max_pages cannot be negative")
	}
	if c.MaxFiles < 0 {
		return fmt.Errorf("max_files cannot be negative")
	}
	if c.MaxPerFilePage < 0 {
		return fmt.Errorf("max_per_file_page cannot be negative")
	}
	return nil
}

// RetryConfig defines retry policies for provider calls.
type RetryConfig struct {
	MaxAttempts int           `json:"max_attempts" yaml:"max_attempts"`
	Backoff     time.Duration `json:"backoff" yaml:"backoff"`
	MaxBackoff  time.Duration `json:"max_backoff" yaml:"max_backoff"`
}

// Validate ensures retry durations are consistent.
func (c RetryConfig) Validate() error {
	if c.MaxAttempts < 0 {
		return fmt.Errorf("max_attempts cannot be negative")
	}
	if c.Backoff < 0 || c.MaxBackoff < 0 {
		return fmt.Errorf("backoff values cannot be negative")
	}
	if c.MaxBackoff > 0 && c.Backoff > c.MaxBackoff {
		return fmt.Errorf("backoff cannot exceed max_backoff")
	}
	return nil
}

// DiffConfig 控制 diff/fuzzy 相关开关。
type DiffConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	Baseline string `json:"baseline" yaml:"baseline"`
	Preview  int    `json:"preview" yaml:"preview"`
}

func (c DiffConfig) Validate() error {
	if c.Preview < 0 {
		return fmt.Errorf("preview cannot be negative")
	}
	return nil
}
