package publicnews

import (
	"strings"
	"testing"
)

func TestSupportingSourceRelevantForIntentSkipsEntityOnlySupport(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:      "苹果最新产品发布：秋季发布会将覆盖多条产品线",
		SourceURL:  "https://primary.example.com/apple-product-launch.html",
		SourceSite: "primary.example.com",
		KeyUpdate:  "苹果最新产品发布计划覆盖手机、电脑、穿戴设备及智能家居等多个领域。",
		Text:       "苹果最新产品发布计划覆盖手机、电脑、穿戴设备及智能家居等多个领域。",
	}
	weak := LatestNewsLookupSource{
		Title:      "Apple security releases",
		SourceURL:  "https://support.apple.com/en-us/100100",
		SourceSite: "support.apple.com",
		KeyUpdate:  "This document lists recent security releases and CVE details.",
		Text:       "Apple security releases and recent CVE details.",
	}
	strong := LatestNewsLookupSource{
		Title:      "苹果新品发布会供应链前瞻",
		SourceURL:  "https://wire.example.net/apple-product-event.html",
		SourceSite: "wire.example.net",
		KeyUpdate:  "第二来源称苹果新品发布会将覆盖多条产品线。",
		Text:       "第二来源称苹果新品发布会将覆盖多条产品线。",
	}
	intent := LatestNewsLookupIntent{
		Topic:          "苹果最新产品发布",
		EntityMentions: []string{"苹果", "Apple"},
	}
	if SupportingSourceRelevantForIntent(primary, weak, intent) {
		t.Fatalf("expected entity-only support source to be rejected")
	}
	if !SupportingSourceRelevantForIntent(primary, strong, intent) {
		t.Fatalf("expected topic-specific support source to be accepted")
	}
	decision := DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
		Primary:   primary,
		Candidate: weak,
		Intent:    intent,
	})
	if decision.Accepted || decision.RuleID != SourceRelevanceRuleTopicSpecificSupportNeeded {
		t.Fatalf("expected structured rejection rule, got %#v", decision)
	}
}

func TestSupportingSourceRelevantForIntentRejectsBodyOnlyTopicEvidence(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "政策委员会公布利率决定",
		KeyUpdate: "政策委员会宣布维持利率不变，并继续观察通胀数据。",
		Text:      "政策委员会宣布维持利率不变。",
	}
	candidate := LatestNewsLookupSource{
		Title:     "另一经济体加息将至",
		KeyUpdate: "另一经济体央行即将议息，加息幅度与节奏仍存在分歧。",
		Text:      "背景部分提到政策委员会利率决定，但本文事件是另一经济体加息。",
	}
	intent := LatestNewsLookupIntent{Topic: "政策委员会利率决定"}
	if SupportingSourceRelevantForIntent(primary, candidate, intent) {
		t.Fatalf("expected body-only topic overlap not to count as cross-check evidence")
	}
}

func TestSupportingSourceRelevantForIntentRequiresSameEvent(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "美联储理事沃勒谈前瞻性指引",
		KeyUpdate: "美联储理事沃勒表示，必要时可以不使用前瞻性指引。",
		Text:      "美联储理事沃勒谈前瞻性指引。",
	}
	sameEvent := LatestNewsLookupSource{
		Title:     "沃勒：前瞻指引并非越多越好",
		KeyUpdate: "另一来源报道沃勒对前瞻指引的最新表态。",
		Text:      "另一来源报道沃勒对前瞻指引的最新表态。",
	}
	differentEvent := LatestNewsLookupSource{
		Title:     "沃什国会证词将谈通胀与改革",
		KeyUpdate: "美联储主席沃什将在国会发表证词，讨论通胀和改革计划。",
		Text:      "美联储主席沃什将在国会发表证词。",
	}
	intent := LatestNewsLookupIntent{Topic: "美联储利率政策"}
	if !SupportingSourceRelevantForIntent(primary, sameEvent, intent) {
		t.Fatalf("expected reports about the same Waller event to cross-check")
	}
	decision := DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
		Primary:   primary,
		Candidate: differentEvent,
		Intent:    intent,
	})
	if decision.Accepted || decision.RuleID != SourceRelevanceRuleEventCoherenceNeeded {
		t.Fatalf("expected different Fed events to fail event coherence, got %#v", decision)
	}
}

