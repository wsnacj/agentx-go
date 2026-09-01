# `tools/weather` 中文 API Reference

成熟度：`Experimental extension`

`weather`提供确定性的`weather_lookup` Tool，支持城市的当前天气和今日预报。portable实现拥有Open-Meteo
地理编码与forecast协议，但不拥有网络放行、代理、DNS/重定向、凭据或审计策略；Host必须注入
`httprequest.Preparer`。

```go
const Name = "weather_lookup"

type Options struct {
    Prepare           httprequest.Preparer
    GeocodingBaseURL string
    ForecastBaseURL  string
    TimeoutMs        int
    UserAgent        string
    Now              func() time.Time
}

type Request struct {
    Location string `json:"location"`
}

type Result struct { /* provider、地点、Current 与 Today */ }
type Current struct { /* 当前观测 */ }
type Today struct { /* 今日摘要 */ }

func Definition() tool.Definition
func Register(tool.Registrar, Options)
func NewHandler(Options) tool.Handler
func Run(context.Context, Request, Options) (tool.Result, error)
func Lookup(context.Context, Request, Options) (Result, error)
```

默认协议端点为Open-Meteo，但Host仍必须通过`Prepare`显式允许两个端点。Tool不支持任意未来日期；默认只查询
current与单日today结果。缺少地点、Host未注入网络端口、网络被拒绝、provider失败或context取消都会显式返回
错误，不会退回需要API key的provider或伪造天气。

`Result`包含provider、解析后的地点/国家/时区、UTC获取时间、当前温度/体感/湿度/风速/天气码，以及今日
最高温、最低温与天气码。温度单位为摄氏度，风速单位为km/h。
