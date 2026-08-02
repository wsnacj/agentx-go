package publicnews

import "testing"

func TestDefaultEvidenceQualityPolicyRuleIDs(t *testing.T) {
	policy := DefaultEvidenceQualityPolicy()
	cases := []struct {
		name   string
		input  EvidenceQualityInput
		ruleID string
	}{
		{
			name:   "source_metadata_line",
			input:  EvidenceQualityInput{Headline: "示例事件", Line: "原标题：示例事件"},
			ruleID: EvidenceQualityRuleSourceMetadataLine,
		},
		{
			name:   "authorship_metadata_line",
			input:  EvidenceQualityInput{Headline: "腾讯AI进展", Line: "定焦One原创 作者 | 陈丹 编辑 | 魏佳"},
			ruleID: EvidenceQualityRuleSourceMetadataLine,
		},
		{
			name:   "promotional_noise",
			input:  EvidenceQualityInput{KeyUpdate: "炒股就看金麒麟分析师研报，权威，专业，及时，全面，助您挖掘潜力主题机会！"},
			ruleID: EvidenceQualityRulePromotionalNoise,
		},
		{
			name:   "hypothetical_rewrite",
			input:  EvidenceQualityInput{Headline: "骂了大半年，马斯克突然改口", KeyUpdate: "换句话说，马斯克要是想搞垮Anthropic，随时能做到。"},
			ruleID: EvidenceQualityRuleHypotheticalRewrite,
		},
		{
			name:   "encoding_noise",
			input:  EvidenceQualityInput{KeyUpdate: "鈥斺€斘恼伦钚路⒉际奔�:2026年7月。"},
			ruleID: EvidenceQualityRuleEncodingNoise,
		},
		{
			name:   "boilerplate_noise",
			input:  EvidenceQualityInput{KeyUpdate: "Stock quote and real-time quotes are available 24 hours a day, five days a week."},
			ruleID: EvidenceQualityRuleBoilerplateNoise,
		},
		{
			name:   "video_description_metadata",
			input:  EvidenceQualityInput{KeyUpdate: "视频简介 关注全球主要央行货币政策动向，欧洲央行行长谈风险平衡。"},
			ruleID: EvidenceQualityRuleBoilerplateNoise,
		},
		{
			name:   "mobile_news_image_cta",
			input:  EvidenceQualityInput{KeyUpdate: "打开网易新闻 查看精彩图片"},
			ruleID: EvidenceQualityRuleBoilerplateNoise,
		},
		{
			name:   "clustered_navigation_controls",
			input:  EvidenceQualityInput{KeyUpdate: "桌面版 最新搜看股票 返回 放大 + 缩小 - 传Anthropic与投资者会面 推荐2利好3利淡1"},
			ruleID: EvidenceQualityRuleBoilerplateNoise,
		},
		{
			name:   "portal_navigation_dominates_announcement",
			input:  EvidenceQualityInput{Headline: "示例公司董事会会议通告", KeyUpdate: "研究报告 定期财报 公司公告 主要行业 今日热门 本周热门 本月热门 最新推荐 精选研报 更多 >> 示例公司公告董事会将于2026年8月召开会议。"},
			ruleID: EvidenceQualityRuleBoilerplateNoise,
		},
		{
			name:   "timestamp_only",
			input:  EvidenceQualityInput{KeyUpdate: "2026-07-13 20:10"},
			ruleID: EvidenceQualityRuleDateCategoryLine,
		},
		{
			name:   "related_news_navigation",
			input:  EvidenceQualityInput{KeyUpdate: "政策利率 相关新闻"},
			ruleID: EvidenceQualityRuleBoilerplateNoise,
		},
		{
			name:   "headline_restatement",
			input:  EvidenceQualityInput{Headline: "美联储，重磅来袭！降息，突变！", KeyUpdate: "美联储，重磅来袭！降息，突变！"},
			ruleID: EvidenceQualityRuleHeadlineRestatement,
		},
		{
			name:   "headline_restatement_with_book_quote",
			input:  EvidenceQualityInput{Headline: "震惊！Anthropic发现大模型有未知潜意识空间J-space", KeyUpdate: "Anthropic发现大模型有未知潜意识空间J-space》"},
			ruleID: EvidenceQualityRuleHeadlineRestatement,
		},
		{
			name:   "headline_restatement_with_publisher_suffix",
			input:  EvidenceQualityInput{Headline: "马斯克为何改口称看错了Anthropic 态度180度急转_中华网", KeyUpdate: "马斯克为何改口称看错了Anthropic 态度180度急转。"},
			ruleID: EvidenceQualityRuleHeadlineRestatement,
		},
		{
			name:   "low_information",
			input:  EvidenceQualityInput{Headline: "美联储会议纪要即将公布", KeyUpdate: "美联储的最新动向，备受市场关注！"},
			ruleID: EvidenceQualityRuleLowInformation,
		},
		{
			name:   "editorial_major_test",
			input:  EvidenceQualityInput{Headline: "政策会议前瞻", KeyUpdate: "政策委员会即将迎来重大考验。"},
			ruleID: EvidenceQualityRuleLowInformation,
		},
		{
			name:   "generic_editorial_lead",
			input:  EvidenceQualityInput{Headline: "2026年波斯湾石油危机引发全球油价飙升50%", KeyUpdate: "家事国事天下事，事事关心。天下所关心，并不限于某一具体事件，更是其后的历史渊源与深层动机。"},
			ruleID: EvidenceQualityRuleLowInformation,
		},
		{
			name:   "customer_thanks_lead",
			input:  EvidenceQualityInput{Headline: "《CC直播》停运公告", KeyUpdate: "感谢您一直以来给予网易《CC直播》的支持与厚爱！"},
			ruleID: EvidenceQualityRuleLowInformation,
		},
		{
			name:   "metaphorical_narrative_lead",
			input:  EvidenceQualityInput{Headline: "2026年波斯湾石油危机引发全球油价飙升50%", KeyUpdate: "在这场战争里，一个幽灵盘旋在美伊战场的上方。70多年以来，它一直在中东的上空徘徊不去。"},
			ruleID: EvidenceQualityRuleLowInformation,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var decision EvidenceQualityDecision
			if tc.input.Line != "" {
				decision = policy.EvaluateLine(tc.input)
			} else {
				decision = policy.EvaluateKeyUpdate(tc.input)
			}
			if decision.Accepted || decision.RuleID != tc.ruleID {
				t.Fatalf("expected reject rule %q, got %#v", tc.ruleID, decision)
			}
		})
	}
}

