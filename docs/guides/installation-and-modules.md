# 安装与多 Module 引用

`agentx-go`当前采用一个仓库、四个独立 Go module。调用方只引入实际使用的
module：

| Module | 用途 |
| --- | --- |
| `github.com/wsnacj/agentx-go` | 根 Client、Run、错误和 ExecutionAdapter合同 |
| `github.com/wsnacj/agentx-go/components` | provider-neutral LLM合同 |
| `github.com/wsnacj/agentx-go/runtime` | Host Kit、toolloop、Workflow和其它 portable Runtime owner |
| `github.com/wsnacj/agentx-go/extensions` | 可选 portable领域扩展合同；当前含 A股 contracts与 Domain Module注册编排 |

自定义 ExecutionAdapter 路径只需要根 module。Host Kit + Model/Tool Adapter
路径需要根、components和 runtime三个 module，因为配置显式使用根合同、LLM响应
和 Runtime类型。Workflow Host Kit路径只需要 Runtime module。A股 portable
contracts和 host-neutral Domain Module注册编排只需要 extensions module；若同时
使用 immutable asset loader，再显式加入 runtime module。

## 当前固定验证版本

```text
github.com/wsnacj/agentx-go
  v0.0.0-20260729101644-c7c26d427ac2
github.com/wsnacj/agentx-go/components
  v0.0.0-20260729125257-bb6949793309
github.com/wsnacj/agentx-go/runtime
  v0.0.0-20260801051155-7203f1b5be0a
github.com/wsnacj/agentx-go/extensions
  v0.0.0-20260801054929-4cae9842b02a
```

这些是不可变 private validation pseudo-version，不是 tag或正式 semver。

```bash
go get github.com/wsnacj/agentx-go@v0.0.0-20260729101644-c7c26d427ac2
go get github.com/wsnacj/agentx-go/components@v0.0.0-20260729125257-bb6949793309
go get github.com/wsnacj/agentx-go/runtime@v0.0.0-20260801051155-7203f1b5be0a
go get github.com/wsnacj/agentx-go/extensions@v0.0.0-20260801054929-4cae9842b02a
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
`extensions/conformance/astock-contract-consumer`同时固定 runtime和 extensions，
用于验证无 HS、无长期 `replace` 的组合接入。
`extensions/conformance/domain-module-consumer`只固定 extensions，验证新项目可以
在无 HS、无 Runner、无长期 `replace` 时实现 config resolver和注册 callback。
长期 consumer不得依赖本地 `replace`；本地 `replace`只能用于临时开发测量，不能
作为 fixed-version或发布证据。

## 升级方式

Developer Preview期间，升级应显式修改 pseudo-version，并运行调用方 focused
tests、API differential和 module cache验证。API正文覆盖不等于兼容承诺；正式
module path、tag、semver、license和发行授权仍是后续 fail-closed门禁。
