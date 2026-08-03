package retrieval

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestWebFetchCacheReadWrite(t *testing.T) {
	ResetWebFetchCache()
	key := WebFetchCacheKey("https://example.com/page", 1000, 2000, "markdown", "sig")
	WriteWebFetchCache(key, []byte(`{"ok":true}`), time.Minute)
	got, ok := ReadWebFetchCache(key)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if !bytes.Equal(got, []byte(`{"ok":true}`)) {
		t.Fatalf("unexpected payload: %q", string(got))
	}
}

func TestWebFetchCacheKeyHashesFullRequestIdentity(t *testing.T) {
	one := WebFetchCacheKey("https://example.com/page?token=one", 1000, 2000, "markdown", "sig")
	again := WebFetchCacheKey("https://example.com/page?token=one", 1000, 2000, "markdown", "sig")
	two := WebFetchCacheKey("https://example.com/page?token=two", 1000, 2000, "markdown", "sig")
	if one == "" || one != again || one == two {
		t.Fatalf("unexpected cache identities one=%q again=%q two=%q", one, again, two)
	}
	for _, raw := range []string{"example.com", "token", "one"} {
		if strings.Contains(one, raw) {
			t.Fatalf("cache key retained %q: %q", raw, one)
		}
	}
}

func TestWebFetchCacheExpires(t *testing.T) {
	ResetWebFetchCache()
	key := WebFetchCacheKey("https://example.com/page", 1000, 2000, "markdown", "sig")
	WriteWebFetchCache(key, []byte(`{"ok":true}`), 10*time.Millisecond)
	time.Sleep(30 * time.Millisecond)
	if _, ok := ReadWebFetchCache(key); ok {
		t.Fatal("expected cache miss after ttl")
	}
}
