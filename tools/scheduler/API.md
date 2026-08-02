# `tools/scheduler` API 参考

`scheduler` 提供 `cron` 工具的可移植命令路由：解析并规范化 `add/list/status/run/remove`，
返回 typed argument error，并把 `context` 与参数的防御性副本交给 Host Backend。

成熟度：`Experimental extension`。当前不构成 Public、Beta 或 Stable 兼容性承诺。

## 接入

```go
scheduler.Register(registry, scheduler.BackendFuncs{
    AddFunc: add,
    ListFunc: list,
    StatusFunc: status,
    RunFunc: run,
    RemoveFunc: remove,
})
```

`Backend` 的五个方法分别对应一个 action。canonical 不提供默认 scheduler，不读取系统 cron，
也不创建 goroutine、进程、队列或数据库。

## Host 责任

Host 继续拥有：

- task/session identity 与 visibility；
- authorization、approval、operator permission；
- concrete scheduler/queue/RunStore backend；
- durable write 顺序、幂等、重试、取消和恢复；
- lane、queue mode、timeout 与产品默认策略；
- credential、真实进程和部署环境。

`Request.Arguments` 是已解析 JSON object 的防御性浅复制。Host 必须按自身合同校验 action-specific
字段，不能把它当成绕过授权的自由命令。

## 错误、取消与并发

- 缺少 `action` 或 action 不受支持时返回 `runtime/toolerrors.ToolArgumentError`；
- Backend 方法缺失时 fail closed；
- 取消/deadline 原样传给 Backend；
- handler 无共享可变状态，Backend 的并发、事务和 durable ordering 由 Host 保证。
