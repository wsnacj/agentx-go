# `tools/memory` API 参考

`memory` 提供 `memory_search` 与 `memory_get` 的可移植模型合同和调用协调。它负责参数解析、
source 别名归一化、上限收口、typed request、稳定 schema、typed error 与 `context` 传递；具体
文件、数据库、Session、召回模型和排名策略由 Host 实现。

成熟度：`Experimental extension / Developer Preview candidate`。当前不构成 Public、Beta 或
Stable 兼容性承诺。

## 接入

```go
registry := tools.NewRegistry()
memory.Register(registry, memory.Options{
    Backend: memory.BackendFuncs{
        SearchFunc: search,
        GetFunc: get,
    },
    MaxSearchResults: 8,
    MaxReadLines: 40,
})
```

`Backend` 是唯一副作用边界：

```go
type Backend interface {
    Search(context.Context, SearchRequest) (string, error)
    Get(context.Context, GetRequest) (string, error)
}
```

未提供 Backend 时不注册工具，不会扫描本机目录或选择默认数据库。

## `SearchRequest`

canonical handler 会：

- 校验 `query` 并收窄 `limit`；
- 把 `files/durable/MEMORY.md/memory/*` 归一为 `memory`；
- 把 `records/memory_records` 归一为 `structured`；
- 把 `session/recall` 归一为 `sessions`；
- 将 Session filter 放入 `SessionSearch`，但不解释 visibility、lineage、hydration 或 scorer。

`Backend.Search` 返回已有的 JSON 文本，canonical 不重排字段或改写 Host 的诊断/排名结果。

## `GetRequest`

`memory_get` 要求 `path`，并把 `lines` 限制在 Host 配置的上限内。`root` 不属于模型参数；
工作区和 memory root 必须由 Host 在 Backend construction 时固定。

## 错误、取消与并发

- 缺少 `query/path`、选择空 source 或尝试传入 `root` 返回
  `runtime/toolerrors.ToolArgumentError`；
- `context` 原样传给 Backend，取消与 deadline 不被吞掉；
- handler 不维护可变状态；Backend 是否支持并发以及读一致性由 Host 声明；
- credential、Session 可见性、具体 Store、embedding/reranker、产品召回策略均不属于本包。
