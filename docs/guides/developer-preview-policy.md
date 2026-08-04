# Developer Preview 兼容政策

本政策适用于首次稳定版本之前的 AgentX Go。

## 受治理的推荐入口

推荐入口包括：

1. 根 `Client` 与自定义 `ExecutionAdapter`；
2. Model / Tool Host Kit；
3. Workflow Host Kit；
4. Objective Host Kit；
5. Session / Subagent Host Kit；
6. 成熟度矩阵中明确列出的 Developer Preview candidate。

其它可导入 package 默认按 Experimental 管理。Go exported 不自动形成稳定兼容承诺。

## 不得静默改变的行为

- `RunRequest`、`RunResult`、identity 与 status 映射；
- typed error code、`errors.Is/As`、retryability 与 display-safe message；
- context cancellation、deadline、并发 Run 与 `Shutdown(ctx)`；
- LLM/tool JSON、Workflow 状态转换与 durable write 顺序；
- credential、授权、网络、进程、文件和 backend 的显式 Host 边界；
- 推荐 package 路径和跨 module 类型身份。

## API 变化要求

推荐 API 的 additive 或 breaking 变化必须同时更新：

- 中文 Reference；
- 可读 API snapshot 与 hash；
- CHANGELOG 和升级说明；
- external-style consumer；
- affected module normal/race/vet/tidy/list；
- 真实 Host 的 focused compatibility。

Experimental API 可以调整，但必须明确影响，不能把已有 consumer 当作不存在。

## 弃用

非紧急删除应先标记 Deprecated、给出替代路径，并至少保留一个可升级版本窗口。安全或
数据完整性问题可以缩短窗口，但必须在发布说明中给出原因和缓解方式。

进入 Beta 前会重新确定：

- 兼容范围与弃用周期；
- 最低 Go 版本和支持平台；
- 安全响应与版本维护周期；
- module release train 和 tag；
- License 与正式发行责任。
