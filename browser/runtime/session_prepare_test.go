package browserruntime

import (
	"context"
	"errors"
	"testing"
)

type preparedProfileTestBackend struct {
	profilesReqs []BrowserProfilesRequest
	profilesResp BrowserProfilesResult
	profilesErr  error
}

func (b *preparedProfileTestBackend) RuntimeStatus(context.Context, BrowserProfileStatusRequest) (BrowserProfileStatusResult, error) {
	return BrowserProfileStatusResult{}, nil
}

func (b *preparedProfileTestBackend) RuntimeStart(context.Context, BrowserProfileLifecycleRequest) (BrowserProfileStatusResult, error) {
	return BrowserProfileStatusResult{}, nil
}

func (b *preparedProfileTestBackend) RuntimeStop(context.Context, BrowserProfileLifecycleRequest) (BrowserProfileStatusResult, error) {
	return BrowserProfileStatusResult{}, nil
}

func (b *preparedProfileTestBackend) RuntimeProfiles(_ context.Context, req BrowserProfilesRequest) (BrowserProfilesResult, error) {
	b.profilesReqs = append(b.profilesReqs, req)
	return b.profilesResp, b.profilesErr
}

func TestResolveSharedSessionPreparedProfileFromProfilesPrefersDefaultThenSelectedThenFirst(t *testing.T) {
	if got := ResolveSharedSessionPreparedProfileFromProfiles(BrowserRuntimeInfo{Profile: "selected"}, BrowserProfilesResult{
		DefaultProfile: "default",
		Profiles: []BrowserProfileInfo{
			{Profile: "first"},
		},
	}); got != "default" {
		t.Fatalf("expected default profile to win, got %q", got)
	}
	if got := ResolveSharedSessionPreparedProfileFromProfiles(BrowserRuntimeInfo{Profile: "selected"}, BrowserProfilesResult{
		Profiles: []BrowserProfileInfo{
			{Profile: "first"},
		},
	}); got != "selected" {
		t.Fatalf("expected selected profile fallback, got %q", got)
	}
	if got := ResolveSharedSessionPreparedProfileFromProfiles(BrowserRuntimeInfo{}, BrowserProfilesResult{
		Profiles: []BrowserProfileInfo{
			{Profile: "first"},
			{Profile: "second"},
		},
	}); got != "first" {
		t.Fatalf("expected first discovered profile fallback, got %q", got)
	}
}

func TestLoadSharedSessionPreparedProfileBypassesProfilesForExplicitRequest(t *testing.T) {
	backend := &preparedProfileTestBackend{}
	profile, profiles, observedAt, err := LoadSharedSessionPreparedProfile(context.Background(), backend, "isolated", BrowserRuntimeInfo{Profile: "selected"})
	if err != nil {
		t.Fatalf("load prepared profile: %v", err)
	}
	if profile != "isolated" || profiles != nil || !observedAt.IsZero() {
		t.Fatalf("expected explicit requested profile without discovery, got profile=%q profiles=%#v observed_at=%v", profile, profiles, observedAt)
	}
	if len(backend.profilesReqs) != 0 {
		t.Fatalf("expected no RuntimeProfiles call for explicit request, got %#v", backend.profilesReqs)
	}
}

func TestLoadSharedSessionPreparedProfileFallsBackToSelectedProfileOnProfilesError(t *testing.T) {
	backend := &preparedProfileTestBackend{profilesErr: errors.New("profiles unavailable")}
	profile, profiles, observedAt, err := LoadSharedSessionPreparedProfile(context.Background(), backend, "", BrowserRuntimeInfo{Profile: "selected"})
	if err != nil {
		t.Fatalf("load prepared profile fallback: %v", err)
	}
	if profile != "selected" || profiles != nil || !observedAt.IsZero() {
		t.Fatalf("expected selected profile fallback on profiles error, got profile=%q profiles=%#v observed_at=%v", profile, profiles, observedAt)
	}
	if len(backend.profilesReqs) != 1 {
		t.Fatalf("expected one RuntimeProfiles call, got %#v", backend.profilesReqs)
	}
}

func TestLoadSharedSessionPreparedProfileReturnsProfilesResultWhenNoManagedProfileExists(t *testing.T) {
	backend := &preparedProfileTestBackend{profilesResp: BrowserProfilesResult{
		Backend: "proxy",
		Note:    "no profiles",
	}}
	profile, profiles, observedAt, err := LoadSharedSessionPreparedProfile(context.Background(), backend, "", BrowserRuntimeInfo{})
	if err == nil {
		t.Fatalf("expected missing prepared profile error")
	}
	if profile != "" || profiles == nil || profiles.Note != "no profiles" || observedAt.IsZero() {
		t.Fatalf("expected missing profile to return profiles result context, got profile=%q profiles=%#v observed_at=%v err=%v", profile, profiles, observedAt, err)
	}
}
