package skills

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	agentxassetfs "github.com/wsnacj/agentx-go/runtime/assetfs"
)

const loadCacheMaxEntries = 64

var (
	loadWatcherRetryBackoff         = 2 * time.Second
	loadFingerprintBuilder          = buildLoadFingerprint
	loadSourceWatcherFactory        = newLoadSourceWatcher
	loadValidationParallelThreshold = 16
	loadValidationMaxWorkers        = 8
)

type loadCacheEntry struct {
	Skills     []Skill
	Report     LoadReport
	Generation uint64
}

var loadCache = struct {
	mu    sync.RWMutex
	order []string
	items map[string]loadCacheEntry
}{
	order: make([]string, 0, loadCacheMaxEntries),
	items: map[string]loadCacheEntry{},
}

type loadGenerationEntry struct {
	mu               sync.RWMutex
	Fingerprint      string
	State            []loadSourceValidationState
	Generation       uint64
	watcher          *loadSourceWatcher
	nextWatchRetryAt time.Time
}

var loadGenerationCache = struct {
	mu    sync.RWMutex
	order []string
	items map[string]*loadGenerationEntry
}{
	order: make([]string, 0, loadCacheMaxEntries),
	items: map[string]*loadGenerationEntry{},
}

func loadFromSourcesCached(sources []loadSource, opts LoadOptions) ([]Skill, LoadReport, error) {
	cacheKey, ok := buildLoadCacheLookupKey(sources, opts)
	generation := uint64(0)
	if ok {
		if cached, found := peekLoadCache(cacheKey); found {
			if current, generationOK := resolveLoadGeneration(cacheKey, sources, opts); generationOK {
				generation = current
			}
			if generation > 0 && cached.Generation == generation {
				skills, report := cloneLoadCacheHit(cached)
				report.CacheHit = true
				report.Generation = generation
				return skills, report, nil
			}
		}
	}
	items, report, err := loadFromSources(sources, opts)
	if err != nil {
		return nil, report, err
	}
	report.CacheHit = false
	if ok && generation == 0 {
		if current, generationOK := resolveLoadGeneration(cacheKey, sources, opts); generationOK {
			generation = current
		}
	}
	report.Generation = generation
	if ok && generation > 0 {
		setLoadCache(cacheKey, loadCacheEntry{
			Skills:     cloneSkills(items),
			Report:     cloneLoadReport(report),
			Generation: generation,
		})
	}
	return items, report, nil
}

func getLoadCache(key string) (loadCacheEntry, bool) {
	loadCache.mu.RLock()
	defer loadCache.mu.RUnlock()
	entry, ok := loadCache.items[key]
	if !ok {
		return loadCacheEntry{}, false
	}
	return loadCacheEntry{
		Skills:     cloneSkills(entry.Skills),
		Report:     cloneLoadReport(entry.Report),
		Generation: entry.Generation,
	}, true
}

func peekLoadCache(key string) (loadCacheEntry, bool) {
	loadCache.mu.RLock()
	defer loadCache.mu.RUnlock()
	entry, ok := loadCache.items[key]
	return entry, ok
}

func cloneLoadCacheHit(entry loadCacheEntry) ([]Skill, LoadReport) {
	return cloneSkills(entry.Skills), cloneLoadReport(entry.Report)
}

func setLoadCache(key string, entry loadCacheEntry) {
	loadCache.mu.Lock()
	defer loadCache.mu.Unlock()
	if _, exists := loadCache.items[key]; exists {
		loadCache.items[key] = entry
		return
	}
	if len(loadCache.order) >= loadCacheMaxEntries {
		evict := loadCache.order[0]
		loadCache.order = loadCache.order[1:]
		delete(loadCache.items, evict)
	}
	loadCache.order = append(loadCache.order, key)
	loadCache.items[key] = entry
}

func resolveLoadGeneration(key string, sources []loadSource, opts LoadOptions) (uint64, bool) {
	if strings.TrimSpace(key) == "" {
		return 0, false
	}
	now := time.Now()
	if entry, ok := getLoadGenerationEntry(key); ok {
		entry.ensureWatcher(sources, now)
		if generation, ok := entry.tryFastValidateGeneration(sources); ok {
			return generation, true
		}
	}
	fingerprint, state, ok := loadFingerprintBuilder(sources, opts)
	if !ok || strings.TrimSpace(fingerprint) == "" {
		return 0, false
	}
	entry := upsertLoadGenerationEntry(key, fingerprint, state)
	if entry == nil {
		return 0, false
	}
	entry.ensureWatcher(sources, now)
	return entry.currentGeneration(), true
}

