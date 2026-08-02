package tools

import (
	"encoding/json"
	"strings"
)

type browserRuntimeTopLevelAliasPayloadDecoder struct {
	payload map[string]json.RawMessage
}

func newBrowserRuntimeTopLevelAliasPayloadDecoder(
	payload map[string]json.RawMessage,
) browserRuntimeTopLevelAliasPayloadDecoder {
	return browserRuntimeTopLevelAliasPayloadDecoder{payload: payload}
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) raw(key string) json.RawMessage {
	if decoder.payload == nil {
		return nil
	}
	return decoder.payload[key]
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) summary(key string) *browserTopLevelSummary {
	return browserRuntimeTopLevelSummaryFromRawPtr(decoder.raw(key))
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) display(key string) *browserTopLevelDisplaySummary {
	return browserRuntimeTopLevelDisplayFromRawPtr(decoder.raw(key))
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) review(key string) *browserReviewSurfaceSummary {
	return browserRuntimeReviewSurfaceFromRawPtr(decoder.raw(key))
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) surface(key string) *browserTopLevelSurfaceSummary {
	return browserRuntimeTopLevelSurfaceFromRawPtr(decoder.raw(key))
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) view(key string) *browserTopLevelViewSummary {
	return browserRuntimeTopLevelViewFromRawPtr(decoder.raw(key))
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) workbench(key string) *browserRuntimeWorkbenchSurfaceSummary {
	return browserRuntimeWorkbenchSurfaceFromRawPtr(decoder.raw(key))
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) workbenchDisplay(key string) *browserRuntimeWorkbenchDisplaySummary {
	return browserRuntimeWorkbenchDisplayFromRawPtr(decoder.raw(key))
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) stringValue(key string) string {
	return browserRuntimeStringFromRaw(decoder.raw(key))
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) stringSlice(key string) []string {
	return browserRuntimeStringSliceFromRaw(decoder.raw(key))
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) routeDescriptor(key string) browserRuntimeRouteDescriptor {
	route, ok := browserRuntimeRouteDescriptorFromRaw(decoder.raw(key))
	if !ok {
		return browserRuntimeRouteDescriptor{}
	}
	return route
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) diagnosticsExplanationSummary(key string) *browserTopLevelSummary {
	return browserRuntimeTopLevelSummaryFromDiagnosticsExplanationRawPtr(decoder.raw(key))
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) runtimeDiagnosticsExplanationSummary(key string) *browserTopLevelSummary {
	return browserRuntimeTopLevelSummaryFromRuntimeDiagnosticsExplanationRawPtr(decoder.raw(key))
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) resolverExplanationSummary(key string) *browserTopLevelSummary {
	return browserRuntimeTopLevelSummaryFromResolverExplanationRawPtr(decoder.raw(key))
}

func (decoder browserRuntimeTopLevelAliasPayloadDecoder) workbenchDiagnosticsSummary(key string) *browserTopLevelSummary {
	return browserRuntimeTopLevelSummaryFromWorkbenchDiagnosticsRawPtr(decoder.raw(key))
}

func browserRuntimeTopLevelSummaryFromRaw(raw json.RawMessage) (browserTopLevelSummary, bool) {
	if summary, ok := browserRuntimeDecodeNonEmptyRaw(raw, browserUnifiedSummaryEmpty); ok {
		return *summary, true
	}
	return browserTopLevelSummary{}, false
}

func browserRuntimeTopLevelSummaryFromRawPtr(raw json.RawMessage) *browserTopLevelSummary {
	summary, ok := browserRuntimeTopLevelSummaryFromRaw(raw)
	if !ok {
		return nil
	}
	return &summary
}

func browserRuntimeTopLevelDisplayFromRaw(raw json.RawMessage) (browserTopLevelDisplaySummary, bool) {
	if display, ok := browserRuntimeDecodeNonEmptyRaw(raw, browserTopLevelDisplayEmpty); ok {
		return *display, true
	}
	return browserTopLevelDisplaySummary{}, false
}

