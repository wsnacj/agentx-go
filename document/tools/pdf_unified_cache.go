package tools

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type pdfUnifiedArtifactCache struct {
	mu         sync.Mutex
	maxEntries int
	order      []string
	entries    map[string]pdfUnifiedDocumentArtifacts
}

func newPDFUnifiedArtifactCache(maxEntries int) *pdfUnifiedArtifactCache {
	if maxEntries <= 0 {
		maxEntries = defaultPDFArtifactCacheEntries
	}
	return &pdfUnifiedArtifactCache{
		maxEntries: maxEntries,
		order:      make([]string, 0, maxEntries),
		entries:    make(map[string]pdfUnifiedDocumentArtifacts, maxEntries),
	}
}

func buildCachedPDFUnifiedDocumentArtifacts(
	ctx context.Context,
	cache *pdfUnifiedArtifactCache,
	input pdfToolResolvedInput,
	params map[string]any,
	runtime pdfBackendRuntime,
	maxPages int,
	maxPageChars int,
	ocrxConfigPath string,
	resolver pdfModelResolverConfig,
) (pdfUnifiedDocumentArtifacts, error) {
	if cache == nil {
		return buildPDFUnifiedDocumentArtifacts(ctx, input, params, runtime, maxPages, maxPageChars, ocrxConfigPath, resolver)
	}
	key, ok := buildPDFUnifiedArtifactCacheKey(input, params, maxPages, maxPageChars, ocrxConfigPath)
	if !ok {
		return buildPDFUnifiedDocumentArtifacts(ctx, input, params, runtime, maxPages, maxPageChars, ocrxConfigPath, resolver)
	}
	if cached, hit := cache.get(key); hit {
		cached.Path = input.Path
		cached.DisplayPath = input.Display
		cached.AnalysisPlan = applyPDFModelResolverToPlan(cached.AnalysisPlan, resolver)
		return cached, nil
	}
	artifacts, err := buildPDFUnifiedDocumentArtifacts(ctx, input, params, runtime, maxPages, maxPageChars, ocrxConfigPath, resolver)
	if err != nil {
		return pdfUnifiedDocumentArtifacts{}, err
	}
	cache.put(key, artifacts)
	return artifacts, nil
}

func buildPDFUnifiedArtifactCacheKey(input pdfToolResolvedInput, params map[string]any, maxPages int, maxPageChars int, ocrxConfigPath string) (string, bool) {
	ocrKey := ""
	ocrKey = resolveOCRXConfig(ocrxConfigPath)
	queryKey := pdfUnifiedSelectionQueryCacheKey(params)
	if trimmed := strings.TrimSpace(input.CacheIdentity); trimmed != "" {
		return fmt.Sprintf("%s|%d|%d|%s|%s|%s", trimmed, maxPages, maxPageChars, normalizePDFUnifiedPageSelectionSpec(params), ocrKey, queryKey), true
	}
	if strings.TrimSpace(input.Path) == "" {
		return "", false
	}
	info, err := os.Stat(input.Path)
	if err != nil || info.IsDir() {
		return "", false
	}
	return fmt.Sprintf("%s|%d|%d|%d|%d|%s|%s|%s", input.Path, info.Size(), info.ModTime().UnixNano(), maxPages, maxPageChars, normalizePDFUnifiedPageSelectionSpec(params), ocrKey, queryKey), true
}

func pdfUnifiedSelectionQueryCacheKey(params map[string]any) string {
	if hasPDFToolPageSelection(params) {
		return "explicit"
	}
	prompt := strings.ToLower(strings.TrimSpace(firstString(params, "prompt", "query", "instruction", "task")))
	if prompt == "" {
		return "query:none"
	}
	if classifyPDFUnifiedQuery(prompt) != pdfUnifiedQueryClassFieldCompare {
		return "query:prefix"
	}
	sum := sha256.Sum256([]byte(prompt))
	return fmt.Sprintf("query:%x", sum[:8])
}

func normalizePDFUnifiedPageSelectionSpec(params map[string]any) string {
	if len(params) == 0 {
		return "all"
	}
	if ints := readIntSlice(params["pages"]); len(ints) > 0 {
		parts := make([]string, 0, len(ints))
		for _, page := range ints {
			parts = append(parts, strconv.Itoa(page))
		}
		return "pages:" + strings.Join(parts, ",")
	}
	if text, ok := params["pages"].(string); ok && strings.TrimSpace(text) != "" {
		return "pages_str:" + strings.TrimSpace(text)
	}
	if raw := strings.TrimSpace(firstString(params, "page_range", "pageRange")); raw != "" {
		return "range:" + raw
	}
	start := firstInt(params, "start_page")
	end := firstInt(params, "end_page")
	if start > 0 || end > 0 {
		return fmt.Sprintf("bounds:%d:%d", start, end)
	}
	return "all"
}

func (c *pdfUnifiedArtifactCache) get(key string) (pdfUnifiedDocumentArtifacts, bool) {
	if c == nil || key == "" {
		return pdfUnifiedDocumentArtifacts{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.entries[key]
	return value, ok
}

func (c *pdfUnifiedArtifactCache) put(key string, artifacts pdfUnifiedDocumentArtifacts) {
	if c == nil || key == "" || c.maxEntries <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.entries[key]; ok {
		c.entries[key] = artifacts
		return
	}
	if len(c.entries) >= c.maxEntries && len(c.order) > 0 {
		evict := c.order[0]
		c.order = append([]string(nil), c.order[1:]...)
		delete(c.entries, evict)
	}
	c.entries[key] = artifacts
	c.order = append(c.order, key)
}
