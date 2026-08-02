package contracts

import (
	"encoding/json"
	"testing"
)

func TestQuotePayloadRoundTripPreservesReadinessAndEvidence(t *testing.T) {
	payload := QuotePayload{
		Tool:          "a_stock_quote_lookup",
		Source:        "test",
		AdapterID:     "tencent_quote",
		AdapterStatus: AdapterStatusOK,
		Subject: Subject{
			EntityName: "同花顺",
			StockCode:  "300033",
			Market:     MarketSZ,
			Verified:   true,
		},
		Freshness: Freshness{
			Mode:           FreshnessModeRealtime,
			AsOf:           "2026-05-15T15:00:00+08:00",
			TradingSession: TradingSessionClosed,
		},
		Evidence: SourceEvidence{
			Provider:  "tencent",
			SourceURL: "https://qt.gtimg.cn/q=sz300033",
			Freshness: FreshnessModeRealtime,
		},
		Readiness: BuildReadiness(AdapterStatusOK, FailureCodeNone, true, nil, nil),
		Quote: QuoteSnapshot{
			Price: MetricValue{Field: "price", Value: "100.00", Unit: "CNY"},
			PETTM: MetricValue{Field: "pe_ttm", Value: "30.5"},
			PB:    MetricValue{Field: "pb", Value: "8.1"},
		},
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal quote payload: %v", err)
	}
	var got QuotePayload
	if err := json.Unmarshal(blob, &got); err != nil {
		t.Fatalf("unmarshal quote payload: %v", err)
	}
	if !got.Readiness.AnswerReady || got.Subject.StockCode != "300033" || got.Quote.PB.Value != "8.1" || got.Evidence.Provider != "tencent" {
		t.Fatalf("unexpected roundtrip payload: %#v", got)
	}
}

func TestResearchPayloadSupportsConsensusAndReportEvidence(t *testing.T) {
	payload := ResearchPayload{
		Subject: Subject{EntityName: "绿的谐波", StockCode: "688017", Market: MarketSH, Verified: true},
		Reports: []ResearchReport{{
			Title:       "公司深度报告",
			Institution: "测试证券",
			PublishedAt: "2026-05-15",
			Rating:      "买入",
			PDFURL:      "https://example.test/report.pdf",
			Forecasts: []ForecastItem{{
				Field: "eps_forecast",
				Year:  "2026",
				Mean:  MetricValue{Value: "1.23", Unit: "CNY/share"},
				Count: 5,
			}},
		}},
	}
	if len(payload.Reports) != 1 || len(payload.Reports[0].Forecasts) != 1 || payload.Reports[0].Forecasts[0].Mean.Value != "1.23" {
		t.Fatalf("unexpected research payload: %#v", payload)
	}
}

func TestSignalAnnouncementAndProfilePayloadShapes(t *testing.T) {
	signal := SignalPayload{
		Subject: Subject{StockCode: "002475", Market: MarketSZ, Verified: true},
		Signals: []SignalEvent{{
			Type:      SignalTypeDragonTigerBoard,
			TradeDate: "2026-05-15",
			NetBuy:    MetricValue{Value: "5000", Unit: "万元"},
		}},
	}
	if signal.Signals[0].Type != SignalTypeDragonTigerBoard || signal.Signals[0].NetBuy.Unit != "万元" {
		t.Fatalf("unexpected signal payload: %#v", signal)
	}

	announcement := AnnouncementPayload{
		Subject: Subject{StockCode: "600000", Market: MarketSH, Verified: true},
		Announcements: []Announcement{{
			Title:             "年度报告",
			Type:              "年报",
			PublishedAt:       "2026-04-30",
			FullTextAvailable: true,
		}},
	}
	if !announcement.Announcements[0].FullTextAvailable {
		t.Fatalf("expected announcement full text availability")
	}

	profile := ProfilePayload{
		Subject: Subject{StockCode: "300033", Market: MarketSZ, Verified: true},
		Profile: CompanyProfile{
			Name:        "同花顺",
			Industry:    "软件开发",
			ListingDate: "2009-12-25",
			F10Sections: map[string]string{"company": "profile"},
		},
		Finance: map[string]MetricValue{"roe": {Value: "12.3", Unit: "%"}},
	}
	if profile.Profile.F10Sections["company"] != "profile" || profile.Finance["roe"].Unit != "%" {
		t.Fatalf("unexpected profile payload: %#v", profile)
	}
}