func browserRuntimeTopLevelDisplayFromRawPtr(raw json.RawMessage) *browserTopLevelDisplaySummary {
	display, ok := browserRuntimeTopLevelDisplayFromRaw(raw)
	if !ok {
		return nil
	}
	return &display
}

func browserRuntimeReviewSurfaceFromRaw(raw json.RawMessage) (browserReviewSurfaceSummary, bool) {
	if review, ok := browserRuntimeDecodeNonEmptyRaw(raw, browserReviewSurfaceSummaryEmpty); ok {
		return *review, true
	}
	return browserReviewSurfaceSummary{}, false
}

func browserRuntimeReviewSurfaceFromRawPtr(raw json.RawMessage) *browserReviewSurfaceSummary {
	review, ok := browserRuntimeReviewSurfaceFromRaw(raw)
	if !ok {
		return nil
	}
	return &review
}

func browserRuntimeTopLevelSurfaceFromRaw(raw json.RawMessage) (browserTopLevelSurfaceSummary, bool) {
	if surface, ok := browserRuntimeDecodeNonEmptyRaw(raw, browserTopLevelSurfaceEmpty); ok {
		return *surface, true
	}
	return browserTopLevelSurfaceSummary{}, false
}

func browserRuntimeTopLevelSurfaceFromRawPtr(raw json.RawMessage) *browserTopLevelSurfaceSummary {
	surface, ok := browserRuntimeTopLevelSurfaceFromRaw(raw)
	if !ok {
		return nil
	}
	return &surface
}

func browserRuntimeTopLevelViewFromRaw(raw json.RawMessage) (browserTopLevelViewSummary, bool) {
	if view, ok := browserRuntimeDecodeNonEmptyRaw(raw, browserTopLevelViewEmpty); ok {
		return *view, true
	}
	return browserTopLevelViewSummary{}, false
}

func browserRuntimeTopLevelViewFromRawPtr(raw json.RawMessage) *browserTopLevelViewSummary {
	view, ok := browserRuntimeTopLevelViewFromRaw(raw)
	if !ok {
		return nil
	}
	return &view
}

func browserRuntimeWorkbenchSurfaceFromRaw(raw json.RawMessage) (browserRuntimeWorkbenchSurfaceSummary, bool) {
	if workbench, ok := browserRuntimeDecodeNonEmptyRaw(raw, browserUnifiedWorkbenchEmpty); ok {
		return *workbench, true
	}
	return browserRuntimeWorkbenchSurfaceSummary{}, false
}

func browserRuntimeWorkbenchSurfaceFromRawPtr(raw json.RawMessage) *browserRuntimeWorkbenchSurfaceSummary {
	workbench, ok := browserRuntimeWorkbenchSurfaceFromRaw(raw)
	if !ok {
		return nil
	}
	return &workbench
}

func browserRuntimeWorkbenchDisplayFromRaw(raw json.RawMessage) (browserRuntimeWorkbenchDisplaySummary, bool) {
	if display, ok := browserRuntimeDecodeNonEmptyRaw(raw, browserUnifiedWorkbenchDisplayEmpty); ok {
		return *display, true
	}
	return browserRuntimeWorkbenchDisplaySummary{}, false
}

func browserRuntimeWorkbenchDisplayFromRawPtr(raw json.RawMessage) *browserRuntimeWorkbenchDisplaySummary {
	display, ok := browserRuntimeWorkbenchDisplayFromRaw(raw)
	if !ok {
		return nil
	}
	return &display
}

func browserRuntimeRouteDescriptorFromRaw(raw json.RawMessage) (browserRuntimeRouteDescriptor, bool) {
	route, ok := browserRuntimeDecodeRaw[browserRuntimeRouteDescriptor](raw)
	if !ok || route == nil || *route == (browserRuntimeRouteDescriptor{}) {
		return browserRuntimeRouteDescriptor{}, false
	}
	return *route, true
}

func browserRuntimeStringFromRaw(raw json.RawMessage) string {
	value, ok := browserRuntimeDecodeRaw[string](raw)
	if !ok {
		return ""
	}
	return strings.TrimSpace(*value)
}

