package browserruntime

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	BrowserDoctorStatusOK      = "ok"
	BrowserDoctorStatusPending = "pending"
	BrowserDoctorStatusWarn    = "warn"
	BrowserDoctorStatusError   = "error"
)

type BrowserDoctorCheckSummary struct {
	Status  string `json:"status,omitempty"`
	Code    string `json:"code,omitempty"`
	Summary string `json:"summary,omitempty"`
	Note    string `json:"note,omitempty"`
}

type BrowserDoctorRouteSummary struct {
	BrowserDoctorCheckSummary
	Source            string `json:"source,omitempty"`
	Backend           string `json:"backend,omitempty"`
	Profile           string `json:"profile,omitempty"`
	RuntimeTarget     string `json:"runtime_target,omitempty"`
	Endpoint          string `json:"endpoint,omitempty"`
	SelectionStrategy string `json:"selection_strategy,omitempty"`
}

type BrowserDoctorLaunchSummary struct {
	BrowserDoctorCheckSummary
	Source                              string `json:"source,omitempty"`
	Backend                             string `json:"backend,omitempty"`
	Profile                             string `json:"profile,omitempty"`
	RuntimeTarget                       string `json:"runtime_target,omitempty"`
	PlaywrightCachePath                 string `json:"playwright_cache_path,omitempty"`
	PlaywrightCacheSource               string `json:"playwright_cache_source,omitempty"`
	NodeVersion                         string `json:"node_version,omitempty"`
	PlaywrightPackage                   string `json:"playwright_package,omitempty"`
	PlaywrightPackageVersion            string `json:"playwright_package_version,omitempty"`
	RuntimeBaselineReady                *bool  `json:"runtime_baseline_ready,omitempty"`
	RuntimeBaselineBlockReason          string `json:"runtime_baseline_block_reason,omitempty"`
	SelectedLaunchSource                string `json:"selected_launch_source,omitempty"`
	SelectedLaunchDeliveryGeneration    string `json:"selected_launch_delivery_generation,omitempty"`
	SelectedLaunchPayloadSource         string `json:"selected_launch_payload_source,omitempty"`
	SelectedLaunchPayloadReady          *bool  `json:"selected_launch_payload_ready,omitempty"`
	SelectedLaunchPayloadBlockReason    string `json:"selected_launch_payload_block_reason,omitempty"`
	SelectedLaunchReady                 *bool  `json:"selected_launch_ready,omitempty"`
	SelectedLaunchBlockReason           string `json:"selected_launch_block_reason,omitempty"`
	SelectedLaunchExecutableReady       *bool  `json:"selected_launch_executable_ready,omitempty"`
	SelectedLaunchExecutableBlockReason string `json:"selected_launch_executable_block_reason,omitempty"`
	DeliveryTransitionPending           *bool  `json:"delivery_transition_pending,omitempty"`
	DeliveryTransitionStage             string `json:"delivery_transition_stage,omitempty"`
	BundleReady                         *bool  `json:"bundle_ready,omitempty"`
	DeliveryReady                       *bool  `json:"delivery_ready,omitempty"`
	NodeModulesReady                    *bool  `json:"node_modules_ready,omitempty"`
	BrowserReady                        *bool  `json:"browser_ready,omitempty"`
	BootstrapState                      string `json:"bootstrap_state,omitempty"`
	BootstrapErrorCode                  string `json:"bootstrap_error_code,omitempty"`
	LaunchBlockReason                   string `json:"launch_block_reason,omitempty"`
}

type BrowserDoctorBringupStep struct {
	Label string `json:"label,omitempty"`
	Run   string `json:"run,omitempty"`
	When  string `json:"when,omitempty"`
}

type BrowserDoctorBringupSummary struct {
	SubstrateSource         string                     `json:"substrate_source,omitempty"`
	SubstrateEndpoint       string                     `json:"substrate_endpoint,omitempty"`
	PreferredRuntimeBackend string                     `json:"preferred_runtime_backend,omitempty"`
	PreferredRuntimeProfile string                     `json:"preferred_runtime_profile,omitempty"`
	PreferredRuntimeTarget  string                     `json:"preferred_runtime_target,omitempty"`
	PrimaryStep             string                     `json:"primary_step,omitempty"`
	DoctorAction            string                     `json:"doctor_action,omitempty"`
	RepairAction            string                     `json:"repair_action,omitempty"`
	ReadyAction             string                     `json:"ready_action,omitempty"`
	AcceptanceCommand       string                     `json:"acceptance_command,omitempty"`
	Steps                   []BrowserDoctorBringupStep `json:"steps,omitempty"`
}

type BrowserDoctorSummary struct {
	Status            string                       `json:"status,omitempty"`
	Ready             bool                         `json:"ready,omitempty"`
	Route             *BrowserDoctorRouteSummary   `json:"route,omitempty"`
	Launch            *BrowserDoctorLaunchSummary  `json:"launch,omitempty"`
	Bringup           *BrowserDoctorBringupSummary `json:"bringup,omitempty"`
	ConfiguredTargets []string                     `json:"configured_targets,omitempty"`
	RepairCommand     string                       `json:"repair_command,omitempty"`
	AcceptanceCommand string                       `json:"acceptance_command,omitempty"`
	Suggestions       []string                     `json:"suggestions,omitempty"`
}