func getLoadGenerationEntry(key string) (*loadGenerationEntry, bool) {
	loadGenerationCache.mu.RLock()
	defer loadGenerationCache.mu.RUnlock()
	entry, ok := loadGenerationCache.items[key]
	return entry, ok
}

func upsertLoadGenerationEntry(key string, fingerprint string, state []loadSourceValidationState) *loadGenerationEntry {
	loadGenerationCache.mu.Lock()
	defer loadGenerationCache.mu.Unlock()
	if entry, exists := loadGenerationCache.items[key]; exists {
		entry.mu.Lock()
		if entry.Fingerprint == "" {
			entry.Fingerprint = fingerprint
			entry.State = cloneLoadValidationStates(state)
		} else if entry.Fingerprint != fingerprint {
			entry.Fingerprint = fingerprint
			entry.State = cloneLoadValidationStates(state)
			entry.Generation++
		} else if len(state) > 0 {
			entry.State = cloneLoadValidationStates(state)
		}
		entry.mu.Unlock()
		return entry
	}
	entry := &loadGenerationEntry{
		Fingerprint: fingerprint,
		State:       cloneLoadValidationStates(state),
		Generation:  1,
	}
	if len(loadGenerationCache.order) >= loadCacheMaxEntries {
		evict := loadGenerationCache.order[0]
		loadGenerationCache.order = loadGenerationCache.order[1:]
		if existing := loadGenerationCache.items[evict]; existing != nil {
			existing.closeWatcher()
		}
		delete(loadGenerationCache.items, evict)
	}
	loadGenerationCache.order = append(loadGenerationCache.order, key)
	loadGenerationCache.items[key] = entry
	return entry
}

func (e *loadGenerationEntry) currentGeneration() uint64 {
	if e == nil {
		return 0
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Generation
}

func (e *loadGenerationEntry) bumpGeneration() {
	if e == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Generation++
	e.Fingerprint = ""
	e.State = nil
}

func (e *loadGenerationEntry) tryFastValidateGeneration(sources []loadSource) (uint64, bool) {
	if e == nil {
		return 0, false
	}
	e.mu.RLock()
	watcherInstalled := e.watcher != nil
	state := cloneLoadValidationStates(e.State)
	generation := e.Generation
	e.mu.RUnlock()
	if !watcherInstalled || generation == 0 || len(state) == 0 {
		return 0, false
	}
	if !fastValidateLoadValidationState(sources, state) {
		return 0, false
	}
	return generation, true
}

func (e *loadGenerationEntry) ensureWatcher(sources []loadSource, now time.Time) {
	if e == nil {
		return
	}
	e.mu.Lock()
	if e.watcher != nil {
		e.mu.Unlock()
		return
	}
	if !e.nextWatchRetryAt.IsZero() && now.Before(e.nextWatchRetryAt) {
		e.mu.Unlock()
		return
	}
	e.mu.Unlock()

	watcher, err := loadSourceWatcherFactory(sources, e)
	if err != nil {
		e.mu.Lock()
		e.nextWatchRetryAt = now.Add(loadWatcherRetryBackoff)
		e.mu.Unlock()
		return
	}
	e.mu.Lock()
	e.watcher = watcher
	e.nextWatchRetryAt = time.Time{}
	e.mu.Unlock()
}

func (e *loadGenerationEntry) closeWatcher() {
	if e == nil {
		return
	}
	e.mu.Lock()
	watcher := e.watcher
	e.watcher = nil
	e.mu.Unlock()
	if watcher != nil {
		_ = watcher.Close()
	}
}

func (e *loadGenerationEntry) releaseWatcher(watcher *loadSourceWatcher) {
	if e == nil || watcher == nil {
		return
	}
	e.mu.Lock()
	if e.watcher == watcher {
		e.watcher = nil
	}
	e.mu.Unlock()
}

type loadSourceWatcher struct {
	watcher *fsnotify.Watcher
	mu      sync.Mutex
	dirs    map[string]struct{}
	roots   []string
}

func newLoadSourceWatcher(sources []loadSource, entry *loadGenerationEntry) (*loadSourceWatcher, error) {
	sources = loadWatcherSources(sources)
	if len(sources) == 0 {
		return &loadSourceWatcher{
			dirs:  map[string]struct{}{},
			roots: nil,
		}, nil
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	loadWatcher := &loadSourceWatcher{
		watcher: watcher,
		dirs:    map[string]struct{}{},
		roots:   loadWatcherRoots(sources),
	}
	for _, source := range sources {
		if err := loadWatcher.addRecursive(source.Dir); err != nil {
			_ = watcher.Close()
			return nil, err
		}
	}
	go loadWatcher.run(entry)
	return loadWatcher, nil
}

func loadWatcherSources(sources []loadSource) []loadSource {
	out := make([]loadSource, 0, len(sources))
	for _, source := range sources {
		if !loadWatcherSourceEnabled(source) {
			continue
		}
		out = append(out, source)
	}
	return out
}

func loadWatcherSourceEnabled(source loadSource) bool {
	if source.FSBacked {
		return false
	}
	if source.Kind == SourceBundled {
		return false
	}
	return strings.TrimSpace(source.Dir) != ""
}

func (w *loadSourceWatcher) Close() error {
	if w == nil || w.watcher == nil {
		return nil
	}
	return w.watcher.Close()
}

func (w *loadSourceWatcher) run(entry *loadGenerationEntry) {
	defer entry.releaseWatcher(w)
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename|fsnotify.Chmod) == 0 {
				continue
			}
			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					_ = w.addRecursive(event.Name)
				}
			}
			if !w.shouldInvalidate(event, entry) {
				continue
			}
			entry.bumpGeneration()
		case _, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			entry.bumpGeneration()
		}
	}
}

