package message

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	channelruntime "github.com/wsnacj/agentx-go/runtime/channel"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
	canonicaltools "github.com/wsnacj/agentx-go/tools"
)

type fakeMessageSender struct {
	targets          []channelruntime.TextTarget
	texts            []string
	replyTargets     []channelruntime.TextTarget
	replyTexts       []string
	editTargets      []channelruntime.TextTarget
	editTexts        []string
	deleteTargets    []channelruntime.TextTarget
	reactTargets     []channelruntime.TextTarget
	reactEmoji       []string
	reactRemove      []bool
	reactionIDs      []string
	forwardTargets   []channelruntime.TextTarget
	forwardSourceIDs []string
	err              error
	failFor          map[string]error
}

func (f *fakeMessageSender) SendText(_ context.Context, target channelruntime.TextTarget, text string) error {
	if f.failFor != nil {
		if err, ok := f.failFor[target.ChatID]; ok {
			return err
		}
	}
	if f.err != nil {
		return f.err
	}
	f.targets = append(f.targets, target)
	f.texts = append(f.texts, text)
	return nil
}

func (f *fakeMessageSender) ReplyText(_ context.Context, target channelruntime.TextTarget, text string) error {
	if f.failFor != nil {
		if err, ok := f.failFor[target.ChatID]; ok {
			return err
		}
	}
	if f.err != nil {
		return f.err
	}
	f.replyTargets = append(f.replyTargets, target)
	f.replyTexts = append(f.replyTexts, text)
	return nil
}

func (f *fakeMessageSender) EditText(_ context.Context, target channelruntime.TextTarget, text string) error {
	if f.failFor != nil {
		if err, ok := f.failFor[target.ChatID]; ok {
			return err
		}
	}
	if f.err != nil {
		return f.err
	}
	f.editTargets = append(f.editTargets, target)
	f.editTexts = append(f.editTexts, text)
	return nil
}

func (f *fakeMessageSender) DeleteMessage(_ context.Context, target channelruntime.TextTarget) error {
	if f.failFor != nil {
		if err, ok := f.failFor[target.ChatID]; ok {
			return err
		}
	}
	if f.err != nil {
		return f.err
	}
	f.deleteTargets = append(f.deleteTargets, target)
	return nil
}

func (f *fakeMessageSender) ReactMessage(_ context.Context, target channelruntime.TextTarget, emoji string, remove bool, reactionID string) error {
	if f.failFor != nil {
		if err, ok := f.failFor[target.ChatID]; ok {
			return err
		}
	}
	if f.err != nil {
		return f.err
	}
	f.reactTargets = append(f.reactTargets, target)
	f.reactEmoji = append(f.reactEmoji, emoji)
	f.reactRemove = append(f.reactRemove, remove)
	f.reactionIDs = append(f.reactionIDs, reactionID)
	return nil
}

func (f *fakeMessageSender) ForwardMessage(_ context.Context, target channelruntime.TextTarget, sourceMessageID string) error {
	if f.failFor != nil {
		if err, ok := f.failFor[target.ChatID]; ok {
			return err
		}
	}
	if f.err != nil {
		return f.err
	}
	f.forwardTargets = append(f.forwardTargets, target)
	f.forwardSourceIDs = append(f.forwardSourceIDs, sourceMessageID)
	return nil
}