func TestSupportingSourceRelevantForIntentRejectsRealizedDecisionVersusFuturePrediction(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "Federal Reserve Holds Interest Rates Steady as Inflation Progress Remains Gradual",
		KeyUpdate: "The U.S. Federal Reserve has announced its decision to maintain current interest rates, citing a need for more consistent evidence that inflation is moving toward its 2% target.",
	}
	candidate := LatestNewsLookupSource{
		Title:     "Will the Fed Raise Rates After the July 2026 Meeting?",
		KeyUpdate: "Prediction markets are assigning overwhelming probability to the Fed holding rates steady at the upcoming July 28-29 meeting.",
	}
	decision := DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
		Primary:   primary,
		Candidate: candidate,
		Intent: LatestNewsLookupIntent{
			Topic:          "Federal Reserve interest rate policy",
			EntityMentions: []string{"Federal Reserve", "Fed"},
		},
	})
	if decision.Accepted || decision.RuleID != SourceRelevanceRuleEventCoherenceNeeded {
		t.Fatalf("expected a future prediction not to cross-check a realized decision, got %#v", decision)
	}
}

func TestSupportingSourceRelevantForIntentRejectsMarketMoveVersusShareBuyback(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "小米集团-W午前涨逾6% MiMo端侧模型通过备案发布具身生成大模型U0",
		KeyUpdate: "小米集团-W（01810）午前涨逾6%，截至发稿，股价上涨5.96%，现报27.40港元，成交额41.29亿港元。",
	}
	candidate := LatestNewsLookupSource{
		Title:     "小米集团-W(01810.HK)7月17日回购1.02亿港元，年内累计回购111.66亿港元",
		KeyUpdate: "证券时报·数据宝统计，小米集团-W在港交所公告显示，7月17日以每股26.880港元至27.000港元的价格回购380.00万股，回购金额达1.02亿港元。",
	}
	intent := LatestNewsLookupIntent{
		Topic:          "小米集团-W",
		EntityMentions: []string{"小米集团-W"},
	}
	decision := DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
		Primary:   primary,
		Candidate: candidate,
		Intent:    intent,
	})
	if decision.Accepted || decision.RuleID != SourceRelevanceRuleEventCoherenceNeeded {
		intentText := strings.ToLower(strings.Join(append([]string{intent.Topic}, intent.EntityMentions...), " "))
		t.Fatalf("expected a share-buyback story not to cross-check a market move/model story, got %#v; primary=%v candidate=%v",
			decision,
			eventFingerprintTokens(primary.KeyUpdate, intentText),
			eventFingerprintTokens(candidate.KeyUpdate, intentText),
		)
	}
}

func TestEventSourceClaimActualityTreatsRealizedResultAsRealizedWhenExpected(t *testing.T) {
	source := LatestNewsLookupSource{
		Title:     "Policy committee officially announced its rate decision",
		KeyUpdate: "The committee has announced its decision to hold rates, as analysts expected it to do.",
	}
	if got := eventSourceClaimActuality(source); got != eventClaimActualityRealized {
		t.Fatalf("expected explicit realization to win over expectation context, got %v", got)
	}
}

