# runtime/cases API

导入路径：

```go
import cases "github.com/wsnacj/agentx-go/runtime/cases"
```

成熟度：**Experimental extension / private validation**。

本包定义 AgentX 的 Case 数据合同、规范化/复制 helper 和最小存储端口。它是
[`extensions/productshell`](../../extensions/productshell/API.md)准备阶段使用的
canonical Case source authority，但不提供数据库、文件或远程存储实现，也不拥有
Pack选择、Workflow执行或产品策略。

## Case

```go
type Case struct {
    ID            string         `json:"id,omitempty"`
    Type          string         `json:"type,omitempty"`
    PackID        string         `json:"pack_id,omitempty"`
    WorkflowID    string         `json:"workflow_id,omitempty"`
    Source        string         `json:"source,omitempty"`
    Intent        string         `json:"intent,omitempty"`
    SessionID     string         `json:"session_id,omitempty"`
    PolicyProfile string         `json:"policy_profile,omitempty"`
    MemorySchema  string         `json:"memory_schema,omitempty"`
    Status        string         `json:"status,omitempty"`
    Inputs        map[string]any `json:"inputs,omitempty"`
    Outcome       map[string]any `json:"outcome,omitempty"`
    CreatedAt     int64          `json:"created_at,omitempty"`
    UpdatedAt     int64          `json:"updated_at,omitempty"`
}
```

`Normalize` 会 trim 所有顶层 string 字段，保留时间戳和 opaque map 中的原始值，
并递归复制 `map[string]any` 与 `[]any`。nil/空 map 会规范化为非 nil 空 map。
它不解释 status、policy、input 或 outcome 的业务含义。

```go
func Normalize(Case) Case
func Clone(*Case) *Case
func IsZero(Case) bool
```

`Clone(nil)` 返回 nil；非 nil 输入先执行 Normalize，因此返回值拥有独立的
map/slice。`IsZero` 同样先 Normalize：空白 string、nil/空 map 都视为空，任意非零
时间戳或非空字段则不是零值。

## Filter 与 Store

```go
type Filter struct {
    PackID   string `json:"pack_id,omitempty"`
    CaseType string `json:"case_type,omitempty"`
    Status   string `json:"status,omitempty"`
    Limit    int    `json:"limit,omitempty"`
}

type Store interface {
    UpsertCase(context.Context, Case) error
    GetCase(context.Context, string) (Case, error)
    ListCases(context.Context, Filter) ([]Case, error)
}
```

`Store`只固定可替换 backend的最小方法集合；错误 identity、排序、并发、事务、
取消响应和 durable guarantee由具体实现声明。本包不提供默认backend。接口接收
`context.Context`，但是否以及何时响应取消或deadline取决于具体backend；调用方不得
假定任意实现都有相同I/O或关闭行为。

## 示例

```go
value := cases.Normalize(cases.Case{
    ID:         " case-1 ",
    Type:       " research ",
    PackID:     " pack-1 ",
    WorkflowID: " workflow-1 ",
    Inputs:     map[string]any{"query": "risk"},
})

if err := store.UpsertCase(ctx, value); err != nil {
    return err
}
```

## 非目标

- 不提供 StoreAdapter 或具体持久化 backend；
- 不定义 not-found、already-exists 等 backend 错误；
- 不选择 Pack、Workflow、policy 或 provider；
- 不依赖 HS、Runner、Scene、credential 或网络；
- 不构成 Public、Beta、Stable 或正式发行声明。

ProductShell接入示例见
[ProductShell两阶段准备指南](../../docs/guides/product-shell-preparation.md)。
