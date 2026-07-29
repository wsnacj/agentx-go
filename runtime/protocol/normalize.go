package protocol

import "strings"

func NormalizeRunEvent(event RunEvent) RunEvent {
	event.Envelope = normalizeEnvelope(event.Envelope, RunEventSchemaV1, "")
	event.SourceEvent = trim(event.SourceEvent)
	event.SourceEventID = trim(event.SourceEventID)
	event.Level = trimLower(event.Level)
	event.Status = trimLower(event.Status)
	event.Reason = trim(event.Reason)
	event.Stage = trim(event.Stage)
	event.ToolName = trimLower(event.ToolName)
	event.ToolCallID = trim(event.ToolCallID)
	event.ModelID = trim(event.ModelID)
	event.Provider = trim(event.Provider)
	event.ExecutionContractID = trim(event.ExecutionContractID)
	event.ExecutionContractDiff = trimStringSlice(event.ExecutionContractDiff)
	event.Error = normalizeErrorInfoPtr(event.Error)
	event.Usage = normalizeUsagePtr(event.Usage)
	event.RuntimeDecision = normalizeRuntimeDecisionPtr(event.RuntimeDecision)
	event.Attrs = copyAttrs(event.Attrs)
	return event
}

func NormalizeTraceSpan(span TraceSpan) TraceSpan {
	span.Envelope = normalizeEnvelope(span.Envelope, TraceSpanSchemaV1, "")
	span.Type = trimLower(span.Type)
	span.Name = trim(span.Name)
	span.Status = trimLower(span.Status)
	span.Usage = normalizeUsagePtr(span.Usage)
	span.Error = normalizeErrorInfoPtr(span.Error)
	span.Attrs = copyAttrs(span.Attrs)
	return span
}

func NormalizeToolExecutionPlan(plan ToolExecutionPlan) ToolExecutionPlan {
	plan.Envelope = normalizeEnvelope(plan.Envelope, ToolExecutionPlanSchemaV1, KindToolPlanCreated)
	plan.PlanID = trim(plan.PlanID)
	plan.ExecutionContractID = trim(plan.ExecutionContractID)
	plan.ExecutionContractDiff = trimStringSlice(plan.ExecutionContractDiff)
	for idx := range plan.Calls {
		plan.Calls[idx] = normalizeToolPlanCall(plan.Calls[idx])
	}
	for idx := range plan.Interruptions {
		plan.Interruptions[idx] = normalizeToolPlanInterruption(plan.Interruptions[idx])
	}
	for idx := range plan.BlockedCalls {
		plan.BlockedCalls[idx] = normalizeToolPlanBlockedCall(plan.BlockedCalls[idx])
	}
	return plan
}

func NormalizeHandoffRecord(record HandoffRecord) HandoffRecord {
	record.Envelope = normalizeEnvelope(record.Envelope, HandoffSchemaV1, "")
	record.HandoffID = trim(record.HandoffID)
	record.HandoffKind = trimLower(record.HandoffKind)
	record.Source = normalizeHandoffEndpoint(record.Source)
	record.Target = normalizeHandoffEndpoint(record.Target)
	record.InputFilter = normalizeHandoffInputFilter(record.InputFilter)
	record.Isolation = normalizeHandoffIsolation(record.Isolation)
	record.Status = trimLower(record.Status)
	record.Reason = trim(record.Reason)
	record.Error = normalizeErrorInfoPtr(record.Error)
	record.Attrs = copyAttrs(record.Attrs)
	return record
}

func NormalizeSandboxManifest(manifest SandboxManifest) SandboxManifest {
	manifest.Envelope = normalizeEnvelope(manifest.Envelope, SandboxManifestSchemaV1, KindSandboxManifest)
	manifest.ManifestID = trim(manifest.ManifestID)
	manifest.Root = trim(manifest.Root)
	manifest.Backend = trimLower(manifest.Backend)
	manifest.Platform.GOOS = trimLower(manifest.Platform.GOOS)
	manifest.Platform.Arch = trimLower(manifest.Platform.Arch)
	for idx := range manifest.Entries {
		manifest.Entries[idx] = normalizeSandboxEntry(manifest.Entries[idx])
	}
	for idx := range manifest.Environment {
		manifest.Environment[idx] = normalizeSandboxEnvVar(manifest.Environment[idx])
	}
	for idx := range manifest.PathGrants {
		manifest.PathGrants[idx] = normalizeSandboxPathGrant(manifest.PathGrants[idx])
	}
	manifest.Network.Mode = trimLower(manifest.Network.Mode)
	manifest.Network.Allow = trimStringSlice(manifest.Network.Allow)
	manifest.Network.Deny = trimStringSlice(manifest.Network.Deny)
	manifest.CommandPolicy.Allow = trimStringSlice(manifest.CommandPolicy.Allow)
	manifest.CommandPolicy.Deny = trimStringSlice(manifest.CommandPolicy.Deny)
	manifest.Reason = trim(manifest.Reason)
	return manifest
}

