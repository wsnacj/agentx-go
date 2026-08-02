package browserruntime

import "testing"

func TestBrowserSessionRegistryTrackResolveAndForget(t *testing.T) {
	reg := NewBrowserSessionRegistry()

	first := reg.TrackTab("s1", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://example.com/a",
		Title:      "A",
		BrowserApp: "Safari",
		Backend:    "system",
	}, true)
	if first.ID == "" {
		t.Fatal("expected tracked target id")
	}

	resolved, ok := reg.ResolveTarget("s1", first.ID)
	if !ok || resolved.TabIndex != 2 || resolved.URL != "https://example.com/a" {
		t.Fatalf("unexpected resolved target: %#v ok=%v", resolved, ok)
	}

	updated := reg.TrackTab("s1", BrowserSessionTarget{
		TabIndex: 2,
		URL:      "https://example.com/b",
		Title:    "B",
	}, true)
	if updated.ID != first.ID {
		t.Fatalf("expected stable id reuse, got first=%q updated=%q", first.ID, updated.ID)
	}
	if updated.URL != "https://example.com/b" || updated.Title != "B" {
		t.Fatalf("expected updated metadata, got %#v", updated)
	}

	current, ok := reg.CurrentTarget("s1")
	if !ok || current.ID != first.ID {
		t.Fatalf("expected current target %q, got %#v ok=%v", first.ID, current, ok)
	}

	reg.ForgetTab("s1", 2)
	if _, ok := reg.ResolveTarget("s1", first.ID); ok {
		t.Fatalf("expected target %q to be forgotten", first.ID)
	}
}

