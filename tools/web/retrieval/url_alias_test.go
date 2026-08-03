package retrieval

import "testing"

func TestFirstHTTPURLAlias(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]any
		wantField string
		wantURL   string
		wantOK    bool
	}{
		{
			name:      "href",
			params:    map[string]any{"href": " https://example.com/article "},
			wantField: "href",
			wantURL:   "https://example.com/article",
			wantOK:    true,
		},
		{
			name:      "sourceUrl",
			params:    map[string]any{"sourceUrl": "https://example.com/source"},
			wantField: "sourceUrl",
			wantURL:   "https://example.com/source",
			wantOK:    true,
		},
		{
			name:   "explicit url blocks alias",
			params: map[string]any{"url": "", "href": "https://example.com/article"},
		},
		{
			name:   "relative URL rejected",
			params: map[string]any{"href": "/article"},
		},
		{
			name:   "non-http URL rejected",
			params: map[string]any{"href": "ftp://example.com/article"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			field, rawURL, ok := FirstHTTPURLAlias(tc.params)
			if ok != tc.wantOK || field != tc.wantField || rawURL != tc.wantURL {
				t.Fatalf("FirstHTTPURLAlias() field=%q url=%q ok=%v, want field=%q url=%q ok=%v", field, rawURL, ok, tc.wantField, tc.wantURL, tc.wantOK)
			}
		})
	}
}
