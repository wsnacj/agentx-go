package wechatarticle_test

import (
	"context"
	"testing"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
	wechatarticle "github.com/wsnacj/agentx-go/scenes/wechatarticle"
)

type client struct{}

func (client) CheckLogin(context.Context) (wechatarticle.LoginStatus, error) {
	return wechatarticle.LoginStatus{Valid: true}, nil
}
func (client) SearchAccounts(context.Context, string) ([]wechatarticle.Account, error) {
	return []wechatarticle.Account{{FakeID: "fixture"}}, nil
}
func (client) ListArticles(context.Context, string, int, int, string) (wechatarticle.ArticleListResult, error) {
	return wechatarticle.ArticleListResult{Articles: []wechatarticle.Article{{Title: "Fixture", Link: "https://mp.weixin.qq.com/s/fixture"}}}, nil
}
func (client) DownloadArticle(context.Context, string, string) (wechatarticle.DownloadResult, error) {
	return wechatarticle.DownloadResult{Text: "fixture"}, nil
}

func TestExternalHostCanComposeCoordinator(t *testing.T) {
	coordinator := wechatarticle.NewCoordinator(client{})
	coordinator.AccountKeyword = "fixture"
	request := control.RuntimeAdapterExecutionRequest{Status: control.HostActionReady, ReadyForHostExecution: true, AdapterRef: wechatarticle.DefaultAdapterRef, StrategyRef: wechatarticle.StrategySyncArticles, IdempotencyRef: "idempotency:external", InputRefs: []control.DisplaySafeRef{"input:wechat_article"}, Frame: control.ObjectiveFrame{ID: "objective-wechat-external", SourceContext: []control.DisplaySafeRef{"source:wechat_external"}}}
	execution := coordinator.Execute(context.Background(), request)
	if execution.Result.Status != control.VerificationSatisfied || len(execution.Articles) != 1 {
		t.Fatalf("execution=%#v", execution)
	}
}
