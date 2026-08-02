# `runtime/objective` 中文 API Reference

状态：**Experimental / private validation**。本状态不构成 Public、Beta或Stable承诺。

本 package是 Objective推荐路径的最小类型入口。当前它使用Go type alias保持既有
`runtime/controlcontract` kernel的类型、JSON和方法identity，使Host不需要继续把大型
Experimental package作为业务入口；后续物理owner拆分可以在该边界后继续进行。

## Managed ingress

```go
func BuildManagedIngress(ManagedObjectiveIngressInput) ManagedObjectiveIngressResult
```

输入只接受Host已经整理好的goal digest、success criteria、evidence、policy、budget、
approval、strategy catalog、adapter registry和display-safe refs。它不会读取原始goal，
不会调用模型、工具、Workflow或adapter，也不会执行持久化和副作用。

结果明确给出Objective loop、strategy planning/final gate和runtime adapter是否ready，以及
status、failure class、missing inputs、boundaries和next host action。

## Host adapter与result

```go
func BuildHostAdapterRegistry(HostAdapterRegistryInput) HostAdapterRegistrySnapshot
func BuildRuntimeAdapterResult(RuntimeAdapterResultInput) RuntimeAdapterResult
```

registry只描述Host拥有的adapter，不包含函数、凭据或backend。Host执行完成后必须报告
adapter/run/strategy identity、structured observations、evidence和output refs；builder不会
代替Host执行adapter。

## Observation与verification

```go
func BuildObservationNormalization(ObservationNormalizationInput) ObservationNormalizationResult
func BuildVerification(ObjectiveVerificationInput) ObjectiveVerificationResult
```

normalization把Host的structured result转换为portable observation合同，并保留adapter、
strategy、run和evidence identity。verification按Objective required evidence逐项判断
satisfied、partial或blocked。两者都拒绝raw output进入display-safe合同。

## 类型入口

本package只重命名Objective Host Kit直接使用的最小类型：activation、control mode、
execution intensity、policy、budget、strategy catalog、adapter descriptor、managed ingress、
runtime-adapter request/result、observation、evidence和verification。

这些名称不是第二份数据模型，不做字段复制或JSON转换。Beta前仍需继续把底层物理owner
从`controlcontract`收拢；当前调用方只应使用本文和`objective/hostkit`展示的子集。

## 非目标

- 不解析自然语言或选择产品策略；
- 不提供provider、Runner、tool、Workflow、scheduler或store backend；
- 不安装credential、authorization、approval或side-effect默认；
- 不表示整个`controlcontract`已经成为公共API。
