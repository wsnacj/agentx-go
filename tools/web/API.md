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

当前成熟度为 Experimental。进入 Developer Preview 前仍需完成 HS fixed-version cutover、兼容差异测试和签名漂移门禁。