func (w *loadSourceWatcher) addRecursive(root string) error {
	trimmed := strings.TrimSpace(root)
	if trimmed == "" {
		return nil
	}
	info, err := os.Stat(trimmed)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}
	return filepath.WalkDir(trimmed, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			return nil
		}
		return w.add(path)
	})
}

func (w *loadSourceWatcher) add(path string) error {
	cleaned := normalizeLoadCachePath(path)
	if cleaned == "" {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, exists := w.dirs[cleaned]; exists {
		return nil
	}
	if err := w.watcher.Add(cleaned); err != nil {
		return err
	}
	w.dirs[cleaned] = struct{}{}
	return nil
}

func loadWatcherRoots(sources []loadSource) []string {
	roots := make([]string, 0, len(sources))
	for _, source := range sources {
		if trimmed := normalizeLoadCachePath(source.Dir); trimmed != "" {
			roots = append(roots, filepath.Clean(trimmed))
		}
	}
	return roots
}

func (w *loadSourceWatcher) shouldInvalidate(event fsnotify.Event, entry *loadGenerationEntry) bool {
	if entry == nil {
		return true
	}
	entry.mu.RLock()
	state := cloneLoadValidationStates(entry.State)
	entry.mu.RUnlock()
	if len(state) == 0 {
		return true
	}
	return shouldInvalidateLoadEvent(w.roots, state, event)
}

type loadSourceValidationState struct {
	Kind           Source
	Dir            string
	FSBacked       bool
	ID             string
	Fingerprint    string
	Missing        bool
	DirExists      bool
	DirModTimeNano int64
	SkillDirs      []loadSkillDirValidationState
}

type loadSkillDirValidationState struct {
	Name         string
	NativeDir    string
	SkillDoc     loadFileValidationState
	ResourceDirs []loadResourceDirValidationState
}

type loadFileValidationState struct {
	NativePath  string
	Exists      bool
	Size        int64
	ModTimeNano int64
}

type loadResourceDirValidationState struct {
	NativePath string
	Name       string
	Exists     bool
	DirStates  []loadDirValidationState
	Files      []string
}

type loadDirValidationState struct {
	Path        string
	NativePath  string
	ModTimeNano int64
}

func buildLoadCacheLookupKey(sources []loadSource, opts LoadOptions) (string, bool) {
	hasher := fnv.New64a()
	writeLoadCacheString(hasher, "v3")
	writeLoadCacheBool(hasher, opts.FailFast)
	writeLoadCacheBool(hasher, opts.StrictFrontmatter)
	writeLoadCacheInt(hasher, opts.MaxCandidatesPerRoot)
	writeLoadCacheInt(hasher, opts.MaxSkillsLoadedPerSource)
	writeLoadCacheInt(hasher, opts.MaxSkillFileBytes)
	for _, source := range sources {
		writeLoadCacheString(hasher, string(source.Kind))
		writeLoadCacheBool(hasher, source.FSBacked)
		if source.FSBacked {
			if !isCacheableFSSource(source) {
				return "", false
			}
			writeLoadCacheString(hasher, strings.TrimSpace(source.ID))
			writeLoadCacheString(hasher, strings.TrimSpace(source.Fingerprint))
			continue
		}
		writeLoadCacheString(hasher, normalizeLoadCachePath(source.Dir))
	}
	return strconv.FormatUint(hasher.Sum64(), 16), true
}

func buildLoadFingerprint(sources []loadSource, opts LoadOptions) (string, []loadSourceValidationState, bool) {
	state, ok := buildLoadValidationState(sources)
	if !ok {
		return "", nil, false
	}
	hasher := fnv.New64a()
	writeLoadCacheString(hasher, "v3")
	writeLoadCacheBool(hasher, opts.FailFast)
	writeLoadCacheBool(hasher, opts.StrictFrontmatter)
	writeLoadCacheInt(hasher, opts.MaxCandidatesPerRoot)
	writeLoadCacheInt(hasher, opts.MaxSkillsLoadedPerSource)
	writeLoadCacheInt(hasher, opts.MaxSkillFileBytes)
	for _, source := range state {
		writeLoadCacheString(hasher, string(source.Kind))
		writeLoadCacheBool(hasher, source.FSBacked)
		if source.FSBacked {
			writeLoadCacheString(hasher, source.ID)
			writeLoadCacheString(hasher, source.Fingerprint)
			continue
		}
		writeLoadCacheString(hasher, source.Dir)
		writeLoadCacheBool(hasher, source.Missing)
		writeLoadCacheBool(hasher, source.DirExists)
		writeLoadCacheInt(hasher, len(source.SkillDirs))
		for _, skillDir := range source.SkillDirs {
			writeLoadCacheString(hasher, skillDir.Name)
			writeLoadCacheBool(hasher, skillDir.SkillDoc.Exists)
			writeLoadCacheInt64(hasher, skillDir.SkillDoc.Size)
			writeLoadCacheInt64(hasher, skillDir.SkillDoc.ModTimeNano)
			writeLoadCacheInt(hasher, len(skillDir.ResourceDirs))
			for _, resource := range skillDir.ResourceDirs {
				writeLoadCacheString(hasher, resource.Name)
				writeLoadCacheBool(hasher, resource.Exists)
				writeLoadCacheInt(hasher, len(resource.Files))
				for _, resourceFile := range resource.Files {
					writeLoadCacheString(hasher, resourceFile)
				}
			}
		}
	}
	return strconv.FormatUint(hasher.Sum64(), 16), state, true
}

func buildLoadValidationState(sources []loadSource) ([]loadSourceValidationState, bool) {
	if len(sources) == 0 {
		return nil, true
	}
	workers := loadValidationWorkers(len(sources))
	if workers <= 1 {
		return buildLoadValidationStateRange(sources)
	}
	return runLoadValidationWorkers(workers, len(sources), func(start, end int) ([]loadSourceValidationState, bool) {
		return buildLoadValidationStateRange(sources[start:end])
	})
}

func buildLoadValidationStateRange(sources []loadSource) ([]loadSourceValidationState, bool) {
	states := make([]loadSourceValidationState, 0, len(sources))
	for _, source := range sources {
		state, ok := buildLoadSourceValidationState(source)
		if !ok {
			return nil, false
		}
		states = append(states, state)
	}
	return states, true
}

func buildLoadSourceValidationState(source loadSource) (loadSourceValidationState, bool) {
	state := loadSourceValidationState{
		Kind:     source.Kind,
		FSBacked: source.FSBacked,
	}
	if source.FSBacked {
		if !isCacheableFSSource(source) {
			return loadSourceValidationState{}, false
		}
		state.ID = strings.TrimSpace(source.ID)
		state.Fingerprint = strings.TrimSpace(source.Fingerprint)
		return state, true
	}
	state.Dir = normalizeLoadCachePath(source.Dir)
	if info, err := os.Stat(source.Dir); err == nil && info.IsDir() {
		state.DirExists = true
		state.DirModTimeNano = info.ModTime().UnixNano()
	}
	entries, err := os.ReadDir(source.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			state.Missing = true
			return state, true
		}
		return loadSourceValidationState{}, false
	}
	skillEntries := make([]fs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillEntries = append(skillEntries, entry)
	}
	skillDirs, ok := buildLoadSkillDirValidationStates(source.Dir, skillEntries)
	if !ok {
		return loadSourceValidationState{}, false
	}
	sort.Slice(skillDirs, func(i, j int) bool {
		return skillDirs[i].Name < skillDirs[j].Name
	})
	state.SkillDirs = skillDirs
	return state, true
}

