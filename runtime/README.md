# AgentX Go Runtime

本目录是 `agentx-go` 的 Runtime owner module：

```text
github.com/wsnacj/agentx-go/runtime
```

当前落地的无外部依赖叶子 package：

- [`assetfs`](./assetfs/API.md)：immutable filesystem snapshot、content
  fingerprint、atomic resolver registration与 `assetfs://`解析。
- [`artifact`](./artifact/API.md)：Run/Session/Node关联的 Artifact identity、
  metadata、lineage与 registry/blob port，并提供并发安全的 `MemoryRegistry`；
  concrete文件、对象存储和清理策略继续由 Host拥有。
- [`protocol`](./protocol/API.md)：版本化 Runtime wire/schema、normalization
  与 validation。
- [`telemetry/safeerror`](./telemetry/safeerror/API.md)：observation-safe
  error projection、identity 与 cause-preserving wrapper。
- [`mediaartifact`](./mediaartifact/API.md)：跨 browser、PDF、video、nodes
  capability 共享的媒体产物元数据 wire descriptor。它不负责 Artifact注册、
  lineage或持久化；需要保存和查询执行产物时使用 `artifact`，两者不会自动互转。
- [`objective`](./objective/API.md)：Objective推荐路径的最小类型与构造名称；当前保持
  `controlcontract` kernel的类型/JSON identity，为后续物理owner拆分建立稳定边界。
- [`objective/hostkit`](./objective/hostkit/API.md)：组合managed ingress、显式Host
  runtime-adapter dispatch、observation normalization和verification；具体handler、policy、
  approval、credential与backend继续由Host拥有。
- [`toolerrors`](./toolerrors/API.md)：结构化工具参数错误、cause chain 与
  deterministic repair hint 数据合同。
- [`budget`](./budget/API.md)：对调用方提供的 limit/snapshot 执行无副作用的
  预算阶段、停止原因与近限额警告判定。
- [`promptcontext`](./promptcontext/API.md)：构造 prompt rendering 所需的时间、
  timezone、session/model identity，并提供 fail-soft RFC3339 时间投影。
- [`runstore`](./runstore/API.md)：Run、NodeExecution和Event的数据合同、存储 port、
  node execution投影，以及并发安全的 `MemoryStore`；durable backend、事务、保留期、
  跨进程一致性和恢复策略继续由 Host拥有。
- [`channel`](./channel/API.md)：portable message/target与sender/runner ports、routing、
  chunking、in-memory dedupe、bounded ingress queue/worker/cancellation/Shutdown和
  display-safe session delivery合同；pairing、access policy、平台SDK、credential、
  durable backend与真实网络发送继续由Host拥有。
- [`construction`](./construction/API.md)：通过窄 `Host` port组合 model、
  runner、adapter和根 Client，拥有阶段顺序、context检查、失败清理与 ownership
  transfer；具体 provider、Runner、adapter与产品策略继续由 host注入。
- [`controlcontract`](./controlcontract/API.md)：公共执行控制状态、evidence、Objective、
  Host adapter catalog/request/readback、Managed Objective Ingress与closeout projection，
  durable-store request/result/readback、trace-backed final answer、async delegation controller/
  completion/runtime handoff与strategy metadata，
  blocker、next action与display-safe合同，以及Objective Spec/strict JSON、capability/
  strategy catalog、intensity/controller、graph validation、evidence verification、
  recovery/replanning、Host effect request-result-readback及admission/invocation kernel；
  具体adapter、delegation normalization、LLM/provider、worker dispatch、执行、调度backend、
  持久化与产品策略留在Host。
- [`execution`](./execution/API.md)：把根 `agentx.Client` 的 adapter request
  确定性分派给窄 `Host` port，组装 adapter result并转发 Shutdown与 error
  classification；具体 engine input/output、model/tool/backend继续由 host拥有。
- [`executionpolicy`](./executionpolicy/API.md)：执行身份、可见性、预算、循环、
  approval、replay、sandbox和 evidence policy DTO及 Host编译 port，并提供Snapshot
  metadata、bounded retry command、decision packet、observation-only execution-loop
  report与soft-rejection确定性reducer；不执行授权、审批、工具、Host adapter或backend。
- [`hostkit`](./hostkit/API.md)：拥有 portable model/tool round adapter，并组合
  per-run `toolloop.Assembly`、`execution.Runtime`和根 Client；普通新项目通过
  `NewModelToolClient`提供 model/tool函数，无需手写 Factory、RoundExecutor或
  依赖 HS Runner，高级 Host仍可使用窄 `Factory`注入完整 assembly policy。
- [`hosthttp/hostserver`](./hosthttp/hostserver/API.md)：Host-deployed Scene的
  bounded transport、request identity、exposure policy与 graceful shutdown。
