package hostkit

import (
	"strings"
	"testing"

	astockcontracts "github.com/wsnacj/agentx-go/extensions/astock/contracts"
	astocktools "github.com/wsnacj/agentx-go/extensions/astock/contracts"
)

func TestFormatAStockAnswerFormatsQuotePayload(t *testing.T) {
	got, err := FormatAStockAnswer(map[string]any{
		"answer_kind": "valuation",
		"payload": astockcontracts.QuotePayload{
			Tool:          astocktools.ToolAStockQuoteLookup,
			AdapterStatus: astockcontracts.AdapterStatusOK,
			Subject: astockcontracts.Subject{
				EntityName: "同花顺",
				StockCode:  "300033",
				Market:     astockcontracts.MarketSZ,
				Verified:   true,
			},
			Freshness: astockcontracts.Freshness{AsOf: "2026-05-16T04:00:00+08:00"},
			Evidence:  astockcontracts.SourceEvidence{SourceURL: "https://qt.gtimg.cn/q=sz300033"},
			Readiness: astockcontracts.Readiness{AnswerReady: true, RequestedFieldsReady: true},
			Quote: astockcontracts.QuoteSnapshot{
				Price:         astockcontracts.MetricValue{Value: "231.45", Unit: "CNY"},
				ChangePercent: astockcontracts.MetricValue{Value: "-2.34", Unit: "%"},
				PETTM:         astockcontracts.MetricValue{Value: "52.15"},
				PB:            astockcontracts.MetricValue{Value: "24.92"},
				MarketCap:     astockcontracts.MetricValue{Value: "1741.99", Unit: "亿元"},
			},
		},
	})
	if err != nil {
		t.Fatalf("format quote answer: %v", err)
	}
	for _, want := range []string{"同花顺（300033.SZ）", "价格：231.45CNY", "PE(TTM)：52.15", "来源：https://qt.gtimg.cn/q=sz300033", "不构成任何投资建议"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected formatted quote answer to contain %q, got:\n%s", want, got)
		}
	}
	if strings.HasPrefix(strings.TrimSpace(got), "{") {
		t.Fatalf("formatter must return natural-language text, got JSON-like reply: %s", got)
	}
}

func TestFormatAStockAnswerToleratesScalarQuoteMetrics(t *testing.T) {
	got, err := FormatAStockAnswer(map[string]any{
		"answer_kind": "valuation",
		"payload": map[string]any{
			"tool":           astocktools.ToolAStockQuoteLookup,
			"adapter_status": "ok",
			"subject": map[string]any{
				"entity_name": "同花顺",
				"stock_code":  "300033",
				"market":      "sz",
			},
			"freshness": map[string]any{"as_of": "2026-05-16"},
			"evidence":  map[string]any{"source_url": "https://qt.gtimg.cn/q=sz300033"},
			"readiness": map[string]any{"answer_ready": true, "requested_fields_ready": true},
			"quote": map[string]any{
				"price":          "231.45",
				"change_percent": -2.34,
				"pe_ttm":         "52.15",
				"pb":             "24.92",
				"market_cap":     "1741.99",
			},
		},
	})
	if err != nil {
		t.Fatalf("format scalar quote metrics: %v", err)
	}
	for _, want := range []string{"同花顺（300033.SZ）", "价格：231.45", "涨跌幅：-2.34", "总市值：1741.99"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected formatted scalar quote answer to contain %q, got:\n%s", want, got)
		}
	}
}

func TestFormatAStockAnswerSkipsMissingPayloadWithoutHardError(t *testing.T) {
	got, err := FormatAStockAnswer(map[string]any{
		"user_message":           "综合看财报和行情",
		"_argument_decode_error": "decode failed",
	})
	if err != nil {
		t.Fatalf("missing formatter payload should not hard fail: %v", err)
	}
	for _, want := range []string{"a_stock_answer_format 未执行格式化", "缺少可验证的标准 A 股工具 payload", "不要把模型自行拼装的混合财报/行情 payload 当成新的事实来源"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected skipped formatter guidance to contain %q, got:\n%s", want, got)
		}
	}
}