func fastValidateLoadValidationState(sources []loadSource, state []loadSourceValidationState) bool {
	if len(sources) != len(state) {
		return false
	}
	workers := loadValidationWorkers(len(sources))
	if workers <= 1 {
		return fastValidateLoadValidationStateRange(sources, state)
	}
	return runLoadValidationBoolWorkers(workers, len(sources), func(start, end int) bool {
		return fastValidateLoadValidationStateRange(sources[start:end], state[start:end])
	})
}

func fastValidateLoadValidationStateRange(sources []loadSource, state []loadSourceValidationState) bool {
	for idx, source := range sources {
		expected := state[idx]
		if expected.Kind != source.Kind || expected.FSBacked != source.FSBacked {
			return false
		}
		if source.FSBacked {
			if !isCacheableFSSource(source) ||
				strings.TrimSpace(source.ID) != expected.ID ||
				strings.TrimSpace(source.Fingerprint) != expected.Fingerprint {
				return false
			}
			continue
		}
		if normalizeLoadCachePath(source.Dir) != expected.Dir {
			return false
		}
		if !fastValidateLoadSourceValidationState(source, expected) {
			return false
		}
	}
	return true
}

func isCacheableFSSource(source loadSource) bool {
	spec := source.fsSource()
	return spec.Valid() && agentxassetfs.IsProviderFS(spec.FS, spec.ID, spec.Fingerprint)
}

