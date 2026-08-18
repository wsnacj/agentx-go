# Web Retrieval API（Experimental）

`tools/web/retrieval` 是 Web 工具的 portable implementation owner。通常应优先使用上层 `tools/web`；只有需要自定义模型工具外壳、搜索 fallback 或审计投影的 Host 才直接使用本包。

主要合同包括：

- `Preparer`：Host-owned 网络构造端口；
- `PrepareSearchRequest`、`ExecutePreparedSearch`、`RunSearch`：搜索校验、provider 协议和响应归一化；
- `ExecuteWebFetch`：抓取、重定向诊断、内容提取、Firecrawl 显式 fallback；
- `Page`、`FindInPage` 与页面缓存：OpenPage/FindInPage 的共享机制；
- `NetworkErrorClassifier`：将 Host 私有网络错误映射到稳定检索错误类别，不转移策略所有权；
- `SearchAuditEvent`：只发出 policy-neutral 观察，审计开关、脱敏和落盘仍由 Host 负责。

豆包搜索Custom协议的identity为`doubao_custom`；旧`ark`名称仅规范化为
`doubao_custom`。Custom协议支持时间区间、站点包含/排除、最高权威等级和Query改写，
要求搜索专用凭据，且provider返回内容不进入进程内搜索缓存。

Global协议的identity为`doubao_global`，使用独立`global_search` endpoint和
`Documents/Snippet/HostInfo`响应结构。`DocCount`由portable count控制；
`MaxSnippetLength`和`IcpHostOnly`由Host配置，不暴露为模型可选参数。Global只支持按量后付费Key，
不支持Custom订阅套餐Key。

该包不会导入 HS、Runner、Scene、具体代理实现或凭据系统。
