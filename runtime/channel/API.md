# `runtime/channel` 中文 API Reference

成熟度：**Experimental extension / private validation**。

本包提供与具体平台无关的消息入口、sender/runner port、有界内存协调和 session delivery
合同。它不连接飞书或其它平台，不读取 credential，不审批账号，不保存 pairing 状态，也不
执行生产网络发送。

## 消息与 Host ports

主要类型：

- `Message`、`TextTarget`：平台适配器提供的消息与回复目标；`Raw` 仍是 Host 持有的
  opaque JSON，不由 Core 解析业务规则。
- `TextSender`、`ReplySender`、`EditSender`、`DeleteSender`、`ReactSender`、
  `ForwardSender`：Host 实现的发送能力 ports。
- `TurnRunner`：Host 实现的单轮执行 port；本包不拥有 engine、模型或工具 backend。
- `ToolContext`：把当前平台、目标和 sender 显式传给 Host 工具层。

## Routing、chunking 与 dedupe

```go
type RoutedRunner struct { ... }
type AccountSenders struct { ... }
type ChunkingSender struct { ... }
type Deduper struct { ... }

func SplitText(text string, limit int) []string
func BuildContentDedupeKey(message Message) string
```

`RoutedRunner`和`AccountSenders`只按显式 binding/account选择调用方提供的实现；
`ChunkingSender`按 rune 边界拆分文本；`Deduper`只提供进程内、带 TTL 的 reservation，
不是 durable idempotency store。

## Bounded ingress runtime

```go
func NewIngressRuntime(IngressRuntimeOptions) *IngressRuntime
func (r *IngressRuntime) Submit(context.Context, InboundProcessor, Message) IngressSubmitResult
func (r *IngressRuntime) Shutdown(context.Context) error
```

`IngressRuntimeOptions`显式控制最大并发、queue容量、`reject`/`wait` overload策略和
`Close`默认等待上限。`Submit`只有返回`accepted`时才转移消息处理所有权；重复、过载、
关闭、runtime缺失和调用方context取消都有稳定`IngressSubmitReason`。

`Shutdown(ctx)`有界且幂等：停止接收、以`ErrIngressRuntimeClosed`取消 active work、释放
queued reservation并等待worker退出；等待超时返回包装后的`ctx.Err()`。关闭后`Submit`
返回`closed`。调用方传给`Submit`的context只约束等待入队，不会在成功入队后取消已转移的
任务；Host shutdown和`InboundProcessor.Timeout`仍提供运行边界。

## Session delivery contract

```go
func BuildSessionChannelContract(SessionChannelContractInput) SessionChannelContract
func DeterministicSessionKey(SessionSource) string
func ChannelCapabilityFromSender(TextSender, int) ChannelCapability
```

合同把session source、lifecycle、channel capability、delivery status、missing input、
blocker和next Host action规范化为display-safe结果。它明确声明平台delivery归Host adapter
所有，channel capability不控制Agent执行policy，raw platform output不会进入合同。

## 最小示例

```go
runtime := channel.NewIngressRuntime(channel.IngressRuntimeOptions{
    MaxConcurrency: 4,
    QueueCapacity:  64,
    OverloadPolicy: channel.IngressOverloadReject,
})
defer runtime.Shutdown(context.Background())

result := runtime.Submit(ctx, channel.InboundProcessor{
    Runner: hostRunner,
    Sender: hostSender,
}, message)
if !result.Accepted {
    // 根据 result.Reason 做 Host 侧重试、降级或用户反馈。
}
```

## Fixed-version external consumer

[`runtime/conformance/channel-consumer`](../conformance/channel-consumer)固定依赖
`v0.0.0-20260801203033-9f67493e1e32`，不使用`replace`，也不import HS、Runner、Scene、
platform SDK或backend。它以调用方提供的内存runner/sender验证消息所有权转移、回复、
有界Shutdown和session delivery合同：

```bash
cd runtime/conformance/channel-consumer
GOWORK=off go test ./...
GOWORK=off go run .
```

预期输出：

```text
agentx-channel-ok:sent:channel-conformance:session_channel_ready
```

## 非目标

- pairing challenge、账号审批、allowlist和durable pairing state；
- concrete Feishu/Slack/Teams或其它平台SDK、credential和transport；
- durable queue、分布式dedupe、scheduler或RunStore backend；
- engine/Runner、模型、工具、access evaluator产品策略；
- Scene业务规则、真实网络发送或Public/Beta/Stable承诺。
