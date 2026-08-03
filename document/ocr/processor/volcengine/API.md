# Volcengine OCR Processor API（Experimental）

`document/ocr/processor/volcengine`把Volcengine OCR原始响应转换为canonical OCR model。

唯一入口`NewProcessor(config.ProviderConfig)`返回`processor.ProviderProcessor`。本包只拥有
response parsing/aggregation，不拥有HTTP client、credential、endpoint、retry或网络授权。
调用方通常通过OCR顶层默认processor registry使用它。

wire字段和支持的operation仍为Experimental；新增provider能力必须保持既有解析错误和
payload JSON差分。
