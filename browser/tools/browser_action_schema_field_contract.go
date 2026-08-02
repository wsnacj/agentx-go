package tools

import (
	"sort"
	"strings"

	types "github.com/wsnacj/agentx-go/components/llm"
)

const (
	BrowserActionSchemaFieldKindAction           = "action"
	BrowserActionSchemaFieldKindTarget           = "target"
	BrowserActionSchemaFieldKindSemanticHint     = "semantic_hint"
	BrowserActionSchemaFieldKindArtifact         = "artifact"
	BrowserActionSchemaFieldKindPageState        = "page_state"
	BrowserActionSchemaFieldKindInput            = "input"
	BrowserActionSchemaFieldKindSnapshot         = "snapshot"
	BrowserActionSchemaFieldKindRuntimeSelection = "runtime_selection"
	BrowserActionSchemaFieldKindRuntimeControl   = "runtime_control"
	BrowserActionSchemaFieldKindOption           = "option"
)

type BrowserActionSchemaFieldContract struct {
	Field                    string   `json:"field"`
	Owner                    string   `json:"owner"`
	Kind                     string   `json:"kind"`
	BackendNeutral           bool     `json:"backend_neutral"`
	LLMActionSchema          bool     `json:"llm_action_schema"`
	BrowserRuntimeOwned      bool     `json:"browser_runtime_owned,omitempty"`
	PlaywrightLocatorDSL     bool     `json:"playwright_locator_dsl,omitempty"`
	BackendSpecificSemantics bool     `json:"backend_specific_semantics,omitempty"`
	Notes                    []string `json:"notes,omitempty"`
}

type BrowserActionSchemaFieldFinding struct {
	Rule            string `json:"rule"`
	Surface         string `json:"surface,omitempty"`
	Field           string `json:"field,omitempty"`
	Message         string `json:"message"`
	SuggestedAction string `json:"suggested_action,omitempty"`
}

func BuiltinBrowserActionSchemaFieldContracts() []BrowserActionSchemaFieldContract {
	var out []BrowserActionSchemaFieldContract
	seen := map[string]bool{}
	add := func(kind string, notes []string, fields ...string) {
		for _, raw := range fields {
			field := strings.ToLower(strings.TrimSpace(raw))
			if field == "" || seen[field] {
				continue
			}
			seen[field] = true
			out = append(out, BrowserActionSchemaFieldContract{
				Field:           field,
				Owner:           "tools/browser_action_schema",
				Kind:            kind,
				BackendNeutral:  true,
				LLMActionSchema: true,
				Notes:           append([]string(nil), notes...),
			})
		}
	}

	add(BrowserActionSchemaFieldKindAction, nil,
		"action", "kind", "operation", "dialog_action", "accept", "dismiss", "submit")
	add(BrowserActionSchemaFieldKindTarget, []string{"DOM selector means a backend-neutral CSS/DOM selector, not a Playwright locator chain."},
		"ref", "element_ref", "input_ref", "selector", "start_ref", "end_ref", "start_selector", "end_selector", "from", "to", "frame", "target", "tab_index", "index")
	add(BrowserActionSchemaFieldKindSemanticHint, []string{"Semantic hints are consumed by browserruntime resolver/actionability; tools schema must not encode Playwright locator DSL."},
		"element", "label", "text", "start_element", "end_element", "start_label", "end_label")
	add(BrowserActionSchemaFieldKindArtifact, nil,
		"path", "paths", "file", "files", "output", "output_path")
	add(BrowserActionSchemaFieldKindPageState, nil,
		"url", "request_url", "response_url", "domain", "origin", "headers", "headers_json", "cookies", "entries", "name", "key", "value", "values", "same_site", "http_only", "secure", "expires", "storage_kind", "clear")
	add(BrowserActionSchemaFieldKindInput, nil,
		"type", "fields", "prompt", "prompt_text", "script", "javascript", "js", "username", "password")
	add(BrowserActionSchemaFieldKindSnapshot, nil,
		"format", "snapshot_format", "mode", "refs", "interactive", "compact", "efficient", "depth", "level", "filter", "max_chars", "max_elements")
	add(BrowserActionSchemaFieldKindRuntimeSelection, nil,
		"browser", "browser_app", "profile", "runtime_target", "remember_target", "remember", "remember_profile", "color", "copy_from")
	add(BrowserActionSchemaFieldKindRuntimeControl, nil,
		"coordination_goal", "include_routes", "force", "enabled")
	add(BrowserActionSchemaFieldKindOption, nil,
		"wait_ms", "post_wait_ms", "delay_ms", "full_page", "landscape", "print_background", "width", "height", "latitude", "longitude", "accuracy", "media", "timezone", "locale", "device")

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Field < out[j].Field
	})
	return out
}

