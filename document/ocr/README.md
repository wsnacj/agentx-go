# ocrx 模块概览

`core/ocrx` 提供新一代 OCR 管线的基础设施，目标是抽离拆页、缓存、并发、Provider 调用等通用能力，便于在同一框架内扩展常规 OCR、表格 OCR、印章检测与后续自定义策略。

## 目录结构

- `service.go`：面向外部的统一入口，封装 OCR/Table/Stamp 三类服务接口。
- `types.go` / `model/`：统一的数据结构定义，可映射至旧版 `core/ocr` 的返回值。
- `config/`：配置结构体及校验逻辑，支持按管线拆分 Provider、Splitter、Cache、Worker、Retry 等段落。
- `splitter/`：拆页策略，内置 Poppler 本地拆页与远程 HTTP splitter；支持分批拆页、并发控制与流式上传。
- `provider/`：Provider 接口及工厂注册，默认包含 TextIn、火山引擎、百度 OCR。
- `pipeline/`：调度管理器，串联拆页、缓存、Provider 调用、diff 与结果聚合。
- `cache/`：缓存构建器，默认提供文件系统缓存。
- `worker/`：简单的并发控制器。
- `util/`：常用辅助方法（例如缓存 Key 计算）。
- `diff/`：为差异化比对预留的占位包。

## 关键设计

1. **弹性配置**：`config.ServiceConfig` 按 Operation（ocr/table/stamp）划分，每个 Operation 可绑定不同的 Provider/Splitter/Cache/Worker 策略。
2. **拆页策略替换**：`splitter.Factory` 支持通过配置选择 Poppler、远程拆页或自定义实现；Poppler 默认使用 300 DPI，并支持 `batch_pages` / `max_parallel` 控制长 PDF 的分批拆页并发度。
3. **可插拔 Provider**：`provider.Factory` 与 `Dependencies` 结合，可在构建 `Service` 时注入不同 Provider；默认实现包含 TextIn 与火山引擎 Volcengine，均封装了鉴权、超时与结构化日志。
4. **更稳的缓存键**：`cache.Builder` 默认返回文件系统实现（BaseDir、TTL 可配置），当 `enabled=false` 时自动降级为 noop；本地文件缓存键优先使用内容指纹，避免仅靠 `mtime/size` 造成误命中。
5. **并发与资源控制**：`worker.Pool` 对调用并发进行限流；`pipeline.Manager` 对多页任务采用 fail-fast cancel，任一页失败会尽早取消同批其它页；`splitter.Result` 暴露 `Cleanup` 回调，用于统一清理拆页产生的临时目录。
6. **可配置重试**：`pipeline.Manager` 结合 `github.com/cenkalti/backoff/v5` 提供指数退避重试策略，通过 `config.Retry` 控制初始/最大间隔及最大尝试次数，并对每次重试输出详细日志。
7. **指标采集**：`internal/metrics` 集成 Prometheus Counter/Histogram，统计调用次数、重试次数、耗时及错误分类，帮助观察运行状况。
8. **Diff/Fuzzy 能力**：`diff` 包提供对 TextIn OCR/Table 原始 JSON 的差异比对，并通过 `FuzzyLocateText` 调用 `internal/fuzzy` 的快速模糊匹配，用于定位识别文本中的目标片段。多页 OCR/Table 场景会先合并原始响应再做 diff。
9. **结构化输出**：TextIn OCR 结果同时保留 `RecognizedText`、更适合规则/审阅的 `NormalizedText`，以及带 `ul/table/td` 结构的 HTML，便于 LLM 消费原始文档内容。
10. **配置加载**：`config.Load` 支持从 YAML 文件读取配置；`config.DefaultTextInConfig` 可基于 TextIn 鉴权信息快速组装默认三条管线。

### 内置 Provider

| kind | 适用场景 | 配置要点 |
| --- | --- | --- |
| `textin` | 通用 OCR / 表格 / 印章识别 | `auth.app_id`、`auth.secret_code` 必填，可复用 `config.DefaultTextInConfig`。 |
| `volcengine` | 火山引擎通用文字识别（OCRNormal） | `auth.access_key_id`、`auth.secret_access_key` 必填；`additional.action`/`version`/`region`/`service` 如不设置将分别默认为 `OCRNormal`、`2020-08-26`、`cn-north-1`、`cv`；支持 `options.approximate_pixel/mode/filter_thresh/half_to_full` 透传，并内置文本 diff 与模糊定位能力。 |
| `baidu` | 百度 OCR（通用 + 表格） | `auth.access_token` 或 `auth.api_key`/`auth.secret_key` 二选一；`additional.token_url` 默认为 `https://aip.baidubce.com/oauth/2.0/token`；可透传 `language_type`、`recognize_granularity`、`return_excel`、`cell_contents` 等参数，OCR/表格均已接入并支持文本 diff。 |

