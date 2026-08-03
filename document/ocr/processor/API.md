# OCR Processor API（Experimental）

`document/ocr/processor`定义provider原始响应到canonical OCR model之间的解析扩展点。

- `OperationProcessor.Build`聚合同一次operation的多份原始响应；
- `OperationProcessor.Diff`计算可选baseline差异；
- `ProviderProcessor.For(kind)`为operation选择处理器；
- `Factory`从显式`ProviderConfig`构造provider processor；
- `Registry.Lookup(kind)`按稳定provider kind解析factory，缺失时返回错误。

本包不执行网络请求，也不拥有credential。实现必须可被并发pipeline安全调用，或由Host为
每次执行构造独立实例。`Registry`是普通map，构造完成后不应并发修改。

内置解析实现：[`baidu`](baidu/API.md)、[`textin`](textin/API.md)和
[`volcengine`](volcengine/API.md)。
