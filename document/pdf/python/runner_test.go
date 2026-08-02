package python

import (
	"reflect"
	"testing"

	pdfparser "github.com/wsnacj/agentx-go/document/pdf"
)

func TestArgumentsPreserveLegacyOrder(t *testing.T) {
	runner := &Runner{config: Config{ScriptPath: "/bundle/pdfparser.py"}}
	got := runner.arguments(pdfparser.RunRequest{
		PDFPath: "/input/report.pdf",
		Options: pdfparser.PDFParserOptions{
			OutputFormat:     "json",
			PageRange:        "1-3",
			TableEngine:      "hybrid",
			NeedCharacter:    true,
			ExtractImages:    true,
			HighAccuracyMode: true,
		},
	})
	want := []string{
		"/bundle/pdfparser.py", "/input/report.pdf",
		"--output-format", "json", "--page-range", "1-3", "--table-engine", "hybrid",
		"--need-character", "--extract-images", "--high-accuracy",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("arguments = %#v, want %#v", got, want)
	}
}

func TestBundledScriptPath(t *testing.T) {
	path, err := BundledScriptPath()
	if err != nil {
		t.Fatal(err)
	}
	if path == "" {
		t.Fatal("expected bundled script path")
	}
}
