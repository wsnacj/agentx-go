package configs

import "testing"

func TestRecommendSpecsForTextRanksMatchingSpec(t *testing.T) {
	zh := &DocSpec{
		Meta: MetaSpec{DocType: "zh_report", Version: "v1"},
		Chapters: []ChapterSpec{{
			Key:           "financial_highlights",
			TitleKeywords: []string{"主要会计数据"},
			Fields: []FieldSpec{{
				Key:     "营业收入",
				Aliases: []string{"营业总收入"},
			}},
		}},
	}
	en := &DocSpec{
		Meta: MetaSpec{DocType: "en_report", Version: "v1"},
		Chapters: []ChapterSpec{{
			Key:           "financial_highlights",
			TitleKeywords: []string{"Financial Highlights"},
			Fields: []FieldSpec{{
				Key:     "Revenue",
				Aliases: []string{"Total revenue"},
			}},
		}},
	}

	got := RecommendSpecsForText("Financial Highlights\nRevenue and total revenue increased.", []*DocSpec{zh, en})
	if len(got) == 0 {
		t.Fatal("expected recommendation")
	}
	if got[0].DocType != "en_report" || got[0].Score <= 0 {
		t.Fatalf("expected English spec to rank first, got %#v", got)
	}
	if len(got[0].MatchedChapters) != 1 || got[0].MatchedChapters[0] != "financial_highlights" {
		t.Fatalf("unexpected matched chapters: %#v", got[0])
	}
	if len(got[0].MatchedFields) == 0 || got[0].MatchedFields[0] != "Revenue" {
		t.Fatalf("unexpected matched fields: %#v", got[0])
	}
}

func TestRecommendSpecsForTextSupportsCompactCJKMatching(t *testing.T) {
	spec := &DocSpec{
		Meta: MetaSpec{DocType: "zh_report"},
		Chapters: []ChapterSpec{{
			Key:           "summary",
			TitleKeywords: []string{"主要 会计 数据"},
			Fields:        []FieldSpec{{Key: "营业 收入"}},
		}},
	}

	got := RecommendSpecsForText("主要会计数据\n营业收入 100", []*DocSpec{spec})
	if len(got) != 1 || got[0].DocType != "zh_report" {
		t.Fatalf("expected compact CJK match, got %#v", got)
	}
}

func TestRecommendSpecsForTextOmitsZeroScoreSpecs(t *testing.T) {
	spec := &DocSpec{
		Meta: MetaSpec{DocType: "unmatched"},
		Chapters: []ChapterSpec{{
			Key:           "summary",
			TitleKeywords: []string{"No Match"},
			Fields:        []FieldSpec{{Key: "Missing"}},
		}},
	}
	if got := RecommendSpecsForText("unrelated text", []*DocSpec{nil, spec}); len(got) != 0 {
		t.Fatalf("expected no recommendation, got %#v", got)
	}
}
