package channel

import (
	"testing"
	"time"
)

func TestNewDeduperDefaultsTTLWhenNonPositive(t *testing.T) {
	d := NewDeduper(0)
	if d.ttl != 10*time.Minute {
		t.Fatalf("expected default ttl, got %s", d.ttl)
	}
}

func TestDeduperBeginForPrunesExpiredEntries(t *testing.T) {
	base := time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC)
	now := base
	d := NewDeduper(time.Minute)
	d.now = func() time.Time { return now }

	if !d.BeginFor("k1", time.Second) {
		t.Fatalf("expected first reservation to succeed")
	}
	if got := d.items["k1"].state; got != "pending" {
		t.Fatalf("unexpected state after begin: %q", got)
	}

	now = base.Add(500 * time.Millisecond)
	if d.BeginFor("k1", time.Second) {
		t.Fatalf("expected duplicate reservation to stay blocked before expiry")
	}

	now = base.Add(2 * time.Second)
	if !d.BeginFor("k1", time.Second) {
		t.Fatalf("expected reservation to succeed after expiry prune")
	}
}

func TestDeduperCompleteForMarksDoneWithCustomTTL(t *testing.T) {
	base := time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC)
	now := base
	d := NewDeduper(time.Minute)
	d.now = func() time.Time { return now }

	d.CompleteFor("k1", 2*time.Second)
	item, ok := d.items["k1"]
	if !ok {
		t.Fatalf("expected completion entry to be recorded")
	}
	if item.state != "done" {
		t.Fatalf("unexpected completion state: %q", item.state)
	}
	if !item.expiresAt.Equal(base.Add(2 * time.Second)) {
		t.Fatalf("unexpected completion expiry: got=%s want=%s", item.expiresAt, base.Add(2*time.Second))
	}

	now = base.Add(time.Second)
	if d.Begin("k1") {
		t.Fatalf("expected completed key to stay blocked before expiry")
	}
}

func TestBuildContentDedupeKeyFallsBackToChatID(t *testing.T) {
	msg := Message{
		ChatID: "chat_1",
		UserID: "user_1",
		Text:   "  hello   world  ",
	}
	if got := BuildContentDedupeKey(msg); got != "content:chat_1:user_1:hello world" {
		t.Fatalf("unexpected dedupe key: %q", got)
	}
}

func TestBuildContentDedupeKeyReturnsBlankWhenScopeOrTextMissing(t *testing.T) {
	tests := []Message{
		{SessionID: "session_1", UserID: "user_1", Text: "   "},
		{UserID: "user_1", Text: "hello"},
	}

	for _, msg := range tests {
		if got := BuildContentDedupeKey(msg); got != "" {
			t.Fatalf("expected blank dedupe key for %#v, got %q", msg, got)
		}
	}
}
