# 安装与多 Module 引用

`agentx-go`当前采用一个仓库、九个独立 Go library module。调用方只引入实际使用的
module：

| Module | 用途 |
| --- | --- |
| `github.com/wsnacj/agentx-go` | 根 Client、Run、错误和 ExecutionAdapter合同 |
| `github.com/wsnacj/agentx-go/components` | provider-neutral LLM合同 |
| `github.com/wsnacj/agentx-go/runtime` | Host Kit、toolloop、Workflow和其它 portable Runtime owner |
| `github.com/wsnacj/agentx-go/extensions` | 可选 portable extension机制；当前含Domain Module、Pack、Skills与ProductShell |
| `github.com/wsnacj/agentx-go/providers` | 可选模型provider；当前含OpenAI-compatible、Anthropic Messages、Codex Responses client与transport/fault/retry/usage机制 |
| `github.com/wsnacj/agentx-go/tools` | 可选通用tool catalog与实现；当前含线程安全Registry、保守名称修复和纯文本diffs |
| `github.com/wsnacj/agentx-go/browser` | 可选Browser runtime、browserd host与推荐Browser tools |
| `github.com/wsnacj/agentx-go/document` | 可选OCR、Document pipeline、PDF与推荐Document tools |
| `github.com/wsnacj/agentx-go/scenes` | 可移植Domain Kits；当前覆盖A股、研究/公开来源、文档、浏览器操作、公共交通、港/美股与财报 |

自定义 ExecutionAdapter 路径只需要根 module。Host Kit + Model/Tool Adapter
路径需要根、components和 runtime三个 module，因为配置显式使用根合同、LLM响应
和 Runtime类型。Workflow Host Kit路径只需要 Runtime module。A股推荐入口、
Host Kit、Pack和内嵌资产由 scenes module提供，并通过其固定依赖使用
runtime/components；调用方只需直接声明自己代码实际 import的 module。只从目录
加载 Skill时直接声明 extensions即可；若代码直接构造经 `runtime/assetfs`证明身份的
immutable `FSSource`，还应把 runtime列为直接依赖。
需要AgentX提供的模型协议实现时直接声明providers；credential和endpoint
仍必须由Host显式提供，Core Runtime不会反向依赖providers。
需要canonical通用tool时直接声明tools；仅使用provider-neutral tool合同的调用方声明
components即可。Runtime不反向依赖tools，Host仍显式拥有授权、安全策略和具体backend。

## 当前固定验证版本

下列矩阵的机器可读事实源为
[`developer-preview-module-versions.txt`](../reference/developer-preview-module-versions.txt)；
文档门禁会核对九个版本、代表性external consumer与本节，避免只验证历史四module。

```text
github.com/wsnacj/agentx-go
  v0.0.0-20260802113655-f41de95ec5be
github.com/wsnacj/agentx-go/components
  v0.0.0-20260802130858-34ec103e09d9
github.com/wsnacj/agentx-go/runtime
  v0.0.0-20260802113655-f41de95ec5be
github.com/wsnacj/agentx-go/extensions
  v0.0.0-20260802113655-f41de95ec5be
github.com/wsnacj/agentx-go/providers
  v0.0.0-20260802124746-c7f90139a1cc
github.com/wsnacj/agentx-go/tools
  v0.0.0-20260802165151-c51d7391dbb4
github.com/wsnacj/agentx-go/browser
  v0.0.0-20260802183055-f15e2f99ed1a
github.com/wsnacj/agentx-go/document
  v0.0.0-20260802203835-ec74047cbe60
github.com/wsnacj/agentx-go/scenes
  v0.0.0-20260803022834-4043fbe78ff3
```

这些是不可变 private validation pseudo-version，不是 tag或正式 semver。

```bash
go get github.com/wsnacj/agentx-go@v0.0.0-20260802113655-f41de95ec5be
go get github.com/wsnacj/agentx-go/components@v0.0.0-20260802130858-34ec103e09d9
go get github.com/wsnacj/agentx-go/runtime@v0.0.0-20260802113655-f41de95ec5be
go get github.com/wsnacj/agentx-go/extensions@v0.0.0-20260802113655-f41de95ec5be
go get github.com/wsnacj/agentx-go/providers@v0.0.0-20260802124746-c7f90139a1cc
go get github.com/wsnacj/agentx-go/tools@v0.0.0-20260802165151-c51d7391dbb4
go get github.com/wsnacj/agentx-go/browser@v0.0.0-20260802183055-f15e2f99ed1a
go get github.com/wsnacj/agentx-go/document@v0.0.0-20260802203835-ec74047cbe60
go get github.com/wsnacj/agentx-go/scenes@v0.0.0-20260803022834-4043fbe78ff3
```

## Private 仓库访问

```bash
export GOPRIVATE=github.com/wsnacj/agentx-go
export GONOSUMDB=github.com/wsnacj/agentx-go
export GOPROXY=direct
export GOWORK=off
```

Git必须能使用组织批准的 HTTPS token或 SSH访问私有仓库。凭据、token和机器级
URL rewrite不得写入源码、`go.mod`、示例或日志。

## 多 Module 验证边界

根目录的 `go test ./...`不会跨越 nested module。至少分别运行：

