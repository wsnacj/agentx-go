package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/wsnacj/agentx-go/document/pipeline/types"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func parseCachePolicyCanRead(policy ParseCachePolicy) bool {
	return policy == ParseCachePolicyRead || policy == ParseCachePolicyReadWrite
}

func parseCachePolicyCanWrite(policy ParseCachePolicy) bool {
	return policy == ParseCachePolicyWrite || policy == ParseCachePolicyReadWrite
}

func parseCacheEntryDir(cacheDir string, cacheKey string) string {
	cacheDir = strings.TrimSpace(cacheDir)
	cacheKey = strings.TrimSpace(cacheKey)
	if cacheKey == "" {
		return ""
	}
	return filepath.Join(cacheDir, cacheKey)
}

func loadParseCacheResult(entryDir string, fingerprint *types.ParseFingerprint) (*types.DocumentResult, bool, error) {
	entryDir = strings.TrimSpace(entryDir)
	if entryDir == "" || fingerprint == nil || strings.TrimSpace(fingerprint.CacheKey) == "" {
		return nil, false, nil
	}
	resultPath := filepath.Join(entryDir, "result.json")
	raw, err := os.ReadFile(resultPath)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("read cache result: %w", err)
	}
	var result types.DocumentResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, false, fmt.Errorf("decode cache result: %w", err)
	}
	if result.Fingerprint == nil || result.Fingerprint.CacheKey != fingerprint.CacheKey {
		return nil, false, fmt.Errorf("cache fingerprint mismatch for %s", entryDir)
	}
	result.OutputDir = entryDir
	return &result, true, nil
}

func saveParseCacheResult(entryDir string, result *types.DocumentResult) error {
	entryDir = strings.TrimSpace(entryDir)
	if entryDir == "" || result == nil || result.Fingerprint == nil || strings.TrimSpace(result.Fingerprint.CacheKey) == "" {
		return nil
	}
	if err := os.MkdirAll(entryDir, 0o755); err != nil {
		return fmt.Errorf("create cache entry: %w", err)
	}
	cached := *result
	cached.OutputDir = entryDir
	cached.Cache = nil
	if err := writeJSON(filepath.Join(entryDir, "result.json"), &cached); err != nil {
		return fmt.Errorf("write cache result: %w", err)
	}
	manifest := map[string]any{
		"cache_key":   result.Fingerprint.CacheKey,
		"fingerprint": result.Fingerprint,
		"result":      "result.json",
		"saved_at":    time.Now().UTC().Format(time.RFC3339),
	}
	if err := writeJSON(filepath.Join(entryDir, "manifest.json"), manifest); err != nil {
		return fmt.Errorf("write cache manifest: %w", err)
	}
	return nil
}

func parseCacheInfo(policy ParseCachePolicy, hit bool, written bool, entryDir string) *types.ParseCacheInfo {
	if policy == ParseCachePolicyNone || policy == ParseCachePolicyDefault {
		return nil
	}
	return &types.ParseCacheInfo{
		Policy:   string(policy),
		Hit:      hit,
		Written:  written,
		EntryDir: strings.TrimSpace(entryDir),
	}
}
