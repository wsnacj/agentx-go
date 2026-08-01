package channel

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const SessionChannelContractVersion = "agentx.session_channel_contract.v1"

type SessionLifecycle string

const (
	SessionLifecycleActive        SessionLifecycle = "active"
	SessionLifecycleSuspended     SessionLifecycle = "suspended"
	SessionLifecycleResumePending SessionLifecycle = "resume_pending"
	SessionLifecycleExpired       SessionLifecycle = "expired"
	SessionLifecycleUnknown       SessionLifecycle = "unknown"
)

type ChannelDeliveryStatus string

const (
	ChannelDeliveryUnknown            ChannelDeliveryStatus = "unknown"
	ChannelDeliverySent               ChannelDeliveryStatus = "sent"
	ChannelDeliveryBuffered           ChannelDeliveryStatus = "buffered"
	ChannelDeliveryFailed             ChannelDeliveryStatus = "failed"
	ChannelDeliveryRequiresUserAction ChannelDeliveryStatus = "requires_user_action"
)

type SessionSource struct {
	SourceKind  string `json:"source_kind,omitempty"`
	Platform    string `json:"platform,omitempty"`
	AccountID   string `json:"account_id,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
	ChannelID   string `json:"channel_id,omitempty"`
	ChatID      string `json:"chat_id,omitempty"`
	ThreadID    string `json:"thread_id,omitempty"`
	WorkspaceID string `json:"workspace_id,omitempty"`
}

type ChannelCapability struct {
	CapabilityRef string   `json:"capability_ref,omitempty"`
	Text          bool     `json:"text"`
	Markdown      bool     `json:"markdown"`
	Reply         bool     `json:"reply"`
	Edit          bool     `json:"edit"`
	Draft         bool     `json:"draft"`
	Thread        bool     `json:"thread"`
	Attachments   bool     `json:"attachments"`
	Delete        bool     `json:"delete"`
	React         bool     `json:"react"`
	Forward       bool     `json:"forward"`
	MaxLength     int      `json:"max_length,omitempty"`
	Ready         bool     `json:"ready"`
	Boundaries    []string `json:"boundaries,omitempty"`
}

type ChannelDeliveryResult struct {
	DeliveryRef        string                `json:"delivery_ref,omitempty"`
	Status             ChannelDeliveryStatus `json:"status,omitempty"`
	Target             TextTarget            `json:"target,omitempty"`
	MessageID          string                `json:"message_id,omitempty"`
	BufferedRef        string                `json:"buffered_ref,omitempty"`
	FailureReason      string                `json:"failure_reason,omitempty"`
	RequiresUserAction bool                  `json:"requires_user_action"`
	Sent               bool                  `json:"sent"`
	Buffered           bool                  `json:"buffered"`
	Failed             bool                  `json:"failed"`
	Boundaries         []string              `json:"boundaries,omitempty"`
	NextHostAction     string                `json:"next_host_action,omitempty"`
}

type SessionChannelContractInput struct {
	Source          SessionSource         `json:"source,omitempty"`
	Lifecycle       SessionLifecycle      `json:"lifecycle,omitempty"`
	Capability      ChannelCapability     `json:"capability,omitempty"`
	Delivery        ChannelDeliveryResult `json:"delivery,omitempty"`
	Boundaries      []string              `json:"boundaries,omitempty"`
	NextHostAction  string                `json:"next_host_action,omitempty"`
	RawOutputLoaded bool                  `json:"raw_output_loaded"`
}

type SessionChannelContract struct {
	ContractVersion                        string                `json:"contract_version,omitempty"`
	SessionKey                             string                `json:"session_key,omitempty"`
	Source                                 SessionSource         `json:"source,omitempty"`
	Lifecycle                              SessionLifecycle      `json:"lifecycle,omitempty"`
	Capability                             ChannelCapability     `json:"capability,omitempty"`
	Delivery                               ChannelDeliveryResult `json:"delivery,omitempty"`
	SessionSourceReady                     bool                  `json:"session_source_ready"`
	ChannelCapabilityReady                 bool                  `json:"channel_capability_ready"`
	DeliveryResultReady                    bool                  `json:"delivery_result_ready"`
	ReadyForProductShellSession            bool                  `json:"ready_for_product_shell_session"`
	ChannelCapabilityControlsExecution     bool                  `json:"channel_capability_controls_execution"`
	PlatformMessageRulesInEngine           bool                  `json:"platform_message_rules_in_engine"`
	HostChannelAdapterOwnsPlatformDelivery bool                  `json:"host_channel_adapter_owns_platform_delivery"`
	MissingInputs                          []string              `json:"missing_inputs,omitempty"`
	BlockedReasons                         []string              `json:"blocked_reasons,omitempty"`
	Boundaries                             []string              `json:"boundaries,omitempty"`
	NextHostAction                         string                `json:"next_host_action,omitempty"`
	RawOutputLoaded                        bool                  `json:"raw_output_loaded"`
}

func SessionSourceFromMessage(message Message) SessionSource {
	return SessionSource{
		SourceKind: "channel_message",
		Platform:   message.Platform,
		AccountID:  message.AccountID,
		UserID:     message.UserID,
		SessionID:  message.SessionID,
		ChannelID:  firstNonEmpty(message.ChatID, message.ThreadID, message.SessionID),
		ChatID:     message.ChatID,
		ThreadID:   message.ThreadID,
	}
}

func BuildSessionChannelContract(input SessionChannelContractInput) SessionChannelContract {
	return SessionChannelContract{
		ContractVersion: SessionChannelContractVersion,
		Source:          input.Source,
		Lifecycle:       input.Lifecycle,
		Capability:      input.Capability,
		Delivery:        input.Delivery,
		Boundaries: sessionContractUniqueStrings([]string{
			"runtime_channel_session_contract",
			"product_shell_session_channel_contract",
			"host_channel_adapter_owns_platform_delivery",
			"channel_capability_does_not_control_execution_policy",
			"no_platform_message_rules_in_engine",
		}, input.Boundaries...),
		NextHostAction:  sessionContractFirstNonEmpty(input.NextHostAction, "review_session_channel_contract"),
		RawOutputLoaded: input.RawOutputLoaded,
	}.Normalize()
}

func (s SessionSource) Normalize() SessionSource {
	out := s
	out.SourceKind = sessionContractToken(out.SourceKind)
	if out.SourceKind == "" {
		out.SourceKind = "channel_message"
	}
	out.Platform = strings.ToLower(strings.TrimSpace(out.Platform))
	out.AccountID = strings.TrimSpace(out.AccountID)
	out.UserID = strings.TrimSpace(out.UserID)
	out.SessionID = strings.TrimSpace(out.SessionID)
	out.ChannelID = strings.TrimSpace(out.ChannelID)
	out.ChatID = strings.TrimSpace(out.ChatID)
	out.ThreadID = strings.TrimSpace(out.ThreadID)
	out.WorkspaceID = strings.TrimSpace(out.WorkspaceID)
	if out.ChannelID == "" {
		out.ChannelID = firstNonEmpty(out.ChatID, out.ThreadID, out.SessionID, out.WorkspaceID)
	}
	return out
}

func (s SessionSource) Ready() bool {
	normalized := s.Normalize()
	return normalized.Platform != "" &&
		normalized.UserID != "" &&
		firstNonEmpty(normalized.ChannelID, normalized.ThreadID, normalized.WorkspaceID, normalized.SessionID) != ""
}

func (s SessionSource) DeterministicSessionKey() string {
	return DeterministicSessionKey(s)
}

func (c ChannelCapability) Normalize() ChannelCapability {
	out := c
	out.CapabilityRef = strings.TrimSpace(out.CapabilityRef)
	if out.MaxLength < 0 {
		out.MaxLength = 0
	}
	out.Boundaries = sessionContractUniqueStrings([]string{
		"channel_capability_contract",
		"channel_capability_does_not_control_execution_policy",
	}, out.Boundaries...)
	out.Ready = out.Text
	return out
}

func ChannelCapabilityFromSender(sender TextSender, maxLength int) ChannelCapability {
	capability := ChannelCapability{
		Text:      sender != nil,
		MaxLength: maxLength,
	}
	if sender == nil {
		return capability.Normalize()
	}
	_, capability.Reply = sender.(ReplySender)
	_, capability.Edit = sender.(EditSender)
	_, capability.Delete = sender.(DeleteSender)
	_, capability.React = sender.(ReactSender)
	_, capability.Forward = sender.(ForwardSender)
	return capability.Normalize()
}

func (r ChannelDeliveryResult) Normalize() ChannelDeliveryResult {
	out := r
	out.DeliveryRef = strings.TrimSpace(out.DeliveryRef)
	out.Target = normalizeTextTarget(out.Target)
	out.MessageID = strings.TrimSpace(out.MessageID)
	out.BufferedRef = strings.TrimSpace(out.BufferedRef)
	out.FailureReason = strings.TrimSpace(out.FailureReason)
	out.Status = NormalizeChannelDeliveryStatus(string(out.Status))
	switch {
	case out.RequiresUserAction:
		out.Status = ChannelDeliveryRequiresUserAction
	case out.FailureReason != "":
		out.Status = ChannelDeliveryFailed
	case out.BufferedRef != "":
		out.Status = ChannelDeliveryBuffered
	case out.MessageID != "":
		out.Status = ChannelDeliverySent
	}
	out.Sent = out.Status == ChannelDeliverySent
	out.Buffered = out.Status == ChannelDeliveryBuffered
	out.Failed = out.Status == ChannelDeliveryFailed
	out.RequiresUserAction = out.Status == ChannelDeliveryRequiresUserAction
	out.Boundaries = sessionContractUniqueStrings([]string{
		"channel_delivery_result",
		"host_channel_adapter_owns_platform_delivery",
	}, out.Boundaries...)
	out.NextHostAction = sessionContractFirstNonEmpty(out.NextHostAction, nextHostActionForDeliveryStatus(out.Status))
	return out
}

func (c SessionChannelContract) Normalize() SessionChannelContract {
	out := c
	out.ContractVersion = strings.TrimSpace(out.ContractVersion)
	if out.ContractVersion == "" {
		out.ContractVersion = SessionChannelContractVersion
	}
	out.Source = out.Source.Normalize()
	out.Lifecycle = NormalizeSessionLifecycle(string(out.Lifecycle))
	out.Capability = out.Capability.Normalize()
	out.Delivery = out.Delivery.Normalize()
	out.SessionKey = DeterministicSessionKey(out.Source)
	out.SessionSourceReady = out.Source.Ready() && out.SessionKey != ""
	out.ChannelCapabilityReady = out.Capability.Ready
	out.DeliveryResultReady = deliveryResultReady(out.Delivery.Status)
	out.ChannelCapabilityControlsExecution = false
	out.PlatformMessageRulesInEngine = false
	out.HostChannelAdapterOwnsPlatformDelivery = true
	out.MissingInputs = sessionContractUniqueStrings(nil, out.MissingInputs...)
	out.BlockedReasons = sessionContractUniqueStrings(nil, out.BlockedReasons...)
	out.Boundaries = sessionContractUniqueStrings([]string{
		"runtime_channel_session_contract",
		"product_shell_session_channel_contract",
		"host_channel_adapter_owns_platform_delivery",
		"channel_capability_does_not_control_execution_policy",
		"no_platform_message_rules_in_engine",
	}, out.Boundaries...)
	out.NextHostAction = sessionContractFirstNonEmpty(out.NextHostAction, "review_session_channel_contract")
	out = out.requireSessionSource()
	if !out.ChannelCapabilityReady {
		out.MissingInputs = sessionContractUniqueStrings(out.MissingInputs, "channel:text_capability")
		out.BlockedReasons = sessionContractUniqueStrings(out.BlockedReasons, "channel_text_capability_missing")
	}
	if out.Lifecycle == SessionLifecycleExpired {
		out.BlockedReasons = sessionContractUniqueStrings(out.BlockedReasons, "session_lifecycle_expired")
	}
	if out.Delivery.Status == ChannelDeliveryFailed {
		out.BlockedReasons = sessionContractUniqueStrings(out.BlockedReasons, "channel_delivery_failed")
	}
	out.ReadyForProductShellSession = !out.RawOutputLoaded &&
		out.SessionSourceReady &&
		out.ChannelCapabilityReady &&
		out.Lifecycle != SessionLifecycleExpired &&
		len(out.MissingInputs) == 0 &&
		len(out.BlockedReasons) == 0
	if !out.ReadyForProductShellSession && out.NextHostAction == "review_session_channel_contract" {
		out.NextHostAction = "provide_session_channel_contract"
	}
	return out
}

func (c SessionChannelContract) requireSessionSource() SessionChannelContract {
	if c.Source.Platform == "" {
		c.MissingInputs = sessionContractUniqueStrings(c.MissingInputs, "channel:session_platform")
	}
	if c.Source.UserID == "" {
		c.MissingInputs = sessionContractUniqueStrings(c.MissingInputs, "channel:session_user")
	}
	if firstNonEmpty(c.Source.ChannelID, c.Source.ThreadID, c.Source.WorkspaceID, c.Source.SessionID) == "" {
		c.MissingInputs = sessionContractUniqueStrings(c.MissingInputs, "channel:session_channel_or_thread_or_workspace")
	}
	if len(c.MissingInputs) > 0 {
		c.BlockedReasons = sessionContractUniqueStrings(c.BlockedReasons, "session_source_incomplete")
	}
	return c
}

func NormalizeSessionLifecycle(raw string) SessionLifecycle {
	switch sessionContractToken(raw) {
	case string(SessionLifecycleSuspended):
		return SessionLifecycleSuspended
	case string(SessionLifecycleResumePending):
		return SessionLifecycleResumePending
	case string(SessionLifecycleExpired):
		return SessionLifecycleExpired
	case string(SessionLifecycleUnknown):
		return SessionLifecycleUnknown
	default:
		return SessionLifecycleActive
	}
}

func NormalizeChannelDeliveryStatus(raw string) ChannelDeliveryStatus {
	switch sessionContractToken(raw) {
	case string(ChannelDeliverySent):
		return ChannelDeliverySent
	case string(ChannelDeliveryBuffered):
		return ChannelDeliveryBuffered
	case string(ChannelDeliveryFailed):
		return ChannelDeliveryFailed
	case string(ChannelDeliveryRequiresUserAction):
		return ChannelDeliveryRequiresUserAction
	default:
		return ChannelDeliveryUnknown
	}
}

func DeterministicSessionKey(source SessionSource) string {
	normalized := source.Normalize()
	if !normalized.Ready() {
		return ""
	}
	material := strings.Join([]string{
		normalized.Platform,
		normalized.AccountID,
		normalized.ChannelID,
		normalized.ChatID,
		normalized.ThreadID,
		normalized.WorkspaceID,
		normalized.UserID,
		normalized.SessionID,
	}, "\x1f")
	sum := sha256.Sum256([]byte(material))
	return "session:" + hex.EncodeToString(sum[:12])
}

func deliveryResultReady(status ChannelDeliveryStatus) bool {
	return status == ChannelDeliverySent ||
		status == ChannelDeliveryBuffered ||
		status == ChannelDeliveryRequiresUserAction
}

func nextHostActionForDeliveryStatus(status ChannelDeliveryStatus) string {
	switch status {
	case ChannelDeliverySent:
		return "monitor_channel_delivery"
	case ChannelDeliveryBuffered:
		return "flush_buffered_channel_delivery"
	case ChannelDeliveryFailed:
		return "repair_channel_delivery"
	case ChannelDeliveryRequiresUserAction:
		return "request_channel_user_action"
	default:
		return "review_channel_delivery"
	}
}

func normalizeTextTarget(target TextTarget) TextTarget {
	target.AccountID = strings.TrimSpace(target.AccountID)
	target.ChatID = strings.TrimSpace(target.ChatID)
	target.ThreadID = strings.TrimSpace(target.ThreadID)
	target.MessageID = strings.TrimSpace(target.MessageID)
	return target
}

func sessionContractToken(raw string) string {
	token := strings.ToLower(strings.TrimSpace(raw))
	token = strings.NewReplacer(" ", "_", "-", "_").Replace(token)
	return token
}

func sessionContractFirstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func sessionContractUniqueStrings(items []string, values ...string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items)+len(values))
	for _, value := range append(append([]string{}, items...), values...) {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
