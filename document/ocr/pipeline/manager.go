package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wsnacj/agentx-go/document/internal/logging"
	"github.com/wsnacj/agentx-go/document/ocr/cache"
	"github.com/wsnacj/agentx-go/document/ocr/config"
	"github.com/wsnacj/agentx-go/document/ocr/internal/metrics"
	"github.com/wsnacj/agentx-go/document/ocr/model"
	"github.com/wsnacj/agentx-go/document/ocr/processor"
	"github.com/wsnacj/agentx-go/document/ocr/provider"
	"github.com/wsnacj/agentx-go/document/ocr/splitter"
	"github.com/wsnacj/agentx-go/document/ocr/util"
	"github.com/wsnacj/agentx-go/document/ocr/worker"

	"github.com/cenkalti/backoff/v5"
	"go.uber.org/zap"
)

// Manager coordinates splitting, caching, provider calls and aggregation
// for a specific pipeline kind.
type Manager struct {
	cfg          config.PipelineConfig
	provider     provider.Provider
	splitter     splitter.Splitter
	cache        cache.Store
	pool         *worker.Pool
	kind         model.OperationKind
	diffEnabled  bool
	diffPreview  int
	diffBaseline []byte
	processor    processor.OperationProcessor
}

type workItem struct {
	SourcePath  string
	PayloadPath string
	IsRemote    bool
	PageIndex   int
	Options     map[string]any
}

// NewManager creates a Manager instance from collaborators.

func NewManager(
	kind model.OperationKind,
	cfg config.PipelineConfig,
	p provider.Provider,
	s splitter.Splitter,
	c cache.Store,
	pool *worker.Pool,
	proc processor.OperationProcessor,
) (*Manager, error) {
	if p == nil {
		return nil, fmt.Errorf("provider cannot be nil")
	}
	if s == nil {
		return nil, fmt.Errorf("splitter cannot be nil")
	}
	if c == nil {
		return nil, fmt.Errorf("cache cannot be nil")
	}
	if pool == nil {
		return nil, fmt.Errorf("worker pool cannot be nil")
	}
	if proc == nil {
		return nil, fmt.Errorf("operation processor cannot be nil")
	}
	var baseline []byte
	if cfg.Diff.Enabled && cfg.Diff.Baseline != "" {
		data, err := os.ReadFile(cfg.Diff.Baseline)
		if err != nil {
			return nil, fmt.Errorf("read diff baseline %s: %w", cfg.Diff.Baseline, err)
		}
		baseline = data
	}

	preview := cfg.Diff.Preview
	if preview <= 0 {
		preview = 3
	}

	return &Manager{
		cfg:          cfg,
		provider:     p,
		splitter:     s,
		cache:        c,
		pool:         pool,
		kind:         kind,
		diffEnabled:  cfg.Diff.Enabled,
		diffPreview:  preview,
		diffBaseline: baseline,
		processor:    proc,
	}, nil
}