- [`hosthttp/requestjson`](./hosthttp/requestjson/API.md)：有界、拒绝未知字段和
  trailing data的严格 JSON请求解码。
- [`hosthttp/resourcepolicy`](./hosthttp/resourcepolicy/API.md)：Host-owned路径、
  opaque value、permission与预算收窄机制。
- [`toolloop`](./toolloop/API.md)：通过窄 `Stepper`/`RoundExecutor`/
  `RoundPhaseExecutor` ports驱动确定性多轮执行，拥有 outcome收口、
  continuation state更新、request→observe→before-action→act与
  failure fuse→host policy→loop detector固定顺序、循环/成功重放检测，并通过
  `Assembly`组合 driver、coordinator、termination capture与 final portable
  state；model request、concrete tool executor、持久化与用户可见回复继续由
  host拥有。
- [`telemetry`](./telemetry/API.md)：Runtime event、tool/semantic projection、
  stored-event replay、summary 与私有 JSONL sink。
- [`workflow`](./workflow/API.md)：Workflow Spec 的 planning/node/execution
  mode 与 nodes/edges/state/artifact/evaluator 数据合同，以及最小 validator
  construction seam；不含 validation 实现或 executor。
- [`workflow/bindingstate`](./workflow/bindingstate/API.md)：portable binding
  materialization、node-result recording、内存 state transition 与 required
  slot validation implementation；不含 lowering、executor 或 durable
  lifecycle。
- [`workflow/transition`](./workflow/transition/API.md)：portable traversal、
  final-status normalization 与 success/failure/always edge-routing
  implementation；不含 node execution 或 durable lifecycle。
- [`workflow/journal`](./workflow/journal/API.md)：portable run/node
  snapshot、upsert 与 lifecycle event 的 fail-fast durable ordering，通过
  五方法 Port 保留 host-owned backend。
- [`workflow/nodeexec`](./workflow/nodeexec/API.md)：portable node context
  binding、Outcome/Node/basic capability priority 与 exact-once invocation；
  concrete executor 和 policy 留在 host。
- [`workflow/orchestration`](./workflow/orchestration/API.md)：组合 portable
  binding、transition、journal 与 node execution 的 lowered-plan run 主循环；
  lowering、具体 executor/backend 和 error display policy 继续由 host 注入。
- [`workflow/schema`](./workflow/schema/API.md)：portable Workflow JSON Schema
  normalization 与递归 definition validation implementation；config
  key/alias/default 和 admission policy 留在 host。
- [`workflow/validation`](./workflow/validation/API.md)：portable Spec/Node
  structural orchestration、graph/binding kernel 和显式 host policy port；
  不提供默认产品或 runtime policy。
- [`workflow/lowering`](./workflow/lowering/API.md)：portable validation 后
  node lowering、argument JSON 编码和 orchestration plan projection；
  tool/model/task mapping 与 default继续由 host注入。
- [`workflow/composition`](./workflow/composition/API.md)：构造并保存已验证的
  lowering/orchestration依赖，按固定顺序执行 lower→run，并同时返回 portable
  plan与 partial/full result；具体 policy/backend继续由 host注入。
- [`workflow/hostkit`](./workflow/hostkit/API.md)：面向普通 Host的标准 Workflow
  construction/access seam，组合现有 lowering、journal、nodeexec、
  orchestration和 composition真实 owner，不复制执行语义。

当前成熟度为 **Core Developer Preview candidate / private validation**。本
module已提供需要调用方注入 `construction.Host`的高级构造生命周期、
`execution.Runtime`，以及不依赖 HS Runner的 `hostkit.NewModelToolClient`。Host
Kit仍需调用方提供 concrete model/tool函数；本 module尚未提供无需任何
host-provided adapter/policy的根 `agentxruntime.New`、真实 backend、provider、
credential、Scene领域实现或
完整 embedded Runtime，不能据此宣称 Runtime 已达到 Public、Beta、Stable 或
production-ready。

依赖方向固定为：

```text
Runtime owner package -> Go 标准库或已批准的 agentx-go contract/component
```

production package 不得 import `hs/...`、`scene/...` 或机器本地路径。新增
package 必须有独立 owner/consumer review、中文 API 文档、focused/race/module
验证和 source-authority cutover，不创建杂项 `common` package。

本地验证：

```bash
GOWORK=off go test ./...
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go mod tidy
GOWORK=off go list -m all
```

`conformance/` 下的 external-style consumer 是独立 nested module，必须单独
验证；根 module 的 `go test ./...` 不会自动跨越 nested module。M5J/M5K 的
`controlcontract-consumer`固定 pseudo-version、无长期 `replace`，覆盖 projection、
budget、lifecycle、display-safe fail-closed、Objective Graph validation、verification/
recovery proposal、Host effect gate以及Objective runtime/executor/productization路径。
