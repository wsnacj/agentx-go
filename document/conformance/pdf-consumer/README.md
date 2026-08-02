# PDF fixed-version consumer

该 consumer 固定使用 `agentx-go/document@v0.0.0-20260802192110-b18be0f45ec5`，不包含
HS、Runner、Scene 或长期 `replace`。它通过 deterministic fake 实现 `pdf.Runner`，证明外部项目
可以在不启动进程、不访问网络的情况下构造 parser、传播 context 并消费格式化结果。
