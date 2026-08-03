package retrieval

import "strings"

type SearchProviderKind string

const (
	SearchProviderKindStructuredSearch    SearchProviderKind = "structured_search"
	SearchProviderKindSynthesizedCitation SearchProviderKind = "synthesized_citation"
)

type SearchResultKind string

const (
	SearchResultKindStructuredResults   SearchResultKind = "structured_results"
	SearchResultKindSynthesizedCitation SearchResultKind = "synthesized_citation"
)

func ResolveSearchProviderKind(provider string) SearchProviderKind {
	switch NormalizeSearchProvider(provider) {
	case "perplexity", "openrouter":
		return SearchProviderKindSynthesizedCitation
	default:
		return SearchProviderKindStructuredSearch
	}
}

func ResolveSearchResultKind(provider string) SearchResultKind {
	switch ResolveSearchProviderKind(provider) {
	case SearchProviderKindSynthesizedCitation:
		return SearchResultKindSynthesizedCitation
	default:
		return SearchResultKindStructuredResults
	}
}

func IsSynthesizedSearchProvider(provider string) bool {
	return ResolveSearchProviderKind(provider) == SearchProviderKindSynthesizedCitation
}

func searchProviderKindString(kind SearchProviderKind) string {
	return strings.TrimSpace(string(kind))
}

func searchResultKindString(kind SearchResultKind) string {
	return strings.TrimSpace(string(kind))
}
