# `tools/message` 中文 API Reference

成熟度：`Experimental extension`

`message`提供portable channel message协调；Host显式注入`runtime/channel.TextSender`和当前
`TextTarget`，canonical implementation负责参数解析、目标覆盖/去重、能力优先级、exact-once
调用、partial failure及稳定JSON结果。

```go
const Name = "message"

type Options struct {
    Sender   channel.TextSender
    Target   channel.TextTarget
    Platform string
}

func Definition() tool.Definition
func Register(tool.Registrar, Options)
func NewHandler(Options) tool.Handler
```

## 行为

- 无`action`且有`text`时默认为`send`；两者都没有时为`current_target`；
- `reply`优先使用`channel.ReplySender`，否则回退到`TextSender.SendText`；
- `edit`、`delete`、`react`、`forward`只在sender实现对应可选能力时执行，否则fail closed；
- `broadcast`和`forward`稳定去重target，并对单目标失败返回display-safe partial result；
- 缺少`text`、`message_id`或目标时返回`runtime/toolerrors` typed error；
- context cancellation和sender error identity不会被吞掉。

## Host边界

本package不选择平台账号，不发现credential，不建立网络连接，也不拥有发送权限、频道
allowlist、审计或速率限制。上述能力由Host的sender adapter和产品policy负责。测试使用内存
sender，不产生真实消息副作用。
