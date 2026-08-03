package wechatarticle

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

func Descriptor() control.ProductionAdapterDescriptor {
	return control.ProductionAdapterDescriptor{
		AdapterRef: DefaultAdapterRef, Owner: "scene", OwnerRef: "scene:agentx_wechat_article", Version: "v1", Kind: control.ProductionAdapterSourceReadback,
		SupportedCandidateRefs: []control.DisplaySafeRef{StrategyAccountSearch, StrategySyncArticles, StrategyDownloadArticle, StrategySearchListDownload},
		ProvidesCapabilityRefs: []control.DisplaySafeRef{"capability:wechat_article_account_search", "capability:wechat_article_article_list", "capability:wechat_article_download"},
		RequiresCapabilityRefs: []control.DisplaySafeRef{"capability:wechat_article_exporter_backend", "capability:wechat_article_basic_auth", "capability:wechat_article_auth_key", "capability:wechat_article_login_state", "capability:wechat_article_account_search", "capability:wechat_article_article_list", "capability:wechat_article_download"},
		InputContractRef:       "contract:wechat_article.search_list_download_request_refs", OutputContractRef: "contract:wechat_article.search_list_download_result_refs", ReadbackContractRef: "contract:wechat_article.exporter_readback",
		RequiredPolicyRefs: []control.DisplaySafeRef{"policy:wechat_article_readonly", "policy:wechat_article_credentials_redacted"}, RequiredApprovalRefs: []control.DisplaySafeRef{"approval:wechat_article_exporter_read"}, RequiredBudgetRef: "budget:wechat_article_exporter_read",
		IdempotencyContractRef: "idempotency:wechat_article_search_list_download", RiskRef: "risk:wechat_article_credentialed_read", SideEffectClass: "read_only", TimeoutPolicyRef: "timeout:wechat_article_exporter_request", RedactionPolicyRef: "redaction:wechat_article_no_secrets",
		DisplaySafeInputRefs: []control.DisplaySafeRef{"input:wechat_article_account_keyword_or_fakeid", "input:wechat_article_page_request", "input:wechat_article_download_format"}, DisplaySafeOutputRefs: []control.DisplaySafeRef{EvidenceAccountSearch, EvidenceArticleList, EvidenceDedupKeys, EvidenceDownloadDigest},
		Boundaries: []control.Boundary{"wechat_article_runtime_adapter_descriptor", "scene_declared_wechat_article_exporter_adapter", "host_owned_exporter_backend_required", "read_only_account_search_article_list_download", "no_basic_auth_secret_in_descriptor", "no_auth_key_in_descriptor", "no_cookie_in_descriptor", "no_browser_automation"},
	}.Normalize()
}

func Registry() control.HostAdapterRegistrySnapshot {
	descriptor := Descriptor()
	return control.BuildHostAdapterRegistry(control.HostAdapterRegistryInput{RegistryRef: "registry:wechat_article_exporter_scene", Descriptors: []control.ProductionAdapterDescriptor{descriptor}, PolicyRefs: descriptor.RequiredPolicyRefs})
}

