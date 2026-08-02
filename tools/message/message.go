// Package message provides the portable AgentX channel message tool.
package message

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	channelruntime "github.com/wsnacj/agentx-go/runtime/channel"
	agentxsafeerror "github.com/wsnacj/agentx-go/runtime/telemetry/safeerror"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
)

// Name is the catalog name of the message tool.
const Name = "message"

const messageToolName = Name

// Options supplies the Host-owned sender and current target.
type Options struct {
	Sender   channelruntime.TextSender
	Target   channelruntime.TextTarget
	Platform string
}

type messageToolArgs struct {
	Action        string              `json:"action"`
	Text          string              `json:"text"`
	AccountID     string              `json:"account_id"`
	ChatID        string              `json:"chat_id"`
	ThreadID      string              `json:"thread_id"`
	MessageID     string              `json:"message_id"`
	SourceMessage string              `json:"source_message_id"`
	ReplyInThread *bool               `json:"reply_in_thread"`
	Emoji         string              `json:"emoji"`
	ReactionID    string              `json:"reaction_id"`
	Remove        bool                `json:"remove"`
	ChatIDs       []string            `json:"chat_ids"`
	Targets       []messageToolTarget `json:"targets"`
}

type messageToolTarget struct {
	AccountID     string `json:"account_id"`
	ChatID        string `json:"chat_id"`
	ThreadID      string `json:"thread_id"`
	MessageID     string `json:"message_id"`
	ReplyInThread *bool  `json:"reply_in_thread"`
}

type messageToolFailure struct {
	Target channelruntime.TextTarget `json:"target"`
	Error  string                    `json:"error"`
}

type messageToolResult struct {
	Action     string                      `json:"action"`
	Platform   string                      `json:"platform,omitempty"`
	Available  bool                        `json:"available"`
	Status     string                      `json:"status,omitempty"`
	SourceID   string                      `json:"source_message_id,omitempty"`
	Target     channelruntime.TextTarget   `json:"target,omitempty"`
	Targets    []channelruntime.TextTarget `json:"targets,omitempty"`
	Failures   []messageToolFailure        `json:"failures,omitempty"`
	TextLen    int                         `json:"text_len,omitempty"`
	SentCount  int                         `json:"sent_count,omitempty"`
	Emoji      string                      `json:"emoji,omitempty"`
	ReactionID string                      `json:"reaction_id,omitempty"`
	Removed    bool                        `json:"removed,omitempty"`
	Actions    []string                    `json:"actions,omitempty"`
	Warning    string                      `json:"warning,omitempty"`
}

// Register adds the message tool when Options identify a usable Host surface.
func Register(reg toolcontract.Registrar, opts Options) {
	if reg == nil {
		return
	}
	handler := NewHandler(opts)
	if handler == nil {
		return
	}
	reg.Register(Definition(), handler)
}

// NewHandler constructs the real portable message coordination implementation.
// It returns nil only when neither a current target, platform nor sender exists.
func NewHandler(opts Options) toolcontract.Handler {
	target := normalizeMessageToolTarget(opts.Target)
	platform := strings.TrimSpace(opts.Platform)
	if target.ChatID == "" && target.ThreadID == "" && target.MessageID == "" && target.AccountID == "" && platform == "" && opts.Sender == nil {
		return nil
	}
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		args, err := decodeMessageToolArgs(call.Arguments)
		if err != nil {
			return "", err
		}
		action := normalizeMessageToolAction(args.Action, args.Text)
		resolvedTarget := target
		if value := strings.TrimSpace(args.AccountID); value != "" {
			resolvedTarget.AccountID = value
		}
		if value := strings.TrimSpace(args.ChatID); value != "" {
			resolvedTarget.ChatID = value
		}
		if value := strings.TrimSpace(args.ThreadID); value != "" {
			resolvedTarget.ThreadID = value
		}
		if value := strings.TrimSpace(args.MessageID); value != "" {
			resolvedTarget.MessageID = value
		}
		if args.ReplyInThread != nil {
			resolvedTarget.ReplyInThread = *args.ReplyInThread
		}
		switch action {
		case "current_target":
			return marshalMessageToolResult(messageToolResult{
				Action:    action,
				Platform:  platform,
				Available: messageToolHasDestination(resolvedTarget) && opts.Sender != nil,
				Status:    messageToolStatus(opts.Sender, resolvedTarget),
				Target:    resolvedTarget,
				Actions:   messageToolActions(),
				Warning:   messageToolWarning(opts.Sender, resolvedTarget),
			})
		case "send", "reply":
			text := strings.TrimSpace(args.Text)
			if action == "reply" && strings.TrimSpace(resolvedTarget.MessageID) == "" {
				resolvedTarget.MessageID = target.MessageID
			}
			return executeMessageSend(ctx, opts.Sender, platform, action, resolvedTarget, text)
		case "edit":
			text := strings.TrimSpace(args.Text)
			if strings.TrimSpace(resolvedTarget.MessageID) == "" {
				resolvedTarget.MessageID = target.MessageID
			}
			return executeMessageEdit(ctx, opts.Sender, platform, resolvedTarget, text)
		case "delete":
			if strings.TrimSpace(resolvedTarget.MessageID) == "" {
				resolvedTarget.MessageID = target.MessageID
			}
			return executeMessageDelete(ctx, opts.Sender, platform, resolvedTarget)
		case "react":
			if strings.TrimSpace(resolvedTarget.MessageID) == "" {
				resolvedTarget.MessageID = target.MessageID
			}
			return executeMessageReact(ctx, opts.Sender, platform, resolvedTarget, args)
		case "forward":
			return executeMessageForward(ctx, opts.Sender, platform, target, resolvedTarget, args)
		case "broadcast":
			return executeMessageBroadcast(ctx, opts.Sender, platform, target, args)
		default:
			return "", fmt.Errorf("%s: unsupported action %q", messageToolName, action)
		}
	}
}

