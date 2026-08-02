# 安装与多 Module 引用

`agentx-go`当前采用一个仓库、四个独立 Go module。调用方只引入实际使用的
module：

| Module | 用途 |
| --- | --- |
| `github.com/wsnacj/agentx-go` | 根 Client、Run、错误和 ExecutionAdapter合同 |
| `github.com/wsnacj/agentx-go/components` | provider-neutral LLM合同 |
| `github.com/wsnacj/agentx-go/runtime` | Host Kit、toolloop、Workflow和其它 portable Runtime owner |
| `github.com/wsnacj/agentx-go/extensions` | 可选 portable extension机制；当前含 A股推荐入口、Domain Module、Pack与 Skills |

自定义 ExecutionAdapter 路径只需要根 module。Host Kit + Model/Tool Adapter
路径需要根、components和 runtime三个 module，因为配置显式使用根合同、LLM响应
和 Runtime类型。Workflow Host Kit路径只需要 Runtime module。A股推荐入口、
Host Kit、Pack和内嵌资产由 extensions module提供，并通过其固定依赖使用
runtime/components；调用方只需直接声明自己代码实际 import的 module。只从目录
加载 Skill时直接声明 extensions即可；若代码直接构造经 `runtime/assetfs`证明身份的
immutable `FSSource`，还应把 runtime列为直接依赖。

## 当前固定验证版本

```text
github.com/wsnacj/agentx-go
  v0.0.0-20260802103826-c7d80001682e
github.com/wsnacj/agentx-go/components
  v0.0.0-20260802103826-c7d80001682e
github.com/wsnacj/agentx-go/runtime
  v0.0.0-20260802103826-c7d80001682e
github.com/wsnacj/agentx-go/extensions
  v0.0.0-20260802103826-c7d80001682e
```

这些是不可变 private validation pseudo-version，不是 tag或正式 semver。

```bash
go get github.com/wsnacj/agentx-go@v0.0.0-20260802103826-c7d80001682e
go get github.com/wsnacj/agentx-go/components@v0.0.0-20260802103826-c7d80001682e
go get github.com/wsnacj/agentx-go/runtime@v0.0.0-20260802103826-c7d80001682e
go get github.com/wsnacj/agentx-go/extensions@v0.0.0-20260802103826-c7d80001682e
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
```

external-style consumer也是独立 nested module，应在 `GOWORK=off`下单独测试。
根[`conformance/consumer`](../../conformance/consumer)同时固定根、components、runtime
和extensions四module版本，在一个无HS、无Runner、无长期`replace`、无网络副作用的
consumer中验证自定义ExecutionAdapter、Model/Tool Host Kit、Workflow Host Kit三条
标准construction路径以及A股推荐extension入口。
`extensions/conformance/astock-contract-consumer`同时固定 runtime和 extensions，
用于验证无 HS、无长期 `replace` 的组合接入。
`extensions/conformance/domain-module-consumer`只固定 extensions，验证新项目可以
在无 HS、无 Runner、无长期 `replace` 时实现 config resolver和注册 callback。
`extensions/conformance/astock-consumer`固定 extensions/runtime，验证 A股 Manifest、
嵌入资产、三组 Pack、route/binding、fixture Host Kit和 evaluator的组合路径。
`extensions/conformance/skills-consumer`固定 extensions/runtime，验证 immutable
Skill加载、缓存、activation、requested semantics和资源引用；它不依赖 HS、Runner、
长期 `replace`、网络或命令执行。
长期 consumer不得依赖本地 `replace`；本地 `replace`只能用于临时开发测量，不能
作为 fixed-version或发布证据。

## Ubuntu实机复跑

仓库内当前候选lane固定Ubuntu 24.04 amd64与Go 1.25.12，在真实Linux进程中运行四module
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

该命令只在临时目录使用`v0.0.0-m6d.0`组装四module同版候选，不修改tracked `go.mod`，
并运行固定漏洞扫描、无replace consumer和只读cache复验。不得把token、URL rewrite或
runner credential写入仓库。

## 升级方式

Developer Preview期间，升级应显式修改 pseudo-version，并运行调用方 focused
tests、API differential和 module cache验证。完整步骤、回滚边界和已知非承诺见
[版本、升级与回滚](versioning-and-upgrades.md)。API正文覆盖不等于兼容承诺；正式
module path、tag、semver、license和发行授权仍是后续 fail-closed门禁。