func TestRegister_CurrentTarget(t *testing.T) {
	reg := canonicaltools.NewRegistry()
	Register(reg, Options{
		Platform: "feishu",
		Target: channelruntime.TextTarget{
			AccountID: "acct",
			ChatID:    "chat_1",
			MessageID: "msg_1",
		},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"current_target"}`,
	})
	if err != nil {
		t.Fatalf("current_target: %v", err)
	}
	var payload struct {
		Action    string                    `json:"action"`
		Platform  string                    `json:"platform"`
		Available bool                      `json:"available"`
		Status    string                    `json:"status"`
		Target    channelruntime.TextTarget `json:"target"`
		Actions   []string                  `json:"actions"`
		Warning   string                    `json:"warning"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "current_target" || payload.Platform != "feishu" {
		t.Fatalf("unexpected current target payload: %#v", payload)
	}
	if payload.Target.ChatID != "chat_1" || payload.Target.MessageID != "msg_1" {
		t.Fatalf("unexpected target payload: %#v", payload)
	}
	if payload.Available {
		t.Fatalf("expected target-only registration to be unavailable, got %#v", payload)
	}
	if payload.Status != "unavailable" {
		t.Fatalf("expected unavailable status, got %#v", payload)
	}
	if len(payload.Actions) != 8 {
		t.Fatalf("expected expanded message actions, got %#v", payload.Actions)
	}
}

func TestRegister_SendUsesDefaultTarget(t *testing.T) {
	reg := canonicaltools.NewRegistry()
	sender := &fakeMessageSender{}
	Register(reg, Options{
		Sender:   sender,
		Platform: "feishu",
		Target: channelruntime.TextTarget{
			AccountID: "acct",
			ChatID:    "chat_1",
			MessageID: "msg_1",
		},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"send","text":"hello"}`,
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if len(sender.targets) != 1 || sender.targets[0].ChatID != "chat_1" || sender.texts[0] != "hello" {
		t.Fatalf("unexpected send capture: targets=%#v texts=%#v", sender.targets, sender.texts)
	}
	if out == "" {
		t.Fatalf("expected non-empty send payload")
	}
}

func TestRegister_SendSupportsTargetOverride(t *testing.T) {
	reg := canonicaltools.NewRegistry()
	sender := &fakeMessageSender{}
	Register(reg, Options{
		Sender: sender,
		Target: channelruntime.TextTarget{
			AccountID: "acct",
			ChatID:    "chat_1",
			MessageID: "msg_1",
		},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"send","text":"hello","chat_id":"chat_2","message_id":"msg_2"}`,
	})
	if err != nil {
		t.Fatalf("send override: %v", err)
	}
	if len(sender.targets) != 1 || sender.targets[0].ChatID != "chat_2" || sender.targets[0].MessageID != "msg_2" {
		t.Fatalf("unexpected override target: %#v", sender.targets)
	}
}

func TestRegister_ReplyUsesCurrentMessageID(t *testing.T) {
	reg := canonicaltools.NewRegistry()
	sender := &fakeMessageSender{}
	Register(reg, Options{
		Sender: sender,
		Target: channelruntime.TextTarget{
			AccountID: "acct",
			ChatID:    "chat_1",
			MessageID: "msg_current",
		},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"reply","text":"follow up"}`,
	})
	if err != nil {
		t.Fatalf("reply: %v", err)
	}
	if len(sender.replyTargets) != 1 || sender.replyTargets[0].ChatID != "chat_1" || sender.replyTargets[0].MessageID != "msg_current" {
		t.Fatalf("unexpected reply target: %#v", sender.replyTargets)
	}
	var payload struct {
		Action    string                    `json:"action"`
		Status    string                    `json:"status"`
		SentCount int                       `json:"sent_count"`
		Target    channelruntime.TextTarget `json:"target"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode reply payload: %v", err)
	}
	if payload.Action != "reply" || payload.Status != "sent" || payload.SentCount != 1 {
		t.Fatalf("unexpected reply payload: %#v", payload)
	}
	if payload.Target.MessageID != "msg_current" {
		t.Fatalf("expected reply payload to preserve message id, got %#v", payload)
	}
}