func normalizeMessageToolTarget(target channelruntime.TextTarget) channelruntime.TextTarget {
	return channelruntime.TextTarget{
		AccountID:     strings.TrimSpace(target.AccountID),
		ChatID:        strings.TrimSpace(target.ChatID),
		ThreadID:      strings.TrimSpace(target.ThreadID),
		MessageID:     strings.TrimSpace(target.MessageID),
		ReplyInThread: target.ReplyInThread,
	}
}

func decodeMessageToolArgs(raw string) (messageToolArgs, error) {
	if strings.TrimSpace(raw) == "" {
		return messageToolArgs{}, nil
	}
	var args messageToolArgs
	if err := json.Unmarshal([]byte(raw), &args); err != nil {
		return messageToolArgs{}, fmt.Errorf("%s: invalid args: %w", messageToolName, err)
	}
	return args, nil
}

func normalizeMessageToolAction(raw string, text string) string {
	action := strings.ToLower(strings.TrimSpace(raw))
	if action != "" {
		return action
	}
	if strings.TrimSpace(text) != "" {
		return "send"
	}
	return "current_target"
}

func messageToolStatus(sender channelruntime.TextSender, target channelruntime.TextTarget) string {
	if sender == nil {
		return "unavailable"
	}
	if !messageToolHasDestination(target) {
		return "degraded"
	}
	return "available"
}

func messageToolWarning(sender channelruntime.TextSender, target channelruntime.TextTarget) string {
	if sender == nil {
		return "channel sender is not configured"
	}
	if !messageToolHasDestination(target) {
		return "current channel target is incomplete"
	}
	return ""
}

func messageToolActions() []string {
	return []string{"current_target", "send", "reply", "broadcast", "edit", "delete", "react", "forward"}
}

func executeMessageSend(
	ctx context.Context,
	sender channelruntime.TextSender,
	platform string,
	action string,
	target channelruntime.TextTarget,
	text string,
) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", missingMessageToolArgument([]string{"text"}, fmt.Sprintf("%s: text is required for action=%s", messageToolName, action))
	}
	if sender == nil {
		return "", fmt.Errorf("%s: sender is unavailable", messageToolName)
	}
	if !messageToolHasDestination(target) {
		return "", missingMessageToolArgument([]string{"chat_id", "thread_id"}, fmt.Sprintf("%s: chat_id or thread_id is required for action=%s", messageToolName, action))
	}
	if action == "reply" {
		if replySender, ok := sender.(channelruntime.ReplySender); ok {
			if err := replySender.ReplyText(ctx, target, text); err != nil {
				return "", fmt.Errorf("%s: %s failed: %w", messageToolName, action, err)
			}
		} else {
			if err := sender.SendText(ctx, target, text); err != nil {
				return "", fmt.Errorf("%s: %s failed: %w", messageToolName, action, err)
			}
		}
	} else {
		if err := sender.SendText(ctx, target, text); err != nil {
			return "", fmt.Errorf("%s: %s failed: %w", messageToolName, action, err)
		}
	}
	return marshalMessageToolResult(messageToolResult{
		Action:    action,
		Platform:  platform,
		Available: true,
		Status:    "sent",
		Target:    target,
		TextLen:   len([]rune(text)),
		SentCount: 1,
	})
}