func browserRuntimeStringSliceFromRaw(raw json.RawMessage) []string {
	values, ok := browserRuntimeDecodeRaw[[]string](raw)
	if !ok {
		return nil
	}
	return mergeToolMetadataStrings(nil, *values)
}

func browserRuntimeTopLevelSummaryFromDiagnosticsExplanationRaw(raw json.RawMessage) (*browserTopLevelSummary, bool) {
	explanation, ok := browserRuntimeDecodeRaw[browserDiagnosticsExplanationSummary](raw)
	if !ok {
		return nil, false
	}
	summary := browserTopLevelSummaryFromDiagnosticsExplanation(explanation)
	return summary, summary != nil
}

func browserRuntimeTopLevelSummaryFromDiagnosticsExplanationRawPtr(raw json.RawMessage) *browserTopLevelSummary {
	summary, ok := browserRuntimeTopLevelSummaryFromDiagnosticsExplanationRaw(raw)
	if !ok {
		return nil
	}
	return summary
}

func browserRuntimeTopLevelSummaryFromRuntimeDiagnosticsExplanationRaw(raw json.RawMessage) (*browserTopLevelSummary, bool) {
	explanation, ok := browserRuntimeDecodeRaw[browserRuntimeDiagnosticsExplanationSummary](raw)
	if !ok {
		return nil, false
	}
	summary := browserTopLevelSummaryFromRuntimeDiagnosticsExplanation(explanation)
	return summary, summary != nil
}

func browserRuntimeTopLevelSummaryFromRuntimeDiagnosticsExplanationRawPtr(raw json.RawMessage) *browserTopLevelSummary {
	summary, ok := browserRuntimeTopLevelSummaryFromRuntimeDiagnosticsExplanationRaw(raw)
	if !ok {
		return nil
	}
	return summary
}

func browserRuntimeTopLevelSummaryFromResolverExplanationRaw(raw json.RawMessage) (*browserTopLevelSummary, bool) {
	explanation, ok := browserRuntimeDecodeRaw[browserRuntimeResolverExplanationSummary](raw)
	if !ok {
		return nil, false
	}
	summary := browserTopLevelSummaryFromResolverExplanation(explanation)
	return summary, summary != nil
}

func browserRuntimeTopLevelSummaryFromResolverExplanationRawPtr(raw json.RawMessage) *browserTopLevelSummary {
	summary, ok := browserRuntimeTopLevelSummaryFromResolverExplanationRaw(raw)
	if !ok {
		return nil
	}
	return summary
}

func browserRuntimeTopLevelSummaryFromWorkbenchDiagnosticsRaw(raw json.RawMessage) (*browserTopLevelSummary, bool) {
	diagnostics, ok := browserRuntimeDecodeRaw[browserRuntimeWorkbenchDiagnosticsSummary](raw)
	if !ok {
		return nil, false
	}
	summary := browserTopLevelSummaryFromWorkbenchDiagnostics(diagnostics)
	return summary, summary != nil
}

func browserRuntimeTopLevelSummaryFromWorkbenchDiagnosticsRawPtr(raw json.RawMessage) *browserTopLevelSummary {
	summary, ok := browserRuntimeTopLevelSummaryFromWorkbenchDiagnosticsRaw(raw)
	if !ok {
		return nil
	}
	return summary
}

func browserRuntimeMarshalTopLevelSummaryAlias(summary *browserTopLevelSummary) (json.RawMessage, bool) {
	if summary == nil {
		return nil, false
	}
	blob, err := json.Marshal(summary)
	if err != nil {
		return nil, false
	}
	return blob, true
}

func browserRuntimeDecodeRaw[T any](raw json.RawMessage) (*T, bool) {
	if len(raw) == 0 {
		return nil, false
	}
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, false
	}
	return &value, true
}

func browserRuntimeDecodeNonEmptyRaw[T any](raw json.RawMessage, empty func(T) bool) (*T, bool) {
	value, ok := browserRuntimeDecodeRaw[T](raw)
	if !ok {
		return nil, false
	}
	if empty != nil && empty(*value) {
		return nil, false
	}
	return value, true
}
