# `extensions/pack` 中文 API Reference

成熟度：**Experimental extension / private validation**。

本包提供 Domain Pack 的 portable 定义、确定性校验编排、内存注册、Workflow选择与
semantic tool物化、自然语言路由选择和运行前 Binding。它不执行模型、工具、审批、
memory/eval backend或真实副作用。

## 构造入口

```go
coordinator, err := pack.NewCoordinator(workflowValidator, toolArgumentLowerer)
registry, err := pack.NewMemoryRegistry(coordinator)
err = registry.Register(definition)
binding, ok, err := coordinator.ResolveBinding(registry, packID, caseType, workflowID)
```

- `NewCoordinator(Validator, ToolArgumentLowerer)`：显式注入 Host拥有的 Workflow
  admission与 tool config lowering。任一依赖为空都会 fail closed。
- `Coordinator.ValidateDefinition`：按稳定顺序校验 Pack字段、引用、schema、
  Workflow和 semantic tool/evaluator/policy/memory合同。
- `Coordinator.MaterializeWorkflow`：选择 Workflow，把 semantic tool映射为 runtime
  tool并验证物化后的输入和 Workflow；不会调用真实工具。
- `NewMemoryRegistry`：构造并发安全的内存注册表；注册时深拷贝 definition并拒绝
  重复 ID。
- `Coordinator.ResolveBinding`：组合 Definition、Workflow、CaseSchema、
  PolicyProfile、EvalSuite和 MemorySchema。

## 定义合同

- `Manifest`、`Definition`：Pack identity、路由提示、支持的 case、默认 Workflow及
  全部内容清单。
- `CaseSchema`、`CaseLibraryCase`、`CaseInputPlaceholder`：case输入和样例合同。
- `PromptTemplate`、`PromptTemplateVariable`、`SourceAttribution`、
  `PackMediaArtifact`：prompt、来源归属和媒体资产描述。
- `SemanticTool`：语义工具名、输入/输出 schema、Host runtime tool映射与默认参数。
- `Evaluator`、`EvalSuite`：评估输出与 gate/shadow suite描述，不执行评估。
- `PolicyProfile`：引用 `runtime/executionpolicy.Contract`，不执行授权。
- `MemorySchema`、`MemoryRecallPolicy`：memory记录与 recall要求，不拥有存储 backend。

`Definition`提供按 ID/name/case type查询、默认项选择与 Workflow选择 helper。这些
helper不会修改原定义，也不会规范化带空白的已存 identity。

## Registry、路由和 Binding

- `Registry`：注册、读取、排序列举及 Workflow解析的最小 port。
- `MemoryRegistry`：并发安全实现；`Get`/`List`返回深拷贝。
- `SelectBinding`：根据 Pack、case、Workflow、route hint和描述进行确定性评分；
  低于阈值或最高分歧义时不选择。
- `SelectOptions`、`RouteSelection`、`RouteSelectionCandidate`：路由输入与证据。
- `Binding`：运行前完整绑定，可验证 case input和 memory record、查询 policy合同。

## Runtime metadata helper

`SemanticToolNameFromConfig`、`ArtifactTypesFromConfig`、
`SemanticToolTagsFromConfig`、`StripSemanticToolRuntimeMetadata`及两个 Normalize
函数只处理 Pack保留 metadata，不执行工具或改变 Host业务字段。

## 并发、错误和复制语义

- `Coordinator`构造后自身只读；调用方注入的 Validator/Lowerer必须满足相同并发
  边界。
- `MemoryRegistry`支持并发读写，注册保持 validate→duplicate check→copy顺序。
- Host validator/lowerer错误 identity通过 `%w`保留；Pack上下文、错误文本和错误
  顺序属于当前 compatibility differential。
- map/slice/Workflow/Definition在 registry与 Binding边界执行深拷贝，避免调用方
  修改注册内容。

## 非目标

- `pack/runtime` memory capture/recall、eval suite执行和持久化；
- concrete model/tool executor、authorization、approval、sandbox backend；
- Scene pack内容、provider、credential或网络；
- Public/Beta/Stable、正式 tag或 semver承诺。
