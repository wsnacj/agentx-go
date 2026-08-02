package browserruntime

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// BrowserSessionTarget is a session-scoped handle for a tracked browser tab.
type BrowserSessionTarget struct {
	ID         string
	TabIndex   int
	URL        string
	Title      string
	BrowserApp string
	Backend    string
	Profile    string
	Target     string
}

// BrowserSessionRoute scopes tab handles to a concrete browser runtime route.
type BrowserSessionRoute struct {
	Backend    string
	Profile    string
	Target     string
	BrowserApp string
}

// BrowserSessionRouteState is a read-only snapshot of tracked targets for a route.
type BrowserSessionRouteState struct {
	Route                    BrowserSessionRoute
	CurrentTargetID          string
	CurrentTargetSource      string
	PendingTargetReview      *BrowserSessionTargetReview
	PendingTargetReviewCount int
	Targets                  []BrowserSessionTarget
}

// BrowserSessionTargetReview is a session-scoped pending target review posture
// for a tracked route target that should not yet be auto-adopted.
type BrowserSessionTargetReview struct {
	ID         string
	TabIndex   int
	URL        string
	Title      string
	BrowserApp string
	Backend    string
	Profile    string
	Target     string
	Decision   string
	Reason     string
}

// BrowserSessionRegistry keeps lightweight browser target handles in memory.
type BrowserSessionRegistry struct {
	mu                              sync.RWMutex
	sessions                        map[string]*browserSessionState
	suppressWatchManagerInvalidates atomic.Int32
}

type browserSessionState struct {
	nextID                      uint64
	currentTargetByRoute        map[string]string
	currentTargetSourceByRoute  map[string]string
	pendingTargetReviewsByRoute map[string][]BrowserSessionTargetReview
	targets                     map[string]BrowserSessionTarget
	tabToTargetByRoute          map[string]map[int]string
}

func NewBrowserSessionRegistry() *BrowserSessionRegistry {
	return &BrowserSessionRegistry{sessions: map[string]*browserSessionState{}}
}

func (r *BrowserSessionRegistry) invalidateWatchManagers() {
	if r == nil || r.suppressWatchManagerInvalidates.Load() > 0 {
		return
	}
	invalidateSharedSessionBrowserObserverManagersForSessionRegistry(r)
}

func (r *BrowserSessionRegistry) withSuppressedWatchManagerInvalidation(fn func()) {
	if r == nil || fn == nil {
		return
	}
	r.suppressWatchManagerInvalidates.Add(1)
	defer r.suppressWatchManagerInvalidates.Add(-1)
	fn()
}

func (r *BrowserSessionRegistry) TrackTab(sessionID string, target BrowserSessionTarget, setCurrent bool) BrowserSessionTarget {
	sessionID = strings.TrimSpace(sessionID)
	target = normalizeBrowserSessionTarget(target)
	if sessionID == "" || target.TabIndex <= 0 {
		return target
	}
	r.mu.Lock()
	state := r.ensureSessionLocked(sessionID)
	tracked := state.trackTabLocked(target)
	if setCurrent {
		routeKey := browserSessionRouteKey(browserSessionRouteFromTarget(tracked))
		priorTargetID := strings.TrimSpace(state.currentTargetByRoute[routeKey])
		priorSource := strings.TrimSpace(state.currentTargetSourceByRoute[routeKey])
		state.currentTargetByRoute[routeKey] = tracked.ID
		if priorTargetID == strings.TrimSpace(tracked.ID) && priorSource != "" {
			state.currentTargetSourceByRoute[routeKey] = priorSource
		} else {
			state.currentTargetSourceByRoute[routeKey] = "tracked_active_tab"
		}
		state.clearPendingTargetReviewForRouteLocked(routeKey, strings.TrimSpace(tracked.ID))
	}
	r.mu.Unlock()
	r.invalidateWatchManagers()
	return tracked
}

func (r *BrowserSessionRegistry) TrackTabs(sessionID string, targets []BrowserSessionTarget, currentTabIndex int) []BrowserSessionTarget {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" || len(targets) == 0 {
		out := make([]BrowserSessionTarget, 0, len(targets))
		for _, item := range targets {
			out = append(out, normalizeBrowserSessionTarget(item))
		}
		return out
	}
	r.mu.Lock()
	state := r.ensureSessionLocked(sessionID)
	seenTabsByRoute := map[string]map[int]bool{}
	out := make([]BrowserSessionTarget, 0, len(targets))
	for _, item := range targets {
		item = normalizeBrowserSessionTarget(item)
		if item.TabIndex <= 0 {
			out = append(out, item)
			continue
		}
		tracked := state.trackTabLocked(item)
		routeKey := browserSessionRouteKey(browserSessionRouteFromTarget(tracked))
		if seenTabsByRoute[routeKey] == nil {
			seenTabsByRoute[routeKey] = map[int]bool{}
		}
		seenTabsByRoute[routeKey][tracked.TabIndex] = true
		if tracked.TabIndex == currentTabIndex {
			priorTargetID := strings.TrimSpace(state.currentTargetByRoute[routeKey])
			priorSource := strings.TrimSpace(state.currentTargetSourceByRoute[routeKey])
			state.currentTargetByRoute[routeKey] = tracked.ID
			if priorTargetID == strings.TrimSpace(tracked.ID) && priorSource != "" {
				state.currentTargetSourceByRoute[routeKey] = priorSource
			} else {
				state.currentTargetSourceByRoute[routeKey] = "tracked_active_tab"
			}
		}
		out = append(out, tracked)
	}
	for routeKey, seenTabs := range seenTabsByRoute {
		tabMap := state.tabToTargetByRoute[routeKey]
		for tabIndex, targetID := range tabMap {
			if seenTabs[tabIndex] {
				continue
			}
			delete(tabMap, tabIndex)
			delete(state.targets, targetID)
			if strings.TrimSpace(state.currentTargetByRoute[routeKey]) == targetID {
				delete(state.currentTargetByRoute, routeKey)
				delete(state.currentTargetSourceByRoute, routeKey)
			}
			state.removePendingTargetReviewForRouteTargetLocked(routeKey, targetID)
		}
		if len(tabMap) == 0 {
			delete(state.tabToTargetByRoute, routeKey)
		}
	}
	r.mu.Unlock()
	r.invalidateWatchManagers()
	return out
}

