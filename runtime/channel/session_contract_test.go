package channel

import "testing"

func TestSessionSourceFromMessageBuildsDeterministicKey(t *testing.T) {
	message := Message{
		Platform:  " Feishu ",
		AccountID: "main",
		SessionID: "session-1",
		ChatID:    "chat-1",
		ThreadID:  "thread-1",
		UserID:    "user-1",
	}

	source := SessionSourceFromMessage(message).Normalize()
	if source.Platform != "feishu" ||
		source.ChannelID != "chat-1" ||
		source.SourceKind != "channel_message" {
		t.Fatalf("unexpected source normalization: %#v", source)
	}
	key := source.DeterministicSessionKey()
	if key == "" || key != DeterministicSessionKey(source) {
		t.Fatalf("expected deterministic key, got %q", key)
	}

	same := SessionSourceFromMessage(message).Normalize()
	same.Platform = "feishu"
	if same.DeterministicSessionKey() != key {
		t.Fatalf("expected stable deterministic key")
	}

	different := same
	different.ThreadID = "thread-2"
	if different.DeterministicSessionKey() == key {
		t.Fatalf("expected thread change to change deterministic key")
	}
}

func TestBuildSessionChannelContractReadyForProductShellSession(t *testing.T) {
	contract := BuildSessionChannelContract(SessionChannelContractInput{
		Source: SessionSource{
			Platform:    "slack",
			AccountID:   "main",
			UserID:      "user-1",
			ChannelID:   "channel-1",
			WorkspaceID: "workspace-1",
		},
		Lifecycle:  "resume-pending",
		Capability: ChannelCapability{Text: true, Markdown: true, Thread: true, MaxLength: 4096},
		Delivery:   ChannelDeliveryResult{Status: ChannelDeliveryBuffered, BufferedRef: "buffer:reply"},
		Boundaries: []string{"test_boundary"},
	})

	if contract.ContractVersion != SessionChannelContractVersion {
		t.Fatalf("contract version = %q", contract.ContractVersion)
	}
	if contract.Lifecycle != SessionLifecycleResumePending {
		t.Fatalf("lifecycle = %q", contract.Lifecycle)
	}
	if contract.SessionKey == "" ||
		!contract.SessionSourceReady ||
		!contract.ChannelCapabilityReady ||
		!contract.DeliveryResultReady ||
		!contract.ReadyForProductShellSession {
		t.Fatalf("expected ready contract: %#v", contract)
	}
	if contract.ChannelCapabilityControlsExecution ||
		contract.PlatformMessageRulesInEngine ||
		!contract.HostChannelAdapterOwnsPlatformDelivery {
		t.Fatalf("unexpected ownership flags: %#v", contract)
	}
	for _, boundary := range []string{
		"runtime_channel_session_contract",
		"product_shell_session_channel_contract",
		"host_channel_adapter_owns_platform_delivery",
		"channel_capability_does_not_control_execution_policy",
		"no_platform_message_rules_in_engine",
		"test_boundary",
	} {
		if !sessionContractTestContains(contract.Boundaries, boundary) {
			t.Fatalf("missing boundary %q in %#v", boundary, contract.Boundaries)
		}
	}
}

func TestBuildSessionChannelContractFailClosedForMissingSourceAndCapability(t *testing.T) {
	contract := BuildSessionChannelContract(SessionChannelContractInput{
		Source:     SessionSource{Platform: "teams"},
		Lifecycle:  SessionLifecycleExpired,
		Capability: ChannelCapability{},
		Delivery:   ChannelDeliveryResult{FailureReason: "rate_limited"},
	})

	if contract.ReadyForProductShellSession {
		t.Fatalf("expected contract to block: %#v", contract)
	}
	for _, missing := range []string{
		"channel:session_user",
		"channel:session_channel_or_thread_or_workspace",
		"channel:text_capability",
	} {
		if !sessionContractTestContains(contract.MissingInputs, missing) {
			t.Fatalf("missing input %q not reported: %#v", missing, contract.MissingInputs)
		}
	}
	for _, reason := range []string{
		"session_source_incomplete",
		"channel_text_capability_missing",
		"session_lifecycle_expired",
		"channel_delivery_failed",
	} {
		if !sessionContractTestContains(contract.BlockedReasons, reason) {
			t.Fatalf("blocked reason %q not reported: %#v", reason, contract.BlockedReasons)
		}
	}
	if contract.NextHostAction != "provide_session_channel_contract" {
		t.Fatalf("next host action = %q", contract.NextHostAction)
	}
}

func TestChannelCapabilityFromSenderDetectsOptionalInterfaces(t *testing.T) {
	sender := &recordingSender{}
	capability := ChannelCapabilityFromSender(sender, -1)

	if !capability.Ready ||
		!capability.Text ||
		!capability.Reply ||
		!capability.Edit ||
		!capability.Delete ||
		!capability.React ||
		!capability.Forward {
		t.Fatalf("unexpected capability from sender: %#v", capability)
	}
	if capability.MaxLength != 0 {
		t.Fatalf("negative max length should normalize to 0: %#v", capability)
	}

	empty := ChannelCapabilityFromSender(nil, 100)
	if empty.Ready || empty.Text {
		t.Fatalf("nil sender should not be ready: %#v", empty)
	}
}

func TestChannelDeliveryResultNormalizesStatus(t *testing.T) {
	sent := ChannelDeliveryResult{MessageID: "msg-1"}.Normalize()
	if sent.Status != ChannelDeliverySent || !sent.Sent || !sent.DeliveryReadyForTest() {
		t.Fatalf("expected sent delivery: %#v", sent)
	}

	action := ChannelDeliveryResult{RequiresUserAction: true}.Normalize()
	if action.Status != ChannelDeliveryRequiresUserAction ||
		!action.RequiresUserAction ||
		action.NextHostAction != "request_channel_user_action" {
		t.Fatalf("expected requires user action delivery: %#v", action)
	}
}

func (r ChannelDeliveryResult) DeliveryReadyForTest() bool {
	return deliveryResultReady(r.Status)
}

func sessionContractTestContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
