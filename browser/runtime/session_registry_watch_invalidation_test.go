package browserruntime

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestBrowserSessionRegistryTrackTabInvalidatesSharedWatchManagerCaches(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "isolated",
			Status:    "running",
			Running:   true,
			Connected: true,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	manager := SharedSessionBrowserObserverManagerFor(registry, nil, stateRegistry, time.Minute)
	sessionID := "sess-registry-track-tab-invalidates-watch"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	req := SharedSessionBrowserObserverRequest{
		SessionID:                   sessionID,
		SelectedInfo:                BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:                route,
		RequestedProfile:            "isolated",
		IncludeStatus:               true,
		IncludeProfiles:             true,
		IncludeSessionView:          true,
		SessionViewInfo:             BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		SessionViewRouteFilter:      route,
		SessionViewRequestedProfile: "isolated",
	}

	registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	first := manager.ObserveWatchLoop(context.Background(), backend, req)
	if len(first.View.Session.Routes) != 1 || len(first.View.Session.Routes[0].Targets) != 1 || first.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected initial watch loop to expose first tracked tab, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	second := manager.ObserveWatchLoop(context.Background(), backend, req)
	if len(second.View.Session.Routes) != 1 || len(second.View.Session.Routes[0].Targets) != 1 || second.View.Session.Routes[0].Targets[0].Title != "Two" {
		t.Fatalf("expected registry track-tab write to invalidate cached watch loop, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected registry track-tab invalidation to reuse cached raw status/profiles, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestBrowserSessionRegistryRecordPendingTargetReviewInvalidatesSharedWatchManagerCaches(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "isolated",
			Status:    "running",
			Running:   true,
			Connected: true,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	manager := SharedSessionBrowserObserverManagerFor(registry, nil, stateRegistry, time.Minute)
	sessionID := "sess-registry-pending-review-invalidates-watch"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	req := SharedSessionBrowserObserverRequest{
		SessionID:                   sessionID,
		SelectedInfo:                BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:                route,
		RequestedProfile:            "isolated",
		IncludeStatus:               true,
		IncludeProfiles:             true,
		IncludeSessionView:          true,
		SessionViewInfo:             BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		SessionViewRouteFilter:      route,
		SessionViewRequestedProfile: "isolated",
	}

	popup := registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://example.com/popup",
		Title:      "Popup",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, false)
	first := manager.ObserveWatchLoop(context.Background(), backend, req)
	if len(first.View.Session.Routes) != 1 || first.View.Session.Routes[0].PendingTargetReview != nil {
		t.Fatalf("expected initial watch loop without pending review, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}

	registry.RecordPendingTargetReviewForRoute(sessionID, route, BrowserSessionTargetReview{
		ID:         popup.ID,
		TabIndex:   popup.TabIndex,
		URL:        popup.URL,
		Title:      popup.Title,
		BrowserApp: popup.BrowserApp,
		Backend:    popup.Backend,
		Profile:    popup.Profile,
		Target:     popup.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})
	second := manager.ObserveWatchLoop(context.Background(), backend, req)
	if len(second.View.Session.Routes) != 1 || second.View.Session.Routes[0].PendingTargetReview == nil {
		t.Fatalf("expected registry pending-review write to invalidate cached watch loop, got %#v", second.View.Session.Routes)
	}
	if got := strings.TrimSpace(second.View.Session.Routes[0].PendingTargetReview.ID); got != popup.ID {
		t.Fatalf("expected refreshed pending review target %q, got %q", popup.ID, got)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected registry pending-review invalidation to reuse cached raw status/profiles, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
}

func TestBrowserSessionRegistryTrackTabRefreshesProjectionCachesBeforeNextWatchLoop(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-registry-track-tab-refreshes-projection": {{RunID: "run-1", Status: "running"}},
		}},
	}
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "isolated",
			Status:    "running",
			Running:   true,
			Connected: true,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	manager := SharedSessionBrowserObserverManagerFor(registry, runRegistry, stateRegistry, time.Minute)
	sessionID := "sess-registry-track-tab-refreshes-projection"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	req := SharedSessionBrowserObserverRequest{
		SessionID:                   sessionID,
		SelectedInfo:                BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:                route,
		RequestedProfile:            "isolated",
		IncludeStatus:               true,
		IncludeProfiles:             true,
		IncludeSessionView:          true,
		SessionViewInfo:             BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		SessionViewRouteFilter:      route,
		SessionViewRequestedProfile: "isolated",
	}

	registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	first := manager.ObserveWatchLoop(context.Background(), backend, req)
	if len(first.View.Session.Routes) != 1 || len(first.View.Session.Routes[0].Targets) != 1 || first.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected initial watch loop to expose first tracked tab, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 1 {
		t.Fatalf("expected initial watch loop to snapshot runs once, got %d", runRegistry.callCount())
	}

	registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected registry write to refresh cached projection once, got %d", runRegistry.callCount())
	}

	second := manager.ObserveWatchLoop(context.Background(), backend, req)
	if len(second.View.Session.Routes) != 1 || len(second.View.Session.Routes[0].Targets) != 1 || second.View.Session.Routes[0].Targets[0].Title != "Two" {
		t.Fatalf("expected refreshed watch loop to expose updated tracked tab, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected projection refresh to reuse cached raw status/profiles, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed projection cache, got %d", runRegistry.callCount())
	}
}

