# Objective Host Kit 接入指南

Objective Host Kit适合“Host拥有真实执行与业务授权，AgentX拥有可移植控制语义”的场景。
它不是自动读取凭据、自动选工具或自动写业务系统的黑盒Agent。

## 标准接入步骤

1. Host把原始请求整理为display-safe goal digest、success criteria和required evidence；
2. Host加载自己的execution policy、budget、approval、strategy catalog和adapter registry；
3. 用`objectivehostkit.New`注入默认handler或按adapter ref注册handlers；
4. 首次可以保持dispatch关闭，只检查ingress/readiness；
5. Host完成authorization/approval后，显式开启并确认dispatch；
6. handler执行真实backend，并用`objective.BuildRuntimeAdapterResult`报告observations/evidence；
7. Host Kit完成normalization和Objective verification，返回completed、partial或blocked。

## 最小构造

```go
runtime, err := objectivehostkit.New(objectivehostkit.Config{
    Handlers: map[objective.DisplaySafeRef]objectivehostkit.Handler{
        "adapter:readonly_metric": func(
            ctx context.Context,
            request objective.RuntimeAdapterRequest,
        ) objective.RuntimeAdapterResult {
            // Host在这里调用自己的backend，并负责authorization、credential和审计。
            return objective.BuildRuntimeAdapterResult(objective.RuntimeAdapterResultInput{
                Request:           request,
                AdapterRef:        request.AdapterRef,
                StrategyRef:       request.StrategyRef,
                HostAdapterRunRef: "adapter_run:readonly_metric",
                Status:            objective.VerificationSatisfied,
                Observations:      observations,
                EvidenceRefs:      evidence,
            })
        },
    },
})
if err != nil {
    return err
}

result := runtime.Run(ctx, objectivehostkit.RunRequest{
    Ingress:               ingress,
    DispatchEnabled:       true,
    DispatchHostConfirmed: true,
})
```

完整`ingress`必须包含policy、budget、approval、strategy catalog、adapter registry、
idempotency和input refs；缺项时返回结构化blocker，而不是静默采用产品默认。

## 两阶段运行

生产Host可以先用dispatch关闭的同一`Run`做preflight；`ReadyForDispatch=false`时读取status、
missing inputs和next action。Host确认后再执行显式dispatch。Host Kit不持久化preflight，
`IdempotencyRef`会进入canonical request，但真正的去重仍由Host backend负责。

## 取消、并发和副作用

- context原样传给handler；Host负责让backend响应取消和deadline；
- 每个`Run`至多调用一个handler一次，Host Kit不自动retry；
- `Runtime`不持有单次Run state，不创建goroutine，handler map在`New`时复制；
- handler并发安全、credential生命周期、连接池和Shutdown由Host负责；
- side-effect adapter仍必须在ingress policy/approval中显式允许。

## 与HS迁移的关系

HS产品入口继续拥有产品refs、配置来源、operator approval和具体handlers；其通用dispatch、
normalization和verification实现改由本package持有。HS兼容类型只用于平滑迁移，不是第二份
source authority。

## 当前边界

本路径完成C5 Objective的一次同步Host adapter闭环，但不等于C6长任务。Task、Session、
Subagent、Scheduler、Resume和durable backend将在后续能力级wave完成。
