package browserruntime

import (
	"context"
	"strings"
	"time"
)

// SharedSessionBrowserRawObservationReplayTTL is the default freshness window
// for replaying source-time raw browser observations into lifecycle-owned
// runtime/session state.
const SharedSessionBrowserRawObservationReplayTTL = 2 * time.Second

// SharedSessionBrowserRawStatusObservation captures the direct RuntimeStatus
// backend load before any registry sync or lifecycle-owned resolution.
type SharedSessionBrowserRawStatusObservation struct {
	RequestedProfile string
	Status           *BrowserProfileStatusResult
	Err              error
	ObservedAt       time.Time
}

// SharedSessionBrowserRawProfilesObservation captures the direct
// RuntimeProfiles backend load before any scoped projection or registry sync.
type SharedSessionBrowserRawProfilesObservation struct {
	RequestedProfile string
	Profiles         *BrowserProfilesResult
	Err              error
	ObservedAt       time.Time
}

// SharedSessionBrowserRawLifecycleObservation captures the direct
// RuntimeStart/RuntimeStop backend load before lifecycle-owned mapping.
type SharedSessionBrowserRawLifecycleObservation struct {
	Profile    string
	Status     *BrowserProfileStatusResult
	Err        error
	ObservedAt time.Time
}

// SharedSessionBrowserRawStatusAndProfilesObservation captures a single raw
// polling cycle for optional RuntimeStatus and RuntimeProfiles loads before any
// lifecycle-owned sync or projection.
type SharedSessionBrowserRawStatusAndProfilesObservation struct {
	RequestedProfile   string
	Status             *BrowserProfileStatusResult
	StatusErr          error
	StatusObservedAt   time.Time
	Profiles           *BrowserProfilesResult
	ProfilesErr        error
	ProfilesObservedAt time.Time
}

// SharedSessionBrowserRawObservationExpired reports whether a source-time raw
// observation is too old to replay into lifecycle-owned state.
func SharedSessionBrowserRawObservationExpired(observedAt time.Time, now time.Time, ttl time.Duration) bool {
	if observedAt.IsZero() {
		return true
	}
	if now.IsZero() {
		now = time.Now()
	}
	return now.Sub(observedAt) > ttl
}