func TestSupportingSourceRelevantForIntentRejectsDifferentCompanyEvent(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "Anthropic旗下最强模型解禁在即，同日新一代主力模型发布_10%公司_澎湃新闻-The Paper",
		KeyUpdate: "在被“封禁”不到三周后，AI（人工智能）初创公司Anthropic旗下最先进模型即将重新上架。",
	}
	candidate := LatestNewsLookupSource{
		Title:     "美国AI企业Anthropic估值9650亿美元登顶全球，为何能在3个月内实现估值翻2.5倍？",
		KeyUpdate: "美国AI企业Anthropic完成H轮650亿美元融资，投后估值达9650亿美元，正式登顶全球最贵AI公司宝座，较3个月前估值翻涨2.5倍。",
	}
	decision := DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
		Primary:   primary,
		Candidate: candidate,
		Intent: LatestNewsLookupIntent{
			Topic:          "Anthropic",
			EntityMentions: []string{"Anthropic"},
		},
	})
	if decision.Accepted || decision.RuleID != SourceRelevanceRuleEventCoherenceNeeded {
		intentText := "anthropic anthropic"
		t.Fatalf("expected unrelated financing event not to cross-check model relisting, got %#v; primary=%#v candidate=%#v",
			decision,
			eventFingerprintTokens(primary.KeyUpdate, intentText),
			eventFingerprintTokens(candidate.KeyUpdate, intentText),
		)
	}
}

func TestSupportingSourceRelevantForIntentRejectsSameActionOnDifferentDates(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "Keep：7月17日回购83,800股股份",
		KeyUpdate: "Keep公告称，7月17日，公司回购83,800股股份，每股最高购回价2.03港元。",
	}
	candidate := LatestNewsLookupSource{
		Title:     "KEEP7月16日斥资19.27万港元回购9.88万股",
		KeyUpdate: "KEEP发布公告，于2026年7月16日斥资19.27万港元回购9.88万股。",
	}
	decision := DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
		Primary:   primary,
		Candidate: candidate,
		Intent: LatestNewsLookupIntent{
			Topic:          "Keep",
			EntityMentions: []string{"Keep"},
		},
	})
	if decision.Accepted || decision.RuleID != SourceRelevanceRuleEventCoherenceNeeded {
		t.Fatalf("expected adjacent buybacks on different dates to remain separate events, got %#v", decision)
	}
}

func TestEventDateAnchorsIgnoreDecimalPrices(t *testing.T) {
	if anchors := eventDateAnchors("每股回购价格2.03-1.88港元，金额16.00万港元。"); len(anchors) != 0 {
		t.Fatalf("expected decimal prices not to become event dates, got %#v", anchors)
	}
	anchors := eventDateAnchors("公告日期为2026-07-17，每股价格2.03港元。")
	if len(anchors) != 1 || !anchors["07-17"] {
		t.Fatalf("expected full numeric date to remain available, got %#v", anchors)
	}
}

func TestSupportingSourceRelevantForIntentRejectsDifferentRecurringQuantitiesWithoutDates(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "Keep完成股份回购",
		KeyUpdate: "Keep回购83,800股，耗资16万港元；回购后已发行股份总数为498,607,087股。",
	}
	candidate := LatestNewsLookupSource{
		Title:     "KEEP公布股份回购",
		KeyUpdate: "Keep回购98,800股，耗资18万港元；回购后已发行股份总数为498,607,087股。",
	}
	decision := DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
		Primary:   primary,
		Candidate: candidate,
		Intent: LatestNewsLookupIntent{
			Topic:          "Keep",
			EntityMentions: []string{"Keep"},
		},
	})
	if decision.Accepted || decision.RuleID != SourceRelevanceRuleEventCoherenceNeeded {
		t.Fatalf("expected different no-date buybacks to remain separate events, got %#v", decision)
	}
}

func TestEventQuantityAnchorsNormalizeEquivalentShareCounts(t *testing.T) {
	primary := eventQuantityAnchors("公司回购83,800股，另有498,607,087股已发行股份。")
	candidate := eventQuantityAnchors("公司本次回购8.38万股。")
	if eventQuantityAnchorsConflict("公司回购83,800股。", "公司本次回购8.38万股。") {
		t.Fatalf("equivalent share counts must not conflict: primary=%#v candidate=%#v", primary, candidate)
	}
	if !eventQuantityValuesOverlap(primary["shares"], candidate["shares"]) {
		t.Fatalf("expected normalized share-count overlap: primary=%#v candidate=%#v", primary, candidate)
	}
	if len(primary["shares"]) != 1 {
		t.Fatalf("issued share capital must not become an event anchor: %#v", primary)
	}
}

