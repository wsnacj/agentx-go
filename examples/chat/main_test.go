package main

import (
	"context"
	"testing"
)

func TestChatExample(t *testing.T) {
	result, err := run(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "completed" || result.Reply != "fixture: hello" || result.SessionID != "chat-example" {
		t.Fatalf("result=%#v", result)
	}
}
