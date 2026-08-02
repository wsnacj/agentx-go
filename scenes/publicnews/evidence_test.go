package publicnews

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func groundedContext() Context {
	return Context{
		UserMessage: "[skill:latest-news-brief] 帮我看下某事件的最新新闻，整理简短摘要并带来源。",
		Title:       "示例事件最新进展",
		SourceURL:   "https://news.example.com/latest-update.html",
		Text:        "发布时间: 2026-04-22 07:30 UTC\n示例事件最新进展：双方于今日发布新声明，局势较前一日出现变化。",
	}
}

func TestBuildGuardPayloadPassesWithGroundedCrossCheck(t *testing.T) {
	payload := BuildGuardPayload(groundedContext(), map[string]any{
		"headline":     "示例事件最新进展",
		"source_url":   "https://news.example.com/latest-update.html",
		"published_at": "2026-04-22 07:30 UTC",
		"key_update":   "示例事件最新进展：双方于今日发布新声明，局势较前一日出现变化。",
		"source_count": 2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "第二来源确认示例事件最新进展",
				"source_url":   "https://wire.example.net/latest-update.html",
				"published_at": "2026-04-22 07:45 UTC",
				"key_update":   "第二来源确认双方发布新声明。",
				"text":         "发布时间: 2026-04-22 07:45 UTC\n第二来源确认示例事件最新进展：第二来源确认双方发布新声明。",
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "passed" || !payload.CrossCheckReady || payload.ObservedSourceCount != 2 {
		t.Fatalf("expected grounded two-source guard pass, got %#v", payload)
	}
	if payload.Evaluation == nil || !payload.Evaluation.Passed {
		t.Fatalf("expected passing pack evaluation, got %#v", payload.Evaluation)
	}
}

func TestBuildGuardPayloadRejectsStalePrimaryAndDifferentSupportingEvent(t *testing.T) {
	payload := BuildGuardPayload(Context{
		UserMessage: "帮我查 Anthropic 最近一周有什么重要新闻。",
		Title:       "Anthropic旗下最强模型解禁在即，同日新一代主力模型发布",
		SourceURL:   "https://news.example.com/anthropic-model-relisting",
		Text:        "Anthropic旗下最先进模型即将重新上架，相关限制此前持续了不到三周。",
	}, map[string]any{
		"topic":           "Anthropic",
		"entity_mentions": []any{"Anthropic"},
		"freshness": map[string]any{
			"published_after": "2026-07-07T00:00:00+08:00",
		},
		"headline":     "Anthropic旗下最强模型解禁在即，同日新一代主力模型发布",
		"source_url":   "https://news.example.com/anthropic-model-relisting",
		"published_at": "2026-07-01T13:25:00+08:00",
		"key_update":   "在被封禁不到三周后，AI初创公司Anthropic旗下最先进模型即将重新上架。",
		"source_count": 2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "美国AI企业Anthropic估值9650亿美元登顶全球",
				"source_url":   "https://wire.example.net/anthropic-financing",
				"published_at": "2026-07-14T12:59:00+08:00",
				"key_update":   "美国AI企业Anthropic完成H轮650亿美元融资，投后估值达到9650亿美元。",
				"text":         "美国AI企业Anthropic完成H轮650亿美元融资，投后估值达到9650亿美元。",
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "needs_cross_check" || payload.CrossCheckReady || payload.ObservedSourceCount != 1 {
		t.Fatalf("expected stale primary and unrelated support not to pass guard, got %#v", payload)
	}
	if payload.Evaluation == nil || payload.Evaluation.FreshnessConfirmed || payload.Evaluation.Passed {
		t.Fatalf("expected freshness evaluation to fail, got %#v", payload.Evaluation)
	}
	joined := strings.Join(payload.ReviewReasons, ",")
	for _, reason := range []string{"primary_source_outside_freshness_window", "supporting_source_event_mismatch"} {
		if !strings.Contains(joined, reason) {
			t.Fatalf("expected review reason %q, got %#v", reason, payload.ReviewReasons)
		}
	}
}

func TestBuildGuardPayloadRejectsFuturePredictionAsDecisionCrossCheck(t *testing.T) {
	payload := BuildGuardPayload(Context{
		UserMessage: "查找美联储最近一周的重要政策新闻。",
		Title:       "Federal Reserve Holds Interest Rates Steady as Inflation Progress Remains Gradual",
		SourceURL:   "https://primary.example.com/fed-rate-decision",
		Text:        "The U.S. Federal Reserve has announced its decision to maintain current interest rates, citing inflation evidence.",
	}, map[string]any{
		"topic":           "Federal Reserve interest rate policy",
		"entity_mentions": []any{"Federal Reserve", "Fed"},
		"freshness": map[string]any{
			"published_after": "2026-07-12T00:00:00+08:00",
		},
		"headline":     "Federal Reserve Holds Interest Rates Steady as Inflation Progress Remains Gradual",
		"source_url":   "https://primary.example.com/fed-rate-decision",
		"published_at": "2026-07-14T23:00:00+08:00",
		"key_update":   "The U.S. Federal Reserve has announced its decision to maintain current interest rates, citing a need for more consistent evidence that inflation is moving toward its 2% target.",
		"source_count": 2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "Will the Fed Raise Rates After the July 2026 Meeting?",
				"source_url":   "https://independent.example.net/fed-july-prediction",
				"published_at": "2026-07-15T19:45:00+08:00",
				"key_update":   "Prediction markets are assigning overwhelming probability to the Fed holding rates steady at the upcoming July 28-29 meeting.",
				"text":         "Prediction markets assign a high probability to the Fed holding rates steady at the upcoming July 28-29 meeting.",
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "needs_cross_check" || payload.CrossCheckReady || payload.ObservedSourceCount != 1 {
		t.Fatalf("expected prediction evidence to be excluded from decision cross-check, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.ReviewReasons, ","), "supporting_source_event_mismatch") {
		t.Fatalf("expected structured event mismatch review reason, got %#v", payload.ReviewReasons)
	}
}

func TestBuildGuardPayloadRejectsStaleSupportingSource(t *testing.T) {
	payload := BuildGuardPayload(Context{
		UserMessage: "帮我查示例公司最近一周的重要新闻。",
		Title:       "示例公司发布新产品",
		SourceURL:   "https://news.example.com/new-product",
		Text:        "示例公司于本周发布新产品，并开始邀请客户参与测试。",
	}, map[string]any{
		"topic": "示例公司",
		"freshness": map[string]any{
			"published_after": "2026-07-07T00:00:00+08:00",
		},
		"headline":     "示例公司发布新产品",
		"source_url":   "https://news.example.com/new-product",
		"published_at": "2026-07-12T09:00:00+08:00",
		"key_update":   "示例公司于本周发布新产品，并开始邀请客户参与测试。",
		"source_count": 2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "合作方确认示例公司产品测试",
				"source_url":   "https://wire.example.net/new-product",
				"published_at": "2026-06-30T10:00:00+08:00",
				"key_update":   "合作方确认示例公司已经启动新产品客户测试。",
				"text":         "合作方确认示例公司已经启动新产品客户测试，并披露了第一批测试安排。",
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "needs_cross_check" || payload.CrossCheckReady || payload.ObservedSourceCount != 1 {
		t.Fatalf("expected stale supporting source not to count, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.ReviewReasons, ","), "supporting_source_outside_freshness_window") {
		t.Fatalf("expected freshness review reason, got %#v", payload.ReviewReasons)
	}
	if !strings.Contains(strings.Join(payload.Warnings, ","), "supporting_source_outside_freshness_window_ignored") {
		t.Fatalf("expected freshness warning, got %#v", payload.Warnings)
	}
}

func TestBuildGuardPayloadRejectsCrossDomainSyndicatedCopy(t *testing.T) {
	body := strings.Repeat("百望股份与行业机构签署合作框架，双方将围绕数据要素价格研究和信用体系建设推进后续工作。", 8)
	payload := BuildGuardPayload(Context{
		UserMessage: "帮我看下百望股份最新新闻",
		Title:       "百望股份：致力于构筑商业信用可信数字底座",
		SourceURL:   "https://finance.sina.com.cn/roll/example.shtml",
		Text:        "来源：上海证券报·中国证券网\n发布时间: 2026-07-07 20:26\n" + body,
	}, map[string]any{
		"headline":     "百望股份：致力于构筑商业信用可信数字底座",
		"source_url":   "https://finance.sina.com.cn/roll/example.shtml",
		"published_at": "2026-07-07 20:26",
		"key_update":   "百望股份与行业机构签署合作框架，双方将推进数据要素价格研究和信用体系建设。",
		"source_count": 2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "百望股份：致力于构筑商业信用可信数字底座",
				"source_url":   "https://mirror.eastmoney.com/news/example",
				"published_at": "2026-07-07 20:29",
				"key_update":   "百望股份与行业机构签署合作框架，双方将推进数据要素价格研究和信用体系建设。",
				"text":         "发布时间: 2026-07-07 20:29\n上海证券报\n" + body,
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "needs_cross_check" || payload.CrossCheckReady || payload.ObservedSourceCount != 1 {
		t.Fatalf("expected syndicated mirror to remain one independent source, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.ReviewReasons, ","), "supporting_source_not_independent") {
		t.Fatalf("expected source-independence review reason, got %#v", payload.ReviewReasons)
	}
	if !strings.Contains(strings.Join(payload.Warnings, ","), "supporting_source_syndicated_copy_ignored") {
		t.Fatalf("expected syndicated-copy warning, got %#v", payload.Warnings)
	}
}

func TestBuildGuardPayloadRejectsAttributedMediaSummaryAsIndependentSource(t *testing.T) {
	payload := BuildGuardPayload(Context{
		UserMessage: "查找 NVIDIA 最近一周的重要公司或监管新闻",
		Title:       "英伟达亚洲 AI 芯片客户名单调整",
		SourceURL:   "https://primary.example.com/nvidia-customers",
		Text:        "发布时间: 2026-07-14 21:30\n消息人士称英伟达调整了亚洲 AI 芯片客户名单，并加强出口合规审查。主报道补充了多个市场的审核范围和后续整改流程。",
	}, map[string]any{
		"topic":           "NVIDIA 公司监管新闻",
		"entity_mentions": []any{"NVIDIA", "英伟达"},
		"headline":        "英伟达亚洲 AI 芯片客户名单调整",
		"source_url":      "https://primary.example.com/nvidia-customers",
		"published_at":    "2026-07-14T21:30:00+08:00",
		"key_update":      "英伟达调整亚洲 AI 芯片客户名单并加强出口合规审查。",
		"source_count":    2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "英伟达大幅调整亚洲 AI 芯片客户名单",
				"source_url":   "https://summary.example.net/nvidia-customers",
				"published_at": "2026-07-14T14:40:00+08:00",
				"key_update":   "英国《金融时报》引述知情人士报道，英伟达已调整亚洲地区获授权购买 AI 芯片的客户数量。",
				"text":         "英国《金融时报》引述知情人士报道，英伟达已将亚洲地区获授权购买 AI 芯片的客户数量减少一半以上。公司建立了新的白名单，并在新加坡、马来西亚和日本加强尽职调查。未通过首次审查的客户可整改后重新申请。",
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "needs_cross_check" || payload.CrossCheckReady || payload.ObservedSourceCount != 1 {
		t.Fatalf("expected attributed summary to remain one source, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.ReviewReasons, ","), "supporting_source_not_independent") {
		t.Fatalf("expected source-independence review reason, got %#v", payload.ReviewReasons)
	}
	if !strings.Contains(strings.Join(payload.Warnings, ","), "supporting_source_attributed_republication_ignored") {
		t.Fatalf("expected attributed-republication warning, got %#v", payload.Warnings)
	}
}

func TestBuildGuardPayloadRejectsNewsCollectionsAsEventEvidence(t *testing.T) {
	primaryBody := strings.Join([]string{
		"1、英伟达确认新系统进入生产阶段",
		"https://news.example.com/story-1",
		"2、监管机构启动行业调查",
		"https://news.example.com/story-2",
		"3、合作伙伴发布网络平台",
		"https://news.example.com/story-3",
		"4、供应商开始交付新内存",
		"https://news.example.com/story-4",
		"5、公司公布机器人路线图",
		"https://news.example.com/story-5",
	}, "\n")
	supportingBody := "**速览：** 多项动态。### 1. 新系统投产 来源:媒体甲>>### 2. 行业调查 来源:媒体乙>>### 3. 网络合作 来源:媒体丙>>### 4. 内存交付 来源:媒体丁>>### 5. 机器人路线 来源:媒体戊>>"
	payload := BuildGuardPayload(Context{
		UserMessage: "查找 NVIDIA 最近一周的重要公司或监管新闻",
		Title:       "新浪英伟达热点小时报_今日实时英伟达热点速递",
		SourceURL:   "https://primary.example.com/nvidia-roundup",
		Text:        primaryBody,
	}, map[string]any{
		"topic":           "NVIDIA 公司新闻",
		"entity_mentions": []any{"NVIDIA", "英伟达"},
		"headline":        "新浪英伟达热点小时报_今日实时英伟达热点速递",
		"source_url":      "https://primary.example.com/nvidia-roundup",
		"published_at":    "2026-07-16T13:00:00+08:00",
		"key_update":      "英伟达确认新系统已经进入生产阶段，并将按计划交付。",
		"source_count":    2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "英伟达动态",
				"source_url":   "https://support.example.net/nvidia-digest",
				"published_at": "2026-07-16T12:00:00+08:00",
				"key_update":   "英伟达确认新系统已经进入生产阶段，并将按计划交付。",
				"text":         supportingBody,
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "needs_cross_check" || payload.CrossCheckReady || payload.ObservedSourceCount != 0 {
		t.Fatalf("expected roundup and digest not to count as event evidence, got %#v", payload)
	}
	reasons := strings.Join(payload.ReviewReasons, ",")
	if !strings.Contains(reasons, "primary_source_collection_surface") ||
		!strings.Contains(reasons, "supporting_source_collection_surface") {
		t.Fatalf("expected collection-surface review reasons, got %#v", payload.ReviewReasons)
	}
}

func TestBuildGuardPayloadRejectsDatedCompanyNewsIndex(t *testing.T) {
	indexBody := strings.Join([]string{
		"News from Example Corp. A wide array of company news stories.",
		"Latest",
		"Jul 08, 2026, 16:05 ET   Example Corp. announces quarterly earnings release date",
		"Jun 24, 2026, 09:00 ET   Example Corp. publishes its annual technology report",
		"Jun 10, 2026, 16:05 ET   Example Corp. announces quarterly dividend",
		"May 14, 2026, 16:07 ET   Example Corp. appoints a new board director",
		"Apr 29, 2026, 16:11 ET   Example Corp. earnings release is available",
	}, "\n\n")
	payload := BuildGuardPayload(Context{
		UserMessage: "查找 Example Corp. 最近新闻",
		Title:       "news",
		SourceURL:   "https://wire.example.com/news/example-corp/",
		Text:        indexBody,
	}, map[string]any{
		"topic":        "Example Corp.",
		"headline":     "news",
		"source_url":   "https://wire.example.com/news/example-corp/",
		"published_at": "2026-07-08T16:05:00+08:00",
		"key_update":   "News from Example Corp.",
		"source_count": 1,
	}, nil, PayloadOptions{})
	if payload.ObservedSourceCount != 0 || payload.GuardStatus != "needs_cross_check" {
		t.Fatalf("expected dated company-news index not to count as event evidence, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.ReviewReasons, ","), "primary_source_collection_surface") {
		t.Fatalf("expected collection-surface reason, got %#v", payload.ReviewReasons)
	}
}

func TestBuildGuardPayloadRejectsPortalBodyWithCleanExtractedEvent(t *testing.T) {
	portalBody := "研究报告 定期财报 公司公告 主要行业 今日热门 本周热门 本月热门 最新推荐 精选研报 更多 >> 百望股份公告董事会批准回购H股计划。客户端下载 正在加载，请稍候"
	payload := BuildGuardPayload(Context{
		UserMessage: "查找百望股份最近新闻",
		Title:       "百望股份回购H股计划",
		SourceURL:   "https://mirror.example.com/report/06657.html",
		Text:        portalBody,
	}, map[string]any{
		"topic":        "百望股份",
		"headline":     "百望股份回购H股计划",
		"source_url":   "https://mirror.example.com/report/06657.html",
		"published_at": "2026-06-26T21:59:00+08:00",
		"key_update":   "百望股份公告董事会批准回购H股计划。",
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "needs_cross_check" || payload.ObservedSourceCount != 0 {
		t.Fatalf("expected portal body not to count as primary event evidence, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.ReviewReasons, ","), "primary_source_collection_surface") {
		t.Fatalf("expected portal body collection review reason, got %#v", payload.ReviewReasons)
	}
}

func TestBuildGuardPayloadRejectsAIGeneratedPrimaryEvidence(t *testing.T) {
	payload := BuildGuardPayload(Context{
		UserMessage: "欧盟 AI 监管最近一周有什么新进展？",
		Title:       "欧盟人工智能法案出现新进展",
		SourceURL:   "https://news.example.com/eu-ai-act",
		Text:        "欧盟监管机构宣布新的人工智能法案执行安排。\n\n免责声明：本文内容由开放的智能模型自动生成，仅供参考。",
	}, map[string]any{
		"topic":        "欧盟 AI 监管",
		"headline":     "欧盟人工智能法案出现新进展",
		"source_url":   "https://news.example.com/eu-ai-act",
		"published_at": "2026-07-17T12:15:03+08:00",
		"key_update":   "欧盟监管机构宣布新的人工智能法案执行安排。",
		"source_count": 1,
	}, nil, PayloadOptions{})
	if payload.ObservedSourceCount != 0 || payload.GuardStatus != "needs_cross_check" {
		t.Fatalf("expected AI-generated primary not to count as evidence, got %#v", payload)
	}
	if payload.Evaluation == nil || payload.Evaluation.SourceAccepted {
		t.Fatalf("expected AI-generated source_accepted=false, got %#v", payload.Evaluation)
	}
	if !strings.Contains(strings.Join(payload.ReviewReasons, ","), "primary_source_ai_generated") {
		t.Fatalf("expected AI-generated review reason, got %#v", payload.ReviewReasons)
	}
}

func TestEvidenceLooksLikeSyndicatedCopyRejectsLightlyRewrittenWireBrief(t *testing.T) {
	primaryText := "人民财讯7月13日电，在第十四届CHINAFIT北京健身大会上，运动科技公司Keep首次展示基于Keepace.ai打造的自研Agent成果。据公司介绍，Keep已将十年运动科学积累封装为200余项运动Skill，支持不同能力之间的调用与嵌套。未来，Keep将持续扩展Skill能力，让Agent覆盖更多运动场景。\n\n（文章来源：证券时报网）\n\n特别声明：以上内容由用户上传并发布，本平台仅提供信息存储空间服务。"
	supportingText := "观点网讯：7月13日，运动科技公司Keep在第十四届CHINAFIT北京健身大会上，首次展示基于Keepace.ai打造的自研AI智能体成果。根据公开资料整理，Keep将十年运动科学积累封装为200余项运动Skill，支持不同能力之间的调用与嵌套。Keep方面表示，未来将持续扩展Skill能力，让AI智能体覆盖更多运动场景。\n\n免责声明：本文根据公开信息整理，不构成投资建议，使用前请核实。"
	primary := Evidence{Headline: "Keep自研AI智能体搭载200余项运动Skill", groundedText: primaryText}
	supporting := Evidence{Headline: "Keep展示自研AI智能体成果", groundedText: supportingText}
	if evidenceLooksLikeSyndicatedCopy(primary, supporting) {
		return
	}
	containment := articleShingleContainment(
		normalizedArticleContent(stripSourceIndependenceBoilerplate(primaryText)),
		normalizedArticleContent(stripSourceIndependenceBoilerplate(supportingText)),
	)
	t.Fatalf("expected lightly rewritten wire brief not to count as independent, containment=%.3f", containment)
}

func TestEvidenceLooksLikeSyndicatedCopyKeepsIndependentShortReports(t *testing.T) {
	primaryText := "示例公司在年度开发者大会发布新的企业协作产品，并披露首批客户将在三个地区参与测试。记者现场采访产品负责人后确认，测试范围暂不包含个人用户，正式价格将在后续公告中公布。该报道还核对了演示版本和上线时间。"
	supportingText := "当地监管登记显示，示例公司的新协作服务已完成数据处理备案。另一家媒体采访两名试用客户，确认他们分别在制造和物流场景使用独立部署版本；客户没有确认发布会所称的全部功能，商业化效果仍待观察。"
	primary := Evidence{Headline: "示例公司发布企业协作产品", groundedText: primaryText}
	supporting := Evidence{Headline: "客户确认示例公司协作服务开始测试", groundedText: supportingText}
	if evidenceLooksLikeSyndicatedCopy(primary, supporting) {
		containment := articleShingleContainment(normalizedArticleContent(primaryText), normalizedArticleContent(supportingText))
		t.Fatalf("expected independently reported short articles to remain distinct, containment=%.3f", containment)
	}
}

func TestEvidenceLooksLikeSyndicatedCopyRejectsMixedLengthRewrittenMirror(t *testing.T) {
	primaryText := strings.Join([]string{
		"快科技官方\n\n关注\n\n快科技7月16日消息，英伟达确认2026年全年不发布新款游戏显卡，这是三十多年来首次出现游戏显卡年度断更的局面。原定基于Rubin架构的RTX 60系列已推迟至2027年底甚至2028年。",
		"从RTX 20系到50系，英伟达保持了约两年一代的更新节奏。20系2018年发布，30系2020年，40系2022年，50系2025年初。如今这一节奏被AI浪潮彻底打破。",
		"财报信号更为明显。英伟达已将游戏业务从独立分类移除，并入边缘计算板块。边缘计算整体营收占比不到8%，而数据中心单季度收入高达752亿美元，约合人民币5107亿元。",
		"黄仁勋在投资者电话会上表示，即便想提升游戏显卡出货量，GDDR7显存供应也无法保障。AI芯片消耗了大部分显存产能，消费级显卡在供应链中只能排在后面。",
		"一块AI加速卡利润是游戏显卡的十几倍，同样一批显存产能卖给AI客户回报远高于消费市场。英伟达把资源全面倾斜到数据中心，这纯粹是商业逻辑驱动下的必然选择。",
		"玩家感到失落可以理解。从1999年GeForce 256开启GPU时代至今，游戏玩家一直是英伟达最核心的用户群。如今游戏业务在公司总营收中占比已不到8%。",
		"需要指出的是，英伟达并未完全放弃游戏市场，只是优先级发生了根本变化。RTX 60系列大概率会出，但更新周期可能从两年延长到三到四年，游戏显卡不再是英伟达的战略重心。",
		"对AMD而言这是一个机会窗口。RX 9070 XT首发时出现抢购盛况，备货远超RTX 50系。如果英伟达放慢游戏卡迭代，AMD有望用RDNA 4产品填补市场空缺。",
		"【本文结束】如需转载请务必注明出处：快科技\n\n责任编辑：红茶",
	}, "\n\n")
	supportingText := strings.Join([]string{
		"英伟达确认2026年全年不发布新款游戏显卡，这是其三十多年来首次出现游戏显卡年度断更，原定基于Rubin架构的RTX 60系列已推迟至2027年底甚至2028年。",
		"此前从RTX 20系到50系，英伟达一直保持约两年一代的更新节奏，如今这一规律已被AI浪潮彻底打破。财报信号更为明确：英伟达已将游戏业务从独立分类移除，并入边缘计算板块，该板块整体营收占比不足8%，而数据中心单季度收入高达752亿美元，约合人民币5107亿元。",
		"黄仁勋在投资者电话会上坦言，即便想提升游戏显卡出货量，GDDR7显存供应也无法保障，AI芯片消耗了大部分显存产能，消费级显卡在供应链中优先级靠后。一块AI加速卡利润是游戏显卡的十几倍，同批显存产能供给AI客户的回报远高于消费市场，英伟达将资源全面倾斜数据中心，完全是商业逻辑驱动的必然选择。",
		"玩家的失落不难理解，自1999年GeForce 256开启GPU时代以来，游戏玩家始终是英伟达核心用户群，如今游戏业务在总营收中占比已不足8%。英伟达并未完全放弃游戏市场，只是优先级发生根本变化，RTX 60系列大概率仍会推出，但更新周期可能从两年拉长至三到四年，游戏显卡不再是其战略重心。",
		"这对AMD而言是难得的机会窗口，RX 9070 XT首发便出现抢购盛况，备货量远超RTX 50系，若英伟达放慢游戏卡迭代，AMD有望凭借RDNA 4产品填补市场空缺。",
	}, "\n\n")
	primaryContent := normalizedArticleContent(primaryText)
	supportingContent := normalizedArticleContent(supportingText)
	if (len(primaryContent) <= sourceIndependenceBriefMaxContentRunes) ==
		(len(supportingContent) <= sourceIndependenceBriefMaxContentRunes) {
		t.Fatalf("fixture must straddle brief boundary: primary=%d supporting=%d", len(primaryContent), len(supportingContent))
	}
	primary := Evidence{Headline: "英伟达2026年不发布游戏显卡", groundedText: primaryText}
	supporting := Evidence{Headline: "英伟达确认2026年不发布新游戏显卡", groundedText: supportingText}
	if !evidenceLooksLikeSyndicatedCopy(primary, supporting) {
		containment := articleShingleContainment(primaryContent, supportingContent)
		t.Fatalf("expected mixed-length rewritten mirror not to count as independent, containment=%.3f", containment)
	}
}

func TestBuildGuardPayloadRejectsSharedPublisherAppBriefAcrossDomains(t *testing.T) {
	primaryText := "智通财经APP讯，阿里巴巴-W(09988)发布公告，于2026年7月7日斥资4999.3万美元回购407.6万股。\n该信息由智通财经网提供"
	supportingText := "智通财经 APP 讯，阿里巴巴-W 发布公告，于 2026 年 7 月 7 日斥资 4999.3 万美元回购 407.6 万股。\n\n智通财经 APP 讯，阿里巴巴-W(09988) 发布公告，于 2026 年 7 月 7 日斥资 4999.3 万美元回购 407.6 万股。"
	payload := BuildGuardPayload(Context{
		UserMessage: "阿里巴巴最近新闻风险怎么看？",
		Title:       "阿里巴巴-W 7月7日回购407.6万股",
		SourceURL:   "http://hkstock.cnfol.com/gangguzixun/example.shtml",
		Text:        primaryText,
	}, map[string]any{
		"topic":        "阿里巴巴-W",
		"headline":     "阿里巴巴-W 7月7日回购407.6万股",
		"source_url":   "http://hkstock.cnfol.com/gangguzixun/example.shtml",
		"published_at": "2026-07-08T18:52:00+08:00",
		"key_update":   "阿里巴巴-W于2026年7月7日斥资4999.3万美元回购407.6万股。",
		"source_count": 2,
		"supporting_sources": []any{map[string]any{
			"headline":     "阿里巴巴-W 7 月 7 日回购 407.6 万股",
			"source_url":   "https://longportapp.cn/zh-HK/news/example",
			"published_at": "2026-07-08T10:35:00+08:00",
			"key_update":   "阿里巴巴-W于2026年7月7日斥资4999.3万美元回购407.6万股。",
			"text":         supportingText,
		}},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "needs_cross_check" || payload.CrossCheckReady || payload.ObservedSourceCount != 1 {
		t.Fatalf("expected shared publisher APP brief to remain one source, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.Warnings, ","), "supporting_source_shared_upstream_publisher_ignored") {
		t.Fatalf("expected shared-publisher warning, got %#v", payload.Warnings)
	}
}

func TestBuildGuardPayloadRejectsLongAbridgedCrossDomainCopy(t *testing.T) {
	sharedBody := strings.Join([]string{
		"研究机构公布季度行业数据，示例公司出货量和市场份额继续位居首位，同行之间的份额差距较上年同期有所扩大。",
		"公司把端侧智能能力作为产品更新重点，新一代设备增加本地推理、知识管理和跨设备协同能力，并开始分批上市。",
		"基础设施业务同步发布计算平台升级方案，覆盖异构芯片调度、集群推理加速和数据中心能源效率等多个环节。",
		"刚结束的完整财年中，公司披露收入与经调整利润均实现增长，智能相关产品占收入比重继续提升。",
		"服务器项目储备较上一财年增加，管理层表示交付节奏仍取决于客户验收、供应链安排和区域监管要求。",
		"企业解决方案强调本地与云端协同，通过统一运行层把模型、算力和业务流程连接起来，并在若干行业试点。",
		"公司同时与芯片、操作系统和行业软件伙伴开展合作，希望降低企业部署推理服务时的硬件与运维成本。",
		"管理层在投资者沟通中再次说明增长路径，包括智能终端、计算基础设施和行业服务三个组成部分。",
		"这些项目仍处在不同商业化阶段，已经披露的订单、收入和产品测试结果不能直接等同于未来利润贡献。",
		"后续需要继续观察交付确认、毛利率变化、资本开支和客户集中度，才能判断相关业务对整体经营的持续影响。",
	}, "\n\n")
	primaryBody := strings.Join([]string{
		"最新盘点文章先介绍行业整体回落，并列出示例公司本季度的具体出货量、市场份额以及与第二名的差距。",
		sharedBody,
		"文章随后又加入另外两家公司的云服务和基础模型发展情况，用较长篇幅比较三条不同的产业路径。",
		"结尾把多家公司放在同一产业链中讨论，并重复评价各家公司已经把智能业务发展为主要增长方向。",
	}, "\n\n")
	supportingBody := strings.Join([]string{
		"较早发布的专题文章从示例公司的终端产品和基础设施布局切入，没有列出最新版行业份额表。",
		sharedBody,
		"文章最后概括智能相关收入和服务器储备，并称这些项目已经成为公司当前业务的重要组成部分。",
	}, "\n\n")
	containment := articleShingleContainment(
		normalizedArticleContent(primaryBody),
		normalizedArticleContent(supportingBody),
	)
	if containment < sourceIndependenceLongContainment || containment >= sourceIndependenceStrongContainment {
		t.Fatalf("invalid long-copy regression fixture containment=%.3f", containment)
	}

	payload := BuildGuardPayload(Context{
		UserMessage: "查找示例公司最近新闻",
		Title:       "行业年中盘点：示例公司终端和基础设施业务进展",
		SourceURL:   "https://primary.example.com/company-midyear",
		Text:        primaryBody,
	}, map[string]any{
		"topic":        "示例公司",
		"headline":     "行业年中盘点：示例公司终端和基础设施业务进展",
		"source_url":   "https://primary.example.com/company-midyear",
		"published_at": "2026-07-18T05:27:58+08:00",
		"key_update":   "研究机构公布季度行业数据，示例公司出货量和市场份额继续位居首位。",
		"source_count": 2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "示例公司终端和基础设施业务进展",
				"source_url":   "https://support.example.net/company-update",
				"published_at": "2026-07-14T16:11:00+08:00",
				"key_update":   "研究机构公布季度行业数据，示例公司出货量和市场份额继续位居首位。",
				"text":         supportingBody,
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "needs_cross_check" || payload.CrossCheckReady || payload.ObservedSourceCount != 1 {
		t.Fatalf("expected long abridged copy to remain one independent source, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.Warnings, ","), "supporting_source_syndicated_copy_ignored") {
		t.Fatalf("expected long-copy warning, got %#v", payload.Warnings)
	}
}

func TestBuildGuardPayloadRejectsSourcesWithSharedWireLineage(t *testing.T) {
	primaryText := strings.Repeat("油价上涨增加了欧洲央行下一次会议的不确定性。", 12) +
		"德国央行行长对路透社表示，政策制定者仍需保持警惕。"
	supportingText := strings.Join([]string{
		"市场正在重新评估欧洲央行下一次会议的利率路径。",
		"路透对74位经济学家的调查显示，多数受访者预计本次会议维持利率不变。",
		"据路透援引知情人士，政策委员会仍会讨论进一步收紧政策的理由。",
		"据路透，欧洲央行还在研究最低准备金率调整方案。",
	}, "\n")
	payload := BuildGuardPayload(Context{
		UserMessage: "欧洲央行最近一周的政策消息是什么？",
		Title:       "油价上涨令欧洲央行利率路径再添变数",
		SourceURL:   "https://primary.example.com/ecb-rates",
		Text:        primaryText,
	}, map[string]any{
		"topic":        "欧洲央行政策消息",
		"headline":     "油价上涨令欧洲央行利率路径再添变数",
		"source_url":   "https://primary.example.com/ecb-rates",
		"published_at": "2026-07-15T20:02:00+08:00",
		"key_update":   "市场正在重新评估欧洲央行下一次会议的利率路径。",
		"source_count": 2,
		"supporting_sources": []any{map[string]any{
			"headline":     "欧洲央行会议预期转向谨慎",
			"source_url":   "https://support.example.net/ecb-rates",
			"published_at": "2026-07-17T16:12:00+08:00",
			"key_update":   "市场普遍预计欧洲央行将在本次会议维持利率不变。",
			"text":         supportingText,
		}},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "needs_cross_check" || payload.CrossCheckReady || payload.ObservedSourceCount != 1 {
		t.Fatalf("expected shared Reuters lineage to remain one source, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.Warnings, ","), "supporting_source_shared_upstream_publisher_ignored") {
		t.Fatalf("expected shared-upstream warning, got %#v", payload.Warnings)
	}
}

func TestSourcesWithSharedWireMentionRemainIndependentWithOriginalReporting(t *testing.T) {
	left := LatestNewsLookupSource{Text: "本报记者采访两名政策委员，并核对会议文件。委员还对路透社表示，政策仍需保持警惕。"}
	right := LatestNewsLookupSource{Text: "路透对74位经济学家的调查显示多数人预计利率不变。据路透，委员会还在讨论准备金率。据路透援引知情人士，讨论尚无结论。"}
	if LatestNewsSourcesShareUpstreamPublisher(left, right) {
		t.Fatal("expected explicit original reporting to preserve source independence")
	}
}

func TestBuildGuardPayloadRejectsLightlyRewrittenCrossDomainBrief(t *testing.T) {
	primaryText := "人民财讯7月13日电，在第十四届CHINAFIT北京健身大会上，运动科技公司Keep首次展示基于Keepace.ai打造的自研Agent成果。据公司介绍，Keep已将十年运动科学积累封装为200余项运动Skill，支持不同能力之间的调用与嵌套。未来，Keep将持续扩展Skill能力，让Agent覆盖更多运动场景。\n\n（文章来源：证券时报网）"
	supportingText := "观点网讯：7月13日，运动科技公司Keep在第十四届CHINAFIT北京健身大会上，首次展示基于Keepace.ai打造的自研AI智能体成果。根据公开资料整理，Keep将十年运动科学积累封装为200余项运动Skill，支持不同能力之间的调用与嵌套。Keep方面表示，未来将持续扩展Skill能力，让AI智能体覆盖更多运动场景。\n\n免责声明：本文根据公开信息整理，不构成投资建议，使用前请核实。"
	payload := BuildGuardPayload(Context{
		UserMessage: "Keep最近新闻风险怎么看？",
		Title:       "Keep自研AI智能体搭载200余项运动Skill",
		SourceURL:   "https://finance.eastmoney.com/a/example.html",
		Text:        primaryText,
	}, map[string]any{
		"topic":           "Keep",
		"entity_mentions": []any{"Keep"},
		"headline":        "Keep自研AI智能体搭载200余项运动Skill",
		"source_url":      "https://finance.eastmoney.com/a/example.html",
		"published_at":    "2026-07-13T15:46:00+08:00",
		"key_update":      "人民财讯7月13日电，运动科技公司Keep首次展示基于Keepace.ai打造的自研Agent成果。",
		"source_count":    2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "Keep展示自研AI智能体成果",
				"source_url":   "https://www.guandian.cn/article/example.html",
				"published_at": "2026-07-13T15:59:00+08:00",
				"key_update":   "观点网讯：运动科技公司Keep首次展示基于Keepace.ai打造的自研AI智能体成果。",
				"text":         supportingText,
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "needs_cross_check" || payload.CrossCheckReady || payload.ObservedSourceCount != 1 {
		t.Fatalf("expected lightly rewritten cross-domain brief to remain one source, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.Warnings, ","), "supporting_source_syndicated_copy_ignored") {
		t.Fatalf("expected syndicated-copy warning, got %#v", payload.Warnings)
	}
}

func TestBuildGuardPayloadRejectsPublisherDomainUserGeneratedPage(t *testing.T) {
	payload := BuildGuardPayload(Context{
		UserMessage: "示例公司最近新闻",
		Title:       "示例公司发布新产品",
		SourceURL:   "https://finance.example.com/self-media/article",
		Text:        "示例公司宣布新产品将于8月开始测试。\n\n特别声明：以上作品内容由自媒体平台用户上传并发布，本平台仅提供信息存储空间服务。",
	}, map[string]any{
		"headline":     "示例公司发布新产品",
		"source_url":   "https://finance.example.com/self-media/article",
		"published_at": "2026-07-13T15:46:00+08:00",
		"key_update":   "示例公司宣布新产品将于8月开始测试。",
		"source_count": 1,
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "needs_cross_check" || payload.CrossCheckReady || payload.ObservedSourceCount != 0 {
		t.Fatalf("expected user-generated page not to count as editorial evidence, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.ReviewReasons, ","), "primary_source_community_surface") {
		t.Fatalf("expected community-surface review reason, got %#v", payload.ReviewReasons)
	}
}

func TestBuildGuardPayloadRejectsKnownAlternateDomainsFromSamePublisher(t *testing.T) {
	payload := BuildGuardPayload(Context{
		UserMessage: "欧洲央行最近政策消息有什么影响？",
		Title:       "欧洲央行发布政策会议纪要",
		SourceURL:   "https://finance.sina.cn/2026-07-09/policy-minutes.html",
		Text:        "欧洲央行发布7月政策会议纪要，委员表示将继续采取逐次会议决策。",
	}, map[string]any{
		"topic":           "欧洲央行政策",
		"entity_mentions": []any{"欧洲央行", "ECB"},
		"headline":        "欧洲央行发布政策会议纪要",
		"source_url":      "https://finance.sina.cn/2026-07-09/policy-minutes.html",
		"published_at":    "2026-07-09T10:00:00+08:00",
		"key_update":      "欧洲央行发布7月政策会议纪要，委员表示将继续采取逐次会议决策。",
		"source_count":    2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "欧洲央行7月政策会议纪要公布",
				"source_url":   "https://finance.sina.com.cn/world/2026-07-09/policy-minutes.shtml",
				"published_at": "2026-07-09T10:15:00+08:00",
				"key_update":   "另一篇报道确认欧洲央行公布7月政策会议纪要，并将继续逐次会议决策。",
				"text":         "另一篇报道确认欧洲央行公布7月政策会议纪要，并说明委员仍将根据新数据逐次会议作出决定。",
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "needs_cross_check" || payload.CrossCheckReady || payload.ObservedSourceCount != 1 {
		t.Fatalf("expected alternate Sina domains to remain one publisher, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.ReviewReasons, ","), "supporting_source_not_independent") {
		t.Fatalf("expected independence review reason, got %#v", payload.ReviewReasons)
	}
	if !strings.Contains(strings.Join(payload.Warnings, ","), "supporting_source_same_publisher_ignored") {
		t.Fatalf("expected same-publisher warning, got %#v", payload.Warnings)
	}
}

func TestBuildGuardPayloadPassesWithIndependentThirdSourceAfterSyndicatedCopy(t *testing.T) {
	body := strings.Repeat("示例公司发布产品更新，计划在多个市场逐步扩大测试范围并披露后续运营数据。", 8)
	payload := BuildGuardPayload(Context{
		UserMessage: "帮我看下示例公司最新新闻",
		Title:       "示例公司发布产品更新",
		SourceURL:   "https://primary.example.com/update",
		Text:        "发布时间: 2026-07-07 20:26\n" + body,
	}, map[string]any{
		"headline":     "示例公司发布产品更新",
		"source_url":   "https://primary.example.com/update",
		"published_at": "2026-07-07 20:26",
		"key_update":   "示例公司发布产品更新，并计划逐步扩大测试范围。",
		"source_count": 3,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "示例公司发布产品更新",
				"source_url":   "https://mirror.example.net/update",
				"published_at": "2026-07-07 20:29",
				"key_update":   "示例公司发布产品更新，并计划逐步扩大测试范围。",
				"text":         "转载稿\n" + body,
			},
			map[string]any{
				"headline":     "行业媒体核实示例公司测试安排",
				"source_url":   "https://independent.example.org/report",
				"published_at": "2026-07-07 21:10",
				"key_update":   "行业媒体独立采访两名合作方，确认测试已在三个地区启动。",
				"text":         strings.Repeat("行业媒体独立采访两名合作方，并核对公开登记信息，确认测试已在三个地区启动且仍处于早期阶段。", 8),
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "passed" || !payload.CrossCheckReady || payload.ObservedSourceCount != 2 {
		t.Fatalf("expected a genuinely independent third source to satisfy cross-check, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.Warnings, ","), "supporting_source_syndicated_copy_ignored") {
		t.Fatalf("expected ignored mirror to remain observable, got %#v", payload.Warnings)
	}
}

func TestBuildGuardPayloadRejectsUngroundedModelFields(t *testing.T) {
	payload := BuildGuardPayload(Context{}, map[string]any{
		"headline":     "示例事件最新进展",
		"source_url":   "https://news.example.com/latest-update.html",
		"published_at": "2026-04-22 07:30 UTC",
		"key_update":   "示例事件最新进展：双方于今日发布新声明，局势较前一日出现变化。",
		"source_count": 1,
	}, nil, PayloadOptions{})
	if payload.NewsFieldsReady || payload.GuardStatus != "missing_news_fields" {
		t.Fatalf("expected ungrounded fields to fail, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.MissingNewsFields, ","), "grounded_page_text") {
		t.Fatalf("expected grounded_page_text missing field, got %#v", payload.MissingNewsFields)
	}
	reasons := strings.Join(payload.ReviewReasons, ",")
	if !strings.Contains(reasons, "no_usable_source") || strings.Contains(reasons, "single_source_only") {
		t.Fatalf("expected zero usable sources to remain distinct from a single source, got %#v", payload.ReviewReasons)
	}
}

func TestBuildGuardPayloadDoesNotReportFreshnessFailureWithoutPrimaryEvidence(t *testing.T) {
	payload := BuildGuardPayload(Context{}, map[string]any{
		"freshness": map[string]any{
			"published_after": "2026-07-12T00:00:00+08:00",
		},
	}, nil, PayloadOptions{})
	reasons := strings.Join(payload.ReviewReasons, ",")
	if !strings.Contains(reasons, "no_usable_source") {
		t.Fatalf("expected explicit no-source reason, got %#v", payload.ReviewReasons)
	}
	if strings.Contains(reasons, "single_source_only") || strings.Contains(reasons, "primary_source_outside_freshness_window") {
		t.Fatalf("zero-source guard must not invent single-source or freshness claims, got %#v", payload.ReviewReasons)
	}
}

func TestBuildGuardPayloadRejectsLowQualityKeyUpdate(t *testing.T) {
	payload := BuildGuardPayload(Context{
		UserMessage: "帮我看下某事件最新新闻",
		Title:       "示例事件最新新闻",
		SourceURL:   "https://news.example.com/topic",
		Text:        "发布时间: 2026-05-15 09:00 UTC\nAll\n更多导航内容",
	}, map[string]any{
		"published_at": "2026-05-15T09:00:00",
		"source_count": 2,
		"supporting_sources": []any{
			map[string]any{
				"source_url":   "https://wire.example.net/a",
				"published_at": "2026-05-15T09:10:00",
				"key_update":   "第二来源确认示例事件出现最新进展。",
				"text":         "第二来源确认示例事件出现最新进展。",
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "missing_news_fields" || payload.NewsFieldsReady {
		t.Fatalf("expected low-quality key_update to fail guard, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.MissingNewsFields, ","), "key_update") {
		t.Fatalf("expected key_update missing field, got %#v", payload.MissingNewsFields)
	}
}

func TestBuildGuardPayloadRejectsTimestampAndNavigationOnlyEvidence(t *testing.T) {
	payload := BuildGuardPayload(Context{
		UserMessage: "帮我看下政策事件最新新闻",
		Title:       "宏观政策速递",
		SourceURL:   "https://news.example.com/subject/7381",
		Text:        "宏观政策速递\n全部内容\n2026-07-13 20:10\n阅 65.85W",
	}, map[string]any{
		"published_at": "2026-07-13T20:10:00",
		"key_update":   "2026-07-13 20:10",
		"source_count": 2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "政策信息披露平台",
				"source_url":   "https://wire.example.net/article/kx-tag-detail.html",
				"published_at": "2026-07-13T20:20:00",
				"key_update":   "政策利率 相关新闻",
				"text":         "政策利率 相关新闻\n没有更多了...\n热门股票\n最新价",
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "missing_news_fields" || payload.NewsFieldsReady {
		t.Fatalf("expected timestamp-only primary evidence to fail guard, got %#v", payload)
	}
	if payload.CrossCheckReady || payload.ObservedSourceCount != 0 {
		t.Fatalf("expected navigation-only supporting evidence to be excluded, got %#v", payload)
	}
}

func TestBuildGuardPayloadRejectsEncodingCorruptKeyUpdate(t *testing.T) {
	payload := BuildGuardPayload(groundedContext(), map[string]any{
		"headline":     "示例事件最新进展",
		"source_url":   "https://news.example.com/latest-update.html",
		"published_at": "2026-04-22 07:30 UTC",
		"key_update":   "鈥斺€斘恼伦钚路⒉际奔�:2026年7月。",
		"source_count": 2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "第二来源确认示例事件最新进展",
				"source_url":   "https://wire.example.net/latest-update.html",
				"published_at": "2026-04-22 07:45 UTC",
				"key_update":   "第二来源确认双方发布新声明。",
				"text":         "发布时间: 2026-04-22 07:45 UTC\n第二来源确认示例事件最新进展：第二来源确认双方发布新声明。",
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "missing_news_fields" || payload.NewsFieldsReady {
		t.Fatalf("expected encoding-corrupt key_update to fail guard, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.MissingNewsFields, ","), "key_update") {
		t.Fatalf("expected key_update missing field, got %#v", payload.MissingNewsFields)
	}
	if !strings.Contains(strings.Join(payload.Warnings, ","), "key_update_encoding_noise") {
		t.Fatalf("expected encoding warning, got %#v", payload.Warnings)
	}
}

func TestBuildGuardPayloadRejectsEncodingCorruptSupportingSource(t *testing.T) {
	payload := BuildGuardPayload(groundedContext(), map[string]any{
		"headline":     "示例事件最新进展",
		"source_url":   "https://news.example.com/latest-update.html",
		"published_at": "2026-04-22 07:30 UTC",
		"key_update":   "示例事件最新进展：双方于今日发布新声明，局势较前一日出现变化。",
		"source_count": 2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "第二来源确认示例事件最新进展",
				"source_url":   "https://wire.example.net/latest-update.html",
				"published_at": "2026-04-22 07:45 UTC",
				"key_update":   "鈥斺€斘恼伦钚路⒉际奔�:2026年7月。",
				"text":         "发布时间: 2026-04-22 07:45 UTC\n第二来源确认示例事件最新进展：第二来源确认双方发布新声明。",
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "needs_cross_check" || payload.CrossCheckReady || payload.ObservedSourceCount != 1 {
		t.Fatalf("expected noisy supporting source to be excluded from cross-check, got %#v", payload)
	}
	joined := strings.Join(payload.ReviewReasons, ",")
	if !strings.Contains(joined, "supporting_source_low_quality") {
		t.Fatalf("expected supporting source quality review reason, got %#v", payload.ReviewReasons)
	}
	if !strings.Contains(joined, "cross_check_evidence_missing") {
		t.Fatalf("expected missing cross-check review reason, got %#v", payload.ReviewReasons)
	}
}

func TestExtractEvidenceSkipsOriginalTitleBoilerplate(t *testing.T) {
	evidence := ExtractEvidence(
		"发布时间: 2026-05-16 09:00 UTC\n（原标题：美联储，重磅来袭！降息，突变！）\n美联储会议纪要即将公布，市场关注通胀数据和6月降息概率变化。",
		"美联储会议纪要即将公布！可能为利率前景提供线索",
		"https://news.example.com/fed-policy.html",
	)
	if strings.Contains(evidence.KeyUpdate, "原标题") {
		t.Fatalf("expected original-title boilerplate to be skipped, got %#v", evidence)
	}
	if !strings.Contains(evidence.KeyUpdate, "会议纪要") || !strings.Contains(evidence.KeyUpdate, "降息概率") {
		t.Fatalf("expected factual body line as key update, got %#v", evidence)
	}
}

func TestExtractEvidenceSkipsHeadlineRestatement(t *testing.T) {
	evidence := ExtractEvidence(
		"发布时间: 2026-05-17 10:49 UTC\n美联储，重磅来袭！降息，突变！\n美联储官员表示仍需观察通胀数据，市场对6月降息概率下调至38%。",
		"美联储，重磅来袭！降息，突变！",
		"https://news.example.com/fed-policy.html",
	)
	if evidence.KeyUpdate == "美联储，重磅来袭！降息，突变！" {
		t.Fatalf("expected headline restatement to be skipped, got %#v", evidence)
	}
	if !strings.Contains(evidence.KeyUpdate, "降息概率") {
		t.Fatalf("expected factual body line as key update, got %#v", evidence)
	}
}

func TestBuildGuardPayloadUsesResolverForSupportingSources(t *testing.T) {
	payload := BuildGuardPayload(groundedContext(), map[string]any{
		"headline":     "示例事件最新进展",
		"source_url":   "https://news.example.com/latest-update.html",
		"published_at": "2026-04-22 07:30 UTC",
		"key_update":   "示例事件最新进展：双方于今日发布新声明，局势较前一日出现变化。",
		"source_count": 2,
		"supporting_sources": []any{
			map[string]any{
				"source_url": "https://wire.example.net/latest-update.html",
			},
		},
	}, func(params map[string]any) Context {
		return Context{
			SourceURL: StringArg(params["source_url"]),
			Title:     "第二来源确认示例事件最新进展",
			Text:      "发布时间: 2026-04-22 07:45 UTC\n第二来源确认示例事件最新进展：第二来源确认双方发布新声明。",
		}
	}, PayloadOptions{})
	if payload.GuardStatus != "passed" || len(payload.SupportingSources) != 1 || !payload.SupportingSources[0].GroundedTextAvailable {
		t.Fatalf("expected resolver-grounded supporting source, got %#v", payload)
	}
}

func TestApplyEvidenceReviewUsesCleanSummaryBeforeGuardProjection(t *testing.T) {
	payload := BuildGuardPayload(groundedContext(), map[string]any{
		"headline":     "示例事件最新进展",
		"source_url":   "https://news.example.com/latest-update.html",
		"published_at": "2026-04-22 07:30 UTC",
		"key_update":   "All",
		"source_count": 2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "第二来源确认示例事件最新进展",
				"source_url":   "https://wire.example.net/latest-update.html",
				"published_at": "2026-04-22 07:45 UTC",
				"key_update":   "第二来源确认双方发布新声明。",
				"text":         "发布时间: 2026-04-22 07:45 UTC\n第二来源确认示例事件最新进展：第二来源确认双方发布新声明。",
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "missing_news_fields" {
		t.Fatalf("expected deterministic guard to miss weak key_update before review, got %#v", payload)
	}
	payload = ApplyEvidenceReview(context.Background(), payload, groundedContext(), LatestNewsLookupIntent{UserMessage: "帮我看下示例事件最新新闻"}, EvidenceReviewerFunc(func(context.Context, EvidenceReviewInput) (EvidenceReviewResult, error) {
		return EvidenceReviewResult{
			Accepted:     true,
			CleanSummary: "示例事件最新进展：双方于今日发布新声明，局势较前一日出现变化。",
			Confidence:   0.91,
			Source:       "test_reviewer",
		}, nil
	}))
	if payload.GuardStatus != "passed" || !payload.NewsFieldsReady {
		t.Fatalf("expected clean review summary to pass guard, got %#v", payload)
	}
	if payload.Evidence.KeyUpdate != "示例事件最新进展：双方于今日发布新声明，局势较前一日出现变化。" {
		t.Fatalf("expected reviewed clean summary to replace key_update, got %#v", payload.Evidence)
	}
	if payload.EvidenceReview == nil || !payload.EvidenceReview.Reviewed || !payload.EvidenceReview.Accepted {
		t.Fatalf("expected evidence review projection, got %#v", payload.EvidenceReview)
	}
}

func TestApplyEvidenceReviewRejectsNoisyEvidence(t *testing.T) {
	payload := BuildGuardPayload(groundedContext(), map[string]any{
		"headline":     "示例事件最新进展",
		"source_url":   "https://news.example.com/latest-update.html",
		"published_at": "2026-04-22 07:30 UTC",
		"key_update":   "示例事件最新进展：双方于今日发布新声明，局势较前一日出现变化。",
		"source_count": 2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "第二来源确认示例事件最新进展",
				"source_url":   "https://wire.example.net/latest-update.html",
				"published_at": "2026-04-22 07:45 UTC",
				"key_update":   "第二来源确认双方发布新声明。",
				"text":         "发布时间: 2026-04-22 07:45 UTC\n第二来源确认示例事件最新进展：第二来源确认双方发布新声明。",
			},
		},
	}, nil, PayloadOptions{})
	if payload.GuardStatus != "passed" {
		t.Fatalf("expected deterministic guard pass before review, got %#v", payload)
	}
	payload = ApplyEvidenceReview(context.Background(), payload, groundedContext(), LatestNewsLookupIntent{UserMessage: "帮我看下示例事件最新新闻"}, EvidenceReviewerFunc(func(context.Context, EvidenceReviewInput) (EvidenceReviewResult, error) {
		return EvidenceReviewResult{
			Accepted:       false,
			CleanSummary:   "这条清洗摘要不应在 rejected 状态下被采用。",
			RequiresReview: true,
			ReasonCodes:    []string{"summary_contains_page_noise"},
			Confidence:     0.83,
			Source:         "test_reviewer",
		}, nil
	}))
	if payload.GuardStatus != "needs_cross_check" || payload.Evaluation == nil || payload.Evaluation.Passed {
		t.Fatalf("expected evidence review rejection to degrade guard, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.ReviewReasons, ","), "evidence_review:summary_contains_page_noise") {
		t.Fatalf("expected reviewer reason code, got %#v", payload.ReviewReasons)
	}
	if payload.Evidence.KeyUpdate == "这条清洗摘要不应在 rejected 状态下被采用。" {
		t.Fatalf("rejected review must not replace evidence summary")
	}
}

func TestApplyEvidenceReviewFailureKeepsDeterministicGuard(t *testing.T) {
	payload := BuildGuardPayload(groundedContext(), map[string]any{
		"headline":     "示例事件最新进展",
		"source_url":   "https://news.example.com/latest-update.html",
		"published_at": "2026-04-22 07:30 UTC",
		"key_update":   "示例事件最新进展：双方于今日发布新声明，局势较前一日出现变化。",
		"source_count": 2,
		"supporting_sources": []any{
			map[string]any{
				"headline":     "第二来源确认示例事件最新进展",
				"source_url":   "https://wire.example.net/latest-update.html",
				"published_at": "2026-04-22 07:45 UTC",
				"key_update":   "第二来源确认双方发布新声明。",
				"text":         "发布时间: 2026-04-22 07:45 UTC\n第二来源确认示例事件最新进展：第二来源确认双方发布新声明。",
			},
		},
	}, nil, PayloadOptions{})
	payload = ApplyEvidenceReview(context.Background(), payload, groundedContext(), LatestNewsLookupIntent{UserMessage: "帮我看下示例事件最新新闻"}, EvidenceReviewerFunc(func(context.Context, EvidenceReviewInput) (EvidenceReviewResult, error) {
		return EvidenceReviewResult{}, errors.New("review model unavailable")
	}))
	if payload.GuardStatus != "passed" {
		t.Fatalf("expected reviewer failure to preserve deterministic guard, got %#v", payload)
	}
	if !strings.Contains(strings.Join(payload.Warnings, ","), "evidence_review_failed") {
		t.Fatalf("expected reviewer failure warning, got %#v", payload.Warnings)
	}
}

func TestSourceSiteNormalizesWWW(t *testing.T) {
	if got := SourceSite("https://www.News.Example.com/latest"); got != "news.example.com" {
		t.Fatalf("expected normalized host, got %q", got)
	}
}

func TestSourcePublisherFamilyCollapsesKnownAlternateDomains(t *testing.T) {
	left := SourcePublisherFamily("https://finance.sina.cn/article/1")
	right := SourcePublisherFamily("https://finance.sina.com.cn/article/2")
	if left != "sina" || right != left {
		t.Fatalf("expected alternate Sina domains to share a publisher family, left=%q right=%q", left, right)
	}
	if SourcePublisherFamily("https://news.example.com/a") == SourcePublisherFamily("https://wire.example.net/b") {
		t.Fatal("unrelated registrable domains must remain independent")
	}
}

func TestCommunitySourceURL(t *testing.T) {
	if !CommunitySourceURL("https://caifuhao.eastmoney.com/news/20260702084545800718600") {
		t.Fatal("expected Eastmoney wealth-account pages to be classified as community sources")
	}
	if !CommunitySourceURL("https://k.sina.cn/article_7857201856_1d45362c0019086b7g.html") {
		t.Fatal("expected Sina Kandian article pages to be classified as community sources")
	}
	if CommunitySourceURL("https://finance.eastmoney.com/a/202607021234.html") {
		t.Fatal("expected Eastmoney editorial finance pages not to be classified as community sources")
	}
}

func TestUserGeneratedSourceText(t *testing.T) {
	ugc := "特别声明：以上作品内容由凤凰网旗下自媒体平台用户上传并发布，本平台仅提供信息存储空间服务。"
	if !UserGeneratedSourceText(ugc) {
		t.Fatal("expected an explicit user-uploaded self-media disclosure to identify a community surface")
	}
	if UserGeneratedSourceText("记者核对公司公告并采访两名客户，确认产品测试已经启动。") {
		t.Fatal("expected an editorial evidence sentence not to be classified as user-generated")
	}
}

func TestKeyUpdateSufficientRejectsNavigationAndHTMLNoise(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "chinese_update", value: "第二来源确认示例事件出现最新进展。", want: true},
		{name: "english_update", value: "OpenAI launches a new product strategy.", want: true},
		{name: "short_navigation", value: "城事", want: false},
		{name: "spaced_chinese_navigation", value: "政治                              教育              法治              社会", want: false},
		{name: "date_category", value: "ProductMay 15, 2026", want: false},
		{name: "timestamp_only", value: "2026-07-13 20:10", want: false},
		{name: "localized_timestamp_only", value: "2026年7月13日 20:10", want: false},
		{name: "related_news_navigation", value: "联邦基金利率           相关新闻", want: false},
		{name: "load_more_navigation", value: "加载更多热门话题", want: false},
		{name: "html_noise", value: "<!doctype html><html><script>window.__DATA__={}</script>", want: false},
		{name: "promotional_registration", value: "欧易OKX    交易所    |    新人注册首选", want: false},
		{name: "promotional_exchange_catalog", value: "Gate 大门    交易所    |    币种最多最全", want: false},
		{name: "promotional_download", value: "Download app now and claim bonus with promo code.", want: false},
		{name: "stock_research_ad", value: "炒股就看金麒麟分析师研报，权威，专业，及时，全面，助您挖掘潜力主题机会！", want: false},
		{name: "first_person_investment_anecdote", value: "我在 BTC 9 万那天满仓追进去，结果两周亏掉 22.4% 后才明白。", want: false},
		{name: "site_navigation_boilerplate", value: "Microsoft发布2026年5月安全更新 欢迎访问网络信息中心 设为首页 加入收藏 当前位置：首页>网络安全>正文 地址：示例路1号 邮编：000000", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := KeyUpdateSufficient(tc.value); got != tc.want {
				t.Fatalf("KeyUpdateSufficient(%q)=%v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestKeyUpdateSufficientForHeadlineRejectsExactHeadlineRestatement(t *testing.T) {
	headline := "比特币市场最新进展"
	if KeyUpdateSufficientForHeadline(headline, headline) {
		t.Fatalf("expected exact headline restatement to be rejected")
	}
	if !KeyUpdateSufficientForHeadline(headline, "比特币市场最新进展：交易员关注资金流向和波动风险。") {
		t.Fatalf("expected headline-qualified factual update to be accepted")
	}
}

func TestExtractEvidenceSkipsPromotionalKeyUpdateLines(t *testing.T) {
	payload := BuildExtractPayload(Context{
		Title:     "比特币市场最新进展",
		SourceURL: "https://news.example.com/bitcoin-market.html",
		Text: strings.Join([]string{
			"发布时间: 2026-05-18 09:00 UTC",
			"欧易OKX    交易所    |    新人注册首选",
			"炒股就看金麒麟分析师研报，权威，专业，及时，全面，助您挖掘潜力主题机会！",
			"比特币市场最新进展：交易员关注资金流向和波动风险。",
		}, "\n"),
	}, PayloadOptions{})
	if payload.Evidence.KeyUpdate != "比特币市场最新进展：交易员关注资金流向和波动风险。" {
		t.Fatalf("expected promotional line to be skipped, got %#v", payload.Evidence)
	}
}

func TestExtractEvidenceSkipsAuthorshipMetadataLead(t *testing.T) {
	evidence := ExtractEvidence(
		"定焦One原创 作者 | 陈丹 编辑 | 魏佳 在大模型密集交卷的上周，团队交出了首份答卷。 7月6日，腾讯发布混元Hy3正式版。",
		"腾讯AI，翻身了吗？",
		"https://news.example.com/tencent-hy3.html",
	)
	if evidence.KeyUpdate != "7月6日，腾讯发布混元Hy3正式版。" {
		t.Fatalf("expected factual event after authorship metadata, got %#v", evidence)
	}
}

func TestExtractEvidenceSkipsListingMetadataBeforeFactualUpdate(t *testing.T) {
	payload := BuildExtractPayload(Context{
		Title:     "宏观政策速递",
		SourceURL: "https://news.example.com/2026/07/policy-update.html",
		Text: strings.Join([]string{
			"发布时间: 2026-07-13 20:10 UTC",
			"全部内容",
			"2026-07-13 20:10",
			"阅 65.85W",
			"政策委员会宣布维持利率不变，并表示将继续观察通胀与就业数据。",
		}, "\n"),
	}, PayloadOptions{})
	if payload.Evidence.KeyUpdate != "政策委员会宣布维持利率不变，并表示将继续观察通胀与就业数据。" {
		t.Fatalf("expected listing metadata to be skipped, got %#v", payload.Evidence)
	}
}

func TestExtractEvidenceSkipsEditorialTransitionBeforeFactualUpdate(t *testing.T) {
	payload := BuildExtractPayload(Context{
		Title:     "政策会议前瞻",
		SourceURL: "https://news.example.com/2026/07/policy-hearing.html",
		Text: strings.Join([]string{
			"发布时间: 2026-07-13 06:29 UTC",
			"政策委员会即将迎来重大考验。",
			"新任主席将于7月14日出席国会听证会，市场将关注其对利率路径的表态。",
		}, "\n"),
	}, PayloadOptions{})
	if payload.Evidence.KeyUpdate != "新任主席将于7月14日出席国会听证会，市场将关注其对利率路径的表态。" {
		t.Fatalf("expected editorial transition to be skipped, got %#v", payload.Evidence)
	}
}

func TestExtractEvidenceSkipsCustomerThanksBeforeShutdownFact(t *testing.T) {
	evidence := ExtractEvidence(strings.Join([]string{
		"感谢您一直以来给予网易《CC直播》的支持与厚爱！",
		"由于产品开发运营策略的调整，《CC直播》将于2026年8月31日15时终止运营。",
	}, "\n"), "《CC直播》停运公告", "https://cc.163.com/2026/06/29/cc20260630final/")
	if evidence.KeyUpdate != "由于产品开发运营策略的调整，《CC直播》将于2026年8月31日15时终止运营。" {
		t.Fatalf("expected shutdown fact after customer-thanks lead, got %#v", evidence)
	}
}

func TestExtractEvidenceWithPolicyAndFilterSkipsOffTopicLead(t *testing.T) {
	evidence := ExtractEvidenceWithPolicyAndFilter(strings.Join([]string{
		"发布时间: 2026-07-07 17:09 UTC",
		"黄金期货上涨1.03%，现货黄金随后回落。",
		"政策委员会将公布会议纪要，投资者将关注未来利率路径。",
	}, "\n"), "政策委员会最新表态", "https://news.example.com/policy.html", DefaultEvidenceQualityPolicy(), func(line string) bool {
		return strings.Contains(line, "政策委员会") || strings.Contains(line, "利率路径")
	})
	if evidence.KeyUpdate != "政策委员会将公布会议纪要，投资者将关注未来利率路径。" {
		t.Fatalf("expected relevance filter to skip off-topic factual lead, got %#v", evidence)
	}
}

func TestExtractEvidenceWithPolicyAndFilterReturnsOneRelevantSentence(t *testing.T) {
	evidence := ExtractEvidenceWithPolicyAndFilter(
		"黄金期货上涨1.03%，随后有所回落。 美联储将于7月9日公布会议纪要，投资者将关注未来利率路径。 同日还将公布其他市场数据。",
		"美联储最新表态",
		"https://news.example.com/fed-minutes.html",
		DefaultEvidenceQualityPolicy(),
		func(line string) bool { return strings.Contains(line, "美联储") },
	)
	if evidence.KeyUpdate != "美联储将于7月9日公布会议纪要，投资者将关注未来利率路径。" {
		t.Fatalf("expected one relevant sentence, got %#v", evidence)
	}
}

func TestExtractEvidenceWithPolicyAndScorerSelectsStrongestHeadlineEvent(t *testing.T) {
	title := "小米集团-W午前涨逾6% MiMo端侧模型通过备案发布具身生成大模型U0"
	intent := LatestNewsLookupIntent{
		Topic:          "小米集团-W",
		EntityMentions: []string{"小米集团-W"},
	}
	evidence := ExtractEvidenceWithPolicyAndScorer(strings.Join([]string{
		"小米集团-W（01810）午前涨逾6%，截至发稿，股价上涨5.96%，现报27.40港元，成交额41.29亿港元。",
		"小米MiMo端侧模型昨日正式通过国家大模型备案。",
		"7月15日，小米正式发布并开源具身生成模型Xiaomi-Robotics-U0。",
	}, "\n"), title, "https://news.example.com/xiaomi-models", DefaultEvidenceQualityPolicy(), func(line string) int {
		score := HeadlineEventEvidenceScore(title, line, intent)
		if score < 2 {
			return 0
		}
		return score
	})
	if !strings.Contains(evidence.KeyUpdate, "MiMo端侧模型") && !strings.Contains(evidence.KeyUpdate, "Xiaomi-Robotics-U0") {
		t.Fatalf("expected the strongest headline event instead of the market lead, got %#v", evidence)
	}
	if strings.Contains(evidence.KeyUpdate, "股价上涨") {
		t.Fatalf("did not expect a price-context lead to outrank the substantive headline event: %#v", evidence)
	}
}
