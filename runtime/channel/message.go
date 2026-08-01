package channel

import (
	"context"
	"encoding/json"
)

type Message struct {
	Platform  string
	AccountID string
	SessionID string
	MessageID string
	ChatID    string
	ThreadID  string
	ChatType  string
	UserID    string
	Text      string
	Mentioned bool
	Raw       json.RawMessage
}

type TextTarget struct {
	AccountID     string
	ChatID        string
	ThreadID      string
	MessageID     string
	ReplyInThread bool
}

type ToolContext struct {
	Platform string
	Target   TextTarget
	Sender   TextSender
}

type TextSender interface {
	SendText(ctx context.Context, target TextTarget, text string) error
}

type ReplySender interface {
	TextSender
	ReplyText(ctx context.Context, target TextTarget, text string) error
}

type EditSender interface {
	TextSender
	EditText(ctx context.Context, target TextTarget, text string) error
}

type DeleteSender interface {
	TextSender
	DeleteMessage(ctx context.Context, target TextTarget) error
}

type ReactSender interface {
	TextSender
	ReactMessage(ctx context.Context, target TextTarget, emoji string, remove bool, reactionID string) error
}

type ForwardSender interface {
	TextSender
	ForwardMessage(ctx context.Context, target TextTarget, sourceMessageID string) error
}

type TurnRunner interface {
	RunTurn(ctx context.Context, inbound Message) (string, error)
	WorkspaceDir() string
	Profile() string
}
