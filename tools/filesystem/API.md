# `tools/filesystem` API 参考

`filesystem` 提供 AgentX 的可移植文件工具协调层。它实现 `read`、`write`、`edit` 和
`apply_patch` 的模型参数兼容、默认值与预算收口、稳定 schema/JSON、文本切片、精确替换和
自定义 patch 语法，但不直接选择或访问真实工作区。

成熟度：`Experimental extension / Developer Preview candidate`。当前不构成 Public、Beta 或
Stable 兼容性承诺。

## Host 边界

调用方必须实现 `Workspace`：

```go
type Workspace interface {
    Read(context.Context, ReadRequest) (ReadResult, error)
    Write(context.Context, WriteRequest) (WriteResult, error)
    Edit(context.Context, EditRequest) (EditResult, error)
    ApplyPatch(context.Context, ApplyPatchRequest) (PatchSummary, error)
}
```

Host 继续拥有工作区根选择、授权/审批、受保护路径、symlink 与路径逃逸防护、原子落盘、
durable write、具体 OS/对象存储后端，以及 `assetfs://` URI 的解析。canonical 包不会读取环境
变量、凭据或默认访问本机文件。

## 注册

```go
registry := tools.NewRegistry()
filesystem.Register(registry, filesystem.Options{
    Workspace: workspace,
})
```

未提供 `Workspace` 时不会注册任何工具，也不会退化为未经授权的本地文件访问。

## 可移植机制

- `SelectText`：从任意 `io.Reader` 执行二进制拒绝、零基行切片和 rune 数量限制；
- `EditText`：执行一次或全部精确文本替换，并在提交前检查输出上限；
- `ParsePatch`：解析 `*** Begin Patch` / `*** End Patch` 自定义语法；
- `ApplyUpdateChunk`：执行唯一、无歧义的 patch 上下文替换；
- `PatchSummary.FilesTouched`：以首次出现顺序返回稳定去重的实际变更文件。

`Workspace` 的 `Edit` 和 `ApplyPatch` 实现应复用这些函数，同时把读取、校验和提交放入 Host
自己的安全/原子事务中。

## 工具合同

| 工具 | 输入要点 | 成功结果 |
|---|---|---|
| `read` | `path`；可选 `start_line/max_lines/max_chars` | `path/start_line/line_count/content/truncated` |
| `write` | `path/content`；可选 `append/create_dirs` | `path/bytes_written/mode/files_touched` |
| `edit` | `path/old_string/new_string`；可选 `replace_all` | `path/replacements/files_touched` |
| `apply_patch` | `input` 自定义 patch 文本 | `added/modified/deleted/files_touched/text` |

运行时仍接受既有别名，例如 `file_path`、`text`、`oldText/newText`、`from/lines` 和 `patch`；
这些别名不会出现在 canonical schema 中。

## 错误与取消

- 缺少 `edit.old_string` 或 `apply_patch.input` 返回 `runtime/toolerrors.ToolArgumentError`；
- 参数过大、patch 边界错误、上下文不存在或歧义会在副作用前失败；
- `context` 原样传给 Host port，Host 必须在提交前尊重取消或 deadline；
- Host 的授权、安全和后端错误原样保留，工具层只在历史合同需要时增加工具名前缀。

## 并发

handler 本身不维护可变状态。`Workspace` 的并发安全、同一路径冲突和事务隔离由 Host 实现并
记录在其自身合同中。
