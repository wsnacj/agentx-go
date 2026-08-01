package channel

import (
	"context"
	"strings"
	"testing"
)

func TestSplitTextSplitsParagraphsAndLongSegments(t *testing.T) {
	got := SplitText("第一段内容\n第二段更长一些", 8)
	if len(got) < 2 {
		t.Fatalf("expected multiple chunks, got %#v", got)
	}
	if strings.Join(got, "") == "" {
		t.Fatalf("expected non-empty chunked output")
	}
}

func TestBindingMatchMatchesScopedMessage(t *testing.T) {
	match := BindingMatch{AccountID: "main", ChatType: "group", ChatID: "oc_group"}
	if !match.Matches(Message{AccountID: "main", ChatType: "group", ChatID: "oc_group"}) {
		t.Fatalf("expected binding to match")
	}
	if match.Matches(Message{AccountID: "main", ChatType: "group", ChatID: "oc_other"}) {
		t.Fatalf("expected binding mismatch")
	}
}

func TestRoutedRunnerUsesMatchingBinding(t *testing.T) {
	defaultRunner := stubRunner{reply: "default"}
	groupRunner := stubRunner{reply: "group"}
	router := RoutedRunner{
		DefaultRunner: defaultRunner,
		Bindings: []RunnerBinding{
			{
				Match:  BindingMatch{AccountID: "main", ChatType: "group", ChatID: "oc_group"},
				Runner: groupRunner,
			},
		},
	}
	reply, err := router.RunTurn(context.Background(), Message{AccountID: "main", ChatType: "group", ChatID: "oc_group"})
	if err != nil {
		t.Fatalf("RunTurn: %v", err)
	}
	if reply != "group" {
		t.Fatalf("expected group binding runner, got %q", reply)
	}
}

func TestAccountSendersUsesDefaultAndFallback(t *testing.T) {
	sender := &recordingSender{}
	router := AccountSenders{
		DefaultAccountID: "main",
		Senders: map[string]TextSender{
			"main": sender,
		},
	}
	if err := router.SendText(context.Background(), TextTarget{ChatID: "oc_dm"}, "hello"); err != nil {
		t.Fatalf("SendText: %v", err)
	}
	if len(sender.texts) != 1 || sender.texts[0] != "hello" {
		t.Fatalf("unexpected sender texts: %#v", sender.texts)
	}
}

func TestAccountSendersReplyUsesReplySenderWhenAvailable(t *testing.T) {
	sender := &recordingSender{}
	router := AccountSenders{
		DefaultAccountID: "main",
		Senders: map[string]TextSender{
			"main": sender,
		},
	}
	if err := router.ReplyText(context.Background(), TextTarget{ChatID: "oc_dm", MessageID: "msg_1", ReplyInThread: true}, "hello"); err != nil {
		t.Fatalf("ReplyText: %v", err)
	}
	if len(sender.replyTexts) != 1 || sender.replyTexts[0] != "hello" {
		t.Fatalf("unexpected reply texts: %#v", sender.replyTexts)
	}
	if len(sender.replyTargets) != 1 || !sender.replyTargets[0].ReplyInThread {
		t.Fatalf("unexpected reply targets: %#v", sender.replyTargets)
	}
}

func TestAccountSendersEditAndDeleteUseOptionalSenderInterfaces(t *testing.T) {
	sender := &recordingSender{}
	router := AccountSenders{
		DefaultAccountID: "main",
		Senders: map[string]TextSender{
			"main": sender,
		},
	}
	target := TextTarget{ChatID: "oc_dm", MessageID: "msg_1"}
	if err := router.EditText(context.Background(), target, "updated"); err != nil {
		t.Fatalf("EditText: %v", err)
	}
	if err := router.DeleteMessage(context.Background(), target); err != nil {
		t.Fatalf("DeleteMessage: %v", err)
	}
	if len(sender.editTexts) != 1 || sender.editTexts[0] != "updated" {
		t.Fatalf("unexpected edit texts: %#v", sender.editTexts)
	}
	if len(sender.deleteTargets) != 1 || sender.deleteTargets[0].MessageID != "msg_1" {
		t.Fatalf("unexpected delete targets: %#v", sender.deleteTargets)
	}
}

