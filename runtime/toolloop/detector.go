package toolloop

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
)

const (
	LoopKindRepeat     = "repeat"
	LoopKindPingPong   = "ping_pong"
	LoopKindNoProgress = "no_progress"
	LoopKindReplay     = "successful_replay"
)

// Call is the portable projection of a model-requested tool call.
type Call struct {
	Name      string
	Arguments string
}

// RunObservation is the portable projection of a completed tool call.
type RunObservation struct {
	Name       string
	Output     string
	Failed     bool
	ErrorClass string
}

// LoopSignal describes a detected loop or successful replay.
type LoopSignal struct {
	Kind            string
	Round           int
	CallSignature   string
	ResultSignature string
	Count           int
}

// LoopDetectorConfig supplies host-owned policy values to the detector.
type LoopDetectorConfig struct {
	Enabled             bool
	RepeatThreshold     int
	PingPongThreshold   int
	NoProgressThreshold int
}

type loopFrame struct {
	Round           int
	CallSignature   string
	ResultSignature string
	Successful      bool
}

// LoopDetector detects repeated, ping-pong, no-progress, and replay patterns.
type LoopDetector struct {
	enabled             bool
	repeatThreshold     int
	pingPongThreshold   int
	noProgressThreshold int
	maxHistory          int
	history             []loopFrame
}

// NewLoopDetector constructs a detector from already resolved host policy.
func NewLoopDetector(config LoopDetectorConfig) *LoopDetector {
	repeat := max(config.RepeatThreshold, 1)
	pingPong := max(config.PingPongThreshold, 2)
	if pingPong%2 != 0 {
		pingPong++
	}
	noProgress := max(config.NoProgressThreshold, 1)
	maxHistory := max(repeat, pingPong, noProgress, 4)
	return &LoopDetector{
		enabled:             config.Enabled,
		repeatThreshold:     repeat,
		pingPongThreshold:   pingPong,
		noProgressThreshold: noProgress,
		maxHistory:          maxHistory + 2,
		history:             make([]loopFrame, 0, maxHistory+2),
	}
}

// Observe records one round and reports a terminal loop signal when present.
func (d *LoopDetector) Observe(round int, calls []Call, runs []RunObservation) (LoopSignal, bool) {
	if d == nil || !d.enabled || len(calls) == 0 {
		return LoopSignal{}, false
	}
	callSignature := buildCallSignature(calls)
	if callSignature == "" {
		return LoopSignal{}, false
	}
	frame := loopFrame{
		Round:           round,
		CallSignature:   callSignature,
		ResultSignature: buildResultSignature(runs),
		Successful:      runsSuccessful(runs),
	}
	d.history = append(d.history, frame)
	if overflow := len(d.history) - d.maxHistory; overflow > 0 {
		d.history = append([]loopFrame(nil), d.history[overflow:]...)
	}
	if signal, ok := d.detectNoProgress(); ok {
		return signal, true
	}
	if signal, ok := d.detectRepeat(); ok {
		return signal, true
	}
	if signal, ok := d.detectPingPong(); ok {
		return signal, true
	}
	return LoopSignal{}, false
}

// ShouldSuppressReplay reports an immediately repeated successful call.
func (d *LoopDetector) ShouldSuppressReplay(calls []Call) (LoopSignal, bool) {
	if d == nil || !d.enabled || len(calls) == 0 || len(d.history) == 0 {
		return LoopSignal{}, false
	}
	callSignature := buildCallSignature(calls)
	if callSignature == "" {
		return LoopSignal{}, false
	}
	last := d.history[len(d.history)-1]
	if last.CallSignature != callSignature ||
		!last.Successful ||
		strings.TrimSpace(last.ResultSignature) == "" {
		return LoopSignal{}, false
	}
	return LoopSignal{
		Kind:            LoopKindReplay,
		Round:           last.Round,
		CallSignature:   last.CallSignature,
		ResultSignature: last.ResultSignature,
		Count:           2,
	}, true
}

