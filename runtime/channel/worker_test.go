package channel

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestInboundProcessorReleasesDedupeOnRunnerError(t *testing.T) {
	d := NewDeduper(time.Minute)
	processor := InboundProcessor{
		Runner:  workerErrorRunner{err: errors.New("run failed")},
		Sender:  &workerRecordingSender{},
		Deduper: d,
		BuildReservations: func(message Message) []DedupeReservation {
			return []DedupeReservation{{Key: message.MessageID}}
		},
	}

	err := processor.Process(context.Background(), Message{MessageID: "m1", ChatID: "c1"})
	if err == nil || !strings.Contains(err.Error(), "agent run failed") {
		t.Fatalf("expected wrapped runner error, got %v", err)
	}
	if !d.Begin("m1") {
		t.Fatalf("expected dedupe reservation to be released after runner error")
	}
}

func TestInboundProcessorReleasesDedupeOnSendError(t *testing.T) {
	d := NewDeduper(time.Minute)
	processor := InboundProcessor{
		Runner:  workerReplyRunner{reply: "hello"},
		Sender:  &workerErrorSender{err: errors.New("send failed")},
		Deduper: d,
		BuildReservations: func(message Message) []DedupeReservation {
			return []DedupeReservation{{Key: message.MessageID}}
		},
	}

	err := processor.Process(context.Background(), Message{MessageID: "m1", ChatID: "c1"})
	if err == nil || !strings.Contains(err.Error(), "send reply failed") {
		t.Fatalf("expected wrapped send error, got %v", err)
	}
	if !d.Begin("m1") {
		t.Fatalf("expected dedupe reservation to be released after send error")
	}
}

func TestInboundProcessorProcessAsyncReturnsFalseWhenReservationBlocked(t *testing.T) {
	d := NewDeduper(time.Minute)
	if !d.Begin("m1") {
		t.Fatalf("expected initial reservation to succeed")
	}

	processor := InboundProcessor{
		Runner:  workerReplyRunner{reply: "hello"},
		Sender:  &workerRecordingSender{},
		Deduper: d,
		Ingress: newTestIngressRuntime(t),
		BuildReservations: func(message Message) []DedupeReservation {
			return []DedupeReservation{{Key: message.MessageID}}
		},
	}

	result := processor.SubmitAsync(context.Background(), Message{MessageID: "m1", ChatID: "c1"})
	if result.Accepted || result.Reason != IngressSubmitDuplicate {
		t.Fatalf("duplicate submission = %#v, want duplicate rejection", result)
	}
}

func TestInboundProcessorLogsHandledEvents(t *testing.T) {
	var buf strings.Builder
	sender := &workerRecordingSender{}
	processor := InboundProcessor{
		Runner:  workerReplyRunner{reply: "hello"},
		Sender:  sender,
		Logger:  NewWriterLogger(&buf),
		Deduper: NewDeduper(time.Minute),
		BuildReservations: func(message Message) []DedupeReservation {
			return []DedupeReservation{{Key: message.MessageID}}
		},
	}

	if err := processor.Process(context.Background(), Message{
		Platform:  "feishu",
		AccountID: "main",
		SessionID: "session-1",
		ChatID:    "c1",
		MessageID: "m1",
	}); err != nil {
		t.Fatalf("Process: %v", err)
	}
	if !strings.Contains(buf.String(), "channel event handled platform=feishu") {
		t.Fatalf("expected handled event log, got %q", buf.String())
	}
	if len(sender.texts) != 1 || sender.texts[0] != "hello" {
		t.Fatalf("unexpected sender texts: %#v", sender.texts)
	}
}

type workerReplyRunner struct {
	reply string
}

func (r workerReplyRunner) RunTurn(context.Context, Message) (string, error) { return r.reply, nil }
func (r workerReplyRunner) WorkspaceDir() string                             { return "." }
func (r workerReplyRunner) Profile() string                                  { return "safe" }

type workerErrorRunner struct {
	err error
}

func (r workerErrorRunner) RunTurn(context.Context, Message) (string, error) { return "", r.err }
func (r workerErrorRunner) WorkspaceDir() string                             { return "." }
func (r workerErrorRunner) Profile() string                                  { return "safe" }

type workerRecordingSender struct {
	texts []string
}

func (s *workerRecordingSender) SendText(_ context.Context, _ TextTarget, text string) error {
	s.texts = append(s.texts, text)
	return nil
}

type workerErrorSender struct {
	err error
}

func (s *workerErrorSender) SendText(context.Context, TextTarget, string) error {
	return s.err
}

func newTestIngressRuntime(t *testing.T) *IngressRuntime {
	t.Helper()
	runtime := NewIngressRuntime(IngressRuntimeOptions{
		MaxConcurrency: 1,
		QueueCapacity:  8,
		CloseTimeout:   time.Second,
	})
	t.Cleanup(func() {
		if err := runtime.Close(); err != nil {
			t.Errorf("close ingress runtime: %v", err)
		}
	})
	return runtime
}
