# OCR Model API（Experimental）

`document/ocr/model`定义OCR、表格和印章三类operation在pipeline内部交换的provider-neutral
数据模型。

主要类型包括`Request`、`Response`、`Meta`、`OperationKind`、`OCRPayload`、
`TablePayload`、`StampPayload`及对应page/box/coordinate、diff summary结构。
`OperationKindOCR`、`OperationKindTable`、`OperationKindStamp`和`OperationKindAny`
用于选择处理器，不表示provider或产品mode。

本包只拥有数据合同，不访问文件、网络或credential。调用方不得依赖未记录的map字段作为
稳定业务schema；当前所有类型均为Experimental，JSON兼容由OCR顶层consumer差分保护。
