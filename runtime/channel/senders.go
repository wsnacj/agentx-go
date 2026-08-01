package channel

import (
	"context"
	"fmt"
	"strings"
)

type AccountSenders struct {
	DefaultAccountID string
	Senders          map[string]TextSender
}

func (s AccountSenders) SendText(ctx context.Context, target TextTarget, text string) error {
	sender, accountID, err := s.resolveSender(target)
	if err != nil {
		return err
	}
	_ = accountID
	return sender.SendText(ctx, target, text)
}

func (s AccountSenders) ReplyText(ctx context.Context, target TextTarget, text string) error {
	sender, _, err := s.resolveSender(target)
	if err != nil {
		return err
	}
	replySender, ok := sender.(ReplySender)
	if !ok {
		return sender.SendText(ctx, target, text)
	}
	return replySender.ReplyText(ctx, target, text)
}

func (s AccountSenders) EditText(ctx context.Context, target TextTarget, text string) error {
	sender, _, err := s.resolveSender(target)
	if err != nil {
		return err
	}
	editSender, ok := sender.(EditSender)
	if !ok {
		return fmt.Errorf("channel edit is unsupported for account %q", strings.TrimSpace(target.AccountID))
	}
	return editSender.EditText(ctx, target, text)
}

func (s AccountSenders) DeleteMessage(ctx context.Context, target TextTarget) error {
	sender, _, err := s.resolveSender(target)
	if err != nil {
		return err
	}
	deleteSender, ok := sender.(DeleteSender)
	if !ok {
		return fmt.Errorf("channel delete is unsupported for account %q", strings.TrimSpace(target.AccountID))
	}
	return deleteSender.DeleteMessage(ctx, target)
}

func (s AccountSenders) ReactMessage(ctx context.Context, target TextTarget, emoji string, remove bool, reactionID string) error {
	sender, _, err := s.resolveSender(target)
	if err != nil {
		return err
	}
	reactSender, ok := sender.(ReactSender)
	if !ok {
		return fmt.Errorf("channel react is unsupported for account %q", strings.TrimSpace(target.AccountID))
	}
	return reactSender.ReactMessage(ctx, target, emoji, remove, reactionID)
}

func (s AccountSenders) ForwardMessage(ctx context.Context, target TextTarget, sourceMessageID string) error {
	sender, _, err := s.resolveSender(target)
	if err != nil {
		return err
	}
	forwardSender, ok := sender.(ForwardSender)
	if !ok {
		return fmt.Errorf("channel forward is unsupported for account %q", strings.TrimSpace(target.AccountID))
	}
	return forwardSender.ForwardMessage(ctx, target, sourceMessageID)
}

func (s AccountSenders) resolveSender(target TextTarget) (TextSender, string, error) {
	accountID := strings.TrimSpace(target.AccountID)
	if accountID == "" {
		accountID = strings.TrimSpace(s.DefaultAccountID)
	}
	sender := s.Senders[accountID]
	if sender == nil {
		sender = s.Senders[""]
	}
	if sender == nil {
		for _, candidate := range s.Senders {
			sender = candidate
			break
		}
	}
	if sender == nil {
		return nil, accountID, fmt.Errorf("no channel sender configured for account %q", accountID)
	}
	return sender, accountID, nil
}
