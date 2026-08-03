# OCR Splitter API（Experimental）

`document/ocr/splitter`定义输入文档到逐页图像的拆分合同。

- `Splitter.Split(ctx, Request)`返回`Result.Images`、stats和必须调用的`Cleanup`；
- `Factory`从显式`SplitterConfig`构造实现；
- `DefaultFactories`提供`poppler`与`remote`实现。

Poppler实现会读取输入文件、创建临时目录并启动显式系统命令；remote实现会读取文件并向
显式`base_url`发送HTTP请求。两者都传播context cancellation/deadline，且调用方必须执行
`Cleanup`。本包不会选择credential、代理、出网规则或进程授权；生产Host必须在构造前完成
路径、命令和endpoint治理。若只需要自定义拆分页，可实现`Splitter`并替换默认factory。
