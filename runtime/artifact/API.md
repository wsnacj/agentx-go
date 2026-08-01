# runtime/artifact API（Experimental）

导入路径：

```go
import artifact "github.com/wsnacj/agentx-go/runtime/artifact"
```

该 package 定义 AgentX Runtime 的可移植 Artifact registry、lineage link 与 BlobStore
合同，并提供并发安全的进程内 `MemoryRegistry`。当前成熟度为 **Experimental /
private validation**，不构成 Public、Beta、Stable 或 production-ready 声明。

## Registry 数据合同

`Record`描述一个 Artifact 的身份、Run/Node/Session关联、producer/source、kind/role、
路径或存储引用、摘要、标签和时间：

```go
type Record struct {
    ArtifactID   string   `json:"artifact_id"`
    RunID        string   `json:"run_id,omitempty"`
    NodeExecID   string   `json:"node_exec_id,omitempty"`
    SessionID    string   `json:"session_id,omitempty"`
    ToolName     string   `json:"tool_name,omitempty"`
    Producer     string   `json:"producer,omitempty"`
    Source       string   `json:"source,omitempty"`
    Kind         string   `json:"kind,omitempty"`
    Role         string   `json:"role,omitempty"`
    Path         string   `json:"path,omitempty"`
    StorageRef   string   `json:"storage_ref,omitempty"`
    URL          string   `json:"url,omitempty"`
    Digest       string   `json:"digest,omitempty"`
    MIMEType     string   `json:"mime_type,omitempty"`
    Format       string   `json:"format,omitempty"`
    Bytes        int64    `json:"bytes,omitempty"`
    Summary      string   `json:"summary,omitempty"`
    Labels       []string `json:"labels,omitempty"`
    MetadataJSON string   `json:"metadata_json,omitempty"`
    CreatedAt    int64    `json:"created_at"`
}
```

`ArtifactID`和 `CreatedAt`即使为零值也会出现在 JSON 中；其它字段按 `omitempty`
处理。`MetadataJSON`只是调用方提供的字符串，本 package 不解析或验证其 JSON 内容。

`Link`用 `SourceArtifactID`、`TargetArtifactID`和 `Relation`表达 lineage；
`LinkFilter`可按 Artifact、relation和 direction过滤。direction为 `inbound`时匹配
target，为 `outbound`时匹配source；其它值按双向匹配处理。

## Registry 接口

```go
type AuthoringRegistry interface {
    Register(context.Context, Record) error
    Link(context.Context, Link) error
}

type QueryRegistry interface {
    ListByRun(context.Context, string) ([]Record, error)
    ListBySession(context.Context, string) ([]Record, error)
}

type Registry interface {
    AuthoringRegistry
    QueryRegistry
    ListLinks(context.Context, LinkFilter) ([]Link, error)
}
```

这些接口不规定 durable backend、事务、跨进程一致性或授权策略。Host可以提供数据库、
对象存储或其它实现。

## `MemoryRegistry`

`NewMemoryRegistry`返回进程内实现。它保持既有兼容语义：

- 写入前 trim全部字符串；Labels trim、去空、去重并保留首次出现顺序；
- 同一 `ArtifactID`重复注册时，字符串字段保留首个非空值，`Bytes`取较大值，
  Labels合并，`CreatedAt`取最早的正值；
- Run/Session索引只在 Artifact首次注册时建立，后续merge不会迁移或补建索引；
- Run/Session查询按 `CreatedAt`、`ArtifactID`升序；
- 相同 source/target/relation的 Link会合并，Metadata保留首个非空值，CreatedAt取最早
  正值；Link查询按 CreatedAt、source、target、relation升序；
- 并发方法调用由互斥锁保护。返回的 Record、Link及其中slice应按只读值使用；
- nil `*MemoryRegistry`的写方法无操作且返回nil，查询返回nil slice和nil error；
- 当前内存方法不检查 context cancellation。调用方不得据此推导durable backend行为。

`MemoryRegistry`不持久化，进程退出后数据丢失。应通过constructor创建；零值结构体不作为
支持的初始化方式。

## BlobStore 仅为合同

```go
type BlobStore interface {
    Put(context.Context, BlobPutInput) (BlobRef, error)
}
```

`BlobPutInput.Data`使用 `json:"-"`，不会被默认JSON编码；`BlobRef`只返回Host定义的
storage ref、digest和字节数。本 package不提供默认 BlobStore，也不决定路径、权限、
digest算法、加密、去重、生命周期或网络行为。

## 明确不包含

- `FileBlobStore`、私有文件创建、原子写入、symlink防护或文件权限策略；
- `artifact/runtime`中的工具持久化、Workflow scope或Runner集成；
- RunStore/backend、provider、credential、网络或真实生产副作用；
- Artifact保留策略、访问控制、内容安全扫描和业务级类型校验。
