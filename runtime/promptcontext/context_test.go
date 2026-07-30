package promptcontext

import (
	"testing"
	"time"
)

func TestBuildPreservesProvidedFieldsAndDefaultsNow(t *testing.T) {
	now := time.Date(2026, 3, 14, 9, 30, 0, 0, time.UTC)
	got := Build(BuildInput{
		Now:       now,
		Timezone:  "Asia/Shanghai",
		SessionID: "session-1",
		Model:     "gpt-5",
	})
	if !got.Now.Equal(now) {
		t.Fatalf("expected explicit now to be preserved, got %v", got.Now)
	}
	if got.Timezone != "Asia/Shanghai" || got.SessionID != "session-1" || got.Model != "gpt-5" {
		t.Fatalf("expected build fields to be preserved, got %#v", got)
	}

	auto := Build(BuildInput{})
	if auto.Now.IsZero() {
		t.Fatalf("expected zero input time to default to current time")
	}
}

func TestContextTimestampTextUsesTimezoneWhenValid(t *testing.T) {
	ctx := Context{
		Now:      time.Date(2026, 3, 14, 1, 30, 0, 0, time.UTC),
		Timezone: "Asia/Shanghai",
	}
	if got := ctx.TimestampText(); got != "2026-03-14T09:30:00+08:00" {
		t.Fatalf("expected timezone-adjusted timestamp, got %q", got)
	}
}

func TestContextTimestampTextFallsBackToRFC3339WhenTimezoneInvalid(t *testing.T) {
	now := time.Date(2026, 3, 14, 1, 30, 0, 0, time.UTC)
	ctx := Context{
		Now:      now,
		Timezone: "Mars/Olympus",
	}
	if got := ctx.TimestampText(); got != now.Format(time.RFC3339) {
		t.Fatalf("expected invalid timezone fallback, got %q", got)
	}

	ctx.Timezone = ""
	if got := ctx.TimestampText(); got != now.Format(time.RFC3339) {
		t.Fatalf("expected empty timezone fallback, got %q", got)
	}
}