func AuditBrowserActionSchemaFieldContracts(contracts []BrowserActionSchemaFieldContract) []BrowserActionSchemaFieldFinding {
	if len(contracts) == 0 {
		return nil
	}
	findings := make([]BrowserActionSchemaFieldFinding, 0)
	seen := map[string]bool{}
	for _, contract := range contracts {
		field := strings.ToLower(strings.TrimSpace(contract.Field))
		if field == "" {
			findings = append(findings, BrowserActionSchemaFieldFinding{
				Rule:            "browser_action_schema_field_name_required",
				Message:         "browser action schema field contracts must declare a stable field name",
				SuggestedAction: "name the field before adding it to browser or browser_act parameters",
			})
			continue
		}
		if seen[field] {
			findings = append(findings, BrowserActionSchemaFieldFinding{
				Rule:            "browser_action_schema_field_duplicate_contract",
				Field:           field,
				Message:         "browser action schema field contracts must be unique by field name",
				SuggestedAction: "merge duplicate contracts so each LLM-facing browser field has one owner and classification",
			})
		}
		seen[field] = true
		if strings.TrimSpace(contract.Owner) == "" {
			findings = append(findings, BrowserActionSchemaFieldFinding{
				Rule:            "browser_action_schema_field_owner_required",
				Field:           field,
				Message:         "browser action schema field contracts must declare an owner",
				SuggestedAction: "assign ownership to the LLM-facing tools schema or the browserruntime resolver/backend surface",
			})
		}
		if !browserActionSchemaFieldKindKnown(contract.Kind) {
			findings = append(findings, BrowserActionSchemaFieldFinding{
				Rule:            "browser_action_schema_field_kind_required",
				Field:           field,
				Message:         "browser action schema field contracts must declare a known field kind",
				SuggestedAction: "classify the field as action, target, semantic_hint, artifact, page_state, input, snapshot, runtime_selection, runtime_control, or option",
			})
		}
		if contract.LLMActionSchema && !contract.BackendNeutral {
			findings = append(findings, BrowserActionSchemaFieldFinding{
				Rule:            "browser_action_schema_field_backend_neutral_required",
				Field:           field,
				Message:         "LLM-facing browser action schema fields must remain backend-neutral",
				SuggestedAction: "keep backend-specific semantics in browserruntime resolver/backend code and expose only neutral hints in tools schema",
			})
		}
		if contract.LLMActionSchema && (contract.PlaywrightLocatorDSL || browserActionSchemaFieldIsPlaywrightOnly(field)) {
			findings = append(findings, BrowserActionSchemaFieldFinding{
				Rule:            "browser_action_schema_playwright_locator_field_forbidden",
				Field:           field,
				Message:         "LLM-facing browser action schema must not expose Playwright-only locator DSL fields",
				SuggestedAction: "represent the target as ref, selector, element, label, text, or another backend-neutral hint; keep Playwright locator details inside browserruntime",
			})
		}
		if contract.LLMActionSchema && contract.BackendSpecificSemantics && !contract.BrowserRuntimeOwned {
			findings = append(findings, BrowserActionSchemaFieldFinding{
				Rule:            "browser_action_schema_backend_specific_owner_required",
				Field:           field,
				Message:         "backend-specific browser semantics require browserruntime ownership",
				SuggestedAction: "move backend-specific behavior behind the browserruntime resolver/backend boundary before exposing it",
			})
		}
	}
	return sortBrowserActionSchemaFieldFindings(findings)
}