func (r *BrowserSessionRegistry) SyncTabsForRoute(sessionID string, route BrowserSessionRoute, targets []BrowserSessionTarget, currentTabIndex int) []BrowserSessionTarget {
	sessionID = strings.TrimSpace(sessionID)
	route = normalizeBrowserSessionRoute(route)
	if len(targets) == 0 {
		if sessionID != "" {
			r.ClearRoute(sessionID, route)
		}
		return nil
	}
	inputs := make([]BrowserSessionTarget, 0, len(targets))
	for _, item := range targets {
		item = normalizeBrowserSessionTarget(item)
		if item.BrowserApp == "" {
			item.BrowserApp = route.BrowserApp
		}
		if item.Backend == "" {
			item.Backend = route.Backend
		}
		if item.Profile == "" {
			item.Profile = route.Profile
		}
		if item.Target == "" {
			item.Target = route.Target
		}
		inputs = append(inputs, item)
	}
	return r.TrackTabs(sessionID, inputs, currentTabIndex)
}

func (r *BrowserSessionRegistry) TrackCurrentTarget(sessionID string, target BrowserSessionTarget, source ...string) BrowserSessionTarget {
	sessionID = strings.TrimSpace(sessionID)
	target = normalizeBrowserSessionTarget(target)
	if sessionID == "" {
		return target
	}
	r.mu.Lock()
	state := r.ensureSessionLocked(sessionID)
	tracked, routeKey := state.trackCurrentTargetLocked(target)
	if routeKey == "" || strings.TrimSpace(tracked.ID) == "" {
		r.mu.Unlock()
		return tracked
	}
	state.currentTargetByRoute[routeKey] = tracked.ID
	state.currentTargetSourceByRoute[routeKey] = firstBrowserSessionTrackedTargetSource(source...)
	state.clearPendingTargetReviewForRouteLocked(routeKey, strings.TrimSpace(tracked.ID))
	r.mu.Unlock()
	r.invalidateWatchManagers()
	return tracked
}

func (r *BrowserSessionRegistry) ResolveTarget(sessionID string, targetID string) (BrowserSessionTarget, bool) {
	sessionID = strings.TrimSpace(sessionID)
	targetID = strings.TrimSpace(targetID)
	if sessionID == "" || targetID == "" {
		return BrowserSessionTarget{}, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	state := r.sessions[sessionID]
	if state == nil {
		return BrowserSessionTarget{}, false
	}
	target, ok := state.targets[targetID]
	return target, ok
}

func (r *BrowserSessionRegistry) ResolveTab(sessionID string, tabIndex int) (BrowserSessionTarget, bool) {
	return r.ResolveTabForRoute(sessionID, BrowserSessionRoute{}, tabIndex)
}

func (r *BrowserSessionRegistry) CurrentTarget(sessionID string) (BrowserSessionTarget, bool) {
	return r.CurrentTargetForRoute(sessionID, BrowserSessionRoute{})
}

func (r *BrowserSessionRegistry) SelectTabForRoute(sessionID string, route BrowserSessionRoute, tabIndex int, source ...string) (BrowserSessionTarget, bool) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" || tabIndex <= 0 {
		return BrowserSessionTarget{}, false
	}
	r.mu.Lock()
	state := r.sessions[sessionID]
	if state == nil {
		r.mu.Unlock()
		return BrowserSessionTarget{}, false
	}
	target, routeKey, ok := state.resolveTabForRouteLocked(route, tabIndex)
	if !ok || routeKey == "" {
		r.mu.Unlock()
		return BrowserSessionTarget{}, false
	}
	state.currentTargetByRoute[routeKey] = strings.TrimSpace(target.ID)
	state.currentTargetSourceByRoute[routeKey] = firstBrowserSessionTargetSource(source...)
	state.clearPendingTargetReviewForRouteLocked(routeKey, strings.TrimSpace(target.ID))
	r.mu.Unlock()
	r.invalidateWatchManagers()
	return target, true
}

func (r *BrowserSessionRegistry) SelectTargetForRoute(sessionID string, route BrowserSessionRoute, targetID string, source ...string) (BrowserSessionTarget, bool) {
	sessionID = strings.TrimSpace(sessionID)
	targetID = strings.TrimSpace(targetID)
	if sessionID == "" || targetID == "" {
		return BrowserSessionTarget{}, false
	}
	r.mu.Lock()
	state := r.sessions[sessionID]
	if state == nil {
		r.mu.Unlock()
		return BrowserSessionTarget{}, false
	}
	target, routeKey, ok := state.resolveTargetForRouteLocked(route, targetID)
	if !ok || routeKey == "" {
		r.mu.Unlock()
		return BrowserSessionTarget{}, false
	}
	state.currentTargetByRoute[routeKey] = strings.TrimSpace(target.ID)
	state.currentTargetSourceByRoute[routeKey] = firstBrowserSessionTargetSource(source...)
	state.clearPendingTargetReviewForRouteLocked(routeKey, strings.TrimSpace(target.ID))
	r.mu.Unlock()
	r.invalidateWatchManagers()
	return target, true
}

func (r *BrowserSessionRegistry) restoreCurrentTargetSelectionAndRecordPendingTargetPopupReview(
	sessionID string,
	route BrowserSessionRoute,
	snapshot *BrowserSessionTargetSelection,
	activeTab BrowserTab,
	decision string,
	reason string,
	source string,
) (*BrowserSessionTargetSelection, *BrowserSessionTargetReview) {
	sessionID = strings.TrimSpace(sessionID)
	route = normalizeBrowserSessionRoute(route)
	activeTab = BrowserTab{
		Index:    activeTab.Index,
		Title:    strings.TrimSpace(activeTab.Title),
		URL:      strings.TrimSpace(activeTab.URL),
		TargetID: strings.TrimSpace(activeTab.TargetID),
		Active:   activeTab.Active,
	}
	source = firstNonEmptyString(strings.TrimSpace(source), "popup_review_restore")
	if sessionID == "" {
		return nil, nil
	}

	r.mu.Lock()
	state := r.ensureSessionLocked(sessionID)
	if state == nil {
		r.mu.Unlock()
		return nil, nil
	}

	selection, changed := state.restoreCurrentTargetSelectionForRouteLocked(route, snapshot, source)
	review := state.recordPendingTargetReviewForRouteLocked(route, BrowserSessionTargetReview{
		ID:         strings.TrimSpace(activeTab.TargetID),
		TabIndex:   activeTab.Index,
		URL:        strings.TrimSpace(activeTab.URL),
		Title:      strings.TrimSpace(activeTab.Title),
		BrowserApp: strings.TrimSpace(route.BrowserApp),
		Backend:    strings.TrimSpace(route.Backend),
		Profile:    strings.TrimSpace(route.Profile),
		Target:     strings.TrimSpace(route.Target),
		Decision:   strings.TrimSpace(decision),
		Reason:     strings.TrimSpace(reason),
	})
	if review != nil {
		changed = true
	}
	r.mu.Unlock()

	if changed {
		r.invalidateWatchManagers()
	}
	return selection, review
}

