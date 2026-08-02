package browserruntime

import "testing"

func TestProjectSharedSessionBrowserTopLevelProfileInventoryFromClearResultPrefersSharedState(t *testing.T) {
	projection := ProjectSharedSessionBrowserTopLevelProfileInventoryFromClearResult(
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		SharedSessionBrowserClearResult{
			ProfileStatus: BrowserProfileStatusResult{
				Backend:    "proxy",
				Profile:    "workbench",
				Status:     "stopped",
				Running:    false,
				Connected:  false,
				BrowserApp: "Chromium",
			},
			ProfileState: SharedSessionBrowserProfileState{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
				Note:          "synced clear state",
			},
			HasProfileState: true,
		},
	)

	if projection == nil || !projection.HasProfileStatus {
		t.Fatalf("expected clear-result inventory projection")
	}
	if projection.ProfileStatus.Profile != "workbench" || projection.ProfileStatus.RuntimeTarget != "node" || projection.ProfileStatus.Status != "running" || !projection.ProfileStatus.Running || !projection.ProfileStatus.Connected || projection.ProfileStatus.Note != "synced clear state" {
		t.Fatalf("expected clear-result projection to prefer shared profile state, got %#v", projection.ProfileStatus)
	}
}

func TestProjectSharedSessionBrowserTopLevelProfileInventoryFromClearResultFallsBackToRawStatus(t *testing.T) {
	projection := ProjectSharedSessionBrowserTopLevelProfileInventoryFromClearResult(
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		SharedSessionBrowserClearResult{
			ProfileStatus: BrowserProfileStatusResult{
				Backend:    "proxy",
				Profile:    "workbench",
				Status:     "stopped",
				Running:    false,
				Connected:  false,
				BrowserApp: "Chromium",
			},
		},
	)

	if projection == nil || !projection.HasProfileStatus {
		t.Fatalf("expected clear-result raw-status projection")
	}
	if projection.ProfileStatus.Profile != "workbench" || projection.ProfileStatus.RuntimeTarget != "node" || projection.ProfileStatus.Status != "stopped" || projection.ProfileStatus.Running || projection.ProfileStatus.Connected {
		t.Fatalf("expected clear-result projection to fall back to raw status, got %#v", projection.ProfileStatus)
	}
}
