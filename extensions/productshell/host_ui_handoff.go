package productshell

import "strings"

const (
	// HostUIHandoffSchemaV1 identifies the product-shell host UI handoff schema.
	HostUIHandoffSchemaV1 = "agentx.productshell.host_ui_handoff.v1"
	// HostUIHandoffSurface is the display-only handoff surface for host adapters.
	HostUIHandoffSurface = "host_ui"

	// HostUIHandoffTargetLog is a host-owned log delivery target.
	HostUIHandoffTargetLog = "log"
	// HostUIHandoffTargetSidePanel is a host-owned side-panel delivery target.
	HostUIHandoffTargetSidePanel = "side_panel"
	// HostUIHandoffTargetAdminSurface is a host-owned admin-surface delivery target.
	HostUIHandoffTargetAdminSurface = "admin_surface"
	// HostUIHandoffTargetRunOutputJSON is a host-owned structured inspection target.
	HostUIHandoffTargetRunOutputJSON = "run_output_json"

	// HostUIHandoffKindHostDiagnosticOperatorLine identifies operator-line observations.
	HostUIHandoffKindHostDiagnosticOperatorLine = "host_diagnostic_operator_line"
)

// HostUIHandoffInput selects the host-owned delivery target for a product-shell handoff.
type HostUIHandoffInput struct {
	Target string `json:"target,omitempty"`
	Source string `json:"source,omitempty"`
}

// HostUIHandoffEnvelope is a display-safe envelope that host adapters can render.
type HostUIHandoffEnvelope struct {
	Schema      string               `json:"schema,omitempty"`
	Surface     string               `json:"surface,omitempty"`
	Target      string               `json:"target,omitempty"`
	Source      string               `json:"source,omitempty"`
	EntryCount  int                  `json:"entry_count,omitempty"`
	Entries     []HostUIHandoffEntry `json:"entries,omitempty"`
	LatestEntry *HostUIHandoffEntry  `json:"latest_entry,omitempty"`
	Boundaries  []string             `json:"boundaries,omitempty"`
}

// HostUIHandoffEntry is one display-safe item inside a host UI handoff envelope.
type HostUIHandoffEntry struct {
	Target         string   `json:"target,omitempty"`
	Source         string   `json:"source,omitempty"`
	Kind           string   `json:"kind,omitempty"`
	Key            string   `json:"key,omitempty"`
	Available      bool     `json:"available"`
	Status         string   `json:"status,omitempty"`
	DisplayLine    string   `json:"display_line,omitempty"`
	MissingInputs  []string `json:"missing_inputs,omitempty"`
	BlockedReasons []string `json:"blocked_reasons,omitempty"`
	Boundaries     []string `json:"boundaries,omitempty"`
	NextHostAction string   `json:"next_host_action,omitempty"`
}

// NormalizeHostUIHandoffToken returns a display-safe token. Unsafe input is
// replaced with "redacted" and empty input remains empty.
func NormalizeHostUIHandoffToken(value string) string {
	return hostUIHandoffRenderToken(value)
}

// NormalizeHostUIHandoffDisplayLine returns a display-safe single-line value.
// Unsafe input is replaced with "redacted" and empty input remains empty.
func NormalizeHostUIHandoffDisplayLine(value string) string {
	return hostUIHandoffRenderDisplayLine(value)
}

// BuildHostUIHandoffEnvelopeFromOperatorLines projects typed operator-line
// observations into a display-safe, host-owned handoff. It does not read raw
// diagnostics, own an observation aggregate, or decide delivery.
func BuildHostUIHandoffEnvelopeFromOperatorLines(observations []HostDiagnosticOperatorLineObservation, input HostUIHandoffInput) *HostUIHandoffEnvelope {
	if len(observations) == 0 {
		return nil
	}
	target := hostUIHandoffRenderToken(input.Target)
	if target == "" {
		target = HostUIHandoffSurface
	}
	source := hostUIHandoffRenderToken(input.Source)
	if source == "" {
		source = "productshell_observation"
	}
	envelope := HostUIHandoffEnvelope{
		Schema:  HostUIHandoffSchemaV1,
		Surface: HostUIHandoffSurface,
		Target:  target,
		Source:  source,
		Boundaries: []string{
			"productshell_host_ui_handoff",
			"display_safe_handoff_fields",
			"host_adapter_owns_delivery",
			"no_host_diagnostics_json_decode",
		},
	}
	for _, observation := range observations {
		entry, ok := buildHostUIHandoffEntryFromOperatorLine(target, observation)
		if !ok {
			continue
		}
		envelope.Entries = append(envelope.Entries, entry)
	}
	if len(envelope.Entries) == 0 {
		return nil
	}
	envelope.EntryCount = len(envelope.Entries)
	latest := envelope.Entries[len(envelope.Entries)-1]
	envelope.LatestEntry = &latest
	return &envelope
}