// Run executes the pipeline synchronously.
func (m *Manager) Run(ctx context.Context, req model.Request) (model.Response, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	start := time.Now()
	meta := model.Meta{
		ProcessedPaths: nil,
		Diagnostics:    map[string]any{},
	}

	logger.InfoWithFields("ocrx pipeline start",
		zap.String("operation", string(m.kind)),
		zap.Int("input_files", len(req.Paths)),
	)

	jobs, cleanups, err := m.buildJobs(ctx, req)
	if err != nil {
		return model.Response{}, err
	}
	defer m.runCleanups(cleanups)

	if len(jobs) == 0 {
		return model.Response{}, fmt.Errorf("no work items prepared")
	}

	needCharacter := optionBool(req.Options, "need_character")
	diffCfg := m.resolveDiffConfig(req.Options)
	var fingerprints *sourceFingerprintCache
	if m.cfg.Cache.Enabled {
		fingerprints = newSourceFingerprintCache()
	}

	rawData := make([][]byte, len(jobs))
	files := make([]string, len(jobs))
	var cacheHits, cacheMisses, retryAttempts int64

	var wg sync.WaitGroup
	var (
		errOnce  sync.Once
		errMu    sync.Mutex
		firstErr error
	)
	recordErr := func(err error) {
		if err == nil {
			return
		}
		errOnce.Do(func() {
			errMu.Lock()
			firstErr = err
			errMu.Unlock()
			cancel()
		})
	}
	getFirstErr := func() error {
		errMu.Lock()
		defer errMu.Unlock()
		return firstErr
	}

	for idx, job := range jobs {
		if ctx.Err() != nil {
			break
		}
		if err := m.pool.Acquire(ctx); err != nil {
			recordErr(fmt.Errorf("acquire worker: %w", err))
			break
		}
		wg.Add(1)
		go func(i int, jb workItem) {
			defer wg.Done()
			defer m.pool.Release()

			jobStart := time.Now()
			data, cached, attempts, errCategory, callErr := m.executeJob(ctx, jb, needCharacter, fingerprints)
			if callErr != nil {
				retryCount := 0
				if attempts > 0 {
					retryCount = attempts - 1
				}
				metrics.ObserveRequest(string(m.kind), cached, false, retryCount, time.Since(jobStart), errCategory)
				recordErr(callErr)
				return
			}
			retryCount := 0
			if attempts > 0 {
				retryCount = attempts - 1
			}
			metrics.ObserveRequest(string(m.kind), cached, true, retryCount, time.Since(jobStart), "none")
			rawData[i] = data
			files[i] = jb.PayloadPath
			if cached {
				atomic.AddInt64(&cacheHits, 1)
			} else {
				atomic.AddInt64(&cacheMisses, 1)
			}
			if attempts > 1 {
				atomic.AddInt64(&retryAttempts, int64(attempts-1))
			}
		}(idx, job)
	}

	wg.Wait()

	if firstErr := getFirstErr(); firstErr != nil {
		logger.ErrorWithFields("ocrx pipeline failed",
			zap.String("operation", string(m.kind)),
			zap.Error(firstErr),
		)
		return model.Response{}, firstErr
	}

	duration := time.Since(start)
	meta.ProcessedPaths = filterEmpty(files)
	meta.Duration = duration
	meta.Diagnostics["cache_hits"] = cacheHits
	meta.Diagnostics["cache_misses"] = cacheMisses
	meta.Diagnostics["jobs"] = len(jobs)
	meta.Diagnostics["provider_retries"] = retryAttempts

	logger.InfoWithFields("ocrx pipeline completed",
		zap.String("operation", string(m.kind)),
		zap.Duration("duration", duration),
		zap.Int("jobs", len(jobs)),
		zap.Int64("cache_hits", cacheHits),
		zap.Int64("cache_misses", cacheMisses),
	)

	payloadAny, err := m.processor.Build(rawData, files)
	if err != nil {
		return model.Response{}, err
	}

	var diffSummary *model.DiffSummary
	if diffCfg.enabled {
		diffSummary, err = m.processor.Diff(rawData, diffCfg.baseline, diffCfg.preview)
		if err != nil {
			logger.WarnWithFields("ocrx diff execution failed",
				zap.String("operation", string(m.kind)),
				zap.Error(err),
			)
			diffSummary = nil
		}
	}

	switch m.kind {
	case model.OperationKindOCR:
		payload, ok := payloadAny.(model.OCRPayload)
		if !ok {
			return model.Response{}, fmt.Errorf("unexpected OCR payload type %T", payloadAny)
		}
		return model.Response{Meta: meta, Payload: payload, Diff: diffSummary}, nil
	case model.OperationKindTable:
		payload, ok := payloadAny.(model.TablePayload)
		if !ok {
			return model.Response{}, fmt.Errorf("unexpected table payload type %T", payloadAny)
		}
		return model.Response{Meta: meta, Payload: payload, Diff: diffSummary}, nil
	case model.OperationKindStamp:
		payload, ok := payloadAny.(model.StampPayload)
		if !ok {
			return model.Response{}, fmt.Errorf("unexpected stamp payload type %T", payloadAny)
		}
		return model.Response{Meta: meta, Payload: payload}, nil
	default:
		return model.Response{}, fmt.Errorf("unsupported operation %s", m.kind)
	}
}

