package tools

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"strconv"
	"strings"
	"sync"

	types "github.com/wsnacj/agentx-go/components/llm"
)

type pdfUnifiedVisualCache struct {
	mu         sync.Mutex
	maxEntries int
	order      []string
	entries    map[string]pdfUnifiedVisualCacheEntry
}

type pdfUnifiedVisualCacheEntry struct {
	Pages   []int
	Visuals []types.VisualContent
}

func newPDFUnifiedVisualCache(maxEntries int) *pdfUnifiedVisualCache {
	if maxEntries <= 0 {
		maxEntries = defaultPDFArtifactCacheEntries
	}
	return &pdfUnifiedVisualCache{
		maxEntries: maxEntries,
		order:      make([]string, 0, maxEntries),
		entries:    make(map[string]pdfUnifiedVisualCacheEntry, maxEntries),
	}
}

func buildCachedPDFUnifiedPromptVisuals(
	ctx context.Context,
	cache *pdfUnifiedVisualCache,
	documents []pdfUnifiedDocumentArtifacts,
	focuses []pdfUnifiedDocumentFocus,
	queryClass string,
	maxVisualPages int,
) ([]types.VisualContent, []pdfUnifiedDocumentArtifacts, []string, func(), error) {
	if cache == nil {
		return buildPDFUnifiedPromptVisuals(ctx, documents, focuses, queryClass, maxVisualPages)
	}
	if len(documents) == 0 {
		return nil, nil, nil, nil, nil
	}
	out := append([]pdfUnifiedDocumentArtifacts(nil), documents...)
	visuals := make([]types.VisualContent, 0, len(documents)*4)
	warnings := make([]string, 0, len(documents))
	cleanups := make([]func() error, 0, len(documents))
	cleanup := func() {
		for _, fn := range cleanups {
			if fn != nil {
				_ = fn()
			}
		}
	}
	for idx := range out {
		var focus pdfUnifiedDocumentFocus
		if idx < len(focuses) {
			focus = focuses[idx]
		}
		pages := selectPDFUnifiedVisualPages(out[idx], focus, queryClass, maxVisualPages)
		if len(pages) == 0 {
			continue
		}
		if len(pages) > maxVisualPages {
			pages = append([]int(nil), pages[:maxVisualPages]...)
		}
		cacheKey, cacheable := buildPDFUnifiedVisualCacheKey(out[idx], pages, maxVisualPages)
		if cacheable {
			if entry, ok := cache.get(cacheKey); ok {
				visuals = append(visuals, clonePDFVisualContentSlice(entry.Visuals)...)
				out[idx].VisualPages = append([]int(nil), entry.Pages...)
				continue
			}
		}
		rendered, fn, err := pdfRenderPDFPages(ctx, out[idx].Path, pages, defaultPDFVisualDPI)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("render %s visuals failed: %v", out[idx].DisplayPath, err))
			continue
		}
		if fn != nil {
			cleanups = append(cleanups, fn)
		}
		docVisuals := []types.VisualContent{types.NewTextBlock(fmt.Sprintf("Rendered pages for %s", out[idx].DisplayPath))}
		renderedVisuals, err := buildPDFVisualContents(rendered, out[idx].PageMap)
		if err != nil {
			cleanup()
			return nil, nil, nil, nil, err
		}
		docVisuals = append(docVisuals, renderedVisuals...)
		visuals = append(visuals, docVisuals...)
		out[idx].VisualPages = append([]int(nil), pages...)
		if cacheable {
			cache.put(cacheKey, pdfUnifiedVisualCacheEntry{
				Pages:   append([]int(nil), pages...),
				Visuals: clonePDFVisualContentSlice(docVisuals),
			})
		}
	}
	return visuals, out, warnings, cleanup, nil
}

func buildPDFUnifiedVisualCacheKey(document pdfUnifiedDocumentArtifacts, pages []int, maxVisualPages int) (string, bool) {
	if trimmed := strings.TrimSpace(document.CacheIdentity); trimmed != "" {
		h := fnv.New64a()
		for _, page := range document.PageMap {
			_, _ = h.Write([]byte(strconv.Itoa(page.Page)))
			_, _ = h.Write([]byte{0})
			_, _ = h.Write([]byte(page.Excerpt))
			_, _ = h.Write([]byte{0})
		}
		return fmt.Sprintf("%s|%d|%s|%x", trimmed, maxVisualPages, formatPDFPageSelection(pages), h.Sum64()), true
	}
	if strings.TrimSpace(document.Path) == "" {
		return "", false
	}
	info, err := os.Stat(document.Path)
	if err != nil || info.IsDir() {
		return "", false
	}
	h := fnv.New64a()
	for _, page := range document.PageMap {
		_, _ = h.Write([]byte(strconv.Itoa(page.Page)))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(page.Excerpt))
		_, _ = h.Write([]byte{0})
	}
	return fmt.Sprintf("%s|%d|%d|%d|%s|%x", document.Path, info.Size(), info.ModTime().UnixNano(), maxVisualPages, formatPDFPageSelection(pages), h.Sum64()), true
}

func clonePDFVisualContentSlice(in []types.VisualContent) []types.VisualContent {
	if len(in) == 0 {
		return nil
	}
	out := make([]types.VisualContent, 0, len(in))
	for _, item := range in {
		cloned := item
		cloned.Labels = append([]string(nil), item.Labels...)
		if item.FPS != nil {
			value := *item.FPS
			cloned.FPS = &value
		}
		out = append(out, cloned)
	}
	return out
}

func (c *pdfUnifiedVisualCache) get(key string) (pdfUnifiedVisualCacheEntry, bool) {
	if c == nil || key == "" {
		return pdfUnifiedVisualCacheEntry{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.entries[key]
	if !ok {
		return pdfUnifiedVisualCacheEntry{}, false
	}
	return pdfUnifiedVisualCacheEntry{
		Pages:   append([]int(nil), value.Pages...),
		Visuals: clonePDFVisualContentSlice(value.Visuals),
	}, true
}

func (c *pdfUnifiedVisualCache) put(key string, entry pdfUnifiedVisualCacheEntry) {
	if c == nil || key == "" || c.maxEntries <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.entries[key]; ok {
		c.entries[key] = pdfUnifiedVisualCacheEntry{
			Pages:   append([]int(nil), entry.Pages...),
			Visuals: clonePDFVisualContentSlice(entry.Visuals),
		}
		return
	}
	if len(c.entries) >= c.maxEntries && len(c.order) > 0 {
		evict := c.order[0]
		c.order = append([]string(nil), c.order[1:]...)
		delete(c.entries, evict)
	}
	c.entries[key] = pdfUnifiedVisualCacheEntry{
		Pages:   append([]int(nil), entry.Pages...),
		Visuals: clonePDFVisualContentSlice(entry.Visuals),
	}
	c.order = append(c.order, key)
}
