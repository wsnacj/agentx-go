package retrieval

import "testing"

func TestDefaultSearchEndpointForBaiduUsesAISearchWebSearch(t *testing.T) {
	got := DefaultSearchEndpointForProvider("baidu")
	want := "https://qianfan.baidubce.com/v2/ai_search/web_search"
	if got != want {
		t.Fatalf("expected baidu default endpoint %q, got %q", want, got)
	}
}

func TestDoubaoSearchProviderNamesAndEndpoints(t *testing.T) {
	for _, alias := range []string{"ark", "arksearch", "doubao", "doubao_search_custom", SearchProviderDoubaoCustom} {
		if got := NormalizeSearchProvider(alias); got != SearchProviderDoubaoCustom {
			t.Fatalf("NormalizeSearchProvider(%q) = %q", alias, got)
		}
	}
	if got := DefaultSearchEndpointForProvider(SearchProviderDoubaoCustom); got != DefaultSearchDoubaoCustomURL {
		t.Fatalf("custom endpoint = %q", got)
	}
	if got := DefaultSearchEndpointForProvider(SearchProviderDoubaoGlobal); got != DefaultSearchDoubaoGlobalURL {
		t.Fatalf("global endpoint = %q", got)
	}
	if !IsSupportedSearchProvider(SearchProviderDoubaoGlobal) {
		t.Fatal("doubao_global should be supported")
	}
	if SearchProviderAllowsCache(SearchProviderDoubaoCustom) || SearchProviderAllowsCache(SearchProviderDoubaoGlobal) {
		t.Fatal("doubao search content must not enter the process cache")
	}
}
