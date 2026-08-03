package main

import (
	"context"
	"fmt"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
	publicsource "github.com/wsnacj/agentx-go/scenes/publicsource"
	wechatarticle "github.com/wsnacj/agentx-go/scenes/wechatarticle"
)

type sourceCollector struct{}

func (sourceCollector) CollectPublicSourceEvidence(context.Context, publicsource.Request) (publicsource.Report, error) {
	return publicsource.Report{
		Status:           control.VerificationSatisfied,
		Evidence:         []publicsource.Evidence{{SourceRef: "source:fixture", QueryRef: "query:fixture", EvidenceRef: "evidence:fixture", Strength: control.EvidenceAdequate}},
		DisplaySummaries: []publicsource.DisplaySummary{{SourceRef: "source:fixture", Title: "Fixture", Summary: "Display safe source", AttestationRef: "attestation:fixture", RedactionRef: "redaction:fixture"}},
	}, nil
}

type articleClient struct{}

func (articleClient) CheckLogin(context.Context) (wechatarticle.LoginStatus, error) {
	return wechatarticle.LoginStatus{Valid: true}, nil
}
func (articleClient) SearchAccounts(context.Context, string) ([]wechatarticle.Account, error) {
	return []wechatarticle.Account{{Nickname: "Fixture", FakeID: "fixture"}}, nil
}
func (articleClient) ListArticles(context.Context, string, int, int, string) (wechatarticle.ArticleListResult, error) {
	return wechatarticle.ArticleListResult{Articles: []wechatarticle.Article{{Title: "Fixture article", Link: "https://mp.weixin.qq.com/s/fixture", DedupKey: wechatarticle.ArticleDedupKey{AID: "1"}}}}, nil
}
func (articleClient) DownloadArticle(context.Context, string, string) (wechatarticle.DownloadResult, error) {
	return wechatarticle.DownloadResult{Format: "text", Text: "fixture body"}, nil
}

func run(ctx context.Context) (string, error) {
	sourceRequest := control.RuntimeAdapterExecutionRequest{Status: control.HostActionReady, ReadyForHostExecution: true, AdapterRef: publicsource.DefaultAdapterRef, StrategyRef: publicsource.DefaultStrategyRef, IdempotencyRef: "idempotency:source_consumer", InputRefs: []control.DisplaySafeRef{"query:fixture"}, Frame: control.ObjectiveFrame{ID: "objective-source-consumer", SourceContext: []control.DisplaySafeRef{"source:fixture"}}}
	sourceExecution := publicsource.NewCoordinator(sourceCollector{}).Execute(ctx, sourceRequest)
	if !publicsource.Evaluate(sourceExecution.Report, true).Passed {
		return "", fmt.Errorf("publicsource evaluation failed: %s", sourceExecution.Result.FailureReason)
	}

	articleCoordinator := wechatarticle.NewCoordinator(articleClient{})
	articleCoordinator.AccountKeyword, articleCoordinator.DownloadFirst = "Fixture", true
	articleRequest := control.RuntimeAdapterExecutionRequest{Status: control.HostActionReady, ReadyForHostExecution: true, AdapterRef: wechatarticle.DefaultAdapterRef, StrategyRef: wechatarticle.StrategySearchListDownload, IdempotencyRef: "idempotency:article_consumer", InputRefs: []control.DisplaySafeRef{"input:wechat_article"}, Frame: control.ObjectiveFrame{ID: "objective-article-consumer", SourceContext: []control.DisplaySafeRef{"source:fixture"}}}
	articleExecution := articleCoordinator.Execute(ctx, articleRequest)
	if !wechatarticle.Evaluate(articleExecution, true).Passed {
		return "", fmt.Errorf("wechatarticle evaluation failed: %s", articleExecution.Result.FailureReason)
	}
	return fmt.Sprintf("agentx-sourceacquisition-ok:%s:%s:%d:%d", publicsource.PackID, wechatarticle.PackID, len(sourceExecution.Report.Evidence), len(articleExecution.Articles)), nil
}

func main() {
	output, err := run(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}