func executeMessageBroadcast(
	ctx context.Context,
	sender channelruntime.TextSender,
	platform string,
	defaultTarget channelruntime.TextTarget,
	args messageToolArgs,
) (string, error) {
	text := strings.TrimSpace(args.Text)
	if text == "" {
		return "", missingMessageToolArgument([]string{"text"}, fmt.Sprintf("%s: text is required for action=broadcast", messageToolName))
	}
	if sender == nil {
		return "", fmt.Errorf("%s: sender is unavailable", messageToolName)
	}
	targets, err := resolveMessageBroadcastTargets(defaultTarget, args)
	if err != nil {
		return "", err
	}
	successes := make([]channelruntime.TextTarget, 0, len(targets))
	failures := make([]messageToolFailure, 0)
	for _, target := range targets {
		if strings.TrimSpace(target.ChatID) == "" {
			failures = append(failures, messageToolFailure{
				Target: target,
				Error:  "chat_id is required",
			})
			continue
		}
		if err := sender.SendText(ctx, target, text); err != nil {
			failures = append(failures, messageToolFailure{
				Target: target,
				Error:  modelFacingErrorSummary(err, "message_delivery", "broadcast_send_failed"),
			})
			continue
		}
		successes = append(successes, target)
	}
	status := "sent"
	switch {
	case len(successes) == 0 && len(failures) > 0:
		status = "failed"
	case len(successes) > 0 && len(failures) > 0:
		status = "partial"
	}
	warning := ""
	if len(failures) > 0 {
		warning = fmt.Sprintf("%d target(s) failed during broadcast", len(failures))
	}
	return marshalMessageToolResult(messageToolResult{
		Action:    "broadcast",
		Platform:  platform,
		Available: len(successes) > 0,
		Status:    status,
		Targets:   successes,
		Failures:  failures,
		TextLen:   len([]rune(text)),
		SentCount: len(successes),
		Warning:   warning,
	})
}

func executeMessageEdit(
	ctx context.Context,
	sender channelruntime.TextSender,
	platform string,
	target channelruntime.TextTarget,
	text string,
) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", missingMessageToolArgument([]string{"text"}, fmt.Sprintf("%s: text is required for action=edit", messageToolName))
	}
	if sender == nil {
		return "", fmt.Errorf("%s: sender is unavailable", messageToolName)
	}
	if strings.TrimSpace(target.MessageID) == "" {
		return "", missingMessageToolArgument([]string{"message_id"}, fmt.Sprintf("%s: message_id is required for action=edit", messageToolName))
	}
	editSender, ok := sender.(channelruntime.EditSender)
	if !ok {
		return "", fmt.Errorf("%s: edit is unsupported for the current channel sender", messageToolName)
	}
	if err := editSender.EditText(ctx, target, text); err != nil {
		return "", fmt.Errorf("%s: edit failed: %w", messageToolName, err)
	}
	return marshalMessageToolResult(messageToolResult{
		Action:    "edit",
		Platform:  platform,
		Available: true,
		Status:    "updated",
		Target:    target,
		TextLen:   len([]rune(text)),
	})
}

func executeMessageDelete(
	ctx context.Context,
	sender channelruntime.TextSender,
	platform string,
	target channelruntime.TextTarget,
) (string, error) {
	if sender == nil {
		return "", fmt.Errorf("%s: sender is unavailable", messageToolName)
	}
	if strings.TrimSpace(target.MessageID) == "" {
		return "", missingMessageToolArgument([]string{"message_id"}, fmt.Sprintf("%s: message_id is required for action=delete", messageToolName))
	}
	deleteSender, ok := sender.(channelruntime.DeleteSender)
	if !ok {
		return "", fmt.Errorf("%s: delete is unsupported for the current channel sender", messageToolName)
	}
	if err := deleteSender.DeleteMessage(ctx, target); err != nil {
		return "", fmt.Errorf("%s: delete failed: %w", messageToolName, err)
	}
	return marshalMessageToolResult(messageToolResult{
		Action:    "delete",
		Platform:  platform,
		Available: true,
		Status:    "deleted",
		Target:    target,
	})
}

