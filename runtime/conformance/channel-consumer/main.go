package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wsnacj/agentx-go/runtime/channel"
)

type scriptedRunner struct{}

func (scriptedRunner) RunTurn(context.Context, channel.Message) (string, error) {
	return "channel-conformance", nil
}

func (scriptedRunner) WorkspaceDir() string { return "" }
func (scriptedRunner) Profile() string      { return "fixed-consumer" }

type memorySender struct {
	mu   sync.Mutex
	text string
	done chan struct{}
}

func (s *memorySender) SendText(_ context.Context, _ channel.TextTarget, text string) error {
	s.mu.Lock()
	s.text = text
	s.mu.Unlock()
	select {
	case s.done <- struct{}{}:
	default:
	}
	return nil
}

func (s *memorySender) value() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.text
}

func run(ctx context.Context) (string, error) {
	sender := &memorySender{done: make(chan struct{}, 1)}
	runtime := channel.NewIngressRuntime(channel.IngressRuntimeOptions{
		MaxConcurrency: 1,
		QueueCapacity:  2,
		OverloadPolicy: channel.IngressOverloadReject,
	})
	processor := channel.InboundProcessor{
		Runner:  scriptedRunner{},
		Sender:  sender,
		Deduper: channel.NewDeduper(time.Minute),
		BuildReservations: func(message channel.Message) []channel.DedupeReservation {
			return []channel.DedupeReservation{{Key: channel.BuildContentDedupeKey(message)}}
		},
	}
	message := channel.Message{
		Platform:  "example",
		AccountID: "account-1",
		SessionID: "session-1",
		MessageID: "message-1",
		ChatID:    "channel-1",
		UserID:    "user-1",
		Text:      "run",
	}
	result := runtime.Submit(ctx, processor, message)
	if !result.Accepted || result.Reason != channel.IngressSubmitAccepted {
		return "", fmt.Errorf("submit: %+v", result)
	}
	select {
	case <-sender.done:
	case <-ctx.Done():
		return "", ctx.Err()
	}
	if err := runtime.Shutdown(ctx); err != nil {
		return "", err
	}
	contract := channel.BuildSessionChannelContract(channel.SessionChannelContractInput{
		Source:     channel.SessionSourceFromMessage(message),
		Lifecycle:  channel.SessionLifecycleActive,
		Capability: channel.ChannelCapabilityFromSender(sender, 1024),
		Delivery: channel.ChannelDeliveryResult{
			Status:    channel.ChannelDeliverySent,
			MessageID: message.MessageID,
		},
	})
	if !contract.ReadyForProductShellSession || !contract.HostChannelAdapterOwnsPlatformDelivery {
		return "", fmt.Errorf("session contract not ready: %+v", contract)
	}
	return fmt.Sprintf("agentx-channel-ok:%s:%s:session_channel_ready", contract.Delivery.Status, sender.value()), nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	output, err := run(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}