func TestEventQuantityAnchorsIgnorePricesAndUnqualifiedNumbers(t *testing.T) {
	anchors := eventQuantityAnchors("每股最高价2.03港元，最低价1.88港元，公告编号20260718。")
	if len(anchors) != 0 {
		t.Fatalf("prices and unqualified numbers must not become event quantity anchors: %#v", anchors)
	}
}

func TestSupportingSourceRelevantForIntentRejectsDifferentAnthropicEventsWithSharedAIContext(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "Anthropic宣布将Claude Fable 5免费使用期限延长至2026年7月19日",
		KeyUpdate: "延期政策确认：据Android Authority于2026年7月13日报道，人工智能公司Anthropic通过社交平台X宣布，将旗下旗舰模型Claude Fable 5的免费使用期延长至2026年7月19日。",
	}
	candidate := LatestNewsLookupSource{
		Title:     "Anthropic警示未授权股份交易平台",
		KeyUpdate: "眼下各类投资者都争相抢购各行各业人工智能企业的股份，Anthropic本周在官网发布警示：一大批私募及二级市场投资平台号称可交易该公司股份，但实际上均未获得官方授权。",
	}
	intent := LatestNewsLookupIntent{
		Topic:          "Anthropic",
		EntityMentions: []string{"Anthropic"},
	}
	decision := DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
		Primary:   primary,
		Candidate: candidate,
		Intent:    intent,
	})
	if decision.Accepted || decision.RuleID != SourceRelevanceRuleEventCoherenceNeeded {
		intentText := "anthropic anthropic"
		t.Fatalf("expected unrelated product-access and equity-warning events to differ, got %#v; primary=%#v candidate=%#v",
			decision,
			eventFingerprintTokens(primary.KeyUpdate, intentText),
			eventFingerprintTokens(candidate.KeyUpdate, intentText),
		)
	}
}

func TestSupportingSourceRelevantForIntentRejectsBrokerValuationAndProductLaunchEvents(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "券商下调示例公司目标价并评估智能投资回报",
		KeyUpdate: "券商研报称，看好示例公司智能投入的投资回报率，预计新智能体、模型升级和云业务增长可能成为估值修复催化剂。",
	}
	candidate := LatestNewsLookupSource{
		Title:     "示例公司在行业大会发布具身智能全栈方案",
		KeyUpdate: "7月18日，示例公司在行业大会宣布一系列具身智能和智能体技术产品与服务更新。",
	}
	intent := LatestNewsLookupIntent{
		Topic:          "示例公司",
		EntityMentions: []string{"示例公司"},
	}
	decision := DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
		Primary:   primary,
		Candidate: candidate,
		Intent:    intent,
	})
	if decision.Accepted || decision.RuleID != SourceRelevanceRuleEventCoherenceNeeded {
		t.Fatalf("expected broker valuation and later product launch to remain different events, got %#v", decision)
	}
}

func TestSupportingSourceRelevantForIntentRejectsIPOAndEquityWarningAsDifferentEvents(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "Anthropic据悉推进IPO筹备 投行正安排投资者未来几周会面",
		KeyUpdate: "财联社7月16日讯，最新消息显示，美国人工智能企业Anthropic正寻求在潜在的大规模IPO前与投资者会面，为加入由人工智能热潮推动的上市潮做准备。",
	}
	candidate := LatestNewsLookupSource{
		Title:     "Anthropic警示未授权股份交易平台",
		KeyUpdate: "眼下各类投资者都争相抢购各行各业人工智能企业的股份，Anthropic本周在官网发布警示：一大批私募及二级市场投资平台号称可交易该公司股份，但实际上均未获得官方授权。",
	}
	intent := LatestNewsLookupIntent{
		Topic:          "Anthropic",
		EntityMentions: []string{"Anthropic"},
	}
	decision := DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
		Primary:   primary,
		Candidate: candidate,
		Intent:    intent,
	})
	if decision.Accepted || decision.RuleID != SourceRelevanceRuleEventCoherenceNeeded {
		intentText := "anthropic anthropic"
		t.Fatalf("expected IPO preparation and unauthorized-equity warning to differ, got %#v; primary=%#v candidate=%#v",
			decision,
			eventFingerprintTokens(primary.KeyUpdate, intentText),
			eventFingerprintTokens(candidate.KeyUpdate, intentText),
		)
	}
}

