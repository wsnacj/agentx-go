package tools

import "strings"

const browserPostNavigationSnapshotMaxElements = 80

type browserPostNavigationSnapshotRecommendation struct {
	Recommendation string `json:"recommendation"`
	Tool           string `json:"tool"`
	Kind           string `json:"kind"`
	Compact        bool   `json:"compact"`
	MaxElements    int    `json:"max_elements,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

type browserBotDetectionWarning struct {
	WarningCode       string `json:"warning_code"`
	Severity          string `json:"severity,omitempty"`
	Source            string `json:"source,omitempty"`
	Signal            string `json:"signal,omitempty"`
	Reason            string `json:"reason,omitempty"`
	RecommendedAction string `json:"recommended_action,omitempty"`
}

func browserPostNavigationSnapshotRecommendationForNavigate(status, finalURL string) *browserPostNavigationSnapshotRecommendation {
	switch strings.TrimSpace(status) {
	case "navigated", "opened", "ok":
	default:
		return nil
	}
	if strings.TrimSpace(finalURL) == "" {
		return nil
	}
	return &browserPostNavigationSnapshotRecommendation{
		Recommendation: "take_compact_snapshot",
		Tool:           "browser_act",
		Kind:           "snapshot",
		Compact:        true,
		MaxElements:    browserPostNavigationSnapshotMaxElements,
		Reason:         "post_navigation_context",
	}
}

func browserPostNavigationSnapshotRecommendationForAct(kind, status, finalURL string) *browserPostNavigationSnapshotRecommendation {
	if strings.TrimSpace(kind) != "navigate" {
		return nil
	}
	return browserPostNavigationSnapshotRecommendationForNavigate(status, finalURL)
}

type browserBotDetectionSignal struct {
	Source string
	Signal string
}

func browserBotDetectionWarningFromNavigation(finalURL, title, note string) *browserBotDetectionWarning {
	signal, ok := browserBotDetectionSignalFromNavigation(finalURL, title, note)
	if !ok {
		return nil
	}
	return &browserBotDetectionWarning{
		WarningCode:       "browser_bot_detection_challenge",
		Severity:          "warning",
		Source:            signal.Source,
		Signal:            signal.Signal,
		Reason:            "navigation_result_exposed_challenge_signal",
		RecommendedAction: "use_snapshot_or_manual_review_before_retrying_automation",
	}
}

func browserBotDetectionSignalFromNavigation(finalURL, title, note string) (browserBotDetectionSignal, bool) {
	candidates := []struct {
		source string
		text   string
	}{
		{source: "note", text: note},
		{source: "title", text: title},
		{source: "final_url", text: finalURL},
	}
	for _, candidate := range candidates {
		text := strings.ToLower(strings.TrimSpace(candidate.text))
		if text == "" {
			continue
		}
		if strings.Contains(text, "captcha") || strings.Contains(text, "recaptcha") || strings.Contains(text, "hcaptcha") {
			return browserBotDetectionSignal{Source: candidate.source, Signal: "captcha"}, true
		}
		if strings.Contains(text, "verify you are human") || strings.Contains(text, "are you human") || strings.Contains(text, "checking your browser") || strings.Contains(text, "checking if the site connection is secure") {
			return browserBotDetectionSignal{Source: candidate.source, Signal: "human_verification"}, true
		}
		if strings.Contains(text, "bot detection") || strings.Contains(text, "automated traffic") || strings.Contains(text, "unusual traffic") || strings.Contains(text, "cloudflare challenge") {
			return browserBotDetectionSignal{Source: candidate.source, Signal: "bot_challenge"}, true
		}
	}
	return browserBotDetectionSignal{}, false
}
