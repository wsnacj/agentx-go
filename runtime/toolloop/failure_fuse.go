package toolloop

import "strings"

const invalidArgumentsFuseThreshold = 2

// FailureFuseConfig supplies host-owned failure policy to the fuse.
type FailureFuseConfig struct {
	Enabled   bool
	Threshold int
}

// FailureObservation is the portable projection of one tool result.
type FailureObservation struct {
	Tool       string
	Failed     bool
	ErrorClass string
}

// FailureSignal describes a consecutive tool failure fuse trigger.
type FailureSignal struct {
	Round      int
	Tool       string
	Count      int
	ErrorClass string
}

// FailureFuse tracks consecutive failures for the same normalized tool.
type FailureFuse struct {
	enabled   bool
	threshold int

	currentTool       string
	currentCount      int
	currentErrorClass string
}

// NewFailureFuse constructs a fuse from already resolved host policy.
func NewFailureFuse(config FailureFuseConfig) *FailureFuse {
	return &FailureFuse{
		enabled:   config.Enabled,
		threshold: max(config.Threshold, 1),
	}
}

// Observe records tool results and reports the first fuse trigger.
func (f *FailureFuse) Observe(round int, observations []FailureObservation) (FailureSignal, bool) {
	if f == nil || !f.enabled || len(observations) == 0 {
		return FailureSignal{}, false
	}
	for _, observation := range observations {
		tool := strings.TrimSpace(observation.Tool)
		if tool == "" {
			tool = "unknown"
		}
		if !observation.Failed {
			f.reset()
			continue
		}
		errorClass := strings.TrimSpace(observation.ErrorClass)
		if errorClass == "" {
			errorClass = "other"
		}
		if f.currentTool == tool {
			f.currentCount++
		} else {
			f.currentTool = tool
			f.currentCount = 1
		}
		f.currentErrorClass = errorClass
		effectiveThreshold := f.threshold
		if errorClass == "invalid_args" && effectiveThreshold > invalidArgumentsFuseThreshold {
			effectiveThreshold = invalidArgumentsFuseThreshold
		}
		if f.currentCount >= effectiveThreshold {
			return FailureSignal{
				Round:      round,
				Tool:       f.currentTool,
				Count:      f.currentCount,
				ErrorClass: f.currentErrorClass,
			}, true
		}
	}
	return FailureSignal{}, false
}

func (f *FailureFuse) reset() {
	if f == nil {
		return
	}
	f.currentTool = ""
	f.currentCount = 0
	f.currentErrorClass = ""
}
