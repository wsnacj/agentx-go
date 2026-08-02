# `runtime/session/resume` API

`runtime/session/resume`提供P1-D长任务恢复机制的Experimental实现：display-safe tick
payload、continuation readback校验、显式Host wake dispatch、单次daemon tick、bounded runner
和bounded service loop。

当前成熟度是 **Experimental extension / private validation**。普通调用方优先使用
[`runtime/session/hostkit.NewResumeRuntime`](../hostkit/API.md)；只有需要替换daemon/service
组合方式的高级Host才应直接依赖本package。

## 推荐短名称

- `Worker` / `WorkerConfig`：一次readback→validate→wake dispatch协调；
- `Daemon` / `DaemonConfig`：enqueue和单次queue tick处理；
- `Service` / `ServiceInput` / `ServiceReport`：有界多cycle运行；
- `EnqueueRequest` / `EnqueueResult`：display-safe tick入队。

历史较长的`ObjectiveRuntimeSchedulerResume*`名称用于HS迁移期的签名兼容，不应作为新项目
的首选入口。

## 固定语义

- tick payload先normalize，再拒绝unsafe ref或缺失readback identity；
- 每个tick固定执行continuation readback、验证、Host wake dispatch、Ack/Fail；
- canonical只记录Host dispatch事实，不直接调用Runner、LLM、tool或Workflow；
- runner和service都受调用方提供的最大tick/cycle约束；
- context取消、readback mismatch、dispatch失败和queue终态失败均fail closed；
- JSON字段、状态、错误顺序和terminal coordination保持HS迁移前行为。

## Host责任

Host必须提供concrete queue、continuation store/readback和wake dispatch，并继续拥有process、
system scheduler、credential、authorization、租户priority/retry policy、durable backend、
部署、监控和真实副作用。

## 验证

```bash
cd runtime
GOWORK=off go test ./session/resume
GOWORK=off go test -race ./session/resume
GOWORK=off go vet ./session/resume
```
