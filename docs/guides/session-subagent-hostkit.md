# Task / Session / Subagent Host Kit 接入

当业务需要把一个父Objective拆给child worker，并在结果回到父上下文前强制完成持久化
回读和verification时，使用`runtime/session/hostkit`。普通模型对话或Open Tool Loop仍应
使用`runtime/hostkit`；显式图执行使用`runtime/workflow/hostkit`。

## Host必须提供什么

1. `WorkerRuntime`：真实调用child worker并提供独立readback；
2. `StateStore`：保存`WorkerRunRecord`并能按`WorkerRunRef`回读；
3. control contract readiness：worker identity、允许工具、budget、approval、policy、
   idempotency、evidence与parent verification引用；
4. 如果需要继续Objective loop，再提供`ObjectiveRuntimeClosureInput`。

Host Kit不读取环境变量、不加载credential、不选择模型、不创建process，也不自动加入队列。

## 最小流程

```go
kit, err := sessionhostkit.New(sessionhostkit.Config{
    Worker:     myWorker,
    Store:      myStore,
    BackendRef: "backend:research_child",
    Durable:    true,
})
if err != nil {
    return err
}

result, err := kit.Run(ctx, sessionhostkit.RunRequest{
    BackendInput: input,
    Closure:      closure, // 不需要Objective handoff时传nil
})
```

调用方应同时检查`err`、`result.Completed`、`result.Status`、
`result.Backend.FailureClass`、`MissingInputs`与`NextHostAction`。`err == nil`并不自动代表
child结果已通过parent verification；被策略或证据门禁阻断的路径使用结构化结果表达。

## 与长任务的边界

本入口拥有一次child worker invoke/record/readback/verification生命周期。以下仍由Host或
后续P1-D Scheduler/Resume入口负责：

- enqueue、lane、priority和并发配额；
- process/container生命周期；
- retry/backoff、dead letter和cancel cascade；
- wake signal、resume token、跨进程恢复；
- production durable backend与运维。

因此它可以作为长任务的一次安全执行单元，但本身不是完整scheduler。

## 取消与关闭

`Run`的context必须传到Worker和Store。Host Kit自身不拥有后台goroutine，
`Shutdown(ctx)`只关闭接入面且幂等；Worker、Store与process由创建它们的Host关闭。

## 可运行consumer

仓库中的`runtime/conformance/session-hostkit-consumer`使用固定private pseudo-version，
不依赖HS、Runner、Scene、provider、网络或长期`replace`，用于验证外部项目的真实引用方式。
