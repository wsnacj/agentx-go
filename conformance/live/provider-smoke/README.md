# 真实 Provider Smoke

本目录是一个独立、固定版本、无 HS、无 Runner、无 `replace` 的 external-style consumer。
它把以下四层组合成一次真实模型对话；可选 `tool-loop` 模式还要求真实模型生成一次
function call，由 Host 验证无副作用 marker tool 后通过 canonical Tool Direct Answer 收口：

```text
agentx.Run
  -> runtime/hostkit.NewChatClient
  -> components/llm
  -> providers/openaicompat
  -> Host 显式提供的 endpoint / model / credential
```

它属于 `conformance/live`，不是教学 `examples`，不会被普通 `go test ./...` 隐式执行。
默认不读取 credential、不访问网络；只有显式设置启用开关和全部必需参数后才会运行。

## 配置

```bash
export AGENTX_LIVE_PROVIDER_SMOKE_ENABLE=1
export AGENTX_LIVE_PROVIDER_SMOKE_BASE_URL=https://provider.example/v1
export AGENTX_LIVE_PROVIDER_SMOKE_API_KEY=your-untracked-secret
export AGENTX_LIVE_PROVIDER_SMOKE_MODEL=your-model
export AGENTX_LIVE_PROVIDER_SMOKE_MODE=chat # 或 tool-loop
```

可选参数：

```bash
export AGENTX_LIVE_PROVIDER_SMOKE_TIMEOUT=45s
export AGENTX_LIVE_PROVIDER_SMOKE_PROMPT='请只输出 AGENTX_LIVE_PROVIDER_SMOKE_OK'
export AGENTX_LIVE_PROVIDER_SMOKE_EXPECT=AGENTX_LIVE_PROVIDER_SMOKE_OK
```

运行：

```bash
GOWORK=off go test ./...
GOWORK=off go run .
```

未设置 `AGENTX_LIVE_PROVIDER_SMOKE_ENABLE=1` 时，`go run .` 会输出明确的 `skipped` JSON，
不会把“没有运行”冒充成功。启用后缺少 endpoint、API key 或 model 会直接失败。

## 验证范围

- 固定版本能够在仓库外 module 中组合；
- Runtime 能把自然语言输入交给真实 provider；
- provider 响应进入统一 `RunResult`；
- `tool-loop` 模式真实经过 model tool call、Host 参数校验和 Tool Direct Answer；
- context deadline、typed execution path 与有界 `Shutdown` 生效；
- 输出不包含 endpoint、credential 或请求 header。

marker tool 是纯函数，不验证生产工具授权、Workflow、Objective、Browser、Document、Scene 或
生产 backend。
这些能力由各自离线 conformance 与 HS 中的有界 live/NL compatibility matrix 验证。