func TestRegister_ReplySupportsThreadFlag(t *testing.T) {
	reg := canonicaltools.NewRegistry()
	sender := &fakeMessageSender{}
	Register(reg, Options{
		Sender: sender,
		Target: channelruntime.TextTarget{
			AccountID: "acct",
			ChatID:    "chat_1",
			MessageID: "msg_current",
		},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"reply","text":"threaded","reply_in_thread":true}`,
	})
	if err != nil {
		t.Fatalf("reply thread: %v", err)
	}
	if len(sender.replyTargets) != 1 || !sender.replyTargets[0].ReplyInThread {
		t.Fatalf("expected reply_in_thread target, got %#v", sender.replyTargets)
	}
}

func TestRegister_ForwardUsesCurrentMessageID(t *testing.T) {
	reg := canonicaltools.NewRegistry()
	sender := &fakeMessageSender{}
	Register(reg, Options{
		Sender: sender,
		Target: channelruntime.TextTarget{
			AccountID: "acct",
			ChatID:    "chat_current",
			MessageID: "msg_current",
		},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"forward","chat_id":"chat_dest"}`,
	})
	if err != nil {
		t.Fatalf("forward: %v", err)
	}
	if len(sender.forwardTargets) != 1 || sender.forwardTargets[0].ChatID != "chat_dest" {
		t.Fatalf("unexpected forward target: %#v", sender.forwardTargets)
	}
	if len(sender.forwardSourceIDs) != 1 || sender.forwardSourceIDs[0] != "msg_current" {
		t.Fatalf("unexpected forward source ids: %#v", sender.forwardSourceIDs)
	}
	var payload struct {
		Action   string                      `json:"action"`
		Status   string                      `json:"status"`
		SourceID string                      `json:"source_message_id"`
		Targets  []channelruntime.TextTarget `json:"targets"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode forward payload: %v", err)
	}
	if payload.Action != "forward" || payload.Status != "sent" || payload.SourceID != "msg_current" || len(payload.Targets) != 1 {
		t.Fatalf("unexpected forward payload: %#v", payload)
	}
}

func TestRegister_ForwardSupportsTargetListAndPartialFailures(t *testing.T) {
	raw := "secret=sk-agentx path=/private/channel.json query=forward provider_response=denied"
	reg := canonicaltools.NewRegistry()
	sender := &fakeMessageSender{
		failFor: map[string]error{
			"chat_bad": errors.New(raw),
		},
	}
	Register(reg, Options{
		Sender: sender,
		Target: channelruntime.TextTarget{
			AccountID: "acct",
			MessageID: "msg_current",
		},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"forward","source_message_id":"msg_src","targets":[{"chat_id":"chat_ok"},{"chat_id":"chat_bad"}]}`,
	})
	if err != nil {
		t.Fatalf("forward partial: %v", err)
	}
	var payload struct {
		Status    string `json:"status"`
		SentCount int    `json:"sent_count"`
		Warning   string `json:"warning"`
		Failures  []struct {
			Error string `json:"error"`
		} `json:"failures"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode partial forward payload: %v", err)
	}
	if payload.Status != "partial" || payload.SentCount != 1 || len(payload.Failures) != 1 {
		t.Fatalf("unexpected partial forward payload: %#v", payload)
	}
	if payload.Warning == "" {
		t.Fatalf("expected partial forward warning, got %#v", payload)
	}
	if strings.Contains(payload.Failures[0].Error, raw) || strings.Contains(payload.Failures[0].Error, "sk-agentx") {
		t.Fatalf("raw forward error reached model-facing payload: %#v", payload.Failures)
	}
	if !strings.Contains(payload.Failures[0].Error, "class=message_delivery code=forward_failed identity=") {
		t.Fatalf("expected safe forward error projection, got %#v", payload.Failures)
	}
}

func TestRegister_ForwardSupportsThreadTarget(t *testing.T) {
	reg := canonicaltools.NewRegistry()
	sender := &fakeMessageSender{}
	Register(reg, Options{
		Sender: sender,
		Target: channelruntime.TextTarget{
			AccountID: "acct",
			MessageID: "msg_current",
		},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"forward","source_message_id":"msg_src","thread_id":"omt_thread"}`,
	})
	if err != nil {
		t.Fatalf("forward thread: %v", err)
	}
	if len(sender.forwardTargets) != 1 || sender.forwardTargets[0].ThreadID != "omt_thread" {
		t.Fatalf("unexpected forward thread target: %#v", sender.forwardTargets)
	}
	var payload struct {
		Action   string                      `json:"action"`
		Status   string                      `json:"status"`
		Targets  []channelruntime.TextTarget `json:"targets"`
		SourceID string                      `json:"source_message_id"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode forward thread payload: %v", err)
	}
	if payload.Action != "forward" || payload.Status != "sent" || payload.SourceID != "msg_src" || len(payload.Targets) != 1 || payload.Targets[0].ThreadID != "omt_thread" {
		t.Fatalf("unexpected forward thread payload: %#v", payload)
	}
}

func TestRegister_BroadcastSupportsChatIDs(t *testing.T) {
	reg := canonicaltools.NewRegistry()
	sender := &fakeMessageSender{}
	Register(reg, Options{
		Sender: sender,
		Target: channelruntime.TextTarget{
			AccountID: "acct",
			ChatID:    "chat_default",
			MessageID: "msg_current",
		},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"broadcast","text":"hello all","chat_ids":["chat_a","chat_b","chat_a"]}`,
	})
	if err != nil {
		t.Fatalf("broadcast: %v", err)
	}
	if len(sender.targets) != 2 || sender.targets[0].ChatID != "chat_a" || sender.targets[1].ChatID != "chat_b" {
		t.Fatalf("unexpected broadcast targets: %#v", sender.targets)
	}
	var payload struct {
		Action    string                      `json:"action"`
		Status    string                      `json:"status"`
		SentCount int                         `json:"sent_count"`
		Targets   []channelruntime.TextTarget `json:"targets"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode broadcast payload: %v", err)
	}
	if payload.Action != "broadcast" || payload.Status != "sent" || payload.SentCount != 2 {
		t.Fatalf("unexpected broadcast payload: %#v", payload)
	}
}

func TestRegister_BroadcastSurfacesPartialFailures(t *testing.T) {
	raw := "secret=sk-agentx path=/private/channel.json query=broadcast provider_response=denied"
	reg := canonicaltools.NewRegistry()
	sender := &fakeMessageSender{
		failFor: map[string]error{
			"chat_bad": errors.New(raw),
		},
	}
	Register(reg, Options{
		Sender: sender,
		Target: channelruntime.TextTarget{
			AccountID: "acct",
			ChatID:    "chat_default",
		},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"broadcast","text":"hello all","targets":[{"chat_id":"chat_ok"},{"chat_id":"chat_bad"}]}`,
	})
	if err != nil {
		t.Fatalf("broadcast partial: %v", err)
	}
	var payload struct {
		Status    string `json:"status"`
		SentCount int    `json:"sent_count"`
		Warning   string `json:"warning"`
		Failures  []struct {
			Error string `json:"error"`
		} `json:"failures"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode partial broadcast payload: %v", err)
	}
	if payload.Status != "partial" || payload.SentCount != 1 || len(payload.Failures) != 1 {
		t.Fatalf("unexpected partial broadcast payload: %#v", payload)
	}
	if payload.Warning == "" {
		t.Fatalf("expected partial broadcast warning, got %#v", payload)
	}
	if strings.Contains(payload.Failures[0].Error, raw) || strings.Contains(payload.Failures[0].Error, "sk-agentx") {
		t.Fatalf("raw broadcast error reached model-facing payload: %#v", payload.Failures)
	}
	if !strings.Contains(payload.Failures[0].Error, "class=message_delivery code=broadcast_send_failed identity=") {
		t.Fatalf("expected safe broadcast error projection, got %#v", payload.Failures)
	}
}

func TestRegister_RejectsMissingTextForSend(t *testing.T) {
	reg := canonicaltools.NewRegistry()
	Register(reg, Options{
		Sender: &fakeMessageSender{},
		Target: channelruntime.TextTarget{ChatID: "chat_1"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"send"}`,
	})
	if err == nil {
		t.Fatalf("expected missing text error")
	}
	argErr, ok := agentxtoolerrors.AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != agentxtoolerrors.ToolArgumentErrorCodeMissingRequiredArgument {
		t.Fatalf("unexpected argument error code: %#v", argErr)
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"text"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
}

