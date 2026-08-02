package credentialhttp

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientRejectsCrossOriginRedirectBeforeSecondRequest(t *testing.T) {
	var targetHits atomic.Int64
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetHits.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer target.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/credential-target", http.StatusTemporaryRedirect)
	}))
	defer origin.Close()

	request, err := http.NewRequest(http.MethodPost, origin.URL, http.NoBody)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	request.Header.Set("X-Credential", "must-not-cross-origin")
	_, err = NewClient(time.Second).Do(request)
	if !errors.Is(err, ErrCrossOriginRedirect) {
		t.Fatalf("expected cross-origin redirect error, got %v", err)
	}
	if hits := targetHits.Load(); hits != 0 {
		t.Fatalf("cross-origin target received %d requests", hits)
	}
}

func TestClientAllowsBoundedSameOriginRedirect(t *testing.T) {
	var receivedCredential string
	server := httptest.NewUnstartedServer(nil)
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, server.URL+"/final", http.StatusFound)
			return
		}
		receivedCredential = r.Header.Get("X-Credential")
		w.WriteHeader(http.StatusNoContent)
	})
	server.Start()
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL+"/start", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	request.Header.Set("X-Credential", "same-origin")
	response, err := NewClient(time.Second).Do(request)
	if err != nil {
		t.Fatalf("same-origin redirect: %v", err)
	}
	response.Body.Close()
	if receivedCredential != "same-origin" {
		t.Fatalf("same-origin redirect credential = %q", receivedCredential)
	}
}

func TestClientRejectsRedirectChainBeyondLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		step, err := strconv.Atoi(r.URL.Query().Get("step"))
		if err != nil {
			http.Error(w, "invalid redirect step", http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/?step=%d", step+1), http.StatusFound)
	}))
	defer server.Close()

	_, err := NewClient(time.Second).Get(server.URL + "/?step=0")
	if !errors.Is(err, ErrRedirectLimit) {
		t.Fatalf("expected redirect limit error, got %v", err)
	}
}

func TestSameOriginUsesSchemeHostnameAndEffectivePort(t *testing.T) {
	tests := []struct {
		name  string
		left  string
		right string
		want  bool
	}{
		{
			name:  "default HTTP port",
			left:  "http://EXAMPLE.com/path",
			right: "http://example.com:80/other",
			want:  true,
		},
		{
			name:  "default HTTPS port",
			left:  "https://example.com/path",
			right: "https://example.com:443/other",
			want:  true,
		},
		{
			name:  "scheme differs",
			left:  "http://example.com/path",
			right: "https://example.com/path",
			want:  false,
		},
		{
			name:  "effective port differs",
			left:  "https://example.com/path",
			right: "https://example.com:8443/path",
			want:  false,
		},
		{
			name:  "hostname differs",
			left:  "https://example.com/path",
			right: "https://api.example.com/path",
			want:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			left, err := url.Parse(test.left)
			if err != nil {
				t.Fatalf("parse left URL: %v", err)
			}
			right, err := url.Parse(test.right)
			if err != nil {
				t.Fatalf("parse right URL: %v", err)
			}
			if got := sameOrigin(left, right); got != test.want {
				t.Fatalf("sameOrigin(%q, %q) = %t, want %t", test.left, test.right, got, test.want)
			}
		})
	}
}
