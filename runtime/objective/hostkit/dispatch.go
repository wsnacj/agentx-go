// Package hostkit composes the canonical Objective ingress, host-owned runtime
// adapter dispatch, observation normalization, and verification mechanisms.
//
// The package is a Developer Preview candidate for private validation. It does
// not provide a provider, Runner, backend, credential, product policy, or
// production side effect.
package hostkit

import (
	"context"
	"strings"

	objective "github.com/wsnacj/agentx-go/runtime/objective"
)

// Handler is a host-owned runtime-adapter execution function.
type Handler func(context.Context, objective.RuntimeAdapterRequest) objective.RuntimeAdapterResult

// DispatchInput controls one explicitly confirmed host dispatch.
type DispatchInput struct {
	Enabled                  bool
	HostConfirmed            bool
	Request                  objective.RuntimeAdapterRequest
	Handler                  Handler
	Handlers                 map[objective.DisplaySafeRef]Handler
	ExpectedObservationKinds []string
	Boundaries               []objective.Boundary
	Context                  context.Context
}

// DispatchResult is the normalized readback of one host-owned adapter call.
type DispatchResult struct {
	Available                        bool   `json:"available"`
	Enabled                          bool   `json:"enabled"`
	HostConfirmed                    bool   `json:"host_confirmed"`
	Status                           string `json:"status,omitempty"`
	RequestReady                     bool   `json:"request_ready"`
	HandlerReady                     bool   `json:"handler_ready"`
	ResultReadbackReady              bool   `json:"result_readback_ready"`
	ReadyForObservationNormalization bool   `json:"ready_for_observation_normalization"`
	ReadyForVerification             bool   `json:"ready_for_verification"`
	Satisfied                        bool   `json:"satisfied"`
	HostExecutionReported            bool   `json:"host_execution_reported"`
	RuntimeAdapterExecuted           bool   `json:"runtime_adapter_executed"`
	RuntimeAdapterExecutedByCore     bool   `json:"runtime_adapter_executed_by_core"`
	RunnerDispatched                 bool   `json:"runner_dispatched"`
	ToolExecuted                     bool   `json:"tool_executed"`
	WorkflowDispatched               bool   `json:"workflow_dispatched"`
	SchedulerApplied                 bool   `json:"scheduler_applied"`
	InstallerExecuted                bool   `json:"installer_executed"`
	RawOutputLoaded                  bool   `json:"raw_output_loaded"`
	AdapterRef                       string `json:"adapter_ref,omitempty"`
	StrategyRef                      string `json:"strategy_ref,omitempty"`
	HostAdapterRunRef                string `json:"host_adapter_run_ref,omitempty"`
	RequestStatus                    string `json:"request_status,omitempty"`
	ResultStatus                     string `json:"result_status,omitempty"`
	NormalizationStatus              string `json:"normalization_status,omitempty"`
	VerificationStatus               string `json:"verification_status,omitempty"`
	FailureClass                     string `json:"failure_class,omitempty"`
	NextHostAction                   string `json:"next_host_action,omitempty"`
	MissingInputs                    []string
	BlockedReasons                   []string
	Boundaries                       []string
	Request                          objective.RuntimeAdapterRequest          `json:"request,omitempty"`
	Result                           objective.RuntimeAdapterResult           `json:"result,omitempty"`
	Normalization                    objective.ObservationNormalizationResult `json:"normalization,omitempty"`
	Verification                     objective.ObjectiveVerificationResult    `json:"verification,omitempty"`
}