func NormalizeArtifactVersion(version ArtifactVersion) ArtifactVersion {
	version.Envelope = normalizeEnvelope(version.Envelope, ArtifactVersionSchemaV1, "")
	version.ArtifactID = trim(version.ArtifactID)
	version.CanonicalURI = trim(version.CanonicalURI)
	version.Scope = trimLower(version.Scope)
	version.MIMEType = trimLower(version.MIMEType)
	version.CreatedBy.ToolName = trimLower(version.CreatedBy.ToolName)
	version.CreatedBy.ToolCallID = trim(version.CreatedBy.ToolCallID)
	version.CreatedBy.AgentID = trim(version.CreatedBy.AgentID)
	version.CreatedBy.Producer = trim(version.CreatedBy.Producer)
	version.Metadata = copyAttrs(version.Metadata)
	version.Payload.Storage = trimLower(version.Payload.Storage)
	version.Payload.BlobRef = trim(version.Payload.BlobRef)
	version.Payload.Path = trim(version.Payload.Path)
	version.Payload.URL = trim(version.Payload.URL)
	version.Payload.Digest = trim(version.Payload.Digest)
	version.Payload.ExternalID = trim(version.Payload.ExternalID)
	return version
}

func NormalizeArtifactLink(link ArtifactLink) ArtifactLink {
	link.SchemaVersion = trim(link.SchemaVersion)
	if link.SchemaVersion == "" {
		link.SchemaVersion = ArtifactLinkSchemaV1
	}
	link.SourceArtifactID = trim(link.SourceArtifactID)
	link.TargetArtifactID = trim(link.TargetArtifactID)
	link.Relation = trimLower(link.Relation)
	return link
}

func normalizeEnvelope(envelope Envelope, schemaVersion string, defaultKind string) Envelope {
	envelope.SchemaVersion = trim(envelope.SchemaVersion)
	if envelope.SchemaVersion == "" {
		envelope.SchemaVersion = schemaVersion
	}
	envelope.Kind = trimLower(envelope.Kind)
	if envelope.Kind == "" {
		envelope.Kind = defaultKind
	}
	envelope.RunID = trim(envelope.RunID)
	envelope.RootRunID = trim(envelope.RootRunID)
	envelope.ParentRunID = trim(envelope.ParentRunID)
	envelope.SessionID = trim(envelope.SessionID)
	envelope.TurnID = trim(envelope.TurnID)
	envelope.BranchID = trim(envelope.BranchID)
	envelope.NodeExecID = trim(envelope.NodeExecID)
	envelope.TraceID = trim(envelope.TraceID)
	envelope.SpanID = trim(envelope.SpanID)
	envelope.ParentSpanID = trim(envelope.ParentSpanID)
	if envelope.TimestampUnixMilli < 0 {
		envelope.TimestampUnixMilli = 0
	}
	return envelope
}

func normalizeToolPlanCall(call ToolPlanCall) ToolPlanCall {
	call.ToolCallID = trim(call.ToolCallID)
	call.ToolName = trimLower(call.ToolName)
	call.ArgumentsHash = trim(call.ArgumentsHash)
	call.ArgumentsSummary = trim(call.ArgumentsSummary)
	call.Origin = trimLower(call.Origin)
	call.Category = trimLower(call.Category)
	call.ExecutionMode = trimLower(call.ExecutionMode)
	call.IdempotencyKey = trim(call.IdempotencyKey)
	call.ExpectedSideEffect = trimLower(call.ExpectedSideEffect)
	call.Status = trimLower(call.Status)
	call.Reason = trim(call.Reason)
	call.ErrorCode = trim(call.ErrorCode)
	call.RuntimeDecision = normalizeRuntimeDecisionPtr(call.RuntimeDecision)
	return call
}

