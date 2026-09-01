package weather

import (
	"os"
	"strings"
	"testing"
)

func TestChineseAPIReferenceCoversExportedContract(t *testing.T) {
	content, err := os.ReadFile("API.md")
	if err != nil {
		t.Fatal(err)
	}
	reference := string(content)
	for _, required := range []string{
		"Experimental extension", "weather_lookup", "Options", "Request", "Result", "Current", "Today",
		"Definition", "Register", "NewHandler", "Run", "Lookup", "httprequest.Preparer", "Host", "取消",
	} {
		if !strings.Contains(reference, required) {
			t.Errorf("API.md missing %q", required)
		}
	}
}