func (coordinator Coordinator) Execute(ctx context.Context, request control.RuntimeAdapterExecutionRequest) Execution {
	if ctx == nil {
		ctx = context.Background()
	}
	coordinator = coordinator.normalize()
	request = request.Normalize()
	base := coordinator.resultInput(request, control.VerificationSatisfied)
	if !request.ReadyForHostExecution {
		base.Status, base.FailureClass, base.FailureReason, base.NextHostAction = control.VerificationBlocked, firstFailure(request.FailureClass, control.FailureConfigMissing), "runtime_adapter_request_not_ready", "provide_runtime_adapter_execution_request"
		base.MissingInputs, base.Boundaries = []control.MissingInput{"host:runtime_adapter_execution_request"}, []control.Boundary{"wechat_article_runtime_adapter_request_not_ready"}
		return Execution{Result: control.BuildRuntimeAdapterExecutionResult(base)}
	}
	if request.AdapterRef != DefaultAdapterRef {
		base.Status, base.FailureClass, base.FailureReason, base.NextHostAction = control.VerificationBlocked, control.FailureHostAdapterMissing, "wechat_article_adapter_ref_mismatch", "provide_wechat_article_runtime_adapter"
		base.MissingInputs, base.Boundaries = []control.MissingInput{"host:wechat_article_runtime_adapter"}, []control.Boundary{"wechat_article_runtime_adapter_ref_mismatch"}
		return Execution{Result: control.BuildRuntimeAdapterExecutionResult(base)}
	}
	if !supportedStrategy(request.StrategyRef) {
		base.Status, base.FailureClass, base.FailureReason, base.NextHostAction = control.VerificationBlocked, control.FailureHostAdapterMissing, "wechat_article_strategy_not_supported", "select_wechat_article_runtime_strategy"
		base.MissingInputs, base.Boundaries = []control.MissingInput{"host:wechat_article_runtime_strategy"}, []control.Boundary{"wechat_article_runtime_strategy_not_supported"}
		return Execution{Result: control.BuildRuntimeAdapterExecutionResult(base)}
	}
	if coordinator.Client == nil {
		base.Status, base.FailureClass, base.FailureReason, base.NextHostAction = control.VerificationBlocked, control.FailureHostAdapterMissing, "wechat_article_exporter_client_missing", "configure_wechat_article_exporter_client"
		base.MissingInputs, base.Boundaries = []control.MissingInput{"host:wechat_article_exporter_client"}, []control.Boundary{"wechat_article_exporter_client_missing"}
		return Execution{Result: control.BuildRuntimeAdapterExecutionResult(base)}
	}
	login, err := coordinator.Client.CheckLogin(ctx)
	if err != nil {
		base = coordinator.applyFailure(base, err)
		return Execution{Login: login, Result: control.BuildRuntimeAdapterExecutionResult(base)}
	}
	report := Execution{Login: login}
	base.Observations = append(base.Observations, coordinator.observation("wechat_article_login_state", "login_valid", fmt.Sprintf("%t", login.Valid), EvidenceLoginStateProbe))
	if !login.Valid {
		base.Status, base.FailureClass, base.FailureReason = control.VerificationBlocked, control.FailureCredentialMissing, "wechat_article_login_invalid_or_expired"
		base.NextHostAction = control.NextHostAction(first(login.NextHostAction, "open_wechat_exporter_scan_login"))
		base.MissingInputs, base.Boundaries = []control.MissingInput{loginMissingInput(login)}, []control.Boundary{"wechat_article_scan_login_required"}
		report.Result = control.BuildRuntimeAdapterExecutionResult(base)
		return report
	}
	switch request.StrategyRef {
	case StrategyAccountSearch:
		return coordinator.runAccountSearch(ctx, request, base, report)
	case StrategySyncArticles:
		return coordinator.runArticleList(ctx, request, base, report)
	case StrategyDownloadArticle:
		return coordinator.runDownload(ctx, request, base, report)
	case StrategySearchListDownload:
		return coordinator.runSearchListDownload(ctx, request, base, report)
	default:
		return report
	}
}

func (coordinator Coordinator) normalize() Coordinator {
	out := coordinator
	out.AccountKeyword, out.FakeID, out.ArticleKeyword = strings.TrimSpace(out.AccountKeyword), strings.TrimSpace(out.FakeID), strings.TrimSpace(out.ArticleKeyword)
	if out.Begin < 0 {
		out.Begin = 0
	}
	out.Size, out.DownloadFormat, out.DownloadURL = ClampPageSize(out.Size), NormalizeDownloadFormat(out.DownloadFormat), strings.TrimSpace(out.DownloadURL)
	if out.Descriptor.Normalize().AdapterRef == "" {
		out.Descriptor = Descriptor()
	} else {
		out.Descriptor = out.Descriptor.Normalize()
	}
	return out
}

