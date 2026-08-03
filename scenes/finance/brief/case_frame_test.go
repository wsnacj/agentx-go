package brief

import "testing"

func TestBuildBriefCaseInput(t *testing.T) {
	input, ok := BuildBriefCaseInput("查一下浦发银行的最新财报，给我一段简报")
	if !ok {
		t.Fatal("expected brief case input")
	}
	if input["output_style"] != "single_paragraph_brief" {
		t.Fatalf("unexpected output style: %#v", input["output_style"])
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "浦发银行" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
	if input["source_policy"] != "public_web_prefer_official_annual_report_pdf" {
		t.Fatalf("unexpected source policy: %#v", input["source_policy"])
	}
}

func TestBuildBriefCaseInputRecognizesOfficialAnnualReportWording(t *testing.T) {
	input, ok := BuildBriefCaseInput("请去巨潮资讯查东方雨虹002271的2025年年度报告，给我一段只包含关键信息的简报。")
	if !ok {
		t.Fatal("expected annual report brief case input")
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "东方雨虹" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
	identifiers, _ := entity["identifiers"].(map[string]any)
	if identifiers["stock_code"] != "002271" {
		t.Fatalf("unexpected identifiers: %#v", identifiers)
	}
	if input["period_policy"] != "explicit_year:2025" {
		t.Fatalf("unexpected period policy: %#v", input["period_policy"])
	}
}

func TestBuildBriefCaseInputRecognizesSEC20FBriefWording(t *testing.T) {
	input, ok := BuildBriefCaseInput("去SEC官方20-F看下百度25年财报，给我一段只包含关键信息的简报")
	if !ok {
		t.Fatal("expected SEC 20-F brief case input")
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "百度" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
	if input["period_policy"] != "explicit_year:2025" {
		t.Fatalf("unexpected period policy: %#v", input["period_policy"])
	}
}

func TestBuildBriefCaseInputRecognizesSEC20FNetLossBriefWording(t *testing.T) {
	input, ok := BuildBriefCaseInput("去SEC官方20-F看下高途25年财报，给我一段只包含关键信息的简报")
	if !ok {
		t.Fatal("expected SEC 20-F net-loss brief case input")
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "高途" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
	if input["period_policy"] != "explicit_year:2025" {
		t.Fatalf("unexpected period policy: %#v", input["period_policy"])
	}
}

func TestBuildBriefCaseInputStripsTrailingSECFormTypeFromEntity(t *testing.T) {
	input, ok := BuildBriefCaseInput("请查一下腾讯音乐最新的20-F财报，给我一段只包含关键信息的简报")
	if !ok {
		t.Fatal("expected SEC 20-F brief case input")
	}
	entity, _ := input["entity"].(map[string]any)
	if entity["name"] != "腾讯音乐" {
		t.Fatalf("unexpected entity: %#v", entity)
	}
}

func TestBuildBriefCaseInputRejectsMetricOnlyQuestion(t *testing.T) {
	if input, ok := BuildBriefCaseInput("查一下浦发银行的最新财报，25年营收和利润是多少"); ok || input != nil {
		t.Fatalf("did not expect brief case for metric-only request: %#v", input)
	}
}
