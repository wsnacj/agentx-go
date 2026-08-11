# 成熟度与兼容边界

AgentX Go 当前处于 Developer Preview。已有真实实现、外部消费路径、中文 API 和跨模块
验证，但尚未声明 Beta、Stable 或生产 SLA。

## 成熟度分级

| 分级 | 含义 | 调用方建议 |
| --- | --- | --- |
| Developer Preview candidate | 推荐接入路径，进入签名、类型闭包、中文 Reference 和 consumer gate | 可以试用；升级时阅读 CHANGELOG 并运行自身回归 |
| Experimental extension | 有真实实现和 consumer，但入口、owner 或字段仍可能调整 | 固定精确版本；避免把低层结构扩散到业务 API |
| Internal | 位于 Go `internal`，只供上层 package 实现使用 | 外部项目不能导入 |

Go exported 只表示语言层可见，不自动等于兼容承诺。逐 package 分级见
[Package API 与成熟度矩阵](reference/package-maturity.md)。

## 推荐能力状态

| 能力 | 当前状态 | 关键边界 |
| --- | --- | --- |
| 根 Client / Run / Error / Shutdown | Developer Preview candidate | 同步 Run；并发调用被串行化；关闭后拒绝新 Run |
| Model Conversation | Developer Preview candidate | 调用方显式提供模型函数 |
| Open Tool Loop / Tool Direct Answer | Developer Preview candidate | 模型、工具授权和执行由 Host 提供 |
| Workflow | Developer Preview candidate | validator、mapper、executor、identity 和可选 durable port 由 Host 提供 |
| Objective | Developer Preview candidate | strategy、policy、approval 和 handler 由 Host 提供 |
| Session / Subagent / Resume | Developer Preview candidate | worker、state、queue 和持久化由 Host 提供 |
| Providers / Tools | Experimental | credential、网络、sandbox 和产品策略显式注入 |
| Browser / Document | Experimental | 外部进程、文件、Python、OCR 和资源预算显式启用 |
| Portable Scenes | Experimental | 真实数据源、客户规则和 backend 留在领域 Host |

## 模块边界

九个 library module 按单向依赖组织：

```text
root / components
        ↓
runtime
        ↓
extensions / providers / tools
        ↓
browser / document
        ↓
scenes
```

重型 Browser 和 Document 不进入根合同。`scenes` 是显式选择的 batteries module；
轻量 `publicsource`、`publictransport` package 的实际 import graph不包含 Browser 或
Document，但 module graph仍会看到这些可选依赖。对依赖成本敏感的调用方应按所选 package
执行 `go list -deps` 和构建测量。

## 兼容承诺

Developer Preview 期间：

- 推荐 API 的变化必须同步 Reference、签名快照、CHANGELOG、consumer 和迁移说明；
- 不得静默改变 error code、JSON、取消、状态转换、durable write 顺序和 `Shutdown` 语义；
- Experimental API 可以调整，但必须说明调用方可见影响；
- 版本升级应使用不可变版本并保留可回滚点；
- 进入 Beta 前会重新明确弃用窗口、最低 Go 版本和支持周期。

## 平台与工具链

- module language baseline：Go 1.25.0；
- 首版发布验证工具链：Go 1.25.12；
- API/build surface 已持续覆盖 macOS arm64 和 Linux amd64；
- Browser、OCR、CGO、Python、ffmpeg/ffprobe 等可选能力需要额外平台依赖；
- 未经测试的平台不应被推断为受支持。

## 当前非目标

- 自动发现 credential、默认 provider 或默认生产网络；
- 无需调用方提供模型、工具或 policy 的全功能生产 Host；
- 内置客户业务规则、授权策略或生产 backend；
- 把所有 Experimental package 自动升级为稳定 API；
- 通过根 `ExecutionProfile` 字符串组合启用 Workflow、Objective 或长任务。

## 正式公开发行条件

首次正式开源发布至少需要：

1. 已完成：提交 Apache-2.0、NOTICE、九个 library module 的分发副本和直接依赖归属摘要；
2. `v0.2.2`九module同版artifact、tag前缀与9包核心兼容候选面一致；
3. 当前源码通过test、race、vet、tidy、list、module zip和clean-room/offline consumer；
4. Go与Node安全扫描通过，九module可达漏洞为0；
5. 文档站、examples、Package API、双平台签名和Release内容一致；
6. `@wsnacj`承担首版安全、发布与回滚责任，暂无backup maintainer。

满足这些条件只表示可以发布Developer Preview，不表示Beta、Stable、production-ready或生产SLA。