func firstBrowserSessionTargetSource(source ...string) string {
	for _, item := range source {
		value := strings.TrimSpace(item)
		if value != "" {
			return value
		}
	}
	return "select_target"
}

func firstBrowserSessionTrackedTargetSource(source ...string) string {
	for _, item := range source {
		value := strings.TrimSpace(item)
		if value != "" {
			return value
		}
	}
	return "tracked_current_target"
}

func (r *BrowserSessionRegistry) clearCurrentTargetForRoute(sessionID string, route BrowserSessionRoute, invalidate bool) bool {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return false
	}
	r.mu.Lock()
	state := r.sessions[sessionID]
	if state == nil {
		r.mu.Unlock()
		return false
	}
	_, routeKey, ok := state.currentTargetForRouteLocked(route)
	if !ok || routeKey == "" {
		r.mu.Unlock()
		return false
	}
	delete(state.currentTargetByRoute, routeKey)
	delete(state.currentTargetSourceByRoute, routeKey)
	r.mu.Unlock()
	if invalidate {
		r.invalidateWatchManagers()
	}
	return true
}

func (r *BrowserSessionRegistry) ClearCurrentTargetForRoute(sessionID string, route BrowserSessionRoute) bool {
	return r.clearCurrentTargetForRoute(sessionID, route, true)
}

func (r *BrowserSessionRegistry) RecordPendingTargetReviewForRoute(sessionID string, route BrowserSessionRoute, review BrowserSessionTargetReview) {
	sessionID = strings.TrimSpace(sessionID)
	review = normalizeBrowserSessionTargetReview(review)
	if sessionID == "" || strings.TrimSpace(review.ID) == "" {
		return
	}
	r.mu.Lock()
	state := r.ensureSessionLocked(sessionID)
	routeKey := browserSessionRouteKey(route)
	if routeKey == "__default__" {
		routeKey = browserSessionRouteKey(browserSessionRouteFromTarget(BrowserSessionTarget{
			ID:         review.ID,
			TabIndex:   review.TabIndex,
			URL:        review.URL,
			Title:      review.Title,
			BrowserApp: review.BrowserApp,
			Backend:    review.Backend,
			Profile:    review.Profile,
			Target:     review.Target,
		}))
	}
	if routeKey == "" {
		routeKey = "__default__"
	}
	state.appendPendingTargetReviewForRouteLocked(routeKey, review)
	r.mu.Unlock()
	r.invalidateWatchManagers()
}

func (r *BrowserSessionRegistry) ClearPendingTargetReviewForRoute(sessionID string, route BrowserSessionRoute) bool {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return false
	}
	r.mu.Lock()
	state := r.sessions[sessionID]
	if state == nil {
		r.mu.Unlock()
		return false
	}
	routeKey := browserSessionRouteKey(route)
	if routeKey == "" {
		routeKey = "__default__"
	}
	if len(state.pendingTargetReviewsByRoute[routeKey]) == 0 {
		r.mu.Unlock()
		return false
	}
	delete(state.pendingTargetReviewsByRoute, routeKey)
	r.mu.Unlock()
	r.invalidateWatchManagers()
	return true
}

func (r *BrowserSessionRegistry) ClearRoute(sessionID string, route BrowserSessionRoute) int {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return 0
	}
	filter := normalizeBrowserSessionRoute(route)
	r.mu.Lock()
	state := r.sessions[sessionID]
	if state == nil {
		r.mu.Unlock()
		return 0
	}
	clearedTargetIDs := map[string]bool{}
	routeKeys := map[string]bool{}
	for routeKey := range state.tabToTargetByRoute {
		routeKeys[routeKey] = true
	}
	for routeKey := range state.currentTargetByRoute {
		routeKeys[routeKey] = true
	}
	for routeKey := range state.pendingTargetReviewsByRoute {
		routeKeys[routeKey] = true
	}
	for routeKey := range routeKeys {
		routeState := browserSessionRouteFromKey(routeKey)
		if !browserSessionRouteMatchesFilter(routeState, filter) {
			continue
		}
		if tabMap := state.tabToTargetByRoute[routeKey]; len(tabMap) > 0 {
			for _, targetID := range tabMap {
				targetID = strings.TrimSpace(targetID)
				if targetID != "" {
					clearedTargetIDs[targetID] = true
				}
			}
			delete(state.tabToTargetByRoute, routeKey)
		}
		if targetID := strings.TrimSpace(state.currentTargetByRoute[routeKey]); targetID != "" {
			clearedTargetIDs[targetID] = true
		}
		delete(state.currentTargetByRoute, routeKey)
		delete(state.currentTargetSourceByRoute, routeKey)
		delete(state.pendingTargetReviewsByRoute, routeKey)
	}
	if len(clearedTargetIDs) == 0 {
		r.mu.Unlock()
		return 0
	}
	state.pruneUnusedTargetsLocked()
	if len(state.targets) == 0 && len(state.tabToTargetByRoute) == 0 && len(state.pendingTargetReviewsByRoute) == 0 {
		delete(r.sessions, sessionID)
	}
	r.mu.Unlock()
	r.invalidateWatchManagers()
	return len(clearedTargetIDs)
}

