package wechatarticle

import (
	"context"
	"time"

	control "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

const (
	DefaultAdapterRef          control.DisplaySafeRef = "adapter:wechat_article_exporter"
	DefaultRunRef              control.DisplaySafeRef = "adapter_run:wechat_article_exporter"
	StrategyAccountSearch      control.DisplaySafeRef = "strategy:wechat_article_account_search"
	StrategySyncArticles       control.DisplaySafeRef = "strategy:wechat_article_sync_articles"
	StrategyDownloadArticle    control.DisplaySafeRef = "strategy:wechat_article_download_article"
	StrategySearchListDownload control.DisplaySafeRef = "strategy:wechat_article_search_list_download"
	EvidenceLoginStateProbe    control.DisplaySafeRef = "evidence:wechat_article_login_state_probe"
	EvidenceAccountSearch      control.DisplaySafeRef = "evidence:wechat_article_account_search_result"
	EvidenceArticleList        control.DisplaySafeRef = "evidence:wechat_article_list_result"
	EvidenceDedupKeys          control.DisplaySafeRef = "evidence:wechat_article_dedup_keys"
	EvidenceDownloadDigest     control.DisplaySafeRef = "evidence:wechat_article_download_body_digest"
	DefaultPageSize                                   = 10
	DefaultDownloadFormat                             = "text"
)

type LoginStatus struct {
	Valid          bool   `json:"valid"`
	Code           int    `json:"code,omitempty"`
	Message        string `json:"message,omitempty"`
	NextHostAction string `json:"next_host_action,omitempty"`
}

type Account struct {
	Nickname     string `json:"nickname,omitempty"`
	FakeID       string `json:"fakeid,omitempty"`
	Alias        string `json:"alias,omitempty"`
	Signature    string `json:"signature,omitempty"`
	Username     string `json:"username,omitempty"`
	RoundHeadImg string `json:"round_head_img,omitempty"`
}

type ArticleDedupKey struct {
	Link       string `json:"link,omitempty"`
	AID        string `json:"aid,omitempty"`
	AppMsgID   string `json:"appmsgid,omitempty"`
	ItemIdx    string `json:"itemidx,omitempty"`
	UpdateTime int64  `json:"update_time,omitempty"`
}

type Article struct {
	Title      string          `json:"title,omitempty"`
	Link       string          `json:"link,omitempty"`
	Digest     string          `json:"digest,omitempty"`
	Author     string          `json:"author,omitempty"`
	Cover      string          `json:"cover,omitempty"`
	AID        string          `json:"aid,omitempty"`
	AppMsgID   string          `json:"appmsgid,omitempty"`
	ItemIdx    string          `json:"itemidx,omitempty"`
	CreateTime int64           `json:"create_time,omitempty"`
	UpdateTime int64           `json:"update_time,omitempty"`
	IsDeleted  bool            `json:"is_deleted,omitempty"`
	DedupKey   ArticleDedupKey `json:"dedup_key"`
}

type ArticleListResult struct {
	FakeID               string    `json:"fakeid,omitempty"`
	Begin                int       `json:"begin"`
	Size                 int       `json:"size"`
	Articles             []Article `json:"articles"`
	RawArticleCount      int       `json:"raw_article_count"`
	DeletedFilteredCount int       `json:"deleted_filtered_count"`
}

// DownloadResult may carry raw body bytes only between a Host Client and its
// caller. Coordinator exposes only digest/count evidence in the AgentX result.
type DownloadResult struct {
	URL         string `json:"url"`
	Format      string `json:"format"`
	ContentType string `json:"content_type,omitempty"`
	Body        []byte `json:"-"`
	Text        string `json:"text,omitempty"`
}

type SyncOptions struct {
	AccountKeyword string
	FakeID         string
	ArticleKeyword string
	Begin          int
	Size           int
	DownloadFirst  bool
	DownloadFormat string
}
type SyncResult struct {
	AccountKeyword string            `json:"account_keyword,omitempty"`
	Account        *Account          `json:"account,omitempty"`
	Accounts       []Account         `json:"accounts,omitempty"`
	FakeID         string            `json:"fakeid,omitempty"`
	ArticleKeyword string            `json:"article_keyword,omitempty"`
	Articles       []Article         `json:"articles,omitempty"`
	Downloaded     *DownloadResult   `json:"downloaded,omitempty"`
	DedupKeys      []ArticleDedupKey `json:"dedup_keys,omitempty"`
}

// Client is implemented by the Host. The canonical package owns no HTTP,
// credential, cookie, login-QR, provider, or filesystem implementation.
type Client interface {
	CheckLogin(context.Context) (LoginStatus, error)
	SearchAccounts(context.Context, string) ([]Account, error)
	ListArticles(context.Context, string, int, int, string) (ArticleListResult, error)
	DownloadArticle(context.Context, string, string) (DownloadResult, error)
}

// Failure is the display-safe classification returned by a Host error mapper.
type Failure struct {
	Class          control.FailureClass
	MissingInputs  []control.MissingInput
	NextHostAction control.NextHostAction
	Reason         string
	Boundaries     []control.Boundary
}

type ErrorClassifier func(error) Failure

type Coordinator struct {
	Client         Client
	Descriptor     control.ProductionAdapterDescriptor
	AccountKeyword string
	FakeID         string
	ArticleKeyword string
	Begin          int
	Size           int
	DownloadFirst  bool
	DownloadFormat string
	DownloadURL    string
	Now            func() time.Time
	ClassifyError  ErrorClassifier
}

type Execution struct {
	Result     control.RuntimeAdapterExecutionResult `json:"result"`
	Login      LoginStatus                           `json:"login,omitempty"`
	Accounts   []Account                             `json:"accounts,omitempty"`
	Articles   []Article                             `json:"articles,omitempty"`
	Downloaded *DownloadResult                       `json:"downloaded,omitempty"`
	DedupKeys  []ArticleDedupKey                     `json:"dedup_keys,omitempty"`
}

func NewCoordinator(client Client) Coordinator {
	return Coordinator{Client: client, Size: DefaultPageSize, DownloadFormat: DefaultDownloadFormat}
}

func NormalizeDownloadFormat(format string) string {
	switch normalizeToken(format) {
	case "html", "markdown", "text":
		return normalizeToken(format)
	default:
		return DefaultDownloadFormat
	}
}

func ClampPageSize(size int) int {
	if size <= 0 {
		return DefaultPageSize
	}
	if size > 50 {
		return 50
	}
	return size
}
