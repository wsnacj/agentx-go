package runstore

import (
	"encoding/json"
	"sort"
	"strings"

	workflownodeexec "github.com/wsnacj/agentx-go/runtime/workflow/nodeexec"
)

// NodeDelegatedExecutionProjection is the compatibility name for the canonical
// portable delegated execution contract.
type NodeDelegatedExecutionProjection = workflownodeexec.DelegatedExecution

// NodeTerminationProjection is the compatibility name for the canonical
// portable termination contract.
type NodeTerminationProjection = workflownodeexec.Termination

type nodeExecutionContractDiff struct {
	ChangedFields []string `json:"changed_fields,omitempty"`
}

// NodeExecutionProjection is the compatibility name for the canonical
// recursive portable node execution projection.
type NodeExecutionProjection = workflownodeexec.NodeExecutionProjection

func CloneNodeExecutionProjection(in *NodeExecutionProjection) *NodeExecutionProjection {
	if in == nil {
		return nil
	}
	out := *in
	out.ExecutionContractDiff = append([]string(nil), in.ExecutionContractDiff...)
	out.Termination = CloneNodeTerminationProjection(in.Termination)
	out.DelegatedExecution = CloneNodeDelegatedExecutionProjection(in.DelegatedExecution)
	if len(in.ChildNodeExecutions) > 0 {
		out.ChildNodeExecutions = CloneNodeExecutionProjections(in.ChildNodeExecutions)
	}
	return &out
}

// NodeDelegatedRoundProjection is the compatibility name for the canonical
// portable delegated round contract.
type NodeDelegatedRoundProjection = workflownodeexec.DelegatedRound

func (n NodeExecution) Projection() *NodeExecutionProjection {
	return &NodeExecutionProjection{
		NodeExecID:            n.NodeExecID,
		RunID:                 n.RunID,
		BranchID:              n.BranchID,
		ParentNodeExecID:      n.ParentNodeExecID,
		NodeID:                n.NodeID,
		Kind:                  n.Kind,
		Status:                n.Status,
		Attempt:               n.Attempt,
		InputStateRef:         n.InputStateRef,
		OutputStateRef:        n.OutputStateRef,
		ExecutionContractID:   n.ExecutionContractID,
		ExecutionContractDiff: n.ExecutionContractDiff(),
		Termination:           CloneNodeTerminationProjection(n.TerminationProjection()),
		DelegatedExecution:    CloneNodeDelegatedExecutionProjection(n.DelegatedExecutionProjection()),
		StartedAt:             n.StartedAt,
		FinishedAt:            n.FinishedAt,
	}
}

func (n NodeExecution) ExecutionContractDiff() []string {
	return NodeExecutionContractDiffFromJSON(n.ExecutionContractDiffJSON)
}

func (n NodeExecution) TerminationProjection() *NodeTerminationProjection {
	return NodeTerminationProjectionFromJSON(n.TerminationJSON)
}

func (n NodeExecution) DelegatedExecutionProjection() *NodeDelegatedExecutionProjection {
	return NodeDelegatedExecutionProjectionFromJSON(n.DelegatedExecutionJSON)
}

func NodeExecutionContractDiffFromJSON(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	var payload nodeExecutionContractDiff
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return nil
	}
	out := make([]string, 0, len(payload.ChangedFields))
	for _, item := range payload.ChangedFields {
		if strings.TrimSpace(item) == "" {
			continue
		}
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func NodeTerminationProjectionFromJSON(raw string) *NodeTerminationProjection {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	var payload NodeTerminationProjection
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return nil
	}
	if strings.TrimSpace(payload.Kind) == "" &&
		strings.TrimSpace(payload.CheckpointStage) == "" &&
		strings.TrimSpace(payload.CheckpointError) == "" &&
		strings.TrimSpace(payload.EventName) == "" &&
		strings.TrimSpace(payload.EventStatus) == "" &&
		!payload.ReplyPersisted {
		return nil
	}
	return &payload
}

func NodeDelegatedExecutionProjectionFromJSON(raw string) *NodeDelegatedExecutionProjection {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	var payload NodeDelegatedExecutionProjection
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return nil
	}
	return CloneNodeDelegatedExecutionProjection(&payload)
}