// RenderHostUIHandoffLogFields renders one handoff entry as display-safe log fields.
func RenderHostUIHandoffLogFields(entry HostUIHandoffEntry) (string, bool) {
	key := hostUIHandoffRenderToken(entry.Key)
	status := hostUIHandoffRenderToken(entry.Status)
	displayLine := hostUIHandoffRenderDisplayLine(entry.DisplayLine)
	missing := hostUIHandoffFirstRenderToken(entry.MissingInputs)
	blocked := hostUIHandoffFirstRenderToken(entry.BlockedReasons)
	next := hostUIHandoffRenderToken(entry.NextHostAction)
	target := hostUIHandoffRenderToken(entry.Target)
	kind := hostUIHandoffRenderToken(entry.Kind)
	source := hostUIHandoffRenderToken(entry.Source)
	if key == "" &&
		status == "" &&
		displayLine == "" &&
		missing == "" &&
		blocked == "" &&
		next == "" &&
		target == "" &&
		kind == "" &&
		source == "" {
		return "", false
	}
	return strings.Join([]string{
		"key=" + hostUIHandoffRenderValue(key),
		"status=" + hostUIHandoffRenderValue(status),
		"available=" + hostUIHandoffBool(entry.Available),
		"line=" + hostUIHandoffRenderValue(displayLine),
		"missing=" + hostUIHandoffRenderValue(missing),
		"blocked=" + hostUIHandoffRenderValue(blocked),
		"next=" + hostUIHandoffRenderValue(next),
		"target=" + hostUIHandoffRenderValue(target),
		"kind=" + hostUIHandoffRenderValue(kind),
		"source=" + hostUIHandoffRenderValue(source),
	}, " "), true
}

func buildHostUIHandoffEntryFromOperatorLine(target string, observation HostDiagnosticOperatorLineObservation) (HostUIHandoffEntry, bool) {
	entry := HostUIHandoffEntry{
		Target:         target,
		Source:         hostUIHandoffRenderToken(observation.Source),
		Kind:           HostUIHandoffKindHostDiagnosticOperatorLine,
		Key:            hostUIHandoffRenderToken(observation.Key),
		Available:      observation.Available,
		Status:         hostUIHandoffRenderToken(observation.Status),
		DisplayLine:    hostUIHandoffRenderDisplayLine(observation.OperatorDisplayLine),
		MissingInputs:  hostUIHandoffRenderTokenList(observation.MissingInputs),
		BlockedReasons: hostUIHandoffRenderTokenList(observation.BlockedReasons),
		Boundaries: hostUIHandoffRenderTokenList(append([]string{
			"host_diagnostic_operator_line_observation",
			"display_safe_operator_line_only",
		}, observation.Boundaries...)),
		NextHostAction: hostUIHandoffRenderToken(observation.NextHostAction),
	}
	if entry.Source == "" {
		entry.Source = "host_diagnostic_operator_line"
	}
	if entry.Key == "" &&
		entry.Status == "" &&
		entry.DisplayLine == "" &&
		len(entry.MissingInputs) == 0 &&
		len(entry.BlockedReasons) == 0 &&
		entry.NextHostAction == "" {
		return HostUIHandoffEntry{}, false
	}
	return entry, true
}

func hostUIHandoffRenderTokenList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		token := hostUIHandoffRenderToken(value)
		if token == "" {
			continue
		}
		out = appendUniqueProductShellString(out, token)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func hostUIHandoffFirstRenderToken(values []string) string {
	for _, value := range values {
		if token := hostUIHandoffRenderToken(value); token != "" {
			return token
		}
	}
	return ""
}

func hostUIHandoffRenderValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}

func hostUIHandoffBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func hostUIHandoffRenderToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-' || r == ':' || r == '.':
		default:
			return "redacted"
		}
	}
	if strings.Contains(value, "://") {
		return "redacted"
	}
	return value
}

func hostUIHandoffRenderDisplayLine(value string) string {
	value = strings.TrimSpace(strings.NewReplacer("\n", " ", "\r", " ", "\t", " ").Replace(value))
	if value == "" {
		return ""
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-' || r == ':' || r == '.' || r == '=' || r == ';' || r == ',':
		default:
			return "redacted"
		}
	}
	if strings.Contains(value, "://") {
		return "redacted"
	}
	return value
}