func (d *LoopDetector) detectRepeat() (LoopSignal, bool) {
	if d == nil || len(d.history) < d.repeatThreshold {
		return LoopSignal{}, false
	}
	last := d.history[len(d.history)-1]
	count := 1
	for index := len(d.history) - 2; index >= 0; index-- {
		if d.history[index].CallSignature != last.CallSignature {
			break
		}
		count++
	}
	if count < d.repeatThreshold {
		return LoopSignal{}, false
	}
	return LoopSignal{
		Kind:            LoopKindRepeat,
		Round:           last.Round,
		CallSignature:   last.CallSignature,
		ResultSignature: last.ResultSignature,
		Count:           count,
	}, true
}

func (d *LoopDetector) detectNoProgress() (LoopSignal, bool) {
	if d == nil || len(d.history) < d.noProgressThreshold {
		return LoopSignal{}, false
	}
	last := d.history[len(d.history)-1]
	if strings.TrimSpace(last.ResultSignature) == "" {
		return LoopSignal{}, false
	}
	count := 1
	for index := len(d.history) - 2; index >= 0; index-- {
		if d.history[index].CallSignature != last.CallSignature ||
			d.history[index].ResultSignature != last.ResultSignature {
			break
		}
		count++
	}
	if count < d.noProgressThreshold {
		return LoopSignal{}, false
	}
	return LoopSignal{
		Kind:            LoopKindNoProgress,
		Round:           last.Round,
		CallSignature:   last.CallSignature,
		ResultSignature: last.ResultSignature,
		Count:           count,
	}, true
}

func (d *LoopDetector) detectPingPong() (LoopSignal, bool) {
	if d == nil || len(d.history) < d.pingPongThreshold {
		return LoopSignal{}, false
	}
	sequence := d.history[len(d.history)-d.pingPongThreshold:]
	first := strings.TrimSpace(sequence[0].CallSignature)
	second := strings.TrimSpace(sequence[1].CallSignature)
	if first == "" || second == "" || first == second {
		return LoopSignal{}, false
	}
	for index := range sequence {
		expected := first
		if index%2 == 1 {
			expected = second
		}
		if sequence[index].CallSignature != expected {
			return LoopSignal{}, false
		}
	}
	last := sequence[len(sequence)-1]
	return LoopSignal{
		Kind:            LoopKindPingPong,
		Round:           last.Round,
		CallSignature:   last.CallSignature,
		ResultSignature: last.ResultSignature,
		Count:           len(sequence),
	}, true
}

func buildCallSignature(calls []Call) string {
	parts := make([]string, 0, len(calls))
	for _, call := range calls {
		name := normalizeText(call.Name)
		if name == "" {
			name = "unknown"
		}
		parts = append(parts, fmt.Sprintf("%s(%s)", name, normalizeArguments(call.Arguments)))
	}
	return strings.Join(parts, " -> ")
}

func buildResultSignature(runs []RunObservation) string {
	parts := make([]string, 0, len(runs))
	for _, run := range runs {
		tool := normalizeText(run.Name)
		if tool == "" {
			tool = "unknown"
		}
		if run.Failed {
			parts = append(parts, tool+":err="+strings.TrimSpace(run.ErrorClass))
			continue
		}
		parts = append(parts, tool+":ok="+shortDigest(normalizeText(run.Output)))
	}
	return strings.Join(parts, "|")
}

func normalizeArguments(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "{}"
	}
	var payload any
	if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
		if normalized, marshalErr := json.Marshal(payload); marshalErr == nil {
			return string(normalized)
		}
	}
	return normalizeText(trimmed)
}

func normalizeText(raw string) string {
	return strings.ToLower(strings.TrimSpace(strings.Join(strings.Fields(raw), " ")))
}

func shortDigest(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "empty"
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(raw))
	return fmt.Sprintf("%016x", hasher.Sum64())
}

func runsSuccessful(runs []RunObservation) bool {
	if len(runs) == 0 {
		return false
	}
	for _, run := range runs {
		if run.Failed {
			return false
		}
	}
	return true
}