func (r *BrowserSessionRegistry) PruneStaleRouteState(sessionID string, route BrowserSessionRoute) bool {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return false
	}
	filter := normalizeBrowserSessionRoute(route)
	r.mu.Lock()
	state := r.sessions[sessionID]
	if state == nil {
		r.mu.Unlock()
		return false
	}
	changed := false
	routeKeys := map[string]bool{}
	for routeKey := range state.tabToTargetByRoute {
		routeKeys[routeKey] = true
	}
	for routeKey := range state.currentTargetByRoute {
		routeKeys[routeKey] = true
	}
	for routeKey := range state.pendingTargetReviewsByRoute {
		routeKeys[routeKey] = true
	}
	for routeKey := range routeKeys {
		routeState := browserSessionRouteFromKey(routeKey)
		if !browserSessionRouteMatchesFilter(routeState, filter) {
			continue
		}
		tabMap := state.tabToTargetByRoute[routeKey]
		for tabIndex, targetID := range tabMap {
			targetID = strings.TrimSpace(targetID)
			if targetID == "" {
				delete(tabMap, tabIndex)
				changed = true
				continue
			}
			target, ok := state.targets[targetID]
			if !ok || !browserSessionSameLogicalRoute(browserSessionRouteFromTarget(target), routeState) {
				delete(tabMap, tabIndex)
				changed = true
			}
		}
		if len(tabMap) == 0 {
			delete(state.tabToTargetByRoute, routeKey)
		}
		if currentTargetID := strings.TrimSpace(state.currentTargetByRoute[routeKey]); currentTargetID != "" {
			target, ok := state.targets[currentTargetID]
			if !ok || !browserSessionSameLogicalRoute(browserSessionRouteFromTarget(target), routeState) {
				delete(state.currentTargetByRoute, routeKey)
				delete(state.currentTargetSourceByRoute, routeKey)
				changed = true
			}
		}
		reviews := state.pendingTargetReviewsByRoute[routeKey]
		if len(reviews) > 0 {
			filtered := reviews[:0]
			for _, review := range reviews {
				targetID := strings.TrimSpace(review.ID)
				target, ok := state.targets[targetID]
				if targetID == "" || !ok || !browserSessionSameLogicalRoute(browserSessionRouteFromTarget(target), routeState) {
					changed = true
					continue
				}
				filtered = append(filtered, review)
			}
			if len(filtered) == 0 {
				delete(state.pendingTargetReviewsByRoute, routeKey)
			} else {
				state.pendingTargetReviewsByRoute[routeKey] = append([]BrowserSessionTargetReview(nil), filtered...)
			}
		}
	}
	if changed {
		state.pruneUnusedTargetsLocked()
		if len(state.targets) == 0 && len(state.tabToTargetByRoute) == 0 && len(state.pendingTargetReviewsByRoute) == 0 {
			delete(r.sessions, sessionID)
		}
	}
	r.mu.Unlock()
	if changed {
		r.invalidateWatchManagers()
	}
	return changed
}

func (r *BrowserSessionRegistry) ForgetTab(sessionID string, tabIndex int) {
	r.ForgetTabForRoute(sessionID, BrowserSessionRoute{}, tabIndex)
}

func (r *BrowserSessionRegistry) Snapshot(sessionID string) []BrowserSessionRouteState {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	state := r.sessions[sessionID]
	if state == nil {
		return nil
	}
	return state.snapshotLocked()
}

func (r *BrowserSessionRegistry) ensureSessionLocked(sessionID string) *browserSessionState {
	if state := r.sessions[sessionID]; state != nil {
		return state
	}
	state := &browserSessionState{
		currentTargetByRoute:        map[string]string{},
		currentTargetSourceByRoute:  map[string]string{},
		pendingTargetReviewsByRoute: map[string][]BrowserSessionTargetReview{},
		targets:                     map[string]BrowserSessionTarget{},
		tabToTargetByRoute:          map[string]map[int]string{},
	}
	r.sessions[sessionID] = state
	return state
}

func (s *browserSessionState) trackTabLocked(target BrowserSessionTarget) BrowserSessionTarget {
	target = normalizeBrowserSessionTarget(target)
	if target.TabIndex <= 0 {
		return target
	}
	route := browserSessionRouteFromTarget(target)
	routeKey := browserSessionRouteKey(route)
	targetID := ""
	if routeKey == "__default__" {
		if matchedRouteKey, matchedTargetID, ok := s.findUniqueTabTargetLocked(target.TabIndex); ok {
			routeKey = matchedRouteKey
			targetID = matchedTargetID
		}
	} else {
		if matchedRouteKey, matchedTargetID, ok := s.findEquivalentTabTargetLocked(route, target.TabIndex); ok {
			routeKey = matchedRouteKey
			targetID = matchedTargetID
		}
	}
	tabToTarget := s.ensureRouteTabMapLocked(routeKey)
	if targetID == "" {
		targetID = strings.TrimSpace(tabToTarget[target.TabIndex])
	}
	if targetID == "" {
		s.nextID++
		targetID = fmt.Sprintf("brw-%d", s.nextID)
	}
	current := s.targets[targetID]
	if target.URL == "" {
		target.URL = current.URL
	}
	if target.Title == "" {
		target.Title = current.Title
	}
	if target.BrowserApp == "" {
		target.BrowserApp = current.BrowserApp
	}
	if target.Backend == "" {
		target.Backend = current.Backend
	}
	if target.Profile == "" {
		target.Profile = current.Profile
	}
	if target.Target == "" {
		target.Target = current.Target
	}
	target.ID = targetID
	s.targets[targetID] = target
	tabToTarget[target.TabIndex] = targetID
	return target
}

func (s *browserSessionState) trackCurrentTargetLocked(target BrowserSessionTarget) (BrowserSessionTarget, string) {
	target = normalizeBrowserSessionTarget(target)
	if target.TabIndex > 0 {
		tracked := s.trackTabLocked(target)
		return tracked, browserSessionRouteKey(browserSessionRouteFromTarget(tracked))
	}
	route := browserSessionRouteFromTarget(target)
	routeKey := browserSessionRouteKey(route)
	targetID := ""
	if currentTargetID := strings.TrimSpace(s.currentTargetByRoute[routeKey]); currentTargetID != "" {
		if current, ok := s.targets[currentTargetID]; ok && current.TabIndex <= 0 {
			targetID = currentTargetID
		}
	}
	if targetID == "" {
		if matchedRouteKey, matchedTargetID, ok := s.findEquivalentCurrentTargetLocked(route); ok {
			routeKey = matchedRouteKey
			targetID = matchedTargetID
		}
	}
	if targetID == "" {
		s.nextID++
		targetID = fmt.Sprintf("brw-%d", s.nextID)
	}
	current := s.targets[targetID]
	if target.URL == "" {
		target.URL = current.URL
	}
	if target.Title == "" {
		target.Title = current.Title
	}
	if target.BrowserApp == "" {
		target.BrowserApp = current.BrowserApp
	}
	if target.Backend == "" {
		target.Backend = current.Backend
	}
	if target.Profile == "" {
		target.Profile = current.Profile
	}
	if target.Target == "" {
		target.Target = current.Target
	}
	target.ID = targetID
	s.targets[targetID] = target
	return target, routeKey
}

