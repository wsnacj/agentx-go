# `tools/process` 中文 API Reference

成熟度：`Experimental extension`

该package提供显式 opt-in 的本地前台命令与只读进程列表adapter，拥有command/result、bounded
stdout/stderr、caller cancellation、adapter timeout、非零exit和进程终止证据等portable语义。
它不会自动注册模型tool，也不会在构造时执行命令。

```go
type LocalOptions struct {
    Root               string
    DefaultTimeout     time.Duration
    MaxOutputBytes     int
    ProcessOutputBytes int
}
type LocalAdapter struct { /* unexported */ }
func NewLocalAdapter(LocalOptions) *LocalAdapter

type Command struct {
    Command        string
    Workdir        string
    Env            map[string]string
    Timeout        time.Duration
    MaxOutputBytes int
}
type CommandResult struct { /* command、workdir、exit、bounded output、duration与termination */ }
type Termination struct { /* reason、signal、process_group与wait_delay_ms */ }
func (*LocalAdapter) Run(context.Context, Command) (CommandResult, error)

type ListRequest struct { Limit int }
type ListResult struct { Lines []string; Truncated bool }
func (*LocalAdapter) List(context.Context, ListRequest) (ListResult, error)
```

`Run`只执行non-interactive foreground command，不启动detached/background worker。非零exit是
`CommandResult`，不是adapter error；由`Command.Timeout`或`DefaultTimeout`触发的超时返回带
`timed_out`和partial output的结果。caller context取消或deadline则返回typed `*Error`，并保持
`errors.Is(err, context.Canceled/DeadlineExceeded)`成立。Unix取消会终止整个process group，避免
shell child脱离超时边界。

```go
type ErrorCode string
type Error struct {
    Code ErrorCode
    Op string
    Message string
    Cause error
}
func AsError(error) (*Error, bool)
```

稳定code包括`invalid_command`、`workdir_resolution_failed`、`command_failed`、
`context_canceled`、`context_deadline_exceeded`、`process_list_failed`和
`output_limit_exceeded`。调用方应使用`errors.As`或`AsError`判断，不依赖错误文本。

## Host 责任与安全边界

`LocalOptions.Root`只提供路径containment和symlink escape防护，是非 sandbox 边界。Host必须在
调用前完成authorization、approval、command allow/deny policy、sandbox/容器、审计、secret
处理、资源配额和生产部署策略；signal选择与kill授权也不属于本package。

外部项目只有在明确接受本地副作用后才应构造`LocalAdapter`。仅需要portable model/tool
coordination且不希望提供本地进程能力的Host，不应构造或暴露它。
