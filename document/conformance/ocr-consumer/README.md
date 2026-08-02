# OCR fixed-version consumer

该 consumer 不依赖 HS、Runner、Scene 或长期 `replace`。它固定使用
`agentx-go/document@v0.0.0-20260802190326-aa1d8a5fdded`，并以纯内存 adapter 验证外部项目可构造
OCR Service、传播 `context` 并得到确定性结果；不会读取 credential、访问网络或启动进程。
