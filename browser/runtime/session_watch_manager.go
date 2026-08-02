package browserruntime

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

const sharedSessionBrowserWatchManagerEventCycleCacheTTL = 100 * time.Millisecond
const sharedSessionBrowserObserverManagerIdleTTL = 10 * time.Minute
const sharedSessionBrowserWatchManagerIdleTTL = 10 * time.Minute

// SharedSessionBrowserObserverManager owns the shared source-time watch cycle
// dependencies for session-scoped browser observation. It is the explicit
// session observer/watchdog manager that binds registries and reconnect timing
// to the watch-loop contract.
type SharedSessionBrowserObserverManager struct {
	SessionRegistry *BrowserSessionRegistry
	RunRegistry     SharedSessionRunRegistry
	StateRegistry   SharedSessionBrowserStateRegistry
	ReconnectWindow time.Duration
	cache           *sharedSessionBrowserWatchManagerCache
	provider        *sharedSessionBrowserObserverManagerProviderEntry
}

// SharedSessionBrowserWatchManager binds the shared observer/watchdog manager
// to a concrete backend control so source-time observation entrypoints can be
// reused as a long-lived watch manager without rethreading control on every
// call.
type SharedSessionBrowserWatchManager struct {
	Observer SharedSessionBrowserObserverManager
	Control  BrowserRuntimeControlBackend
	state    *sharedSessionBrowserWatchManagerState
}

type sharedSessionBrowserWatchManagerCacheKey struct {
	typ reflect.Type
	ptr uintptr
}

type sharedSessionBrowserWatchManagerCache struct {
	mu         sync.RWMutex
	managers   map[sharedSessionBrowserWatchManagerCacheKey]SharedSessionBrowserWatchManager
	generation atomic.Uint64
	rawGen     atomic.Uint64
}

type sharedSessionBrowserObserverManagerProviderCacheKey struct {
	sessionRegistry sharedSessionBrowserWatchManagerCacheKey
	runRegistry     sharedSessionBrowserWatchManagerCacheKey
	stateRegistry   sharedSessionBrowserWatchManagerCacheKey
	reconnectWindow time.Duration
}

type sharedSessionBrowserObserverManagerProviderEntry struct {
	mu           sync.RWMutex
	lastActiveAt time.Time
	manager      SharedSessionBrowserObserverManager
}

type sharedSessionBrowserWatchManagerState struct {
	mu                  sync.RWMutex
	generation          uint64
	rawGeneration       uint64
	lastActiveAt        time.Time
	rawStatus           map[string]sharedSessionBrowserCachedRawStatusObservation
	rawProfiles         map[string]sharedSessionBrowserCachedRawProfilesObservation
	rawRouteMutations   map[sharedSessionBrowserRouteMutationSourceKey]sharedSessionBrowserCachedRawRouteMutationObservation
	routeMutations      map[sharedSessionBrowserRouteMutationSourceKey]sharedSessionBrowserCachedEventCycleObservation
	rawStatusInFlight   map[string]*sharedSessionBrowserInFlightRawStatusObservation
	rawProfilesInFlight map[string]*sharedSessionBrowserInFlightRawProfilesObservation
	rawStartsInFlight   map[string]*sharedSessionBrowserInFlightRawLifecycleObservation
	rawStopsInFlight    map[string]*sharedSessionBrowserInFlightRawLifecycleObservation
	eventCycles         map[SharedSessionBrowserObserverRequest]sharedSessionBrowserCachedEventCycleObservation
	bindings            map[SharedSessionBrowserObserverRequest]sharedSessionBrowserCachedBindingObservation
	views               map[SharedSessionBrowserObserverRequest]sharedSessionBrowserCachedViewObservation
	watchLoops          map[SharedSessionBrowserObserverRequest]sharedSessionBrowserCachedWatchLoopObservation
	eventCyclesInFlight map[SharedSessionBrowserObserverRequest]*sharedSessionBrowserInFlightEventCycleObservation
	bindingsInFlight    map[SharedSessionBrowserObserverRequest]*sharedSessionBrowserInFlightBindingObservation
	viewsInFlight       map[SharedSessionBrowserObserverRequest]*sharedSessionBrowserInFlightViewObservation
	watchLoopsInFlight  map[SharedSessionBrowserObserverRequest]*sharedSessionBrowserInFlightWatchLoopObservation
}

