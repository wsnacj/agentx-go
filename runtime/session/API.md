# `runtime/session` API

`runtime/session`是Task、Session与Subagent公共命名边界。当前成熟度为
**Experimental extension / private validation**。

它复用已经验证的delegation、evidence、parent merge、async completion与Objective loop
类型身份，使推荐入口`runtime/session/hostkit`不直接泄漏巨大的
`runtime/controlcontract`底层包。该包自身不执行worker、不创建session、不排队、不调度、
不持久化，也不拥有credential或生产副作用。

## 主要合同

- `DisplaySafeRef`、`AttemptRef`、`Boundary`、`EvidenceRef`、`Observation`；
- `HostOwnedDelegationWorkerRuntimeReadiness`与
  `HostOwnedDelegationWorkerRuntimeInvocation`；
- `DelegationWorkerParentMerge`与`DelegationObjectiveRuntimeHandoff`；
- async child status/backend/role、completion projection与Objective loop step；
- 对应的normalize、append、merge和builder函数。

这些alias/forwarder保持原JSON和类型身份，不代表全部底层control contract自动成为
Developer Preview API。普通项目应优先依赖`runtime/session/hostkit`；只有实现自定义Host
port或检查typed readback时才直接导入本包。

## Non-goal

- Task queue、scheduler、retry、resume和wake；
- child prompt、model/tool选择和authorization；
- Runner、process、provider、credential或backend；
- 把child输出直接接受为parent事实。

## 验证

```bash
cd runtime
GOWORK=off go test ./session/... 
GOWORK=off go test -race ./session/...
GOWORK=off go vet ./session/...
```
