package diffs_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/diffs"
)

func TestDiffsCatalogExecution(t *testing.T) {
	registry := tools.NewRegistry()
	diffs.Register(registry)
	result, err := registry.Execute(context.Background(), toolcontract.Call{Name: "diffs", Arguments: `{"before":"alpha\nbeta\n","after":"alpha\ngamma\n","path":"sample.txt"}`})
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["tool"] != "diffs" || payload["format"] != "unified" || !strings.Contains(payload["unified"].(string), "-beta") || !strings.Contains(payload["unified"].(string), "+gamma") {
		t.Fatalf("payload=%#v", payload)
	}
}

func TestDefinitionDocumentsTextOnlyBoundary(t *testing.T) {
	description := diffs.Definition().Function.Description
	for _, token := range []string{"does not inspect files", "write or mutate files", "write/edit/apply_patch", "git status"} {
		if !strings.Contains(description, token) {
			t.Fatalf("description missing %q: %s", token, description)
		}
	}
}
