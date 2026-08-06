# ProductShell 两阶段准备

`extensions/productshell`把“用户/Host输入如何成为可执行的Case和Workflow上下文”收口为
一个Experimental portable package。除两阶段preparation外，它还提供一个可选的临时
Workflow planning mechanism；具体模型、工具策略、执行器和产品策略仍不进入通用
Runtime。

```go
import productshell "github.com/wsnacj/agentx-go/extensions/productshell"
```

## 安装与固定版本

安装当前版本：

```bash
go get github.com/wsnacj/agentx-go/extensions@v0.2.1
```

如果代码直接导入 `runtime/cases`：

```bash
go get github.com/wsnacj/agentx-go/runtime@v0.2.1
```

可重复构建的项目应固定`go.mod`和`go.sum`。完整的无专有Runner、无长期
`replace` 示例位于
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
顺序，而不需要引入专有Runner、具体Scene或provider。

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

## 准备后可选：临时 Workflow 规划

没有显式Workflow且没有Pack binding时，Host可以选择进入temporary planning。启用判断
必须显式：typed `AutoWorkflowPlanning`或受支持option优先，产品默认值由Host传入
`ShouldAttemptTemporaryWorkflowPlanning`，canonical package不会根据自然语言或
ProductShell名称自行开启。

接入分成三步：

1. Host从raw visible tools应用自身alias、denylist、authorization和产品策略，转换为
   `[]TemporaryWorkflowPlanningTool`，并单独决定是否允许LLM step；
2. `TemporaryWorkflowPlanner`通过Host提供的generator生成typed plan，完成有限重试、
   binding lowering、Workflow Spec构造和validator调用；
3. `TemporaryWorkflowPlanningPipeline`固定`Should → Resolve → Apply`顺序，成功后由Host
   继续编译execution snapshot、过滤capability并执行Workflow。

```go
planner := productshell.NewTemporaryWorkflowPlanner(
    productshell.TemporaryWorkflowPlannerConfig{
        Generator:         generator, // Host model adapter；可用确定性fixture测试
        Validator:         validator,
        WorkflowIDFactory: workflowIDFactory,
        NormalizeToolName: normalizeToolName,
    },
)

prepared, err := planner.ResolveTemporaryWorkflowPlan(
    ctx,
    productshell.TemporaryWorkflowPlannerInput{
        Input:            preparedInput,
        UserMessage:      userMessage,
        PlanningTools:    hostFilteredTools,
        AllowLLMSteps:    hostAllowsLLMSteps,
        VisibleToolCount: len(rawVisibleTools),
        LLMTaskTimeoutMs: timeoutMs,
    },
)
```

代码片段只展示portable planner。完整stage adapter和可运行fixture见
[`extensions/conformance/productshell-consumer`](../../extensions/conformance/productshell-consumer)。
该consumer固定不可变module版本，不使用专有Runner、长期`replace`、真实provider、
credential或网络。

### 失败和生命周期

- generation request timeout默认45秒，由Host generator负责执行；只对generator返回的
  deadline重试一次，retry最多90秒，取消不重试；
- 当前request存在非空session input清单时，引用清单外的`session.input.*`会得到一次
  带可用路径的planning feedback；
- generator、binding、validator失败统一成为`TemporaryWorkflowPlanningError`，cause可用
  `errors.Is/As`读取；
- planner不启动goroutine也不持有backend；generator、validator和identity由Host管理；
- `PreparedTemporaryWorkflowPlan`和metrics表示plan已经生成/校验，不表示Workflow已经
  执行或持久化。

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

## 既有 Host 接入边界

既有Host应直接复用portable package，只在自身代码中保留：

- ProductShell选择、command/skill推断和产品默认值；
- 临时规划的tool alias/denylist、可见性和启用策略；
- 具体model/provider adapter、credential和LLM输出解码；
- Pack/Workflow registry及具体 materializer；
- Case草拟、产品timeout policy和validation policy；
- session binding持久化、RunStore及其它 backend adapter；
- authorization、approval、sandbox、provider、credential和观测产品投影。

兼容层可以做必要的Host类型转换，但不得复制 canonical stage order、option codec或
Shell Binding算法形成长期双写。

## 明确 non-goal

- 默认自然语言模型、具体LLM/provider调用、credential或model/tool执行；
- tool policy、默认planning启用策略、execution snapshot/capability filter；
- 默认Pack/Workflow/Case产品策略；
- concrete Store、registry、durable backend或网络；
- Objective、Scene、CLI/HTTP和真实副作用；
- 完整hostless Runtime construction；
- Public、Beta、Stable或正式发行承诺。

具体签名和复制/并发语义见
[`extensions/productshell` API](../../extensions/productshell/API.md)和
[`runtime/cases` API](../../runtime/cases/API.md)。

准备完成后的typed session/process观测和display-safe Host交接不属于preparation stage；
需要该路径时继续阅读
[ProductShell Observation与Host Handoff](product-shell-observation-handoff.md)。
