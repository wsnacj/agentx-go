package requestjson_test

import (
	"strings"
	"testing"

	"github.com/wsnacj/agentx-go/runtime/hosthttp/requestjson"
)

func TestExternalConsumerDecodesStrictJSON(t *testing.T) {
	var request struct {
		Prompt string `json:"prompt"`
	}
	if err := requestjson.Decode(strings.NewReader(`{"prompt":"inspect"}`), 64, &request); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if request.Prompt != "inspect" {
		t.Fatalf("prompt = %q", request.Prompt)
	}
}
