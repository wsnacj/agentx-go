package extractors_test

import (
	"github.com/wsnacj/agentx-go/document/pipeline/extractors"
	"testing"
)

func TestParseLooseJSONObject_Basic(t *testing.T) {
	cases := []string{
		"{\"a\":1}",
		"```json\n{\"a\":1}\n```",
		"Answer: {\"a\":1} thanks",
		"{\"a\":1,}", // trailing comma
	}
	for _, s := range cases {
		m, err := extractors.ParseLooseJSONObject(s)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", s, err)
		}
		if _, ok := m["a"]; !ok {
			t.Fatalf("key 'a' not found for %q", s)
		}
	}
}

func TestParseLoosePagesMap(t *testing.T) {
	s := "```json\n{\n  \"cover\": [1],\n  \"balance_sheet\": [88, 89,],\n  \"extra\": [1.0, 2, 3]\n}\n```"
	m, err := extractors.ParseLoosePagesMap(s)
	if err != nil {
		t.Fatalf("ParseLoosePagesMap failed: %v", err)
	}
	if len(m["balance_sheet"]) != 2 || m["balance_sheet"][0] != 88 || m["balance_sheet"][1] != 89 {
		t.Fatalf("unexpected balance_sheet: %v", m["balance_sheet"])
	}
	if len(m["extra"]) != 3 {
		t.Fatalf("unexpected extra: %v", m["extra"])
	}
}
