package tools

import "strings"

func browserTabByIndex(tabs []BrowserTab, index int) BrowserTab {
	for _, tab := range tabs {
		if tab.Index == index {
			return tab
		}
	}
	return BrowserTab{Index: index}
}

func browserRememberDecisionRequiresPopupReview(decision string, ready bool) bool {
	return strings.EqualFold(strings.TrimSpace(decision), "session_target_popup_review_required") && !ready
}

func browserRememberDecisionSurfacesTopLevelReview(decision string, ready bool) bool {
	normalized := strings.TrimSpace(decision)
	switch {
	case strings.EqualFold(normalized, "session_target_popup_review_required") && !ready:
		return true
	case strings.EqualFold(normalized, "session_target_popup_review_confirmed") && ready:
		return true
	default:
		return false
	}
}