func (m *Manager) buildJobs(ctx context.Context, req model.Request) ([]workItem, []func() error, error) {
	if len(req.Paths) == 0 {
		return nil, nil, fmt.Errorf("request paths empty")
	}

	maxPages := req.MaxPages
	if m.cfg.Limits.MaxPages > 0 {
		if maxPages == 0 || m.cfg.Limits.MaxPages < maxPages {
			maxPages = m.cfg.Limits.MaxPages
		}
	}

	var jobs []workItem
	var cleanups []func() error
	pagesUsed := 0
	filesSeen := 0
	maxFiles := m.cfg.Limits.MaxFiles
	maxPerFilePage := m.cfg.Limits.MaxPerFilePage

	for _, src := range req.Paths {
		if maxFiles > 0 && filesSeen >= maxFiles {
			logger.WarnWithFields("max files limit reached",
				zap.Int("max_files", maxFiles),
				zap.String("operation", string(m.kind)),
			)
			break
		}
		if maxPages > 0 && pagesUsed >= maxPages {
			break
		}

		perFilePages := 0

		if isRemotePath(src) {
			perFilePages++
			if maxPerFilePage > 0 && perFilePages > maxPerFilePage {
				logger.WarnWithFields("max per-file pages limit reached",
					zap.Int("limit", maxPerFilePage),
					zap.String("file", src),
					zap.String("operation", string(m.kind)))
				filesSeen++
				continue
			}
			jobs = append(jobs, workItem{
				SourcePath:  src,
				PayloadPath: src,
				IsRemote:    true,
				PageIndex:   1,
				Options:     req.Options,
			})
			pagesUsed++
			filesSeen++
			continue
		}

		ext := strings.ToLower(filepath.Ext(src))
		if ext == ".pdf" {
			remaining := 0
			if maxPages > 0 {
				remaining = maxPages - pagesUsed
			}
			splitRes, err := m.splitter.Split(ctx, splitter.Request{Path: src, MaxPages: remaining, Options: req.Options})
			if err != nil {
				return nil, cleanups, fmt.Errorf("split %s: %w", src, err)
			}
			if splitRes.Cleanup != nil {
				cleanups = append(cleanups, splitRes.Cleanup)
			}
			for _, img := range splitRes.Images {
				if maxPages > 0 && pagesUsed >= maxPages {
					break
				}
				if maxPerFilePage > 0 && perFilePages >= maxPerFilePage {
					logger.WarnWithFields("max per-file pages limit reached",
						zap.Int("limit", maxPerFilePage),
						zap.String("file", src),
						zap.String("operation", string(m.kind)))
					break
				}
				perFilePages++
				jobs = append(jobs, workItem{
					SourcePath:  src,
					PayloadPath: img,
					IsRemote:    isRemotePath(img),
					PageIndex:   perFilePages,
					Options:     req.Options,
				})
				pagesUsed++
			}
			if len(splitRes.Images) == 0 {
				logger.WarnWithFields("splitter produced no images",
					zap.String("file", src),
					zap.String("operation", string(m.kind)),
				)
			}
			filesSeen++
			continue
		}

		perFilePages++
		if maxPerFilePage > 0 && perFilePages > maxPerFilePage {
			logger.WarnWithFields("max per-file pages limit reached",
				zap.Int("limit", maxPerFilePage),
				zap.String("file", src),
				zap.String("operation", string(m.kind)))
			filesSeen++
			continue
		}
		jobs = append(jobs, workItem{
			SourcePath:  src,
			PayloadPath: src,
			IsRemote:    false,
			PageIndex:   1,
			Options:     req.Options,
		})
		pagesUsed++
		filesSeen++
	}

	return jobs, cleanups, nil
}

func (m *Manager) runCleanups(cleanups []func() error) {
	for _, fn := range cleanups {
		if fn == nil {
			continue
		}
		if err := fn(); err != nil {
			logger.WarnWithFields("ocrx cleanup failed", zap.Error(err))
		}
	}
}

