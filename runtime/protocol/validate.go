package protocol

import (
	"fmt"
	"strings"
)

func ValidateRunEvent(event RunEvent) error {
	event = NormalizeRunEvent(event)
	if err := validateEnvelope(event.Envelope, RunEventSchemaV1, true, true); err != nil {
		return err
	}
	return nil
}

func ValidateTraceSpan(span TraceSpan) error {
	span = NormalizeTraceSpan(span)
	if err := validateEnvelope(span.Envelope, TraceSpanSchemaV1, true, true); err != nil {
		return err
	}
	if span.TraceID == "" {
		return requiredField("trace_id")
	}
	if span.SpanID == "" {
		return requiredField("span_id")
	}
	if span.Type == "" {
		return requiredField("type")
	}
	return nil
}

func ValidateToolExecutionPlan(plan ToolExecutionPlan) error {
	plan = NormalizeToolExecutionPlan(plan)
	if err := validateEnvelope(plan.Envelope, ToolExecutionPlanSchemaV1, true, true); err != nil {
		return err
	}
	if plan.PlanID == "" {
		return requiredField("plan_id")
	}
	for idx, call := range plan.Calls {
		if call.ToolCallID == "" {
			return indexedRequiredField("calls", idx, "tool_call_id")
		}
		if call.ToolName == "" {
			return indexedRequiredField("calls", idx, "tool_name")
		}
	}
	return nil
}

func ValidateHandoffRecord(record HandoffRecord) error {
	record = NormalizeHandoffRecord(record)
	if err := validateEnvelope(record.Envelope, HandoffSchemaV1, true, true); err != nil {
		return err
	}
	if record.HandoffID == "" {
		return requiredField("handoff_id")
	}
	if !handoffEndpointPresent(record.Source) {
		return requiredField("source")
	}
	if !handoffEndpointPresent(record.Target) {
		return requiredField("target")
	}
	return nil
}

func ValidateSandboxManifest(manifest SandboxManifest) error {
	manifest = NormalizeSandboxManifest(manifest)
	if err := validateEnvelope(manifest.Envelope, SandboxManifestSchemaV1, true, true); err != nil {
		return err
	}
	if manifest.ManifestID == "" {
		return requiredField("manifest_id")
	}
	if manifest.Root == "" {
		return requiredField("root")
	}
	for idx, entry := range manifest.Entries {
		if entry.Path == "" {
			return indexedRequiredField("entries", idx, "path")
		}
		if pathHasParentSegment(entry.Path) {
			return fmt.Errorf("agentx/runtime/protocol: entries[%d].path must not escape root", idx)
		}
	}
	return nil
}

func ValidateArtifactVersion(version ArtifactVersion) error {
	version = NormalizeArtifactVersion(version)
	if err := validateEnvelope(version.Envelope, ArtifactVersionSchemaV1, false, true); err != nil {
		return err
	}
	if version.ArtifactID == "" {
		return requiredField("artifact_id")
	}
	if version.Version <= 0 {
		return requiredField("version")
	}
	if version.Kind == "" {
		return requiredField("kind")
	}
	return nil
}

func ValidateArtifactLink(link ArtifactLink) error {
	link = NormalizeArtifactLink(link)
	if link.SchemaVersion != ArtifactLinkSchemaV1 {
		return fmt.Errorf("agentx/runtime/protocol: unsupported schema_version %q", link.SchemaVersion)
	}
	if link.SourceArtifactID == "" {
		return requiredField("source_artifact_id")
	}
	if link.TargetArtifactID == "" {
		return requiredField("target_artifact_id")
	}
	if link.Relation == "" {
		return requiredField("relation")
	}
	return nil
}

func validateEnvelope(envelope Envelope, schemaVersion string, requireRunID bool, requireKind bool) error {
	if envelope.SchemaVersion != schemaVersion {
		return fmt.Errorf("agentx/runtime/protocol: unsupported schema_version %q", envelope.SchemaVersion)
	}
	if requireKind && envelope.Kind == "" {
		return requiredField("kind")
	}
	if requireRunID && envelope.RunID == "" {
		return requiredField("run_id")
	}
	return nil
}

func handoffEndpointPresent(endpoint HandoffEndpoint) bool {
	return endpoint.AgentID != "" ||
		endpoint.PackID != "" ||
		endpoint.WorkflowID != "" ||
		endpoint.NodeID != "" ||
		endpoint.RunID != "" ||
		endpoint.ExpectedRunID != "" ||
		endpoint.SessionID != "" ||
		endpoint.TaskID != "" ||
		endpoint.ToolCallID != ""
}

func pathHasParentSegment(value string) bool {
	normalized := strings.ReplaceAll(value, "\\", "/")
	for _, part := range strings.Split(normalized, "/") {
		if part == ".." {
			return true
		}
	}
	return false
}

func requiredField(name string) error {
	return fmt.Errorf("agentx/runtime/protocol: %s is required", name)
}

func indexedRequiredField(collection string, index int, name string) error {
	return fmt.Errorf("agentx/runtime/protocol: %s[%d].%s is required", collection, index, name)
}
