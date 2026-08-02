# `scenes/publicnews/hostkit` 中文 API Reference

成熟度：**Experimental extension**。

本包提供公开新闻查询的 provider-neutral Host 协调层。它按照既有顺序调用 Host 注入的
来源、抽取和 guard handler，并将结果投影为 `publicnews.LatestNewsLookupPayload`。

## 推荐入口

- `LatestNewsLookupConfig`：声明 source、默认 source policy、回答边界开关、重试策略、
  evidence reviewer 和 handlers。
- `LatestNewsLookupHandlers`：`Sources`、`Extract`、`Guard` 三个显式注入点。
- `BuildLatestNewsLookupPayload`：执行一次 bounded coordination，遵守
  `context.Context`；不会绕过 cancellation/deadline。
- `BuildLatestNewsLookupHandler`：把同一协调逻辑包装为 portable tool handler。
- `LatestNewsRetryPolicy`：只控制已注入来源 handler 的有界重试；不会隐式切换
  provider。

当来源已经是 terminal config/auth/quota failure 时，本包保持既有 fail-closed 策略，
不会继续抽取或 guard；当 evidence 不足时，返回结构化 readiness、failure code 和
answer contract，由产品层决定展示方式。

调用方必须自行提供搜索/页面读取实现、credential、安全授权、source policy、审计和
真实副作用。本包不依赖 HS、Runner 或 Scene。
