package tools

import (
	"context"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserManagedResolverRecoveryResult struct {
	Snapshot                 BrowserSnapshotResult
	SnapshotRecovered        bool
	SnapshotText             string
	SnapshotTruncated        bool
	FinalURL                 string
	Title                    string
	BrowserApp               string
	Backend                  string
	Note                     string
	InvalidateSessionTargets bool
}

type browserManagedResolverFailurePolicyArgs struct {
	Route   browserResolvedExecutionRoute
	Request browserManagedRouteExecutionArgs
}

type browserManagedResolverFailurePolicyResult struct {
	Status         string
	Outcome        *agentxbrowserruntime.BrowserElementResolverOutcome
	RecoveryAction string
	FinalURL       string
	Title          string
	BrowserApp     string
	Backend        string
	Note           string
	Recovery       browserManagedResolverRecoveryResult
}

type browserManagedResolverExecutionResult[T any] struct {
	Result   T
	Recovery browserManagedResolverRecoveryResult
}

type browserManagedResolverSnapshotPayload struct {
	Text        string
	Format      string
	Mode        string
	Refs        string
	Frame       string
	Elements    []BrowserSnapshotElement
	Interactive bool
	Compact     bool
	Depth       int
	Truncated   bool
}

func browserManagedResolverFailurePolicy(
	ctx context.Context,
	err error,
	args browserManagedResolverFailurePolicyArgs,
) (browserManagedResolverFailurePolicyResult, bool) {
	status, outcome, note, ok := browserManagedResolverFailure(err)
	if !ok {
		return browserManagedResolverFailurePolicyResult{}, false
	}
	route := args.Route
	req := args.Request
	result := browserManagedResolverFailurePolicyResult{
		Status:         status,
		Outcome:        outcome,
		RecoveryAction: browserResolverRecoveryAction(outcome),
		FinalURL:       strings.TrimSpace(req.FinalURL),
		Title:          strings.TrimSpace(req.Title),
		BrowserApp:     firstNonEmpty(strings.TrimSpace(req.ResultBrowserApp), strings.TrimSpace(req.BrowserApp)),
		Backend:        firstNonEmpty(strings.TrimSpace(req.ResultBackend), strings.TrimSpace(route.RuntimeInfo.Backend)),
		Note:           strings.TrimSpace(note),
	}
	result.Recovery = route.managedResolverRecovery(
		ctx,
		req,
		outcome,
		result.FinalURL,
		result.Title,
		result.BrowserApp,
		result.Backend,
		result.Note,
	)
	result.FinalURL = firstNonEmpty(strings.TrimSpace(result.Recovery.FinalURL), result.FinalURL)
	result.Title = firstNonEmpty(strings.TrimSpace(result.Recovery.Title), result.Title)
	result.BrowserApp = firstNonEmpty(strings.TrimSpace(result.Recovery.BrowserApp), result.BrowserApp)
	result.Backend = firstNonEmpty(strings.TrimSpace(result.Recovery.Backend), result.Backend)
	result.Note = firstNonEmpty(strings.TrimSpace(result.Recovery.Note), result.Note)
	return result, true
}

func browserManagedResolverExecute[T any](
	ctx context.Context,
	invoke func() (T, error),
	args browserManagedResolverFailurePolicyArgs,
	fromPolicy func(browserManagedResolverFailurePolicyResult) T,
) (browserManagedResolverExecutionResult[T], error) {
	result, err := invoke()
	if err == nil {
		return browserManagedResolverExecutionResult[T]{Result: result}, nil
	}
	policy, ok := browserManagedResolverFailurePolicy(ctx, err, args)
	if !ok {
		var zero T
		return browserManagedResolverExecutionResult[T]{Result: zero}, err
	}
	return browserManagedResolverExecutionResult[T]{
		Result:   fromPolicy(policy),
		Recovery: policy.Recovery,
	}, nil
}

func browserManagedResolverSnapshotPayloadForRecovery(recovery browserManagedResolverRecoveryResult) browserManagedResolverSnapshotPayload {
	return browserManagedResolverSnapshotPayload{
		Text:   recovery.SnapshotText,
		Format: strings.TrimSpace(recovery.Snapshot.Format),
		Mode:   strings.TrimSpace(recovery.Snapshot.Mode),
		Refs: func() string {
			if recovery.SnapshotRecovered {
				return firstNonEmpty(strings.TrimSpace(recovery.Snapshot.Refs), "role")
			}
			return ""
		}(),
		Frame:       strings.TrimSpace(recovery.Snapshot.Frame),
		Elements:    append([]BrowserSnapshotElement(nil), recovery.Snapshot.Elements...),
		Interactive: recovery.Snapshot.Interactive || recovery.SnapshotRecovered,
		Compact:     recovery.Snapshot.Compact,
		Depth:       recovery.Snapshot.Depth,
		Truncated:   recovery.SnapshotTruncated,
	}
}

func browserManagedResolverApplyTargetInvalidation(targetID string, recovery browserManagedResolverRecoveryResult) string {
	if recovery.InvalidateSessionTargets {
		return ""
	}
	return strings.TrimSpace(targetID)
}

func browserManagedResolverRefreshRecoveryNote(base string, result browserRuntimePrepareResult, err error) string {
	base = strings.TrimSpace(base)
	decision := strings.TrimSpace(result.Decision)
	suffix := ""
	switch {
	case err != nil && decision != "":
		suffix = "refresh recovery via browser action=refresh failed (" + decision + ")"
	case err != nil:
		suffix = "refresh recovery via browser action=refresh failed"
	case decision == "restart_reconnect_in_progress":
		suffix = "refresh recovery via browser action=refresh is already in progress"
	case decision == "restart_blocked_active_node_run":
		suffix = "refresh recovery via browser action=refresh was blocked by an active node run"
	case decision != "":
		suffix = "refresh recovery via browser action=refresh returned " + decision
	default:
		return base
	}
	if base == "" {
		return suffix
	}
	return base + "; " + suffix
}