func TestRegister_RejectsMissingTargetsForBroadcast(t *testing.T) {
	reg := canonicaltools.NewRegistry()
	Register(reg, Options{
		Sender: &fakeMessageSender{},
		Target: channelruntime.TextTarget{ChatID: "chat_1"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"broadcast","text":"hello"}`,
	})
	if err == nil {
		t.Fatalf("expected missing targets error")
	}
	argErr, ok := agentxtoolerrors.AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != agentxtoolerrors.ToolArgumentErrorCodeMissingRequiredArgument {
		t.Fatalf("unexpected argument error code: %#v", argErr)
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"chat_ids", "targets"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
}

func TestRegister_EditUsesCurrentMessageID(t *testing.T) {
	reg := canonicaltools.NewRegistry()
	sender := &fakeMessageSender{}
	Register(reg, Options{
		Sender: sender,
		Target: channelruntime.TextTarget{
			AccountID: "acct",
			ChatID:    "chat_1",
			MessageID: "msg_current",
		},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"edit","text":"updated"}`,
	})
	if err != nil {
		t.Fatalf("edit: %v", err)
	}
	if len(sender.editTargets) != 1 || sender.editTargets[0].MessageID != "msg_current" || sender.editTexts[0] != "updated" {
		t.Fatalf("unexpected edit capture: targets=%#v texts=%#v", sender.editTargets, sender.editTexts)
	}
	var payload struct {
		Action string `json:"action"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode edit payload: %v", err)
	}
	if payload.Action != "edit" || payload.Status != "updated" {
		t.Fatalf("unexpected edit payload: %#v", payload)
	}
}

func TestRegister_DeleteUsesCurrentMessageID(t *testing.T) {
	reg := canonicaltools.NewRegistry()
	sender := &fakeMessageSender{}
	Register(reg, Options{
		Sender: sender,
		Target: channelruntime.TextTarget{
			AccountID: "acct",
			ChatID:    "chat_1",
			MessageID: "msg_current",
		},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"delete"}`,
	})
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(sender.deleteTargets) != 1 || sender.deleteTargets[0].MessageID != "msg_current" {
		t.Fatalf("unexpected delete capture: %#v", sender.deleteTargets)
	}
	var payload struct {
		Action string `json:"action"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode delete payload: %v", err)
	}
	if payload.Action != "delete" || payload.Status != "deleted" {
		t.Fatalf("unexpected delete payload: %#v", payload)
	}
}

func TestRegister_ReactUsesCurrentMessageID(t *testing.T) {
	reg := canonicaltools.NewRegistry()
	sender := &fakeMessageSender{}
	Register(reg, Options{
		Sender: sender,
		Target: channelruntime.TextTarget{
			AccountID: "acct",
			ChatID:    "chat_1",
			MessageID: "msg_current",
		},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"react","emoji":"smile"}`,
	})
	if err != nil {
		t.Fatalf("react: %v", err)
	}
	if len(sender.reactTargets) != 1 || sender.reactTargets[0].MessageID != "msg_current" || sender.reactEmoji[0] != "SMILE" || sender.reactRemove[0] {
		t.Fatalf("unexpected react capture: targets=%#v emoji=%#v remove=%#v", sender.reactTargets, sender.reactEmoji, sender.reactRemove)
	}
	var payload struct {
		Action string `json:"action"`
		Status string `json:"status"`
		Emoji  string `json:"emoji"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode react payload: %v", err)
	}
	if payload.Action != "react" || payload.Status != "reacted" || payload.Emoji != "SMILE" {
		t.Fatalf("unexpected react payload: %#v", payload)
	}
}

func TestRegister_ReactRemoveSupportsReactionID(t *testing.T) {
	reg := canonicaltools.NewRegistry()
	sender := &fakeMessageSender{}
	Register(reg, Options{
		Sender: sender,
		Target: channelruntime.TextTarget{
			AccountID: "acct",
			ChatID:    "chat_1",
			MessageID: "msg_current",
		},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      messageToolName,
		Arguments: `{"action":"react","remove":true,"reaction_id":"rid_1"}`,
	})
	if err != nil {
		t.Fatalf("react remove: %v", err)
	}
	if len(sender.reactTargets) != 1 || !sender.reactRemove[0] || sender.reactionIDs[0] != "rid_1" {
		t.Fatalf("unexpected react remove capture: remove=%#v ids=%#v", sender.reactRemove, sender.reactionIDs)
	}
	var payload struct {
		Status     string `json:"status"`
		Removed    bool   `json:"removed"`
		ReactionID string `json:"reaction_id"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode react remove payload: %v", err)
	}
	if payload.Status != "reaction_removed" || !payload.Removed || payload.ReactionID != "rid_1" {
		t.Fatalf("unexpected react remove payload: %#v", payload)
	}
}