```bash
GOWORK=off go test ./...
GOWORK=off go -C components test ./...
GOWORK=off go -C runtime test ./...
GOWORK=off go -C extensions test ./...
GOWORK=off go -C providers test ./...
GOWORK=off go -C tools test ./...
GOWORK=off go -C browser test ./...
GOWORK=off go -C document test ./...
GOWORK=off go -C scenes test ./...
```

external-style consumer也是独立 nested module，应在 `GOWORK=off`下单独测试。
根[`conformance/consumer`](../../conformance/consumer)同时固定根、components、runtime、
extensions和scenes五module版本，在一个无HS、无Runner、无长期`replace`、无网络副作用的
consumer中验证自定义ExecutionAdapter、Model/Tool Host Kit、Workflow Host Kit三条
标准construction路径以及A股推荐extension入口。
`extensions/conformance/domain-module-consumer`只固定 extensions，验证新项目可以
在无 HS、无 Runner、无长期 `replace` 时实现 config resolver、注册 callback和
无模型Domain Kit fixture执行。
`scenes/conformance/astock-consumer`固定 scenes/extensions/runtime，验证 A股 Manifest、
嵌入资产、三组 Pack、route/binding、fixture Host Kit和 evaluator的组合路径。
`extensions/conformance/skills-consumer`固定 extensions/runtime，验证 immutable
Skill加载、缓存、activation、requested semantics和资源引用；它不依赖 HS、Runner、
长期 `replace`、网络或命令执行。
长期 consumer不得依赖本地 `replace`；本地 `replace`只能用于临时开发测量，不能
作为 fixed-version或发布证据。
`providers/conformance/openaicompat-consumer`固定components/providers，不使用
`replace`、HS、Runner、Scene或真实网络，验证显式HTTPDoer、credential injection、
chat/tool-call解码和usage合同。
`providers/conformance/provider-cohort-consumer`固定P2-B版本，并在同一无网络fixture中
验证Anthropic Messages与Codex Responses/SSE；Codex token store和OAuth刷新不属于
canonical consumer。
`tools/conformance/catalog-diffs-consumer`固定components/runtime/tools版本，不使用
`replace`、HS、Runner、Scene、网络或credential；它验证catalog注册、保守名称修复和
纯文本diffs真实执行。该consumer不证明Host授权、sandbox或具体backend已经迁入。

`tools/conformance/general-tools-consumer`固定P2-D tools版本和其直接依赖，在独立module中
组合invocation registry、diffs以及message/filesystem/HTTP/memory/scheduler五类工具。它
通过内存workspace和显式fake ports执行全部10个注册入口，不使用`replace`、HS、Runner、
Scene、真实网络、文件、credential或scheduler backend；生产Host仍必须显式提供并治理
这些能力。

`tools/conformance/llm-task-consumer`固定P6-A2 tools版本，不使用`replace`，通过显式fake
model adapter验证`llm_task`的JSON/schema/model identity与取消合同。它不发现provider、
credential或全局模型配置；真实模型接入仍由Host显式注入。

Browser Runtime使用独立可选module：

```bash
go get github.com/wsnacj/agentx-go/browser@v0.0.0-20260802183055-f15e2f99ed1a
```

P3已完成`browser/runtime`、`browser/host/browserd`、`browser/tools`与统一fixed consumer。
module不会自动启动browserd，也不提供默认credential、proxy、登录态或网络；这些继续由Host
显式注入和授权。

Document与Portable Scenes分别使用独立固定版本。对应external-style consumer位于
`document/conformance/*-consumer`与`scenes/conformance/*-consumer`；它们覆盖推荐高层
入口，不把低层OCR/Pipeline实现包自动升级为Developer Preview API。完整能力与入口见
[七类能力矩阵](capability-map.md)，package分级见
[API索引与成熟度矩阵](../reference/package-maturity.md)。

## Ubuntu实机复跑

Pre-Beta候选lane固定Ubuntu 24.04 amd64与Go 1.25.12，在真实Linux进程中运行九module
normal/race/vet/tidy/list、双平台API gate、fixed-version空缓存consumer和module
artifact provenance：

```bash
GOWORK=off CGO_ENABLED=1 go run ./scripts/check_ubuntu_runtime.go
```

该命令需要Linux amd64、Go 1.25.12、GitHub私有仓库读取权限和公共Go依赖访问；它会使用临时空
`GOMODCACHE`获取固定版本，并在冻结cache、`GOPROXY=off`后再次运行consumer。GitHub
Actions的完整入口为`.github/workflows/m6d-pre-beta-candidate.yml`，支持当前迁移分支
push和`workflow_dispatch`，不要求PR；M6C workflow保留为手动单项复跑。M6D还会运行：

```bash
GOWORK=off go run ./scripts/check_pre_beta_candidate.go
```

该命令只在临时目录使用`v0.0.0-m6d.0`组装九module同版候选，不修改tracked `go.mod`，
并运行固定漏洞扫描、无replace consumer和只读cache复验。不得把token、URL rewrite或
runner credential写入仓库。技术门禁覆盖九module仍不自动解除License/NOTICE、具名
security/release责任、兼容性晋级或正式发行授权。

## 升级方式

Developer Preview期间，升级应显式修改 pseudo-version，并运行调用方 focused
tests、API differential和 module cache验证。完整步骤、回滚边界和已知非承诺见
[版本、升级与回滚](versioning-and-upgrades.md)。API正文覆盖不等于兼容承诺；正式
module path、tag、semver、license和发行授权仍是后续 fail-closed门禁。
