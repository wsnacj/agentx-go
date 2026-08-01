# ProductShell Preparation 固定版本 Consumer

这个隔离 consumer 只依赖已经推送的 `agentx-go/extensions` 与
`agentx-go/runtime` 固定 pseudo-version，不使用 HS、Runner、长期 `replace`、
真实 provider、网络、凭据或生产副作用。

示例由宿主显式提供确定性回调，验证以下完整准备链路：

1. `Input` 合并请求的 Case；
2. 显式 skill directive 解析为 requested skills；
3. portable Pack 选择与 binding；
4. Case binding 与有效 Case 校验；
5. Workflow materialization；
6. `PreparationPipeline` 的固定阶段顺序和最终结果。

它只证明外部 module 能消费 Experimental ProductShell preparation contract，
不代表默认 provider、持久化、完整 ProductShell Runtime 或正式发行承诺。

```bash
GOWORK=off GOPROXY=off go test ./...
GOWORK=off GOPROXY=off go run .
```