func (coordinator Coordinator) runAccountSearch(ctx context.Context, request control.RuntimeAdapterExecutionRequest, input control.RuntimeAdapterExecutionResultInput, report Execution) Execution {
	if coordinator.AccountKeyword == "" {
		input.Status, input.FailureClass, input.FailureReason, input.NextHostAction = control.VerificationBlocked, control.FailureConfigMissing, "wechat_article_account_keyword_missing", "provide_wechat_article_account_keyword"
		input.MissingInputs, input.Boundaries = []control.MissingInput{"input:wechat_article_account_keyword"}, []control.Boundary{"wechat_article_account_keyword_required"}
		report.Result = control.BuildRuntimeAdapterExecutionResult(input)
		return report
	}
	accounts, err := coordinator.Client.SearchAccounts(ctx, coordinator.AccountKeyword)
	if err != nil {
		report.Result = control.BuildRuntimeAdapterExecutionResult(coordinator.applyFailure(input, err))
		return report
	}
	report.Accounts = accounts
	input.Observations = append(input.Observations, coordinator.observation("wechat_article_account_search", "account_count", fmt.Sprintf("%d", len(accounts)), EvidenceAccountSearch))
	input.EvidenceRefs, input.OutputRefs = control.MergeEvidenceRefs(input.EvidenceRefs, coordinator.evidence("wechat_article_account_search", EvidenceAccountSearch)), append(input.OutputRefs, EvidenceAccountSearch)
	report.Result = control.BuildRuntimeAdapterExecutionResult(input)
	return report
}

func (coordinator Coordinator) runArticleList(ctx context.Context, request control.RuntimeAdapterExecutionRequest, input control.RuntimeAdapterExecutionResultInput, report Execution) Execution {
	fakeID := coordinator.FakeID
	if fakeID == "" {
		if coordinator.AccountKeyword == "" {
			input.Status, input.FailureClass, input.FailureReason, input.NextHostAction = control.VerificationBlocked, control.FailureConfigMissing, "wechat_article_fakeid_or_keyword_missing", "provide_wechat_article_fakeid_or_account_keyword"
			input.MissingInputs, input.Boundaries = []control.MissingInput{"input:wechat_article_fakeid_or_account_keyword"}, []control.Boundary{"wechat_article_fakeid_or_keyword_required"}
			report.Result = control.BuildRuntimeAdapterExecutionResult(input)
			return report
		}
		accounts, err := coordinator.Client.SearchAccounts(ctx, coordinator.AccountKeyword)
		if err != nil {
			report.Result = control.BuildRuntimeAdapterExecutionResult(coordinator.applyFailure(input, err))
			return report
		}
		report.Accounts = accounts
		input.Observations = append(input.Observations, coordinator.observation("wechat_article_account_search", "account_count", fmt.Sprintf("%d", len(accounts)), EvidenceAccountSearch))
		if len(accounts) == 0 || strings.TrimSpace(accounts[0].FakeID) == "" {
			input.Status, input.FailureClass, input.FailureReason, input.NextHostAction = control.VerificationBlocked, control.FailureTargetUnavailable, "wechat_article_account_not_found", "refine_wechat_article_account_keyword"
			input.MissingInputs, input.Boundaries = []control.MissingInput{"input:wechat_article_account_keyword"}, []control.Boundary{"wechat_article_account_search_returned_no_fakeid"}
			report.Result = control.BuildRuntimeAdapterExecutionResult(input)
			return report
		}
		fakeID = strings.TrimSpace(accounts[0].FakeID)
	}
	list, err := coordinator.Client.ListArticles(ctx, fakeID, coordinator.Begin, coordinator.Size, coordinator.ArticleKeyword)
	if err != nil {
		report.Result = control.BuildRuntimeAdapterExecutionResult(coordinator.applyFailure(input, err))
		return report
	}
	report.Articles, report.DedupKeys = list.Articles, dedupKeys(list.Articles)
	input.Observations = append(input.Observations, coordinator.observation("wechat_article_list", "article_count", fmt.Sprintf("%d", len(list.Articles)), EvidenceArticleList), coordinator.observation("dedup_keys", "dedup_key_count", fmt.Sprintf("%d", len(report.DedupKeys)), EvidenceDedupKeys))
	input.EvidenceRefs = control.MergeEvidenceRefs(input.EvidenceRefs, coordinator.evidence("wechat_article_list", EvidenceArticleList), coordinator.evidence("dedup_keys", EvidenceDedupKeys))
	input.OutputRefs = append(input.OutputRefs, EvidenceArticleList, EvidenceDedupKeys)
	report.Result = control.BuildRuntimeAdapterExecutionResult(input)
	return report
}