func TestFormatAStockAnswerSkipsModelAssembledCompositePayload(t *testing.T) {
	got, err := FormatAStockAnswer(map[string]any{
		"user_message": "综合看财报和行情",
		"answer_kind":  "research",
		"payload": map[string]any{
			"finance_report":  map[string]any{"brief": "收入增长"},
			"quote_valuation": map[string]any{"price": "10元"},
			"signals":         map[string]any{"fund_flow": "暂缺"},
		},
	})
	if err != nil {
		t.Fatalf("composite formatter payload should not hard fail: %v", err)
	}
	for _, want := range []string{"non_standard_composite_payload", "不要把模型自行拼装的混合财报/行情 payload 当成新的事实来源"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected composite skip guidance to contain %q, got:\n%s", want, got)
		}
	}
}

func TestFormatAStockAnswerSkipsModelReassembledQuotePayload(t *testing.T) {
	got, err := FormatAStockAnswer(map[string]any{
		"user_message": "海康威视现在股价、PE、PB、市值大概是多少？数据时间点要说清楚。",
		"answer_kind":  "valuation",
		"payload": map[string]any{
			"subject": map[string]any{
				"entity_name": "海康威视",
				"stock_code":  "002415",
				"market":      "sz",
			},
			"as_of":      "2026-07-14T17:42:04+08:00",
			"freshness":  "realtime",
			"source":     "腾讯行情",
			"source_url": "https://qt.gtimg.cn/q=sz002415",
			"quote": map[string]any{
				"price":      "33.58元",
				"pe_ttm":     "20.60",
				"pb":         "3.89",
				"market_cap": "3077.56亿元",
			},
		},
	})
	if err != nil {
		t.Fatalf("model-reassembled quote payload should not hard fail: %v", err)
	}
	for _, want := range []string{"missing_standard_tool_identity", "基于前面已经返回的"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected reconstructed quote skip guidance to contain %q, got:\n%s", want, got)
		}
	}
}

func TestFormatAStockAnswerSkipsMalformedIdentifiedQuotePayload(t *testing.T) {
	got, err := FormatAStockAnswer(map[string]any{
		"answer_kind": "valuation",
		"payload": map[string]any{
			"tool":      astocktools.ToolAStockQuoteLookup,
			"freshness": "realtime",
		},
	})
	if err != nil {
		t.Fatalf("malformed identified quote payload should not hard fail: %v", err)
	}
	if !strings.Contains(got, "invalid_valuation_payload") {
		t.Fatalf("expected invalid valuation skip guidance, got:\n%s", got)
	}
}

func TestFormatAStockAnswerSkipsIdentifiedQuoteWithoutReadinessContract(t *testing.T) {
	got, err := FormatAStockAnswer(map[string]any{
		"answer_kind": "valuation",
		"payload": map[string]any{
			"tool":           astocktools.ToolAStockQuoteLookup,
			"adapter_status": "ok",
			"subject": map[string]any{
				"entity_name": "海康威视",
				"stock_code":  "002415",
				"market":      "sz",
				"verified":    true,
			},
			"freshness": map[string]any{"as_of": "2026-07-17T23:01:12+08:00"},
			"quote": map[string]any{
				"price":      map[string]any{"value": "33.15", "unit": "CNY"},
				"pe_ttm":     map[string]any{"value": "20.34"},
				"pb":         map[string]any{"value": "3.84"},
				"market_cap": map[string]any{"value": "3038.15", "unit": "亿元"},
			},
		},
	})
	if err != nil {
		t.Fatalf("identified quote without readiness should not hard fail: %v", err)
	}
	if !strings.Contains(got, "missing_readiness_contract") {
		t.Fatalf("expected missing readiness skip guidance, got:\n%s", got)
	}
	if strings.Contains(got, "当前不能形成可靠回答：ok") {
		t.Fatalf("formatter must not turn absent readiness into a contradictory ok failure: %s", got)
	}
}

