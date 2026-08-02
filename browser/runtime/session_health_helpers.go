package browserruntime

import "strings"

func sharedSessionBrowserUniqueActions(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func sharedSessionBrowserWithoutAction(items []string, action string) []string {
	if len(items) == 0 || strings.TrimSpace(action) == "" {
		return items
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(action)) {
			continue
		}
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func sharedSessionBrowserFollowPolicyPriority(state string) int {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "popup_storm_review_required":
		return 3
	case "redirect_review_required":
		return 2
	case "popup_review_required":
		return 1
	default:
		return 0
	}
}

// SharedSessionBrowserSelectedFollowPolicyState returns the highest-priority
// follow-policy blocker across the scoped route inputs.
func SharedSessionBrowserSelectedFollowPolicyState(routes []SharedSessionBrowserRouteCoordinationInput) string {
	selectedState := ""
	for _, route := range routes {
		state := strings.TrimSpace(route.FollowPolicyState)
		if sharedSessionBrowserFollowPolicyPriority(state) > sharedSessionBrowserFollowPolicyPriority(selectedState) {
			selectedState = state
		}
	}
	return selectedState
}