func CloneNodeTerminationProjection(in *NodeTerminationProjection) *NodeTerminationProjection {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func CloneNodeDelegatedExecutionProjection(in *NodeDelegatedExecutionProjection) *NodeDelegatedExecutionProjection {
	if in == nil {
		return nil
	}
	out := *in
	out.Rounds = nil
	for _, round := range in.Rounds {
		if nodeDelegatedRoundProjectionEmpty(round) {
			continue
		}
		out.Rounds = append(out.Rounds, round)
	}
	if nodeDelegatedExecutionProjectionEmpty(out) {
		return nil
	}
	return &out
}

func nodeDelegatedRoundProjectionEmpty(in NodeDelegatedRoundProjection) bool {
	return strings.TrimSpace(in.NodeExecID) == "" &&
		in.Round == 0 &&
		strings.TrimSpace(in.OutcomeKind) == "" &&
		strings.TrimSpace(in.StopReason) == "" &&
		in.ToolCalls == 0 &&
		in.ToolRuns == 0
}

func nodeDelegatedExecutionProjectionEmpty(in NodeDelegatedExecutionProjection) bool {
	return strings.TrimSpace(in.Driver) == "" &&
		strings.TrimSpace(in.OutcomeKind) == "" &&
		in.RoundCount == 0 &&
		in.ToolCalls == 0 &&
		len(in.Rounds) == 0
}

func NodeDelegatedExecutionLastStopReason(in *NodeDelegatedExecutionProjection) string {
	if in == nil || len(in.Rounds) == 0 {
		return ""
	}
	for idx := len(in.Rounds) - 1; idx >= 0; idx-- {
		stopReason := in.Rounds[idx].StopReason
		if strings.TrimSpace(stopReason) != "" {
			return stopReason
		}
	}
	return ""
}

func NodeDelegatedExecutionProjectionFromChildNodeExecutions(items []NodeExecutionProjection) *NodeDelegatedExecutionProjection {
	flattened := flattenNodeExecutionProjectionTree(items)
	if len(flattened) == 0 {
		return nil
	}
	rounds := make([]NodeDelegatedRoundProjection, 0, len(flattened))
	totalToolCalls := 0
	driver := ""
	maxRound := 0
	for idx, item := range flattened {
		round := delegatedRoundProjectionFromNodeExecution(item, idx+1)
		if round.Round <= 0 {
			continue
		}
		if driver == "" && item.DelegatedExecution != nil && strings.TrimSpace(item.DelegatedExecution.Driver) != "" {
			driver = item.DelegatedExecution.Driver
		}
		if round.Round > maxRound {
			maxRound = round.Round
		}
		totalToolCalls += round.ToolCalls
		rounds = append(rounds, round)
	}
	if len(rounds) == 0 {
		return nil
	}
	out := &NodeDelegatedExecutionProjection{
		Driver:      driver,
		OutcomeKind: delegatedOutcomeKindFromRounds(rounds),
		RoundCount:  maxRound,
		ToolCalls:   totalToolCalls,
		Rounds:      rounds,
	}
	if out.RoundCount <= 0 {
		out.RoundCount = len(rounds)
	}
	return out
}

func SelectTerminalNodeExecutionProjection(items []NodeExecutionProjection) *NodeExecutionProjection {
	if len(items) == 0 {
		return nil
	}
	candidates := items
	topLevel := make([]NodeExecutionProjection, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.ParentNodeExecID) == "" {
			topLevel = append(topLevel, item)
		}
	}
	if len(topLevel) > 0 {
		candidates = topLevel
	}
	best := candidates[0]
	for idx := 1; idx < len(candidates); idx++ {
		candidate := candidates[idx]
		if candidate.FinishedAt != best.FinishedAt {
			if candidate.FinishedAt > best.FinishedAt {
				best = candidate
			}
			continue
		}
		if candidate.StartedAt != best.StartedAt {
			if candidate.StartedAt > best.StartedAt {
				best = candidate
			}
			continue
		}
		if candidate.Attempt != best.Attempt {
			if candidate.Attempt > best.Attempt {
				best = candidate
			}
			continue
		}
		if candidate.NodeExecID > best.NodeExecID {
			best = candidate
		}
	}
	return CloneNodeExecutionProjection(&best)
}

func delegatedRoundProjectionFromNodeExecution(item NodeExecutionProjection, defaultRound int) NodeDelegatedRoundProjection {
	round := NodeDelegatedRoundProjection{}
	if item.DelegatedExecution != nil && len(item.DelegatedExecution.Rounds) > 0 {
		last := item.DelegatedExecution.Rounds[len(item.DelegatedExecution.Rounds)-1]
		round.NodeExecID = last.NodeExecID
		round.Round = last.Round
		round.OutcomeKind = last.OutcomeKind
		round.StopReason = last.StopReason
		round.ToolCalls = last.ToolCalls
		round.ToolRuns = last.ToolRuns
	}
	if round.NodeExecID == "" {
		round.NodeExecID = item.NodeExecID
	}
	if round.Round <= 0 {
		round.Round = nodeExecutionProjectionRoundNumber(item)
	}
	if round.Round <= 0 {
		round.Round = defaultRound
	}
	if round.OutcomeKind == "" {
		round.OutcomeKind = nodeExecutionProjectionOutcomeKind(item)
	}
	if round.StopReason == "" {
		if stopReason := nodeExecutionProjectionStopReason(item); stopReason != "" {
			round.StopReason = stopReason
		}
	}
	if round.ToolCalls <= 0 && item.DelegatedExecution != nil {
		round.ToolCalls = item.DelegatedExecution.ToolCalls
	}
	return round
}

