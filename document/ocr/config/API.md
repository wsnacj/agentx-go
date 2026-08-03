# OCR Config API（Experimental）

`document/ocr/config`定义OCR service、provider、splitter、cache、worker、retry、limit和diff
的配置合同。

## 主要类型与入口

- `ServiceConfig`聚合服务配置；`PipelineConfig`描述单类OCR执行；
- `ProviderConfig`、`SplitterConfig`、`CacheConfig`和`WorkerConfig`分别配置显式依赖；
- `RetryConfig`、`LimitConfig`与`DiffConfig`约束重试、资源上限和结果差异；
- `Load(path)`从调用方指定文件读取配置；
- `DefaultTextInConfig(appID, secret)`只使用显式参数构造TextIn配置。

`Load`会产生文件读取副作用，但不搜索工作区、环境变量或默认credential路径。
credential生命周期、endpoint allowlist、配置文件权限与secret注入由Host负责。配置结构仍为
Experimental；字段默认和JSON/YAML兼容变更必须经过consumer差分。
