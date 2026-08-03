package retrieval

import "testing"

func TestDefaultSearchEndpointForBaiduUsesAISearchWebSearch(t *testing.T) {
	got := DefaultSearchEndpointForProvider("baidu")
	want := "https://qianfan.baidubce.com/v2/ai_search/web_search"
	if got != want {
		t.Fatalf("expected baidu default endpoint %q, got %q", want, got)
	}
}
