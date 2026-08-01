# `runtime/executionpolicy` 中文 API Reference

成熟度：**Experimental extension / private validation**。

本包定义执行身份和 policy 的 portable 数据合同。它只描述“允许什么、限制什么”，
不执行授权、审批、sandbox、provider调用或真实副作用。

## 主要类型

- `Contract`：一次执行使用的完整 policy快照输入。
- `Identity`：Profile、Pack、Workflow、Case、Run、Node和租户身份。
- `VisibilityPolicy`：工具 allow/deny、声明要求与风险上限。
- `BudgetPolicy`、`LoopPolicy`：工具调用、时限、token、成本、并发、重试和循环限制。
- `ApprovalPolicy`、`ReplayPolicy`、`RuntimeControlPolicy`：审批、replay和控制面声明。
- `SideEffectPolicy`、`SandboxPolicy`、`EvidencePolicy`：副作用、sandbox和证据要求。
- `Snapshot`、`Diff`、`CompileInput`：Host编译和持久化边界。
- `Compiler`：由 Host实现的 policy编译 port。

所有字段保持现有 JSON tag。空值是否生效、policy合并、授权决定和 backend行为由
Host拥有；本包不提供默认策略，也不暗示生产安全授权。

## 并发与错误

DTO本身没有后台生命周期。`Compiler`实现必须遵守调用方 `context.Context`；错误
identity由实现方拥有，本合同不会自动包装或改写。

## 非目标

- 具体 authorization/approval执行；
- sandbox或网络隔离 backend；
- provider、credential、Scene业务策略；
- Public/Beta/Stable或 semver承诺。