func normalizeToolPlanInterruption(interruption ToolPlanInterruption) ToolPlanInterruption {
	interruption.InterruptionID = trim(interruption.InterruptionID)
	interruption.ToolCallID = trim(interruption.ToolCallID)
	interruption.Type = trimLower(interruption.Type)
	interruption.ApprovalID = trim(interruption.ApprovalID)
	interruption.Status = trimLower(interruption.Status)
	interruption.Reason = trim(interruption.Reason)
	return interruption
}

func normalizeToolPlanBlockedCall(call ToolPlanBlockedCall) ToolPlanBlockedCall {
	call.ToolCallID = trim(call.ToolCallID)
	call.ToolName = trimLower(call.ToolName)
	call.Reason = trim(call.Reason)
	call.ErrorCode = trim(call.ErrorCode)
	return call
}

func normalizeHandoffEndpoint(endpoint HandoffEndpoint) HandoffEndpoint {
	endpoint.AgentID = trim(endpoint.AgentID)
	endpoint.PackID = trim(endpoint.PackID)
	endpoint.WorkflowID = trim(endpoint.WorkflowID)
	endpoint.NodeID = trim(endpoint.NodeID)
	endpoint.RunID = trim(endpoint.RunID)
	endpoint.ExpectedRunID = trim(endpoint.ExpectedRunID)
	endpoint.SessionID = trim(endpoint.SessionID)
	endpoint.TaskID = trim(endpoint.TaskID)
	endpoint.ToolCallID = trim(endpoint.ToolCallID)
	return endpoint
}

func normalizeHandoffInputFilter(filter HandoffInputFilter) HandoffInputFilter {
	filter.Mode = trimLower(filter.Mode)
	filter.IncludedArtifactIDs = trimStringSlice(filter.IncludedArtifactIDs)
	filter.IncludedMessageIDs = trimStringSlice(filter.IncludedMessageIDs)
	filter.Reason = trim(filter.Reason)
	return filter
}

func normalizeHandoffIsolation(isolation HandoffIsolation) HandoffIsolation {
	isolation.Scope = trim(isolation.Scope)
	isolation.Branch = trim(isolation.Branch)
	return isolation
}

func normalizeSandboxEntry(entry SandboxEntry) SandboxEntry {
	entry.Path = trim(entry.Path)
	entry.Kind = trimLower(entry.Kind)
	entry.Source = trim(entry.Source)
	entry.Mode = trimLower(entry.Mode)
	return entry
}

func normalizeSandboxEnvVar(env SandboxEnvVar) SandboxEnvVar {
	env.Name = trim(env.Name)
	env.Source = trimLower(env.Source)
	env.Description = trim(env.Description)
	return env
}

func normalizeSandboxPathGrant(grant SandboxPathGrant) SandboxPathGrant {
	grant.Path = trim(grant.Path)
	grant.Mode = trimLower(grant.Mode)
	grant.Reason = trim(grant.Reason)
	return grant
}

func normalizeErrorInfoPtr(info *ErrorInfo) *ErrorInfo {
	if info == nil {
		return nil
	}
	out := *info
	out.Class = trimLower(out.Class)
	out.Code = trim(out.Code)
	out.Reason = trim(out.Reason)
	out.Message = trim(out.Message)
	return &out
}

func normalizeRuntimeDecisionPtr(decision *RuntimeDecisionSnapshot) *RuntimeDecisionSnapshot {
	if decision == nil {
		return nil
	}
	out := *decision
	out.Action = trimLower(out.Action)
	out.Reason = trim(out.Reason)
	out.Detail = trim(out.Detail)
	out.DecisionSubject = trim(out.DecisionSubject)
	out.TargetKind = trimLower(out.TargetKind)
	out.PolicySource = trim(out.PolicySource)
	out.ControlSource = trim(out.ControlSource)
	out.EnforcementSurface = trim(out.EnforcementSurface)
	return &out
}

func normalizeUsagePtr(usage *Usage) *Usage {
	if usage == nil {
		return nil
	}
	out := *usage
	out.ModelID = trim(out.ModelID)
	out.Provider = trimLower(out.Provider)
	return &out
}

func copyAttrs(attrs map[string]any) map[string]any {
	if len(attrs) == 0 {
		return nil
	}
	out := make(map[string]any, len(attrs))
	for key, value := range attrs {
		trimmed := trim(key)
		if trimmed == "" {
			continue
		}
		out[trimmed] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func trimStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := trim(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func trim(value string) string {
	return strings.TrimSpace(value)
}

func trimLower(value string) string {
	return strings.ToLower(trim(value))
}
