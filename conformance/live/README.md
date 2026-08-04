# Live Conformance

`conformance/live` 保存需要真实 provider、真实网络或其它显式外部依赖的可选验收程序。
它与 `examples` 的职责严格分开：

- `examples` 是默认可运行、确定性、无 credential、无网络的教学代码；
- 普通 `conformance/*-consumer` 是固定版本、无 HS、无长期 `replace` 的离线合同证据；
- `conformance/live` 是显式 opt-in 的真实集成证据，不进入默认 test、文档示例或发行 module。

当前入口：

- [`provider-smoke`](provider-smoke)：用固定版本的根合同、Runtime Host Kit、LLM 合同和
  OpenAI-compatible provider 运行一次真实模型对话。

所有 live consumer 必须满足：

1. 默认关闭，未显式启用时不得访问网络；
2. credential 只从调用环境读取，不写入源码、`go.mod`、artifact 或日志；
3. endpoint、model 和超时由 Host 显式提供，不内置生产 provider 默认值；
4. 不允许 import HS、Runner、Scene Host 或使用长期 `replace`；
5. 只输出 display-safe 结果，不打印 credential、请求头或原始敏感响应；
6. 失败只表示对应 live 环境或集成链路未通过，不能替代离线 module gate。

