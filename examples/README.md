# 示例与可运行消费证据

`examples` 是 AgentX 的中文教学入口。示例固定到下方的 Developer Preview 版本，
在 `GOWORK=off`、无本地 `replace`、无 HS、无 Runner 条件下构建，
因此既能阅读，也能作为真实外部项目的最小接入样板。

当前固定版本：

```text
v0.2.0
```

如果仓库尚未公开，首次运行前需要配置 `GOPRIVATE`、`GONOSUMDB`、`GOPROXY=direct`
以及可读取仓库的Git凭据；凭据不能写入本目录、源码或`go.mod`。

## 七条核心能力与两条重型扩展路径

| 示例 | 覆盖能力 | Host显式提供 | 默认副作用 |
| --- | --- | --- | --- |
| [`chat`](chat) | 模型对话 / A0默认控制状态 | fixture model函数 | 无 |
| [`tool-loop`](tool-loop) | Open Tool Loop + Tool Direct Answer | fixture model、`diffs`注册与结果确认 | 无 |
| [`workflow`](workflow) | 显式Workflow图 | validator、mapper、executor、identity、clock | 无 |
| [`objective`](objective) | Objective Runtime Loop | strategy、policy、approval与fixture handler | 无 |
| [`session-subagent`](session-subagent) | Session、Subagent、Scheduler、跨construction Resume | worker、readback、Host-owned内存state/queue | 无 |
| [`deterministic-scene`](deterministic-scene) | Deterministic Scene | fixture行情handler与确定性evaluator | 无网络 |
| [`reference-host`](reference-host) | 可配置的Chat或Tool Loop参考Host | provider、tools、RunStore选择 | 默认全离线 |
| [`browser`](browser) | Browser Tool最小接入 | 显式memory backend | 无网络 |
| [`document`](document) | PDF Tool最小接入 | 显式memory PDF/model/input ports | 无文件/网络 |

逐个运行：

```bash
GOWORK=off go run ./chat
GOWORK=off go run ./tool-loop
GOWORK=off go run ./workflow
GOWORK=off go run ./objective
GOWORK=off go run ./session-subagent
GOWORK=off go run ./deterministic-scene
GOWORK=off go run ./browser
GOWORK=off go run ./document
```

所有示例使用fixture、pure tool或in-memory port，不读取credential，不访问真实网络，
不启动外部进程，不写文件，也不表示生产backend已经由AgentX默认提供。

`examples`不会加入需要credential或真实网络的“隐藏模式”。真实provider验收位于
[`conformance/live/provider-smoke`](../conformance/live/provider-smoke)，它是独立、显式
opt-in的external-style consumer，不是教学示例，也不会被默认测试隐式执行。

## Reference Host

`reference-host` 演示“小而稳定的Core + 可选Batteries + 显式Host”的最终组合方式。
默认配置只运行离线fixture模型和内存RunStore：

```bash
GOWORK=off go run ./reference-host \
  -mode chat -provider fixture -tools none -store memory -input "hello"

GOWORK=off go run ./reference-host \
  -mode tool-loop -provider fixture -tools diffs -store memory -input "compare"
```

当前只接受下列显式组合：

- `provider=fixture`；
- `tools=none|diffs`，其中Chat必须为`none`，Tool Loop必须为`diffs`；
- `store=memory`。

未知provider、隐式环境credential、文件store或不匹配的tool配置会直接失败。替换为真实
provider、授权tool或durable backend时，调用方修改Host组合，不修改AgentX Runtime语义。
完整说明见[Reference Host接入指南](../docs/guides/reference-host.md)。

## 根合同与高级扩展

| 示例 | 说明 |
| --- | --- |
| [`contract-basic`](contract-basic) | 根`Client`、`Run`、identity和有界`Shutdown` |
| [`custom-adapter`](custom-adapter) | 高级自定义`ExecutionAdapter`、typed error与`errors.Is/As` |

`custom-adapter`是已有完整Runtime的Host使用的高级construction seam，不是普通新项目
使用AgentX的默认入口。

## 示例与conformance的职责

示例解释“如何接入”，各module下的`conformance/*-consumer`证明固定版本的合同、
错误、并发、取消、跨module依赖和无反向依赖。示例通过不替代module gate，conformance
通过也不自动构成Public、Beta、Stable或生产发行声明。