func normalizeBrowserSessionTarget(target BrowserSessionTarget) BrowserSessionTarget {
	if target.TabIndex < 0 {
		target.TabIndex = 0
	}
	return BrowserSessionTarget{
		ID:         strings.TrimSpace(target.ID),
		TabIndex:   target.TabIndex,
		URL:        strings.TrimSpace(target.URL),
		Title:      strings.TrimSpace(target.Title),
		BrowserApp: strings.TrimSpace(target.BrowserApp),
		Backend:    strings.TrimSpace(target.Backend),
		Profile:    strings.TrimSpace(target.Profile),
		Target:     strings.TrimSpace(target.Target),
	}
}

func normalizeBrowserSessionTargetReview(review BrowserSessionTargetReview) BrowserSessionTargetReview {
	if review.TabIndex < 0 {
		review.TabIndex = 0
	}
	return BrowserSessionTargetReview{
		ID:         strings.TrimSpace(review.ID),
		TabIndex:   review.TabIndex,
		URL:        strings.TrimSpace(review.URL),
		Title:      strings.TrimSpace(review.Title),
		BrowserApp: strings.TrimSpace(review.BrowserApp),
		Backend:    strings.TrimSpace(review.Backend),
		Profile:    strings.TrimSpace(review.Profile),
		Target:     strings.TrimSpace(review.Target),
		Decision:   strings.TrimSpace(review.Decision),
		Reason:     strings.TrimSpace(review.Reason),
	}
}

func browserSessionRouteFromTarget(target BrowserSessionTarget) BrowserSessionRoute {
	target = normalizeBrowserSessionTarget(target)
	return BrowserSessionRoute{
		Backend:    target.Backend,
		Profile:    target.Profile,
		Target:     target.Target,
		BrowserApp: target.BrowserApp,
	}
}

func normalizeBrowserSessionRoute(route BrowserSessionRoute) BrowserSessionRoute {
	return BrowserSessionRoute{
		Backend:    strings.ToLower(strings.TrimSpace(route.Backend)),
		Profile:    strings.ToLower(strings.TrimSpace(route.Profile)),
		Target:     strings.ToLower(strings.TrimSpace(route.Target)),
		BrowserApp: strings.ToLower(strings.TrimSpace(route.BrowserApp)),
	}
}

func browserSessionRouteKey(route BrowserSessionRoute) string {
	route = normalizeBrowserSessionRoute(route)
	if route.Backend == "" && route.Profile == "" && route.Target == "" && route.BrowserApp == "" {
		return "__default__"
	}
	return strings.Join([]string{route.Backend, route.Profile, route.Target, route.BrowserApp}, "\x00")
}

func (s *browserSessionState) ensureRouteTabMapLocked(routeKey string) map[int]string {
	if routeKey == "" {
		routeKey = "__default__"
	}
	if s.tabToTargetByRoute == nil {
		s.tabToTargetByRoute = map[string]map[int]string{}
	}
	if s.tabToTargetByRoute[routeKey] == nil {
		s.tabToTargetByRoute[routeKey] = map[int]string{}
	}
	return s.tabToTargetByRoute[routeKey]
}

func (s *browserSessionState) findUniqueTabTargetLocked(tabIndex int) (string, string, bool) {
	matchedRouteKey := ""
	matchedTargetID := ""
	for routeKey, tabMap := range s.tabToTargetByRoute {
		targetID := strings.TrimSpace(tabMap[tabIndex])
		if targetID == "" {
			continue
		}
		if matchedTargetID != "" && matchedTargetID != targetID {
			return "", "", false
		}
		if matchedTargetID == "" {
			matchedRouteKey = routeKey
			matchedTargetID = targetID
		}
	}
	if matchedTargetID == "" {
		return "", "", false
	}
	return matchedRouteKey, matchedTargetID, true
}

func (s *browserSessionState) findEquivalentTabTargetLocked(route BrowserSessionRoute, tabIndex int) (string, string, bool) {
	if tabIndex <= 0 {
		return "", "", false
	}
	matchedRouteKey := ""
	matchedTargetID := ""
	for candidateRouteKey, tabMap := range s.tabToTargetByRoute {
		targetID := strings.TrimSpace(tabMap[tabIndex])
		if targetID == "" {
			continue
		}
		if !browserSessionSameLogicalRoute(browserSessionRouteFromKey(candidateRouteKey), route) {
			continue
		}
		if matchedTargetID != "" && matchedTargetID != targetID {
			return "", "", false
		}
		if matchedTargetID == "" {
			matchedRouteKey = candidateRouteKey
			matchedTargetID = targetID
		}
	}
	if matchedTargetID == "" {
		return "", "", false
	}
	return matchedRouteKey, matchedTargetID, true
}

func (s *browserSessionState) findEquivalentCurrentTargetLocked(route BrowserSessionRoute) (string, string, bool) {
	matchedRouteKey := ""
	matchedTargetID := ""
	for candidateRouteKey, targetID := range s.currentTargetByRoute {
		targetID = strings.TrimSpace(targetID)
		if targetID == "" {
			continue
		}
		target, ok := s.targets[targetID]
		if !ok || target.TabIndex > 0 {
			continue
		}
		if !browserSessionSameLogicalRoute(browserSessionRouteFromKey(candidateRouteKey), route) {
			continue
		}
		if matchedTargetID != "" && matchedTargetID != targetID {
			return "", "", false
		}
		if matchedTargetID == "" {
			matchedRouteKey = candidateRouteKey
			matchedTargetID = targetID
		}
	}
	if matchedTargetID == "" {
		return "", "", false
	}
	return matchedRouteKey, matchedTargetID, true
}

