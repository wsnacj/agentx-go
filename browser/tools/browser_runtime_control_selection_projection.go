package tools

type browserRuntimeSessionSelectionProjection struct {
	ProfileSelection   *browserRuntimeSessionProfileSelection
	TargetSelection    *browserRuntimeSessionTargetSelection
	ApplyTargetToRoute bool
}

func browserRuntimeApplySessionSelectionProjection(
	payload *browserRuntimePayload,
	projection browserRuntimeSessionSelectionProjection,
) {
	if payload == nil {
		return
	}
	if projection.ProfileSelection != nil {
		payload.SessionProfileSelection = projection.ProfileSelection
	}
	if projection.TargetSelection != nil {
		payload.SessionTargetSelection = projection.TargetSelection
		if projection.ApplyTargetToRoute {
			browserRuntimeApplySessionSelectionsToRoute(
				payload.SelectedRoute,
				payload.SessionProfileSelection,
				projection.TargetSelection,
			)
		}
	}
}
