# 使用 Workflow Host Kit

这是 M3E 推荐给显式图执行调用方的标准接入路径。它适用于已经拥有业务 validation
policy、node mapping、executor和可选存储 backend，但不希望了解或重复组合
lowering、journal、node execution coordination、orchestration和 composition的 Host。

## 1. 安装与导入

当前 private validation固定版本：

```bash
go get github.com/wsnacj/agentx-go/runtime@v0.0.0-20260802113655-f41de95ec5be
```

调用方只需两个 AgentX import：

```go
import (
    workflow "github.com/wsnacj/agentx-go/runtime/workflow"
    workflowhostkit "github.com/wsnacj/agentx-go/runtime/workflow/hostkit"
)
```

不需要直接导入 composition、lowering、journal、nodeexec或 orchestration。

## 2. 最小可运行示例

```go
package main

import (
    "context"
    "fmt"

    workflow "github.com/wsnacj/agentx-go/runtime/workflow"
    workflowhostkit "github.com/wsnacj/agentx-go/runtime/workflow/hostkit"
)

type validator struct{}

func (validator) ValidateSpec(workflow.Spec) error     { return nil }
func (validator) ValidateNode(workflow.NodeSpec) error { return nil }

type mapper struct{}

func (mapper) MapNode(
    workflow.NodeSpec,
    workflow.ExecutionMode,
) (workflowhostkit.MappedCall, error) {
    return workflowhostkit.MappedCall{Name: "echo"}, nil
}

type executor struct{}

func (executor) Execute(
    context.Context,
    workflowhostkit.Call,
) (string, error) {
    return "hello workflow", nil
}

func main() {
    runtime, err := workflowhostkit.New(workflowhostkit.Config{
        Validator:          validator{},
        Mapper:             mapper{},
        BasicExecutor:      executor{},
        NewRunID:           func() string { return "run-example" },
        NewNodeExecutionID: func() string { return "nodeexec-example" },
        NowUnixMilli:       func() int64 { return 1 },
    })
    if err != nil {
        panic(err)
    }

    result, err := runtime.Run(context.Background(), workflow.Spec{
        ID:        "hello",
        EntryNode: "echo",
        Nodes: []workflow.NodeSpec{{
            ID:   "echo",
            Kind: workflow.NodeTool,
        }},
    }, workflowhostkit.Inputs{})
    if err != nil {
        panic(err)
    }
    fmt.Println(result.Execution.FinalStatus)
    fmt.Println(result.Execution.NodeOutput["echo"])
}
```

输出：

```text
completed
hello workflow
```

仓库内同一代码路径的 fixed-version证据位于
[`runtime/conformance/workflow-hostkit-consumer`](../../runtime/conformance/workflow-hostkit-consumer)。
该 nested module没有 `replace`，也不依赖 HS或 Runner。

## 3. Host 必须拥有的能力

| 能力 | 是否必需 | Owner |
| --- | --- | --- |
| `Validator` | 是 | Host产品 admission policy |
| `Mapper` | 是 | Host tool/model/task mapping |
| Basic/Node/Outcome executor之一 | 是 | Host concrete execution |
| `NewRunID` | 是 | Host identity policy |
| `NewNodeExecutionID` | 是 | Host identity policy |
| `NowUnixMilli` | 是 | Host clock |
| `JournalPort` | 否 | Host durable backend adapter |
| `NewEventID` | 配置 `JournalPort` 时必需 | Host identity policy |
| context binder/error projection | 否 | Host observation与兼容 policy |

executor能力优先级为 Outcome、Node、Basic，每个节点只调用一个。Host Kit不根据
node kind猜测 provider或工具，也不会为缺失能力安装默认值。

## 4. 输入与结果

`workflowhostkit.Inputs`包含可选 runtime Workflow ID以及 `RunInputs`：Run、Case、
Branch identity和 initial/session/case binding roots。Workflow ID exact-empty时使用
`Spec.ID`，其它值原样保留。

`Result`同时保留：

- lowering plan，供 Host做诊断或兼容投影；
- partial或完整 execution，包括 RunID、final status、node results、node output和
  state。

lowering失败返回零结果；durable写入或 node execution开始后失败会保留已经产生的
partial result。所有 dependency error和 cancellation/deadline都保留
`errors.Is/As` identity。

## 5. Durable 与生命周期

不配置 `JournalPort`时，Workflow仍执行同一个 canonical journal/composition路径，
但不会写具体 backend。配置 port后，Run、state snapshot、node upsert和 lifecycle
event顺序由 canonical journal保证；Host只负责五方法 backend adapter。

Workflow Host Kit不拥有 concrete backend资源，因此没有 `Shutdown`。如果 executor
或 storage需要关闭，Host必须在自己的生命周期 owner中有界、幂等地关闭。不要添加
一个永远返回 nil的伪 Shutdown来暗示资源已经被管理。

## 6. 取消与并发

调用方 context原样传入 journal和 executor，不会被替换，也不会在取消后由 Host Kit
后台继续。`Runtime`自身不可变、不持有单次 Run state、不创建 goroutine；并发安全
取决于调用方注入的 validator、mapper、executor、port、identity和clock。

如果这些依赖不是并发安全的，Host应为每次请求创建独立 Runtime或在依赖 owner处
同步；Host Kit不会偷偷加锁改变吞吐和顺序语义。

## 7. 与 Open Tool Loop 的选择

- 输入是自然语言、模型动态选择工具并进行有界多轮执行：使用根 Client +
  `runtime/hostkit.NewModelToolClient`。
- 输入是显式 `workflow.Spec`，节点和边在执行前已声明：使用本 Workflow Host Kit。
- 已有完整 Runtime且只需要根 Run合同：实现自定义 `ExecutionAdapter`。

Workflow Host Kit不会把显式图包装为根 Client的假 `mode`。两条执行路径可以由同一
业务 Host分别组合，但状态、重试、资源和产品 policy仍由其真正 owner管理。

## 8. 非目标

- 默认 validation/mapping、provider、credential或网络 client；
- concrete RunStore、queue、scheduler或 filesystem backend；
- retry、resume、Objective、长任务或子 Session编排；
- 根 `agentx.Client` Workflow mode；
- Scene、HTTP、CLI或生产副作用；
- Public、Beta、Stable或正式发布承诺。
