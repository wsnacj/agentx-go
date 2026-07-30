# runtime/promptcontext API

导入路径：

```go
import promptcontext "github.com/wsnacj/agentx-go/runtime/promptcontext"
```

成熟度：**Experimental / private validation**。

该 package 定义 prompt rendering 所需的当前时间、timezone、session identity
和 model identity。它不是 Go `context.Context` 的替代品，也不拥有 Run
生命周期、Runner、execution controls、provider 或 Runtime construction。

## 数据类型

```go
type Context struct {
    Now       time.Time
    Timezone  string
    SessionID string
    Model     string
}

type BuildInput struct {
    Now       time.Time
    Timezone  string
    SessionID string
    Model     string
}
```

字段保持调用方输入，不执行 trim、normalization 或 validation。`Context` 是值
类型；本 package 不保存全局状态。

## 构造

```go
func Build(input BuildInput) Context
```

- `input.Now` 非零时原样保留；
- `input.Now.IsZero()` 时使用调用时的 `time.Now()`；
- 其它字段原样复制；
- 不读取环境变量、配置文件或 provider。

需要确定性时间的调用方应显式提供 `Now`。

## 时间文本

```go
func (c Context) TimestampText() string
```

- `Timezone` 为空时，按 `c.Now.Format(time.RFC3339)` 输出；
- `time.LoadLocation(Timezone)` 成功时，先转换到该 location，再输出 RFC3339；
- timezone 无效时 fail-soft，回退为原 `Now` 的 RFC3339；
- 不返回 error，不推断 locale，也不改变 `Context`。

location 与夏令时规则由当前 Go runtime 的 timezone database 决定。

## 与公共 Runtime 的边界

`github.com/wsnacj/agentx-go/runtime` module root 仍保留给未来窄 Runtime
construction 合同。`promptcontext` 不提供：

- `agentxruntime.New`、Client、Run 或 Shutdown；
- cancellation/deadline 传播；
- prompt 模板或 prompt module；
- model/provider 选择；
- session persistence、telemetry 或 evidence；
- Public、Beta、Stable 或 production-ready 声明。