type sharedSessionBrowserCachedRawStatusObservation struct {
	cachedAt    time.Time
	observation SharedSessionBrowserRawStatusObservation
}

type sharedSessionBrowserCachedRawProfilesObservation struct {
	cachedAt    time.Time
	observation SharedSessionBrowserRawProfilesObservation
}

type sharedSessionBrowserCachedRawRouteMutationObservation struct {
	cachedAt    time.Time
	observation SharedSessionBrowserRawRouteMutationObservation
}

type sharedSessionBrowserRouteMutationSourceKey struct {
	sessionID        string
	selectedInfo     BrowserRuntimeInfo
	requestedProfile string
}

type sharedSessionBrowserCachedEventCycleObservation struct {
	cachedAt    time.Time
	observation SharedSessionBrowserEventCycleObservation
}

type sharedSessionBrowserCachedBindingObservation struct {
	cachedAt    time.Time
	observation SharedSessionBrowserBindingObservation
}

type sharedSessionBrowserCachedViewObservation struct {
	cachedAt    time.Time
	observation SharedSessionBrowserViewObservation
}

type sharedSessionBrowserCachedWatchLoopObservation struct {
	cachedAt    time.Time
	observation SharedSessionBrowserWatchLoopObservation
}

type sharedSessionBrowserInFlightRawStatusObservation struct {
	ready       chan struct{}
	observation SharedSessionBrowserRawStatusObservation
}

type sharedSessionBrowserInFlightRawProfilesObservation struct {
	ready       chan struct{}
	observation SharedSessionBrowserRawProfilesObservation
}

type sharedSessionBrowserInFlightRawLifecycleObservation struct {
	ready       chan struct{}
	observation SharedSessionBrowserRawLifecycleObservation
}

type sharedSessionBrowserInFlightEventCycleObservation struct {
	ready       chan struct{}
	observation SharedSessionBrowserEventCycleObservation
}

type sharedSessionBrowserInFlightBindingObservation struct {
	ready       chan struct{}
	observation SharedSessionBrowserBindingObservation
}

type sharedSessionBrowserInFlightViewObservation struct {
	ready       chan struct{}
	observation SharedSessionBrowserViewObservation
}

type sharedSessionBrowserInFlightWatchLoopObservation struct {
	ready       chan struct{}
	observation SharedSessionBrowserWatchLoopObservation
}

var sharedSessionBrowserObserverManagerProviders = struct {
	mu        sync.RWMutex
	providers map[sharedSessionBrowserObserverManagerProviderCacheKey]*sharedSessionBrowserObserverManagerProviderEntry
}{
	providers: make(map[sharedSessionBrowserObserverManagerProviderCacheKey]*sharedSessionBrowserObserverManagerProviderEntry),
}

func sharedSessionBrowserObserverManager(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) SharedSessionBrowserObserverManager {
	key, ok := sharedSessionBrowserObserverManagerProviderCacheKeyForDependencies(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	)
	if !ok {
		return NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, reconnectWindow)
	}
	now := time.Now()
	sharedSessionBrowserObserverManagerProviders.mu.Lock()
	defer sharedSessionBrowserObserverManagerProviders.mu.Unlock()
	pruneIdleSharedSessionBrowserObserverManagerProvidersLocked(now)
	if entry, found := sharedSessionBrowserObserverManagerProviders.providers[key]; found {
		entry.touch(now)
		cached := entry.manager
		if cached.provider == nil {
			cached.provider = entry
			entry.manager = cached
		}
		return cached
	}
	entry := &sharedSessionBrowserObserverManagerProviderEntry{
		lastActiveAt: now,
	}
	manager := NewSharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, reconnectWindow)
	manager.provider = entry
	entry.manager = manager
	sharedSessionBrowserObserverManagerProviders.providers[key] = entry
	return manager
}

