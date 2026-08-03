package wechatarticle

import (
	"context"
	"testing"
	"time"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

type fixtureClient struct {
	loginCalls, searchCalls, listCalls, downloadCalls int
	login                                             LoginStatus
	err                                               error
}

func (client *fixtureClient) CheckLogin(context.Context) (LoginStatus, error) {
	client.loginCalls++
	return client.login, client.err
}
func (client *fixtureClient) SearchAccounts(context.Context, string) ([]Account, error) {
	client.searchCalls++
	return []Account{{Nickname: "AgentX", FakeID: "fakeid:1"}}, client.err
}
func (client *fixtureClient) ListArticles(context.Context, string, int, int, string) (ArticleListResult, error) {
	client.listCalls++
	article := Article{Title: "Article", Link: "https://mp.weixin.qq.com/s/example", DedupKey: ArticleDedupKey{AID: "1"}}
	return ArticleListResult{Articles: []Article{article}}, client.err
}
func (client *fixtureClient) DownloadArticle(context.Context, string, string) (DownloadResult, error) {
	client.downloadCalls++
	return DownloadResult{URL: "https://mp.weixin.qq.com/s/example", Format: "text", Text: "content"}, client.err
}

func readyWeChatRequest() control.RuntimeAdapterExecutionRequest {
	return control.RuntimeAdapterExecutionRequest{Status: control.HostActionReady, ReadyForHostExecution: true, AdapterRef: DefaultAdapterRef, StrategyRef: StrategySearchListDownload, IdempotencyRef: "idempotency:wechat_test", InputRefs: []control.DisplaySafeRef{"input:wechat_article"}, Frame: control.ObjectiveFrame{ID: "objective-wechat-test", SourceContext: []control.DisplaySafeRef{"source:wechat_test"}}}
}

func TestCoordinatorRunsSearchListDownloadExactlyOnce(t *testing.T) {
	client := &fixtureClient{login: LoginStatus{Valid: true}}
	coordinator := NewCoordinator(client)
	coordinator.AccountKeyword = "AgentX"
	coordinator.DownloadFirst = true
	coordinator.Now = func() time.Time { return time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC) }
	execution := coordinator.Execute(context.Background(), readyWeChatRequest())
	if client.loginCalls != 1 || client.searchCalls != 1 || client.listCalls != 1 || client.downloadCalls != 1 {
		t.Fatalf("calls=%d/%d/%d/%d", client.loginCalls, client.searchCalls, client.listCalls, client.downloadCalls)
	}
	if execution.Result.Status != control.VerificationSatisfied || len(execution.Articles) != 1 || execution.Downloaded == nil {
		t.Fatalf("execution=%#v", execution)
	}
	if evaluation := Evaluate(execution, true); !evaluation.Passed {
		t.Fatalf("evaluation=%#v", evaluation)
	}
}

func TestCoordinatorFailsClosedForInvalidLogin(t *testing.T) {
	client := &fixtureClient{login: LoginStatus{Valid: false, NextHostAction: "open_wechat_exporter_scan_login"}}
	coordinator := NewCoordinator(client)
	coordinator.AccountKeyword = "AgentX"
	execution := coordinator.Execute(context.Background(), readyWeChatRequest())
	if execution.Result.Status != control.VerificationBlocked || execution.Result.FailureClass != control.FailureCredentialMissing || client.searchCalls != 0 {
		t.Fatalf("execution=%#v calls=%d", execution, client.searchCalls)
	}
}

func TestDefinitionIsReadOnlyAndBounded(t *testing.T) {
	definition := Definition()
	if definition.Manifest.ID != PackID || len(definition.Tools) != 1 || definition.PolicyProfiles[0].Contract.Budget.MaxToolCalls != 1 {
		t.Fatalf("definition=%#v", definition)
	}
}
