package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wsnacj/agentx-go/scenes/docparse"
	"github.com/wsnacj/agentx-go/scenes/docparse/hostkit"
)

func result() (string, error) {
	kit := hostkit.New(hostkit.Config{Source: "fixed-external-consumer"})
	out, err := kit.ExtractFields(context.Background(), map[string]any{
		"parse_result": map[string]any{
			"status": "success",
			"fields": []any{map[string]any{
				"key": "amount", "value": "10", "page_refs": []any{1},
			}},
		},
		"requested_fields": []string{"amount"},
	})
	if err != nil {
		return "", err
	}
	blob, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	var payload map[string]any
	if err := json.Unmarshal(blob, &payload); err != nil {
		return "", err
	}
	return fmt.Sprintf("agentx-docparse-ok:%s:%d:%s:%t", docparse.PackID, len(docparse.ToolNames()), payload["status"], payload["evidence_complete"]), nil
}

func main() {
	out, err := result()
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
