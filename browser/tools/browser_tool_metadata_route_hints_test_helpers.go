package tools

func browserResolvedRouteHintsForMetadataEntry(name string, metadata map[string]ToolMetadata) BrowserToolMetadataRouteHints {
	return ResolveBrowserToolMetadataRouteHints([]string{name}, metadata)
}

func browserResolvedRouteHintsForCapabilities(name string, capabilities []string) BrowserToolMetadataRouteHints {
	return browserResolvedRouteHintsForMetadataEntry(name, map[string]ToolMetadata{
		name: {
			Type:         "browser",
			Capabilities: append([]string(nil), capabilities...),
		},
	})
}
