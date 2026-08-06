# `runtime/execution` API

状态：**v0.2.1 Developer Preview candidate**。本包进入9包核心兼容候选面及
签名/文档漂移门禁，但不是Beta、Stable或生产SLA。

`runtime/execution` 是根 `agentx.Client` 与具体执行宿主之间的最窄 Run adapter。
它负责请求分派、结果组装、Shutdown转发和 error classification委托，不负责模型、
工具、Workflow、存储或业务策略。

## 构造

```go
runtime, err := execution.New(host)
client, err := agentx.New(agentx.Config{Adapter: runtime})
```

`host == nil` 时 `New` 返回错误。成功把 Runtime交给 `agentx.Client` 后，调用方不应
再直接调用 Runtime或 Host；Client拥有 Run和 Shutdown访问权。

## Host 合同

```go
type Host interface {
    Run(context.Context, Request) (*Result, error)
    Shutdown(context.Context) error
    ClassifyError(error) agentx.ErrorCode
}
```

- `Run` 持有 concrete model/tool/backend执行；必须遵守调用方 context；
- `Shutdown` 必须有界、支持重复调用，并允许前一次 context到期后继续收敛；
- `ClassifyError` 只把宿主错误投影为根 AgentX稳定 error code；
- Host不得把 credential、provider secret或 display-unsafe错误写入 Result。

## 请求与结果

```go
type Request struct {
    Input     string
    SessionID string
}

type Result struct {
    RunID     string
    SessionID string
    Status    string
    Reply     string
}
```

Runtime不二次 trim或验证字段。根 `agentx.Client` 负责空输入校验、SessionID规范化、
status/blocker/next-action映射和 display-safe typed error。

Host可在返回 error时同时返回 partial Result；Runtime会保留该结果和原始 error
identity。Host返回 nil Result时，Runtime不伪造成功结果。

## 并发与生命周期

Runtime本身不新增并发 gate。根 `agentx.Client` 串行化同一 Client的重叠 Run，并
定义 Shutdown开始后的拒绝语义。Runtime只把 Run和 Shutdown原样传递给 Host。

## 当前不支持

- 直接构造 concrete model、tool、provider或 backend；
- 根 Facade的 Workflow、Objective、Resume或长任务入口；
- Scene registry、credential discovery或真实网络副作用；
- Public、Beta或Stable兼容性承诺。

Workflow portable Runtime继续由 `runtime/workflow`及其子 package独立提供；本
package没有以空实现或枚举宣称首版根 Facade支持 Workflow。
