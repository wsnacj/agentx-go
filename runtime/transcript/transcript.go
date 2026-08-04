package transcript

import (
	"fmt"
	"regexp"
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

const (
	compactedToolOutputPlaceholder = "[compacted tool output to fit context window]"
	compactedHistoryPlaceholder    = "[compacted earlier context]"
)

// AnchorSelector locates a host-significant suffix that should be preserved
// when a tool result is truncated. It must be deterministic and safe for
// concurrent use by the caller.
type AnchorSelector func(content string) int

// EstimateInput describes the complete character-budget input. RoleAware
// selects Messages even when the conversation is empty; otherwise Chunks are
// counted.
type EstimateInput struct {
	SystemPrompt string
	Chunks       []string
	Messages     llm.Conversation
	RoleAware    bool
}

// GuardPolicy contains independent warning and hard-limit thresholds. Zero
// disables the corresponding threshold.
type GuardPolicy struct {
	WarnChars int
	MaxChars  int
}

// GuardDecision is a side-effect-free budget evaluation. Hosts own logging,
// telemetry and recovery behavior.
type GuardDecision struct {
	EstimatedChars int
	Warn           bool
	Overflow       bool
}

// CompactionPolicy controls deterministic in-memory compaction.
type CompactionPolicy struct {
	MaxChars         int
	ToolOutputAnchor AnchorSelector
}

// SanitizePolicy controls provider protocol repair without naming a provider.
// The host maps model/provider configuration to this policy.
type SanitizePolicy struct {
	StrictToolProtocol     bool
	StripInternalReasoning bool
}

// HistoryPolicy controls protocol-aware tail pruning.
type HistoryPolicy struct {
	MaxEvents          int
	StrictToolProtocol bool
}

// Diagnostic reports observable transformations performed by this package.
type Diagnostic struct {
	Passes                    int
	StrictProvider            bool
	SynthesizedToolCallIDs    int
	RecoveredToolResults      int
	DowngradedToolResults     int
	StrippedReasoningMsgs     int
	MergedMessages            int
	CompactedToolOutputs      int
	CompactedHistoryBodies    int
	ProtocolAwareHistoryDrops int
}

// EstimateChars returns a conservative character count for prompt, messages
// and tool-call arguments. It intentionally does not guess tokenization.
func EstimateChars(input EstimateInput) int {
	total := len(input.SystemPrompt)
	if input.RoleAware {
		for _, message := range input.Messages {
			total += len(message.Content)
			for _, call := range message.ToolCalls {
				total += len(call.Name) + len(call.Arguments)
			}
		}
		return total
	}
	for _, chunk := range input.Chunks {
		total += len(chunk)
	}
	return total
}

// Evaluate evaluates warning and overflow thresholds without side effects.
func Evaluate(input EstimateInput, policy GuardPolicy) GuardDecision {
	estimated := EstimateChars(input)
	return GuardDecision{
		EstimatedChars: estimated,
		Warn:           policy.WarnChars > 0 && estimated >= policy.WarnChars,
		Overflow:       policy.MaxChars > 0 && estimated >= policy.MaxChars,
	}
}

// OverflowMessage preserves the canonical display text used by AgentX hosts.
func OverflowMessage(estimatedChars, maxChars int) string {
	return fmt.Sprintf("context limit exceeded for this model (estimated %d chars, limit %d)", estimatedChars, maxChars)
}

// Compact reduces tool results first and older user/assistant bodies second.
// It always returns a defensive conversation copy when compaction is enabled.
func Compact(messages llm.Conversation, policy CompactionPolicy) (llm.Conversation, Diagnostic) {
	toolCompacted, toolCount := compactToolOutputs(messages, policy)
	historyCompacted, historyCount := compactHistoryBodies(toolCompacted, policy.MaxChars)
	return historyCompacted, Diagnostic{
		CompactedToolOutputs:   toolCount,
		CompactedHistoryBodies: historyCount,
	}
}

// Sanitize repairs role/tool-call protocol, optionally strips internal
// reasoning blocks, and merges safe adjacent messages.
func Sanitize(messages llm.Conversation, policy SanitizePolicy) (llm.Conversation, Diagnostic) {
	if len(messages) == 0 {
		return nil, Diagnostic{}
	}
	diagnostic := Diagnostic{Passes: 1, StrictProvider: policy.StrictToolProtocol}
	repaired := repairProtocol(messages, policy.StrictToolProtocol, &diagnostic)
	if len(repaired) == 0 {
		return nil, diagnostic
	}
	if policy.StripInternalReasoning {
		repaired = stripInternalReasoningBlocks(repaired, &diagnostic)
	}
	repaired = mergeAdjacentMessages(repaired, &diagnostic)
	return repaired, diagnostic
}

// Prune keeps the newest complete protocol segments within MaxEvents.
func Prune(messages llm.Conversation, policy HistoryPolicy) (llm.Conversation, int) {
	return prune(messages, policy.MaxEvents, policy.StrictToolProtocol)
}

// PruneTailPreservingSystemPrefix preserves a leading system-message prefix
// while applying Prune to the remaining conversation.
func PruneTailPreservingSystemPrefix(messages llm.Conversation, policy HistoryPolicy) (llm.Conversation, int) {
	if policy.MaxEvents <= 0 || len(messages) == 0 {
		return messages, 0
	}
	prefixEnd := 0
	for prefixEnd < len(messages) && normalizeRole(messages[prefixEnd].Role) == "system" {
		prefixEnd++
	}
	pruned, drops := prune(messages[prefixEnd:], policy.MaxEvents, policy.StrictToolProtocol)
	if prefixEnd == 0 {
		return pruned, drops
	}
	out := make(llm.Conversation, 0, prefixEnd+len(pruned))
	out = append(out, cloneConversation(messages[:prefixEnd])...)
	out = append(out, pruned...)
	return out, drops
}

func compactToolOutputs(messages llm.Conversation, policy CompactionPolicy) (llm.Conversation, int) {
	if len(messages) == 0 || policy.MaxChars <= 0 {
		return messages, 0
	}
	if estimateMessages(messages) < policy.MaxChars {
		return messages, 0
	}
	out := cloneConversation(messages)
	indexes := make([]int, 0, len(out))
	for index, message := range out {
		if normalizeRole(message.Role) == "tool" && strings.TrimSpace(message.Content) != "" {
			indexes = append(indexes, index)
		}
	}
	if len(indexes) == 0 {
		return out, 0
	}
	count := 0
	cap := toolOutputSoftCap(policy.MaxChars, len(indexes))
	for _, index := range indexes {
		if estimateMessages(out) < policy.MaxChars {
			return out, count
		}
		content := out[index].Content
		if len(content) <= cap {
			continue
		}
		out[index].Content = truncateToolOutput(content, cap, policy.ToolOutputAnchor)
		if out[index].Content != content {
			count++
		}
	}
	for _, index := range indexes {
		if estimateMessages(out) < policy.MaxChars {
			return out, count
		}
		content := out[index].Content
		if content == compactedToolOutputPlaceholder || len(content) <= len(compactedToolOutputPlaceholder) {
			continue
		}
		out[index].Content = compactedToolOutputPlaceholder
		count++
		if estimateMessages(out) < policy.MaxChars {
			break
		}
	}
	return out, count
}

func compactHistoryBodies(messages llm.Conversation, maxChars int) (llm.Conversation, int) {
	if len(messages) == 0 || maxChars <= 0 {
		return messages, 0
	}
	if estimateMessages(messages) < maxChars {
		return messages, 0
	}
	out := cloneConversation(messages)
	indexes := make([]int, 0, len(out))
	last := len(out) - 1
	for index, message := range out {
		if index == last || strings.TrimSpace(message.Content) == "" || len(message.ToolCalls) > 0 || strings.TrimSpace(message.ToolCallID) != "" {
			continue
		}
		role := normalizeRole(message.Role)
		if role == "assistant" || role == "user" {
			indexes = append(indexes, index)
		}
	}
	if len(indexes) == 0 {
		return out, 0
	}
	count := 0
	cap := historyBodySoftCap(maxChars, len(indexes))
	for _, index := range indexes {
		if estimateMessages(out) < maxChars {
			return out, count
		}
		content := out[index].Content
		if len(content) <= cap {
			continue
		}
		out[index].Content = truncateHistoryBody(content, cap)
		if out[index].Content != content {
			count++
		}
	}
	for _, index := range indexes {
		if estimateMessages(out) < maxChars {
			return out, count
		}
		content := out[index].Content
		if content == compactedHistoryPlaceholder || len(content) <= len(compactedHistoryPlaceholder) {
			continue
		}
		out[index].Content = compactedHistoryPlaceholder
		count++
		if estimateMessages(out) < maxChars {
			break
		}
	}
	return out, count
}

func estimateMessages(messages llm.Conversation) int {
	return EstimateChars(EstimateInput{Messages: messages, RoleAware: true})
}

func toolOutputSoftCap(maxChars, toolCount int) int {
	cap := maxChars / 4
	if toolCount > 0 {
		if perTool := maxChars / (toolCount + 2); perTool < cap {
			cap = perTool
		}
	}
	if cap < 160 {
		cap = 160
	}
	if cap > 2048 {
		cap = 2048
	}
	return cap
}

func truncateToolOutput(content string, maxChars int, selector AnchorSelector) string {
	if maxChars <= 0 || len(content) <= maxChars {
		return content
	}
	marker := "\n...[tool output truncated for context]...\n"
	if maxChars <= len(marker) {
		return marker[:maxChars]
	}
	remaining := maxChars - len(marker)
	headChars := remaining / 2
	tailChars := remaining - headChars
	if selector != nil {
		if anchor := selector(content); anchor >= 0 && anchor < len(content) {
			minHeadChars := remaining / 6
			if minHeadChars < 24 {
				minHeadChars = 24
			}
			tailFromAnchor := len(content) - anchor
			if tailFromAnchor > 0 && tailFromAnchor <= remaining-minHeadChars {
				tailChars = tailFromAnchor
				headChars = remaining - tailChars
			}
		}
	}
	if headChars <= 0 || tailChars <= 0 {
		return content[:remaining] + marker
	}
	if headChars+tailChars >= len(content) {
		return content
	}
	return content[:headChars] + marker + content[len(content)-tailChars:]
}

func historyBodySoftCap(maxChars, count int) int {
	cap := maxChars / 6
	if count > 0 {
		if perMessage := maxChars / (count + 3); perMessage < cap {
			cap = perMessage
		}
	}
	if cap < 120 {
		cap = 120
	}
	if cap > 1024 {
		cap = 1024
	}
	return cap
}

func truncateHistoryBody(content string, maxChars int) string {
	if maxChars <= 0 || len(content) <= maxChars {
		return content
	}
	suffix := "\n...[history truncated for context]..."
	if maxChars <= len(suffix) {
		return suffix[:maxChars]
	}
	return content[:maxChars-len(suffix)] + suffix
}

var internalReasoningPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?is)<thinking>\s*.*?\s*</thinking>`),
	regexp.MustCompile(`(?is)<reasoning>\s*.*?\s*</reasoning>`),
	regexp.MustCompile(`(?is)<analysis>\s*.*?\s*</analysis>`),
	regexp.MustCompile("(?is)```(?:thinking|reasoning|analysis)\\s+.*?```"),
}

func repairProtocol(messages llm.Conversation, strict bool, diagnostic *Diagnostic) llm.Conversation {
	out := make(llm.Conversation, 0, len(messages))
	pending := map[string]bool{}
	order := make([]string, 0, 4)
	sequence := 0
	for _, message := range messages {
		role := normalizeRole(message.Role)
		content := message.Content
		toolCallID := strings.TrimSpace(message.ToolCallID)
		toolCalls := normalizeCalls(message.ToolCalls)
		if role == "assistant" && len(toolCalls) > 0 && strict {
			var count int
			toolCalls, count = synthesizeCallIDs(toolCalls, &sequence)
			diagnostic.SynthesizedToolCallIDs += count
		}
		switch role {
		case "tool":
			if toolCallID == "" && strict {
				if recovered, ok := solePendingCallID(order, pending); ok {
					toolCallID = recovered
					diagnostic.RecoveredToolResults++
				}
			}
			if toolCallID == "" || len(pending) == 0 || !pending[toolCallID] {
				role = "assistant"
				toolCallID = ""
				diagnostic.DowngradedToolResults++
			} else {
				delete(pending, toolCallID)
			}
		default:
			clear(pending)
			order = order[:0]
			if role == "assistant" && len(toolCalls) > 0 {
				for _, call := range toolCalls {
					id := strings.TrimSpace(call.ID)
					if id != "" {
						pending[id] = true
						order = append(order, id)
					}
				}
			}
		}
		if strings.TrimSpace(content) == "" && toolCallID == "" && len(toolCalls) == 0 {
			continue
		}
		out = append(out, llm.Message{Role: role, Content: content, ToolCallID: toolCallID, ToolCalls: toolCalls})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func synthesizeCallIDs(calls []llm.FunctionCall, sequence *int) ([]llm.FunctionCall, int) {
	out := cloneCalls(calls)
	count := 0
	for index := range out {
		if strings.TrimSpace(out[index].ID) != "" {
			continue
		}
		*sequence++
		out[index].ID = fmt.Sprintf("agentx_call_%d", *sequence)
		count++
	}
	return out, count
}

func solePendingCallID(order []string, pending map[string]bool) (string, bool) {
	if len(pending) != 1 {
		return "", false
	}
	for _, id := range order {
		id = strings.TrimSpace(id)
		if id != "" && pending[id] {
			return id, true
		}
	}
	for id, active := range pending {
		if active && strings.TrimSpace(id) != "" {
			return strings.TrimSpace(id), true
		}
	}
	return "", false
}

func stripInternalReasoningBlocks(messages llm.Conversation, diagnostic *Diagnostic) llm.Conversation {
	out := make(llm.Conversation, 0, len(messages))
	for _, message := range messages {
		if normalizeRole(message.Role) == "assistant" && strings.TrimSpace(message.Content) != "" {
			original := message.Content
			message.Content = compactWhitespace(stripInternalReasoningText(message.Content))
			if message.Content != compactWhitespace(original) {
				diagnostic.StrippedReasoningMsgs++
			}
		}
		if strings.TrimSpace(message.Content) == "" && strings.TrimSpace(message.ToolCallID) == "" && len(message.ToolCalls) == 0 {
			continue
		}
		out = append(out, message)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func stripInternalReasoningText(content string) string {
	stripped := content
	for _, pattern := range internalReasoningPatterns {
		stripped = pattern.ReplaceAllString(stripped, "")
	}
	return stripped
}

func compactWhitespace(content string) string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	blank := false
	for _, line := range lines {
		rightTrimmed := strings.TrimRight(line, " \t")
		if strings.TrimSpace(rightTrimmed) == "" {
			if blank {
				continue
			}
			blank = true
			out = append(out, "")
			continue
		}
		blank = false
		out = append(out, rightTrimmed)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func mergeAdjacentMessages(messages llm.Conversation, diagnostic *Diagnostic) llm.Conversation {
	out := make(llm.Conversation, 0, len(messages))
	lastInput := len(messages) - 1
	for index, message := range messages {
		if len(out) == 0 {
			out = append(out, message)
			continue
		}
		last := &out[len(out)-1]
		if canMerge(*last, message, index == lastInput) {
			last.Content = mergeContent(last.Content, message.Content)
			diagnostic.MergedMessages++
			continue
		}
		out = append(out, message)
	}
	return out
}

type historySegment struct{ start, end int }

func prune(messages llm.Conversation, max int, strict bool) (llm.Conversation, int) {
	if max <= 0 || len(messages) <= max {
		return messages, 0
	}
	if !strict {
		return cloneConversation(messages[len(messages)-max:]), 0
	}
	segments := buildHistorySegments(messages)
	remaining := max
	kept := make([]historySegment, 0, len(segments))
	for index := len(segments) - 1; index >= 0; index-- {
		segment := segments[index]
		size := segment.end - segment.start
		if size > remaining {
			break
		}
		kept = append(kept, segment)
		remaining -= size
		if remaining == 0 {
			break
		}
	}
	if len(kept) == 0 {
		return cloneConversation(messages[len(messages)-max:]), 0
	}
	out := make(llm.Conversation, 0, max)
	for index := len(kept) - 1; index >= 0; index-- {
		segment := kept[index]
		out = append(out, cloneConversation(messages[segment.start:segment.end])...)
	}
	if len(out) > max {
		out = out[len(out)-max:]
	}
	drops := len(messages) - len(out)
	if drops < 0 {
		drops = 0
	}
	return out, drops
}

func buildHistorySegments(messages llm.Conversation) []historySegment {
	out := make([]historySegment, 0, len(messages))
	for index := 0; index < len(messages); {
		start := index
		if normalizeRole(messages[index].Role) == "assistant" && len(messages[index].ToolCalls) > 0 {
			index++
			for index < len(messages) && normalizeRole(messages[index].Role) == "tool" {
				index++
			}
			out = append(out, historySegment{start: start, end: index})
			continue
		}
		index++
		out = append(out, historySegment{start: start, end: index})
	}
	return out
}

func canMerge(left, right llm.Message, rightIsLast bool) bool {
	if normalizeRole(left.Role) != normalizeRole(right.Role) || strings.TrimSpace(left.ToolCallID) != "" || strings.TrimSpace(right.ToolCallID) != "" || len(left.ToolCalls) > 0 || len(right.ToolCalls) > 0 {
		return false
	}
	role := normalizeRole(left.Role)
	if role != "system" && role != "user" && role != "assistant" {
		return false
	}
	return !(rightIsLast && role == "user")
}

func mergeContent(left, right string) string {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" {
		return right
	}
	if right == "" {
		return left
	}
	return left + "\n\n" + right
}

func normalizeRole(role string) string {
	role = strings.ToLower(strings.TrimSpace(role))
	switch role {
	case "system", "user", "assistant", "tool":
		return role
	default:
		return "assistant"
	}
}

func normalizeCalls(calls []llm.FunctionCall) []llm.FunctionCall {
	out := make([]llm.FunctionCall, 0, len(calls))
	for _, call := range calls {
		name := strings.TrimSpace(call.Name)
		arguments := strings.TrimSpace(call.Arguments)
		id := strings.TrimSpace(call.ID)
		if id == "" && name == "" && arguments == "" {
			continue
		}
		callType := strings.TrimSpace(call.Type)
		if callType == "" {
			callType = "function"
		}
		out = append(out, llm.FunctionCall{ID: id, Type: callType, Name: name, Arguments: arguments})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cloneConversation(messages llm.Conversation) llm.Conversation {
	if len(messages) == 0 {
		return nil
	}
	out := make(llm.Conversation, len(messages))
	copy(out, messages)
	for index := range out {
		out[index].ToolCalls = cloneCalls(out[index].ToolCalls)
	}
	return out
}

func cloneCalls(calls []llm.FunctionCall) []llm.FunctionCall {
	if len(calls) == 0 {
		return nil
	}
	return append([]llm.FunctionCall(nil), calls...)
}
