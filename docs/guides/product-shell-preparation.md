# ProductShell 两阶段准备

`extensions/productshell`把“用户/Host输入如何成为可执行的Case和Workflow上下文”收口为
一个 Experimental portable owner。它只负责准备，不拥有执行器，也不会把 HS产品策略
迁入通用 Runtime。

```go
import productshell "github.com/wsnacj/agentx-go/extensions/productshell"
```

## 安装与固定版本

当前 private-preview验证版本：

```bash
go get github.com/wsnacj/agentx-go/extensions@v0.0.0-20260801114445-cd5d97b84728
```

如果代码直接导入 `runtime/cases`，当前验证版本为：

```bash
go get github.com/wsnacj/agentx-go/runtime@v0.0.0-20260801112448-651dab4f0a53
```

这些 pseudo-version用于可重复验证，不是正式 semver。完整的无 HS、无 Runner、无
长期 `replace` 示例位于
[`extensions/conformance/productshell-consumer`](../../extensions/conformance/productshell-consumer)；
该 consumer的 `go.mod`固定实际版本，是本轮 external-style接入证据。

## 为什么拆成两阶段

ProductShell输入同时包含两类工作：

1. **可移植、无副作用的输入投影**：解码 option、规范化 Case/Shell Binding、解析显式
   skill/path和 Workflow绑定；
2. **依赖宿主事实的准备编排**：读取 session binding、选择 Pack、草拟 Case、从注册表
   物化 Workflow并执行产品校验。

第一类由 canonical helper直接实现；第二类由 `PreparationPipeline`固定顺序，再通过
`PreparationRuntime`把产品策略和 backend留给 Host。这样新项目可以复用稳定的准备
顺序，而不需要引入 HS、Runner、Scene或具体 provider。

## 第一阶段：投影输入

```go
shell, requestedBinding, sessionInput, workflowState, residual :=
    productshell.ProjectInputOptions(rawOptions)

input := productshell.Input{
    UserMessage:           "[skill:review] 检查这份改动",
    ProductShell:          "coding",
    RequestedShellBinding: requestedBinding,
    SessionInput:          sessionInput,
    WorkflowState:         workflowState,
    ShellOptions:          shell,
    Options:               residual,
}
```

typed field和显式输入优先；map/slice在绑定边界复制，避免调用方后续修改污染准备结果。
`ParseRequestedSkills`不会因为 `ProductShell == "coding"`而自动增加技能。任何业务推断、
推荐或授权均应在Host adapter中显式实现。

若调用方需要严格错误返回，应在投影前直接调用：

```go
binding, present, err := productshell.ParseRequestedShellBindingOption(rawOptions)
if err != nil {
    return err
}
_ = binding
_ = present
```

## 第二阶段：实现最窄 Host port

小型接入可以用 `PreparationRuntimeFuncs`，只填写自身真正拥有的能力：

```go
host := productshell.PreparationRuntimeFuncs{
    ResolveShellBindingFn: func(ctx context.Context, sessionID string, in productshell.Input) (*productshell.PreparedShellBinding, error) {
        // Host读取自身session backend；没有绑定时返回nil。
        return nil, nil
    },
    ResolveWorkflowFn: func(in productshell.Input) (productshell.ResolvedWorkflow, error) {
        // Host从自身Pack/Workflow registry解析并物化。
        return workflowResolver.ResolveWorkflow(in)
    },
    ResolveEffectiveCaseFn: func(sessionID string, in productshell.Input, message string, spec *workflow.Spec, binding *pack.Binding) (*cases.Case, error) {
        // Host决定Case identity、draft/policy和持久化边界。
        return cases.Clone(in.Case), nil
    },
    ValidateEffectiveCaseFn: func(binding *pack.Binding, value *cases.Case) error {
        // Host执行产品准入；返回原始typed/sentinel错误。
        return nil
    },
}

result, err := productshell.NewPreparationPipeline(host).PrepareWithInput(
    ctx,
    productshell.PrepareInput{SessionID: sessionID, Input: input},
)
if err != nil {
    return err
}
```

示例中的 `workflow`、`pack`和 `cases`分别来自 canonical Runtime/Extensions module。
Host port不得使用这些DTO来暗示backend也已迁入：registry、storage、authorization和
具体产品默认值仍属于调用方。

## 固定顺序与失败边界

pipeline按以下顺序运行：

```text
Case输入 → Shell Binding → Command Dispatch → Skills/Paths
→ Pack选择 → Case Binding → Workflow解析 → Effective Case校验 → Metrics
```

- 只有 Host的 `ShouldAttemptPackSelection`/`ShouldAttemptCaseBinding`返回true才进入相关
  分支；默认不会虚构选择或Case草拟能力；
- 每个需要I/O的hook接收原始 `context.Context`；Host应遵守取消和deadline；
- 首个错误立即返回且不执行后续阶段，`errors.Is/As` identity保持；
- `PrepareResult`只表示“准备完成”，不会启动 Open Tool Loop、Workflow或Objective；
- nil runtime只提供无副作用 pass-through，不等于完整ProductShell已配置。

## Workflow接入

若 Host同时支持显式 Workflow和Pack绑定Workflow，可复用
`WorkflowResolutionRuntime`：

```go
resolver := productshell.WorkflowResolutionRuntime{
    HasRegisteredPackFn: hasPack,
    ResolvePackWorkflowFn: resolvePackWorkflow,
    ResolveExplicitPackBindingFn: resolveBinding,
    MaterializeRegisteredWorkflowFn: materializeWorkflow,
}
```

显式raw Workflow必须 opt in；带Pack的显式 Workflow必须与Host注册的Pack Workflow
execution semantics一致。该检查只保护绑定边界，完整 structural validation和执行仍交给
`runtime/workflow`及 `runtime/workflow/hostkit`。

## HS迁移边界

HS production consumer应把 canonical package作为portable source authority，只保留：

- ProductShell选择、command/skill推断和产品默认值；
- Pack/Workflow registry及具体 materializer；
- Case草拟、LLM调用、timeout policy和validation policy；
- session binding持久化、RunStore及其它 backend adapter；
- authorization、approval、sandbox、provider、credential和观测产品投影。

兼容层可以做必要的HS类型转换，但不得复制 canonical stage order、option codec或
Shell Binding算法形成长期双写。

## 明确 non-goal

- 自然语言规划、LLM、provider或model/tool执行；
- 默认Pack/Workflow/Case产品策略；
- concrete Store、registry、durable backend或网络；
- Objective、Scene、CLI/HTTP和真实副作用；
- 完整hostless Runtime construction；
- Public、Beta、Stable或正式发行承诺。

具体签名和复制/并发语义见
[`extensions/productshell` API](../../extensions/productshell/API.md)和
[`runtime/cases` API](../../runtime/cases/API.md)。
