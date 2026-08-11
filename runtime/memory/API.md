# `runtime/memory` API 参考

导入路径：

```go
import "github.com/wsnacj/agentx-go/runtime/memory"
```

成熟度：**Experimental**。本包不属于 `v0.2.2` Developer Preview
兼容候选面，也不构成 Public、Beta 或 Stable 承诺。

`memory` 提供 provider-neutral 的长期记忆 lifecycle 协调：显式 scope、provenance、recall
上限、幂等写入、revision CAS、归档、typed error 与 backend readback 校验。它不保存状态，
不选择数据库，不执行 embedding/rerank，也不决定用户可见性、retention 或自动写入策略。

## 与 `tools/memory` 的关系

- `runtime/memory` 是 Host 授权后的 durable lifecycle mechanism；
- `tools/memory` 是 `memory_search`/`memory_get` 的模型可见 Tool adapter；
- 模型参数不得直接成为 `ScopeRef` 或 durable write。Host 必须先从认证上下文映射 scope，并在
  control/approval policy 通过后调用本包。

两者不是两套 Store。Host 可以让 `tools/memory.Backend` 查询同一个 Platform backend，但 Tool
schema、授权和 JSON projection 继续由 Host 适配。

## Coordinator

```go
coordinator := memory.Coordinator{
    Policy: memory.Policy{
        MaxRecallLimit:    8,
        MaxContentBytes:   16 << 10,
        MaxReferenceCount: 32,
    },
    Backend: backend,
}
```

`Coordinator` 无跨调用状态，可并发使用；Backend 的并发与持久化能力由 Host 声明。三个操作
最多调用一次 Backend，canonical 不自动重试不确定的副作用。

### Recall

```go
result, err := coordinator.Recall(ctx, memory.RecallRequest{
    ScopeRef: "opaque:user-scope",
    Query:    "用户偏好的输出格式",
    Limit:    4,
})
```

空 `States` 只召回 `active`。`Limit=0` 使用 policy 上限，超限会被收窄。Backend 拥有候选选择
和排序；canonical 保留返回顺序，并拒绝跨 scope、超量、无 identity/revision/provenance 或非法
state 的 readback。

### Write

```go
written, err := coordinator.Write(ctx, memory.WriteRequest{
    ScopeRef:       scopeFromAuthenticatedHost,
    Content:        "用户希望示例优先使用 Go。",
    IdempotencyKey: requestID,
    Provenance: memory.Provenance{
        SourceKind: "session",
        SourceRef:  sessionID,
        SessionID:  sessionID,
        RunID:      runID,
    },
})
```

创建时 `RecordID=""`、`ExpectedRevision=0`；更新时两者必须同时提供。Backend 必须原子检查
revision 和 idempotency，并返回 authoritative `Record`。canonical 要求内容、scope、identity、
revision 与 provenance readback 一致，避免成功响应掩盖静默改写。

### Archive

`Archive` 要求 record ID、非零 expected revision 和 idempotency key，合法转换只有
`active -> archived`。它不物理删除内容。删除、retention、自动 supersede、合并、衰减和用户
画像是后续 Host policy，不在首版合同中。

## Backend

```go
type Backend interface {
    Recall(context.Context, RecallRequest) (RecallResult, error)
    Write(context.Context, WriteRequest) (WriteResult, error)
    Archive(context.Context, ArchiveRequest) (ArchiveResult, error)
}
```

Backend 负责 concrete store、visibility、ranking、原子持久化、revision 和 idempotency。CAS
失败返回 `memory.ErrConflict`；Coordinator 会映射为稳定 `ErrorCodeConflict`。

## 错误、取消和敏感信息

`Error` 提供稳定 code 和 display-safe message，底层原因只通过 `errors.Unwrap` 暴露。稳定 code
包括 `invalid_policy`、`invalid_request`、`canceled`、`deadline_exceeded`、
`backend_unavailable`、`recall_failed`、`write_failed`、`archive_failed`、`conflict` 和
`invalid_backend_result`。

错误展示文本与 report 不包含 content、query、scope、provenance、credential 或 backend cause。
caller cancellation/deadline 在调用前后检查；Backend 必须传播 `ctx`。

## Non-goal

本包不拥有 SQLite/vector、filesystem root、embedding、reranker、跨租户 visibility、credential、
approval、自动记忆提取、后台刷新、备份、迁移、Tool schema 或 Session/RunStore 执行事实。上述
能力属于 `agentx-platform` 或业务 Host。