func TestFormatAStockAnswerFormatsResearchPayload(t *testing.T) {
	got, err := FormatAStockAnswer(map[string]any{
		"answer_kind": "research",
		"payload": astockcontracts.ResearchPayload{
			Tool:          astocktools.ToolAStockResearchLookup,
			AdapterStatus: astockcontracts.AdapterStatusOK,
			Subject:       astockcontracts.Subject{EntityName: "贵州茅台", StockCode: "600519", Market: astockcontracts.MarketSH},
			Freshness:     astockcontracts.Freshness{AsOf: "2026-05-16T04:00:00+08:00"},
			Evidence:      astockcontracts.SourceEvidence{SourceURL: "https://data.eastmoney.com/report/stock.jshtml"},
			Readiness:     astockcontracts.Readiness{AnswerReady: true, RequestedFieldsReady: true},
			Reports: []astockcontracts.ResearchReport{{
				Title:       "公司事件点评报告",
				Institution: "华鑫证券",
				PublishedAt: "2026-05-05",
				Rating:      "买入",
				PDFURL:      "https://pdf.dfcfw.com/a.pdf",
			}},
		},
	})
	if err != nil {
		t.Fatalf("format research answer: %v", err)
	}
	for _, want := range []string{"贵州茅台（600519.SH）", "华鑫证券", "买入", "https://pdf.dfcfw.com/a.pdf", "不构成任何投资建议"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected formatted research answer to contain %q, got:\n%s", want, got)
		}
	}
}

func TestFormatAStockAnswerToleratesFlexibleResearchForecastValues(t *testing.T) {
	got, err := FormatAStockAnswer(map[string]any{
		"answer_kind": "research",
		"payload": map[string]any{
			"tool":           astocktools.ToolAStockResearchLookup,
			"adapter_status": "ok",
			"subject": map[string]any{
				"entity_name": "贵州茅台",
				"stock_code":  "600519",
				"market":      "sh",
			},
			"readiness": map[string]any{"answer_ready": true},
			"freshness": map[string]any{"as_of": "2026-05-16"},
			"evidence":  map[string]any{"source_url": "https://data.eastmoney.com/report"},
			"reports": []any{
				map[string]any{
					"title":        "贵州茅台点评",
					"institution":  "测试证券",
					"published_at": "2026-05-15",
					"rating":       "买入",
					"forecasts": []any{
						map[string]any{"field": "eps", "year": "2026", "value": "12.3"},
						map[string]any{"field": "net_profit", "year": "2027", "value": 456.7},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("format flexible research answer: %v", err)
	}
	for _, want := range []string{"贵州茅台（600519.SH）", "贵州茅台点评（测试证券，2026-05-15，买入）", "来源：https://data.eastmoney.com/report"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected formatted flexible research answer to contain %q, got:\n%s", want, got)
		}
	}
}

func TestFormatAStockAnswerFormatsSignalPayload(t *testing.T) {
	got, err := FormatAStockAnswer(map[string]any{
		"answer_kind": "signal",
		"payload": astockcontracts.SignalPayload{
			Tool:          astocktools.ToolAStockSignalLookup,
			AdapterStatus: astockcontracts.AdapterStatusOK,
			Subject:       astockcontracts.Subject{EntityName: "宁德时代", StockCode: "300750", Market: astockcontracts.MarketSZ},
			Freshness:     astockcontracts.Freshness{TradeDate: "2026-05-15"},
			Evidence:      astockcontracts.SourceEvidence{SourceURL: "https://data.eastmoney.com/zjlx/sz300750.html"},
			Readiness:     astockcontracts.Readiness{AnswerReady: true, RequestedFieldsReady: true},
			Signals: []astockcontracts.SignalEvent{{
				Type:      astockcontracts.SignalTypeFundFlow,
				Title:     "宁德时代主力资金流向",
				TradeDate: "2026-05-15",
				NetBuy:    astockcontracts.MetricValue{Value: "-75278.81", Unit: "万元"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("format signal answer: %v", err)
	}
	for _, want := range []string{"宁德时代（300750.SZ）", "主力资金流向", "净买额：-75278.81万元"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected formatted signal answer to contain %q, got:\n%s", want, got)
		}
	}
}
