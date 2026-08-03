package finance

import (
	"encoding/json"
	"testing"
)

func TestMetricsToolPayloadJSONShape(t *testing.T) {
	payload := MetricsToolPayload{
		Tool:                 ToolReportMetricsGuard,
		AdapterID:            "test_adapter",
		RequestedFieldsReady: true,
		Evidence: MetricsEvidence{
			CompanyName:  "测试公司",
			StockCode:    "000001",
			ReportPeriod: "2025",
			Revenue:      "100亿元",
			TrendSeries: []MetricsTrendSeriesPoint{{
				Period:  "2025",
				Revenue: "100亿元",
			}},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal metrics payload: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal metrics payload: %v", err)
	}
	if decoded["tool"] != ToolReportMetricsGuard {
		t.Fatalf("expected tool name, got %#v", decoded["tool"])
	}
	evidence, ok := decoded["evidence"].(map[string]any)
	if !ok || evidence["company_name"] != "测试公司" || evidence["revenue"] != "100亿元" {
		t.Fatalf("unexpected evidence shape: %#v", decoded["evidence"])
	}
}

func TestMetricsCandidatesPayloadJSONShape(t *testing.T) {
	payload := MetricsCandidatesPayload{
		Tool:            ToolReportMetricsCandidates,
		AdapterID:       "test_candidates",
		AdapterStatus:   "ok",
		ResolvedCode:    "000001",
		ResolvedMarket:  "SZ",
		ResolvedCompany: "测试公司",
		ResolvedEntities: []ResolvedEntityCandidate{{
			EntityName:   "测试公司",
			CodeOrTicker: "000001",
			Market:       "A",
			Confidence:   0.95,
			Provenance: []ResolvedEntityProvenanceRef{{
				Source: "test",
				Field:  "code",
				Value:  "000001",
			}},
		}},
		Candidates: []MetricsCandidate{{
			URL:                 "https://example.com/report",
			SourceKind:          "official_report",
			StockCode:           "000001",
			PreferredNextAction: "open_page",
		}},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal candidates payload: %v", err)
	}
	if !json.Valid(raw) {
		t.Fatalf("expected valid json: %s", raw)
	}
	EnsureMetricsCandidatesIdentityResolution(&payload)
	if payload.IdentityResolution == nil ||
		payload.IdentityResolution.SelectedCandidate == nil ||
		payload.IdentityResolution.SelectedCandidate.CodeOrTicker != "000001" {
		t.Fatalf("unexpected identity resolution: %#v", payload.IdentityResolution)
	}
}

func TestBriefToolPayloadJSONShape(t *testing.T) {
	payload := BriefToolPayload{
		Tool:       ToolReportBriefGuard,
		BriefReady: true,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal brief payload: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal brief payload: %v", err)
	}
	if decoded["tool"] != ToolReportBriefGuard || decoded["brief_ready"] != true {
		t.Fatalf("unexpected brief payload shape: %#v", decoded)
	}
}