func (coordinator Coordinator) runDownload(ctx context.Context, request control.RuntimeAdapterExecutionRequest, input control.RuntimeAdapterExecutionResultInput, report Execution) Execution {
	downloadURL := coordinator.DownloadURL
	if downloadURL == "" && coordinator.FakeID != "" {
		list, err := coordinator.Client.ListArticles(ctx, coordinator.FakeID, coordinator.Begin, coordinator.Size, coordinator.ArticleKeyword)
		if err != nil {
			report.Result = control.BuildRuntimeAdapterExecutionResult(coordinator.applyFailure(input, err))
			return report
		}
		report.Articles, report.DedupKeys = list.Articles, dedupKeys(list.Articles)
		input.Observations = append(input.Observations, coordinator.observation("wechat_article_list", "article_count", fmt.Sprintf("%d", len(list.Articles)), EvidenceArticleList))
		if len(list.Articles) > 0 {
			downloadURL = list.Articles[0].Link
		}
	}
	if downloadURL == "" {
		input.Status, input.FailureClass, input.FailureReason, input.NextHostAction = control.VerificationBlocked, control.FailureConfigMissing, "wechat_article_download_url_missing", "provide_wechat_article_download_url_or_fakeid"
		input.MissingInputs, input.Boundaries = []control.MissingInput{"input:wechat_article_download_url_or_fakeid"}, []control.Boundary{"wechat_article_download_url_required"}
		report.Result = control.BuildRuntimeAdapterExecutionResult(input)
		return report
	}
	downloaded, err := coordinator.Client.DownloadArticle(ctx, downloadURL, coordinator.DownloadFormat)
	if err != nil {
		report.Result = control.BuildRuntimeAdapterExecutionResult(coordinator.applyFailure(input, err))
		return report
	}
	report.Downloaded = &downloaded
	coordinator.appendDownload(&input, downloaded)
	report.Result = control.BuildRuntimeAdapterExecutionResult(input)
	return report
}

func (coordinator Coordinator) runSearchListDownload(ctx context.Context, request control.RuntimeAdapterExecutionRequest, input control.RuntimeAdapterExecutionResultInput, report Execution) Execution {
	report = coordinator.runArticleList(ctx, request, input, report)
	if !report.Result.ReadyForObservationNormalization {
		return report
	}
	input = coordinator.resultInput(request, control.VerificationSatisfied)
	input.Observations, input.EvidenceRefs, input.OutputRefs = append(input.Observations, report.Result.Observations...), control.MergeEvidenceRefs(input.EvidenceRefs, report.Result.EvidenceRefs), append(input.OutputRefs, report.Result.OutputRefs...)
	if coordinator.DownloadFirst && len(report.Articles) > 0 {
		downloaded, err := coordinator.Client.DownloadArticle(ctx, report.Articles[0].Link, coordinator.DownloadFormat)
		if err != nil {
			report.Result = control.BuildRuntimeAdapterExecutionResult(coordinator.applyFailure(input, err))
			return report
		}
		report.Downloaded = &downloaded
		coordinator.appendDownload(&input, downloaded)
	}
	report.Result = control.BuildRuntimeAdapterExecutionResult(input)
	return report
}

func (coordinator Coordinator) appendDownload(input *control.RuntimeAdapterExecutionResultInput, downloaded DownloadResult) {
	body := downloaded.Body
	if len(body) == 0 {
		body = []byte(downloaded.Text)
	}
	digest := sha256.Sum256(body)
	input.Observations = append(input.Observations, coordinator.observation("download_body_digest", "download_sha256", "sha256:"+fmt.Sprintf("%x", digest), EvidenceDownloadDigest), coordinator.observation("download_body_digest", "download_char_count", fmt.Sprintf("%d", len([]rune(downloaded.Text))), EvidenceDownloadDigest))
	input.EvidenceRefs, input.OutputRefs = control.MergeEvidenceRefs(input.EvidenceRefs, coordinator.evidence("download_body_digest", EvidenceDownloadDigest)), append(input.OutputRefs, EvidenceDownloadDigest)
}