// Dispatch invokes exactly one host handler after request and confirmation
// gates, then runs canonical observation normalization and verification.
func Dispatch(input DispatchInput) DispatchResult {
	report := DispatchResult{
		Status: "blocked",
		Boundaries: []string{
			"host_runtime_adapter_dispatch_facade",
			"host_owned_runtime_adapter_dispatch_facade",
			"explicit_opt_in_host_dispatch",
			"display_safe_refs_only",
			"core_does_not_execute_adapter",
			"no_runner_dispatch",
			"no_tool_execution_by_core",
			"no_workflow_dispatch_by_core",
			"no_scheduler_apply_by_core",
			"no_install_apply_by_core",
		},
		NextHostAction: "review_runtime_adapter_dispatch_facade",
	}
	report.Boundaries = appendUniqueStrings(report.Boundaries, boundariesToStrings(input.Boundaries)...)
	if !input.Enabled {
		report.MissingInputs = appendUniqueStrings(report.MissingInputs, "host:runtime_adapter_dispatch_facade_enabled")
		report.BlockedReasons = appendUniqueStrings(report.BlockedReasons, "runtime_adapter_dispatch_facade_disabled")
		report.Boundaries = appendUniqueStrings(report.Boundaries, "runtime_adapter_dispatch_facade_default_off")
		report.NextHostAction = "enable_runtime_adapter_dispatch_facade"
		return report.Normalize()
	}
	report.Enabled = true
	if !input.HostConfirmed {
		report.MissingInputs = appendUniqueStrings(report.MissingInputs, "host:runtime_adapter_dispatch_confirmation")
		report.BlockedReasons = appendUniqueStrings(report.BlockedReasons, "runtime_adapter_dispatch_confirmation_missing")
		report.Boundaries = appendUniqueStrings(report.Boundaries, "runtime_adapter_dispatch_requires_host_confirmation")
		report.NextHostAction = "request_runtime_adapter_dispatch_confirmation"
		return report.Normalize()
	}
	report.HostConfirmed = true

	request := input.Request.Normalize()
	report = bindRequest(report, request)
	if !request.ReadyForHostExecution {
		report.MissingInputs = appendUniqueStrings(report.MissingInputs, "host:runtime_adapter_execution_request")
		report.BlockedReasons = appendUniqueStrings(report.BlockedReasons, "runtime_adapter_request_not_ready")
		report.NextHostAction = firstNonEmptyString(string(request.NextHostAction), "provide_runtime_adapter_execution_request")
		return report.Normalize()
	}

	handler := selectHandler(input, request.AdapterRef)
	if handler == nil {
		report.MissingInputs = appendUniqueStrings(report.MissingInputs, "host:runtime_adapter_dispatch_handler")
		report.BlockedReasons = appendUniqueStrings(report.BlockedReasons, "runtime_adapter_dispatch_handler_missing")
		report.Boundaries = appendUniqueStrings(report.Boundaries, "runtime_adapter_dispatch_handler_missing")
		report.NextHostAction = "provide_runtime_adapter_dispatch_handler"
		return report.Normalize()
	}
	report.HandlerReady = true
	report.Available = true
	report.Boundaries = appendUniqueStrings(report.Boundaries, "runtime_adapter_dispatch_handler_ready")

	ctx := input.Context
	if ctx == nil {
		ctx = context.Background()
	}
	result := handler(ctx, request).Normalize()
	report = bindResult(report, result)
	if !result.ReadyForObservationNormalization {
		report.BlockedReasons = appendUniqueStrings(report.BlockedReasons, "runtime_adapter_result_not_ready_for_normalization")
		report.NextHostAction = firstNonEmptyString(string(result.NextHostAction), "review_runtime_adapter_result")
		return report.Normalize()
	}

	normalization := objective.BuildObservationNormalization(objective.ObservationNormalizationInput{
		Frame:                    request.Frame,
		SourceKind:               "host_adapter_result",
		SourceRef:                result.HostAdapterRunRef,
		RuntimeAdapterResult:     result,
		ExpectedObservationKinds: input.ExpectedObservationKinds,
		Boundaries: objective.AppendBoundaries(input.Boundaries,
			"runtime_adapter_dispatch_facade_observation_normalization",
		),
	})
	report = bindNormalization(report, normalization)
	if !normalization.ReadyForVerification {
		report.BlockedReasons = appendUniqueStrings(report.BlockedReasons, "runtime_adapter_normalization_not_ready")
		report.NextHostAction = firstNonEmptyString(string(normalization.NextHostAction), "normalize_observations")
		return report.Normalize()
	}

	verification := objective.BuildVerification(objective.ObjectiveVerificationInput{
		Frame:         request.Frame,
		Normalization: normalization,
		Boundaries: objective.AppendBoundaries(input.Boundaries,
			"runtime_adapter_dispatch_facade_objective_verification",
		),
	})
	report = bindVerification(report, verification)
	return report.Normalize()
}