// SharedSessionBrowserRawObservationKeys returns the profile keys that should
// index a raw source-time observation.
func SharedSessionBrowserRawObservationKeys(requestedProfile string, resolvedProfile string) []string {
	keys := make([]string, 0, 2)
	for _, key := range []string{strings.TrimSpace(requestedProfile), strings.TrimSpace(resolvedProfile)} {
		if len(keys) == 0 || keys[len(keys)-1] != key {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return []string{""}
	}
	return keys
}

// SharedSessionBrowserRawStatusObservationLookupKeysForProfiles returns the
// status observation cache lookup keys implied by a raw profiles result.
func SharedSessionBrowserRawStatusObservationLookupKeysForProfiles(requestedProfile string, profiles *BrowserProfilesResult) []string {
	return SharedSessionBrowserRawObservationKeys(requestedProfile, sharedSessionBrowserRawProfilesDefaultProfile(profiles))
}

// SharedSessionBrowserRawProfilesObservationLookupKeysForStatus returns the
// profiles observation cache lookup keys implied by a raw status result.
func SharedSessionBrowserRawProfilesObservationLookupKeysForStatus(requestedProfile string, status *BrowserProfileStatusResult) []string {
	statusProfile := sharedSessionBrowserRawStatusProfile(status)
	keys := SharedSessionBrowserRawObservationKeys(requestedProfile, statusProfile)
	if statusProfile != "" {
		keys = appendSharedSessionBrowserRawObservationUniqueKeys(
			keys,
			SharedSessionBrowserRawObservationKeys(statusProfile, statusProfile)...,
		)
	}
	return keys
}

func sharedSessionBrowserRawStatusProfile(status *BrowserProfileStatusResult) string {
	if status == nil {
		return ""
	}
	return strings.TrimSpace(status.Profile)
}

func sharedSessionBrowserRawProfilesDefaultProfile(profiles *BrowserProfilesResult) string {
	if profiles == nil {
		return ""
	}
	return strings.TrimSpace(profiles.DefaultProfile)
}

func appendSharedSessionBrowserRawObservationUniqueKeys(keys []string, candidates ...string) []string {
	for _, candidate := range candidates {
		key := strings.TrimSpace(candidate)
		seen := false
		for _, existing := range keys {
			if existing == key {
				seen = true
				break
			}
		}
		if !seen {
			keys = append(keys, key)
		}
	}
	return keys
}

// SharedSessionBrowserRawObservationProfileFilter is a normalized set of
// profiles used to invalidate raw source-time observation caches.
type SharedSessionBrowserRawObservationProfileFilter map[string]struct{}

// BuildSharedSessionBrowserRawObservationProfileFilter returns a normalized
// profile filter for raw source-time observation cache invalidation.
func BuildSharedSessionBrowserRawObservationProfileFilter(profiles ...string) SharedSessionBrowserRawObservationProfileFilter {
	filter := make(SharedSessionBrowserRawObservationProfileFilter, len(profiles))
	for _, profile := range profiles {
		if trimmed := strings.TrimSpace(profile); trimmed != "" {
			filter[trimmed] = struct{}{}
		}
	}
	return filter
}

// Empty reports whether the filter has no usable profile entries.
func (filter SharedSessionBrowserRawObservationProfileFilter) Empty() bool {
	return len(filter) == 0
}

// Matches reports whether any candidate profile is present in the filter.
func (filter SharedSessionBrowserRawObservationProfileFilter) Matches(profiles ...string) bool {
	if len(filter) == 0 {
		return false
	}
	for _, profile := range profiles {
		if _, ok := filter[strings.TrimSpace(profile)]; ok {
			return true
		}
	}
	return false
}

// SharedSessionBrowserRawStatusObservationProfileCandidates returns the
// profile values that should be considered when invalidating a raw status
// observation cached under cacheKey.
func SharedSessionBrowserRawStatusObservationProfileCandidates(cacheKey string, observation SharedSessionBrowserRawStatusObservation) []string {
	return appendSharedSessionBrowserRawObservationUniqueKeys(nil, cacheKey, sharedSessionBrowserRawStatusProfile(observation.Status))
}

// SharedSessionBrowserRawTabsObservationProfileCandidates returns the profile
// values that should be considered when invalidating a raw tabs observation.
func SharedSessionBrowserRawTabsObservationProfileCandidates(cacheKey string, observation SharedSessionBrowserRawTabsObservation) []string {
	return sharedSessionBrowserRawRequestedProfileCandidates(cacheKey, observation.RequestedProfile)
}

// SharedSessionBrowserRawRouteMutationObservationProfileCandidates returns the
// profile values that should be considered when invalidating a raw route
// mutation observation.
func SharedSessionBrowserRawRouteMutationObservationProfileCandidates(cacheKey string, observation SharedSessionBrowserRawRouteMutationObservation) []string {
	return appendSharedSessionBrowserRawObservationUniqueKeys(nil, cacheKey, observation.RequestedProfile, observation.Route.Profile)
}

// SharedSessionBrowserRawNavigationObservationProfileCandidates returns the
// profile values that should be considered when invalidating a raw navigation
// observation.
func SharedSessionBrowserRawNavigationObservationProfileCandidates(cacheKey string, observation SharedSessionBrowserRawNavigationObservation) []string {
	return sharedSessionBrowserRawRequestedProfileCandidates(cacheKey, observation.RequestedProfile)
}

// SharedSessionBrowserRawOpenObservationProfileCandidates returns the profile
// values that should be considered when invalidating a raw open observation.
func SharedSessionBrowserRawOpenObservationProfileCandidates(cacheKey string, observation SharedSessionBrowserRawOpenObservation) []string {
	return sharedSessionBrowserRawRequestedProfileCandidates(cacheKey, observation.RequestedProfile)
}

// SharedSessionBrowserRawTargetObservationProfileCandidates returns the
// profile values that should be considered when invalidating a raw target
// observation.
func SharedSessionBrowserRawTargetObservationProfileCandidates(cacheKey string, observation SharedSessionBrowserRawTargetObservation) []string {
	return sharedSessionBrowserRawRequestedProfileCandidates(cacheKey, observation.RequestedProfile)
}

// SharedSessionBrowserRawLifecycleObservationProfileCandidates returns the
// profile values that should be considered when invalidating a raw lifecycle
// observation.
func SharedSessionBrowserRawLifecycleObservationProfileCandidates(cacheKey string, observation SharedSessionBrowserRawLifecycleObservation) []string {
	return appendSharedSessionBrowserRawObservationUniqueKeys(nil, cacheKey, observation.Profile)
}

// SharedSessionBrowserRawStatusAndProfilesObservationProfileCandidates returns
// the profile values that should be considered when invalidating a combined raw
// status/profiles observation.
func SharedSessionBrowserRawStatusAndProfilesObservationProfileCandidates(cacheKey string, observation SharedSessionBrowserRawStatusAndProfilesObservation) []string {
	return appendSharedSessionBrowserRawObservationUniqueKeys(
		nil,
		cacheKey,
		sharedSessionBrowserRawStatusProfile(observation.Status),
		sharedSessionBrowserRawProfilesDefaultProfile(observation.Profiles),
	)
}

func sharedSessionBrowserRawRequestedProfileCandidates(cacheKey string, requestedProfile string) []string {
	return appendSharedSessionBrowserRawObservationUniqueKeys(nil, cacheKey, requestedProfile)
}

// SharedSessionBrowserRawTabsObservation captures a direct route-scoped tabs
// observation before any shared session registry sync or projection rebuild.
// The observation also carries the tabs action/requested tab index so watch
// recovery can reuse the full tabs-result owner instead of only replaying a
// bare tab sync. Empty tab lists are still meaningful when ObservedAt is
// present because they can represent a detach/close cycle that cleared the
// route.
type SharedSessionBrowserRawTabsObservation struct {
	RequestedProfile       string
	Action                 string
	RequestedTabIndex      int
	Force                  bool
	RememberTarget         bool
	Review                 SharedSessionBrowserPendingTargetReviewState
	Actor                  string
	ExplicitTargetID       string
	PriorSelection         *BrowserSessionTargetSelection
	PriorActiveTargetID    string
	PriorRequestedTargetID string
	Note                   string
	Tabs                   []BrowserTab
	ActiveIndex            int
	ObservedAt             time.Time
}

// SharedSessionBrowserRawNavigationObservation captures a direct route-scoped
// navigation result before any shared session registry sync or projection
// rebuild. The observation preserves the original requested URL together with
// the final URL/title resolved by the runtime backend so redirect review can be
// reconstructed from a source-time backend result.
type SharedSessionBrowserRawNavigationObservation struct {
	RequestedProfile string
	RequestedURL     string
	FinalURL         string
	Title            string
	TabIndex         int
	Force            bool
	ExplicitTargetID string
	PriorSelection   *BrowserSessionTargetSelection
	Note             string
	ObservedAt       time.Time
}

// SharedSessionBrowserRawOpenObservation captures a direct route-scoped open
// result before any shared session registry sync or projection rebuild.
type SharedSessionBrowserRawOpenObservation struct {
	RequestedProfile string
	URL              string
	Title            string
	ObservedAt       time.Time
}

// SharedSessionBrowserRawTargetObservation captures a direct route-scoped
// generic target result before any shared session registry sync or projection
// rebuild. The observation can also carry source-time page-action or generic
// action posture so watch replay can reuse the richer owner instead of
// degrading everything to a bare target/current-target sync.
type SharedSessionBrowserRawTargetObservation struct {
	RequestedProfile  string
	TabIndex          int
	TrackCurrent      bool
	SetCurrent        bool
	URL               string
	Title             string
	Source            string
	PreferredTargetID string
	Actor             string
	Force             bool
	Review            SharedSessionBrowserPendingTargetReviewState
	ReviewDecision    string
	ReviewReady       bool
	Note              string
	ObservedAt        time.Time
}

// SharedSessionBrowserRawRouteMutationObservation captures a direct route-
// scoped mutation event before any shared session registry sync or projection
// rebuild. Unlike cached route-mutation event-cycles, this preserves the
// original attach/detach/current-target mutation payload so watch replay can
// re-apply the owner contract after projection caches have been drained.
type SharedSessionBrowserRawRouteMutationObservation struct {
	RequestedProfile       string
	Route                  BrowserSessionRoute
	Kind                   string
	RequestedURL           string
	Action                 string
	Force                  bool
	Ready                  bool
	RememberTarget         bool
	Review                 SharedSessionBrowserPendingTargetReviewState
	Actor                  string
	ExplicitTargetID       string
	CandidateTargetID      string
	RequestedTabIndex      int
	PriorActiveTargetID    string
	PriorRequestedTargetID string
	ActiveIndex            int
	Tabs                   []BrowserTab
	Tab                    BrowserTab
	SetCurrent             bool
	TabIndex               int
	TargetID               string
	FinalURL               string
	Decision               string
	Reason                 string
	PriorSelection         *BrowserSessionTargetSelection
	PendingTargetID        string
	URL                    string
	Title                  string
	Note                   string
	Source                 string
	ObservedAt             time.Time
}

// BrowserRuntimeRawStatusObservationBackend allows a runtime control backend to
// provide a raw status observation directly, preserving source-time timestamps
// from an event-driven cache instead of forcing the watch manager to poll
// RuntimeStatus itself.
type BrowserRuntimeRawStatusObservationBackend interface {
	ObserveRawBrowserRuntimeStatus(context.Context, string) SharedSessionBrowserRawStatusObservation
}

// BrowserRuntimeRawProfilesObservationBackend allows a runtime control backend
// to provide a raw profiles observation directly, preserving source-time
// timestamps from an event-driven cache instead of forcing the watch manager to
// poll RuntimeProfiles itself.
type BrowserRuntimeRawProfilesObservationBackend interface {
	ObserveRawBrowserRuntimeProfiles(context.Context, string) SharedSessionBrowserRawProfilesObservation
}

// BrowserRuntimeRawLifecycleObservationBackend allows a runtime control backend
// to provide raw start/stop observations directly, preserving source-time
// timestamps from an event-driven cache instead of forcing the watch manager to
// poll RuntimeStart/RuntimeStop itself.
type BrowserRuntimeRawLifecycleObservationBackend interface {
	ObserveRawBrowserRuntimeStart(context.Context, string) SharedSessionBrowserRawLifecycleObservation
	ObserveRawBrowserRuntimeStop(context.Context, string) SharedSessionBrowserRawLifecycleObservation
}

// BrowserRuntimeRawStatusAndProfilesObservationBackend allows a runtime
// control backend to provide a single raw status/profiles observation cycle
// directly, so watch-manager event cycles can reuse an event-driven source of
// truth instead of forcing two independent polling calls.
type BrowserRuntimeRawStatusAndProfilesObservationBackend interface {
	ObserveRawBrowserRuntimeStatusAndProfiles(context.Context, string, bool, bool) SharedSessionBrowserRawStatusAndProfilesObservation
}

// BrowserRuntimeRawTabsObservationBackend allows a runtime backend to provide a
// direct route-scoped tabs observation, so watch-manager event cycles can
// reuse attach/detach source-time state instead of forcing tool-side mutation
// writeback first.
type BrowserRuntimeRawTabsObservationBackend interface {
	ObserveRawBrowserTabs(context.Context, string) SharedSessionBrowserRawTabsObservation
}

// BrowserRuntimeRawNavigationObservationBackend allows a runtime backend to
// provide a direct route-scoped navigation observation, so watch-manager event
// cycles can reuse popup/redirect source-time state instead of forcing
// tool-side result writeback first.
type BrowserRuntimeRawNavigationObservationBackend interface {
	ObserveRawBrowserNavigation(context.Context, string) SharedSessionBrowserRawNavigationObservation
}

// BrowserRuntimeRawOpenObservationBackend allows a runtime backend to provide
// a direct route-scoped open observation, so watch-manager event cycles can
// reuse target-created source-time state instead of forcing tool-side result
// writeback first.
type BrowserRuntimeRawOpenObservationBackend interface {
	ObserveRawBrowserOpen(context.Context, string) SharedSessionBrowserRawOpenObservation
}

// BrowserRuntimeRawTargetObservationBackend allows a runtime backend to
// provide a direct route-scoped generic target observation, so watch-manager
// event cycles can reuse source-time target tracking for console/request/
// response/error style results instead of forcing tool-side writeback first.
type BrowserRuntimeRawTargetObservationBackend interface {
	ObserveRawBrowserTarget(context.Context, string) SharedSessionBrowserRawTargetObservation
}

// BrowserRuntimeRawRouteMutationObservationBackend allows a runtime backend to
// provide a direct route-scoped mutation observation, so watch-manager event
// cycles can replay attach/detach/current-target mutations from a source-time
// backend result instead of relying only on synthetic local mutation caches.
type BrowserRuntimeRawRouteMutationObservationBackend interface {
	ObserveRawBrowserRouteMutation(context.Context, string) SharedSessionBrowserRawRouteMutationObservation
}

func normalizeSharedSessionBrowserRawStatusObservation(observation SharedSessionBrowserRawStatusObservation, requestedProfile string) SharedSessionBrowserRawStatusObservation {
	observation.RequestedProfile = firstNonEmptyString(
		strings.TrimSpace(observation.RequestedProfile),
		strings.TrimSpace(requestedProfile),
	)
	if observation.ObservedAt.IsZero() && (observation.Status != nil || observation.Err != nil) {
		observation.ObservedAt = time.Now()
	}
	return observation
}

func sharedSessionBrowserRawStatusObservationProvided(observation SharedSessionBrowserRawStatusObservation) bool {
	return observation.Status != nil || observation.Err != nil || !observation.ObservedAt.IsZero()
}

// BuildSharedSessionBrowserRawStatusObservation projects a runtime status
// source-time result into the raw status observation shape used by lifecycle
// recovery.
func BuildSharedSessionBrowserRawStatusObservation(
	requestedProfile string,
	status *BrowserProfileStatusResult,
	err error,
	observedAt time.Time,
) (SharedSessionBrowserRawStatusObservation, bool) {
	observation := SharedSessionBrowserRawStatusObservation{
		RequestedProfile: strings.TrimSpace(requestedProfile),
		Err:              err,
		ObservedAt:       observedAt,
	}
	if status != nil {
		statusCopy := *status
		observation.Status = &statusCopy
	}
	observation = normalizeSharedSessionBrowserRawStatusObservation(observation, requestedProfile)
	if !sharedSessionBrowserRawStatusObservationProvided(observation) {
		return SharedSessionBrowserRawStatusObservation{}, false
	}
	return observation, true
}

// FreshSharedSessionBrowserRawStatusObservation returns a replayable raw status
// observation only when the source-time timestamp is still within ttl.
func FreshSharedSessionBrowserRawStatusObservation(
	observation SharedSessionBrowserRawStatusObservation,
	now time.Time,
	ttl time.Duration,
) (SharedSessionBrowserRawStatusObservation, bool) {
	if !sharedSessionBrowserRawStatusObservationProvided(observation) ||
		SharedSessionBrowserRawObservationExpired(observation.ObservedAt, now, ttl) {
		return SharedSessionBrowserRawStatusObservation{}, false
	}
	return normalizeSharedSessionBrowserRawStatusObservation(observation, observation.RequestedProfile), true
}

func normalizeSharedSessionBrowserRawProfilesObservation(observation SharedSessionBrowserRawProfilesObservation, requestedProfile string) SharedSessionBrowserRawProfilesObservation {
	observation.RequestedProfile = firstNonEmptyString(
		strings.TrimSpace(observation.RequestedProfile),
		strings.TrimSpace(requestedProfile),
	)
	if observation.ObservedAt.IsZero() && (observation.Profiles != nil || observation.Err != nil) {
		observation.ObservedAt = time.Now()
	}
	return observation
}

func sharedSessionBrowserRawProfilesObservationProvided(observation SharedSessionBrowserRawProfilesObservation) bool {
	return observation.Profiles != nil || observation.Err != nil || !observation.ObservedAt.IsZero()
}

// BuildSharedSessionBrowserRawProfilesObservation projects a runtime profiles
// source-time result into the raw profiles observation shape used by lifecycle
// recovery.
func BuildSharedSessionBrowserRawProfilesObservation(
	requestedProfile string,
	profiles *BrowserProfilesResult,
	err error,
	observedAt time.Time,
) (SharedSessionBrowserRawProfilesObservation, bool) {
	observation := SharedSessionBrowserRawProfilesObservation{
		RequestedProfile: strings.TrimSpace(requestedProfile),
		Err:              err,
		ObservedAt:       observedAt,
	}
	if profiles != nil {
		profilesCopy := *profiles
		observation.Profiles = &profilesCopy
	}
	observation = normalizeSharedSessionBrowserRawProfilesObservation(observation, requestedProfile)
	if !sharedSessionBrowserRawProfilesObservationProvided(observation) {
		return SharedSessionBrowserRawProfilesObservation{}, false
	}
	return observation, true
}

// FreshSharedSessionBrowserRawProfilesObservation returns a replayable raw
// profiles observation only when the source-time timestamp is still within ttl.
func FreshSharedSessionBrowserRawProfilesObservation(
	observation SharedSessionBrowserRawProfilesObservation,
	now time.Time,
	ttl time.Duration,
) (SharedSessionBrowserRawProfilesObservation, bool) {
	if !sharedSessionBrowserRawProfilesObservationProvided(observation) ||
		SharedSessionBrowserRawObservationExpired(observation.ObservedAt, now, ttl) {
		return SharedSessionBrowserRawProfilesObservation{}, false
	}
	return normalizeSharedSessionBrowserRawProfilesObservation(observation, observation.RequestedProfile), true
}

func normalizeSharedSessionBrowserRawLifecycleObservation(observation SharedSessionBrowserRawLifecycleObservation, profile string) SharedSessionBrowserRawLifecycleObservation {
	observation.Profile = firstNonEmptyString(strings.TrimSpace(observation.Profile), strings.TrimSpace(profile))
	if observation.ObservedAt.IsZero() && (observation.Status != nil || observation.Err != nil) {
		observation.ObservedAt = time.Now()
	}
	return observation
}

func sharedSessionBrowserRawLifecycleObservationProvided(observation SharedSessionBrowserRawLifecycleObservation) bool {
	return observation.Status != nil || observation.Err != nil || !observation.ObservedAt.IsZero()
}

// BuildSharedSessionBrowserRawLifecycleObservation projects a runtime
// start/stop source-time result into the raw lifecycle observation shape used by
// lifecycle recovery.
func BuildSharedSessionBrowserRawLifecycleObservation(
	requestedProfile string,
	resolvedProfile string,
	status *BrowserProfileStatusResult,
	err error,
	observedAt time.Time,
) (SharedSessionBrowserRawLifecycleObservation, bool) {
	profile := firstNonEmptyString(strings.TrimSpace(requestedProfile), strings.TrimSpace(resolvedProfile))
	observation := SharedSessionBrowserRawLifecycleObservation{
		Profile:    profile,
		Err:        err,
		ObservedAt: observedAt,
	}
	if status != nil {
		statusCopy := *status
		observation.Status = &statusCopy
		observation.Profile = firstNonEmptyString(strings.TrimSpace(statusCopy.Profile), observation.Profile)
	}
	observation = normalizeSharedSessionBrowserRawLifecycleObservation(observation, profile)
	if !sharedSessionBrowserRawLifecycleObservationProvided(observation) {
		return SharedSessionBrowserRawLifecycleObservation{}, false
	}
	return observation, true
}

// FreshSharedSessionBrowserRawLifecycleObservation returns a replayable raw
// lifecycle observation only when the source-time timestamp is still within ttl.
func FreshSharedSessionBrowserRawLifecycleObservation(
	observation SharedSessionBrowserRawLifecycleObservation,
	now time.Time,
	ttl time.Duration,
) (SharedSessionBrowserRawLifecycleObservation, bool) {
	if !sharedSessionBrowserRawLifecycleObservationProvided(observation) ||
		SharedSessionBrowserRawObservationExpired(observation.ObservedAt, now, ttl) {
		return SharedSessionBrowserRawLifecycleObservation{}, false
	}
	return normalizeSharedSessionBrowserRawLifecycleObservation(observation, observation.Profile), true
}

func normalizeSharedSessionBrowserRawStatusAndProfilesObservation(
	observation SharedSessionBrowserRawStatusAndProfilesObservation,
	requestedProfile string,
	includeStatus bool,
	includeProfiles bool,
) SharedSessionBrowserRawStatusAndProfilesObservation {
	observation.RequestedProfile = firstNonEmptyString(
		strings.TrimSpace(observation.RequestedProfile),
		strings.TrimSpace(requestedProfile),
	)
	if !includeStatus {
		observation.Status = nil
		observation.StatusErr = nil
		observation.StatusObservedAt = time.Time{}
	} else if observation.StatusObservedAt.IsZero() && (observation.Status != nil || observation.StatusErr != nil) {
		observation.StatusObservedAt = time.Now()
	}
	if !includeProfiles {
		observation.Profiles = nil
		observation.ProfilesErr = nil
		observation.ProfilesObservedAt = time.Time{}
	} else if observation.ProfilesObservedAt.IsZero() && (observation.Profiles != nil || observation.ProfilesErr != nil) {
		observation.ProfilesObservedAt = time.Now()
	}
	return observation
}

func sharedSessionBrowserRawStatusAndProfilesObservationProvided(observation SharedSessionBrowserRawStatusAndProfilesObservation) bool {
	return observation.Status != nil ||
		observation.StatusErr != nil ||
		!observation.StatusObservedAt.IsZero() ||
		observation.Profiles != nil ||
		observation.ProfilesErr != nil ||
		!observation.ProfilesObservedAt.IsZero()
}

// SharedSessionBrowserRawStatusAndProfilesObservationProvided reports whether
// a combined raw status/profiles observation contains source-time state.
func SharedSessionBrowserRawStatusAndProfilesObservationProvided(observation SharedSessionBrowserRawStatusAndProfilesObservation) bool {
	return sharedSessionBrowserRawStatusAndProfilesObservationProvided(observation)
}

// BuildSharedSessionBrowserRawStatusAndProfilesObservation combines raw status
// and profiles source-time observations into the shared combined polling shape.
func BuildSharedSessionBrowserRawStatusAndProfilesObservation(
	statusObservation SharedSessionBrowserRawStatusObservation,
	profilesObservation SharedSessionBrowserRawProfilesObservation,
) (SharedSessionBrowserRawStatusAndProfilesObservation, bool) {
	observation := normalizeSharedSessionBrowserRawStatusAndProfilesObservation(
		SharedSessionBrowserRawStatusAndProfilesObservation{
			RequestedProfile:   firstNonEmptyString(statusObservation.RequestedProfile, profilesObservation.RequestedProfile),
			Status:             statusObservation.Status,
			StatusErr:          statusObservation.Err,
			StatusObservedAt:   statusObservation.ObservedAt,
			Profiles:           profilesObservation.Profiles,
			ProfilesErr:        profilesObservation.Err,
			ProfilesObservedAt: profilesObservation.ObservedAt,
		},
		firstNonEmptyString(statusObservation.RequestedProfile, profilesObservation.RequestedProfile),
		true,
		true,
	)
	if !sharedSessionBrowserRawStatusAndProfilesObservationProvided(observation) {
		return SharedSessionBrowserRawStatusAndProfilesObservation{}, false
	}
	return observation, true
}

// BuildFreshSharedSessionBrowserRawStatusAndProfilesObservation combines raw
// status and profiles observations only when both source-time observations are
// fresh enough to replay.
func BuildFreshSharedSessionBrowserRawStatusAndProfilesObservation(
	statusObservation SharedSessionBrowserRawStatusObservation,
	profilesObservation SharedSessionBrowserRawProfilesObservation,
	now time.Time,
	ttl time.Duration,
) (SharedSessionBrowserRawStatusAndProfilesObservation, bool) {
	if SharedSessionBrowserRawObservationExpired(statusObservation.ObservedAt, now, ttl) ||
		SharedSessionBrowserRawObservationExpired(profilesObservation.ObservedAt, now, ttl) {
		return SharedSessionBrowserRawStatusAndProfilesObservation{}, false
	}
	return BuildSharedSessionBrowserRawStatusAndProfilesObservation(statusObservation, profilesObservation)
}

// BuildSharedSessionBrowserRawStatusAndProfilesObservationForRequest combines
// optional raw status and profiles observations according to the caller's
// requested polling surface.
func BuildSharedSessionBrowserRawStatusAndProfilesObservationForRequest(
	requestedProfile string,
	includeStatus bool,
	statusObservation SharedSessionBrowserRawStatusObservation,
	includeProfiles bool,
	profilesObservation SharedSessionBrowserRawProfilesObservation,
) SharedSessionBrowserRawStatusAndProfilesObservation {
	observation := SharedSessionBrowserRawStatusAndProfilesObservation{
		RequestedProfile: strings.TrimSpace(requestedProfile),
	}
	if includeStatus && sharedSessionBrowserRawStatusObservationProvided(statusObservation) {
		observation.RequestedProfile = firstNonEmptyString(observation.RequestedProfile, statusObservation.RequestedProfile)
		observation.Status = statusObservation.Status
		observation.StatusErr = statusObservation.Err
		observation.StatusObservedAt = statusObservation.ObservedAt
	}
	if includeProfiles && sharedSessionBrowserRawProfilesObservationProvided(profilesObservation) {
		observation.RequestedProfile = firstNonEmptyString(observation.RequestedProfile, profilesObservation.RequestedProfile)
		observation.Profiles = profilesObservation.Profiles
		observation.ProfilesErr = profilesObservation.Err
		observation.ProfilesObservedAt = profilesObservation.ObservedAt
	}
	return normalizeSharedSessionBrowserRawStatusAndProfilesObservation(
		observation,
		requestedProfile,
		includeStatus,
		includeProfiles,
	)
}

// PruneExpiredSharedSessionBrowserRawStatusAndProfilesObservation clears stale
// status or profiles halves from a combined raw source-time observation.
func PruneExpiredSharedSessionBrowserRawStatusAndProfilesObservation(
	observation SharedSessionBrowserRawStatusAndProfilesObservation,
	now time.Time,
	ttl time.Duration,
) SharedSessionBrowserRawStatusAndProfilesObservation {
	if SharedSessionBrowserRawObservationExpired(observation.StatusObservedAt, now, ttl) {
		observation.Status = nil
		observation.StatusErr = nil
		observation.StatusObservedAt = time.Time{}
	}
	if SharedSessionBrowserRawObservationExpired(observation.ProfilesObservedAt, now, ttl) {
		observation.Profiles = nil
		observation.ProfilesErr = nil
		observation.ProfilesObservedAt = time.Time{}
	}
	return observation
}

// FreshSharedSessionBrowserRawStatusAndProfilesObservation returns a combined
// raw status/profiles observation after stale halves have been removed.
func FreshSharedSessionBrowserRawStatusAndProfilesObservation(
	observation SharedSessionBrowserRawStatusAndProfilesObservation,
	now time.Time,
	ttl time.Duration,
) (SharedSessionBrowserRawStatusAndProfilesObservation, bool) {
	observation = PruneExpiredSharedSessionBrowserRawStatusAndProfilesObservation(observation, now, ttl)
	if !sharedSessionBrowserRawStatusAndProfilesObservationProvided(observation) {
		return SharedSessionBrowserRawStatusAndProfilesObservation{}, false
	}
	return normalizeSharedSessionBrowserRawStatusAndProfilesObservation(
		observation,
		observation.RequestedProfile,
		true,
		true,
	), true
}

// SharedSessionBrowserRawStatusAndProfilesObservationKeys returns the profile
// keys that should index a combined raw status/profiles observation.
func SharedSessionBrowserRawStatusAndProfilesObservationKeys(
	statusObservation SharedSessionBrowserRawStatusObservation,
	profilesObservation SharedSessionBrowserRawProfilesObservation,
) []string {
	keys := make([]string, 0, 4)
	appendKeys := func(values ...string) {
		for _, key := range values {
			trimmed := strings.TrimSpace(key)
			if trimmed == "" {
				if len(keys) == 0 {
					keys = append(keys, "")
				}
				continue
			}
			seen := false
			for _, existing := range keys {
				if existing == trimmed {
					seen = true
					break
				}
			}
			if !seen {
				keys = append(keys, trimmed)
			}
		}
	}
	statusProfile := ""
	if statusObservation.Status != nil {
		statusProfile = sharedSessionBrowserRawStatusProfile(statusObservation.Status)
	}
	defaultProfile := ""
	if profilesObservation.Profiles != nil {
		defaultProfile = sharedSessionBrowserRawProfilesDefaultProfile(profilesObservation.Profiles)
	}
	appendKeys(statusObservation.RequestedProfile, statusProfile, profilesObservation.RequestedProfile, defaultProfile)
	if len(keys) == 0 {
		return []string{""}
	}
	return keys
}

func normalizeSharedSessionBrowserRawTabsObservation(
	observation SharedSessionBrowserRawTabsObservation,
	requestedProfile string,
) SharedSessionBrowserRawTabsObservation {
	observation.RequestedProfile = firstNonEmptyString(
		strings.TrimSpace(observation.RequestedProfile),
		strings.TrimSpace(requestedProfile),
	)
	observation.Action = strings.TrimSpace(observation.Action)
	observation.Review.PolicyState = strings.TrimSpace(observation.Review.PolicyState)
	observation.Review.PolicyReason = strings.TrimSpace(observation.Review.PolicyReason)
	observation.Actor = strings.TrimSpace(observation.Actor)
	observation.ExplicitTargetID = strings.TrimSpace(observation.ExplicitTargetID)
	observation.PriorActiveTargetID = strings.TrimSpace(observation.PriorActiveTargetID)
	observation.PriorRequestedTargetID = strings.TrimSpace(observation.PriorRequestedTargetID)
	observation.Note = strings.TrimSpace(observation.Note)
	if observation.Review.Review != nil {
		reviewCopy := *observation.Review.Review
		reviewCopy.ID = strings.TrimSpace(reviewCopy.ID)
		reviewCopy.URL = strings.TrimSpace(reviewCopy.URL)
		reviewCopy.Title = strings.TrimSpace(reviewCopy.Title)
		reviewCopy.Decision = strings.TrimSpace(reviewCopy.Decision)
		reviewCopy.Reason = strings.TrimSpace(reviewCopy.Reason)
		observation.Review.Review = &reviewCopy
	}
	if observation.ObservedAt.IsZero() &&
		(len(observation.Tabs) > 0 ||
			observation.ActiveIndex > 0 ||
			observation.Action != "" ||
			observation.RequestedTabIndex > 0 ||
			observation.Force ||
			observation.RememberTarget ||
			observation.Review.Review != nil ||
			observation.Review.Count > 0 ||
			observation.Review.PolicyState != "" ||
			observation.Review.PolicyReason != "" ||
			observation.Actor != "" ||
			observation.ExplicitTargetID != "" ||
			observation.PriorSelection != nil ||
			observation.PriorActiveTargetID != "" ||
			observation.PriorRequestedTargetID != "" ||
			observation.Note != "") {
		observation.ObservedAt = time.Now()
	}
	if observation.PriorSelection != nil {
		selectionCopy := *observation.PriorSelection
		observation.PriorSelection = &selectionCopy
	}
	if len(observation.Tabs) > 0 {
		observation.Tabs = append([]BrowserTab(nil), observation.Tabs...)
	}
	return observation
}

func sharedSessionBrowserRawTabsObservationProvided(observation SharedSessionBrowserRawTabsObservation) bool {
	return len(observation.Tabs) > 0 ||
		observation.ActiveIndex > 0 ||
		observation.Action != "" ||
		observation.RequestedTabIndex > 0 ||
		observation.Force ||
		observation.RememberTarget ||
		observation.Review.Review != nil ||
		observation.Review.Count > 0 ||
		observation.Review.PolicyState != "" ||
		observation.Review.PolicyReason != "" ||
		observation.Actor != "" ||
		observation.ExplicitTargetID != "" ||
		observation.PriorSelection != nil ||
		observation.PriorActiveTargetID != "" ||
		observation.PriorRequestedTargetID != "" ||
		observation.Note != "" ||
		!observation.ObservedAt.IsZero()
}

// FreshSharedSessionBrowserRawTabsObservation returns a replayable raw tabs
// observation only when the source-time timestamp is still within ttl.
func FreshSharedSessionBrowserRawTabsObservation(
	observation SharedSessionBrowserRawTabsObservation,
	now time.Time,
	ttl time.Duration,
) (SharedSessionBrowserRawTabsObservation, bool) {
	if !sharedSessionBrowserRawTabsObservationProvided(observation) ||
		SharedSessionBrowserRawObservationExpired(observation.ObservedAt, now, ttl) {
		return SharedSessionBrowserRawTabsObservation{}, false
	}
	return normalizeSharedSessionBrowserRawTabsObservation(observation, observation.RequestedProfile), true
}

func normalizeSharedSessionBrowserRawNavigationObservation(
	observation SharedSessionBrowserRawNavigationObservation,
	requestedProfile string,
) SharedSessionBrowserRawNavigationObservation {
	observation.RequestedProfile = firstNonEmptyString(
		strings.TrimSpace(observation.RequestedProfile),
		strings.TrimSpace(requestedProfile),
	)
	observation.RequestedURL = strings.TrimSpace(observation.RequestedURL)
	observation.FinalURL = strings.TrimSpace(observation.FinalURL)
	observation.Title = strings.TrimSpace(observation.Title)
	observation.ExplicitTargetID = strings.TrimSpace(observation.ExplicitTargetID)
	observation.Note = strings.TrimSpace(observation.Note)
	if observation.ObservedAt.IsZero() &&
		(observation.RequestedURL != "" ||
			observation.FinalURL != "" ||
			observation.Title != "" ||
			observation.TabIndex > 0 ||
			observation.Force ||
			observation.ExplicitTargetID != "" ||
			observation.PriorSelection != nil ||
			observation.Note != "") {
		observation.ObservedAt = time.Now()
	}
	if observation.PriorSelection != nil {
		selectionCopy := *observation.PriorSelection
		observation.PriorSelection = &selectionCopy
	}
	return observation
}

func sharedSessionBrowserRawNavigationObservationProvided(observation SharedSessionBrowserRawNavigationObservation) bool {
	return observation.RequestedURL != "" ||
		observation.FinalURL != "" ||
		observation.Title != "" ||
		observation.TabIndex > 0 ||
		observation.Force ||
		observation.ExplicitTargetID != "" ||
		observation.PriorSelection != nil ||
		observation.Note != "" ||
		!observation.ObservedAt.IsZero()
}

// FreshSharedSessionBrowserRawNavigationObservation returns a replayable raw
// navigation observation only when the source-time timestamp is still within ttl.
func FreshSharedSessionBrowserRawNavigationObservation(
	observation SharedSessionBrowserRawNavigationObservation,
	now time.Time,
	ttl time.Duration,
) (SharedSessionBrowserRawNavigationObservation, bool) {
	if !sharedSessionBrowserRawNavigationObservationProvided(observation) ||
		SharedSessionBrowserRawObservationExpired(observation.ObservedAt, now, ttl) {
		return SharedSessionBrowserRawNavigationObservation{}, false
	}
	return normalizeSharedSessionBrowserRawNavigationObservation(observation, observation.RequestedProfile), true
}

func normalizeSharedSessionBrowserRawOpenObservation(
	observation SharedSessionBrowserRawOpenObservation,
	requestedProfile string,
) SharedSessionBrowserRawOpenObservation {
	observation.RequestedProfile = firstNonEmptyString(
		strings.TrimSpace(observation.RequestedProfile),
		strings.TrimSpace(requestedProfile),
	)
	observation.URL = strings.TrimSpace(observation.URL)
	observation.Title = strings.TrimSpace(observation.Title)
	if observation.ObservedAt.IsZero() && (observation.URL != "" || observation.Title != "") {
		observation.ObservedAt = time.Now()
	}
	return observation
}

func sharedSessionBrowserRawOpenObservationProvided(observation SharedSessionBrowserRawOpenObservation) bool {
	return observation.URL != "" || observation.Title != "" || !observation.ObservedAt.IsZero()
}

// FreshSharedSessionBrowserRawOpenObservation returns a replayable raw open
// observation only when the source-time timestamp is still within ttl.
func FreshSharedSessionBrowserRawOpenObservation(
	observation SharedSessionBrowserRawOpenObservation,
	now time.Time,
	ttl time.Duration,
) (SharedSessionBrowserRawOpenObservation, bool) {
	if !sharedSessionBrowserRawOpenObservationProvided(observation) ||
		SharedSessionBrowserRawObservationExpired(observation.ObservedAt, now, ttl) {
		return SharedSessionBrowserRawOpenObservation{}, false
	}
	return normalizeSharedSessionBrowserRawOpenObservation(observation, observation.RequestedProfile), true
}

func normalizeSharedSessionBrowserRawTargetObservation(
	observation SharedSessionBrowserRawTargetObservation,
	requestedProfile string,
) SharedSessionBrowserRawTargetObservation {
	observation.RequestedProfile = firstNonEmptyString(
		strings.TrimSpace(observation.RequestedProfile),
		strings.TrimSpace(requestedProfile),
	)
	observation.URL = strings.TrimSpace(observation.URL)
	observation.Title = strings.TrimSpace(observation.Title)
	observation.Source = strings.TrimSpace(observation.Source)
	observation.PreferredTargetID = strings.TrimSpace(observation.PreferredTargetID)
	observation.Actor = strings.TrimSpace(observation.Actor)
	observation.ReviewDecision = strings.TrimSpace(observation.ReviewDecision)
	observation.Note = strings.TrimSpace(observation.Note)
	observation.Review.PolicyState = strings.TrimSpace(observation.Review.PolicyState)
	observation.Review.PolicyReason = strings.TrimSpace(observation.Review.PolicyReason)
	if observation.Review.Review != nil {
		reviewCopy := *observation.Review.Review
		observation.Review.Review = &reviewCopy
	}
	if observation.ObservedAt.IsZero() &&
		(observation.TabIndex > 0 ||
			observation.TrackCurrent ||
			observation.SetCurrent ||
			observation.URL != "" ||
			observation.Title != "" ||
			observation.Source != "" ||
			observation.PreferredTargetID != "" ||
			observation.Actor != "" ||
			observation.Force ||
			observation.ReviewDecision != "" ||
			observation.ReviewReady ||
			observation.Note != "" ||
			observation.Review.Review != nil ||
			observation.Review.Count > 0 ||
			observation.Review.PolicyState != "" ||
			observation.Review.PolicyReason != "") {
		observation.ObservedAt = time.Now()
	}
	return observation
}

func sharedSessionBrowserRawTargetObservationProvided(observation SharedSessionBrowserRawTargetObservation) bool {
	return observation.TabIndex > 0 ||
		observation.TrackCurrent ||
		observation.SetCurrent ||
		observation.URL != "" ||
		observation.Title != "" ||
		observation.Source != "" ||
		observation.PreferredTargetID != "" ||
		observation.Actor != "" ||
		observation.Force ||
		observation.ReviewDecision != "" ||
		observation.ReviewReady ||
		observation.Note != "" ||
		observation.Review.Review != nil ||
		observation.Review.Count > 0 ||
		observation.Review.PolicyState != "" ||
		observation.Review.PolicyReason != "" ||
		!observation.ObservedAt.IsZero()
}

// NormalizeSharedSessionBrowserRawTargetObservation returns the canonical raw
// target source-time observation shape before cache or route-mutation replay
// handoff.
func NormalizeSharedSessionBrowserRawTargetObservation(
	observation SharedSessionBrowserRawTargetObservation,
	requestedProfile string,
) SharedSessionBrowserRawTargetObservation {
	return normalizeSharedSessionBrowserRawTargetObservation(observation, requestedProfile)
}

// SharedSessionBrowserRawTargetObservationProvided reports whether a raw target
// observation contains replayable source-time state.
func SharedSessionBrowserRawTargetObservationProvided(observation SharedSessionBrowserRawTargetObservation) bool {
	observation = normalizeSharedSessionBrowserRawTargetObservation(observation, observation.RequestedProfile)
	return sharedSessionBrowserRawTargetObservationProvided(observation)
}

// FreshSharedSessionBrowserRawTargetObservation returns a replayable raw target
// observation only when the source-time timestamp is still within ttl.
func FreshSharedSessionBrowserRawTargetObservation(
	observation SharedSessionBrowserRawTargetObservation,
	now time.Time,
	ttl time.Duration,
) (SharedSessionBrowserRawTargetObservation, bool) {
	if !sharedSessionBrowserRawTargetObservationProvided(observation) ||
		SharedSessionBrowserRawObservationExpired(observation.ObservedAt, now, ttl) {
		return SharedSessionBrowserRawTargetObservation{}, false
	}
	return normalizeSharedSessionBrowserRawTargetObservation(observation, observation.RequestedProfile), true
}

// BuildSharedSessionBrowserRawTargetObservation projects a generic target
// source-time result into the raw target observation shape used by
// watch-manager recovery.
func BuildSharedSessionBrowserRawTargetObservation(
	requestedProfile string,
	tabIndex int,
	trackCurrent bool,
	setCurrent bool,
	finalURL string,
	title string,
	source string,
	preferredTargetID string,
	actor string,
	force bool,
	review SharedSessionBrowserPendingTargetReviewState,
	reviewDecision string,
	reviewReady bool,
	note string,
	observedAt time.Time,
) (SharedSessionBrowserRawTargetObservation, bool) {
	observation := SharedSessionBrowserRawTargetObservation{
		RequestedProfile:  requestedProfile,
		TabIndex:          tabIndex,
		TrackCurrent:      trackCurrent,
		SetCurrent:        setCurrent,
		URL:               finalURL,
		Title:             title,
		Source:            source,
		PreferredTargetID: preferredTargetID,
		Actor:             actor,
		Force:             force,
		Review:            review,
		ReviewDecision:    reviewDecision,
		ReviewReady:       reviewReady,
		Note:              note,
		ObservedAt:        observedAt,
	}
	if !observation.TrackCurrent && observation.TabIndex <= 0 && !observation.SetCurrent {
		return SharedSessionBrowserRawTargetObservation{}, false
	}
	observation = normalizeSharedSessionBrowserRawTargetObservation(observation, requestedProfile)
	if !sharedSessionBrowserRawTargetObservationProvided(observation) {
		return SharedSessionBrowserRawTargetObservation{}, false
	}
	return observation, true
}

func normalizeSharedSessionBrowserRawRouteMutationObservation(
	observation SharedSessionBrowserRawRouteMutationObservation,
	requestedProfile string,
) SharedSessionBrowserRawRouteMutationObservation {
	observation.RequestedProfile = firstNonEmptyString(
		strings.TrimSpace(observation.RequestedProfile),
		strings.TrimSpace(requestedProfile),
	)
	observation.Route = normalizeBrowserSessionRoute(observation.Route)
	observation.Kind = strings.TrimSpace(observation.Kind)
	observation.RequestedURL = strings.TrimSpace(observation.RequestedURL)
	observation.Action = strings.TrimSpace(observation.Action)
	observation.Review.PolicyState = strings.TrimSpace(observation.Review.PolicyState)
	observation.Review.PolicyReason = strings.TrimSpace(observation.Review.PolicyReason)
	observation.Actor = strings.TrimSpace(observation.Actor)
	observation.ExplicitTargetID = strings.TrimSpace(observation.ExplicitTargetID)
	observation.CandidateTargetID = strings.TrimSpace(observation.CandidateTargetID)
	observation.PriorActiveTargetID = strings.TrimSpace(observation.PriorActiveTargetID)
	observation.PriorRequestedTargetID = strings.TrimSpace(observation.PriorRequestedTargetID)
	observation.TargetID = strings.TrimSpace(observation.TargetID)
	observation.FinalURL = strings.TrimSpace(observation.FinalURL)
	observation.Decision = strings.TrimSpace(observation.Decision)
	observation.Reason = strings.TrimSpace(observation.Reason)
	observation.PendingTargetID = strings.TrimSpace(observation.PendingTargetID)
	observation.URL = strings.TrimSpace(observation.URL)
	observation.Title = strings.TrimSpace(observation.Title)
	observation.Note = strings.TrimSpace(observation.Note)
	observation.Source = strings.TrimSpace(observation.Source)
	if observation.Review.Review != nil {
		reviewCopy := *observation.Review.Review
		reviewCopy.ID = strings.TrimSpace(reviewCopy.ID)
		reviewCopy.URL = strings.TrimSpace(reviewCopy.URL)
		reviewCopy.Title = strings.TrimSpace(reviewCopy.Title)
		reviewCopy.Decision = strings.TrimSpace(reviewCopy.Decision)
		reviewCopy.Reason = strings.TrimSpace(reviewCopy.Reason)
		observation.Review.Review = &reviewCopy
	}
	if observation.PriorSelection != nil {
		copied := *observation.PriorSelection
		observation.PriorSelection = &copied
	}
	if observation.ObservedAt.IsZero() &&
		(observation.Kind != "" ||
			observation.RequestedURL != "" ||
			observation.Action != "" ||
			observation.Force ||
			observation.Ready ||
			observation.RememberTarget ||
			observation.Review.Review != nil ||
			observation.Review.Count > 0 ||
			observation.Review.PolicyState != "" ||
			observation.Review.PolicyReason != "" ||
			observation.Actor != "" ||
			observation.ExplicitTargetID != "" ||
			observation.CandidateTargetID != "" ||
			observation.RequestedTabIndex > 0 ||
			observation.PriorActiveTargetID != "" ||
			observation.PriorRequestedTargetID != "" ||
			observation.ActiveIndex > 0 ||
			len(observation.Tabs) > 0 ||
			observation.TabIndex > 0 ||
			observation.SetCurrent ||
			observation.Tab.Index > 0 ||
			observation.TargetID != "" ||
			observation.FinalURL != "" ||
			observation.Decision != "" ||
			observation.Reason != "" ||
			observation.PriorSelection != nil ||
			observation.PendingTargetID != "" ||
			observation.URL != "" ||
			observation.Title != "" ||
			observation.Note != "" ||
			observation.Source != "") {
		observation.ObservedAt = time.Now()
	}
	if len(observation.Tabs) > 0 {
		observation.Tabs = append([]BrowserTab(nil), observation.Tabs...)
	}
	return observation
}

func sharedSessionBrowserRawRouteMutationObservationProvided(observation SharedSessionBrowserRawRouteMutationObservation) bool {
	return observation.Kind != "" ||
		observation.RequestedURL != "" ||
		observation.Action != "" ||
		observation.Force ||
		observation.Ready ||
		observation.RememberTarget ||
		observation.Review.Review != nil ||
		observation.Review.Count > 0 ||
		observation.Review.PolicyState != "" ||
		observation.Review.PolicyReason != "" ||
		observation.Actor != "" ||
		observation.ExplicitTargetID != "" ||
		observation.CandidateTargetID != "" ||
		observation.RequestedTabIndex > 0 ||
		observation.PriorActiveTargetID != "" ||
		observation.PriorRequestedTargetID != "" ||
		observation.ActiveIndex > 0 ||
		len(observation.Tabs) > 0 ||
		observation.TabIndex > 0 ||
		observation.SetCurrent ||
		observation.Tab.Index > 0 ||
		observation.TargetID != "" ||
		observation.FinalURL != "" ||
		observation.Decision != "" ||
		observation.Reason != "" ||
		observation.PriorSelection != nil ||
		observation.PendingTargetID != "" ||
		observation.URL != "" ||
		observation.Title != "" ||
		observation.Note != "" ||
		observation.Source != "" ||
		!observation.ObservedAt.IsZero()
}

// NormalizeSharedSessionBrowserRawRouteMutationObservation returns the
// canonical raw route-mutation source-time observation shape before cache or
// watch replay handoff.
func NormalizeSharedSessionBrowserRawRouteMutationObservation(
	observation SharedSessionBrowserRawRouteMutationObservation,
	requestedProfile string,
) SharedSessionBrowserRawRouteMutationObservation {
	return normalizeSharedSessionBrowserRawRouteMutationObservation(observation, requestedProfile)
}

// SharedSessionBrowserRawRouteMutationObservationProvided reports whether a
// raw route mutation observation contains replayable source-time state.
func SharedSessionBrowserRawRouteMutationObservationProvided(observation SharedSessionBrowserRawRouteMutationObservation) bool {
	observation = normalizeSharedSessionBrowserRawRouteMutationObservation(observation, observation.RequestedProfile)
	return sharedSessionBrowserRawRouteMutationObservationProvided(observation)
}

// FreshSharedSessionBrowserRawRouteMutationObservation returns a replayable raw
// route-mutation observation only when the source-time timestamp is still within ttl.
func FreshSharedSessionBrowserRawRouteMutationObservation(
	observation SharedSessionBrowserRawRouteMutationObservation,
	now time.Time,
	ttl time.Duration,
) (SharedSessionBrowserRawRouteMutationObservation, bool) {
	if !sharedSessionBrowserRawRouteMutationObservationProvided(observation) ||
		SharedSessionBrowserRawObservationExpired(observation.ObservedAt, now, ttl) {
		return SharedSessionBrowserRawRouteMutationObservation{}, false
	}
	return normalizeSharedSessionBrowserRawRouteMutationObservation(observation, observation.RequestedProfile), true
}

// BuildSharedSessionBrowserRawTabsObservation projects a tabs source-time
// result into the raw tabs observation shape used by watch-manager recovery.
func BuildSharedSessionBrowserRawTabsObservation(
	requestedProfile string,
	req *BrowserTabsRequest,
	result *BrowserTabsResult,
	observedAt time.Time,
) (SharedSessionBrowserRawTabsObservation, bool) {
	if result == nil {
		return SharedSessionBrowserRawTabsObservation{}, false
	}
	observation := normalizeSharedSessionBrowserRawTabsObservation(
		SharedSessionBrowserRawTabsObservation{
			RequestedProfile:       strings.TrimSpace(requestedProfile),
			Action:                 firstNonEmptyString(result.Action, sharedSessionBrowserTabsRequestAction(req)),
			RequestedTabIndex:      sharedSessionBrowserTabsRequestTabIndex(req),
			Force:                  sharedSessionBrowserTabsRequestForce(req),
			RememberTarget:         sharedSessionBrowserTabsRequestRememberTarget(req),
			Review:                 sharedSessionBrowserTabsRequestReview(req),
			Actor:                  sharedSessionBrowserTabsRequestActor(req),
			ExplicitTargetID:       sharedSessionBrowserTabsRequestExplicitTargetID(req),
			PriorSelection:         sharedSessionBrowserTabsRequestPriorSelection(req),
			PriorActiveTargetID:    sharedSessionBrowserTabsRequestPriorActiveTargetID(req),
			PriorRequestedTargetID: sharedSessionBrowserTabsRequestPriorRequestedTargetID(req),
			Note:                   result.Note,
			Tabs:                   result.Tabs,
			ActiveIndex:            result.ActiveIndex,
			ObservedAt:             observedAt,
		},
		requestedProfile,
	)
	if !sharedSessionBrowserRawTabsObservationProvided(observation) {
		return SharedSessionBrowserRawTabsObservation{}, false
	}
	return observation, true
}

// BuildSharedSessionBrowserRawTabsRouteMutationObservation projects a tabs
// source-time result into the route-mutation replay shape used by watch-manager
// recovery.
func BuildSharedSessionBrowserRawTabsRouteMutationObservation(
	requestedProfile string,
	route BrowserSessionRoute,
	req *BrowserTabsRequest,
	result *BrowserTabsResult,
	observedAt time.Time,
) (SharedSessionBrowserRawRouteMutationObservation, bool) {
	if result == nil {
		return SharedSessionBrowserRawRouteMutationObservation{}, false
	}
	route = normalizeBrowserSessionRoute(route)
	route.BrowserApp = firstNonEmptyString(route.BrowserApp, strings.TrimSpace(result.BrowserApp))
	observation := normalizeSharedSessionBrowserRawRouteMutationObservation(
		SharedSessionBrowserRawRouteMutationObservation{
			RequestedProfile:       strings.TrimSpace(requestedProfile),
			Route:                  route,
			Kind:                   "tabs_result",
			Action:                 firstNonEmptyString(result.Action, sharedSessionBrowserTabsRequestAction(req)),
			Force:                  sharedSessionBrowserTabsRequestForce(req),
			RememberTarget:         sharedSessionBrowserTabsRequestRememberTarget(req),
			Review:                 sharedSessionBrowserTabsRequestReview(req),
			Actor:                  sharedSessionBrowserTabsRequestActor(req),
			ExplicitTargetID:       sharedSessionBrowserTabsRequestExplicitTargetID(req),
			CandidateTargetID:      sharedSessionBrowserRawTabsCandidateTargetID(req, result),
			RequestedTabIndex:      sharedSessionBrowserTabsRequestTabIndex(req),
			PriorSelection:         sharedSessionBrowserTabsRequestPriorSelection(req),
			PriorActiveTargetID:    sharedSessionBrowserTabsRequestPriorActiveTargetID(req),
			PriorRequestedTargetID: sharedSessionBrowserTabsRequestPriorRequestedTargetID(req),
			ActiveIndex:            result.ActiveIndex,
			Tabs:                   result.Tabs,
			Note:                   result.Note,
			ObservedAt:             observedAt,
		},
		requestedProfile,
	)
	if !sharedSessionBrowserRawRouteMutationObservationProvided(observation) {
		return SharedSessionBrowserRawRouteMutationObservation{}, false
	}
	return observation, true
}

// BuildSharedSessionBrowserRawNavigationObservation projects a navigation
// source-time result into the raw navigation observation shape used by
// watch-manager recovery.
func BuildSharedSessionBrowserRawNavigationObservation(
	requestedProfile string,
	req *BrowserNavigateRequest,
	result *BrowserNavigateResult,
	observedAt time.Time,
) (SharedSessionBrowserRawNavigationObservation, bool) {
	if result == nil {
		return SharedSessionBrowserRawNavigationObservation{}, false
	}
	observation := SharedSessionBrowserRawNavigationObservation{
		RequestedProfile: strings.TrimSpace(requestedProfile),
		FinalURL:         strings.TrimSpace(result.FinalURL),
		Title:            strings.TrimSpace(result.Title),
		Note:             strings.TrimSpace(result.Note),
		ObservedAt:       observedAt,
	}
	if req != nil {
		observation.RequestedURL = strings.TrimSpace(req.URL)
		observation.TabIndex = req.TabIndex
		observation.Force = req.Force
		observation.ExplicitTargetID = strings.TrimSpace(req.ExplicitTargetID)
		if req.PriorSelection != nil {
			selectionCopy := *req.PriorSelection
			observation.PriorSelection = &selectionCopy
		}
	}
	if observation.FinalURL == "" {
		observation.FinalURL = observation.RequestedURL
	}
	observation = normalizeSharedSessionBrowserRawNavigationObservation(observation, requestedProfile)
	if !sharedSessionBrowserRawNavigationObservationProvided(observation) {
		return SharedSessionBrowserRawNavigationObservation{}, false
	}
	return observation, true
}

// BuildSharedSessionBrowserRawNavigationRouteMutationObservation projects a raw
// navigation source-time observation into the route-mutation replay shape used
// by watch-manager recovery.
func BuildSharedSessionBrowserRawNavigationRouteMutationObservation(
	observation SharedSessionBrowserRawNavigationObservation,
	route BrowserSessionRoute,
) (SharedSessionBrowserRawRouteMutationObservation, bool) {
	observation = normalizeSharedSessionBrowserRawNavigationObservation(observation, observation.RequestedProfile)
	if observation.FinalURL == "" {
		observation.FinalURL = observation.RequestedURL
	}
	if observation.RequestedURL == "" && observation.FinalURL == "" && observation.Title == "" {
		return SharedSessionBrowserRawRouteMutationObservation{}, false
	}
	routeMutation := normalizeSharedSessionBrowserRawRouteMutationObservation(
		SharedSessionBrowserRawRouteMutationObservation{
			RequestedProfile: observation.RequestedProfile,
			Route:            normalizeBrowserSessionRoute(route),
			Kind:             "navigation_result",
			RequestedURL:     observation.RequestedURL,
			TargetID:         observation.ExplicitTargetID,
			TabIndex:         observation.TabIndex,
			URL:              firstNonEmptyString(observation.FinalURL, observation.RequestedURL),
			FinalURL:         observation.FinalURL,
			Title:            observation.Title,
			Force:            observation.Force,
			PriorSelection:   observation.PriorSelection,
			Note:             observation.Note,
			Source:           "runtime_navigation_source",
			ObservedAt:       observation.ObservedAt,
		},
		observation.RequestedProfile,
	)
	if !sharedSessionBrowserRawRouteMutationObservationProvided(routeMutation) {
		return SharedSessionBrowserRawRouteMutationObservation{}, false
	}
	return routeMutation, true
}

// BuildSharedSessionBrowserRawOpenObservation projects an open source-time
// result into the raw open observation shape used by watch-manager recovery.
func BuildSharedSessionBrowserRawOpenObservation(
	requestedProfile string,
	requestedURL string,
	result *BrowserOpenResult,
	observedAt time.Time,
) (SharedSessionBrowserRawOpenObservation, bool) {
	if result == nil {
		return SharedSessionBrowserRawOpenObservation{}, false
	}
	observation := normalizeSharedSessionBrowserRawOpenObservation(
		SharedSessionBrowserRawOpenObservation{
			RequestedProfile: strings.TrimSpace(requestedProfile),
			URL:              strings.TrimSpace(requestedURL),
			ObservedAt:       observedAt,
		},
		requestedProfile,
	)
	if !sharedSessionBrowserRawOpenObservationProvided(observation) {
		return SharedSessionBrowserRawOpenObservation{}, false
	}
	return observation, true
}

// BuildSharedSessionBrowserRawOpenRouteMutationObservation projects a raw open
// source-time observation into the route-mutation replay shape used by
// watch-manager recovery.
func BuildSharedSessionBrowserRawOpenRouteMutationObservation(
	observation SharedSessionBrowserRawOpenObservation,
	route BrowserSessionRoute,
) (SharedSessionBrowserRawRouteMutationObservation, bool) {
	observation = normalizeSharedSessionBrowserRawOpenObservation(observation, observation.RequestedProfile)
	if observation.URL == "" {
		return SharedSessionBrowserRawRouteMutationObservation{}, false
	}
	routeMutation := normalizeSharedSessionBrowserRawRouteMutationObservation(
		SharedSessionBrowserRawRouteMutationObservation{
			RequestedProfile: observation.RequestedProfile,
			Route:            normalizeBrowserSessionRoute(route),
			Kind:             "open_result",
			URL:              observation.URL,
			Title:            observation.Title,
			Source:           "runtime_open_source",
			ObservedAt:       observation.ObservedAt,
		},
		observation.RequestedProfile,
	)
	if !sharedSessionBrowserRawRouteMutationObservationProvided(routeMutation) {
		return SharedSessionBrowserRawRouteMutationObservation{}, false
	}
	return routeMutation, true
}

// BuildSharedSessionBrowserRawTargetRouteMutationObservation projects a raw
// target source-time observation into the route-mutation replay shape used by
// watch-manager recovery.
func BuildSharedSessionBrowserRawTargetRouteMutationObservation(
	observation SharedSessionBrowserRawTargetObservation,
	route BrowserSessionRoute,
) (SharedSessionBrowserRawRouteMutationObservation, bool) {
	observation = normalizeSharedSessionBrowserRawTargetObservation(observation, observation.RequestedProfile)
	if !sharedSessionBrowserRawTargetObservationProvided(observation) {
		return SharedSessionBrowserRawRouteMutationObservation{}, false
	}
	if observation.TabIndex <= 0 && !observation.TrackCurrent && !observation.SetCurrent {
		return SharedSessionBrowserRawRouteMutationObservation{}, false
	}
	routeMutation := SharedSessionBrowserRawRouteMutationObservation{
		RequestedProfile: observation.RequestedProfile,
		Route:            normalizeBrowserSessionRoute(route),
		Source:           observation.Source,
		ObservedAt:       observation.ObservedAt,
	}
	if observation.Actor != "" ||
		observation.Force ||
		observation.Review.Review != nil ||
		observation.Review.Count > 0 ||
		observation.Review.PolicyState != "" ||
		observation.Review.PolicyReason != "" {
		routeMutation.Kind = "page_action_result"
		routeMutation.TargetID = observation.PreferredTargetID
		routeMutation.TabIndex = observation.TabIndex
		routeMutation.SetCurrent = observation.SetCurrent
		routeMutation.URL = observation.URL
		routeMutation.Title = observation.Title
		routeMutation.Actor = observation.Actor
		routeMutation.Force = observation.Force
		routeMutation.Review = observation.Review
	} else if observation.PreferredTargetID != "" ||
		observation.ReviewDecision != "" ||
		observation.ReviewReady ||
		observation.Note != "" {
		routeMutation.Kind = "action_result"
		routeMutation.TargetID = observation.PreferredTargetID
		routeMutation.TabIndex = observation.TabIndex
		routeMutation.SetCurrent = observation.SetCurrent
		routeMutation.URL = observation.URL
		routeMutation.Title = observation.Title
		routeMutation.Decision = observation.ReviewDecision
		routeMutation.Ready = observation.ReviewReady
		routeMutation.Note = observation.Note
	} else if observation.TabIndex > 0 {
		routeMutation.Kind = "track_tab"
		routeMutation.SetCurrent = observation.SetCurrent
		routeMutation.Tab = BrowserTab{
			Index: observation.TabIndex,
			URL:   observation.URL,
			Title: observation.Title,
		}
	} else if observation.TrackCurrent || observation.SetCurrent {
		if observation.URL != "" || observation.Title != "" {
			routeMutation.Kind = "track_current"
			routeMutation.URL = observation.URL
			routeMutation.Title = observation.Title
		}
	}
	routeMutation = normalizeSharedSessionBrowserRawRouteMutationObservation(routeMutation, observation.RequestedProfile)
	if !sharedSessionBrowserRawRouteMutationObservationProvided(routeMutation) {
		return SharedSessionBrowserRawRouteMutationObservation{}, false
	}
	return routeMutation, true
}

func sharedSessionBrowserRawTabsTargetIDByIndex(tabs []BrowserTab, index int) string {
	for _, tab := range tabs {
		if tab.Index == index {
			return strings.TrimSpace(tab.TargetID)
		}
	}
	return ""
}

func sharedSessionBrowserRawTabsCandidateTargetID(req *BrowserTabsRequest, result *BrowserTabsResult) string {
	targetID := sharedSessionBrowserTabsRequestExplicitTargetID(req)
	if targetID == "" {
		targetID = sharedSessionBrowserTabsRequestPriorRequestedTargetID(req)
	}
	if result == nil {
		return strings.TrimSpace(targetID)
	}
	action := strings.TrimSpace(firstNonEmptyString(result.Action, sharedSessionBrowserTabsRequestAction(req)))
	switch action {
	case "list":
		if targetID == "" && result.ActiveIndex > 0 {
			targetID = sharedSessionBrowserRawTabsTargetIDByIndex(result.Tabs, result.ActiveIndex)
		}
	case "focus":
		requestedTabIndex := sharedSessionBrowserTabsRequestTabIndex(req)
		if targetID == "" && requestedTabIndex > 0 {
			targetID = sharedSessionBrowserRawTabsTargetIDByIndex(result.Tabs, requestedTabIndex)
		}
		if targetID == "" && result.ActiveIndex > 0 {
			targetID = sharedSessionBrowserRawTabsTargetIDByIndex(result.Tabs, result.ActiveIndex)
		}
	case "close":
		if targetID == "" {
			targetID = sharedSessionBrowserTabsRequestPriorRequestedTargetID(req)
		}
	}
	return strings.TrimSpace(targetID)
}

func sharedSessionBrowserTabsRequestAction(req *BrowserTabsRequest) string {
	if req == nil {
		return ""
	}
	return req.Action
}

func sharedSessionBrowserTabsRequestTabIndex(req *BrowserTabsRequest) int {
	if req == nil {
		return 0
	}
	return req.TabIndex
}

func sharedSessionBrowserTabsRequestForce(req *BrowserTabsRequest) bool {
	if req == nil {
		return false
	}
	return req.Force
}

func sharedSessionBrowserTabsRequestRememberTarget(req *BrowserTabsRequest) bool {
	if req == nil {
		return false
	}
	return req.RememberTarget
}

func sharedSessionBrowserTabsRequestReview(req *BrowserTabsRequest) SharedSessionBrowserPendingTargetReviewState {
	if req == nil {
		return SharedSessionBrowserPendingTargetReviewState{}
	}
	return req.Review
}

func sharedSessionBrowserTabsRequestActor(req *BrowserTabsRequest) string {
	if req == nil {
		return ""
	}
	return req.Actor
}

func sharedSessionBrowserTabsRequestExplicitTargetID(req *BrowserTabsRequest) string {
	if req == nil {
		return ""
	}
	return strings.TrimSpace(req.ExplicitTargetID)
}

func sharedSessionBrowserTabsRequestPriorSelection(req *BrowserTabsRequest) *BrowserSessionTargetSelection {
	if req == nil || req.PriorSelection == nil {
		return nil
	}
	selectionCopy := *req.PriorSelection
	return &selectionCopy
}

func sharedSessionBrowserTabsRequestPriorActiveTargetID(req *BrowserTabsRequest) string {
	if req == nil {
		return ""
	}
	return strings.TrimSpace(req.PriorActiveTargetID)
}

func sharedSessionBrowserTabsRequestPriorRequestedTargetID(req *BrowserTabsRequest) string {
	if req == nil {
		return ""
	}
	return strings.TrimSpace(req.PriorRequestedTargetID)
}

// ObserveSharedSessionBrowserRawStatus loads the raw RuntimeStatus source of
// truth used by higher-level observation and execution helpers.
func ObserveSharedSessionBrowserRawStatus(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawStatusObservation {
	return sharedSessionBrowserObserverManager(nil, nil, nil, 0).
		ObserveRawStatus(ctx, control, requestedProfile)
}

// ObserveSharedSessionBrowserRawProfiles loads the raw RuntimeProfiles source
// of truth used by higher-level observation and selection helpers.
func ObserveSharedSessionBrowserRawProfiles(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawProfilesObservation {
	return sharedSessionBrowserObserverManager(nil, nil, nil, 0).
		ObserveRawProfiles(ctx, control, requestedProfile)
}

// ObserveSharedSessionBrowserRawStart loads the raw RuntimeStart source of
// truth used by higher-level lifecycle execution helpers.
func ObserveSharedSessionBrowserRawStart(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	profile string,
) SharedSessionBrowserRawLifecycleObservation {
	return sharedSessionBrowserObserverManager(nil, nil, nil, 0).
		ObserveRawStart(ctx, control, profile)
}

// ObserveSharedSessionBrowserRawStop loads the raw RuntimeStop source of truth
// used by higher-level lifecycle execution helpers.
func ObserveSharedSessionBrowserRawStop(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	profile string,
) SharedSessionBrowserRawLifecycleObservation {
	return sharedSessionBrowserObserverManager(nil, nil, nil, 0).
		ObserveRawStop(ctx, control, profile)
}

// ObserveSharedSessionBrowserRawStatusAndProfiles loads a single raw polling
// cycle for optional RuntimeStatus and RuntimeProfiles sources.
func ObserveSharedSessionBrowserRawStatusAndProfiles(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
	includeStatus bool,
	includeProfiles bool,
) SharedSessionBrowserRawStatusAndProfilesObservation {
	return sharedSessionBrowserObserverManager(nil, nil, nil, 0).
		ObserveRawStatusAndProfiles(ctx, control, requestedProfile, includeStatus, includeProfiles)
}

// ObserveSharedSessionBrowserRawTabs loads a direct route-scoped tabs source of
// truth used by watch-manager mutation recovery helpers.
func ObserveSharedSessionBrowserRawTabs(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawTabsObservation {
	requestedProfile = strings.TrimSpace(requestedProfile)
	if control == nil {
		return SharedSessionBrowserRawTabsObservation{RequestedProfile: requestedProfile}
	}
	source, ok := control.(BrowserRuntimeRawTabsObservationBackend)
	if !ok {
		return SharedSessionBrowserRawTabsObservation{RequestedProfile: requestedProfile}
	}
	return normalizeSharedSessionBrowserRawTabsObservation(
		source.ObserveRawBrowserTabs(ctx, requestedProfile),
		requestedProfile,
	)
}

// ObserveSharedSessionBrowserRawNavigation loads a direct route-scoped
// navigation source of truth used by watch-manager redirect/current-target
// recovery helpers.
func ObserveSharedSessionBrowserRawNavigation(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawNavigationObservation {
	requestedProfile = strings.TrimSpace(requestedProfile)
	if control == nil {
		return SharedSessionBrowserRawNavigationObservation{RequestedProfile: requestedProfile}
	}
	source, ok := control.(BrowserRuntimeRawNavigationObservationBackend)
	if !ok {
		return SharedSessionBrowserRawNavigationObservation{RequestedProfile: requestedProfile}
	}
	return normalizeSharedSessionBrowserRawNavigationObservation(
		source.ObserveRawBrowserNavigation(ctx, requestedProfile),
		requestedProfile,
	)
}

// ObserveSharedSessionBrowserRawOpen loads a direct route-scoped open source
// of truth used by watch-manager target-created recovery helpers.
func ObserveSharedSessionBrowserRawOpen(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawOpenObservation {
	requestedProfile = strings.TrimSpace(requestedProfile)
	if control == nil {
		return SharedSessionBrowserRawOpenObservation{RequestedProfile: requestedProfile}
	}
	source, ok := control.(BrowserRuntimeRawOpenObservationBackend)
	if !ok {
		return SharedSessionBrowserRawOpenObservation{RequestedProfile: requestedProfile}
	}
	return normalizeSharedSessionBrowserRawOpenObservation(
		source.ObserveRawBrowserOpen(ctx, requestedProfile),
		requestedProfile,
	)
}

// ObserveSharedSessionBrowserRawTarget loads a direct route-scoped generic
// target source of truth used by watch-manager target-tracking recovery
// helpers.
func ObserveSharedSessionBrowserRawTarget(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawTargetObservation {
	requestedProfile = strings.TrimSpace(requestedProfile)
	if control == nil {
		return SharedSessionBrowserRawTargetObservation{RequestedProfile: requestedProfile}
	}
	source, ok := control.(BrowserRuntimeRawTargetObservationBackend)
	if !ok {
		return SharedSessionBrowserRawTargetObservation{RequestedProfile: requestedProfile}
	}
	return normalizeSharedSessionBrowserRawTargetObservation(
		source.ObserveRawBrowserTarget(ctx, requestedProfile),
		requestedProfile,
	)
}

// ObserveSharedSessionBrowserRawRouteMutation loads a direct route-scoped
// mutation source of truth used by watch-manager attach/detach/current-target
// recovery helpers.
func ObserveSharedSessionBrowserRawRouteMutation(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	requestedProfile string,
) SharedSessionBrowserRawRouteMutationObservation {
	requestedProfile = strings.TrimSpace(requestedProfile)
	if control == nil {
		return SharedSessionBrowserRawRouteMutationObservation{RequestedProfile: requestedProfile}
	}
	source, ok := control.(BrowserRuntimeRawRouteMutationObservationBackend)
	if !ok {
		return SharedSessionBrowserRawRouteMutationObservation{RequestedProfile: requestedProfile}
	}
	return normalizeSharedSessionBrowserRawRouteMutationObservation(
		source.ObserveRawBrowserRouteMutation(ctx, requestedProfile),
		requestedProfile,
	)
}