func delegatedOutcomeKindFromRounds(rounds []NodeDelegatedRoundProjection) string {
	if len(rounds) == 0 {
		return ""
	}
	last := rounds[len(rounds)-1].OutcomeKind
	switch strings.TrimSpace(last) {
	case "terminated", "error", "completed":
		return last
	}
	for idx := len(rounds) - 1; idx >= 0; idx-- {
		raw := rounds[idx].OutcomeKind
		switch strings.TrimSpace(raw) {
		case "terminated", "error", "completed":
			return raw
		}
	}
	if strings.TrimSpace(last) == "" {
		return ""
	}
	return last
}

func nodeExecutionProjectionRoundNumber(item NodeExecutionProjection) int {
	if item.DelegatedExecution != nil {
		for idx := len(item.DelegatedExecution.Rounds) - 1; idx >= 0; idx-- {
			if item.DelegatedExecution.Rounds[idx].Round > 0 {
				return item.DelegatedExecution.Rounds[idx].Round
			}
		}
		if item.DelegatedExecution.RoundCount > 0 {
			return item.DelegatedExecution.RoundCount
		}
	}
	nodeID := item.NodeID
	roundIdx := strings.LastIndex(nodeID, "round:")
	if roundIdx < 0 {
		return 0
	}
	value := nodeID[roundIdx+len("round:"):]
	if value == "" {
		return 0
	}
	round := 0
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			break
		}
		round = round*10 + int(ch-'0')
	}
	return round
}

func nodeExecutionProjectionOutcomeKind(item NodeExecutionProjection) string {
	if item.DelegatedExecution != nil {
		for idx := len(item.DelegatedExecution.Rounds) - 1; idx >= 0; idx-- {
			if outcome := item.DelegatedExecution.Rounds[idx].OutcomeKind; strings.TrimSpace(outcome) != "" {
				return outcome
			}
		}
		if outcome := item.DelegatedExecution.OutcomeKind; strings.TrimSpace(outcome) != "" {
			return outcome
		}
	}
	switch strings.ToLower(strings.TrimSpace(item.Status)) {
	case "failed":
		return "error"
	case "incomplete", "canceled":
		return "terminated"
	case "completed":
		return "completed"
	default:
		return strings.ToLower(strings.TrimSpace(item.Status))
	}
}

func nodeExecutionProjectionStopReason(item NodeExecutionProjection) string {
	if item.DelegatedExecution != nil {
		if stopReason := NodeDelegatedExecutionLastStopReason(item.DelegatedExecution); stopReason != "" {
			return stopReason
		}
	}
	if item.Termination != nil {
		if kind := item.Termination.Kind; strings.TrimSpace(kind) != "" {
			return kind
		}
	}
	return ""
}

func flattenNodeExecutionProjectionTree(items []NodeExecutionProjection) []NodeExecutionProjection {
	if len(items) == 0 {
		return nil
	}
	out := make([]NodeExecutionProjection, 0, len(items))
	for _, item := range items {
		cloned := CloneNodeExecutionProjection(&item)
		if cloned == nil {
			continue
		}
		cloned.ChildNodeExecutions = nil
		out = append(out, *cloned)
		if len(item.ChildNodeExecutions) > 0 {
			out = append(out, flattenNodeExecutionProjectionTree(item.ChildNodeExecutions)...)
		}
	}
	if len(out) == 0 {
		return nil
	}
	sort.SliceStable(out, func(i, j int) bool {
		leftRound := nodeExecutionProjectionRoundNumber(out[i])
		rightRound := nodeExecutionProjectionRoundNumber(out[j])
		if leftRound > 0 || rightRound > 0 {
			if leftRound == 0 {
				return false
			}
			if rightRound == 0 {
				return true
			}
			if leftRound != rightRound {
				return leftRound < rightRound
			}
		}
		if out[i].StartedAt != out[j].StartedAt {
			return out[i].StartedAt < out[j].StartedAt
		}
		if out[i].FinishedAt != out[j].FinishedAt {
			return out[i].FinishedAt < out[j].FinishedAt
		}
		return out[i].NodeExecID < out[j].NodeExecID
	})
	return out
}

