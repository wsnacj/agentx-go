# 安装与多 Module 引用

`agentx-go`当前采用一个仓库、三个独立 Go module。调用方只引入实际使用的
module：

| Module | 用途 |
| --- | --- |
| `github.com/wsnacj/agentx-go` | 根 Client、Run、错误和 ExecutionAdapter合同 |
| `github.com/wsnacj/agentx-go/components` | provider-neutral LLM合同 |
| `github.com/wsnacj/agentx-go/runtime` | Host Kit、toolloop、Workflow和其它 portable Runtime owner |

自定义 ExecutionAdapter 路径只需要根 module。Host Kit + Model/Tool Adapter
路径需要三个 module，因为配置显式使用根合同、LLM响应和 Runtime类型。

## 当前固定验证版本

```text
github.com/wsnacj/agentx-go
  v0.0.0-20260729101644-c7c26d427ac2
github.com/wsnacj/agentx-go/components
  v0.0.0-20260729125257-bb6949793309
github.com/wsnacj/agentx-go/runtime
  v0.0.0-20260801024524-d545bfb941c5
```

这些是不可变 private validation pseudo-version，不是 tag或正式 semver。

```bash
go get github.com/wsnacj/agentx-go@v0.0.0-20260729101644-c7c26d427ac2
go get github.com/wsnacj/agentx-go/components@v0.0.0-20260729125257-bb6949793309
go get github.com/wsnacj/agentx-go/runtime@v0.0.0-20260801024524-d545bfb941c5
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
```

external-style consumer也是独立 nested module，应在 `GOWORK=off`下单独测试。
长期 consumer不得依赖本地 `replace`；本地 `replace`只能用于临时开发测量，不能
作为 fixed-version或发布证据。

## 升级方式

Developer Preview期间，升级应显式修改 pseudo-version，并运行调用方 focused
tests、API differential和 module cache验证。API正文覆盖不等于兼容承诺；正式
module path、tag、semver、license和发行授权仍是后续 fail-closed门禁。