func TestAccountSendersReactUsesOptionalSenderInterface(t *testing.T) {
	sender := &recordingSender{}
	router := AccountSenders{
		DefaultAccountID: "main",
		Senders: map[string]TextSender{
			"main": sender,
		},
	}
	target := TextTarget{ChatID: "oc_dm", MessageID: "msg_1"}
	if err := router.ReactMessage(context.Background(), target, "SMILE", false, ""); err != nil {
		t.Fatalf("ReactMessage: %v", err)
	}
	if len(sender.reactTargets) != 1 || sender.reactTargets[0].MessageID != "msg_1" || len(sender.reactEmoji) != 1 || sender.reactEmoji[0] != "SMILE" {
		t.Fatalf("unexpected react capture: targets=%#v emoji=%#v", sender.reactTargets, sender.reactEmoji)
	}
}

func TestAccountSendersForwardUsesOptionalSenderInterface(t *testing.T) {
	sender := &recordingSender{}
	router := AccountSenders{
		DefaultAccountID: "main",
		Senders: map[string]TextSender{
			"main": sender,
		},
	}
	target := TextTarget{ThreadID: "omt_thread"}
	if err := router.ForwardMessage(context.Background(), target, "msg_1"); err != nil {
		t.Fatalf("ForwardMessage: %v", err)
	}
	if len(sender.forwardTargets) != 1 || sender.forwardTargets[0].ThreadID != "omt_thread" || len(sender.forwardSourceIDs) != 1 || sender.forwardSourceIDs[0] != "msg_1" {
		t.Fatalf("unexpected forward capture: targets=%#v source=%#v", sender.forwardTargets, sender.forwardSourceIDs)
	}
}

type stubRunner struct {
	reply string
}

func (s stubRunner) RunTurn(_ context.Context, _ Message) (string, error) { return s.reply, nil }
func (s stubRunner) WorkspaceDir() string                                 { return "." }
func (s stubRunner) Profile() string                                      { return "safe" }

type recordingSender struct {
	texts            []string
	replyTexts       []string
	replyTargets     []TextTarget
	editTexts        []string
	editTargets      []TextTarget
	deleteTargets    []TextTarget
	reactTargets     []TextTarget
	reactEmoji       []string
	reactRemove      []bool
	reactionIDs      []string
	forwardTargets   []TextTarget
	forwardSourceIDs []string
}

func (s *recordingSender) SendText(_ context.Context, _ TextTarget, text string) error {
	s.texts = append(s.texts, text)
	return nil
}

func (s *recordingSender) ReplyText(_ context.Context, target TextTarget, text string) error {
	s.replyTargets = append(s.replyTargets, target)
	s.replyTexts = append(s.replyTexts, text)
	return nil
}

func (s *recordingSender) EditText(_ context.Context, target TextTarget, text string) error {
	s.editTargets = append(s.editTargets, target)
	s.editTexts = append(s.editTexts, text)
	return nil
}

func (s *recordingSender) DeleteMessage(_ context.Context, target TextTarget) error {
	s.deleteTargets = append(s.deleteTargets, target)
	return nil
}

func (s *recordingSender) ReactMessage(_ context.Context, target TextTarget, emoji string, remove bool, reactionID string) error {
	s.reactTargets = append(s.reactTargets, target)
	s.reactEmoji = append(s.reactEmoji, emoji)
	s.reactRemove = append(s.reactRemove, remove)
	s.reactionIDs = append(s.reactionIDs, reactionID)
	return nil
}

func (s *recordingSender) ForwardMessage(_ context.Context, target TextTarget, sourceMessageID string) error {
	s.forwardTargets = append(s.forwardTargets, target)
	s.forwardSourceIDs = append(s.forwardSourceIDs, sourceMessageID)
	return nil
}