func TestSupportingSourceRelevantForIntentUsesKeyUpdateBeforeBroadHeadline(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "美联储表态与金价波动",
		KeyUpdate: "美联储将公布会议纪要，投资者将关注官员对通胀的分歧。",
	}
	candidate := LatestNewsLookupSource{
		Title:     "黄金市场担忧美联储加息",
		KeyUpdate: "中东冲突推高通胀担忧，黄金交易员预计高利率可能维持更久。",
	}
	decision := DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
		Primary:   primary,
		Candidate: candidate,
		Intent:    LatestNewsLookupIntent{Topic: "美联储利率政策"},
	})
	if decision.Accepted || decision.RuleID != SourceRelevanceRuleEventCoherenceNeeded {
		t.Fatalf("expected broad headline overlap not to hide different key events, got %#v", decision)
	}
}

func TestHeadlineEventEvidenceCoherentRejectsMarketLeadAndKeepsProductEvent(t *testing.T) {
	intent := LatestNewsLookupIntent{
		Topic:          "腾讯控股",
		EntityMentions: []string{"腾讯控股"},
	}
	headline := "腾讯控股早盘涨近5% 日前发布新一代大模型混元Hy3"
	if HeadlineEventEvidenceCoherent(headline, "腾讯控股盘中涨近6%，现报474.20港元。", intent) {
		t.Fatal("expected market-context lead to be insufficient as the product-event anchor")
	}
	if !HeadlineEventEvidenceCoherent(headline, "腾讯日前正式发布新一代大模型混元Hy3，并披露智能体任务能力提升。", intent) {
		t.Fatal("expected Hy3 release sentence to cohere with the headline event")
	}
}

func TestTextMatchesLatestNewsIntentAcceptsBoundedCompanyAliases(t *testing.T) {
	tests := []struct {
		name   string
		entity string
		text   string
		want   bool
	}{
		{name: "Chinese holding suffix", entity: "腾讯控股", text: "腾讯混元Hy3正式发布。", want: true},
		{name: "Chinese legal suffix", entity: "中国银行股份有限公司", text: "中国银行发布年度业绩。", want: true},
		{name: "English corporation suffix", entity: "Microsoft Corporation", text: "Microsoft announced a product update.", want: true},
		{name: "unrelated company", entity: "腾讯控股", text: "阿里巴巴集团发布季度业绩。", want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := TextMatchesLatestNewsIntent(test.text, LatestNewsLookupIntent{
				Topic:          test.entity,
				EntityMentions: []string{test.entity},
			})
			if got != test.want {
				t.Fatalf("TextMatchesLatestNewsIntent(%q, %q)=%v, want %v", test.text, test.entity, got, test.want)
			}
		})
	}
}

func TestSupportingSourceRelevantForIntentRejectsMirroredEvidenceCopy(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "政策会议纪要显示内部分歧",
		SourceURL: "https://m.publisher.example/article/123",
		KeyUpdate: "政策委员会会议纪要显示，成员对下一步加息存在显著分歧。",
	}
	candidate := LatestNewsLookupSource{
		Title:     "政策会议纪要显示内部分歧",
		SourceURL: "https://www.publisher.example/news/123",
		KeyUpdate: "政策委员会会议纪要显示，成员对下一步加息存在显著分歧。",
	}
	decision := DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
		Primary:   primary,
		Candidate: candidate,
		Intent:    LatestNewsLookupIntent{Topic: "政策委员会利率政策"},
	})
	if decision.Accepted || decision.RuleID != SourceRelevanceRuleDuplicateEvidenceCopy {
		t.Fatalf("expected mirrored evidence to be rejected, got %#v", decision)
	}
}

