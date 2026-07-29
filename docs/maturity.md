# 成熟度与兼容边界

## 当前状态

| 项目 | 状态 |
| --- | --- |
| 仓库 | private validation |
| 根 contract | W1 候选 |
| Runtime construction | 未提供，计划进入 W2 |
| Examples/conformance | W1-B 后提供 |
| HS canonical import | W1-C 后切换 |
| Public/Beta/Stable | 未授权 |
| 正式 tag/semver | 未授权 |
| License/NOTICE/release security | 仍是发行门禁 |

W1 的 exact export、签名、中文 Reference 和行为测试用于发现意外漂移，但它们是
候选合同门禁，不等同于长期兼容承诺。

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

W2 才验证普通使用者需要的窄 Runtime construction。Pre-Beta 还必须关闭版本、
兼容、维护 owner、Changelog、license、security 和发布授权门禁。任何一个步骤都
不能自动把当前合同升级为 Public、Beta 或 Stable。
