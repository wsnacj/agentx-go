package tools

import (
	"fmt"
	"strings"
)

func BrowserDoctorStatusText(doctor *BrowserDoctorSummary) string {
	if doctor == nil {
		return ""
	}
	routeStatus := "-"
	launchStatus := "-"
	if doctor.Route != nil && strings.TrimSpace(doctor.Route.Status) != "" {
		routeStatus = strings.TrimSpace(doctor.Route.Status)
	}
	if doctor.Launch != nil && strings.TrimSpace(doctor.Launch.Status) != "" {
		launchStatus = strings.TrimSpace(doctor.Launch.Status)
	}
	return fmt.Sprintf(
		"browser doctor: status=%s ready=%t route=%s launch=%s",
		strings.TrimSpace(doctor.Status),
		doctor.Ready,
		routeStatus,
		launchStatus,
	)
}

func BrowserDoctorRouteText(route *BrowserDoctorRouteSummary) string {
	if route == nil || strings.TrimSpace(route.Summary) == "" {
		return ""
	}
	return fmt.Sprintf(
		"- browser route [%s/%s]: %s",
		strings.TrimSpace(route.Status),
		strings.TrimSpace(route.Code),
		strings.TrimSpace(route.Summary),
	)
}

func BrowserDoctorLaunchText(launch *BrowserDoctorLaunchSummary) string {
	if launch == nil || strings.TrimSpace(launch.Summary) == "" {
		return ""
	}
	return fmt.Sprintf(
		"- browser launch [%s/%s]: %s",
		strings.TrimSpace(launch.Status),
		strings.TrimSpace(launch.Code),
		strings.TrimSpace(launch.Summary),
	)
}

func BrowserDoctorLaunchDetailText(launch *BrowserDoctorLaunchSummary) string {
	if launch == nil {
		return ""
	}
	items := make([]string, 0, 6)
	if state := strings.TrimSpace(launch.BootstrapState); state != "" {
		items = append(items, "bootstrap="+state)
	}
	if code := strings.TrimSpace(launch.BootstrapErrorCode); code != "" {
		items = append(items, "code="+code)
	}
	if node := strings.TrimSpace(launch.NodeVersion); node != "" {
		items = append(items, "node="+node)
	}
	if pkg := strings.TrimSpace(launch.PlaywrightPackage); pkg != "" {
		items = append(items, "playwright="+firstNonEmpty(strings.TrimSpace(launch.PlaywrightPackageVersion), pkg))
	}
	if launch.RuntimeBaselineReady != nil && !*launch.RuntimeBaselineReady {
		items = append(items, "baseline="+firstNonEmpty(strings.TrimSpace(launch.RuntimeBaselineBlockReason), "not_ready"))
	}
	if launch.SelectedLaunchExecutableReady != nil && !*launch.SelectedLaunchExecutableReady {
		items = append(items, "executable="+firstNonEmpty(strings.TrimSpace(launch.SelectedLaunchExecutableBlockReason), "not_ready"))
	} else if launch.SelectedLaunchReady != nil && !*launch.SelectedLaunchReady {
		items = append(items, "selected_launch="+firstNonEmpty(strings.TrimSpace(launch.SelectedLaunchBlockReason), "not_ready"))
	}
	return strings.Join(items, " ")
}

func BuildBrowserDoctorDisplayLines(
	doctor *BrowserDoctorSummary,
	bringup *BrowserDoctorBringupReport,
) []string {
	if doctor == nil {
		return nil
	}
	lines := []string{BrowserDoctorStatusText(doctor)}
	if line := BrowserDoctorRouteText(doctor.Route); line != "" {
		lines = append(lines, line)
	}
	if line := BrowserDoctorLaunchText(doctor.Launch); line != "" {
		lines = append(lines, line)
	}
	if summary := strings.TrimSpace(BrowserDoctorBringupDisplayText(bringup, doctor.Bringup)); summary != "" {
		lines = append(lines, "- browser bring-up: "+summary)
	}
	if detail := strings.TrimSpace(BrowserDoctorLaunchDetailText(doctor.Launch)); detail != "" {
		lines = append(lines, "- browser launch detail: "+detail)
	}
	if cmd := strings.TrimSpace(doctor.RepairCommand); cmd != "" {
		lines = append(lines, "- browser repair: "+cmd)
	}
	if cmd := strings.TrimSpace(doctor.AcceptanceCommand); cmd != "" {
		lines = append(lines, "- browser acceptance: "+cmd)
	}
	for _, item := range doctor.Suggestions {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		lines = append(lines, "- browser suggestion: "+item)
	}
	return lines
}