func SelectChildNodeExecutionProjections(items []NodeExecutionProjection, parentNodeExecID string) []NodeExecutionProjection {
	if strings.TrimSpace(parentNodeExecID) == "" || len(items) == 0 {
		return nil
	}
	out := make([]NodeExecutionProjection, 0, len(items))
	for _, item := range items {
		if item.ParentNodeExecID != parentNodeExecID {
			continue
		}
		cloned := CloneNodeExecutionProjection(&item)
		if cloned == nil {
			continue
		}
		out = append(out, *cloned)
	}
	if len(out) == 0 {
		return nil
	}
	sort.SliceStable(out, func(i, j int) bool {
		leftRound := nodeExecutionProjectionRoundOrder(out[i])
		rightRound := nodeExecutionProjectionRoundOrder(out[j])
		if leftRound != rightRound {
			return leftRound < rightRound
		}
		if out[i].StartedAt != out[j].StartedAt {
			return out[i].StartedAt < out[j].StartedAt
		}
		if out[i].FinishedAt != out[j].FinishedAt {
			return out[i].FinishedAt < out[j].FinishedAt
		}
		return out[i].NodeExecID < out[j].NodeExecID
	})
	return out
}

func CloneNodeExecutionProjections(in []NodeExecutionProjection) []NodeExecutionProjection {
	if len(in) == 0 {
		return nil
	}
	out := make([]NodeExecutionProjection, 0, len(in))
	for _, item := range in {
		cloned := CloneNodeExecutionProjection(&item)
		if cloned == nil {
			continue
		}
		out = append(out, *cloned)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func AttachChildNodeExecutionProjections(items []NodeExecutionProjection) []NodeExecutionProjection {
	if len(items) == 0 {
		return nil
	}
	cloned := CloneNodeExecutionProjections(items)
	if len(cloned) == 0 {
		return nil
	}
	childrenByParent := make(map[string][]NodeExecutionProjection)
	topLevel := make([]NodeExecutionProjection, 0, len(cloned))
	for _, item := range cloned {
		parentID := item.ParentNodeExecID
		if strings.TrimSpace(parentID) == "" {
			topLevel = append(topLevel, item)
			continue
		}
		childrenByParent[parentID] = append(childrenByParent[parentID], item)
	}
	for idx := range topLevel {
		topLevel[idx].ChildNodeExecutions = attachChildNodeExecutionProjections(childrenByParent, topLevel[idx].NodeExecID)
	}
	if len(topLevel) == 0 {
		return nil
	}
	return topLevel
}

func FindNodeExecutionProjection(items []NodeExecutionProjection, nodeExecID string) *NodeExecutionProjection {
	if strings.TrimSpace(nodeExecID) == "" || len(items) == 0 {
		return nil
	}
	for _, item := range items {
		if item.NodeExecID == nodeExecID {
			return CloneNodeExecutionProjection(&item)
		}
		if found := FindNodeExecutionProjection(item.ChildNodeExecutions, nodeExecID); found != nil {
			return found
		}
	}
	return nil
}

func attachChildNodeExecutionProjections(childrenByParent map[string][]NodeExecutionProjection, parentNodeExecID string) []NodeExecutionProjection {
	if strings.TrimSpace(parentNodeExecID) == "" {
		return nil
	}
	children := childrenByParent[parentNodeExecID]
	if len(children) == 0 {
		return nil
	}
	sort.SliceStable(children, func(i, j int) bool {
		leftRound := nodeExecutionProjectionRoundOrder(children[i])
		rightRound := nodeExecutionProjectionRoundOrder(children[j])
		if leftRound != rightRound {
			return leftRound < rightRound
		}
		if children[i].StartedAt != children[j].StartedAt {
			return children[i].StartedAt < children[j].StartedAt
		}
		if children[i].FinishedAt != children[j].FinishedAt {
			return children[i].FinishedAt < children[j].FinishedAt
		}
		return children[i].NodeExecID < children[j].NodeExecID
	})
	out := make([]NodeExecutionProjection, 0, len(children))
	for _, child := range children {
		child.ChildNodeExecutions = attachChildNodeExecutionProjections(childrenByParent, child.NodeExecID)
		out = append(out, child)
	}
	return out
}

func nodeExecutionProjectionRoundOrder(item NodeExecutionProjection) int {
	if item.DelegatedExecution != nil && len(item.DelegatedExecution.Rounds) > 0 && item.DelegatedExecution.Rounds[0].Round > 0 {
		return item.DelegatedExecution.Rounds[0].Round
	}
	nodeID := item.NodeID
	if nodeID == "" {
		return 0
	}
	idx := strings.LastIndex(nodeID, "round:")
	if idx < 0 {
		return 0
	}
	value := nodeID[idx+len("round:"):]
	round := 0
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			break
		}
		round = round*10 + int(ch-'0')
	}
	return round
}
