package browserruntime

import (
	"context"
	"testing"
	"time"
)

func TestObserveSharedSessionBrowserWatchLoopForScopeProjectsCycleObserverAndWatch(t *testing.T) {
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
			Note:           "profiles ok",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
			},
		},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-watch-loop"
	runRegistry := testSharedSessionRunRegistry{items: map[string][]SharedSessionRunInfo{
		sessionID: {{RunID: "run-1", Status: "running"}},
	}}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://example.com",
		Title:      "Example",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "Chromium",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})

	loop := ObserveSharedSessionBrowserWatchLoopForScope(
		context.Background(),
		backend,
		SharedSessionBrowserObserverRequest{
			SessionID:                   sessionID,
			SelectedInfo:                BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			BindingRoute:                BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
			RequestedProfile:            "isolated",
			IncludeStatus:               true,
			IncludeProfiles:             true,
			IncludeSessionView:          true,
			SessionViewInfo:             BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			SessionViewRouteFilter:      BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
			SessionViewRequestedProfile: "isolated",
		},
		sessionRegistry,
		runRegistry,
		stateRegistry,
		time.Minute,
	)

	if loop.Cycle.Observation.Status == nil || loop.Cycle.Observation.Profiles == nil {
		t.Fatalf("expected cycle to include raw status/profiles observation, got %#v", loop.Cycle)
	}
	if loop.Observer.Observation.Status == nil || loop.Watch.View.Observation.Status == nil {
		t.Fatalf("expected observer/watch to reuse shared cycle observation, got %#v %#v", loop.Observer, loop.Watch)
	}
	if len(loop.Watch.Profiles) != 1 || !loop.Watch.Profiles[0].Selected {
		t.Fatalf("expected projected selected profile, got %#v", loop.Watch.Profiles)
	}
	if len(loop.View.Session.Routes) != 1 || len(loop.View.Session.Runs) != 1 || len(loop.View.Session.Profiles) != 1 {
		t.Fatalf("expected session view projection, got %#v", loop.View)
	}
	if loop.ReferenceTime.IsZero() || !loop.ReferenceTime.Equal(loop.Watch.ReferenceTime) || !loop.ReferenceTime.Equal(loop.Observer.ReferenceTime) {
		t.Fatalf("expected watch loop reference time to align across loop/watch/observer, got loop=%v watch=%v observer=%v", loop.ReferenceTime, loop.Watch.ReferenceTime, loop.Observer.ReferenceTime)
	}
}
