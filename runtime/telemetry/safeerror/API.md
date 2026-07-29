# `runtime/telemetry/safeerror` 中文 API Reference

## 当前定位

导入路径：

```go
import agentxsafeerror "github.com/wsnacj/agentx-go/runtime/telemetry/safeerror"
```

本 package 把 Go `error` 投影为可以进入 telemetry、日志、artifact 或 operator
界面的安全结构。它从 HS `core/agentx/telemetry/safeerror` 原样迁移，是该
projection 算法的唯一 source authority，目前处于
**private validation / Experimental**。

它不是根 `agentx.Error` 的替代品，也不负责业务错误枚举、retryability、
HTTP status、blocker、next action 或 provider policy。调用方仍拥有 class/code
的选择与业务解释。

## Projection

```go
type Projection struct {
    Class    string `json:"class,omitempty"`
    Code     string `json:"code,omitempty"`
    Identity string `json:"identity,omitempty"`
}
```

- `Class`：错误大类，例如 `runtime`、`permission`；
- `Code`：调用方拥有的稳定错误码；
- `Identity`：由错误链或显式 material 生成的稳定 SHA-256 hex；
- 三个 JSON 字段都使用 `omitempty`。

projection 不包含原始 error message。不要把 secret、credential、完整请求或
provider payload 放进 class/code。

## Wrap 与 cause

```go
func Wrap(cause error, message string) error
func WrapWithIdentity(cause error, message string, identity string) error
```

行为：

- nil cause 返回 nil；
- `message` trim 后作为 display-safe `Error()`；
- 空 message 固定回退为 `safe error`；
- `Unwrap` 保留 cause，`errors.Is/As` 可继续检查底层错误；
- `WrapWithIdentity` 的非空 identity 优先；空 identity 自动从 cause chain
  计算；
- display message 不拼入 cause message。

调用方必须先决定可展示的 message，不能把 provider/credential 原文直接传入。

## Project

```go
func Project(err error, class, code string) Projection
func ProjectWithIdentity(err error, class, code, identity string) Projection
func ProjectText(value, class, code string) Projection
```

class/code normalization：

- trim 并转为 lowercase；
- 保留 `a-z`、`0-9`、`-`、`.`；
- 其它连续字符折叠为单个 `_`；
- 最大 64 chars；
- 空 class/code 回退为 `error` / `unknown`。

identity precedence：

1. `ProjectWithIdentity` 的显式非空 identity；
2. `WrapWithIdentity` 或 `Wrap` 保存的 identity；
3. 当前 error chain 的 identity。

`Project(nil, ...)` 仍返回 normalized class/code，但 identity 为空。
`ProjectText` 只对非空 trimmed value 计算 identity。

## Identity

```go
func Identity(material string) string
```

输入先 trim；空输入返回空字符串，其它输入返回 lowercase SHA-256 hex。

error chain 的自动 material 按外层到内层连接：

```text
%T:%s
%T:%s
...
```

因此 error concrete type、message 或 wrap 层级变化会改变 identity。identity
适合去重和关联 observation，不是授权凭据、签名或跨任意实现永久不变的业务 ID。

## Summary 与 map projection

```go
func Summary(Projection) string
func AppendAttrs(map[string]any, string, Projection) map[string]any
func AppendDetails(map[string]string, string, Projection) map[string]string
```

`Summary` 的字段顺序固定为 class、code、identity；全空 projection 返回：

```text
class=error code=unknown
```

`AppendAttrs`/`AppendDetails`：

- nil map 自动分配；
- 非 nil map 原位追加；
- prefix 只 trim，不自动添加 `_` 或 `.`；
- 只写非空字段；
- key 为 `<prefix>error_class`、`<prefix>error_code`、
  `<prefix>error_identity`。

## 并发与生命周期

package 没有 global mutable state、goroutine、I/O 或 shutdown 生命周期。
`Projection` 是普通 value。map helper 会修改传入 map；同一个 map 不应被多个
goroutine 无同步并发写入。

## 安全边界

- identity 是不可逆摘要，不等于原文已安全保存；仍应避免把高敏原文作为
  material 传播到其它位置；
- class/code 必须来自调用方受控枚举，不能直接接收用户输入；
- 本 package 不做 redaction、审计批准或 retention policy；
- telemetry/operator surface 应保存 projection，不保存 raw cause。

## 当前非目标

- 根 AgentX typed error / ErrorCode；
- retryability、HTTP status、blocker、next action；
- provider error mapping；
- telemetry backend、exporter 或 storage；
- Scene 业务错误分类；
- Public/Beta/Stable 或 production-ready 声明。
