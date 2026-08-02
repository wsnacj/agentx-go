package browserruntime

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func ResolveSharedSessionPreparedProfileFromProfiles(selectedInfo BrowserRuntimeInfo, result BrowserProfilesResult) string {
	profile := firstNonEmptyString(strings.TrimSpace(result.DefaultProfile), strings.TrimSpace(selectedInfo.Profile))
	if profile == "" && len(result.Profiles) > 0 {
		profile = strings.TrimSpace(result.Profiles[0].Profile)
	}
	return profile
}

func LoadSharedSessionPreparedProfile(ctx context.Context, control BrowserRuntimeControlBackend, requestedProfile string, selectedInfo BrowserRuntimeInfo) (string, *BrowserProfilesResult, time.Time, error) {
	profile := strings.TrimSpace(requestedProfile)
	if profile != "" {
		return profile, nil, time.Time{}, nil
	}
	observation := sharedSessionBrowserObserverManager(nil, nil, nil, 0).
		ObserveProfiles(ctx, control, "", selectedInfo, profile)
	if observation.ProfilesErr != nil {
		fallback := strings.TrimSpace(selectedInfo.Profile)
		if fallback == "" {
			return "", nil, time.Time{}, observation.ProfilesErr
		}
		return fallback, nil, time.Time{}, nil
	}
	if observation.Profiles == nil {
		fallback := strings.TrimSpace(selectedInfo.Profile)
		if fallback == "" {
			return "", nil, time.Time{}, fmt.Errorf("browser_runtime: no managed browser profile is available for prepare")
		}
		return fallback, nil, time.Time{}, nil
	}
	result := *observation.Profiles
	profile = ResolveSharedSessionPreparedProfileFromProfiles(selectedInfo, result)
	if profile == "" {
		return "", &result, observation.ObservedAt, fmt.Errorf("browser_runtime: no managed browser profile is available for prepare")
	}
	return profile, &result, observation.ObservedAt, nil
}
