# `extensions/domainkit` 中文 API Reference

成熟度：**Experimental extension**。本包提供C7 Deterministic Domain Kit的无模型执行
边界，不构成Public、Beta、Stable或semver承诺。

## 推荐构造

```go
runtime, err := domainkit.New(domainkit.Config{Modules: []domainkit.Module{{
    Manifest: mymodule.Manifest(),
    Handlers: map[string]domainkit.Handler{
        "lookup": lookupFixtureOrHostAdapter,
    },
}}})

result, err := runtime.Run(ctx, domainkit.RunRequest{
    ModuleID: "my-module",
    CaseID:   "lookup-fixture",
    Tool:     "lookup",
    Arguments: map[string]any{"id": "fixture-1"},
})
```

`New`先规范化全部`domainmodule.Manifest`，拒绝重复module、manifest外handler、nil handler
以及缺失handler，不会产生部分注册。每个manifest tool都必须有唯一Host handler。

`Run`只按明确`ModuleID + Tool`调用一次handler：不会调用模型，不进行自然语言路由，不选择
provider，也不隐式读取环境、credential或网络。`CaseID`只是display-safe的conformance/Host
关联标识，不参与dispatch。

## 结果与确定性

`RunResult.Output`保持既有工具协议：`string`与`[]byte`原样返回，其它payload使用
`encoding/json`编码。`OutputDigest`是Output字节的lowercase SHA-256。相同输入与确定性
handler应得到相同digest；Runtime不会谎称带网络、时钟或随机数的Host handler是确定性的。

Runtime会递归复制常见JSON参数map/slice，handler不能通过这些值修改调用方请求。Runtime
查找表是只读且可并发调用；具体handler是否支持并发仍由Host声明和保证。

## 错误合同

`Error`支持`errors.Is/As`，稳定code包括：

- `invalid_config`
- `invalid_request`
- `module_not_found`
- `tool_not_found`
- `handler_failed`
- `encoding_failed`

`Error.Error()`为兼容现有Host工具链保留cause文本；界面或外部日志必须使用
`DisplaySafeMessage`，不能把cause直接展示给用户。handler错误与context cause仍可通过
`errors.Is`检查。调用开始前已取消的context原样返回`context.Canceled`或
`context.DeadlineExceeded`。

## Host责任与non-goal

- authorization、approval、tenant policy与tool allowlist；
- provider、credential、network、filesystem和真实副作用；
- handler retry、rate limit、cache与durable backend；
- Scene业务判断、LLM路由、自动workflow或Objective调度；
- 把普通生产handler自动认定为deterministic；
- Public/Beta/Stable、正式tag或发行。
