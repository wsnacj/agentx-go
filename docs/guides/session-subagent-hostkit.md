# Task / Session / Subagent Host Kit 接入

> 当前可复现的private Developer Preview版本：
> `v0.0.0-20260802091415-920282587efc`。该pseudo-version用于验证，不是正式tag或兼容承诺。

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

## Scheduler / Resume / Long Task

需要把display-safe continuation ref入队，并按有界service loop恢复执行时，组合同一
Host Kit中的`ResumeRuntime`：

```go
resume, err := sessionhostkit.NewResumeRuntime(sessionhostkit.ResumeConfig{
    Queue:  myQueue,
    Worker: myResumeWorker,
    Lane:   scheduler.LaneBackground,
})
if err != nil {
    return err
}

_, err = resume.Enqueue(ctx, sessionhostkit.ResumeEnqueueRequest{
    Enabled:       true,
    Payload:       payload,
    TrustedCaller: true,
})
if err != nil {
    return err
}

report, err := resume.Run(ctx, sessionhostkit.ResumeRunRequest{
    Enabled:          true,
    MaxCycles:        4,
    MaxTicksPerCycle: 8,
})
```

`myQueue`可以是测试用`runtime/scheduler.NewMemoryQueue`，生产环境则应是Host提供的
durable backend。`myResumeWorker`显式提供continuation readback和wake dispatch；canonical
不解析credential、不选择provider，也不直接调用Runner、LLM或Scene。

## 与长任务平台的边界

本入口拥有一次child worker invoke/record/readback/verification生命周期。以下仍由Host或
canonical Scheduler/Resume入口负责enqueue、bounded tick处理和terminal coordination；
以下继续由Host负责：

- 产品priority、租户配额和动态lane policy；
- process/container生命周期；
- retry/backoff、dead letter和cancel cascade；
- concrete wake transport、resume token存储和跨进程恢复；
- production durable backend与运维。

因此它已经提供长任务的portable调度/恢复内核，但不是完整系统scheduler或生产平台。

## 取消与关闭

child `Run`的context必须传到Worker和Store。`ResumeRuntime`同一时刻只允许运行一个
service loop，重入返回`ErrResumeRuntimeBusy`；`Shutdown(ctx)`幂等、取消活动loop并有界
等待，关闭后的`Enqueue`和`Run`返回`ErrResumeRuntimeClosed`。Queue、Worker、Store与
process仍由创建它们的Host关闭。

## 可运行consumer

仓库中的`runtime/conformance/session-hostkit-consumer`使用固定private pseudo-version，
不依赖HS、Runner、Scene、provider、网络或长期`replace`，用于验证外部项目的真实引用方式。
