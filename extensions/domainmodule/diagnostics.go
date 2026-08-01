package domainmodule

import "strings"

// Severity classifies a diagnostic without defining host policy.
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

const (
	DiagnosticModuleRegistered       = "module_registered"
	DiagnosticModuleDuplicateID      = "module_duplicate_id"
	DiagnosticExtensionRootLoaded    = "extension_root_loaded"
	DiagnosticExtensionRootEmpty     = "extension_root_empty"
	DiagnosticExtensionRootLoadError = "extension_root_load_error"
	DiagnosticDegradedCapability     = "degraded_capability"
	DiagnosticMissingConfig          = "missing_config"
	DiagnosticConfigResolved         = "config_resolved"
	DiagnosticConfigResolveError     = "config_resolve_error"
)

// Diagnostic is a display-safe module registration observation.
type Diagnostic struct {
	ModuleID string            `json:"module_id,omitempty"`
	Severity Severity          `json:"severity,omitempty"`
	Code     string            `json:"code,omitempty"`
	Message  string            `json:"message,omitempty"`
	Details  map[string]string `json:"details,omitempty"`
}

// Diagnostics is an ordered diagnostic sequence.
type Diagnostics []Diagnostic

// ModuleReport groups diagnostics under the normalized module manifest.
type ModuleReport struct {
	Manifest    Manifest    `json:"manifest"`
	Diagnostics Diagnostics `json:"diagnostics,omitempty"`
}

// Report is the ordered result of registration orchestration.
type Report struct {
	Modules []ModuleReport `json:"modules,omitempty"`
}

// NewDiagnostic constructs a normalized, display-safe diagnostic record.
func NewDiagnostic(moduleID string, severity Severity, code string, message string, details map[string]string) Diagnostic {
	return Diagnostic{
		ModuleID: NormalizeID(moduleID),
		Severity: normalizeSeverity(severity),
		Code:     strings.TrimSpace(code),
		Message:  strings.TrimSpace(message),
		Details:  cleanDetails(details),
	}
}

func (r *Report) appendDiagnostics(manifest Manifest, diagnostics Diagnostics) {
	if r == nil {
		return
	}
	diagnostics = normalizeDiagnostics(manifest.ID, diagnostics)
	if len(diagnostics) == 0 {
		return
	}
	for i := range r.Modules {
		if NormalizeID(r.Modules[i].Manifest.ID) == NormalizeID(manifest.ID) {
			r.Modules[i].Diagnostics = append(r.Modules[i].Diagnostics, diagnostics...)
			return
		}
	}
	r.Modules = append(r.Modules, ModuleReport{Manifest: manifest, Diagnostics: diagnostics})
}

// Diagnostics returns all module diagnostics in registration order.
func (r Report) Diagnostics() Diagnostics {
	out := Diagnostics{}
	for _, module := range r.Modules {
		out = append(out, module.Diagnostics...)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// HasErrors reports whether any diagnostic has error severity.
func (r Report) HasErrors() bool {
	for _, diagnostic := range r.Diagnostics() {
		if diagnostic.Severity == SeverityError {
			return true
		}
	}
	return false
}

func normalizeDiagnostics(moduleID string, diagnostics Diagnostics) Diagnostics {
	if len(diagnostics) == 0 {
		return nil
	}
	out := make(Diagnostics, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		if strings.TrimSpace(diagnostic.ModuleID) == "" {
			diagnostic.ModuleID = NormalizeID(moduleID)
		} else {
			diagnostic.ModuleID = NormalizeID(diagnostic.ModuleID)
		}
		diagnostic.Severity = normalizeSeverity(diagnostic.Severity)
		diagnostic.Code = strings.TrimSpace(diagnostic.Code)
		diagnostic.Message = strings.TrimSpace(diagnostic.Message)
		diagnostic.Details = cleanDetails(diagnostic.Details)
		out = append(out, diagnostic)
	}
	return out
}

func normalizeSeverity(severity Severity) Severity {
	switch Severity(strings.ToLower(strings.TrimSpace(string(severity)))) {
	case SeverityWarning:
		return SeverityWarning
	case SeverityError:
		return SeverityError
	default:
		return SeverityInfo
	}
}

func cleanDetails(details map[string]string) map[string]string {
	if len(details) == 0 {
		return nil
	}
	out := map[string]string{}
	for key, value := range details {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