// Normalize returns a defensive, canonical view of the dispatch result.
func (report DispatchResult) Normalize() DispatchResult {
	out := report
	out.Status = strings.TrimSpace(out.Status)
	if out.Status == "" {
		out.Status = "blocked"
	}
	out.AdapterRef = strings.TrimSpace(out.AdapterRef)
	out.StrategyRef = strings.TrimSpace(out.StrategyRef)
	out.HostAdapterRunRef = strings.TrimSpace(out.HostAdapterRunRef)
	out.RequestStatus = strings.TrimSpace(out.RequestStatus)
	out.ResultStatus = strings.TrimSpace(out.ResultStatus)
	out.NormalizationStatus = strings.TrimSpace(out.NormalizationStatus)
	out.VerificationStatus = strings.TrimSpace(out.VerificationStatus)
	out.FailureClass = strings.TrimSpace(out.FailureClass)
	out.NextHostAction = strings.TrimSpace(out.NextHostAction)
	if out.NextHostAction == "" {
		out.NextHostAction = "review_runtime_adapter_dispatch_facade"
	}
	out.MissingInputs = appendUniqueStrings(nil, out.MissingInputs...)
	out.BlockedReasons = appendUniqueStrings(nil, out.BlockedReasons...)
	out.Boundaries = appendUniqueStrings(nil, out.Boundaries...)
	out.Request = out.Request.Normalize()
	out.Result = out.Result.Normalize()
	out.Normalization = out.Normalization.Normalize()
	out.Verification = out.Verification.Normalize()
	out.RequestReady = out.RequestReady && out.Request.ReadyForHostExecution
	out.ResultReadbackReady = out.ResultReadbackReady && out.Result.ReadyForObservationNormalization
	out.ReadyForObservationNormalization = out.ReadyForObservationNormalization && out.ResultReadbackReady
	out.ReadyForVerification = out.ReadyForVerification && out.Normalization.ReadyForVerification
	out.Satisfied = out.Satisfied && out.Verification.Satisfied
	out.HostExecutionReported = out.HostExecutionReported || out.Result.HostExecutionReported
	out.RuntimeAdapterExecuted = out.RuntimeAdapterExecuted || out.HostAdapterRunRef != ""
	out.RuntimeAdapterExecutedByCore = false
	out.RunnerDispatched = false
	out.ToolExecuted = false
	out.WorkflowDispatched = false
	out.SchedulerApplied = false
	out.InstallerExecuted = false
	out.RawOutputLoaded = out.RawOutputLoaded || out.Request.RawOutputLoaded || out.Result.RawOutputLoaded || out.Normalization.RawOutputLoaded || out.Verification.RawOutputLoaded
	if out.RawOutputLoaded {
		out.Status = "review_required"
		out.Satisfied = false
		out.MissingInputs = appendUniqueStrings(out.MissingInputs, "host:display_safe_refs")
		out.BlockedReasons = appendUniqueStrings(out.BlockedReasons, "unsafe_input_ref")
		out.Boundaries = appendUniqueStrings(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if len(out.MissingInputs) > 0 || len(out.BlockedReasons) > 0 {
		if out.Status == "satisfied" {
			out.Status = "partial"
		}
	}
	return out
}

func selectHandler(input DispatchInput, adapterRef objective.DisplaySafeRef) Handler {
	if input.Handler != nil {
		return input.Handler
	}
	normalized, ok := objective.NormalizeDisplaySafeRef(string(adapterRef))
	if !ok || normalized == "" {
		return nil
	}
	for ref, handler := range input.Handlers {
		if handler == nil {
			continue
		}
		if candidate, ok := objective.NormalizeDisplaySafeRef(string(ref)); ok && candidate == normalized {
			return handler
		}
	}
	return nil
}

func bindRequest(report DispatchResult, request objective.RuntimeAdapterRequest) DispatchResult {
	report.Request = request
	report.RequestReady = request.ReadyForHostExecution
	report.AdapterRef = string(request.AdapterRef)
	report.StrategyRef = string(request.StrategyRef)
	report.RequestStatus = string(request.Status)
	report.FailureClass = firstFailureClass(objective.FailureClass(report.FailureClass), request.FailureClass)
	report.MissingInputs = appendUniqueStrings(report.MissingInputs, missingInputsToStrings(request.MissingInputs)...)
	report.Boundaries = appendUniqueStrings(report.Boundaries, boundariesToStrings(request.Boundaries)...)
	report.NextHostAction = firstNonEmptyString(string(request.NextHostAction), report.NextHostAction)
	return report
}

func bindResult(report DispatchResult, result objective.RuntimeAdapterResult) DispatchResult {
	report.Result = result
	report.ResultStatus = string(result.Status)
	if result.Status != "" && result.Status != objective.VerificationSatisfied {
		report.Status = string(result.Status)
	}
	report.ResultReadbackReady = result.ReadyForObservationNormalization
	report.ReadyForObservationNormalization = result.ReadyForObservationNormalization
	report.HostExecutionReported = result.HostExecutionReported
	report.RuntimeAdapterExecuted = result.HostExecutionReported || result.HostAdapterRunRef != ""
	report.HostAdapterRunRef = string(result.HostAdapterRunRef)
	report.FailureClass = firstFailureClass(objective.FailureClass(report.FailureClass), result.FailureClass)
	report.MissingInputs = appendUniqueStrings(report.MissingInputs, missingInputsToStrings(result.MissingInputs)...)
	report.Boundaries = appendUniqueStrings(report.Boundaries, append(boundariesToStrings(result.Boundaries), "host_runtime_adapter_handler_invoked")...)
	report.NextHostAction = firstNonEmptyString(string(result.NextHostAction), report.NextHostAction)
	return report
}

func bindNormalization(report DispatchResult, normalization objective.ObservationNormalizationResult) DispatchResult {
	report.Normalization = normalization
	report.NormalizationStatus = string(normalization.Status)
	report.ReadyForVerification = normalization.ReadyForVerification
	report.FailureClass = firstFailureClass(objective.FailureClass(report.FailureClass), normalization.FailureClass)
	report.MissingInputs = appendUniqueStrings(report.MissingInputs, missingInputsToStrings(normalization.MissingInputs)...)
	report.Boundaries = appendUniqueStrings(report.Boundaries, boundariesToStrings(normalization.Boundaries)...)
	report.NextHostAction = firstNonEmptyString(string(normalization.NextHostAction), report.NextHostAction)
	return report
}

func bindVerification(report DispatchResult, verification objective.ObjectiveVerificationResult) DispatchResult {
	report.Verification = verification
	report.VerificationStatus = string(verification.Status)
	report.Status = string(verification.Status)
	report.Satisfied = verification.Satisfied
	report.FailureClass = firstFailureClass(verification.FailureClass, objective.FailureClass(report.FailureClass))
	report.MissingInputs = appendUniqueStrings(report.MissingInputs, missingInputsToStrings(verification.MissingInputs)...)
	report.Boundaries = appendUniqueStrings(report.Boundaries, boundariesToStrings(verification.Boundaries)...)
	report.NextHostAction = firstNonEmptyString(string(verification.NextHostAction), report.NextHostAction)
	return report
}

func firstFailureClass(values ...objective.FailureClass) string {
	for _, value := range values {
		if normalized := objective.NormalizeFailureClass(string(value)); normalized != objective.FailureNone {
			return string(normalized)
		}
	}
	return ""
}

func missingInputsToStrings(inputs []objective.MissingInput) []string {
	out := make([]string, 0, len(inputs))
	for _, input := range objective.MergeMissingInputs(inputs) {
		if value := strings.TrimSpace(string(input)); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func boundariesToStrings(boundaries []objective.Boundary) []string {
	out := make([]string, 0, len(boundaries))
	for _, boundary := range boundaries {
		if value := strings.TrimSpace(string(boundary)); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func appendUniqueStrings(items []string, values ...string) []string {
	seen := make(map[string]struct{}, len(items)+len(values))
	out := make([]string, 0, len(items)+len(values))
	for _, value := range append(append([]string(nil), items...), values...) {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
