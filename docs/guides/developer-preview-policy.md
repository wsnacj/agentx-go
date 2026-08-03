# Developer Preview兼容与分发政策

本政策只适用于AgentX Go private Developer Preview，不是Public、Beta、Stable或生产SLA。

## 候选合同

兼容审阅面只包含成熟度矩阵中的14个Developer Preview candidate，以及以下五条标准
construction；七类能力的概念映射和扩展入口见[能力矩阵](capability-map.md)：

1. 根`agentx.Client`与自定义`ExecutionAdapter`；
2. `runtime/hostkit`与调用方提供的Model/Tool Adapter；
3. `runtime/workflow/hostkit`与调用方提供的Workflow ports；
4. `runtime/objective/hostkit`与调用方提供的Objective policy/catalog/adapter handler；
5. `runtime/session/hostkit`与调用方提供的child worker、state store和verification合同。

其它Experimental package仍可能调整；Go exported不等于产品公共API。

## 允许和禁止的变更

Developer Preview允许在Owner审阅后进行additive或breaking变更，但禁止静默改变：

- `RunRequest`、`RunResult`、identity和status映射；
- typed error code、`errors.Is/As`、retryability和display-safe message；
- context cancellation/deadline、并发Run和`Shutdown(ctx)`的有界/幂等/关闭后语义；
- LLM/tool JSON、Workflow state transition、durable write顺序；
- 推荐package路径、九module依赖方向或公开类型闭包。

任何候选API变化必须同时完成Owner/consumer review、中文Reference、可读API snapshot与
hash、CHANGELOG、升级/回滚说明、fixed consumer和双平台gate。breaking change还必须
使用新的不可变版本，并给出旧版本回滚点。

安全修复可以缩短迁移窗口，但仍需记录行为影响；不得以“安全”为名隐藏公共合同变化。

## Deprecation

当前不承诺固定日历窗口。非紧急删除应先标记Deprecated、提供替代路径，并至少跨越下一
个已接受Developer Preview baseline；安全或数据完整性问题可以更快移除，但必须记录
原因和迁移步骤。进入Beta前必须重新批准明确时间窗口和EOL政策。

## 九Module release train与tag前缀

以下是九module的候选tag设计，不是当前发行授权。当前九module仍使用各自固定
pseudo-version；首次Beta建议与完整准入边界见
[Pre-Beta准入合同](../reference/pre-beta-admission.md)：

| Module | Go tag前缀 |
| --- | --- |
| `github.com/wsnacj/agentx-go` | `vX.Y.Z...` |
| `github.com/wsnacj/agentx-go/components` | `components/vX.Y.Z...` |
| `github.com/wsnacj/agentx-go/runtime` | `runtime/vX.Y.Z...` |
| `github.com/wsnacj/agentx-go/extensions` | `extensions/vX.Y.Z...` |
| `github.com/wsnacj/agentx-go/providers` | `providers/vX.Y.Z...` |
| `github.com/wsnacj/agentx-go/tools` | `tools/vX.Y.Z...` |
| `github.com/wsnacj/agentx-go/browser` | `browser/vX.Y.Z...` |
| `github.com/wsnacj/agentx-go/document` | `document/vX.Y.Z...` |
| `github.com/wsnacj/agentx-go/scenes` | `scenes/vX.Y.Z...` |

如果未来批准同一release train，应明确哪些module同版、哪些独立版，并避免调用方组合
未经验证的checkpoint。当前pseudo-version只用于private validation，不预先批准首次
tag或版本号。

## Version epoch必须分离

- distribution semver：未来Go module release train；
- Git历史tag：旧仓库历史，不自动映射为新产品版本；
- manifest`since: v1`：HS内部legacy contract epoch；
- HTTP `/v1/...`：传输协议版本；
- serialized schema中的`*_v1`：数据格式版本。

任何一个出现`v1`都不表示AgentX Go已经Stable v1。映射必须由未来release ADR显式批准。

## 工具链与平台

当前module的`go 1.24.1`是语言/module graph基线，不构成Go 1.24 patch安全支持承诺。
darwin/arm64与linux/amd64的CGO-disabled API/build surface已经对账；M6C历史证据覆盖
macOS arm64/Go 1.25.5和Ubuntu 24.04.4/amd64/Go 1.25.5/CGO=1。M6D扫描证明Go 1.25.5
标准库存在AgentX Runtime可达漏洞，因此当前Pre-Beta技术候选工具链已收紧为Go 1.25.12；
正式最低/最新支持策略仍需Owner批准。任一结果都不自动形成更宽Linux、架构、Go版本或
native支持承诺。

## Readiness边界

本地preflight和fresh-cache VCS consumer通过后可以声明
`private_validation_ready=true`。Public Beta还要求Owner选择并批准license/NOTICE、
具名security和release approver、正式tag/release授权、正式兼容等级以及剩余Pre-Beta
门禁。M6C已关闭Ubuntu实机证据项，但缺其它任一项时`public_beta_ready=false`。
M6D历史证据只覆盖当时四module snapshot；当前本地与Ubuntu入口已经升级为九module，
必须从当前source revision重新运行后才形成新的技术证据。即使技术门禁通过，也不代替
Owner对License、具名security/release责任和正式兼容等级的批准。
