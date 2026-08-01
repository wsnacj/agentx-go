package channel_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wsnacj/agentx-go/runtime/channel"
)

type cancelRunner struct {
	started chan struct{}
	calls   atomic.Int32
}

func (r *cancelRunner) RunTurn(ctx context.Context, _ channel.Message) (string, error) {
	r.calls.Add(1)
	select {
	case r.started <- struct{}{}:
	default:
	}
	<-ctx.Done()
	return "", context.Cause(ctx)
}

func (*cancelRunner) WorkspaceDir() string { return "" }
func (*cancelRunner) Profile() string      { return "external-test" }

func TestBoundedIngressOverloadShutdownAndClosedContract(t *testing.T) {
	runner := &cancelRunner{started: make(chan struct{}, 1)}
	runtime := channel.NewIngressRuntime(channel.IngressRuntimeOptions{
		MaxConcurrency: 1,
		QueueCapacity:  1,
		OverloadPolicy: channel.IngressOverloadReject,
	})
	processor := channel.InboundProcessor{Runner: runner}

	first := runtime.Submit(context.Background(), processor, channel.Message{MessageID: "first"})
	if !first.Accepted || first.Reason != channel.IngressSubmitAccepted {
		t.Fatalf("first submit = %+v", first)
	}
	select {
	case <-runner.started:
	case <-time.After(time.Second):
		t.Fatal("worker did not start")
	}

	second := runtime.Submit(context.Background(), processor, channel.Message{MessageID: "second"})
	if !second.Accepted || second.Reason != channel.IngressSubmitAccepted {
		t.Fatalf("second submit = %+v", second)
	}
	third := runtime.Submit(context.Background(), processor, channel.Message{MessageID: "third"})
	if third.Accepted || third.Reason != channel.IngressSubmitOverloaded {
		t.Fatalf("third submit = %+v", third)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runtime.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	if err := runtime.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("second Shutdown: %v", err)
	}
	closed := runtime.Submit(context.Background(), processor, channel.Message{MessageID: "closed"})
	if closed.Accepted || closed.Reason != channel.IngressSubmitClosed {
		t.Fatalf("closed submit = %+v", closed)
	}
	if got := runner.calls.Load(); got != 1 {
		t.Fatalf("runner calls = %d, want 1", got)
	}
}

func TestSessionDeliveryContractIsDisplaySafeAndHostOwned(t *testing.T) {
	contract := channel.BuildSessionChannelContract(channel.SessionChannelContractInput{
		Source: channel.SessionSource{
			Platform:  "example",
			UserID:    "user-1",
			ChannelID: "channel-1",
		},
		Lifecycle:  channel.SessionLifecycleActive,
		Capability: channel.ChannelCapability{Text: true, Ready: true},
		Delivery: channel.ChannelDeliveryResult{
			Status:      channel.ChannelDeliveryBuffered,
			BufferedRef: "opaque-buffer-ref",
		},
	})
	if !contract.ReadyForProductShellSession {
		t.Fatalf("contract not ready: %+v", contract)
	}
	if !contract.HostChannelAdapterOwnsPlatformDelivery || contract.ChannelCapabilityControlsExecution {
		t.Fatalf("owner boundary changed: %+v", contract)
	}
	if contract.Delivery.NextHostAction != "flush_buffered_channel_delivery" {
		t.Fatalf("next host action = %q", contract.Delivery.NextHostAction)
	}
}
