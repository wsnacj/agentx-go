package channel

import (
	"context"
	"testing"
)

func TestBindingMatchTreatsEmptyFieldsAsWildcards(t *testing.T) {
	match := BindingMatch{AccountID: "main"}
	msg := Message{
		AccountID: "main",
		ChatType:  "group",
		ChatID:    "oc_group",
		UserID:    "ou_user",
	}
	if !match.Matches(msg) {
		t.Fatalf("expected partial binding to match message: %#v", msg)
	}
}

func TestBindingMatchRejectsUserMismatch(t *testing.T) {
	match := BindingMatch{UserID: "ou_expected"}
	if match.Matches(Message{UserID: "ou_other"}) {
		t.Fatalf("expected user-scoped binding mismatch")
	}
}

func TestRoutedRunnerResolveFallsBackWhenBindingRunnerNil(t *testing.T) {
	defaultRunner := stubRunner{reply: "default"}
	router := RoutedRunner{
		DefaultRunner: defaultRunner,
		Bindings: []RunnerBinding{
			{
				Match:  BindingMatch{AccountID: "main", ChatID: "oc_group"},
				Runner: nil,
			},
		},
	}

	got := router.Resolve(Message{AccountID: "main", ChatID: "oc_group"})
	if got != defaultRunner {
		t.Fatalf("expected default runner fallback when binding runner is nil")
	}
}

func TestRoutedRunnerWorkspaceDirAndProfileHandleNilDefault(t *testing.T) {
	var router RoutedRunner
	if got := router.WorkspaceDir(); got != "" {
		t.Fatalf("expected blank workspace dir for nil default runner, got %q", got)
	}
	if got := router.Profile(); got != "" {
		t.Fatalf("expected blank profile for nil default runner, got %q", got)
	}
}

func TestRoutedRunnerRunTurnUsesDefaultRunnerWhenNoBindingMatches(t *testing.T) {
	defaultRunner := stubRunner{reply: "default"}
	router := RoutedRunner{
		DefaultRunner: defaultRunner,
		Bindings: []RunnerBinding{
			{
				Match:  BindingMatch{AccountID: "main", ChatID: "oc_group"},
				Runner: stubRunner{reply: "group"},
			},
		},
	}

	reply, err := router.RunTurn(context.Background(), Message{AccountID: "main", ChatID: "oc_other"})
	if err != nil {
		t.Fatalf("RunTurn: %v", err)
	}
	if reply != "default" {
		t.Fatalf("expected default runner reply, got %q", reply)
	}
}
