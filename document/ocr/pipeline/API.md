# OCR Pipeline Manager API（Experimental）

`document/ocr/pipeline`组合显式注入的provider、splitter、cache、worker pool和operation
processor，执行一次OCR、表格或印章pipeline。

```go
manager, err := pipeline.NewManager(kind, cfg, provider, splitter, cache, pool, processor)
result, err := manager.Run(ctx, request)
```

`NewManager`拒绝nil依赖；启用diff且配置baseline时会读取调用方指定文件。`Run`负责拆分、
并发限流、cache、provider重试、聚合和可选diff，并传播context cancellation/deadline。
首个job错误会取消同批剩余工作。

真实网络、进程、文件与cache副作用来自注入实现和显式配置；本包不发现credential或默认
provider。普通调用方优先使用`document/ocr.NewService`，只有需要替换底层协作者时才直接
构造Manager。