func AuditBrowserActionToolSchema(def types.Tool, contracts []BrowserActionSchemaFieldContract) []BrowserActionSchemaFieldFinding {
	contractFindings := AuditBrowserActionSchemaFieldContracts(contracts)
	byField := map[string]BrowserActionSchemaFieldContract{}
	for _, contract := range contracts {
		field := strings.ToLower(strings.TrimSpace(contract.Field))
		if field == "" {
			continue
		}
		if _, exists := byField[field]; !exists {
			byField[field] = contract
		}
	}
	surface := strings.TrimSpace(def.Function.Name)
	fields := browserActionSchemaPropertyKeys(def.Function.Parameters)
	findings := make([]BrowserActionSchemaFieldFinding, 0, len(contractFindings)+len(fields))
	findings = append(findings, contractFindings...)
	for _, field := range fields {
		if browserActionSchemaFieldIsPlaywrightOnly(field) {
			findings = append(findings, BrowserActionSchemaFieldFinding{
				Rule:            "browser_action_schema_playwright_locator_field_forbidden",
				Surface:         surface,
				Field:           field,
				Message:         "browser action schema must not expose Playwright-only locator DSL fields",
				SuggestedAction: "use backend-neutral target hints and keep Playwright locator details inside browserruntime resolver/backend code",
			})
			continue
		}
		contract, ok := byField[field]
		if !ok {
			findings = append(findings, BrowserActionSchemaFieldFinding{
				Rule:            "browser_action_schema_field_contract_required",
				Surface:         surface,
				Field:           field,
				Message:         "browser action schema fields must have an explicit backend-neutral field contract",
				SuggestedAction: "add a BrowserActionSchemaFieldContract before exposing the field to browser or browser_act",
			})
			continue
		}
		if contract.LLMActionSchema && !contract.BackendNeutral {
			findings = append(findings, BrowserActionSchemaFieldFinding{
				Rule:            "browser_action_schema_field_backend_neutral_required",
				Surface:         surface,
				Field:           field,
				Message:         "browser action schema field contract is not backend-neutral",
				SuggestedAction: "move backend-specific semantics behind browserruntime or reframe the field as a neutral hint",
			})
		}
	}
	return sortBrowserActionSchemaFieldFindings(findings)
}

func browserActionSchemaFieldKindKnown(kind string) bool {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case BrowserActionSchemaFieldKindAction,
		BrowserActionSchemaFieldKindTarget,
		BrowserActionSchemaFieldKindSemanticHint,
		BrowserActionSchemaFieldKindArtifact,
		BrowserActionSchemaFieldKindPageState,
		BrowserActionSchemaFieldKindInput,
		BrowserActionSchemaFieldKindSnapshot,
		BrowserActionSchemaFieldKindRuntimeSelection,
		BrowserActionSchemaFieldKindRuntimeControl,
		BrowserActionSchemaFieldKindOption:
		return true
	default:
		return false
	}
}

func browserActionSchemaFieldIsPlaywrightOnly(field string) bool {
	field = strings.ToLower(strings.TrimSpace(field))
	if field == "" {
		return false
	}
	switch field {
	case "get_by_role",
		"get_by_text",
		"get_by_label",
		"get_by_placeholder",
		"get_by_alt_text",
		"get_by_title",
		"get_by_test_id",
		"test_id",
		"data_test_id",
		"locator",
		"locator_chain",
		"locator_plan",
		"locator_order",
		"match_plan",
		"element_resolver",
		"frame_locator",
		"has",
		"has_not",
		"has_text",
		"has_not_text",
		"nth",
		"exact":
		return true
	default:
		return strings.HasPrefix(field, "playwright_")
	}
}

func browserActionSchemaPropertyKeys(value any) []string {
	seen := map[string]bool{}
	var walk func(any)
	walk = func(current any) {
		switch typed := current.(type) {
		case map[string]any:
			if props, ok := typed["properties"].(map[string]any); ok {
				for key, nested := range props {
					field := strings.ToLower(strings.TrimSpace(key))
					if field != "" {
						seen[field] = true
					}
					walk(nested)
				}
			}
			for _, nested := range typed {
				walk(nested)
			}
		case []any:
			for _, nested := range typed {
				walk(nested)
			}
		case []map[string]any:
			for _, nested := range typed {
				walk(nested)
			}
		}
	}
	walk(value)
	out := make([]string, 0, len(seen))
	for field := range seen {
		out = append(out, field)
	}
	sort.Strings(out)
	return out
}

func sortBrowserActionSchemaFieldFindings(findings []BrowserActionSchemaFieldFinding) []BrowserActionSchemaFieldFinding {
	if len(findings) == 0 {
		return nil
	}
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Rule != findings[j].Rule {
			return findings[i].Rule < findings[j].Rule
		}
		if findings[i].Surface != findings[j].Surface {
			return findings[i].Surface < findings[j].Surface
		}
		return findings[i].Field < findings[j].Field
	})
	return findings
}
