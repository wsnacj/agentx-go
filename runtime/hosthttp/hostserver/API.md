# Host HTTP Server API（Experimental）

`hostserver` 是 Host-deployed Scene reference server 的通用 transport owner。
它只依赖 Go标准库，负责监听、暴露保护、请求大小、request identity和有界关闭；
它不拥有 Scene路由、领域 readiness、provider、credential或 backend。

## 构造与运行

```go
config := hostserver.DefaultConfig("127.0.0.1:8789")
server, err := config.NewServer(sceneHandler)

err = config.Serve(ctx, hostserver.ServeOptions{
    Handler: sceneHandler,
    OnListening: func(addr string) { /* publish loopback address */ },
})
```

- loopback默认允许无 token启动；
- 非 loopback同时要求 bearer token和 immediate trusted-proxy CIDR；
- `MaxBodyBytes`、header/read/write/idle timeout和 graceful shutdown都有有界默认值；
- `Serve`在 context结束后执行有界 `Shutdown`，失败时关闭 server并返回错误；
- `ServeUntilSignal`只适用于 reference binary/operator入口。

## Request identity

reference server自动应用 `RequestIdentityHandler`：

- 接受长度1～64、只含 ASCII字母、数字、`-_.:`的 `X-Request-ID`；
- 未提供时生成不可预测的 `req_<hex>` identity；
- 通过同名响应头回写，并可用 `RequestIDFromContext`读取；
- 非法 caller identity返回 `400 invalid_request_id`，生成新的 display-safe响应
  identity，且不调用 application handler；
- transport层认证、proxy和 body-limit错误也包含同一 response identity。

`/healthz`和`/readyz`可作为不带 bearer token的 probe路径，但非 loopback部署仍先
检查 trusted immediate peer。`/healthz`只应表达进程/liveness；具体 Scene必须在
自己的 handler中定义 `/readyz` capability/backend语义。

## Embedded handler责任

直接调用 Scene的 `http.Handler`不会自动获得本包的认证、request identity、body
limit、TLS、proxy、server timeout或 graceful shutdown。产品宿主若不使用
`Config.NewServer/Serve`，必须提供等价 transport边界并自行验证。