func fastValidateLoadSourceValidationState(source loadSource, expected loadSourceValidationState) bool {
	info, err := os.Stat(source.Dir)
	if err != nil {
		return expected.Missing && os.IsNotExist(err)
	}
	if !info.IsDir() || expected.Missing || !expected.DirExists {
		return false
	}
	if info.ModTime().UnixNano() != expected.DirModTimeNano {
		if !fastValidateLoadSkillDirNames(source.Dir, expected.SkillDirs) {
			return false
		}
	}
	workers := loadValidationWorkers(len(expected.SkillDirs))
	if workers <= 1 {
		return fastValidateLoadSkillDirValidationStatesRange(source.Dir, expected.SkillDirs)
	}
	return runLoadValidationBoolWorkers(workers, len(expected.SkillDirs), func(start, end int) bool {
		return fastValidateLoadSkillDirValidationStatesRange(source.Dir, expected.SkillDirs[start:end])
	})
}

func fastValidateLoadSkillDirValidationState(sourceDir string, expected loadSkillDirValidationState) bool {
	skillDir := expected.NativeDir
	if strings.TrimSpace(skillDir) == "" {
		skillDir = filepath.Join(sourceDir, expected.Name)
	}
	if !fastValidateLoadFile(filepath.Join(skillDir, skillDocName), expected.SkillDoc) {
		return false
	}
	for _, resource := range expected.ResourceDirs {
		if !fastValidateLoadResourceDir(skillDir, resource) {
			return false
		}
	}
	return true
}

func fastValidateLoadSkillDirNames(sourceDir string, expected []loadSkillDirValidationState) bool {
	actual, ok := loadSourceSkillDirNames(sourceDir)
	if !ok {
		return false
	}
	if len(actual) != len(expected) {
		return false
	}
	for idx, name := range actual {
		if expected[idx].Name != name {
			return false
		}
	}
	return true
}

func fastValidateLoadSkillDirValidationStatesRange(sourceDir string, expected []loadSkillDirValidationState) bool {
	for _, skillState := range expected {
		if !fastValidateLoadSkillDirValidationState(sourceDir, skillState) {
			return false
		}
	}
	return true
}

func fastValidateLoadFile(path string, expected loadFileValidationState) bool {
	if trimmed := strings.TrimSpace(expected.NativePath); trimmed != "" {
		path = trimmed
	}
	info, err := os.Stat(path)
	if err != nil {
		return !expected.Exists && os.IsNotExist(err)
	}
	if info.IsDir() || !expected.Exists {
		return false
	}
	return info.Size() == expected.Size && info.ModTime().UnixNano() == expected.ModTimeNano
}

func fastValidateLoadResourceDir(skillDir string, expected loadResourceDirValidationState) bool {
	base := expected.NativePath
	if strings.TrimSpace(base) == "" {
		base = filepath.Join(skillDir, expected.Name)
	}
	info, err := os.Stat(base)
	if err != nil {
		return !expected.Exists && os.IsNotExist(err)
	}
	if !info.IsDir() || !expected.Exists {
		return false
	}
	actualFiles, ok := listLoadResourceFiles(skillDir, expected.Name)
	if !ok {
		return false
	}
	if len(actualFiles) != len(expected.Files) {
		return false
	}
	for idx, path := range actualFiles {
		if expected.Files[idx] != path {
			return false
		}
	}
	return true
}

func buildLoadSkillDirValidationState(skillDir string, name string) (loadSkillDirValidationState, bool) {
	state := loadSkillDirValidationState{
		Name:      name,
		NativeDir: filepath.Clean(skillDir),
	}
	skillDoc, ok := statLoadValidationFile(filepath.Join(skillDir, skillDocName))
	if !ok {
		return loadSkillDirValidationState{}, false
	}
	state.SkillDoc = skillDoc
	resourceDirs := make([]loadResourceDirValidationState, 0, 3)
	for _, dirName := range []string{"scripts", "references", "assets"} {
		resourceState, ok := buildLoadResourceDirValidationState(skillDir, dirName)
		if !ok {
			return loadSkillDirValidationState{}, false
		}
		resourceDirs = append(resourceDirs, resourceState)
	}
	state.ResourceDirs = resourceDirs
	return state, true
}

