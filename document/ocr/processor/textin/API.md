# TextIn OCR Processor API（Experimental）

`document/ocr/processor/textin`把TextIn OCR、表格和印章原始响应聚合为canonical OCR model，
并提供对应diff投影。

`NewProcessor(config.ProviderConfig)`返回`processor.ProviderProcessor`；处理器不调用网络、
不读取credential，也不决定重试。多页顺序由pipeline传入的raw/files顺序决定，解析失败会
返回错误而不是静默生成成功payload。

本包面向OCR service assembly，不是独立产品API；当前保持Experimental。