func executeMessageReact(
	ctx context.Context,
	sender channelruntime.TextSender,
	platform string,
	target channelruntime.TextTarget,
	args messageToolArgs,
) (string, error) {
	if sender == nil {
		return "", fmt.Errorf("%s: sender is unavailable", messageToolName)
	}
	if strings.TrimSpace(target.MessageID) == "" {
		return "", missingMessageToolArgument([]string{"message_id"}, fmt.Sprintf("%s: message_id is required for action=react", messageToolName))
	}
	emoji := strings.ToUpper(strings.TrimSpace(args.Emoji))
	reactionID := strings.TrimSpace(args.ReactionID)
	if emoji == "" && !(args.Remove && reactionID != "") {
		return "", missingMessageToolArgument([]string{"emoji"}, fmt.Sprintf("%s: emoji is required for action=react unless reaction_id is provided for removal", messageToolName))
	}
	reactSender, ok := sender.(channelruntime.ReactSender)
	if !ok {
		return "", fmt.Errorf("%s: react is unsupported for the current channel sender", messageToolName)
	}
	if err := reactSender.ReactMessage(ctx, target, emoji, args.Remove, reactionID); err != nil {
		return "", fmt.Errorf("%s: react failed: %w", messageToolName, err)
	}
	status := "reacted"
	if args.Remove {
		status = "reaction_removed"
	}
	return marshalMessageToolResult(messageToolResult{
		Action:     "react",
		Platform:   platform,
		Available:  true,
		Status:     status,
		Target:     target,
		Emoji:      emoji,
		ReactionID: reactionID,
		Removed:    args.Remove,
	})
}

func executeMessageForward(
	ctx context.Context,
	sender channelruntime.TextSender,
	platform string,
	defaultTarget channelruntime.TextTarget,
	resolvedTarget channelruntime.TextTarget,
	args messageToolArgs,
) (string, error) {
	if sender == nil {
		return "", fmt.Errorf("%s: sender is unavailable", messageToolName)
	}
	forwardSender, ok := sender.(channelruntime.ForwardSender)
	if !ok {
		return "", fmt.Errorf("%s: forward is unsupported for the current channel sender", messageToolName)
	}
	sourceMessageID := firstNonEmptyMessageValue(strings.TrimSpace(args.SourceMessage), strings.TrimSpace(args.MessageID), strings.TrimSpace(defaultTarget.MessageID))
	if sourceMessageID == "" {
		return "", missingMessageToolArgument([]string{"source_message_id"}, fmt.Sprintf("%s: source_message_id is required for action=forward", messageToolName))
	}
	var targets []channelruntime.TextTarget
	switch {
	case len(args.ChatIDs) > 0 || len(args.Targets) > 0:
		resolved, err := resolveMessageTargetList(defaultTarget, args)
		if err != nil {
			return "", err
		}
		targets = resolved
	default:
		target := normalizeMessageToolTarget(resolvedTarget)
		target.MessageID = ""
		target.ReplyInThread = false
		if !messageToolHasDestination(target) {
			return "", missingMessageToolArgument([]string{"chat_id", "thread_id", "chat_ids", "targets"}, fmt.Sprintf("%s: chat_id, thread_id, chat_ids, or targets are required for action=forward", messageToolName))
		}
		targets = []channelruntime.TextTarget{target}
	}
	successes := make([]channelruntime.TextTarget, 0, len(targets))
	failures := make([]messageToolFailure, 0)
	for _, target := range targets {
		if !messageToolHasDestination(target) {
			failures = append(failures, messageToolFailure{
				Target: target,
				Error:  "chat_id or thread_id is required",
			})
			continue
		}
		if err := forwardSender.ForwardMessage(ctx, target, sourceMessageID); err != nil {
			failures = append(failures, messageToolFailure{
				Target: target,
				Error:  modelFacingErrorSummary(err, "message_delivery", "forward_failed"),
			})
			continue
		}
		successes = append(successes, target)
	}
	status := "sent"
	switch {
	case len(successes) == 0 && len(failures) > 0:
		status = "failed"
	case len(successes) > 0 && len(failures) > 0:
		status = "partial"
	}
	warning := ""
	if len(failures) > 0 {
		warning = fmt.Sprintf("%d target(s) failed during forward", len(failures))
	}
	return marshalMessageToolResult(messageToolResult{
		Action:    "forward",
		Platform:  platform,
		Available: len(successes) > 0,
		Status:    status,
		SourceID:  sourceMessageID,
		Targets:   successes,
		Failures:  failures,
		SentCount: len(successes),
		Warning:   warning,
	})
}

func resolveMessageBroadcastTargets(defaultTarget channelruntime.TextTarget, args messageToolArgs) ([]channelruntime.TextTarget, error) {
	return resolveMessageTargetList(defaultTarget, args)
}

