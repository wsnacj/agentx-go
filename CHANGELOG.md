# Changelog

本文件只记录调用方能够观察到的能力、合同和兼容性变化。内部迁移过程、提交批次和
验证编号不属于产品变更。

格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)。

## Unreleased

当前没有尚未发布的调用方可见变化。

## [0.2.1] - 2026-08-07

### Fixed

- 重新发布九个不可变 Go module tag，修复 `v0.2.0` 曾被移动而导致公共模块代理与
  clean-room consumer 出现 checksum mismatch 的问题；API 与运行时行为不变。
- 安装文档、模块矩阵、API 版本常量与签名快照统一到 `v0.2.1`。

## [0.2.0] - 2026-08-06

### Added

- 根 `Client`、`Run`、typed error、context cancellation 和有界 `Shutdown` 合同；
- Model Conversation、Open Tool Loop 和 Tool Direct Answer Host Kit；
- Workflow、Objective、Session、Subagent、Scheduler 与 Resume 组合能力；
- provider-neutral LLM/Tool 合同及可选 Providers；
- 通用 Tools、Browser、Document 和 Portable Scenes 模块；
- 七类能力示例、Reference Host、external-style consumers 和中文 API Reference；
- `DefaultExecutionProfile`，明确根 Client 当前唯一支持的执行画像；
- `hostserver.ConfigFromEnv`，用于显式加载 Host transport 环境配置。
- Experimental Context Window、Memory lifecycle、Plugin、Connector/MCP、Expert与Team声明及Host组合能力。

### Changed

- `hostserver.DefaultConfig` 不再隐式读取 token 或 trusted-proxy 环境变量。需要约定环境
  配置的 Host 应调用 `ConfigFromEnv`；
- Workflow 和 Docparse 的低层可导入 package 明确标记为 Experimental，不作为推荐稳定入口。
- 首版核心兼容候选面收窄为根contract、`components/llm`及7个Runtime/Host Kit package；
  A股、Browser Ops、Company Research、Docparse和Public News入口保持可用但降为Experimental。
- 最低 Go 版本统一为1.25.0，以采用`x/net`与`x/text`的安全修复版本；发布验证使用
  Go 1.25.12。

### Security

- Browser bundle将Playwright升级到`1.55.1`，修复`CVE-2025-59288`涉及的浏览器安装脚本证书校验风险。

### Known limitations

- 当前是 Developer Preview，Experimental package 可能调整；
- Workflow、Objective 和长任务需要调用方提供 policy、executor 或 backend；
- 真实 provider、浏览器、OCR、进程和网络副作用必须显式配置；
- 不提供Beta/Stable兼容承诺、生产SLA或长期支持周期；
- Developer Preview不代表所有可导入package都具备相同兼容等级。

[0.2.1]: https://github.com/wsnacj/agentx-go/releases/tag/v0.2.1
[0.2.0]: https://github.com/wsnacj/agentx-go/releases/tag/v0.2.0
