# `runtime/scheduler` API

`runtime/scheduler`提供AgentX长任务使用的portable调度内核：typed Job/Result、三条默认
lane、并发安全内存队列、可热切换QueueProxy、bounded Dispatcher、lease heartbeat、
cancellation、retry/dead-letter终态协调和metrics。

当前成熟度是 **Experimental extension / private validation**。它不创建系统进程、不读取
环境变量、不安装launchd/systemd、不拥有SQL/file backend，也不决定产品优先级、授权或
重试策略。

## Queue合同

```go
type Queue interface {
    Enqueue(context.Context, Job) error
    Dequeue(context.Context, Lane) (Job, error)
    Ack(context.Context, Result) error
    Fail(context.Context, Result) error
    Result(context.Context, string) (Result, bool, error)
    Pending(context.Context, string) (bool, error)
}
```

`CancelableQueue`、`KindAwareQueue`、`HeartbeatCapableQueue`、
`RuntimeVisibleQueue`和`LeaseIdentityProvider`都是可选能力。调用方必须按capability
判断，不得假定所有backend都支持取消、按kind租约、heartbeat或跨进程readback。

## `MemoryQueue`

`NewMemoryQueue(QueueConfig)`提供并发安全、进程内实现，支持lane limit、result TTL/
limit、typed cancellation和kind-aware dequeue。它用于测试、示例和短生命周期Host，
不提供durability、跨进程lease、备份或生产恢复。

## `QueueProxy`

`QueueProxy`允许Host替换concrete backend，并把已租约Job的Ack/Fail/Heartbeat路由回原
backend，避免切换期间把终态写入错误队列。Host仍负责backend生命周期和迁移策略。

## `Dispatcher`

`NewDispatcher`按lane启动有界worker，固定执行Dequeue→handler→Ack/Fail顺序，并在
context取消、handler panic、lease lost和terminal persistence失败时fail closed。
`HeartbeatFailureObserver`是唯一日志/遥测port；canonical不依赖具体logger。

`WaitContext`只等待Dispatcher拥有的worker退出。Queue、handler创建的外部资源和具体
backend必须由Host关闭。

## 错误与安全边界

- `ErrQueueEmpty`不是终态失败；worker会按poll interval继续；
- `ErrLeaseLost`阻止stale handler成功结果覆盖新租约；
- handler panic通过`telemetry/safeerror`转换为display-safe摘要；
- `Result.Outcome`是控制流依据，`Result.Error`只用于安全诊断；
- raw credential、provider、Scene payload和真实副作用策略不得进入该package。

## 验证

```bash
cd runtime
GOWORK=off go test ./scheduler
GOWORK=off go test -race ./scheduler
GOWORK=off go vet ./scheduler
```