func buildLoadSkillDirValidationStates(sourceDir string, entries []fs.DirEntry) ([]loadSkillDirValidationState, bool) {
	if len(entries) == 0 {
		return nil, true
	}
	workers := loadValidationWorkers(len(entries))
	if workers <= 1 {
		return buildLoadSkillDirValidationStatesRange(sourceDir, entries)
	}
	return runLoadSkillDirValidationWorkers(workers, len(entries), func(start, end int) ([]loadSkillDirValidationState, bool) {
		return buildLoadSkillDirValidationStatesRange(sourceDir, entries[start:end])
	})
}

func buildLoadSkillDirValidationStatesRange(sourceDir string, entries []fs.DirEntry) ([]loadSkillDirValidationState, bool) {
	skillDirs := make([]loadSkillDirValidationState, 0, len(entries))
	for _, entry := range entries {
		skillState, ok := buildLoadSkillDirValidationState(filepath.Join(sourceDir, entry.Name()), entry.Name())
		if !ok {
			return nil, false
		}
		if !skillState.SkillDoc.Exists {
			continue
		}
		skillDirs = append(skillDirs, skillState)
	}
	return skillDirs, true
}

func loadSourceSkillDirNames(sourceDir string) ([]string, bool) {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, false
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillDoc, ok := statLoadValidationFile(filepath.Join(sourceDir, entry.Name(), skillDocName))
		if !ok {
			return nil, false
		}
		if !skillDoc.Exists {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, true
}

func statLoadValidationFile(path string) (loadFileValidationState, bool) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return loadFileValidationState{}, true
		}
		return loadFileValidationState{}, false
	}
	if info.IsDir() {
		return loadFileValidationState{}, true
	}
	return loadFileValidationState{
		NativePath:  filepath.Clean(path),
		Exists:      true,
		Size:        info.Size(),
		ModTimeNano: info.ModTime().UnixNano(),
	}, true
}

func buildLoadResourceDirValidationState(skillDir string, dirName string) (loadResourceDirValidationState, bool) {
	base := filepath.Join(skillDir, dirName)
	info, err := os.Stat(base)
	if err != nil {
		if os.IsNotExist(err) {
			return loadResourceDirValidationState{Name: dirName}, true
		}
		return loadResourceDirValidationState{}, false
	}
	if !info.IsDir() {
		return loadResourceDirValidationState{Name: dirName}, true
	}
	dirStates := make([]loadDirValidationState, 0, 4)
	files := make([]string, 0, 8)
	walkErr := filepath.WalkDir(base, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			rel, relErr := filepath.Rel(skillDir, path)
			if relErr != nil {
				return relErr
			}
			files = append(files, filepath.ToSlash(rel))
			return nil
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}
		rel, relErr := filepath.Rel(skillDir, path)
		if relErr != nil {
			return relErr
		}
		dirStates = append(dirStates, loadDirValidationState{
			Path:        filepath.ToSlash(rel),
			NativePath:  filepath.Clean(path),
			ModTimeNano: info.ModTime().UnixNano(),
		})
		return nil
	})
	if walkErr != nil {
		return loadResourceDirValidationState{}, false
	}
	sort.Slice(dirStates, func(i, j int) bool {
		return dirStates[i].Path < dirStates[j].Path
	})
	sort.Strings(files)
	return loadResourceDirValidationState{
		NativePath: filepath.Clean(base),
		Name:       dirName,
		Exists:     true,
		DirStates:  dirStates,
		Files:      files,
	}, true
}

