package main

import "testing"

func TestFixedVersionResearchConsumer(t *testing.T) {
	got, err := run()
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	const want = "agentx-research-ok:public-news-brief-pack:public_news_brief_lookup_v1:company-research-pack:company_research_lookup_v1:true:true"
	if got != want {
		t.Fatalf("run() = %q, want %q", got, want)
	}
}
