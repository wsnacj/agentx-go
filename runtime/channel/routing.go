package channel

import "context"

type BindingMatch struct {
	AccountID string
	ChatType  string
	ChatID    string
	UserID    string
}

func (b BindingMatch) Matches(message Message) bool {
	if b.AccountID != "" && b.AccountID != message.AccountID {
		return false
	}
	if b.ChatType != "" && b.ChatType != message.ChatType {
		return false
	}
	if b.ChatID != "" && b.ChatID != message.ChatID {
		return false
	}
	if b.UserID != "" && b.UserID != message.UserID {
		return false
	}
	return true
}

type RunnerBinding struct {
	Match  BindingMatch
	Runner TurnRunner
}

type RoutedRunner struct {
	DefaultRunner TurnRunner
	Bindings      []RunnerBinding
}

func (r RoutedRunner) RunTurn(ctx context.Context, inbound Message) (string, error) {
	runner := r.Resolve(inbound)
	return runner.RunTurn(ctx, inbound)
}

func (r RoutedRunner) WorkspaceDir() string {
	if r.DefaultRunner == nil {
		return ""
	}
	return r.DefaultRunner.WorkspaceDir()
}

func (r RoutedRunner) Profile() string {
	if r.DefaultRunner == nil {
		return ""
	}
	return r.DefaultRunner.Profile()
}

func (r RoutedRunner) Resolve(inbound Message) TurnRunner {
	for _, binding := range r.Bindings {
		if binding.Match.Matches(inbound) && binding.Runner != nil {
			return binding.Runner
		}
	}
	return r.DefaultRunner
}
