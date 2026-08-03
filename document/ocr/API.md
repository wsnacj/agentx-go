# OCR API（Experimental）

`ocr` 提供既有 OCRX 的可移植 config/model、splitter、cache、worker、diff、processor 和
provider-neutral pipeline。package 名暂保留为 `ocrx`，以降低 HS consumer 迁移风险；import
path 以 `github.com/wsnacj/agentx-go/document/ocr` 为准。

低层扩展入口：[`config`](config/API.md)、[`model`](model/API.md)、
[`provider`](provider/API.md)、[`processor`](processor/API.md)、[`splitter`](splitter/API.md)、
[`cache`](cache/API.md)、[`worker`](worker/API.md)、[`pipeline`](pipeline/API.md)、
[`diff`](diff/API.md)和[`util`](util/API.md)。它们全部保持Experimental。

## 推荐构造

1. 构造 `config.ServiceConfig`；
2. 通过 `Dependencies` 注入 `ProviderFactories`、`SplitterFactories`、`CacheBuilder` 和
   `ProcessorFactories`；
3. 调用 `NewService`；
4. 使用 `RecognizeOCR`、`RecognizeTable` 或 `RecognizeStamp`，并传入调用方 `context.Context`。

`NewClientFromConfig` 仅接受调用方显式提供的配置路径。环境变量 credential/config 发现继续
由 Host adapter 拥有，不进入 canonical OCR package。

## 并发、取消与副作用

- `Service` 可并发读取已构造的 pipeline；worker 并发上限由配置决定；
- `context` cancellation/deadline 会终止等待和 provider call；
- cache 与 splitter 可能访问调用方配置的路径；真实 provider 只有在调用相应方法时才访问网络；
- module 不扫描 HS 配置、Scene、credential 文件或默认工作目录。

## Provider

Baidu、TextIn、Volcengine wire adapter 接收显式 `ProviderConfig`。credential 必须由调用方放入
显式 config，HTTP redirect 使用 bounded same-origin policy；生产 Host 仍应在执行合同中完成
endpoint allowlist、credential 管理、审计和限流。
