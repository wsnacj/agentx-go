# ProductShell Preparation 与临时 Workflow 规划固定版本 Consumer

这个隔离 consumer 只依赖已经推送的 `agentx-go/extensions` 与
`agentx-go/runtime` 固定 pseudo-version，不使用 HS、Runner、长期 `replace`、
真实 provider、网络、凭据或生产副作用。

示例由宿主显式提供确定性回调，验证以下两条路径。

Preparation路径：

1. `Input` 合并请求的 Case；
2. 显式 skill directive 解析为 requested skills；
3. portable Pack 选择与 binding；
4. Case binding 与有效 Case 校验；
5. Workflow materialization；
6. `PreparationPipeline` 的固定阶段顺序和最终结果。

临时 Workflow规划路径：

1. Host从raw visible tools显式选择并转换可规划工具；
2. 无网络的确定性 `TemporaryWorkflowPlanGenerator`生成typed计划；
3. canonical planner构造prompt/schema、执行binding lowering并生成Workflow Spec；
4. `TemporaryWorkflowPlanningPipeline`固定`Should → Resolve → Apply`顺序；
5. metrics、Workflow ID、raw Workflow opt-in及输入/输出binding可被外部module读取。

它只证明外部 module 能消费 Experimental ProductShell preparation及temporary workflow
planning合同。generator、validator、tool policy和identity均由Host显式注入；示例不代表
默认 provider、真实LLM、持久化、完整ProductShell Runtime或正式发行承诺。

```bash
GOWORK=off GOPROXY=off go test ./...
GOWORK=off GOPROXY=off go run .
```

预期输出：

```text
agentx-productshell-ok:portable-research:research.lookup:collect-v1:portable-review:case-001:AgentX
agentx-productshell-planning-ok:temp_workflow_external_consumer:1:lookup:true
```
