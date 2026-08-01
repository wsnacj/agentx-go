# Run 与 Artifact 数据平面

M5F把执行记录和产物索引的 portable source authority分别放入：

```go
import (
    artifact "github.com/wsnacj/agentx-go/runtime/artifact"
    runstore "github.com/wsnacj/agentx-go/runtime/runstore"
)
```

两个 package当前均为 **Experimental / private validation**。它们提供数据合同、窄
port和可用于测试或单进程接入的内存实现，不是完整 durable Runtime，也不构成
Public、Beta、Stable或 production-ready声明。

## 数据关系

一次执行的数据关系是：

```text
Run
 ├─ NodeExecution：节点尝试、状态及 execution projection
 ├─ Event：Run或Node关联的有序生命周期/观察事件
 └─ Artifact Record：通过RunID/NodeExecID关联产物元数据
      └─ Link：表达Artifact之间的lineage关系
```

`runstore.Store`拥有 Run、NodeExecution和Event的保存/读取合同；
`artifact.Registry`拥有 Artifact Record注册、按Run/Session查询和lineage查询合同。
两者通过字符串 identity显式关联，不隐藏事务，也不会在写入 RunStore时自动创建
Artifact。Host必须决定何时写入、失败时如何补偿，以及跨 backend的一致性边界。

## 最小内存接入

```go
package main

import (
    "context"
    "fmt"

    artifact "github.com/wsnacj/agentx-go/runtime/artifact"
    runstore "github.com/wsnacj/agentx-go/runtime/runstore"
)

func main() {
    ctx := context.Background()
    runs := runstore.NewMemoryStore()
    artifacts := artifact.NewMemoryRegistry()

    if err := runs.CreateRun(ctx, runstore.Run{
        RunID: "run-1",
        Status: "running",
    }); err != nil {
        panic(err)
    }
    if err := runs.UpsertNodeExecution(ctx, runstore.NodeExecution{
        NodeExecID: "nodeexec-1",
        RunID:      "run-1",
        NodeID:     "collect",
        Status:     "completed",
    }); err != nil {
        panic(err)
    }
    if err := runs.AppendEvent(ctx, runstore.Event{
        EventID:    "event-1",
        RunID:      "run-1",
        NodeExecID: "nodeexec-1",
        Name:       "node.completed",
        CreatedAt:  1,
    }); err != nil {
        panic(err)
    }
    if err := artifacts.Register(ctx, artifact.Record{
        ArtifactID: "artifact-1",
        RunID:      "run-1",
        NodeExecID: "nodeexec-1",
        Kind:       "report",
        StorageRef: "host://reports/1",
        CreatedAt:  1,
    }); err != nil {
        panic(err)
    }

    records, err := artifacts.ListByRun(ctx, "run-1")
    if err != nil {
        panic(err)
    }
    fmt.Println(records[0].ArtifactID)
}
```

`MemoryStore`要求先创建 Run，才允许写入该 Run的Event或NodeExecution；
`UpdateRun`只更新已存在的 Run。Event按 `CreatedAt`、`EventID`稳定排序，NodeExecution
按 `StartedAt`、`NodeExecID`稳定排序。`MemoryRegistry`按 Artifact ID归并重复注册，
并可按 Run、Session或Link筛选查询。

## 并发与 context

`MemoryStore`和`MemoryRegistry`使用读写锁保护内部映射，支持并发调用其公开方法。
调用方仍应把查询返回值当作只读快照，不应在多个 goroutine之间并发修改同一个DTO或
其 slice字段。

当前内存实现不创建goroutine、不持有外部资源，也没有 `Shutdown`。方法接收
`context.Context`是为了与Host backend port保持同形，但纯内存操作当前不会主动轮询
`ctx.Done()`；需要取消、deadline、I/O中断或有界关闭的durable实现必须由Host提供并
遵守传入context。

## Error 合同

RunStore提供两个可通过 `errors.Is`判断的sentinel：

```go
errors.Is(err, runstore.ErrNotFound)
errors.Is(err, runstore.ErrAlreadyExists)
```

- 重复创建 Run或重复 Event ID返回 `ErrAlreadyExists`；
- 更新不存在的 Run，或向不存在的 Run写Event/NodeExecution，返回 `ErrNotFound`；
- 缺失必需 identity返回普通validation error，不应按上述sentinel分类。

Artifact内存Registry当前没有稳定error code：无效Link会被忽略，重复Record或Link按
既定字段规则归并。调用方不得把这一宽松内存行为误当成所有durable backend必须采用的
产品validation策略；严格校验应由Host admission层拥有。

## JSON 边界

Run、NodeExecution、Event、Record和Link使用稳定的显式JSON tag；可选字段使用
`omitempty`。以下字段是“包含JSON文本的字符串”，而不是自动校验的
`json.RawMessage`：

- `Event.PayloadJSON`；
- `NodeExecution.ExecutionContractDiffJSON`、`TerminationJSON`和
  `DelegatedExecutionJSON`；
- `Artifact.Record.MetadataJSON`与`Artifact.Link.MetadataJSON`。

写入前的schema、大小、敏感信息和display-safe检查仍由Host负责。
`artifact.BlobPutInput.Data`标记为 `json:"-"`，不会因序列化DTO而意外进入JSON。

## `artifact` 与 `mediaartifact` 的区别

- `runtime/artifact`描述可注册、查询、关联lineage的执行产物identity和metadata，
  并定义 `BlobStore` port；
- `runtime/mediaartifact`只描述截图、PDF页、视频等媒体输出的尺寸、格式、时长等
  wire metadata；它没有registry、lineage或持久化语义。

两者不会自动互转。Host可以把媒体Descriptor投影成Artifact Record，但必须显式决定
Artifact ID、Run/Node关联、storage reference、digest和安全策略。

## 明确 non-goal

- 数据库、对象存储、文件BlobStore或private file writer；
- 跨进程durability、事务、锁、迁移、索引、保留期、压缩或清理；
- Run、Event、NodeExecution与Artifact之间的原子提交；
- queue、scheduler、resume、replay policy或Objective生命周期；
- provider、credential、Scene、HTTP/CLI和真实网络副作用；
- 默认状态机、业务事件名称、Artifact分类或产品validation策略；
- Public、Beta、Stable、正式tag或兼容性承诺。

具体签名见 [`runtime/runstore` API](../../runtime/runstore/API.md)和
[`runtime/artifact` API](../../runtime/artifact/API.md)。
