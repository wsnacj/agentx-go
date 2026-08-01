package channel

import (
	"context"
	"strings"
	"testing"
)

func TestChunkingSenderReplyFallsBackToSendTextWhenReplyUnsupported(t *testing.T) {
	sender := &textOnlySender{}
	chunking := ChunkingSender{Base: sender, Limit: 6}

	if err := chunking.ReplyText(context.Background(), TextTarget{ChatID: "oc_dm"}, "hello world"); err != nil {
		t.Fatalf("ReplyText: %v", err)
	}
	if len(sender.texts) < 2 {
		t.Fatalf("expected chunked send fallback, got %#v", sender.texts)
	}
}

func TestChunkingSenderEditTextTrimsAndRequiresCapability(t *testing.T) {
	sender := &recordingSender{}
	chunking := ChunkingSender{Base: sender}

	if err := chunking.EditText(context.Background(), TextTarget{MessageID: "msg_1"}, "  updated  "); err != nil {
		t.Fatalf("EditText: %v", err)
	}
	if len(sender.editTexts) != 1 || sender.editTexts[0] != "updated" {
		t.Fatalf("unexpected edit capture: %#v", sender.editTexts)
	}

	unsupported := ChunkingSender{Base: &textOnlySender{}}
	if err := unsupported.EditText(context.Background(), TextTarget{MessageID: "msg_1"}, "updated"); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported edit error, got %v", err)
	}
}

func TestChunkingSenderReactAndForwardTrimInputs(t *testing.T) {
	sender := &recordingSender{}
	chunking := ChunkingSender{Base: sender}

	if err := chunking.ReactMessage(context.Background(), TextTarget{MessageID: "msg_1"}, "  SMILE  ", false, "  reaction_1  "); err != nil {
		t.Fatalf("ReactMessage: %v", err)
	}
	if len(sender.reactEmoji) != 1 || sender.reactEmoji[0] != "SMILE" {
		t.Fatalf("unexpected react emoji: %#v", sender.reactEmoji)
	}
	if len(sender.reactionIDs) != 1 || sender.reactionIDs[0] != "reaction_1" {
		t.Fatalf("unexpected reaction ids: %#v", sender.reactionIDs)
	}

	if err := chunking.ForwardMessage(context.Background(), TextTarget{ChatID: "oc_dm"}, "  msg_1  "); err != nil {
		t.Fatalf("ForwardMessage: %v", err)
	}
	if len(sender.forwardSourceIDs) != 1 || sender.forwardSourceIDs[0] != "msg_1" {
		t.Fatalf("unexpected forward source ids: %#v", sender.forwardSourceIDs)
	}
}

func TestChunkingSenderDeleteMessageIsNoopWhenDeleteUnsupported(t *testing.T) {
	chunking := ChunkingSender{Base: &textOnlySender{}}
	if err := chunking.DeleteMessage(context.Background(), TextTarget{MessageID: "msg_1"}); err != nil {
		t.Fatalf("expected delete noop, got %v", err)
	}
}

func TestChunkingSenderNilBaseIsNoop(t *testing.T) {
	chunking := ChunkingSender{}
	if err := chunking.SendText(context.Background(), TextTarget{ChatID: "oc_dm"}, "hello"); err != nil {
		t.Fatalf("SendText: %v", err)
	}
	if err := chunking.ReplyText(context.Background(), TextTarget{ChatID: "oc_dm"}, "hello"); err != nil {
		t.Fatalf("ReplyText: %v", err)
	}
	if err := chunking.EditText(context.Background(), TextTarget{MessageID: "msg_1"}, "hello"); err != nil {
		t.Fatalf("EditText: %v", err)
	}
	if err := chunking.DeleteMessage(context.Background(), TextTarget{MessageID: "msg_1"}); err != nil {
		t.Fatalf("DeleteMessage: %v", err)
	}
	if err := chunking.ReactMessage(context.Background(), TextTarget{MessageID: "msg_1"}, "SMILE", false, ""); err != nil {
		t.Fatalf("ReactMessage: %v", err)
	}
	if err := chunking.ForwardMessage(context.Background(), TextTarget{ChatID: "oc_dm"}, "msg_1"); err != nil {
		t.Fatalf("ForwardMessage: %v", err)
	}
}

type textOnlySender struct {
	texts []string
}

func (s *textOnlySender) SendText(_ context.Context, _ TextTarget, text string) error {
	s.texts = append(s.texts, text)
	return nil
}
