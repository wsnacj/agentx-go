package extractors

import "testing"

func TestRunRegexHonorsHeaderAndFooterScope(t *testing.T) {
	text := "Report No: 2025\nHeader Only\nBody value: 100\nFooter marker: signed"

	header, ok := RunRegex(RegexInput{
		Text:        text,
		Scope:       "header",
		Pattern:     `Report No:\s*(\d+)`,
		HeaderLines: 2,
	})
	if !ok || header.Value != "2025" {
		t.Fatalf("expected header scoped match, got %#v ok=%t", header, ok)
	}

	if _, ok := RunRegex(RegexInput{
		Text:        text,
		Scope:       "header",
		Pattern:     `Body value:\s*(\d+)`,
		HeaderLines: 2,
	}); ok {
		t.Fatal("did not expect body text to match in header scope")
	}

	footer, ok := RunRegex(RegexInput{
		Text:        text,
		Scope:       "footer",
		Pattern:     `Footer marker:\s*(\w+)`,
		FooterLines: 1,
	})
	if !ok || footer.Value != "signed" {
		t.Fatalf("expected footer scoped match, got %#v ok=%t", footer, ok)
	}
}

func TestRunRegexCandidatesReturnsAllMatches(t *testing.T) {
	got := RunRegexCandidates(RegexInput{
		Text:    "Revenue: 100\nRevenue: 200\n",
		Pattern: `Revenue:\s*(\d+)`,
	})
	if len(got) != 2 {
		t.Fatalf("expected 2 regex candidates, got %#v", got)
	}
	if got[0].Value != "100" || got[1].Value != "200" {
		t.Fatalf("unexpected regex candidates: %#v", got)
	}
}