func (m *Manager) executeJob(ctx context.Context, job workItem, needCharacter bool, fingerprints *sourceFingerprintCache) ([]byte, bool, int, string, error) {
	providerOpts := sanitizeProviderOptions(job.Options)
	key, keyErr := m.cacheKey(job, needCharacter, providerOpts, fingerprints)
	if keyErr != nil {
		logger.WarnWithFields("ocrx cache key fingerprint fallback",
			zap.String("path", job.SourcePath),
			zap.Error(keyErr),
		)
	}
	entry, ok, err := m.cache.Get(ctx, key)
	if err != nil {
		logger.WarnWithFields("ocrx cache read failed",
			zap.String("path", job.PayloadPath),
			zap.Error(err),
		)
	}
	if ok {
		logger.DebugWithFields("ocrx cache hit",
			zap.String("path", job.PayloadPath),
			zap.String("operation", string(m.kind)),
		)
		return entry.Data, true, 1, "cache", nil
	}

	retryCfg := m.cfg.Retry
	maxAttempts := retryCfg.MaxAttempts
	var resp provider.Response
	attempts := 0

	var bo backoff.BackOff
	if maxAttempts == 0 || maxAttempts > 1 {
		expo := backoff.NewExponentialBackOff()
		if retryCfg.Backoff > 0 {
			expo.InitialInterval = retryCfg.Backoff
		}
		if retryCfg.MaxBackoff > 0 {
			expo.MaxInterval = retryCfg.MaxBackoff
		}
		expo.Reset()
		bo = expo
	}

	for {
		attempts++
		r, callErr := m.provider.Call(ctx, provider.Request{
			FilePath:      job.PayloadPath,
			IsRemote:      job.IsRemote,
			NeedCharacter: needCharacter,
			Options:       providerOpts,
		})
		if callErr == nil {
			resp = r
			break
		}

		if ctx.Err() != nil {
			return nil, false, attempts, "context_canceled", fmt.Errorf("provider call %s: %w", job.PayloadPath, ctx.Err())
		}

		if maxAttempts > 0 && attempts >= maxAttempts {
			return nil, false, attempts, categorizeError(callErr), fmt.Errorf("provider call %s: %w", job.PayloadPath, callErr)
		}

		if bo == nil {
			return nil, false, attempts, categorizeError(callErr), fmt.Errorf("provider call %s: %w", job.PayloadPath, callErr)
		}

		next := bo.NextBackOff()
		if next == backoff.Stop {
			return nil, false, attempts, categorizeError(callErr), fmt.Errorf("provider call %s: %w", job.PayloadPath, callErr)
		}

		logger.WarnWithFields("provider retry scheduled",
			zap.String("operation", string(m.kind)),
			zap.String("path", job.PayloadPath),
			zap.Error(callErr),
			zap.Int("attempt", attempts),
			zap.Duration("next_in", next),
		)

		timer := time.NewTimer(next)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return nil, false, attempts, "context_canceled", fmt.Errorf("provider call %s: %w", job.PayloadPath, ctx.Err())
		}
	}

	cacheErr := m.cache.Set(ctx, key, cache.Entry{
		Data:     resp.Raw,
		StoredAt: time.Now(),
		Attributes: map[string]any{
			"source_path": job.SourcePath,
			"page_index":  job.PageIndex,
		},
	})
	if cacheErr != nil {
		logger.WarnWithFields("ocrx cache write failed",
			zap.String("path", job.PayloadPath),
			zap.Error(cacheErr),
		)
	}
	return resp.Raw, false, attempts, "none", nil
}

func (m *Manager) cacheKey(job workItem, needCharacter bool, providerOpts map[string]any, fingerprints *sourceFingerprintCache) (string, error) {
	pageIndex := job.PageIndex
	if pageIndex <= 0 {
		pageIndex = 1
	}
	base := fmt.Sprintf("%s|page=%d", job.SourcePath, pageIndex)
	var fingerprintErr error
	if !isRemotePath(job.SourcePath) {
		base, fingerprintErr = withLocalSourceIdentity(base, job.SourcePath, fingerprints)
	}
	if job.IsRemote {
		base = fmt.Sprintf("%s|payload=%s", base, job.PayloadPath)
	}
	if optsHash := hashOptions(providerOpts); optsHash != "" {
		base = fmt.Sprintf("%s|opts=%s", base, optsHash)
	}
	cacheScope := fmt.Sprintf(
		"char=%t|operation=%s|provider=%s|base_url=%s",
		needCharacter,
		m.kind,
		m.cfg.Provider.Kind,
		m.cfg.Provider.BaseURL,
	)
	return util.HashPath(base, cacheScope), fingerprintErr
}

func withLocalSourceIdentity(base, sourcePath string, fingerprints *sourceFingerprintCache) (string, error) {
	if fingerprints != nil {
		fingerprint, err := fingerprints.Fingerprint(sourcePath)
		if err == nil && fingerprint != "" {
			return fmt.Sprintf("%s|sha256=%s", base, fingerprint), nil
		}
		base = appendSourceMetadata(base, sourcePath)
		return base, err
	}
	return appendSourceMetadata(base, sourcePath), nil
}

func appendSourceMetadata(base, sourcePath string) string {
	if info, err := os.Stat(sourcePath); err == nil {
		return fmt.Sprintf("%s|mtime=%d|size=%d", base, info.ModTime().UnixNano(), info.Size())
	}
	return base
}

