package pack

const (
	semanticToolNameConfigKey = "_pack_semantic_tool"
	artifactTypesConfigKey    = "_pack_artifact_types"
	semanticToolTagsConfigKey = "_pack_semantic_tool_tags"
)

func SemanticToolNameFromConfig(config map[string]any) string {
	if len(config) == 0 {
		return ""
	}
	raw, _ := config[semanticToolNameConfigKey].(string)
	return raw
}

func ArtifactTypesFromConfig(config map[string]any) []string {
	if len(config) == 0 {
		return nil
	}
	return runtimeMetadataStringsFromConfig(config[artifactTypesConfigKey])
}

func SemanticToolTagsFromConfig(config map[string]any) []string {
	if len(config) == 0 {
		return nil
	}
	return runtimeMetadataStringsFromConfig(config[semanticToolTagsConfigKey])
}

func applySemanticToolRuntimeMetadata(config map[string]any, semantic SemanticTool) {
	if config == nil {
		return
	}
	if semantic.Name != "" {
		config[semanticToolNameConfigKey] = semantic.Name
	}
	if len(semantic.ArtifactTypes) > 0 {
		config[artifactTypesConfigKey] = append([]string(nil), semantic.ArtifactTypes...)
	}
	if len(semantic.Tags) > 0 {
		config[semanticToolTagsConfigKey] = append([]string(nil), semantic.Tags...)
	}
}

func StripSemanticToolRuntimeMetadata(config map[string]any) {
	if len(config) == 0 {
		return
	}
	delete(config, semanticToolNameConfigKey)
	delete(config, artifactTypesConfigKey)
	delete(config, semanticToolTagsConfigKey)
}

func NormalizeArtifactTypes(raw any) []string {
	return normalizeRuntimeMetadataStrings(raw)
}

func NormalizeSemanticToolTags(raw any) []string {
	return normalizeRuntimeMetadataStrings(raw)
}

func runtimeMetadataStringsFromConfig(raw any) []string {
	if raw == nil {
		return nil
	}
	switch typed := raw.(type) {
	case []string:
		if len(typed) == 0 {
			return nil
		}
		return append([]string(nil), typed...)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				out = append(out, text)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	case string:
		if typed == "" {
			return nil
		}
		return []string{typed}
	default:
		return nil
	}
}

func normalizeRuntimeMetadataStrings(raw any) []string {
	if raw == nil {
		return nil
	}
	values := make([]string, 0)
	switch typed := raw.(type) {
	case []string:
		values = append(values, typed...)
	case []any:
		for _, item := range typed {
			if text, ok := item.(string); ok {
				values = append(values, text)
			}
		}
	case string:
		values = append(values, typed)
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, item := range values {
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
