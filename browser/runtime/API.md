# `browser/runtime` 中文 API Reference

成熟度：`Experimental extension`

该package拥有provider-neutral Browser Runtime合同与状态机制，不启动浏览器、不读取
credential，也不选择proxy、企业网络或产品policy。所有真实浏览器动作均通过调用方实现的
`BrowserBackend`及其细分capability接口进入。

## 核心入口

```go
type BrowserBackend interface { ... }
type BrowserToolOptions struct { ... }
type BrowserRuntimeInfo struct { ... }
type BrowserSessionRoute struct { ... }
type BrowserSessionTarget struct { ... }

func NewBrowserSessionRegistry() *BrowserSessionRegistry
func NewBrowserSessionStateRegistry() *BrowserSessionStateRegistry
```

`BrowserSessionRegistry`并发安全，按session与route跟踪tab/current target/pending review；
`BrowserSessionStateRegistry`维护profile lifecycle与selection。调用方仍拥有持久化、授权、
进程生命周期和租户隔离。

## Action与backend能力

基础`BrowserBackend`覆盖open、navigate、tabs、extract、snapshot、screenshot、click、type和
eval。可选细分接口覆盖artifact、console/request/response、cookie/storage、trace、highlight、
dialog/upload/press/hover/drag/select/fill、geolocation、device、headers、locale、media、offline、
profile lifecycle与raw observation。Host只实现实际支持的能力；缺失能力必须fail closed。

request/result类型保留现有JSON字段和错误/状态语义。runtime不拥有tool schema或模型参数解析，
这些属于后续`browser/tools`。

## Session协调

`ApplySharedSessionBrowser*`、`ResolveSharedSessionBrowser*`、`SyncSharedSessionBrowser*`与
`ProjectSharedSessionBrowser*`函数组合target tracking、profile health、recovery、selection、
observation、watch和tool-facing projection。它们不执行真实浏览器动作，也不替Host决定
approval、credential或客户策略。

## Tool surface识别

```go
func BrowserUnifiedToolNames() []string
func BrowserSpecialistToolNames() []string
func BrowserCompatToolNames() []string
func BrowserAllToolNames() []string
func IsBrowserToolName(string) bool
func BrowserRuntimeActionForToolCall(string, map[string]any) string
```

这些函数只固定现有名称与action映射，不注册handler。返回的slice为防御性副本。

## 并发、取消与关闭

- registry/watch manager内部状态具备并发保护；调用方不得绕过公开方法写入；
- 接受`context.Context`的执行/观察函数传播取消与deadline；
- 本package不拥有进程或网络资源，因此没有隐式`Shutdown`；真实Host必须为自身backend提供
  有界、幂等关闭合同；
- package当前为Experimental，当前导出面尚未形成Public兼容承诺。

## Non-goal

- 默认browserd、Playwright/CDP provider或自动下载；
- proxy、cookie/登录态、credential、企业网络、authorization/approval；
- AgentX tool注册、Scene业务规则、客户allowlist与生产部署；
- Public/Beta/Stable或跨版本兼容承诺。
