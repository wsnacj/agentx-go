# Baidu OCR Processor API（Experimental）

`document/ocr/processor/baidu`把Baidu OCR原始响应解析为canonical OCR model。

唯一构造入口`NewProcessor(config.ProviderConfig)`返回`processor.ProviderProcessor`，
并由`For(OperationKind)`选择OCR、表格或印章处理器。它不发送HTTP请求、不读取credential，
只处理provider返回的bytes和可选diff baseline。

调用方通常通过`ocr.Dependencies.ProcessorFactories`或默认registry使用本包。Baidu wire
字段和错误兼容仍为Experimental，不构成Baidu服务可用性或商业授权承诺。
