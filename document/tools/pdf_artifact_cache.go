package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

const defaultPDFArtifactCacheEntries = 64

type pdfArtifactCache struct {
	mu         sync.Mutex
	maxEntries int
	order      []string
	entries    map[string]pdfAnalysisArtifacts
}

func newPDFArtifactCache(maxEntries int) *pdfArtifactCache {
	if maxEntries <= 0 {
		maxEntries = defaultPDFArtifactCacheEntries
	}
	return &pdfArtifactCache{
		maxEntries: maxEntries,
		order:      make([]string, 0, maxEntries),
		entries:    make(map[string]pdfAnalysisArtifacts, maxEntries),
	}
}

func buildCachedPDFAnalysisArtifacts(
	ctx context.Context,
	cache *pdfArtifactCache,
	runtime pdfBackendRuntime,
	resolver pdfModelResolverConfig,
	surface pdfToolSurface,
	input pdfToolResolvedInput,
	query string,
	maxExcerptChars int,
	ocrxConfigPath string,
	includeVisualAnalysis bool,
	forceVisualAnalysis bool,
	visionModel string,
	maxVisualPages int,
) (pdfAnalysisArtifacts, error) {
	if cache == nil {
		artifacts, err := buildPDFAnalysisArtifacts(ctx, runtime, input.Path, input.Display, query, maxExcerptChars, ocrxConfigPath, false, visionModel, maxVisualPages)
		if err != nil {
			return pdfAnalysisArtifacts{}, err
		}
		artifacts.AnalysisPlan = adaptPDFAnalysisPlanForToolSurface(artifacts.AnalysisPlan, resolver, surface, forceVisualAnalysis)
		if includeVisualAnalysis {
			artifacts = enrichPDFAnalysisArtifactsWithVisualAnalysis(ctx, artifacts, query, visionModel, maxVisualPages)
		}
		return artifacts, nil
	}
	key, ok := buildPDFAnalysisArtifactCacheKey(input, maxExcerptChars, ocrxConfigPath)
	if !ok {
		return buildCachedPDFAnalysisArtifacts(ctx, nil, runtime, resolver, surface, input, query, maxExcerptChars, ocrxConfigPath, includeVisualAnalysis, forceVisualAnalysis, visionModel, maxVisualPages)
	}
	if cached, ok := cache.get(key); ok {
		cached.DisplayPath = input.Display
		cached.Path = input.Path
		cached.AnalysisPlan = adaptPDFAnalysisPlanForToolSurface(cached.AnalysisPlan, resolver, surface, forceVisualAnalysis)
		if includeVisualAnalysis {
			cached = enrichPDFAnalysisArtifactsWithVisualAnalysis(ctx, cached, query, visionModel, maxVisualPages)
		}
		return cached, nil
	}
	artifacts, err := buildPDFAnalysisArtifacts(ctx, runtime, input.Path, input.Display, query, maxExcerptChars, ocrxConfigPath, false, visionModel, maxVisualPages)
	if err != nil {
		return pdfAnalysisArtifacts{}, err
	}
	cache.put(key, artifacts)
	artifacts.AnalysisPlan = adaptPDFAnalysisPlanForToolSurface(artifacts.AnalysisPlan, resolver, surface, forceVisualAnalysis)
	if includeVisualAnalysis {
		artifacts = enrichPDFAnalysisArtifactsWithVisualAnalysis(ctx, artifacts, query, visionModel, maxVisualPages)
	}
	return artifacts, nil
}

func buildPDFAnalysisArtifactCacheKey(input pdfToolResolvedInput, maxExcerptChars int, ocrxConfigPath string) (string, bool) {
	ocrKey := ""
	ocrKey = resolveOCRXConfig(ocrxConfigPath)
	if trimmed := strings.TrimSpace(input.CacheIdentity); trimmed != "" {
		return fmt.Sprintf("%s|%d|%s", trimmed, maxExcerptChars, ocrKey), true
	}
	if input.Path == "" {
		return "", false
	}
	info, err := os.Stat(input.Path)
	if err != nil || info.IsDir() {
		return "", false
	}
	return fmt.Sprintf("%s|%d|%d|%d|%s", input.Path, info.Size(), info.ModTime().UnixNano(), maxExcerptChars, ocrKey), true
}

func (c *pdfArtifactCache) get(key string) (pdfAnalysisArtifacts, bool) {
	if c == nil || key == "" {
		return pdfAnalysisArtifacts{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.entries[key]
	return value, ok
}

func (c *pdfArtifactCache) put(key string, artifacts pdfAnalysisArtifacts) {
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