func (s *browserSessionState) resolveTargetForRouteLocked(route BrowserSessionRoute, targetID string) (BrowserSessionTarget, string, bool) {
	targetID = strings.TrimSpace(targetID)
	if targetID == "" {
		return BrowserSessionTarget{}, "", false
	}
	target, ok := s.targets[targetID]
	if !ok {
		return BrowserSessionTarget{}, "", false
	}
	targetRoute := browserSessionRouteFromTarget(target)
	if !browserSessionRouteMatchesFilter(targetRoute, route) {
		return BrowserSessionTarget{}, "", false
	}
	return target, browserSessionRouteKey(targetRoute), true
}

func (s *browserSessionState) resolveTabForRouteLocked(route BrowserSessionRoute, tabIndex int) (BrowserSessionTarget, string, bool) {
	if tabIndex <= 0 {
		return BrowserSessionTarget{}, "", false
	}
	routeKey := browserSessionRouteKey(route)
	if targetID := strings.TrimSpace(s.tabToTargetByRoute[routeKey][tabIndex]); targetID != "" {
		target, ok := s.targets[targetID]
		if ok && browserSessionTargetMatchesRouteKey(target, routeKey) {
			return target, routeKey, true
		}
	}
	var (
		matched     BrowserSessionTarget
		matchedKey  string
		foundTarget bool
	)
	for candidateRouteKey, tabMap := range s.tabToTargetByRoute {
		targetID := strings.TrimSpace(tabMap[tabIndex])
		if targetID == "" {
			continue
		}
		target, ok := s.targets[targetID]
		if !ok {
			continue
		}
		if !browserSessionRouteMatchesFilter(browserSessionRouteFromTarget(target), route) {
			continue
		}
		if foundTarget {
			if strings.EqualFold(strings.TrimSpace(matched.ID), strings.TrimSpace(target.ID)) {
				continue
			}
			return BrowserSessionTarget{}, "", false
		}
		matched = target
		matchedKey = candidateRouteKey
		foundTarget = true
	}
	return matched, matchedKey, foundTarget
}

func (s *browserSessionState) currentTargetForRouteLocked(route BrowserSessionRoute) (BrowserSessionTarget, string, bool) {
	routeKey := browserSessionRouteKey(route)
	if targetID := strings.TrimSpace(s.currentTargetByRoute[routeKey]); targetID != "" {
		target, ok := s.targets[targetID]
		if ok && browserSessionTargetMatchesRouteKey(target, routeKey) {
			return target, routeKey, true
		}
	}
	var (
		matched     BrowserSessionTarget
		matchedKey  string
		foundTarget bool
	)
	for candidateRouteKey, targetID := range s.currentTargetByRoute {
		targetID = strings.TrimSpace(targetID)
		if targetID == "" {
			continue
		}
		target, ok := s.targets[targetID]
		if !ok {
			continue
		}
		if !browserSessionRouteMatchesFilter(browserSessionRouteFromTarget(target), route) {
			continue
		}
		if foundTarget {
			if strings.EqualFold(strings.TrimSpace(matched.ID), strings.TrimSpace(target.ID)) {
				continue
			}
			return BrowserSessionTarget{}, "", false
		}
		matched = target
		matchedKey = candidateRouteKey
		foundTarget = true
	}
	return matched, matchedKey, foundTarget
}

func (s *browserSessionState) clearCurrentTargetForRouteLocked(route BrowserSessionRoute) bool {
	target, routeKey, ok := s.currentTargetForRouteLocked(route)
	if !ok || routeKey == "" || strings.TrimSpace(target.ID) == "" {
		return false
	}
	delete(s.currentTargetByRoute, routeKey)
	delete(s.currentTargetSourceByRoute, routeKey)
	return true
}

func (s *browserSessionState) selectTargetForRouteLocked(route BrowserSessionRoute, targetID string, source string) (*BrowserSessionTargetSelection, bool) {
	target, routeKey, ok := s.resolveTargetForRouteLocked(route, targetID)
	if !ok || routeKey == "" {
		return nil, false
	}
	source = firstBrowserSessionTargetSource(source)
	s.currentTargetByRoute[routeKey] = strings.TrimSpace(target.ID)
	s.currentTargetSourceByRoute[routeKey] = source
	s.clearPendingTargetReviewForRouteLocked(routeKey, strings.TrimSpace(target.ID))
	return sharedSessionBrowserTargetSelectionFromTracked(target, source), true
}

func (s *browserSessionState) restoreCurrentTargetSelectionForRouteLocked(route BrowserSessionRoute, snapshot *BrowserSessionTargetSelection, source string) (*BrowserSessionTargetSelection, bool) {
	route = normalizeBrowserSessionRoute(route)
	if snapshot == nil || strings.TrimSpace(snapshot.ID) == "" {
		return nil, s.clearCurrentTargetForRouteLocked(route)
	}
	source = firstNonEmptyString(strings.TrimSpace(snapshot.Source), strings.TrimSpace(source), "popup_review_restore")
	if selection, ok := s.selectTargetForRouteLocked(route, strings.TrimSpace(snapshot.ID), source); ok {
		return selection, true
	}
	if selection, ok := s.selectTargetForRouteLocked(BrowserSessionRoute{}, strings.TrimSpace(snapshot.ID), source); ok {
		return selection, true
	}
	return nil, s.clearCurrentTargetForRouteLocked(route)
}

func browserSessionTargetMatchesRouteKey(target BrowserSessionTarget, routeKey string) bool {
	routeKey = strings.TrimSpace(routeKey)
	if routeKey == "" {
		return false
	}
	return browserSessionSameLogicalRoute(browserSessionRouteFromTarget(target), browserSessionRouteFromKey(routeKey))
}

func (s *browserSessionState) pruneUnusedTargetsLocked() {
	if s == nil {
		return
	}
	used := map[string]bool{}
	for _, tabMap := range s.tabToTargetByRoute {
		for _, targetID := range tabMap {
			targetID = strings.TrimSpace(targetID)
			if targetID != "" {
				used[targetID] = true
			}
		}
	}
	for _, targetID := range s.currentTargetByRoute {
		targetID = strings.TrimSpace(targetID)
		if targetID != "" {
			used[targetID] = true
		}
	}
	for _, reviews := range s.pendingTargetReviewsByRoute {
		for _, review := range reviews {
			targetID := strings.TrimSpace(review.ID)
			if targetID != "" {
				used[targetID] = true
			}
		}
	}
	for targetID := range s.targets {
		if used[targetID] {
			continue
		}
		delete(s.targets, targetID)
	}
}

