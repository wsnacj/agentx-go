# Web Tools API（Experimental）

`tools/web` 提供可复用的 Web Search、Web Fetch、OpenPage 与 FindInPage 实现。它不是只包含 DTO 的 Facade：搜索 provider 协议、HTTP 响应解析、正文提取、readability、缓存、页面质量诊断和页内查找均由该包执行。

## 安全边界

调用方必须注入 `retrieval.Preparer`。Host 负责初始 URL 与每次重定向校验、代理、传输层、超时上限和连接关闭；本包不读取环境变量、不发现凭据，也不提供默认 provider。API key、endpoint 和 model 必须通过 `ProviderConfig` 显式传入。

```go
options := web.Options{
    Search: web.SearchOptions{
        DefaultProvider: "brave",
        Providers: map[string]web.ProviderConfig{
            "brave": {APIKey: braveKey},
        },
        Prepare: hostPreparer,
    },
    Fetch: web.FetchOptions{
        Prepare: hostPreparer,
        CacheTTL: 10 * time.Minute,
    },
}
web.Register(registry, options)
```

### 豆包搜索 Custom

豆包搜索使用独立的搜索凭据，不能复用方舟模型的`ARK_API_KEY`。Host应从豆包搜索控制台获取Key，
在调用点解析后显式注入；推荐的环境变量名是官方MCP使用的
`ASK_ECHO_SEARCH_INFINITY_API_KEY`。`ark`只作为旧配置兼容别名保留，新的provider identity为
`doubao_custom`：

```go
options := web.SearchOptions{
    DefaultProvider: retrieval.SearchProviderDoubaoCustom,
    Providers: map[string]web.ProviderConfig{
        retrieval.SearchProviderDoubaoCustom: {
            APIKey: searchKey,
        },
    },
    Prepare: hostPreparer,
}

result, err := web.RunSearch(ctx, web.SearchRequest{
    Query:              "AgentX Go 最新进展",
    MaxResults:         5,
    Freshness:          "pw",
    DomainFilter:       []string{"github.com", "-spam.example"},
    AuthoritativeOnly:  true,
    QueryRewrite:       true,
}, options)
```

Custom会把portable日期区间映射为`YYYY-MM-DD..YYYY-MM-DD`，把正负`domain_filter`分别映射为
`Sites`与`BlockHosts`，并固定请求`NeedUrl=true`、`NeedContent=false`，完整正文仍由`open_page`
按Host网络策略读取。响应优先使用适合模型消费的`Summary`，并保留`score`、`authority`、
`authority_level`与display-safe `request_id`。

受豆包搜索服务条款约束，Custom返回内容不会进入本包的进程内搜索缓存；调用方也不应把原始
Summary或Content写入durable store。URL、标题、请求ID和引用关系的持久化策略仍由Host审阅并负责。

### 豆包搜索 Global

Global使用`doubao_global`和独立`global_search` endpoint，不复用Custom解析器。Host可以固定
`MaxSnippetTokens`和是否只搜索国内ICP备案站点：

```go
options := web.SearchOptions{
    DefaultProvider: retrieval.SearchProviderDoubaoGlobal,
    Providers: map[string]web.ProviderConfig{
        retrieval.SearchProviderDoubaoGlobal: {
            APIKey: searchKey,
            DoubaoGlobal: web.DoubaoGlobalConfig{
                MaxSnippetTokens: 800,
                ICPHostOnly: true,
            },
        },
    },
    Prepare: hostPreparer,
}
```

Global只支持搜索控制台创建的按量后付费Key，订阅套餐Key会返回稳定的不支持/套餐错误。
Global返回的text snippet会合并为`SearchResult.Description`，`HostInfo.AuthorityLevel`
会写入`SearchResult.Authority`；图片snippet暂不进入web搜索结果。返回内容同样不进入进程搜索缓存。

## Go API

- `RunSearch(ctx, SearchRequest, SearchOptions)`：执行显式配置的搜索 provider；不自动读取凭据，也不实施产品级 provider 降级策略。
- `RunFetch(ctx, FetchRequest, FetchOptions)`：抓取 URL，执行 HTML/Markdown/Text/readability 提取，并返回外部内容标记与 provider diagnostics。
- `RunOpenPage(ctx, FetchRequest, FetchOptions)`：生成可读 `Page`，写入进程内有界缓存，返回供 `find_in_page` 使用的 `page_id`。
- `RunFindInPage(FindRequest)`：只访问页面缓存，不触发网络请求。
- `Register*` / `New*Handler`：将同一实现注册为模型可调用工具。

## 并发、取消与生命周期

公共入口可并发调用；内部缓存带锁。取消与 deadline 由 `context.Context` 传播到 Host `Preparer` 和 HTTP request。`PreparedRequest.Close` 在每次调用结束时执行。缓存是进程内优化，不是 durable store；需要跨进程恢复时应由 Host 在更高层提供持久化能力。

## Non-goal

- provider 凭据发现、Secret 管理；
- CIDR、端口、代理、重定向安全策略；
- provider backoff、授权、审批、审计落盘；
- 浏览器交互、登录态和 JavaScript 页面自动化；
- 将缓存作为 durable lifecycle 或 RunStore。

当前成熟度仍为 Experimental。已有fixed-version Platform consumer、无HS Tool-only真实Custom smoke与兼容
测试，但这不自动升级公共承诺；进入Developer Preview仍需完成候选surface冻结、签名漂移门禁和Global
按量后付费Key的正向证据，或明确把Global继续标为未验收扩展。
