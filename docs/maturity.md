# 成熟度与兼容边界

## 当前状态

| 项目 | 状态 |
| --- | --- |
| 仓库 | private validation |
| 根 contract | W1 候选 |
| LLM contract component | W3-01 已落地，Experimental |
| Runtime construction | W2-A 已测量，W2-B 未获准 |
| Examples/conformance | W1-B private consumer 验证已提供 |
| HS canonical import | W1-C 已切换固定 private pseudo-version |
| HS LLM contract authority | W3-01 已切换到 `components/llm`，旧路径为 Deprecated shim |
| Public/Beta/Stable | 未授权 |
| 正式 tag/semver | 未授权 |
| License/NOTICE/release security | 仍是发行门禁 |

W1 的 exact export、签名、中文 Reference、examples 和 external-style consumer
用于发现意外漂移，但它们是
候选合同门禁，不等同于长期兼容承诺。

根合同的 authoring/source authority 已在 W1-C 后转移到本仓库。HS 旧
`experimental/facade` 只保留 Deprecated alias/forwarder，不再接受独立行为修改。

LLM 合同类型的 source authority 已在 W3-01 后转移到
`github.com/wsnacj/agentx-go/components/llm`。HS
`core/llmx/types` 只保留 Deprecated alias/forwarder；其余 consumer 将在后续
小批次切换，不能继续向旧路径增加行为。

## 明确 non-goal

以下能力没有进入 W1，不得根据 package 名、文档愿景或未来目录推断已支持：

- 官方 embedded Runtime 构造；
- Tool Direct Answer 的独立结果策略；
- Workflow 图执行；
- Objective Runtime Loop；
- 长任务调度、子 Session 和 durable lifecycle；
- Resume、progress/event stream；
- Scene、HTTP、CLI、provider、credential；
- 真实网络或生产副作用。

## 后续晋级

W1 checkpoint 只证明：

- 公共合同可以作为独立零第三方依赖 module 构建和测试；
- external-style consumer 能通过 canonical import 使用它；
- HS adapter 可以依赖同一份合同而不反向污染 owner package。

W2-A 已验证普通使用者需要的窄 Runtime construction 形状，但 closure 结论不允许
直接进入 W2-B。W3-01 只证明首个标准库闭包组件可以通过独立 module、
external-style consumer 和 HS compatibility shim 转移 source authority，不代表
45 个类型已通过公共 API 审批。Pre-Beta 还必须关闭版本、兼容、维护 owner、
Changelog、license、security 和发布授权门禁。任何一个步骤都不能自动把当前
合同升级为 Public、Beta 或 Stable。
