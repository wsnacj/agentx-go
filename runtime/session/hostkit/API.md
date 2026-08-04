# `runtime/session/hostkit` API

`runtime/session/hostkit` 是 AgentX C6 Task/Session/Subagent 的首个推荐构造入口。它把一次
Host-owned child worker 生命周期固定为：

```text
readiness -> invoke exactly once -> durable record -> record readback
-> worker result readback -> parent verification -> optional Objective handoff
```

当前成熟度是 **v0.1.0 Developer Preview candidate**，进入首版9包核心兼容候选面。
它同时提供一次child worker闭环和bounded Resume Runtime；它不是进程管理器、系统
scheduler、Runner或开箱即用provider，也不构成Beta、Stable或生产SLA。

## 最小入口

```go
runtime, err := sessionhostkit.New(sessionhostkit.Config{
    Worker:     worker,
    Store:      store,
    BackendRef: "backend:my_child_worker",
    Durable:    true,
})
if err != nil {
    return err
}
defer runtime.Shutdown(context.Background())

result, err := runtime.Run(ctx, sessionhostkit.RunRequest{
    BackendInput: input,
})
```

`Worker`与`Store`必须由Host显式注入。`New`不启动goroutine、process、queue或网络。

## `Config`

| 字段 | 语义 |
| --- | --- |
| `Worker WorkerRuntime` | 调用并回读一个Host-owned child worker；必填 |
| `Store StateStore` | 写入并回读`WorkerRunRecord`；必填 |
| `BackendRef DisplaySafeRef` | display-safe backend identity |
| `BackendKind string` | 可选Host backend分类；空值使用canonical默认 |
| `Durable bool` | 声明Host store是否durable；不会自动创建durable backend |
| `Now func() time.Time` | 测试或Host提供的clock；空值使用UTC当前时间 |

缺少或typed-nil port时，`New`返回`*ConfigError`；
`errors.Is(err, sessionhostkit.ErrInvalidConfig)`为true。

## `Runtime.Run`

```go
type RunRequest struct {
    BackendInput BackendInput
    Closure      *ObjectiveRuntimeClosureInput
    Boundaries   []controlcontract.Boundary
}
```

- `BackendInput.Enabled`必须显式为true；默认关闭并以结构化blocked结果返回；
- readiness、policy、approval、budget、allowed tools、context refs和idempotency均由Host
  在canonical control contract中准备；
- 每次`Run`至多调用一次`WorkerRuntime.InvokeDelegationWorker`；
- completed result先写`StateStore`，再从store回读并逐字段核对，之后才调用
  `ReadDelegationWorkerResult`；
- store record或worker readback不一致时fail closed，`Completed=false`；
- child output始终保持`WorkerOutputAcceptedAsFact=false`与
  `WorkerResultRequiresVerification=true`；
- `Closure != nil`时，Host Kit继续完成parent merge、evidence verification、Objective
  handoff与下一步Host persist投影；不替Host写Objective RunStore。

`RunResult.Completed`只表示本次Host Kit路径完成到可接受的readback/parent handoff阶段，
不表示业务目标已经正式发布、外部副作用已批准或后台任务已经被scheduler管理。

## Host ports

### `WorkerRuntime`

```go
type WorkerRuntime interface {
    InvokeDelegationWorker(context.Context, WorkerRequest) (WorkerResult, error)
    ReadDelegationWorkerResult(context.Context, WorkerReadbackRequest) (WorkerReadback, error)
}
```

Host可用本地process、远端worker、现有Runner adapter或测试double实现该接口。API不包含
command、credential、provider或具体transport字段；这些必须封装在Host实现内部。

### `StateStore`

```go
type StateStore interface {
    WriteWorkerRun(context.Context, WorkerRunRecord) error
    ReadWorkerRun(context.Context, controlcontract.DisplaySafeRef) (WorkerRunRecord, bool, error)
}
```

