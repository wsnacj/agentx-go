# runtime/toolerrors API

导入路径：

```go
import toolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
```

成熟度：**Experimental / private validation**。

该 package 定义工具参数无法解码、缺少必填字段或字段值无效时的结构化错误，
并携带 deterministic repair hint。它不校验具体工具 schema，不执行 repair，
也不决定模型重试、approval、sandbox、budget 或 HTTP status。

## 错误码

```go
const (
    ToolArgumentErrorCodeInvalidJSON             = "invalid_json"
    ToolArgumentErrorCodeInvalidArgumentObject   = "invalid_argument_object"
    ToolArgumentErrorCodeInvalidArgument         = "invalid_argument"
    ToolArgumentErrorCodeMissingRequiredArgument = "missing_required_argument"
)
```

## Repair kind

```go
const (
    ToolArgumentRepairReturnValidJSONObject = "return_valid_json_object"
    ToolArgumentRepairProvideRequiredField  = "provide_required_field"
    ToolArgumentRepairFixInvalidField       = "fix_invalid_field"
    ToolArgumentRepairUseAliasURL           = "use_alias_url"
)
```

这些值是 repair hint，不等同于 repair 已获授权或已执行。

## 类型

```go
type ToolArgumentRepair struct {
    Kind string `json:"kind,omitempty"`
    From string `json:"from,omitempty"`
    To   string `json:"to,omitempty"`
}

type ToolArgumentErrorOptions struct {
    Code             string
    Detail           string
    Repairable       bool
    SafeAutorepair   bool
    MissingFields    []string
    InvalidFields    []string
    DisallowedFields []string
    AllowedRepairs   []ToolArgumentRepair
    Cause            error
}

type ToolArgumentError struct {
    Tool             string
    Code             string
    Detail           string
    Repairable       bool
    SafeAutorepair   bool
    MissingFields    []string
    InvalidFields    []string
    DisallowedFields []string
    AllowedRepairs   []ToolArgumentRepair
    Cause            error
}
```

`ToolArgumentError` 实现 `error` 和 `Unwrap() error`。`Error()` 优先使用
trim 后的 `Detail`，其次使用 cause 文本，最后返回 `invalid arguments`。
cause 可通过 `errors.Is/As` 检查。

## 构造与检查

```go
func AsToolArgumentError(error) (*ToolArgumentError, bool)
func NewToolArgumentError(string, ToolArgumentErrorOptions) error
func NewInvalidJSONToolArgumentError(string, error) error
func NewInvalidToolArgumentError(string, []string, string) error
func NewMissingRequiredToolArgumentError(string, []string, string) error
```

- `NewToolArgumentError` trim tool/code/detail，并复制四个 slice。
- invalid/missing constructor trim、去空、去重 field，保留首次出现顺序。
- invalid JSON constructor 区分语法错误与 non-object top-level JSON。
- 特化 constructor 只生成 repair hint；`SafeAutorepair` 固定为 false。

## 非目标

- 不定义具体工具参数是否合法；
- 不执行 repair 或模型重试；
- 不拥有 tool body、provider/backend 或控制面策略；
- 不构成 Public、Beta、Stable 或 production-ready 声明。
