# OCR Provider API（Experimental）

`document/ocr/provider`定义OCR上游服务的最小调用合同，并提供Baidu、TextIn和Volcengine
显式HTTP adapter。

## 主要合同

- `Provider.Call(ctx, Request)`返回原始`Response`；context必须传入HTTP请求；
- `Factory`和`Registry.Lookup`负责按kind构造provider；
- `DefaultFactories`与`DefaultConfigValidators`提供内置kind集合；
- `NewBaiduProvider`、`NewTextInProvider`、`NewVolcEngineProvider`只使用显式
  `ProviderConfig`；
- `BaiduError`、`TextInError`、`VolcError`和`ErrorCategorizer`保留provider错误分类。

构造不会从环境变量或credential store自动取密钥。调用方法可能访问真实网络；endpoint、
credential、代理、配额、审计和出网授权由Host负责。生产调用方应在更高层执行allowlist和
display-safe error投影。
