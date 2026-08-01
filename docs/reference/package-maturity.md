# Package API 索引与成熟度矩阵

本页只评估当前四个 module 中实际存在的29个 production package。它不是历史
surface inventory，也不会把任何符号自动升级为 Public、Beta 或 Stable。

## 分级含义

| 分级 | 本阶段含义 | 兼容性含义 |
| --- | --- | --- |
| Developer Preview candidate | M3E 选定的 Open Tool Loop与 Workflow标准路径直接需要，签名和中文 Reference 进入 focused gate | 仅表示候选；当前没有 semver 或长期兼容承诺 |
| Experimental extension | 已有真实 implementation 和 consumer，但仍可能在 Beta 前调整 owner 或入口 | 调用方应固定伪版本并评估升级差异 |
| internalization candidate | 当前为迁移和 source-authority 收口而导出，Beta 前优先评估收进 internal 或由上层 Facade 隐藏 | 不建议新项目直接依赖 |

## 当前矩阵

| Package | 分级 | 中文 Reference | 主要用途 |
| --- | --- | --- | --- |
| `agentx` | Developer Preview candidate | [根 API](agentx.md) | Client、Run、错误、画像和生命周期 |
| `components/llm` | Developer Preview candidate | [LLM API](../../components/llm/API.md) | provider-neutral 模型、消息、工具和 usage 合同 |
| `runtime/execution` | Developer Preview candidate | [Execution API](../../runtime/execution/API.md) | 根 Client 与 Host 之间的 typed dispatch |
| `runtime/hostkit` | Developer Preview candidate | [Host Kit API](../../runtime/hostkit/API.md) | Model/Tool Adapter 组合与低样板 Client 构造 |
| `runtime/assetfs` | Experimental extension | [API](../../runtime/assetfs/API.md) | immutable asset snapshot、fingerprint与 resolver |
| `extensions/astock/contracts` | Experimental extension | [API](../../extensions/astock/contracts/API.md) | A股 portable DTO、JSON normalization与 assessment |
| `extensions/domainmodule` | Experimental extension | [API](../../extensions/domainmodule/API.md) | portable manifest、config、diagnostics与顺序注册编排 |
| `runtime/hosthttp/hostserver` | Experimental extension | [API](../../runtime/hosthttp/hostserver/API.md) | Host-deployed HTTP transport、request identity 与有界关闭 |
| `runtime/hosthttp/requestjson` | Experimental extension | [API](../../runtime/hosthttp/requestjson/API.md) | 有界严格 JSON 请求解码 |
| `runtime/hosthttp/resourcepolicy` | Experimental extension | [API](../../runtime/hosthttp/resourcepolicy/API.md) | Host 资源 allowlist 与预算收窄 |
| `runtime/toolloop` | Developer Preview candidate | [Tool Loop API](../../runtime/toolloop/API.md) | portable 多轮驱动、phase、检测和 assembly |
| `runtime/workflow` | Developer Preview candidate | [API](../../runtime/workflow/API.md) | Workflow Spec 数据合同与 Host admission seam |
| `runtime/workflow/hostkit` | Developer Preview candidate | [Workflow Host Kit API](../../runtime/workflow/hostkit/API.md) | Workflow标准构造入口；组合 portable lowering、journal、node execution和orchestration |
| `runtime/budget` | Experimental extension | [API](../../runtime/budget/API.md) | 预算判定 mechanism |
| `runtime/construction` | Experimental extension | [API](../../runtime/construction/API.md) | 高级 Host 构造生命周期 |
| `runtime/mediaartifact` | Experimental extension | [API](../../runtime/mediaartifact/API.md) | 媒体产物描述合同 |
| `runtime/promptcontext` | Experimental extension | [API](../../runtime/promptcontext/API.md) | prompt context 投影 |
| `runtime/protocol` | Experimental extension | [API](../../runtime/protocol/API.md) | Runtime wire/schema |
| `runtime/telemetry` | Experimental extension | [API](../../runtime/telemetry/API.md) | portable event 与 replay |
| `runtime/telemetry/safeerror` | Experimental extension | [API](../../runtime/telemetry/safeerror/API.md) | observation-safe error 投影 |
| `runtime/toolerrors` | Experimental extension | [API](../../runtime/toolerrors/API.md) | 结构化工具参数错误 |
| `runtime/workflow/composition` | Experimental extension | [API](../../runtime/workflow/composition/API.md) | lower→run 组合入口 |
| `runtime/workflow/journal` | Experimental extension | [API](../../runtime/workflow/journal/API.md) | durable 顺序与 host port |
| `runtime/workflow/lowering` | Experimental extension | [API](../../runtime/workflow/lowering/API.md) | portable lowering |
| `runtime/workflow/validation` | Experimental extension | [API](../../runtime/workflow/validation/API.md) | structural validation 与 policy port |
| `runtime/workflow/bindingstate` | internalization candidate | [API](../../runtime/workflow/bindingstate/API.md) | binding/state mechanism |
| `runtime/workflow/nodeexec` | internalization candidate | [API](../../runtime/workflow/nodeexec/API.md) | node execution coordination |
| `runtime/workflow/orchestration` | internalization candidate | [API](../../runtime/workflow/orchestration/API.md) | lowered-plan orchestration 内核 |
| `runtime/workflow/schema` | internalization candidate | [API](../../runtime/workflow/schema/API.md) | schema normalization/validation 内核 |
| `runtime/workflow/transition` | internalization candidate | [API](../../runtime/workflow/transition/API.md) | state transition 内核 |

机器可检查的同源清单位于
[`developer-preview-packages.tsv`](developer-preview-packages.tsv)。它只服务当前
29个 package 的覆盖与漂移门禁，不是新的全仓 API registry。

## 漂移门禁

```bash
GOWORK=off go run scripts/check_developer_preview_api.go
```

门禁会确认：

1. 四个 module 当前 production package 与矩阵一一对应；
2. 每个 package 都有非空中文 Reference；
3. 七个 Developer Preview candidate 的 `go doc -all` 签名未漂移。

更新候选签名必须同时完成 focused owner/consumer review、中文 Reference 修订和
baseline 更新。该门禁不生成完整文档站，也不替代未来 semver/API compatibility
工具。
