# `tools/agent` API 参考

`agent` 提供面向模型的 Task、Subagent 与 bounded Agent Step 工具协调层。它拥有工具定义、
兼容参数解析、typed argument error、Subagent action 路由、`message -> seed_message` 规范化与
fanout 数量合同；具体 Session Store、Scheduler、worker、visibility 和 durable lifecycle 由 Host 注入。

成熟度：`Experimental extension`。当前不构成 Public、Beta 或 Stable 兼容性承诺。

## 最小接入

```go
registry := tools.NewRegistry()
agent.Register(registry, agent.Options{
    Backend: agent.BackendFuncs{
        TaskFunc: func(ctx context.Context, req agent.Request) (string, error) {
            return hostTasks.Execute(ctx, req.Name, req.Arguments)
        },
        SubagentFunc: func(ctx context.Context, req agent.Request) (string, error) {
            return hostSubagents.Execute(ctx, req.Action, req.Arguments)
        },
        AgentStepFunc: runBoundedChildStep,
    },
})
```

`Options.Enabled` 为空时注册完整 cohort；非空时只注册列出的工具。支持：

- `tasks_spawn`、`tasks_wait`、`tasks_run`；
- `tasks_cancel`、`tasks_replay`、`tasks_collect`、`tasks_deadletter_list`；
- `subagents` 的 `list/status/run/fanout/cancel/replay/steer`；
- `agent_step`。

## 主要合同

```go
type Request struct {
    Name      string
    Action    string
    Arguments map[string]any
}

type Backend interface {
    ExecuteTask(context.Context, Request) (string, error)
    ExecuteSubagent(context.Context, Request) (string, error)
    ExecuteAgentStep(context.Context, Request) (string, error)
}

func Register(tool.Registrar, Options)
func Definitions() []tool.Definition
func NewTaskHandler(string, Backend) tool.Handler
func NewSubagentsHandler(Backend) tool.Handler
func NewAgentStepHandler(Backend) tool.Handler
```

工具定义可单独用于catalog、审计或自定义注册：

```go
const TasksSpawnName = "tasks_spawn"
const TasksWaitName = "tasks_wait"
const TasksRunName = "tasks_run"
const TasksCancelName = "tasks_cancel"
const TasksReplayName = "tasks_replay"
const TasksCollectName = "tasks_collect"
const TasksDeadletterListName = "tasks_deadletter_list"
const SubagentsName = "subagents"
const AgentStepName = "agent_step"

func TasksSpawnDefinition() tool.Definition
func TasksWaitDefinition() tool.Definition
func TasksRunDefinition() tool.Definition
func TasksCancelDefinition() tool.Definition
func TasksReplayDefinition() tool.Definition
func TasksCollectDefinition() tool.Definition
func TasksDeadletterListDefinition() tool.Definition
func SubagentsDefinition() tool.Definition
func AgentStepDefinition() tool.Definition
```

每个 `Request.Arguments` 都是防御性浅复制。Backend 不应修改调用方保存的 map；嵌套 object 的
深复制、事务边界与持久化隔离仍由 Host 负责。

## Host 责任

canonical 明确不拥有：

- Session/Task identity、Store schema、visibility tree；
- concrete Queue/Scheduler、lane、retry、timeout 与 durable write 顺序；
- worker、subprocess、goroutine 生命周期和 resume backend；
- approval、authorization、sandbox、tenant、credential 与产品默认策略；
- child model/provider 选择、工具授权和真实网络/文件副作用。

因此，新项目需要实现三组窄 Backend 方法，但不需要复制 HS 的历史兼容层。可从内存开发
Backend 起步，再替换为自己的 durable backend。

## 错误、取消与并发

- JSON 无效、缺少 `subagents.action`、action 不支持、run 缺少 instruction、fanout 数量不一致时，
  返回 `runtime/toolerrors.ToolArgumentError`；
- `context` 原样传给 Backend，取消与 deadline 不被吞掉；
- handler 不创建 goroutine、不保存 mutable request state；Backend 自身的并发、幂等和事务语义
  由 Host 保证；
- Backend 未注入时不注册工具，单个 Backend function 缺失时 fail closed。