`NewInMemoryStateStore`只用于测试、示例和单进程短生命周期。它是并发安全实现，但不提供
durability、事务、跨进程恢复、retention或生产备份。

## Scheduler / Resume / Long Task入口

```go
runtime, err := sessionhostkit.NewResumeRuntime(sessionhostkit.ResumeConfig{
    Queue:  queue,
    Worker: resumeWorker,
    Lane:   scheduler.LaneBackground,
})

_, err = runtime.Enqueue(ctx, sessionhostkit.ResumeEnqueueRequest{
    Enabled:       true,
    Payload:       payload,
    TrustedCaller: true,
})
result, err := runtime.Run(ctx, sessionhostkit.ResumeRunRequest{
    Enabled:          true,
    MaxCycles:        4,
    MaxTicksPerCycle: 8,
})
```

`ResumeRuntime`组合`runtime/scheduler`、continuation readback和Host wake dispatch：

- Queue、continuation store readback和dispatch必须由Host显式注入；
- tick payload只接受display-safe refs，并按kind-aware queue租约；
- 每个tick固定执行readback→validation→dispatch request→Ack/Fail；
- `MaxCycles`和`MaxTicksPerCycle`保证单次`Run`有界；
- 同一runtime只允许一个service loop，重入返回`ErrResumeRuntimeBusy`；
- `Shutdown(ctx)`幂等、取消当前loop并有界等待；关闭后Enqueue/Run返回
  `ErrResumeRuntimeClosed`；
- `HostRuntimeDispatchByHost=true`只说明Host callback记录了dispatch，不表示canonical直接
  调用了LLM、Runner、tool或workflow。

内存Queue只适合测试或单进程短生命周期。生产Host必须注入durable queue/continuation
store，并拥有lease、retention、backup、operator approval和process/service管理。

## 取消、并发和关闭

- `Run`将调用方`context.Context`原样传给Worker和Store；Host实现必须响应取消与deadline；
- Worker返回错误时，Host Kit把它收口为结构化failure review，不泄漏任意原始输出；
- 同一`Runtime`可并发调用；前提是注入的Worker/Store也支持相同并发度；
- `Shutdown(ctx)`有界且幂等，因为Host Kit不拥有后台worker；
- canceled shutdown context返回`ctx.Err()`且不改变状态；
- 成功Shutdown后的`Run`返回`ErrClosed`，可用`errors.Is`判断；
- Worker、process、queue与store的真实关闭仍由Host负责。

## JSON与display-safe边界

`BackendReport`、`ObjectiveRuntimeClosureProfileReport`及嵌套control contract保持既有JSON
字段、状态和错误顺序。任何无法通过`NormalizeDisplaySafeRef`的引用、raw output标记、
record mismatch或readback mismatch都会阻断完成。Host不得把原始child tool日志作为
`DisplaySafeRef`伪装输入。

## 明确不拥有

- 系统scheduler、平台service、具体queue/store backend和产品retry/priority policy；
- child prompt、model routing、tool catalog、approval与authorization策略；
- Runner、process、container、remote worker或provider；
- concrete RunStore/SQL/object-store backend；
- credential、Scene、网络和生产副作用；
- 自动接受child输出为parent事实。

provider/tools/browser/document/Scene分别属于P2-P4；launchd/systemd、真实durable backend
和生产运行授权仍由Host/发行阶段负责。

## 验证

```bash
cd runtime
GOWORK=off go test ./session/hostkit
GOWORK=off go test -race ./session/hostkit
GOWORK=off go vet ./session/hostkit
```

`hostkit_external_test.go`覆盖成功、cancellation、readback mismatch、typed-nil config、
幂等Shutdown和关闭后调用；`resume_hostkit_external_test.go`覆盖enqueue/run、typed busy、
bounded shutdown和关闭后调用；同包测试复用迁移前HS差分用例验证record/readback顺序、
parent verification、async completion、Objective handoff和resume daemon状态聚合。
