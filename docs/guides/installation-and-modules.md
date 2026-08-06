# 安装与多 Module 引用

AgentX Go 由九个 library module 组成。项目只应安装实际使用的模块。

当前私有开发基线的最低 Go 版本是 1.25.0，验证使用 Go 1.25.12。历史`v0.1.0`tag
存在重建/checksum冲突，不得作为首次公开安装基线；下列命令固定到已推送、不可变的开发提交。

## 最小安装

根 Client：

```bash
go get github.com/wsnacj/agentx-go@v0.1.1-0.20260805062037-378edb9eb58a
```

模型合同与 Runtime Host Kit：

```bash
go get github.com/wsnacj/agentx-go/components@v0.1.1-0.20260805062037-378edb9eb58a
go get github.com/wsnacj/agentx-go/runtime@v0.1.1-0.20260805062037-378edb9eb58a
```

按需增加：

```bash
go get github.com/wsnacj/agentx-go/extensions@v0.1.1-0.20260805062037-378edb9eb58a
go get github.com/wsnacj/agentx-go/providers@v0.1.1-0.20260805062037-378edb9eb58a
go get github.com/wsnacj/agentx-go/tools@v0.1.1-0.20260805062037-378edb9eb58a
go get github.com/wsnacj/agentx-go/browser@v0.1.1-0.20260805062037-378edb9eb58a
go get github.com/wsnacj/agentx-go/document@v0.1.1-0.20260805062037-378edb9eb58a
go get github.com/wsnacj/agentx-go/scenes@v0.1.1-0.20260805062037-378edb9eb58a
```

当前pseudo-version是私有开发基线，不是正式公开release。需要可重复构建的项目应固定精确版本，
不要依赖branch或`@latest`；首次公开版本必须使用从未发布的新tag并重做clean readback。

## 私有仓库访问

仓库保持私有期间，开发和 CI 环境可以配置：

```bash
go env -w GOPRIVATE=github.com/wsnacj/agentx-go
go env -w GONOSUMDB=github.com/wsnacj/agentx-go
```

Git 仍需通过 SSH 或 HTTPS credential 获得仓库权限。不要把 token、SSH rewrite、
`.netrc` 或 credential helper 内容写入源码、`go.mod`、示例和日志。

## 模块依赖方向

```text
root / components
        ↓
runtime
        ↓
extensions / providers / tools
        ↓
browser / document
        ↓
scenes
```

调用方可以只使用 root、components 和 runtime。Browser、Document 与 Scenes 是显式选择
的能力模块。

`scenes` 的 `go.mod` 包含 Browser 和 Document 依赖，但轻量 package 的实际 import graph
保持按需加载。例如 `publicsource` 和 `publictransport` 不直接导入 Browser 或 Document。
如果依赖下载、构建时间或二进制体积敏感，请针对具体入口运行：

```bash
GOWORK=off go list -deps github.com/wsnacj/agentx-go/scenes/publicsource
GOWORK=off go list -m all
```

## 多 Module 版本

首个 Developer Preview 采用九module同版release-train。后续版本是否继续锁步，以对应
Release说明为准。升级时：

1. 先升级依赖图底层的 root/components；
2. 再升级 runtime；
3. 然后升级 extensions/providers/tools；
4. 最后升级 browser/document/scenes；
5. 运行调用方自己的 test、race 和集成验证。

不要长期使用本地 `replace` 或永久 `go.work` 作为发布证据。短期本地联调可以使用，但提交
前必须移除，并以远端不可变版本重新验证。

## Examples

`examples` 是独立教学 module，固定一组经过验证的依赖组合。它不属于 library module，
不会被普通调用方传递依赖。推荐从以下路径开始：

- `examples/chat`
- `examples/tool-loop`
- `examples/workflow`
- `examples/objective`
- `examples/session-subagent`
- `examples/deterministic-scene`
- `examples/reference-host`

完整说明见 [examples README](../../examples/README.md)。