当前仓库只提供 TextIn 的 `config/example.yaml`。其中凭据均为明显不可用的占位值；请把配置复制到仓库外的受控位置后再填入真实凭据，不要把真实凭据提交到源码仓库。Volcengine、Baidu 暂无随包示例文件，应由 host 根据上表字段显式提供配置。

### HTTP 鉴权与重定向边界

- TextIn 的自定义 headers 不得覆盖 `x-ti-app-id` 或 `x-ti-secret-code`；字段名按大小写不敏感处理，发现冲突时配置校验会直接拒绝，鉴权值只由 `auth.app_id` 和 `auth.secret_code` 提供。
- 内置 TextIn、Baidu、Volcengine provider 的默认带凭据 HTTP client 最多跟随 `5` 次重定向，并要求每一跳与初始请求保持相同 scheme、hostname 和 effective port；未显式填写端口时，HTTP/HTTPS 分别按 `80`/`443` 比较。
- 这项约束只限制默认 provider client 的凭据随重定向跨 origin 传播，不等价于 SSRF 防护、egress allowlist、DNS 解析或重绑定防护、TLS 策略、代理策略、响应体大小或读取预算。Host 自行实现 provider 时，其 HTTP client 必须负责这些策略以及对应验证。

## 当前建议

- TextIn 主产物建议优先使用 HTML：对表格、列表和层次化内容保留标签，通常比纯文本更适合后续 LLM 解析。
- 规则抽取、日志审阅和 diff baseline 建议使用 `NormalizedText` / `NormalizedPageTexts`，可读性比原始拼接文本更好。
- 长 PDF 可以打开 `splitter.batch_pages`，并配合 `splitter.max_parallel` 控制拆页吞吐；若环境缺少 `pdfinfo`，会自动退回单次拆分。
- 如果接远程 splitter，可通过 `splitter.options.timeout` 调整 HTTP 超时；上传默认是流式 multipart，不会先把整份文件拼成内存字符串。

## 命令行示例

| 命令 | 说明 |
| --- | --- |
| `go run ./core/ocrx/cmd/ocrxtool -config /absolute/path/ocrx.yaml -mode ocr path/to/file.pdf` | 通用命令行工具。`-config` 必须是绝对路径；也可以省略该参数并设置绝对 `OCRX_CONFIG_PATH`，或使用 `TEXTIN_APP_ID`、`TEXTIN_SECRET_CODE` 让 OCRX owner 在内存中构造 TextIn 配置 |
| `go run ./core/ocrx/cmd/demo -config /absolute/path/ocrx.yaml -provider textin` | Demo 脚本要求显式绝对 `-config` 或绝对 `OCRX_CONFIG_PATH`，读取 `core/ocrx/cmd/demo/data` 下的样例，结果写入 `core/ocrx/cmd/demo/output` |

`ocrxtool` 的输出目录当前包含两层产物：
- 主 `*.json`：snake_case 的稳定结果快照，适合脚本和回归对比。
- `*_artifacts/`：`raw_files` 保留 provider 原始返回，`derived_files` 则按模式写出 `recognized.txt`、`normalized.txt`、`combined.html`、分页文本/HTML 或印章摘要，便于人工排查。

> 后续可考虑补充的命令：`serve`（提供 HTTP 接口）、`diff`（对比两份识别结果）、`watch`（监听目录自动识别）等。

## 快速调用示例

```go
client, err := ocrx.NewClientFromConfigOrEnv("/absolute/path/ocrx.yaml")
if err != nil {
    log.Fatal(err)
}

textRes, _ := client.RecognizeText("foo.pdf")
htmlRes, _ := client.RecognizeHTML("foo.pdf")
tableRes, _ := client.RecognizeTable("bar.pdf")
stampRes, _ := client.RecognizeStamp("stamp.png")

ocrx.WriteText("out/foo.txt", textRes.Text)
ocrx.WriteText("out/foo.normalized.txt", textRes.NormalizedText)
ocrx.WriteText("out/foo.html", htmlRes.HTML)
ocrx.WriteJSONResult("out/table.json", tableRes)
ocrx.WriteJSONResult("out/stamp.json", stampRes)
```

传入空路径时，`NewClientFromConfigOrEnv` 只读取 OCRX owner 管理的 `OCRX_CONFIG_PATH` 或 TextIn 环境凭据；它不会从当前工作目录、最近的 `go.mod` 或源码仓库中猜测配置。文件路径必须为绝对路径。

通过 TextIn 环境凭据在内存中构造的默认配置不会启用文件缓存，避免把敏感派生数据写入隐式的工作目录。需要缓存时，应由 host 在显式配置文件中选择缓存类型和受控绝对目录。
