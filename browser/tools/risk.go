package tools

import "strings"

// RiskLevel classifies the side-effect risk communicated by browser tool metadata.
type RiskLevel int

const (
	RiskUnknown RiskLevel = iota
	RiskLow
	RiskMedium
	RiskHigh
)

func (r RiskLevel) String() string {
	switch r {
	case RiskLow:
		return "low"
	case RiskMedium:
		return "medium"
	case RiskHigh:
		return "high"
	default:
		return "unknown"
	}
}

func parseRiskLevel(raw string) RiskLevel {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "low":
		return RiskLow
	case "medium":
		return RiskMedium
	case "high":
		return RiskHigh
	default:
		return RiskUnknown
	}
}
