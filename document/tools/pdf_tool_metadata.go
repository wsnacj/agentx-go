package tools

import types "github.com/wsnacj/agentx-go/components/llm"
import agentxtools "github.com/wsnacj/agentx-go/tools"

func pdfToolMetadata(defs []types.Tool) map[string]ToolMetadata {
	if len(defs) == 0 {
		return nil
	}
	out := map[string]ToolMetadata{}
	for _, def := range defs {
		name := agentxtools.NormalizeToolName(def.Function.Name)
		if name == "" || !isPDFToolName(name) {
			continue
		}
		meta := pdfToolMetadataForDefinition(name)
		if meta.Source == "" && IsBuiltinCoreTool(name) {
			meta.Source = ToolSourceBuiltin
		}
		out[name] = meta
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func PDFToolMetadata(defs []types.Tool) map[string]ToolMetadata {
	return pdfToolMetadata(defs)
}

func inferPDFToolMetadata(defs []types.Tool) map[string]ToolMetadata {
	return pdfToolMetadata(defs)
}

func isPDFToolName(name string) bool {
	return IsPDFUnifiedToolName(name) || IsPDFSpecialistToolName(name)
}

func pdfToolMetadataForDefinition(name string) ToolMetadata {
	capabilities := []string{"pdf", "document", "read", "network", "pdf_remote_input", "external_content"}
	auditTags := []string{"pdf", "external_content"}
	groups := []string{PDFSurfaceForToolName(name)}
	switch name {
	case "pdf":
		capabilities = append(capabilities, "pdf_kind:unified", "analysis", "document_set", "pdf_multi_input", "pdf_native_provider", "pdf_native_vendor:anthropic", "pdf_native_vendor:gemini")
	case "pdf_extract":
		capabilities = append(capabilities, "pdf_kind:extract")
	case "pdf_read_pages":
		capabilities = append(capabilities, "pdf_kind:read_pages")
	case "pdf_outline":
		capabilities = append(capabilities, "pdf_kind:outline")
	case "pdf_analyze":
		capabilities = append(capabilities, "pdf_kind:analyze", "analysis", "vision", "document_set", "pdf_multi_input", "pdf_native_provider", "pdf_native_vendor:anthropic", "pdf_native_vendor:gemini", "artifact_output")
		auditTags = append(auditTags, "artifact_output")
	case "pdf_extract_structured":
		capabilities = append(capabilities, "pdf_kind:extract_structured", "analysis", "structured", "vision", "document_set", "pdf_multi_input", "pdf_native_provider", "pdf_native_vendor:anthropic", "pdf_native_vendor:gemini", "artifact_output")
		auditTags = append(auditTags, "artifact_output")
	}
	readOnly := builtinToolMetadataBoolPtr(true)
	concurrencySafe := builtinToolMetadataBoolPtr(true)
	switch name {
	case "pdf_analyze", "pdf_extract_structured":
		readOnly = builtinToolMetadataBoolPtr(false)
		concurrencySafe = builtinToolMetadataBoolPtr(false)
	}
	return ToolMetadata{
		Type:            "pdf",
		Groups:          mergeToolMetadataStrings(nil, groups),
		Capabilities:    mergeToolMetadataStrings(nil, capabilities),
		AuditTags:       mergeToolMetadataStrings(nil, auditTags),
		ReadOnly:        readOnly,
		ConcurrencySafe: concurrencySafe,
	}
}

// PDFToolMetadataForName returns canonical descriptive metadata for one PDF
// entrypoint. It grants no authorization or backend access.
func PDFToolMetadataForName(name string) ToolMetadata { return pdfToolMetadataForDefinition(name) }
