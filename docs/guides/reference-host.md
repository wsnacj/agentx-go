# Reference Host：显式组合Core与可选能力

Reference Host回答一个常见问题：新项目使用AgentX时，是否必须重新实现大型产品宿主的
全部兼容代码？答案是否定的。新项目只需要实现自己的少量Host选择。

AgentX的责任分为三层：

1. 根合同和Runtime拥有Run、错误、取消、Shutdown、Tool Loop、Workflow、Objective、
   Session/Subagent/Resume等稳定执行语义；
2. providers、tools、browser、document和scenes提供可选能力；
3. Host选择provider、credential、授权、backend和产品策略，并把这些能力显式注入。

[`examples/reference-host`](../../examples/reference-host)是第三层的默认安全参考实现。
它不依赖专有Runner，也不会把产品策略下沉到Runtime。

## 快速运行

```bash
cd examples

GOWORK=off go run ./reference-host \
  -mode chat \
  -provider fixture \
  -tools none \
  -store memory \
  -input "hello AgentX"
```

Tool Loop与Tool Direct Answer：

```bash
GOWORK=off go run ./reference-host \
  -mode tool-loop \
  -provider fixture \
  -tools diffs \
  -store memory \
  -input "compare"
```

## 安全默认值

Reference Host默认满足：

- 不发现或读取环境credential；
- 不访问真实网络；
- 不启动进程；
- 不读取或写入文件；
- 不选择生产provider；
- 不安装隐式tool；
- 不假设durable backend。

`fixture` provider、`diffs` pure tool和`MemoryStore`都由配置显式选中。未知值或
不匹配组合直接返回错误，不会回退到环境或隐藏默认值。

## 如何替换

真实项目通常只替换三个窄位置：

| 位置 | Reference Host | 生产项目实现 |
| --- | --- | --- |
| Model provider | `fixtureProvider.Request` | 调用显式构造的provider client |
| Tool execution | `tools.Registry` + `diffs` | 注册经过授权、sandbox和结果校验的tool |
| Run data plane | `runstore.MemoryStore` | 实现`runstore.Store`的数据库或远程adapter |

真实credential、tenant、审批、网络allowlist、数据库schema、queue部署和审计继续属于
产品Host。AgentX不会替调用方决定这些策略，但调用方也不需要重写Tool Loop、结果收口、
Run生命周期或typed error。

## 与七类能力示例的关系

Reference Host只演示普通Chat和Open Tool Loop的产品组合，不用空实现伪装Workflow、
Objective或长任务已被统一Facade接管。其它能力使用对应的独立入口：

- [Workflow](../../examples/workflow)；
- [Objective](../../examples/objective)；
- [Session/Subagent/Resume](../../examples/session-subagent)；
- [Deterministic Scene](../../examples/deterministic-scene)。

这些入口共享AgentX的执行语义，但保留各自需要的Host port。进入Beta前可以继续评估是否
增加更高层Facade；当前Developer Preview不为减少API数量而隐藏真实治理输入。

## 当前边界

Reference Host是Developer Experience样板，不是production server或正式发行物。它不提供
默认provider、文件RunStore、系统scheduler、HTTP服务、credential manager或多租户产品壳。
这些能力只有出现清晰的通用owner、两个真实consumer和可验证的平台/取消/错误合同时，
才适合进入可选Batteries。
