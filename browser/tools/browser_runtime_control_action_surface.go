package tools

import (
	"context"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserRuntimeApplySharedActionSurface(
	callCtx context.Context,
	payload *browserRuntimePayload,
	surface agentxbrowserruntime.SharedSessionBrowserActionSurfaceProjection,
) {
	if payload == nil {
		return
	}
	browserRuntimeApplyActionTerminalStatus(payload, browserRuntimeActionTerminalStatus{
		Status:               surface.Status,
		Note:                 surface.Note,
		PreserveExistingNote: true,
	})
	if surface.WorkbenchSurface != nil {
		browserRuntimeApplySharedWorkbenchSessionSurface(callCtx, payload, *surface.WorkbenchSurface, nil, false)
		return
	}
	if surface.ProfileInventory != nil {
		browserRuntimeApplyTopLevelProfileInventory(
			payload,
			*browserRuntimeTopLevelProfileInventoryProjectionFromShared(
				surface.ConfiguredInfo,
				surface.ProfileInventory,
			),
		)
	}
	if projection := browserRuntimeTopLevelSessionProjectionApplicationFromShared(surface.SessionProjection); projection != nil {
		browserRuntimeApplyTopLevelSessionProjection(payload, *projection)
	}
	if browserRuntimeMaybeSyncSharedGuidanceProjection(payload, false) {
		browserRuntimeSyncTopLevelSurfaceSummary(payload)
	}
}