func TestBrowserSessionRegistryScopesTabsByRoute(t *testing.T) {
	reg := NewBrowserSessionRegistry()

	host := reg.TrackTab("s1", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://host.example",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, true)
	node := reg.TrackTab("s1", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	if host.ID == "" || node.ID == "" || host.ID == node.ID {
		t.Fatalf("expected distinct route-scoped target ids, got host=%#v node=%#v", host, node)
	}

	hostResolved, ok := reg.ResolveTabForRoute("s1", BrowserSessionRoute{
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
		BrowserApp: "Safari",
	}, 1)
	if !ok || hostResolved.ID != host.ID || hostResolved.URL != host.URL {
		t.Fatalf("unexpected host route resolution: %#v ok=%v", hostResolved, ok)
	}
	nodeResolved, ok := reg.ResolveTabForRoute("s1", BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "Chromium",
	}, 1)
	if !ok || nodeResolved.ID != node.ID || nodeResolved.URL != node.URL {
		t.Fatalf("unexpected node route resolution: %#v ok=%v", nodeResolved, ok)
	}
}

func TestBrowserSessionRegistryTrackTabsRemovesStaleEntries(t *testing.T) {
	reg := NewBrowserSessionRegistry()

	initial := reg.TrackTabs("s1", []BrowserSessionTarget{
		{TabIndex: 1, URL: "https://a.example"},
		{TabIndex: 2, URL: "https://b.example"},
	}, 2)
	if len(initial) != 2 || initial[0].ID == "" || initial[1].ID == "" {
		t.Fatalf("expected tracked tabs with ids, got %#v", initial)
	}

	next := reg.TrackTabs("s1", []BrowserSessionTarget{
		{TabIndex: 2, URL: "https://b.example/next"},
	}, 2)
	if len(next) != 1 || next[0].ID != initial[1].ID {
		t.Fatalf("expected tab 2 id retained, got %#v", next)
	}
	if _, ok := reg.ResolveTab("s1", 1); ok {
		t.Fatal("expected stale tab 1 to be removed")
	}
}

func TestBrowserSessionRegistrySyncTabsForRouteClearsEmptyRoute(t *testing.T) {
	reg := NewBrowserSessionRegistry()
	sessionID := "s1"
	route := BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "chromium",
	}

	tracked := reg.TrackTabs(sessionID, []BrowserSessionTarget{
		{TabIndex: 1, URL: "https://node.example/home", Title: "Home", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 3, URL: "https://popup.example/offer", Title: "Offer", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
	}, 1)
	if len(tracked) != 2 {
		t.Fatalf("expected tracked tabs, got %#v", tracked)
	}
	reg.RecordPendingTargetReviewForRoute(sessionID, route, BrowserSessionTargetReview{
		ID:         tracked[1].ID,
		TabIndex:   tracked[1].TabIndex,
		URL:        tracked[1].URL,
		Title:      tracked[1].Title,
		BrowserApp: tracked[1].BrowserApp,
		Backend:    tracked[1].Backend,
		Profile:    tracked[1].Profile,
		Target:     tracked[1].Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	cleared := reg.SyncTabsForRoute(sessionID, route, nil, 0)
	if len(cleared) != 0 {
		t.Fatalf("expected empty sync result, got %#v", cleared)
	}
	if _, ok := reg.CurrentTargetForRoute(sessionID, route); ok {
		t.Fatal("expected current target to clear after empty route sync")
	}
	if _, ok := reg.ResolveTabForRoute(sessionID, route, 1); ok {
		t.Fatal("expected route tab 1 to clear after empty route sync")
	}
	if _, ok := reg.ResolveTarget(sessionID, tracked[1].ID); ok {
		t.Fatalf("expected popup target %q to clear after empty route sync", tracked[1].ID)
	}
	if snapshot := reg.Snapshot(sessionID); len(snapshot) != 0 {
		t.Fatalf("expected route snapshot to clear after empty route sync, got %#v", snapshot)
	}
}

func TestBrowserSessionRegistryPruneStaleRouteState(t *testing.T) {
	reg := NewBrowserSessionRegistry()
	sessionID := "s1"
	route := BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "chromium",
	}

	tracked := reg.TrackTabs(sessionID, []BrowserSessionTarget{
		{TabIndex: 1, URL: "https://node.example/home", Title: "Home", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 3, URL: "https://popup.example/offer", Title: "Offer", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
	}, 1)
	if len(tracked) != 2 {
		t.Fatalf("expected tracked targets, got %#v", tracked)
	}
	reg.RecordPendingTargetReviewForRoute(sessionID, route, BrowserSessionTargetReview{
		ID:         tracked[1].ID,
		TabIndex:   tracked[1].TabIndex,
		URL:        tracked[1].URL,
		Title:      tracked[1].Title,
		BrowserApp: tracked[1].BrowserApp,
		Backend:    tracked[1].Backend,
		Profile:    tracked[1].Profile,
		Target:     tracked[1].Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	reg.mu.Lock()
	state := reg.sessions[sessionID]
	delete(state.targets, tracked[0].ID)
	delete(state.targets, tracked[1].ID)
	reg.mu.Unlock()

	if !reg.PruneStaleRouteState(sessionID, route) {
		t.Fatal("expected prune to report changed state")
	}
	if snapshot := reg.Snapshot(sessionID); len(snapshot) != 0 {
		t.Fatalf("expected stale route state to be fully pruned, got %#v", snapshot)
	}
}

func TestBrowserSessionRegistryPruneStaleRouteStateClearsCurrentOnlyTarget(t *testing.T) {
	reg := NewBrowserSessionRegistry()
	sessionID := "s1"
	route := BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "chromium",
	}

	current := reg.TrackCurrentTarget(sessionID, BrowserSessionTarget{
		URL:        "https://node.example/current",
		Title:      "Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, "select_target")
	if current.ID == "" {
		t.Fatalf("expected current-only target id, got %#v", current)
	}

	reg.mu.Lock()
	state := reg.sessions[sessionID]
	delete(state.targets, current.ID)
	reg.mu.Unlock()

	if !reg.PruneStaleRouteState(sessionID, route) {
		t.Fatal("expected prune to report changed state")
	}
	if _, _, ok := reg.CurrentTargetSelectionForRoute(sessionID, route); ok {
		t.Fatal("expected stale current-only target selection to be pruned")
	}
	if snapshot := reg.Snapshot(sessionID); len(snapshot) != 0 {
		t.Fatalf("expected stale current-only route state to be removed, got %#v", snapshot)
	}
}

func TestBrowserSessionRegistryCurrentTargetForRouteRejectsStaleMismatchedDirectRouteState(t *testing.T) {
	reg := NewBrowserSessionRegistry()
	sessionID := "s1"
	route := BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "chromium",
	}

	current := reg.TrackCurrentTarget(sessionID, BrowserSessionTarget{
		URL:        "https://node.example/current",
		Title:      "Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, "select_target")
	if current.ID == "" {
		t.Fatalf("expected current target id, got %#v", current)
	}

	reg.mu.Lock()
	state := reg.sessions[sessionID]
	target := state.targets[current.ID]
	target.Backend = "system"
	target.Profile = "default"
	target.Target = "host"
	state.targets[current.ID] = target
	reg.mu.Unlock()

	if _, _, ok := reg.CurrentTargetSelectionForRoute(sessionID, route); ok {
		t.Fatal("expected stale mismatched current target selection to be ignored")
	}
}

func TestBrowserSessionRegistryResolveTabForRouteRejectsStaleMismatchedDirectRouteState(t *testing.T) {
	reg := NewBrowserSessionRegistry()
	sessionID := "s1"
	route := BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "chromium",
	}

	tracked := reg.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/home",
		Title:      "Home",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	if tracked.ID == "" {
		t.Fatalf("expected tracked target id, got %#v", tracked)
	}

	reg.mu.Lock()
	state := reg.sessions[sessionID]
	target := state.targets[tracked.ID]
	target.Backend = "system"
	target.Profile = "default"
	target.Target = "host"
	state.targets[tracked.ID] = target
	reg.mu.Unlock()

	if _, ok := reg.ResolveTabForRoute(sessionID, route, 1); ok {
		t.Fatal("expected stale mismatched tab mapping to be ignored")
	}
}

func TestBrowserSessionRegistryTrackTabsOnlyPrunesWithinRoute(t *testing.T) {
	reg := NewBrowserSessionRegistry()

	host := reg.TrackTabs("s1", []BrowserSessionTarget{
		{TabIndex: 1, URL: "https://host.example", BrowserApp: "Safari", Backend: "system", Profile: "default", Target: "host"},
	}, 1)
	node := reg.TrackTabs("s1", []BrowserSessionTarget{
		{TabIndex: 1, URL: "https://node.example", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
	}, 1)
	if len(host) != 1 || len(node) != 1 {
		t.Fatalf("expected tracked tabs, got host=%#v node=%#v", host, node)
	}

	reg.TrackTabs("s1", []BrowserSessionTarget{
		{TabIndex: 2, URL: "https://host.example/next", BrowserApp: "Safari", Backend: "system", Profile: "default", Target: "host"},
	}, 2)

	if _, ok := reg.ResolveTabForRoute("s1", BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "Chromium",
	}, 1); !ok {
		t.Fatal("expected node route tab to survive host-only prune")
	}
	if _, ok := reg.ResolveTabForRoute("s1", BrowserSessionRoute{
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
		BrowserApp: "Safari",
	}, 1); ok {
		t.Fatal("expected stale host tab 1 to be pruned")
	}
}

func TestBrowserSessionRegistrySnapshot(t *testing.T) {
	reg := NewBrowserSessionRegistry()

	reg.TrackTab("s1", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://host.example/b",
		Title:      "Host B",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, false)
	current := reg.TrackTab("s1", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/a",
		Title:      "Node A",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)

	snapshot := reg.Snapshot("s1")
	if len(snapshot) != 2 {
		t.Fatalf("expected 2 route snapshots, got %#v", snapshot)
	}
	nodeFound := false
	for _, route := range snapshot {
		if route.Route.Backend != "proxy" || route.Route.Profile != "isolated" || route.Route.Target != "node" {
			continue
		}
		nodeFound = true
		if route.CurrentTargetID != current.ID {
			t.Fatalf("expected current target id %q, got %#v", current.ID, route)
		}
		if len(route.Targets) != 1 || route.Targets[0].ID != current.ID || route.Targets[0].TabIndex != 1 {
			t.Fatalf("unexpected node route targets: %#v", route.Targets)
		}
	}
	if !nodeFound {
		t.Fatalf("expected node route snapshot, got %#v", snapshot)
	}
}

func TestBrowserSessionRegistryPendingTargetReviewCountTracksDistinctTargets(t *testing.T) {
	reg := NewBrowserSessionRegistry()
	sessionID := "s1"
	route := BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "chromium",
	}

	tracked := reg.TrackTabs(sessionID, []BrowserSessionTarget{
		{TabIndex: 1, URL: "https://node.example/home", Title: "Home", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 3, URL: "https://popup.example/offer", Title: "Offer", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 4, URL: "https://popup.example/bonus", Title: "Bonus", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
	}, 1)
	if len(tracked) != 3 {
		t.Fatalf("expected tracked tabs, got %#v", tracked)
	}

	first := BrowserSessionTargetReview{
		ID:         tracked[1].ID,
		TabIndex:   tracked[1].TabIndex,
		URL:        tracked[1].URL,
		Title:      tracked[1].Title,
		BrowserApp: tracked[1].BrowserApp,
		Backend:    tracked[1].Backend,
		Profile:    tracked[1].Profile,
		Target:     tracked[1].Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	}
	reg.RecordPendingTargetReviewForRoute(sessionID, route, first)
	reg.RecordPendingTargetReviewForRoute(sessionID, route, first)

	snapshot := reg.Snapshot(sessionID)
	if len(snapshot) != 1 || snapshot[0].PendingTargetReview == nil || snapshot[0].PendingTargetReviewCount != 1 {
		t.Fatalf("expected single pending popup review count, got %#v", snapshot)
	}

	second := BrowserSessionTargetReview{
		ID:         tracked[2].ID,
		TabIndex:   tracked[2].TabIndex,
		URL:        tracked[2].URL,
		Title:      tracked[2].Title,
		BrowserApp: tracked[2].BrowserApp,
		Backend:    tracked[2].Backend,
		Profile:    tracked[2].Profile,
		Target:     tracked[2].Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	}
	reg.RecordPendingTargetReviewForRoute(sessionID, route, second)

	snapshot = reg.Snapshot(sessionID)
	if len(snapshot) != 1 || snapshot[0].PendingTargetReview == nil {
		t.Fatalf("expected pending popup review snapshot, got %#v", snapshot)
	}
	if snapshot[0].PendingTargetReview.ID != second.ID || snapshot[0].PendingTargetReviewCount != 2 {
		t.Fatalf("expected latest review id %q with count 2, got %#v", second.ID, snapshot[0])
	}

	if !reg.ClearPendingTargetReviewForRoute(sessionID, route) {
		t.Fatal("expected pending popup review to clear")
	}
	snapshot = reg.Snapshot(sessionID)
	if len(snapshot) != 1 || snapshot[0].PendingTargetReview != nil || snapshot[0].PendingTargetReviewCount != 0 {
		t.Fatalf("expected pending popup review count reset, got %#v", snapshot)
	}
}

func TestBrowserSessionRegistryForgetTabFallsBackToEarlierPendingTargetReview(t *testing.T) {
	reg := NewBrowserSessionRegistry()
	sessionID := "s1"
	route := BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "chromium",
	}

	tracked := reg.TrackTabs(sessionID, []BrowserSessionTarget{
		{TabIndex: 1, URL: "https://node.example/home", Title: "Home", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 3, URL: "https://popup.example/offer", Title: "Offer", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 4, URL: "https://popup.example/bonus", Title: "Bonus", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
	}, 1)
	first := BrowserSessionTargetReview{
		ID:         tracked[1].ID,
		TabIndex:   tracked[1].TabIndex,
		URL:        tracked[1].URL,
		Title:      tracked[1].Title,
		BrowserApp: tracked[1].BrowserApp,
		Backend:    tracked[1].Backend,
		Profile:    tracked[1].Profile,
		Target:     tracked[1].Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	}
	second := BrowserSessionTargetReview{
		ID:         tracked[2].ID,
		TabIndex:   tracked[2].TabIndex,
		URL:        tracked[2].URL,
		Title:      tracked[2].Title,
		BrowserApp: tracked[2].BrowserApp,
		Backend:    tracked[2].Backend,
		Profile:    tracked[2].Profile,
		Target:     tracked[2].Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	}
	reg.RecordPendingTargetReviewForRoute(sessionID, route, first)
	reg.RecordPendingTargetReviewForRoute(sessionID, route, second)

	reg.ForgetTabForRoute(sessionID, route, 4)

	snapshot := reg.Snapshot(sessionID)
	if len(snapshot) != 1 {
		t.Fatalf("expected one route snapshot, got %#v", snapshot)
	}
	if snapshot[0].PendingTargetReview == nil || snapshot[0].PendingTargetReview.ID != first.ID || snapshot[0].PendingTargetReview.TabIndex != 3 || snapshot[0].PendingTargetReviewCount != 1 {
		t.Fatalf("expected fallback to earlier pending popup review, got %#v", snapshot[0])
	}
	if _, ok := reg.ResolveTabForRoute(sessionID, route, 4); ok {
		t.Fatal("expected forgotten popup tab to be removed")
	}
}

func TestBrowserSessionRegistryTrackCurrentTargetWithoutTabIndex(t *testing.T) {
	reg := NewBrowserSessionRegistry()
	route := BrowserSessionRoute{
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
		BrowserApp: "safari",
	}

	first := reg.TrackCurrentTarget("s1", BrowserSessionTarget{
		URL:        "https://example.com/one",
		Title:      "One",
		BrowserApp: "Safari",
		Backend:    "system_open",
		Profile:    "default",
		Target:     "host",
	}, "browser_open")
	if first.ID == "" || first.TabIndex != 0 {
		t.Fatalf("expected current-only target id, got %#v", first)
	}
	current, source, ok := reg.CurrentTargetSelectionForRoute("s1", route)
	if !ok || current.ID != first.ID || current.URL != "https://example.com/one" {
		t.Fatalf("unexpected current target selection: %#v source=%q ok=%v", current, source, ok)
	}
	if source != "browser_open" {
		t.Fatalf("expected current target source browser_open, got %q", source)
	}

	next := reg.TrackCurrentTarget("s1", BrowserSessionTarget{
		URL:        "https://example.com/two",
		Title:      "Two",
		BrowserApp: "Safari",
		Backend:    "system_open",
		Profile:    "default",
		Target:     "host",
	})
	if next.ID != first.ID {
		t.Fatalf("expected current-only target id reuse, got first=%q next=%q", first.ID, next.ID)
	}
	if current, ok := reg.CurrentTargetForRoute("s1", route); !ok || current.URL != "https://example.com/two" || current.Title != "Two" {
		t.Fatalf("expected current-only target metadata update, got %#v ok=%v", current, ok)
	}
}

func TestBrowserSessionRegistryTrackCurrentTargetReusesEquivalentSystemRoute(t *testing.T) {
	reg := NewBrowserSessionRegistry()
	sessionID := "s1"
	route := BrowserSessionRoute{
		Backend: "system",
		Profile: "default",
		Target:  "host",
	}

	first := reg.TrackCurrentTarget(sessionID, BrowserSessionTarget{
		URL:     "https://example.com/iphone",
		Title:   "Apple",
		Backend: "system_open",
		Profile: "default",
		Target:  "host",
	}, "browser_open")
	next := reg.TrackCurrentTarget(sessionID, BrowserSessionTarget{
		URL:     "https://example.com/iphone",
		Title:   "iPhone",
		Backend: "http_extract_fallback",
		Profile: "default",
		Target:  "host",
	}, "browser_extract")
	if first.ID == "" || next.ID == "" {
		t.Fatalf("expected tracked current target ids, got first=%#v next=%#v", first, next)
	}
	if next.ID != first.ID {
		t.Fatalf("expected equivalent system route current target reuse, got first=%q next=%q", first.ID, next.ID)
	}
	current, source, ok := reg.CurrentTargetSelectionForRoute(sessionID, route)
	if !ok || current.ID != first.ID || current.URL != "https://example.com/iphone" || current.Title != "iPhone" {
		t.Fatalf("unexpected current target selection: %#v source=%q ok=%v", current, source, ok)
	}
	if source != "browser_extract" {
		t.Fatalf("expected latest current target source browser_extract, got %q", source)
	}
	snapshot := reg.Snapshot(sessionID)
	if len(snapshot) != 1 {
		t.Fatalf("expected equivalent system backends to share one route snapshot, got %#v", snapshot)
	}
}

func TestBrowserSessionRegistrySelectAndClearCurrentTargetForRoute(t *testing.T) {
	reg := NewBrowserSessionRegistry()

	first := reg.TrackTab("s1", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/a",
		Title:      "A",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, false)
	second := reg.TrackTab("s1", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/b",
		Title:      "B",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, false)

	route := BrowserSessionRoute{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}
	selected, ok := reg.SelectTabForRoute("s1", route, 2)
	if !ok || selected.ID != second.ID {
		t.Fatalf("expected tab 2 to be selected, got %#v ok=%v", selected, ok)
	}
	current, source, ok := reg.CurrentTargetSelectionForRoute("s1", route)
	if !ok || current.ID != second.ID {
		t.Fatalf("expected current target %q, got %#v ok=%v", second.ID, current, ok)
	}
	if source != "select_target" {
		t.Fatalf("expected current target source select_target, got %q", source)
	}
	selected, ok = reg.SelectTargetForRoute("s1", route, first.ID)
	if !ok || selected.ID != first.ID {
		t.Fatalf("expected target %q to be selected directly, got %#v ok=%v", first.ID, selected, ok)
	}
	current, source, ok = reg.CurrentTargetSelectionForRoute("s1", route)
	if !ok || current.ID != first.ID {
		t.Fatalf("expected current target %q, got %#v ok=%v", first.ID, current, ok)
	}
	if source != "select_target" {
		t.Fatalf("expected current target source select_target after direct selection, got %q", source)
	}
	if !reg.ClearCurrentTargetForRoute("s1", route) {
		t.Fatal("expected current target clear to succeed")
	}
	if _, ok := reg.CurrentTargetForRoute("s1", route); ok {
		t.Fatal("expected route current target to be cleared")
	}
}

func TestBrowserSessionRegistryClearRoute(t *testing.T) {
	reg := NewBrowserSessionRegistry()
	nodeRoute := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}
	hostRoute := BrowserSessionRoute{Backend: "system", Profile: "default", Target: "host"}
	first := reg.TrackTab("s1", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	reg.TrackTab("s1", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	host := reg.TrackTab("s1", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://host.example/one",
		Title:      "Host",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, true)

	if cleared := reg.ClearRoute("s1", nodeRoute); cleared != 2 {
		t.Fatalf("expected to clear 2 node targets, got %d", cleared)
	}
	if _, ok := reg.CurrentTargetForRoute("s1", nodeRoute); ok {
		t.Fatal("expected node current target to clear after route reset")
	}
	if _, ok := reg.ResolveTabForRoute("s1", nodeRoute, 1); ok {
		t.Fatal("expected node tab lookup to clear after route reset")
	}
	if target, ok := reg.CurrentTargetForRoute("s1", hostRoute); !ok || target.ID != host.ID {
		t.Fatalf("expected host route target to remain intact, got %#v ok=%v", target, ok)
	}
	if target, ok := reg.ResolveTabForRoute("s1", BrowserSessionRoute{}, 1); ok && target.ID == first.ID {
		t.Fatalf("expected cleared node target not to resolve through broad lookup, got %#v", target)
	}
}

func TestBrowserSessionRegistryClearRouteClearsCurrentOnlyTarget(t *testing.T) {
	reg := NewBrowserSessionRegistry()
	route := BrowserSessionRoute{Backend: "system", Profile: "default", Target: "host", BrowserApp: "safari"}

	tracked := reg.TrackCurrentTarget("s1", BrowserSessionTarget{
		URL:        "https://example.com/current",
		BrowserApp: "Safari",
		Backend:    "system_open",
		Profile:    "default",
		Target:     "host",
	}, "browser_open")
	if tracked.ID == "" {
		t.Fatalf("expected current-only target id")
	}
	if cleared := reg.ClearRoute("s1", route); cleared != 1 {
		t.Fatalf("expected to clear current-only route target, got %d", cleared)
	}
	if _, ok := reg.CurrentTargetForRoute("s1", route); ok {
		t.Fatal("expected current-only route target to clear")
	}
	if _, ok := reg.ResolveTarget("s1", tracked.ID); ok {
		t.Fatalf("expected current-only target %q to be removed", tracked.ID)
	}
}

func TestBrowserSessionRegistryRouteFilterMatchesBackendFamilies(t *testing.T) {
	reg := NewBrowserSessionRegistry()
	reg.TrackTab("s1", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy-tabs",
		Profile:    "workbench",
		Target:     "node",
	}, true)

	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}
	if target, ok := reg.CurrentTargetForRoute("s1", route); !ok || target.TabIndex != 3 {
		t.Fatalf("expected current target lookup to match proxy backend family, got %#v ok=%v", target, ok)
	}
	if cleared := reg.ClearRoute("s1", route); cleared != 1 {
		t.Fatalf("expected family-matched route clear to remove 1 target, got %d", cleared)
	}
	if _, ok := reg.ResolveTabForRoute("s1", route, 3); ok {
		t.Fatal("expected route clear to remove proxy family-matched target")
	}
}

func TestBrowserSessionRouteHelpersNormalizeAndMatch(t *testing.T) {
	normalized := normalizeBrowserSessionRoute(BrowserSessionRoute{
		Backend:    " Proxy ",
		Profile:    " Workbench ",
		Target:     " Node ",
		BrowserApp: " Chromium ",
	})
	if normalized.Backend != "proxy" || normalized.Profile != "workbench" || normalized.Target != "node" || normalized.BrowserApp != "chromium" {
		t.Fatalf("expected route normalization to trim+lowercase fields, got %#v", normalized)
	}

	if key := browserSessionRouteKey(BrowserSessionRoute{}); key != "__default__" {
		t.Fatalf("expected empty route to map to default route key, got %q", key)
	}

	candidate := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node", BrowserApp: "chromium"}
	if !browserSessionRouteMatchesFilter(candidate, BrowserSessionRoute{Backend: "proxy", Target: "node"}) {
		t.Fatalf("expected partial route filter to match candidate")
	}
	if browserSessionRouteMatchesFilter(candidate, BrowserSessionRoute{Backend: "system"}) {
		t.Fatalf("expected mismatched backend filter to fail")
	}
}
