package main

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	agentx "github.com/wsnacj/agentx-go"
)

func TestReferenceHostChat(t *testing.T) {
	value, err := parseConfig([]string{"-input", "hello"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	host, err := newReferenceHost(value)
	if err != nil {
		t.Fatal(err)
	}
	result, err := host.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "completed" || result.Reply != "fixture: hello" {
		t.Fatalf("unexpected result: %#v", result)
	}
	stored, err := host.store.GetRun(context.Background(), host.runID)
	if err != nil || stored.Status != "completed" {
		t.Fatalf("unexpected stored run: %#v err=%v", stored, err)
	}
	if err := host.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := host.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	_, err = host.client.Run(context.Background(), agentx.RunRequest{Input: "closed"})
	if !errors.Is(err, &agentx.Error{Code: agentx.CodeClientClosed}) {
		t.Fatalf("closed error = %v", err)
	}
}

func TestReferenceHostToolLoop(t *testing.T) {
	host, err := newReferenceHost(config{
		Mode: modeToolLoop, Provider: providerFixture, Tools: toolsDiffs, Store: storeMemory, Input: "compare",
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := host.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "completed" || !strings.Contains(result.Reply, "reference.txt") {
		t.Fatalf("unexpected result: %#v", result)
	}
	if err := host.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestReferenceHostPreservesCancellationContract(t *testing.T) {
	host, err := newReferenceHost(config{
		Mode: modeChat, Provider: providerFixture, Tools: toolsNone, Store: storeMemory, Input: "hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = host.Run(ctx)
	if !errors.Is(err, &agentx.Error{Code: agentx.CodeCanceled}) {
		t.Fatalf("canceled error = %v", err)
	}
	if err := host.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestReferenceHostRejectsImplicitCapability(t *testing.T) {
	tests := []config{
		{Mode: modeChat, Provider: "environment", Tools: toolsNone, Store: storeMemory, Input: "hello"},
		{Mode: modeChat, Provider: providerFixture, Tools: toolsDiffs, Store: storeMemory, Input: "hello"},
		{Mode: modeToolLoop, Provider: providerFixture, Tools: toolsNone, Store: storeMemory, Input: "hello"},
		{Mode: modeChat, Provider: providerFixture, Tools: toolsNone, Store: "file", Input: "hello"},
	}
	for _, test := range tests {
		if _, err := newReferenceHost(test); err == nil {
			t.Fatalf("expected rejection for %#v", test)
		}
	}
}
