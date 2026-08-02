# `tools/httprequest` 中文 API Reference

成熟度：`Experimental extension`

该package拥有HTTP tool的portable request shaping、query/header/body组装、timeout/max-char/
redirect预算收窄、bounded response读取、external-content标记、provider diagnostics和稳定错误
格式。Host通过`Preparer`显式注入URL校验、redirect policy、proxy/transport与HTTP client。

```go
const Name = "http_request"

type HTTPDoer interface {
    Do(*http.Request) (*http.Response, error)
}
type Preparer func(context.Context, PrepareInput) (PreparedRequest, error)
type Options struct { /* Prepare、预算、UserAgent与可选clock */ }
type Request struct { /* method/url/headers/query/body与调用方预算 */ }

func Definition() tool.Definition
func Register(tool.Registrar, Options)
func NewHandler(Options) tool.Handler
func Run(context.Context, Request, Options) (tool.Result, error)
func NormalizeMethod(string) string
```

`Run`只通过调用方提供的`Preparer`和`HTTPDoer`执行一次请求，不读取环境变量、credential或
proxy配置。初始URL与每次redirect的SSRF/private-host/port/CIDR策略、trusted proxy、TLS、
审计和生产出网授权属于Host。`Register`在未提供`Preparer`时不注册工具。

测试使用内存`HTTPDoer`，不监听端口、不访问真实网络。