func resolveMessageTargetList(defaultTarget channelruntime.TextTarget, args messageToolArgs) ([]channelruntime.TextTarget, error) {
	targets := make([]channelruntime.TextTarget, 0, len(args.ChatIDs)+len(args.Targets))
	for _, chatID := range args.ChatIDs {
		target := defaultTarget
		target.ChatID = strings.TrimSpace(chatID)
		target.ThreadID = ""
		target.MessageID = ""
		target.ReplyInThread = false
		targets = append(targets, normalizeMessageToolTarget(target))
	}
	for _, item := range args.Targets {
		target := channelruntime.TextTarget{
			AccountID: firstNonEmptyMessageValue(strings.TrimSpace(item.AccountID), strings.TrimSpace(defaultTarget.AccountID)),
			ChatID:    strings.TrimSpace(item.ChatID),
			ThreadID:  strings.TrimSpace(item.ThreadID),
			MessageID: strings.TrimSpace(item.MessageID),
		}
		if item.ReplyInThread != nil {
			target.ReplyInThread = *item.ReplyInThread
		}
		targets = append(targets, normalizeMessageToolTarget(target))
	}
	targets = dedupeMessageTargets(targets)
	if len(targets) == 0 {
		return nil, missingMessageToolArgument([]string{"chat_ids", "targets"}, fmt.Sprintf("%s: chat_ids or targets are required", messageToolName))
	}
	return targets, nil
}

func missingMessageToolArgument(fields []string, detail string) error {
	return agentxtoolerrors.NewMissingRequiredToolArgumentError(messageToolName, fields, detail)
}

func dedupeMessageTargets(targets []channelruntime.TextTarget) []channelruntime.TextTarget {
	if len(targets) < 2 {
		return targets
	}
	seen := make(map[string]struct{}, len(targets))
	out := make([]channelruntime.TextTarget, 0, len(targets))
	for _, target := range targets {
		key := strings.Join([]string{
			strings.TrimSpace(target.AccountID),
			strings.TrimSpace(target.ChatID),
			strings.TrimSpace(target.ThreadID),
			strings.TrimSpace(target.MessageID),
			fmt.Sprintf("%t", target.ReplyInThread),
		}, "|")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, target)
	}
	return out
}

func firstNonEmptyMessageValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func marshalMessageToolResult(result messageToolResult) (string, error) {
	blob, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(blob), nil
}

// Definition returns the stable model-facing schema.
func Definition() toolcontract.Definition {
	return toolcontract.Definition{
		Type: "function",
		Function: toolcontract.Function{
			Name:        messageToolName,
			Description: "Inspect the active channel target and send, reply, broadcast, forward, edit, delete, or react to channel messages through the current channel adapter.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action":            stringEnumSchema("Message operation to perform. Defaults to send when text is provided, otherwise current_target.", messageToolActions()...),
					"text":              stringSchema("Text content for send, reply, broadcast, or edit actions."),
					"account_id":        stringSchema("Channel account or workspace identifier for the target."),
					"chat_id":           stringSchema("Channel chat or room identifier for the target."),
					"thread_id":         stringSchema("Thread identifier for threaded channel targets."),
					"message_id":        stringSchema("Message identifier for reply, edit, delete, react, or forward source selection."),
					"source_message_id": stringSchema("Source message identifier for forward actions."),
					"reply_in_thread":   boolSchema("Whether replies should be posted inside the selected thread when supported."),
					"emoji":             stringSchema("Emoji or reaction token for react actions."),
					"reaction_id":       stringSchema("Existing reaction identifier to remove when remove=true."),
					"remove":            boolSchema("Remove the selected reaction instead of adding one."),
					"chat_ids": map[string]any{
						"type":        "array",
						"description": "Chat identifiers for broadcast or forward actions.",
						"items":       map[string]any{"type": "string"},
					},
					"targets": map[string]any{
						"type":        "array",
						"description": "Explicit channel targets for broadcast or forward actions.",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"account_id":      stringSchema("Channel account or workspace identifier for this target."),
								"chat_id":         stringSchema("Channel chat or room identifier for this target."),
								"thread_id":       stringSchema("Thread identifier for this target."),
								"message_id":      stringSchema("Message identifier for this target when an action needs a message context."),
								"reply_in_thread": boolSchema("Whether this target should reply inside its selected thread when supported."),
							},
						},
					},
				},
			},
		},
	}
}

func modelFacingErrorSummary(err error, class string, code string) string {
	return agentxsafeerror.Summary(agentxsafeerror.Project(err, class, code))
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func boolSchema(description string) map[string]any {
	return map[string]any{"type": "boolean", "description": description}
}

func stringEnumSchema(description string, values ...string) map[string]any {
	return map[string]any{"type": "string", "description": description, "enum": append([]string(nil), values...)}
}

func messageToolHasDestination(target channelruntime.TextTarget) bool {
	return strings.TrimSpace(target.ChatID) != "" || strings.TrimSpace(target.ThreadID) != ""
}