func browserSessionRouteMatchesFilter(candidate BrowserSessionRoute, filter BrowserSessionRoute) bool {
	candidate = normalizeBrowserSessionRoute(candidate)
	filter = normalizeBrowserSessionRoute(filter)
	if filter.Backend != "" && !browserSessionBackendMatches(candidate.Backend, filter.Backend) {
		return false
	}
	if filter.Profile != "" && candidate.Profile != filter.Profile {
		return false
	}
	if filter.Target != "" && candidate.Target != filter.Target {
		return false
	}
	if filter.BrowserApp != "" && candidate.BrowserApp != filter.BrowserApp {
		return false
	}
	return true
}

func browserSessionSameLogicalRoute(left BrowserSessionRoute, right BrowserSessionRoute) bool {
	left = normalizeBrowserSessionRoute(left)
	right = normalizeBrowserSessionRoute(right)
	return browserSessionCanonicalBackend(left.Backend) == browserSessionCanonicalBackend(right.Backend) &&
		left.Profile == right.Profile &&
		left.Target == right.Target &&
		left.BrowserApp == right.BrowserApp
}

func browserSessionBackendMatches(candidate string, filter string) bool {
	candidate = browserSessionCanonicalBackend(candidate)
	filter = browserSessionCanonicalBackend(filter)
	if candidate == "" || filter == "" {
		return candidate == filter
	}
	return candidate == filter
}

func browserSessionCanonicalBackend(backend string) string {
	backend = strings.ToLower(strings.TrimSpace(backend))
	switch {
	case strings.HasPrefix(backend, "proxy-"):
		return "proxy"
	case strings.HasPrefix(backend, "proxy_"):
		return "proxy"
	case strings.HasPrefix(backend, "system-"):
		return "system"
	case strings.HasPrefix(backend, "system_"):
		return "system"
	case strings.HasPrefix(backend, "safari_"):
		return "system"
	case strings.HasPrefix(backend, "http_"):
		return "system"
	case strings.HasPrefix(backend, "sandbox-"):
		return "sandbox"
	case strings.HasPrefix(backend, "sandbox_"):
		return "sandbox"
	case strings.HasPrefix(backend, "custom-"):
		return "custom"
	case strings.HasPrefix(backend, "custom_"):
		return "custom"
	default:
		return backend
	}
}

func (r *BrowserSessionRegistry) ResolveTabForRoute(sessionID string, route BrowserSessionRoute, tabIndex int) (BrowserSessionTarget, bool) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" || tabIndex <= 0 {
		return BrowserSessionTarget{}, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	state := r.sessions[sessionID]
	if state == nil {
		return BrowserSessionTarget{}, false
	}
	target, _, ok := state.resolveTabForRouteLocked(route, tabIndex)
	return target, ok
}

func (r *BrowserSessionRegistry) CurrentTargetForRoute(sessionID string, route BrowserSessionRoute) (BrowserSessionTarget, bool) {
	target, _, ok := r.CurrentTargetSelectionForRoute(sessionID, route)
	return target, ok
}

func (r *BrowserSessionRegistry) CurrentTargetSelectionForRoute(sessionID string, route BrowserSessionRoute) (BrowserSessionTarget, string, bool) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return BrowserSessionTarget{}, "", false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	state := r.sessions[sessionID]
	if state == nil {
		return BrowserSessionTarget{}, "", false
	}
	target, routeKey, ok := state.currentTargetForRouteLocked(route)
	if !ok {
		return BrowserSessionTarget{}, "", false
	}
	return target, strings.TrimSpace(state.currentTargetSourceByRoute[routeKey]), true
}

func (r *BrowserSessionRegistry) ForgetTabForRoute(sessionID string, route BrowserSessionRoute, tabIndex int) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" || tabIndex <= 0 {
		return
	}
	r.mu.Lock()
	state := r.sessions[sessionID]
	if state == nil {
		r.mu.Unlock()
		return
	}
	target, routeKey, ok := state.resolveTabForRouteLocked(route, tabIndex)
	if !ok || routeKey == "" {
		r.mu.Unlock()
		return
	}
	tabMap := state.tabToTargetByRoute[routeKey]
	if tabMap == nil {
		r.mu.Unlock()
		return
	}
	targetID := strings.TrimSpace(target.ID)
	delete(tabMap, tabIndex)
	if len(tabMap) == 0 {
		delete(state.tabToTargetByRoute, routeKey)
	}
	delete(state.targets, targetID)
	if strings.TrimSpace(state.currentTargetByRoute[routeKey]) == targetID {
		delete(state.currentTargetByRoute, routeKey)
		delete(state.currentTargetSourceByRoute, routeKey)
	}
	state.removePendingTargetReviewForRouteTargetLocked(routeKey, targetID)
	if len(state.targets) == 0 && len(state.tabToTargetByRoute) == 0 && len(state.pendingTargetReviewsByRoute) == 0 {
		delete(r.sessions, sessionID)
	}
	r.mu.Unlock()
	r.invalidateWatchManagers()
}