type sourceFingerprintCache struct {
	entries sync.Map
}

type sourceFingerprintResult struct {
	once  sync.Once
	value string
	err   error
}

func newSourceFingerprintCache() *sourceFingerprintCache {
	return &sourceFingerprintCache{}
}

func (c *sourceFingerprintCache) Fingerprint(path string) (string, error) {
	if c == nil {
		return util.HashFileContents(path)
	}

	entryAny, _ := c.entries.LoadOrStore(path, &sourceFingerprintResult{})
	entry := entryAny.(*sourceFingerprintResult)
	entry.once.Do(func() {
		entry.value, entry.err = util.HashFileContents(path)
	})
	return entry.value, entry.err
}

func categorizeError(err error) string {
	if err == nil {
		return "none"
	}
	var categorized provider.ErrorCategorizer
	if errors.As(err, &categorized) {
		if cat := categorized.ErrorCategory(); cat != "" {
			return cat
		}
	}
	var tex *provider.TextInError
	if errors.As(err, &tex) {
		return tex.ErrorCategory()
	}
	return "unknown"
}

type diffConfig struct {
	enabled  bool
	baseline []byte
	preview  int
}

func (m *Manager) resolveDiffConfig(opts map[string]any) diffConfig {
	cfg := diffConfig{
		enabled:  m.diffEnabled,
		baseline: m.diffBaseline,
		preview:  m.diffPreview,
	}
	if val, ok := optionBoolValue(opts, "diff_enabled"); ok {
		cfg.enabled = val
	}
	if preview, ok := optionIntValue(opts, "diff_preview"); ok && preview > 0 {
		cfg.preview = preview
	}
	if path, ok := optionStringValue(opts, "diff_baseline"); ok {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			cfg.baseline = nil
		} else if trimmed == m.cfg.Diff.Baseline {
			cfg.baseline = m.diffBaseline
		} else {
			data, err := os.ReadFile(trimmed)
			if err != nil {
				logger.WarnWithFields("ocrx diff baseline override failed",
					zap.String("path", trimmed),
					zap.String("operation", string(m.kind)),
					zap.Error(err),
				)
				cfg.enabled = false
				cfg.baseline = nil
			} else {
				cfg.baseline = data
			}
		}
	}
	return cfg
}

func optionBool(opts map[string]any, key string) bool {
	val, _ := optionBoolValue(opts, key)
	return val
}

func optionBoolValue(opts map[string]any, key string) (bool, bool) {
	if opts == nil {
		return false, false
	}
	v, ok := opts[key]
	if !ok {
		return false, false
	}
	switch val := v.(type) {
	case bool:
		return val, true
	case *bool:
		if val == nil {
			return false, true
		}
		return *val, true
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(val))
		if err != nil {
			return false, true
		}
		return parsed, true
	case int:
		return val != 0, true
	case int64:
		return val != 0, true
	case float64:
		return val != 0, true
	default:
		return false, true
	}
}

func optionIntValue(opts map[string]any, key string) (int, bool) {
	if opts == nil {
		return 0, false
	}
	v, ok := opts[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case int:
		return val, true
	case int32:
		return int(val), true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func optionStringValue(opts map[string]any, key string) (string, bool) {
	if opts == nil {
		return "", false
	}
	v, ok := opts[key]
	if !ok {
		return "", false
	}
	switch val := v.(type) {
	case string:
		return val, true
	case fmt.Stringer:
		return val.String(), true
	default:
		return fmt.Sprintf("%v", val), true
	}
}

func sanitizeProviderOptions(opts map[string]any) map[string]any {
	if len(opts) == 0 {
		return nil
	}
	clean := make(map[string]any, len(opts))
	for k, v := range opts {
		switch k {
		case "diff_enabled", "diff_baseline", "diff_preview":
			continue
		default:
			clean[k] = v
		}
	}
	if len(clean) == 0 {
		return nil
	}
	return clean
}

func hashOptions(opts map[string]any) string {
	if len(opts) == 0 {
		return ""
	}
	data, err := json.Marshal(opts)
	if err != nil {
		return fmt.Sprintf("marshal_err:%T", err)
	}
	return util.HashPath(string(data), "provider_opts")
}

func filterEmpty(in []string) []string {
	var out []string
	for _, v := range in {
		if strings.TrimSpace(v) == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

func isRemotePath(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}
