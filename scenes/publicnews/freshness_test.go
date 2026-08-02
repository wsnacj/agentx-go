package publicnews

import (
	"testing"
	"time"
)

func TestLatestNewsPublishedAfterAtResolvesRecentWeek(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	now := time.Date(2026, 7, 14, 18, 45, 0, 0, location)
	intent := LatestNewsLookupIntent{Freshness: map[string]any{
		"relative_date_hint": "最近一周",
	}}
	if got := LatestNewsPublishedAfterAt(intent, now); got != "2026-07-07T00:00:00+08:00" {
		t.Fatalf("recent-week cutoff = %q", got)
	}
	if LatestNewsSourceWithinFreshnessWindowAt(LatestNewsLookupSource{
		PublishedAt: "2026-07-01T13:25:00+08:00",
	}, intent, now) {
		t.Fatal("expected source before the resolved recent-week cutoff to fail")
	}
	if !LatestNewsSourceWithinFreshnessWindowAt(LatestNewsLookupSource{
		PublishedAt: "2026-07-07T00:00:00+08:00",
	}, intent, now) {
		t.Fatal("expected source on the resolved cutoff to pass")
	}
}

func TestLatestNewsPublishedAfterAtPrefersExplicitCutoff(t *testing.T) {
	intent := LatestNewsLookupIntent{Freshness: map[string]any{
		"relative_date_hint": "最近一周",
		"published_after":    "2026-07-10T00:00:00Z",
	}}
	if got := LatestNewsPublishedAfterAt(intent, time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)); got != "2026-07-10T00:00:00Z" {
		t.Fatalf("explicit cutoff = %q", got)
	}
}

func TestLatestNewsPublishedAfterAtUsesExplicitUserPhraseWhenStructuredHintIsMissing(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	intent := LatestNewsLookupIntent{UserMessage: "帮我查 Anthropic 最近一周有什么重要新闻。"}
	got := LatestNewsPublishedAfterAt(intent, time.Date(2026, 7, 14, 18, 45, 0, 0, location))
	if got != "2026-07-07T00:00:00+08:00" {
		t.Fatalf("user-message cutoff = %q", got)
	}
}

func TestLatestNewsPublishedAfterAtDefaultsGenericRecentHintToThirtyDays(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	intent := LatestNewsLookupIntent{Freshness: map[string]any{
		"mode":               "latest",
		"relative_date_hint": "最近",
		"require_latest":     true,
	}}
	got := LatestNewsPublishedAfterAt(intent, time.Date(2026, 7, 18, 18, 45, 0, 0, location))
	if got != "2026-06-18T00:00:00+08:00" {
		t.Fatalf("generic-recent cutoff = %q", got)
	}
}