func listLoadResourceFiles(skillDir string, dirName string) ([]string, bool) {
	base := filepath.Join(skillDir, dirName)
	info, err := os.Stat(base)
	if err != nil {
		return nil, os.IsNotExist(err)
	}
	if !info.IsDir() {
		return nil, true
	}
	files := make([]string, 0, 8)
	walkErr := filepath.WalkDir(base, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(skillDir, path)
		if relErr != nil {
			return relErr
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if walkErr != nil {
		return nil, false
	}
	sort.Strings(files)
	return files, true
}

func normalizeLoadCachePath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if abs, err := filepath.Abs(trimmed); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(trimmed)
}

func shouldInvalidateLoadEvent(roots []string, state []loadSourceValidationState, event fsnotify.Event) bool {
	name := normalizeLoadCachePath(event.Name)
	if name == "" {
		return true
	}
	for _, source := range state {
		root := filepath.Clean(normalizeLoadCachePath(source.Dir))
		if root == "" {
			continue
		}
		if root != name && !strings.HasPrefix(name, root+string(os.PathSeparator)) {
			continue
		}
		rel, err := filepath.Rel(root, name)
		if err != nil {
			return true
		}
		rel = filepath.Clean(rel)
		if rel == "." {
			return false
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) == 0 {
			return false
		}
		skillName := parts[0]
		skillExists := false
		for _, skillDir := range source.SkillDirs {
			if skillDir.Name == skillName {
				skillExists = true
				break
			}
		}
		if len(parts) == 1 {
			if !skillExists {
				if event.Op&fsnotify.Create == 0 {
					return false
				}
				info, statErr := os.Stat(name)
				if statErr != nil || !info.IsDir() {
					return false
				}
				skillDoc, ok := statLoadValidationFile(filepath.Join(name, skillDocName))
				return ok && skillDoc.Exists
			}
			return true
		}
		second := parts[1]
		if second == skillDocName {
			return true
		}
		if second == "scripts" || second == "references" || second == "assets" {
			return event.Op&(fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0
		}
		if !skillExists {
			return false
		}
		return false
	}
	for _, root := range roots {
		cleanedRoot := filepath.Clean(normalizeLoadCachePath(root))
		if cleanedRoot == "" {
			continue
		}
		if cleanedRoot == name || strings.HasPrefix(name, cleanedRoot+string(os.PathSeparator)) {
			return true
		}
	}
	return true
}

func writeLoadCacheString(hasher hash.Hash64, value string) {
	_, _ = hasher.Write([]byte(value))
	_, _ = hasher.Write([]byte{0})
}

func writeLoadCacheBool(hasher hash.Hash64, value bool) {
	if value {
		_, _ = hasher.Write([]byte{1})
		return
	}
	_, _ = hasher.Write([]byte{0})
}

func writeLoadCacheInt(hasher hash.Hash64, value int) {
	writeLoadCacheInt64(hasher, int64(value))
}

func writeLoadCacheInt64(hasher hash.Hash64, value int64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(value))
	_, _ = hasher.Write(buf[:])
}

func loadValidationWorkers(count int) int {
	if count < loadValidationParallelThreshold {
		return 1
	}
	workers := runtime.GOMAXPROCS(0)
	if workers <= 1 {
		return 1
	}
	if workers > loadValidationMaxWorkers {
		workers = loadValidationMaxWorkers
	}
	if workers > count {
		workers = count
	}
	if workers < 1 {
		return 1
	}
	return workers
}

type loadValidationResult struct {
	items []loadSourceValidationState
	ok    bool
}

func runLoadValidationWorkers(workers int, count int, fn func(start, end int) ([]loadSourceValidationState, bool)) ([]loadSourceValidationState, bool) {
	if workers <= 1 || count <= 0 {
		return fn(0, count)
	}
	chunkSize := (count + workers - 1) / workers
	results := make(chan loadValidationResult, workers)
	launched := 0
	for start := 0; start < count; start += chunkSize {
		end := start + chunkSize
		if end > count {
			end = count
		}
		launched++
		go func(start, end int) {
			items, ok := fn(start, end)
			results <- loadValidationResult{items: items, ok: ok}
		}(start, end)
	}
	merged := make([]loadSourceValidationState, 0, count)
	for i := 0; i < launched; i++ {
		result := <-results
		if !result.ok {
			return nil, false
		}
		merged = append(merged, result.items...)
	}
	return merged, true
}

type loadSkillDirValidationResult struct {
	items []loadSkillDirValidationState
	ok    bool
}

func runLoadSkillDirValidationWorkers(workers int, count int, fn func(start, end int) ([]loadSkillDirValidationState, bool)) ([]loadSkillDirValidationState, bool) {
	if workers <= 1 || count <= 0 {
		return fn(0, count)
	}
	chunkSize := (count + workers - 1) / workers
	results := make(chan loadSkillDirValidationResult, workers)
	launched := 0
	for start := 0; start < count; start += chunkSize {
		end := start + chunkSize
		if end > count {
			end = count
		}
		launched++
		go func(start, end int) {
			items, ok := fn(start, end)
			results <- loadSkillDirValidationResult{items: items, ok: ok}
		}(start, end)
	}
	merged := make([]loadSkillDirValidationState, 0, count)
	for i := 0; i < launched; i++ {
		result := <-results
		if !result.ok {
			return nil, false
		}
		merged = append(merged, result.items...)
	}
	return merged, true
}

func runLoadValidationBoolWorkers(workers int, count int, fn func(start, end int) bool) bool {
	if workers <= 1 || count <= 0 {
		return fn(0, count)
	}
	chunkSize := (count + workers - 1) / workers
	results := make(chan bool, workers)
	launched := 0
	for start := 0; start < count; start += chunkSize {
		end := start + chunkSize
		if end > count {
			end = count
		}
		launched++
		go func(start, end int) {
			results <- fn(start, end)
		}(start, end)
	}
	for i := 0; i < launched; i++ {
		if !<-results {
			return false
		}
	}
	return true
}

func cloneSkills(items []Skill) []Skill {
	if len(items) == 0 {
		return nil
	}
	out := make([]Skill, 0, len(items))
	for _, item := range items {
		out = append(out, cloneSkill(item))
	}
	return out
}

func cloneSkill(item Skill) Skill {
	cloned := item
	cloned.Keywords = append([]string(nil), item.Keywords...)
	cloned.Tags = append([]string(nil), item.Tags...)
	cloned.WhenToUse = append([]string(nil), item.WhenToUse...)
	cloned.WhenNotToUse = append([]string(nil), item.WhenNotToUse...)
	cloned.NegativeExamples = append([]string(nil), item.NegativeExamples...)
	cloned.Steps = append([]string(nil), item.Steps...)
	cloned.Paths = append([]string(nil), item.Paths...)
	cloned.ToolHints = append([]string(nil), item.ToolHints...)
	cloned.Examples = append([]string(nil), item.Examples...)
	cloned.EvalAssertions = append([]string(nil), item.EvalAssertions...)
	cloned.AllowedTools = append([]string(nil), item.AllowedTools...)
	cloned.OS = append([]string(nil), item.OS...)
	cloned.Requires.Bins = append([]string(nil), item.Requires.Bins...)
	cloned.Requires.AnyBins = append([]string(nil), item.Requires.AnyBins...)
	cloned.Requires.Env = append([]string(nil), item.Requires.Env...)
	cloned.Requires.Config = append([]string(nil), item.Requires.Config...)
	cloned.Install = cloneInstallSpecs(item.Install)
	cloned.Resources.Scripts = append([]string(nil), item.Resources.Scripts...)
	cloned.Resources.References = append([]string(nil), item.Resources.References...)
	cloned.Resources.Assets = append([]string(nil), item.Resources.Assets...)
	if item.Metadata != nil {
		cloned.Metadata = make(map[string]string, len(item.Metadata))
		for key, value := range item.Metadata {
			cloned.Metadata[key] = value
		}
	}
	if item.Dispatch != nil {
		dispatch := *item.Dispatch
		dispatch.Aliases = append([]string(nil), item.Dispatch.Aliases...)
		cloned.Dispatch = &dispatch
	}
	return cloned
}

func cloneInstallSpecs(items []InstallSpec) []InstallSpec {
	if len(items) == 0 {
		return nil
	}
	out := make([]InstallSpec, 0, len(items))
	for _, item := range items {
		cloned := item
		cloned.DependsOn = append([]string(nil), item.DependsOn...)
		cloned.Bins = append([]string(nil), item.Bins...)
		cloned.OS = append([]string(nil), item.OS...)
		cloned.Command = append([]string(nil), item.Command...)
		cloned.Rollback = append([]string(nil), item.Rollback...)
		out = append(out, cloned)
	}
	return out
}

func cloneLoadReport(report LoadReport) LoadReport {
	cloned := report
	cloned.Issues = append([]LoadIssue(nil), report.Issues...)
	return cloned
}

func cloneLoadValidationStates(items []loadSourceValidationState) []loadSourceValidationState {
	if len(items) == 0 {
		return nil
	}
	out := make([]loadSourceValidationState, 0, len(items))
	for _, item := range items {
		cloned := item
		cloned.SkillDirs = cloneLoadSkillDirValidationStates(item.SkillDirs)
		out = append(out, cloned)
	}
	return out
}

func cloneLoadSkillDirValidationStates(items []loadSkillDirValidationState) []loadSkillDirValidationState {
	if len(items) == 0 {
		return nil
	}
	out := make([]loadSkillDirValidationState, 0, len(items))
	for _, item := range items {
		cloned := item
		cloned.ResourceDirs = cloneLoadResourceDirValidationStates(item.ResourceDirs)
		out = append(out, cloned)
	}
	return out
}

func cloneLoadResourceDirValidationStates(items []loadResourceDirValidationState) []loadResourceDirValidationState {
	if len(items) == 0 {
		return nil
	}
	out := make([]loadResourceDirValidationState, 0, len(items))
	for _, item := range items {
		cloned := item
		cloned.DirStates = append([]loadDirValidationState(nil), item.DirStates...)
		cloned.Files = append([]string(nil), item.Files...)
		out = append(out, cloned)
	}
	return out
}
