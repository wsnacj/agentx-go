# runtime/workflow/schema API

导入路径：

```go
import schema "github.com/wsnacj/agentx-go/runtime/workflow/schema"
```

成熟度：**Experimental / private validation**。

该 package 是 portable Workflow JSON Schema normalization 和 definition
validation 的 canonical implementation owner。它接受普通 Go/JSON 值，只依赖
标准库，不依赖 HS、Runner、Scene、provider 或 backend。

## Normalize

```go
func Normalize(raw any, label string) (map[string]any, error)
```

`raw` 可以是 `map[string]any` 或没有外围空白的 JSON object string。空 map
和空 string 返回 `nil, nil`；其它类型、whitespace-only、外围空白或无效 JSON
按稳定 `workflow: <label> ...` 文本报错。

函数不会复制传入的 map，也不验证 schema definition。host 若需要 definition
validation，应显式调用 `ValidateDefinition`。

## ValidateDefinition

```go
func ValidateDefinition(definition map[string]any, path string) error
```

`ValidateDefinition` 递归验证 AgentX Workflow 当前支持的 JSON Schema 子集：

- `type`：object、array、string、number、integer、boolean、null；
- object：properties、required、additionalProperties、min/maxProperties；
- array：items、min/maxItems；
- string：pattern、min/maxLength；
- number/integer：minimum、maximum、exclusiveMinimum、exclusiveMaximum；
- literal：const、enum。

它验证 keyword/type 适用性、canonical lowercase、重复项、literal type、
嵌套 definition 和 range，但不执行 Workflow runtime value validation。
`path` 原样参与错误路径，调用方应传稳定的配置路径，如 `config.schema`。

## Owner 边界

该 package 不拥有：

- Workflow config key、alias、default 或 validation 调用顺序；
- Spec/Node kind policy、pack/product policy；
- provider、credential、Scene 或具体 executor/backend；
- 完整 Agent Runtime construction。

这些能力由 host 或其它 canonical owner 负责。

## 并发与生命周期

两个函数都不创建 goroutine，不保存全局或请求状态。调用方不并发修改传入
map 时，可以并发调用；package 不提供 Shutdown。

## 稳定性

当前只属于 Experimental/private validation，不构成 Public、Beta、Stable 或
production-ready 声明。进入 Beta 前仍需决定它是保留独立扩展 API，还是由
更高层 Workflow facade 统一暴露。