// SharedSessionBrowserObserverManagerFor returns a shared provider for the
// given session registries and reconnect window. Callers that do not need a
// bespoke provider should prefer this over constructing a fresh manager so
// per-control bound managers and short-lived source-time caches can be reused
// across entrypoints.
func SharedSessionBrowserObserverManagerFor(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) SharedSessionBrowserObserverManager {
	return sharedSessionBrowserObserverManager(sessionRegistry, runRegistry, stateRegistry, reconnectWindow)
}

// NewSharedSessionBrowserObserverManager constructs the explicit shared
// session observer/watchdog manager used by watch, binding, session-view, and
// inspection surfaces.
func NewSharedSessionBrowserObserverManager(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) SharedSessionBrowserObserverManager {
	return SharedSessionBrowserObserverManager{
		SessionRegistry: sessionRegistry,
		RunRegistry:     runRegistry,
		StateRegistry:   stateRegistry,
		ReconnectWindow: reconnectWindow,
		cache: &sharedSessionBrowserWatchManagerCache{
			managers: make(map[sharedSessionBrowserWatchManagerCacheKey]SharedSessionBrowserWatchManager),
		},
	}
}

func sharedSessionBrowserWatchManagerCacheKeyForControl(control BrowserRuntimeControlBackend) (sharedSessionBrowserWatchManagerCacheKey, bool) {
	return sharedSessionBrowserWatchManagerCacheKeyForValue(control)
}

func sharedSessionBrowserWatchManagerCacheKeyForValue(value any) (sharedSessionBrowserWatchManagerCacheKey, bool) {
	reflected := reflect.ValueOf(value)
	if !reflected.IsValid() || reflected.Kind() != reflect.Pointer || reflected.IsNil() {
		return sharedSessionBrowserWatchManagerCacheKey{}, false
	}
	return sharedSessionBrowserWatchManagerCacheKey{
		typ: reflected.Type(),
		ptr: reflected.Pointer(),
	}, true
}

func sharedSessionBrowserObserverManagerProviderCacheKeyForDependencies(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) (sharedSessionBrowserObserverManagerProviderCacheKey, bool) {
	key := sharedSessionBrowserObserverManagerProviderCacheKey{
		reconnectWindow: reconnectWindow,
	}
	if sessionRegistry != nil {
		cacheKey, ok := sharedSessionBrowserWatchManagerCacheKeyForValue(sessionRegistry)
		if !ok {
			return sharedSessionBrowserObserverManagerProviderCacheKey{}, false
		}
		key.sessionRegistry = cacheKey
	}
	if runRegistry != nil {
		cacheKey, ok := sharedSessionBrowserWatchManagerCacheKeyForValue(runRegistry)
		if !ok {
			return sharedSessionBrowserObserverManagerProviderCacheKey{}, false
		}
		key.runRegistry = cacheKey
	}
	if stateRegistry != nil {
		cacheKey, ok := sharedSessionBrowserWatchManagerCacheKeyForValue(stateRegistry)
		if !ok {
			return sharedSessionBrowserObserverManagerProviderCacheKey{}, false
		}
		key.stateRegistry = cacheKey
	}
	return key, true
}

// NewSharedSessionBrowserWatchManager constructs the explicit source-time
// watch manager that binds backend control together with the shared observer
// dependencies.
func NewSharedSessionBrowserWatchManager(
	control BrowserRuntimeControlBackend,
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) SharedSessionBrowserWatchManager {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).Bind(control)
}