func (coordinator Coordinator) resultInput(request control.RuntimeAdapterExecutionRequest, status control.VerificationStatus) control.RuntimeAdapterExecutionResultInput {
	return control.RuntimeAdapterExecutionResultInput{Request: request, AdapterRef: DefaultAdapterRef, StrategyRef: request.StrategyRef, HostAdapterRunRef: DefaultRunRef, Status: status, Boundaries: control.AppendBoundaries(coordinator.Descriptor.Boundaries, "wechat_article_runtime_adapter_execution", "scene_owned_wechat_article_adapter", "host_owned_exporter_client", "display_safe_refs_only", "no_basic_auth_secret_in_result", "no_auth_key_in_result", "no_cookie_in_result", "no_browser_automation")}
}

func (coordinator Coordinator) applyFailure(input control.RuntimeAdapterExecutionResultInput, err error) control.RuntimeAdapterExecutionResultInput {
	failure := Failure{Class: control.FailureExternalDependencyUnavailable, MissingInputs: []control.MissingInput{"host:wechat_article_exporter_readback"}, NextHostAction: "review_wechat_article_exporter_readback", Reason: "wechat_article_exporter_error", Boundaries: []control.Boundary{"wechat_article_exporter_readback_failed"}}
	if coordinator.ClassifyError != nil {
		failure = coordinator.ClassifyError(err)
	}
	input.Status, input.FailureClass, input.MissingInputs, input.NextHostAction, input.FailureReason, input.Boundaries = control.VerificationBlocked, control.NormalizeFailureClass(string(failure.Class)), append([]control.MissingInput(nil), failure.MissingInputs...), failure.NextHostAction, normalizeToken(failure.Reason), append([]control.Boundary(nil), failure.Boundaries...)
	if input.FailureClass == control.FailureNone {
		input.FailureClass = control.FailureExternalDependencyUnavailable
	}
	if input.FailureReason == "" {
		input.FailureReason = "wechat_article_exporter_error"
	}
	return input
}

func (coordinator Coordinator) observation(kind, name, value string, evidenceRef control.DisplaySafeRef) control.Observation {
	observedAt := coordinator.observedAt()
	evidence := control.EvidenceRef{Ref: evidenceRef, Kind: kind, Strength: control.EvidenceAdequate, Source: DefaultRunRef, ObservedAt: observedAt}
	return control.Observation{Kind: kind, Source: DefaultRunRef, Subject: "objective:wechat_article_runtime_adapter", Name: name, Value: value, Strength: control.EvidenceAdequate, ObservedAt: observedAt, EvidenceRefs: []control.EvidenceRef{evidence}, DisplaySafeRefs: []control.DisplaySafeRef{evidenceRef}}
}

func (coordinator Coordinator) evidence(kind string, ref control.DisplaySafeRef) []control.EvidenceRef {
	return []control.EvidenceRef{{Ref: ref, Kind: kind, Strength: control.EvidenceAdequate, Source: DefaultRunRef, ObservedAt: coordinator.observedAt()}}
}
func (coordinator Coordinator) observedAt() string {
	now := coordinator.Now
	if now == nil {
		now = time.Now
	}
	return now().UTC().Format(time.RFC3339)
}
func dedupKeys(articles []Article) []ArticleDedupKey {
	out := make([]ArticleDedupKey, 0, len(articles))
	for _, article := range articles {
		out = append(out, article.DedupKey)
	}
	return out
}
func supportedStrategy(ref control.DisplaySafeRef) bool {
	return ref == StrategyAccountSearch || ref == StrategySyncArticles || ref == StrategyDownloadArticle || ref == StrategySearchListDownload
}
func loginMissingInput(login LoginStatus) control.MissingInput {
	if strings.EqualFold(strings.TrimSpace(login.NextHostAction), "set_auth_key_or_auth_key_file") {
		return "host:wechat_article_auth_key"
	}
	return "host:wechat_article_scan_login"
}
func firstFailure(values ...control.FailureClass) control.FailureClass {
	for _, value := range values {
		if normalized := control.NormalizeFailureClass(string(value)); normalized != control.FailureNone {
			return normalized
		}
	}
	return control.FailureNone
}
func normalizeToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer(" ", "_", "/", "_", "\\", "_", ".", "_", "-", "_").Replace(value)
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == ':' {
			b.WriteRune(r)
		}
	}
	return strings.Trim(b.String(), "_:")
}

func first(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