func BrowserDoctorAggregateStatus(statuses ...string) string {
	best := ""
	bestRank := 0
	for _, status := range statuses {
		status = strings.ToLower(strings.TrimSpace(status))
		rank := browserDoctorStatusRank(status)
		if rank <= 0 || rank <= bestRank {
			continue
		}
		best = status
		bestRank = rank
	}
	return best
}

func BrowserDoctorAcceptanceScript(root string) string {
	return browserDoctorRepoScript(root, "browserd-platform-acceptance-check.sh")
}

func BrowserDoctorBootstrapRepairScript(root string) string {
	return browserDoctorRepoScript(root, "browserd-bootstrap-repair.sh")
}

func browserDoctorRepoScript(root string, filename string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return ""
	}
	searchRoots := browserDoctorSearchRoots(root)
	for _, searchRoot := range searchRoots {
		current := searchRoot
		for current != "" {
			candidate := filepath.Join(current, "core", "agentx", "browserdaemon", filename)
			if browserDoctorFileExists(candidate) {
				return filepath.Clean(candidate)
			}
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}
	}
	return ""
}

func BrowserDoctorAcceptanceCommand(root string) string {
	script := strings.TrimSpace(BrowserDoctorAcceptanceScript(root))
	if script == "" {
		return ""
	}
	return "bash " + script
}

func BrowserDoctorBootstrapRepairCommand(root string) string {
	script := strings.TrimSpace(BrowserDoctorBootstrapRepairScript(root))
	if script == "" {
		return ""
	}
	return "bash " + script
}

func BrowserDoctorBringupSteps(
	doctorAction string,
	repairAction string,
	readyAction string,
	acceptanceCommand string,
) []BrowserDoctorBringupStep {
	steps := make([]BrowserDoctorBringupStep, 0, 4)
	appendStep := func(label string, run string, when string) {
		run = strings.TrimSpace(run)
		if run == "" {
			return
		}
		steps = append(steps, BrowserDoctorBringupStep{
			Label: strings.TrimSpace(label),
			Run:   run,
			When:  strings.TrimSpace(when),
		})
	}
	appendStep("doctor", doctorAction, "always")
	appendStep("repair", repairAction, "if_suggested")
	appendStep("ready", readyAction, "after_doctor_or_repair")
	appendStep("acceptance", acceptanceCommand, "optional_validation")
	if len(steps) == 0 {
		return nil
	}
	return steps
}

func BrowserDoctorBringupSummaryText(bringup *BrowserDoctorBringupSummary) string {
	if bringup == nil {
		return ""
	}
	steps := make([]string, 0, len(bringup.Steps))
	for _, step := range bringup.Steps {
		run := strings.TrimSpace(step.Run)
		if run == "" {
			continue
		}
		switch strings.TrimSpace(step.When) {
		case "if_suggested":
			run += " (if suggested)"
		case "optional_validation":
			run = "validate: " + run
		}
		steps = append(steps, run)
	}
	if len(steps) == 0 {
		return ""
	}
	return strings.Join(steps, " -> ")
}

func BrowserDoctorBringupDetailText(bringup *BrowserDoctorBringupSummary) string {
	summary := strings.TrimSpace(BrowserDoctorBringupSummaryText(bringup))
	extras := browserDoctorBringupDetailExtras(bringup)
	if len(extras) == 0 {
		return summary
	}
	if summary == "" {
		return strings.Join(extras, " ")
	}
	return summary + " [" + strings.Join(extras, " ") + "]"
}

func browserDoctorBringupDetailExtras(bringup *BrowserDoctorBringupSummary) []string {
	if bringup == nil {
		return nil
	}
	extras := make([]string, 0, 6)
	if value := strings.TrimSpace(bringup.PrimaryStep); value != "" {
		extras = append(extras, "primary_step="+value)
	}
	if value := strings.TrimSpace(bringup.SubstrateSource); value != "" {
		extras = append(extras, "substrate_source="+value)
	}
	if value := strings.TrimSpace(bringup.SubstrateEndpoint); value != "" {
		extras = append(extras, "substrate_endpoint="+value)
	}
	if value := strings.TrimSpace(bringup.PreferredRuntimeBackend); value != "" {
		extras = append(extras, "preferred_runtime_backend="+value)
	}
	if value := strings.TrimSpace(bringup.PreferredRuntimeProfile); value != "" {
		extras = append(extras, "preferred_runtime_profile="+value)
	}
	if value := strings.TrimSpace(bringup.PreferredRuntimeTarget); value != "" {
		extras = append(extras, "preferred_runtime_target="+value)
	}
	if len(extras) == 0 {
		return nil
	}
	return extras
}

func browserDoctorStatusRank(status string) int {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case BrowserDoctorStatusError:
		return 4
	case BrowserDoctorStatusWarn:
		return 3
	case BrowserDoctorStatusPending:
		return 2
	case BrowserDoctorStatusOK:
		return 1
	default:
		return 0
	}
}

func browserDoctorSearchRoots(root string) []string {
	seen := map[string]bool{}
	roots := make([]string, 0, 2)
	appendRoot := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		value = filepath.Clean(value)
		if seen[value] {
			return
		}
		seen[value] = true
		roots = append(roots, value)
	}
	appendRoot(root)
	if cwd, err := os.Getwd(); err == nil {
		appendRoot(cwd)
	}
	return roots
}

func browserDoctorFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