func TestSupportingSourceRelevantForIntentIgnoresSharedEvidenceLabels(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "欧洲央行施纳贝尔表态后市场押注7月行动",
		SourceURL: "https://primary.example.com/ecb-policy",
		KeyUpdate: "核心事件：欧洲央行委员施纳贝尔6月讲话改变政策讨论方向，议题转为政策需收紧至何种程度、维持多久。",
	}
	candidate := LatestNewsLookupSource{
		Title:     "欧洲央行7月加息预期随能源价格降温消退",
		SourceURL: "https://independent.example.net/ecb-energy",
		KeyUpdate: "核心事件：能源价格意外快速回落，欧洲央行7月会议的加息争议基本平息，市场关注点转向9月。",
	}
	intent := LatestNewsLookupIntent{
		Topic:          "欧洲央行政策对市场影响",
		EntityMentions: []string{"欧洲央行", "ECB"},
	}
	comparable, coherent := eventEvidenceCoherence(primary, candidate, intent)
	if !comparable || coherent {
		t.Fatalf("expected shared evidence labels not to make different events coherent, comparable=%t coherent=%t", comparable, coherent)
	}
	decision := DefaultSourceRelevancePolicy().EvaluateSupportingSource(SourceRelevanceInput{
		Primary:   primary,
		Candidate: candidate,
		Intent:    intent,
	})
	if decision.Accepted {
		t.Fatalf("expected different ECB policy events to be rejected after label removal, got %#v", decision)
	}
}

func TestSupportingSourceRelevantForIntentRejectsDifferentECBPolicyAnalysis(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "为避免通胀收尾代价飙升，欧洲央行或采用机会主义通缩放缓达标节奏",
		SourceURL: "https://primary.example.com/ecb-opportunistic-disinflation",
		KeyUpdate: "上世纪90年代美联储学者提出的机会主义式通缩框架，恰好适配欧元区当下增长低迷、财政承压、通胀小幅超标的环境，欧洲央行或隐性采用该思路，减少加息幅度、拉长紧缩周期。",
	}
	candidate := LatestNewsLookupSource{
		Title:     "欧央行官员：欧洲央行可能不再需要加息",
		SourceURL: "https://independent.example.net/ecb-stournaras",
		KeyUpdate: "格隆汇7月2日，欧洲央行管委斯图纳拉斯表示，随着能源价格意外下跌以及欧元区通胀放缓，央行可能不需要在6月加息后继续收紧政策。",
	}
	intent := LatestNewsLookupIntent{
		Topic:          "欧洲央行政策对市场影响",
		EntityMentions: []string{"欧洲央行", "ECB", "European Central Bank"},
	}
	comparable, coherent := eventEvidenceCoherence(primary, candidate, intent)
	if !comparable || coherent {
		t.Fatalf("expected separate ECB analysis and official-comment events to be incoherent; primary=%v candidate=%v", eventFingerprintTokens(primary.KeyUpdate, strings.ToLower(strings.Join(append([]string{intent.Topic}, intent.EntityMentions...), " "))), eventFingerprintTokens(candidate.KeyUpdate, strings.ToLower(strings.Join(append([]string{intent.Topic}, intent.EntityMentions...), " "))))
	}
}

func TestSupportingSourceRelevantForIntentRejectsDifferentECBPolicyDates(t *testing.T) {
	primary := LatestNewsLookupSource{
		Title:     "欧洲央行会议纪要：即使预计加息三次，通胀仍将高于目标水平直至明年",
		SourceURL: "https://primary.example.com/ecb-minutes",
		KeyUpdate: "7月9日，欧洲央行公布会议记录，预测显示即使进行近三次加息，通胀率仍将高于目标水平直至明年。",
	}
	candidate := LatestNewsLookupSource{
		Title:     "欧洲央行行长：通胀上行和经济下行的风险更趋平衡",
		SourceURL: "https://independent.example.net/ecb-risk-balance",
		KeyUpdate: "7月2日，欧洲央行行长表示，通胀上行和经济下行的风险更趋平衡。",
	}
	intent := LatestNewsLookupIntent{
		Topic:          "欧洲央行政策对市场影响",
		EntityMentions: []string{"欧洲央行", "ECB", "European Central Bank"},
	}
	comparable, coherent := eventEvidenceCoherence(primary, candidate, intent)
	if !comparable || coherent {
		intentText := strings.ToLower(strings.Join(append([]string{intent.Topic}, intent.EntityMentions...), " "))
		t.Fatalf("expected different ECB policy dates and actions to be incoherent; primary=%v candidate=%v", eventFingerprintTokens(primary.KeyUpdate, intentText), eventFingerprintTokens(candidate.KeyUpdate, intentText))
	}
}

