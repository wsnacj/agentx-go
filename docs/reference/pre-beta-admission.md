# Pre-Beta准入合同

本页定义AgentX九module从private Developer Preview进入Pre-Beta候选前必须满足的统一
合同。它是技术和责任准入清单，不是发行授权；任一具名决定缺失时，
`public_beta_ready`必须保持`false`。

## Release train

九个library module组成一个受控release train：root、components、runtime、extensions、
providers、tools、browser、document和scenes。技术候选必须在同一source revision上使用
同一个可删除prerelease版本组装，并同时通过module zip、test、race、vet、tidy、list、
固定漏洞扫描、无`replace`聚合consumer和只读离线cache验证。

当前只批准临时file proxy中的`v0.0.0-m6d.0`用于技术测量。它不会写入tracked
`go.mod`、tag或公共下载入口。

2026-08-03已在macOS、Go 1.25.12和干净source revision
`f57ecda758b5d43614bac2e5f9b005e24dff8228`上完成一次九module候选复验：132个package的
test/vet/tidy/list、九module zip、无HS且无`replace`的聚合consumer、`GOPROXY=off`只读
cache和固定`govulncheck@v1.6.0`均通过，已知可达漏洞为0。该结论只解除本地技术门禁，
不解除下表的责任与发行决定。

## 正式tag方案

建议的首次Beta版本为`v0.1.0-beta.1`，但尚未授权创建。Go多module仓库使用相同语义
版本、不同tag前缀：

| Module | 候选tag |
| --- | --- |
| root | `v0.1.0-beta.1` |
| components | `components/v0.1.0-beta.1` |
| runtime | `runtime/v0.1.0-beta.1` |
| extensions | `extensions/v0.1.0-beta.1` |
| providers | `providers/v0.1.0-beta.1` |
| tools | `tools/v0.1.0-beta.1` |
| browser | `browser/v0.1.0-beta.1` |
| document | `document/v0.1.0-beta.1` |
| scenes | `scenes/v0.1.0-beta.1` |

tag必须全部指向同一个批准source revision，并保持immutable。失败版本不得重写tag；应在
后续版本中使用`retract`或发布修复版本。HTTP/schema中的`v1`、历史Git tag和旧manifest
`since`不属于这个distribution version epoch。

## Fail-closed决定

| 决定 | 当前状态 | 解除条件 |
| --- | --- | --- |
| License/NOTICE | pending | Owner选择许可模式，必要法律审阅后提交正式文本，并确认九个module zip均携带正确文本 |
| Security approval | pending | 指定具名security approver，审阅九module扫描、响应流程和残余依赖风险 |
| Compatibility promotion | Developer Preview only | 明确批准14个候选package的Beta兼容范围和允许变更规则 |
| Release authorization | pending | release owner与独立reviewer批准版本、tag、回滚、支持平台和发布窗口 |
| Ubuntu九module证据 | passed | Actions run `30792532517`在Ubuntu 24.04、Go 1.25.12、CGO=1与source `2ae4fbd671fa...`上通过同一门禁 |

固定扫描仍报告少量“module已依赖、但生产调用图不可达”的上游记录：extensions和document
中的`golang.org/x/sys`，以及scenes中的`golang.org/x/net`。直接升级extensions到当前修复
版本会同时把该module的最低Go版本从1.24提升到1.25，因此本轮没有把工具链合同变化伪装成
安全修复。具名security approver应在Beta前选择兼容升级、最低Go版本提升或带平台边界的
显式风险接受。

许可证选择涉及产品与法律责任，本仓库不会自动写入MIT、Apache-2.0、专有许可或双许可
文本，也不会因为`LICENSE`文件出现就自动解除法律和发行审批。

## Inventory与旧治理快照

进入本轮Pre-Beta技术准入不重新生成HS的全仓inventory、C4/C5或15文件旧归档。旧快照
基于早期source scope，已经不能代表当前九module release train。只有License、具名安全
责任、Beta surface和release owner全部确定，并冻结最终候选commit后，才评估是否需要一次
面向发布范围的最小重建；不得恢复54至81分钟的全仓日常治理路径。

HS中`AGENTX_RELEASE_GOVERNANCE_CHECK=1`仍可显式显示旧快照stale/mismatch，它是
deferred release blocker，不属于agentx-go九module技术候选的运行时回归。

## 本地入口

```bash
GOWORK=off go run ./scripts/check_developer_preview_distribution.go
GOWORK=off go run ./scripts/check_pre_beta_candidate.go
```

第一条验证当前九module固定Developer Preview版本；第二条从当前tracked source构建同版
九module临时候选。两条命令均不得创建tag或发布版本。Ubuntu复验使用
`GOWORK=off CGO_ENABLED=1 go run ./scripts/check_ubuntu_runtime.go`。

远端run `30792532517`的job `91618919616`耗时13分11秒，候选artifact
`8847878284`的归档digest为
`sha256:0438657392a3d3779e0fa4d37371961284ae39212294b2c637c8c477755f70c4`。
它同时通过Ubuntu runtime gate和九module候选门禁；`public_beta_ready`仍因License、
具名安全、兼容范围和发行授权保持`false`。
