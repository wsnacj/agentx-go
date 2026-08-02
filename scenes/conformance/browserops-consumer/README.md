# Browser Ops fixed-version consumer

该 consumer 只依赖 `agentx-go/scenes` 及其固定依赖版本，不使用 `replace`、HS、Runner、旧
Scene import、真实网络、浏览器、文件副作用或 credential。它用内存 fixture 验证 Browser Ops
Pack identity、canonical tools Registry、Host Kit runtime-call 投影和 evidence readiness 成功路径。

当前 fixed scenes 版本为 `v0.0.0-20260802230400-97a7e59508f5`；它是 private repository 的
Developer Preview 验证证据，不是 Public/Beta/Stable 或正式发行声明。
