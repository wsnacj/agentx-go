package browserruntime

import "time"

type SharedSessionBrowserHealthSummary struct {
	State                       string
	Reason                      string
	RecoveryAction              string
	ReconnectHint               string
	DisconnectCount             int
	DisconnectBurstCount        int
	DisconnectBurstWindowMs     int
	CooldownRemainingMs         int
	RetryBackoffRemainingMs     int
	RestartAttemptCount         int
	RestartFailureCount         int
	LastDisconnectUnixMilli     int64
	LastReconnectUnixMilli      int64
	LastRestartAttemptUnixMilli int64
	LastRestartResult           string
	LastRestartError            string
	RecommendedBackoffMs        int
	ResolverBlockedBy           string
	AmbiguityClass              string
	CandidateKind               string
	CandidateStrength           string
	RetryDisposition            string
	ManualRetryHint             string
	NextStepAlias               string
	SpecificityFields           []string
}

type SharedSessionBrowserHealthEvaluation struct {
	Summary           *SharedSessionBrowserHealthSummary
	Profile           SharedSessionBrowserProfileState
	HasProfile        bool
	ReconnectTimedOut bool
}

type SharedSessionBrowserHealthInput struct {
	ActiveNodeRunID                   string
	RouteTargetCount                  int
	StoredState                       string
	StoredReason                      string
	StoredRecoveryAction              string
	StoredReconnectHint               string
	StoredDisconnectCount             int
	StoredDisconnectBurstCount        int
	StoredDisconnectBurstWindowMs     int
	StoredCooldownRemainingMs         int
	StoredRetryBackoffRemainingMs     int
	StoredRestartAttemptCount         int
	StoredRestartFailureCount         int
	StoredLastDisconnectUnixMilli     int64
	StoredLastReconnectUnixMilli      int64
	StoredLastRestartAttemptUnixMilli int64
	StoredLastRestartResult           string
	StoredLastRestartError            string
	StoredRecommendedBackoffMs        int
	StoredResolverBlockedBy           string
	StoredAmbiguityClass              string
	StoredCandidateKind               string
	StoredCandidateStrength           string
	StoredRetryDisposition            string
	StoredManualRetryHint             string
	StoredNextStepAlias               string
	StoredSpecificityFields           []string
	ReferenceTime                     time.Time
	Profiles                          []SharedSessionBrowserProfileState
}

type SharedSessionBrowserHealthActions struct {
	PrimaryAction      string
	RestartAction      string
	RecommendedActions []string
	ClearRestartAction bool
	SuppressRefresh    bool
}

type SharedSessionBrowserHealthGuidance struct {
	RestartAction      string
	PrimaryAction      string
	PrimaryNodeAction  string
	NextStep           string
	RecommendedActions []string
}

type SharedSessionBrowserProfileRecoveryAssessment struct {
	EffectiveStatus          BrowserProfileStatusResult
	NeedsRefreshRecovery     bool
	ShouldStopBeforeRecovery bool
	ReconnectInProgress      bool
	HasSyntheticStatus       bool
	SyntheticStatus          BrowserProfileStatusResult
}

type SharedSessionBrowserCoordinationInput struct {
	ActiveNodeRunID         string
	RouteTargetCount        int
	SelectedBrowserProfile  string
	SelectedBrowserTargetID string
	Profiles                []SharedSessionBrowserProfileState
}

type SharedSessionBrowserCoordinationPlan struct {
	State                     string
	BrowserOnNode             bool
	HasActiveNodeRun          bool
	HasRunningBrowserProfile  bool
	NeedsSessionSync          bool
	SyncAction                string
	PrepareAction             string
	RestartAction             string
	TeardownAction            string
	PrimaryBrowserAction      string
	PrimaryNodeAction         string
	NextStep                  string
	RecommendedBrowserActions []string
	RecommendedNodeActions    []string
}

type SharedSessionBrowserCoordinationActions struct {
	PrimaryAction      string
	RecommendedActions []string
}

type SharedSessionBrowserCoordinationGuidance struct {
	PrimaryAction      string
	NextStep           string
	RecommendedActions []string
}

type SharedSessionBrowserCoordinationStatus struct {
	Decision string
	Ready    bool
}

type SharedSessionBrowserCoordinationEvaluationInput struct {
	Coordination      SharedSessionBrowserCoordinationInput
	Routes            []SharedSessionBrowserRouteCoordinationInput
	HealthEvaluation  SharedSessionBrowserHealthEvaluation
	BlockedAutoFollow bool
}

type SharedSessionBrowserCoordinationEvaluation struct {
	Plan          SharedSessionBrowserCoordinationPlan
	RestartAction string
	Guidance      SharedSessionBrowserCoordinationGuidance
}

type SharedSessionBrowserRouteCoordinationInput struct {
	FollowPolicyState string
	ManagedRuntime    bool
}

type SharedSessionBrowserManagedRouteRecoveryInput struct {
	ActiveNodeRunID string
	Profiles        []SharedSessionBrowserProfileState
	Routes          []SharedSessionBrowserRouteCoordinationInput
}