func TestDefaultEvidenceQualityPolicyKeepsAttributedFactualUpdate(t *testing.T) {
	decision := DefaultEvidenceQualityPolicy().EvaluateKeyUpdate(EvidenceQualityInput{
		Headline:  "骂了大半年，马斯克突然改口",
		KeyUpdate: "马斯克表示自己此前看错Anthropic，并称其当前模型处于AI行业领先位置。",
	})
	if !decision.Accepted {
		t.Fatalf("expected attributed factual update to remain usable, got %#v", decision)
	}
}

func TestDefaultEvidenceQualityPolicyKeepsArticleWithLimitedRelatedNavigation(t *testing.T) {
	decision := DefaultEvidenceQualityPolicy().EvaluateKeyUpdate(EvidenceQualityInput{
		Headline:  "示例公司发布季度业绩",
		KeyUpdate: "示例公司7月18日发布季度业绩，收入同比增长12%；页面下方另有今日热门和更多内容入口。",
	})
	if !decision.Accepted {
		t.Fatalf("limited related-navigation text must not reject a factual article update, got %#v", decision)
	}
}

func TestDefaultEvidenceQualityPolicyScoresSpecificity(t *testing.T) {
	weak := EvidenceSpecificityScore(
		"美联储，重磅来袭！降息，突变！",
		"美联储，重磅来袭！降息，突变！",
		"美联储，重磅来袭！降息，突变！",
	)
	strong := EvidenceSpecificityScore(
		"美联储官员称仍需观察通胀数据，降息预期降温",
		"美联储官员表示仍需观察通胀数据，市场对6月降息概率下调至38%。",
		"美联储官员表示仍需观察通胀数据，市场对6月降息概率下调至38%。",
	)
	if weak >= strong {
		t.Fatalf("expected factual key update to score higher than headline restatement, weak=%d strong=%d", weak, strong)
	}
}