func TestTopicSpecificIntentTokensAvoidEntityGenericBoundaryNoise(t *testing.T) {
	tokens := TopicSpecificIntentTokens(LatestNewsLookupIntent{
		Topic:          "比特币市场",
		EntityMentions: []string{"比特币", "BTC"},
	})
	if len(tokens) != 0 {
		t.Fatalf("expected entity plus generic market topic to have no specific tokens, got %#v", tokens)
	}
	tokens = TopicSpecificIntentTokens(LatestNewsLookupIntent{
		Topic:          "苹果最新产品发布",
		EntityMentions: []string{"苹果", "Apple"},
	})
	if !containsIntentToken(tokens, "产品") || !containsIntentToken(tokens, "发布") {
		t.Fatalf("expected product/release tokens after removing entity and generic terms, got %#v", tokens)
	}
}

func TestDefaultSourceRelevancePolicyScoresSourceForIntent(t *testing.T) {
	score := DefaultSourceRelevancePolicy().ScoreSourceForIntent(LatestNewsLookupSource{
		Title:     "苹果产品发布会供应链前瞻",
		KeyUpdate: "第二来源称苹果产品发布会将覆盖多条产品线。",
		Text:      "苹果产品发布会。",
	}, LatestNewsLookupIntent{
		Topic:          "苹果最新产品发布",
		EntityMentions: []string{"苹果", "Apple"},
	})
	if score < 4 {
		t.Fatalf("expected topic-specific source score, got %d", score)
	}
}

func TestLatestNewsIntentSubjectAnchorsRemoveFacetTerms(t *testing.T) {
	cases := []struct {
		intent LatestNewsLookupIntent
		want   string
	}{
		{intent: LatestNewsLookupIntent{Topic: "美联储利率政策"}, want: "美联储"},
		{intent: LatestNewsLookupIntent{Topic: "苹果最新产品发布"}, want: "苹果"},
		{intent: LatestNewsLookupIntent{Topic: "Federal Reserve interest rate policy"}, want: "federalreserve"},
		{intent: LatestNewsLookupIntent{Topic: "OpenAI latest product release"}, want: "openai"},
	}
	for _, tc := range cases {
		anchors := LatestNewsIntentSubjectAnchors(tc.intent)
		if len(anchors) != 1 || anchors[0] != tc.want {
			t.Fatalf("anchors(%q)=%#v, want %q", tc.intent.Topic, anchors, tc.want)
		}
	}
}

func TestTextMatchesLatestNewsIntentRequiresSubjectNotFacetOnly(t *testing.T) {
	intent := LatestNewsLookupIntent{Topic: "美联储利率政策"}
	if TextMatchesLatestNewsIntent("韩国央行重申政策利率需要在适当时候上调。", intent) {
		t.Fatalf("expected policy/rate facet-only evidence to be rejected")
	}
	if !TextMatchesLatestNewsIntent("美联储会议纪要显示内部对加息分歧显著。", intent) {
		t.Fatalf("expected subject-specific evidence to match")
	}
	aliasIntent := LatestNewsLookupIntent{Topic: "苹果产品发布", EntityMentions: []string{"苹果", "Apple"}}
	if !TextMatchesLatestNewsIntent("Apple announced a new product event.", aliasIntent) {
		t.Fatalf("expected an explicit alias to match")
	}
}

func containsIntentToken(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