func (s *browserSessionState) snapshotLocked() []BrowserSessionRouteState {
	if s == nil {
		return nil
	}
	routeKeys := map[string]bool{}
	for routeKey := range s.tabToTargetByRoute {
		routeKeys[routeKey] = true
	}
	for routeKey, targetID := range s.currentTargetByRoute {
		if strings.TrimSpace(targetID) != "" {
			routeKeys[routeKey] = true
		}
	}
	for routeKey, reviews := range s.pendingTargetReviewsByRoute {
		if len(reviews) > 0 {
			routeKeys[routeKey] = true
		}
	}
	keys := make([]string, 0, len(routeKeys))
	for routeKey := range routeKeys {
		keys = append(keys, routeKey)
	}
	sort.Strings(keys)
	out := make([]BrowserSessionRouteState, 0, len(keys))
	for _, routeKey := range keys {
		state := BrowserSessionRouteState{
			Route:                    browserSessionRouteFromKey(routeKey),
			CurrentTargetID:          strings.TrimSpace(s.currentTargetByRoute[routeKey]),
			CurrentTargetSource:      strings.TrimSpace(s.currentTargetSourceByRoute[routeKey]),
			PendingTargetReviewCount: len(s.pendingTargetReviewsByRoute[routeKey]),
		}
		if review, ok := s.pendingTargetReviewForRouteLocked(routeKey); ok && strings.TrimSpace(review.ID) != "" {
			review = normalizeBrowserSessionTargetReview(review)
			state.PendingTargetReview = &review
		}
		targetIDs := map[string]bool{}
		tabIndexes := make([]int, 0, len(s.tabToTargetByRoute[routeKey]))
		for tabIndex := range s.tabToTargetByRoute[routeKey] {
			tabIndexes = append(tabIndexes, tabIndex)
		}
		sort.Ints(tabIndexes)
		for _, tabIndex := range tabIndexes {
			targetID := strings.TrimSpace(s.tabToTargetByRoute[routeKey][tabIndex])
			if targetID == "" || targetIDs[targetID] {
				continue
			}
			target, ok := s.targets[targetID]
			if !ok {
				continue
			}
			state.Targets = append(state.Targets, target)
			targetIDs[targetID] = true
		}
		if currentTargetID := strings.TrimSpace(state.CurrentTargetID); currentTargetID != "" && !targetIDs[currentTargetID] {
			if target, ok := s.targets[currentTargetID]; ok {
				state.Targets = append(state.Targets, target)
			}
		}
		if state.PendingTargetReview != nil && !targetIDs[strings.TrimSpace(state.PendingTargetReview.ID)] {
			if target, ok := s.targets[strings.TrimSpace(state.PendingTargetReview.ID)]; ok {
				state.Targets = append(state.Targets, target)
			}
		}
		if len(state.Targets) == 0 {
			continue
		}
		out = append(out, state)
	}
	return out
}

func (s *browserSessionState) clearPendingTargetReviewForRouteLocked(routeKey string, targetID string) {
	routeKey = strings.TrimSpace(routeKey)
	targetID = strings.TrimSpace(targetID)
	if routeKey == "" || targetID == "" || s == nil || s.pendingTargetReviewsByRoute == nil {
		return
	}
	s.removePendingTargetReviewForRouteTargetLocked(routeKey, targetID)
}

func (s *browserSessionState) pendingTargetReviewForRouteLocked(routeKey string) (BrowserSessionTargetReview, bool) {
	if s == nil || s.pendingTargetReviewsByRoute == nil {
		return BrowserSessionTargetReview{}, false
	}
	reviews := s.pendingTargetReviewsByRoute[strings.TrimSpace(routeKey)]
	if len(reviews) == 0 {
		return BrowserSessionTargetReview{}, false
	}
	return reviews[len(reviews)-1], true
}

func (s *browserSessionState) appendPendingTargetReviewForRouteLocked(routeKey string, review BrowserSessionTargetReview) {
	if s == nil {
		return
	}
	routeKey = strings.TrimSpace(routeKey)
	review = normalizeBrowserSessionTargetReview(review)
	if routeKey == "" || strings.TrimSpace(review.ID) == "" {
		return
	}
	if s.pendingTargetReviewsByRoute == nil {
		s.pendingTargetReviewsByRoute = map[string][]BrowserSessionTargetReview{}
	}
	reviews := append([]BrowserSessionTargetReview(nil), s.pendingTargetReviewsByRoute[routeKey]...)
	filtered := reviews[:0]
	for _, current := range reviews {
		if strings.EqualFold(strings.TrimSpace(current.ID), strings.TrimSpace(review.ID)) {
			continue
		}
		filtered = append(filtered, current)
	}
	filtered = append(filtered, review)
	s.pendingTargetReviewsByRoute[routeKey] = filtered
}

func (s *browserSessionState) recordPendingTargetReviewForRouteLocked(route BrowserSessionRoute, review BrowserSessionTargetReview) *BrowserSessionTargetReview {
	if s == nil {
		return nil
	}
	route = normalizeBrowserSessionRoute(route)
	review = normalizeBrowserSessionTargetReview(review)
	if strings.TrimSpace(review.ID) == "" {
		target, _, ok := s.resolveTabForRouteLocked(route, review.TabIndex)
		if !ok || strings.TrimSpace(target.ID) == "" {
			return nil
		}
		review.ID = strings.TrimSpace(target.ID)
	}
	routeKey := browserSessionRouteKey(route)
	if routeKey == "__default__" {
		routeKey = browserSessionRouteKey(browserSessionRouteFromTarget(BrowserSessionTarget{
			ID:         review.ID,
			TabIndex:   review.TabIndex,
			URL:        review.URL,
			Title:      review.Title,
			BrowserApp: review.BrowserApp,
			Backend:    review.Backend,
			Profile:    review.Profile,
			Target:     review.Target,
		}))
	}
	if routeKey == "" {
		routeKey = "__default__"
	}
	s.appendPendingTargetReviewForRouteLocked(routeKey, review)
	return &review
}

func (s *browserSessionState) removePendingTargetReviewForRouteTargetLocked(routeKey string, targetID string) {
	if s == nil || s.pendingTargetReviewsByRoute == nil {
		return
	}
	routeKey = strings.TrimSpace(routeKey)
	targetID = strings.TrimSpace(targetID)
	if routeKey == "" || targetID == "" {
		return
	}
	reviews := s.pendingTargetReviewsByRoute[routeKey]
	if len(reviews) == 0 {
		return
	}
	filtered := reviews[:0]
	for _, review := range reviews {
		if strings.EqualFold(strings.TrimSpace(review.ID), targetID) {
			continue
		}
		filtered = append(filtered, review)
	}
	if len(filtered) == 0 {
		delete(s.pendingTargetReviewsByRoute, routeKey)
		return
	}
	s.pendingTargetReviewsByRoute[routeKey] = append([]BrowserSessionTargetReview(nil), filtered...)
}

func browserSessionRouteFromKey(routeKey string) BrowserSessionRoute {
	if routeKey == "" || routeKey == "__default__" {
		return BrowserSessionRoute{}
	}
	parts := strings.Split(routeKey, "\x00")
	for len(parts) < 4 {
		parts = append(parts, "")
	}
	return BrowserSessionRoute{
		Backend:    parts[0],
		Profile:    parts[1],
		Target:     parts[2],
		BrowserApp: parts[3],
	}
}
