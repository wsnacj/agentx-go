# `browser/tools` 中文 API Reference

成熟度：`Experimental extension / Developer Preview candidate`

`github.com/wsnacj/agentx-go/browser/tools` 把 provider-neutral 的
[`browser/runtime`](../runtime/API.md) 合同注册成 AgentX Tool。它拥有参数解析、schema、
tool handler、session 协调、能力投影、错误收口和 Browser Local Planner 机制；不拥有具体
浏览器进程、provider、credential、approval、企业网络策略或产品默认值。

## 最小接入

调用方至少提供一个实现 `BrowserBackend` 的 backend，并显式选择要注册的 tool：

```go
registry := agentxtools.NewRegistry()
browsertools.RegisterBrowserTools(registry, browsertools.BrowserToolOptions{
    Backend: backend,
    EnabledTools: []string{"browser"},
})

result, err := registry.Execute(ctx, llm.FunctionCall{
    Name: "browser",
    Arguments: `{"action":"open","url":"https://example.com"}`,
})
```

推荐新接入只暴露统一入口 `browser`。注册 `browser` 时，内部会同时建立必要的
`browser_runtime` 与 `browser_act` handler；旧调用方可继续选择 specialist/compat surface。

## Tool surface

- Unified：`browser`；
- Specialist：`browser_runtime`、`browser_act`；
- Compat：`browser_open`、`browser_navigate`、`browser_tabs`、`browser_extract`、
  `browser_screenshot`、`browser_click`、`browser_type`、`browser_eval`。

`BrowserUnifiedToolNames`、`BrowserSpecialistToolNames`、`BrowserCompatToolNames` 和
`BrowserAllToolNames` 返回防御性副本；`BrowserSurfaceForToolName`、`ResolveBrowserSurface`
用于兼容入口与统一入口之间的映射。它们不代表 Public/Beta/Stable 承诺。

## `BrowserToolOptions`

主要字段分为四组：

1. 执行参数：`Root`、`TimeoutMs`、`MaxChars`、`OpenWaitMs`、`ScreenshotWaitMs`；
2. backend：`Backend`、`NodeBackend`、`SandboxBackend`；未注入 backend 时 capability
   fail closed，不创建隐式系统浏览器；
3. session：`SessionRegistry`、`SessionRunRegistry`、`SessionStateRegistry`；registry 可被多个
   handler 并发使用，调用方仍负责租户隔离和持久化；
4. host ports：`PublishArtifact`、`LocalPlannerChat`、`RunCommand`、`RepairScript`、
   `AcceptanceScript`。这些能力必须由 Host 显式注入；canonical package 不搜索 HS 路径，
   不自行执行命令，也不读取模型/provider 配置。

`EnabledTools` 为空时会按 backend 实际 capability 注册可支持的 surface；显式列表只会缩窄
surface，不会把 backend 不支持的能力伪装为可用。

## Backend 与可选能力

`BrowserBackend` 覆盖 open、navigate、tabs、extract、snapshot、screenshot、click、type、eval。
下载/PDF/HTML、console、request/response、cookie/storage、trace、profile lifecycle、upload、
drag/fill 等能力通过独立小接口按需实现。`BrowserCapabilityProvider` 可声明更窄能力；缺失能力
返回稳定的 unsupported/fail-closed 结果。

具体 browserd 进程可由 [`browser/host/browserd`](../host/browserd/API.md) 管理，但本 package
不会自动构造或启动它。代理 transport、系统浏览器、登录态和真实网络策略属于 Host adapter。

### Experimental Host Adapter seam

具体 backend 若需要 route、artifact、doctor metadata 或 locator 投影，可使用
`BrowserResolvedExecutionRoute`、`BrowserArtifactResolveRequest`、
`BrowserDoctorRouteMetadataProvider` 以及 `ResolveBrowserElementTargetWithHint` 等
`host_adapter.go` 合同。该 seam 只暴露 portable value/helper，不拥有具体 proxy、provider、
credential 或产品策略；当前仍是 Experimental，Beta 前需要进一步收口。

## 参数、错误与 identity

- `DecodeToolArguments` 与 `CanonicalizeToolArguments` 接受现有兼容输入并输出 canonical JSON；
- 参数错误可通过 `errors.As(err, *ToolArgumentError)` 识别，保留稳定 code、field 与 repair；
- `WithToolSessionID` 把调用方 session identity 传入 Browser session 协调；
- `WithToolRuntimeNetworkGuard` 只注入宿主已决定的网络 guard，不替代 authorization、approval
  或 sandbox；
- backend error、resolver outcome、JSON 字段和结果状态沿用 `browser/runtime` 合同。

## 取消、并发与副作用

- handler 和 backend port 传播调用方 `context.Context` cancellation/deadline；
- registry/state registry 可并发使用；自定义 backend 与注入的 port 必须自行声明并保证并发边界；
- artifact 发布只经 `PublishArtifact`，command 只经 `RunCommand`，模型决策只经
  `LocalPlannerChat`；未注入时对应动作 fail closed；
- 本 package 不拥有全局进程，因此没有隐式 `Shutdown`。具体 Host 必须提供有界、幂等关闭。

## Browser Local Planner

`EvaluateBrowserLocalPlannerEligibilityForActResult`、
`BuildBrowserLocalPlannerContextForActResult` 和 `BrowserLocalPlannerDryRunForActResult` 是纯机制。
只有显式启用 `BrowserLocalPlannerExecute` 且注入 `LocalPlannerChat` 后，才会调用模型 adapter；
模型、prompt policy、配额和安全审批仍由 Host 决定。

## Non-goal

- 默认 provider、credential、proxy、登录态或生产网络；
- 自动下载/启动 browserd 或系统浏览器；
- Host command 授权、sandbox、approval 和客户 allowlist；
- Scene/业务 route、真实 RunStore backend 或产品默认值；
- Public/Beta/Stable、semver 兼容承诺或正式发行授权。