func TestBrowserSessionRegistryTrackTabRefreshesProjectionCachesFromCachedEventCycleWhenRawDrained(t *testing.T) {
	registry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	runRegistry := &countingSharedSessionRunRegistry{
		testSharedSessionRunRegistry: testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
			"sess-registry-track-tab-refreshes-from-cached-cycle": {{RunID: "run-1", Status: "running"}},
		}},
	}
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "isolated",
			Status:    "running",
			Running:   true,
			Connected: true,
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	manager := SharedSessionBrowserObserverManagerFor(registry, runRegistry, stateRegistry, time.Minute)
	bound := manager.Bind(backend)
	sessionID := "sess-registry-track-tab-refreshes-from-cached-cycle"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	req := SharedSessionBrowserObserverRequest{
		SessionID:                   sessionID,
		SelectedInfo:                BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BindingRoute:                route,
		RequestedProfile:            "isolated",
		IncludeStatus:               true,
		IncludeProfiles:             true,
		IncludeSessionView:          true,
		SessionViewInfo:             BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		SessionViewRouteFilter:      route,
		SessionViewRequestedProfile: "isolated",
	}

	registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	first := bound.ObserveWatchLoop(context.Background(), req)
	if len(first.View.Session.Routes) != 1 || len(first.View.Session.Routes[0].Targets) != 1 || first.View.Session.Routes[0].Targets[0].Title != "One" {
		t.Fatalf("expected initial watch loop to expose first tracked tab, got %#v", first.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected initial watch loop to poll backend once, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 1 {
		t.Fatalf("expected initial watch loop to snapshot runs once, got %d", runRegistry.callCount())
	}

	bound.state.mu.Lock()
	clear(bound.state.rawStatus)
	clear(bound.state.rawProfiles)
	eventCycleCount := len(bound.state.eventCycles)
	watchLoopCount := len(bound.state.watchLoops)
	bound.state.mu.Unlock()
	if eventCycleCount == 0 || watchLoopCount == 0 {
		t.Fatalf("expected cached event-cycle/watch-loop source before draining raw caches, got eventCycles=%d watchLoops=%d", eventCycleCount, watchLoopCount)
	}

	registry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com/one",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected registry write to refresh cached projection from event-cycle fallback once, got %d", runRegistry.callCount())
	}

	second := bound.ObserveWatchLoop(context.Background(), req)
	if len(second.View.Session.Routes) != 1 || len(second.View.Session.Routes[0].Targets) != 1 || second.View.Session.Routes[0].Targets[0].Title != "Two" {
		t.Fatalf("expected refreshed watch loop to expose updated tracked tab, got %#v", second.View.Session.Routes)
	}
	if len(backend.statusReqs) != 1 || len(backend.profilesReqs) != 1 {
		t.Fatalf("expected projection refresh to reuse cached event-cycle source after raw drain, got status=%d profiles=%d", len(backend.statusReqs), len(backend.profilesReqs))
	}
	if runRegistry.callCount() != 2 {
		t.Fatalf("expected next ObserveWatchLoop to reuse refreshed projection cache, got %d", runRegistry.callCount())
	}
}
