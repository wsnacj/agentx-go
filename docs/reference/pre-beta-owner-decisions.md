# Pre-Beta Owner决策单

九module本地与Ubuntu技术门禁已经通过。本页只把剩余人工决定压缩到一个可审阅入口；
它不是法律意见，也不自动生成License、创建tag或升级API成熟度。

## 推荐方案

| 决定 | 推荐 | 当前状态 |
| --- | --- | --- |
| 首次版本 | 九module同一release train使用`v0.1.0-beta.1`及各module tag前缀 | proposed，未授权 |
| 首批Beta兼容面 | 只冻结9个Core推荐入口；5个Scene candidate继续Experimental | proposed，未授权 |
| License | 优先评估Apache-2.0；如需商业限制或双许可，应先完成法律审阅 | pending Owner/legal |
| Security approver | 指定一名对九module扫描、响应流程和残余依赖负责的人 | pending named approver |
| Release owner | `@wsnacj`可作为仓库release owner候选 | pending explicit acceptance |
| 独立reviewer | 必须与release owner分离，复核tag source、manifest、rollback和授权 | pending named reviewer |

推荐的9个Core Beta候选为：根`agentx`、`components/llm`、`runtime/execution`、
`runtime/hostkit`、`runtime/toolloop`、`runtime/workflow`、`runtime/workflow/hostkit`、
`runtime/objective/hostkit`和`runtime/session/hostkit`。`scenes/astock`、
`scenes/browserops`、`scenes/publicnews`、`scenes/companyresearch`和`scenes/docparse`虽然已进入
Developer Preview signature gate，但建议首个Beta继续标为Experimental，避免把Domain Kit
与Core兼容承诺捆绑。

## 为什么现在不能自动关闭

- License决定会改变复制、修改、专利授权和商业分发权利；不能由技术门禁代签。
- security approver和独立reviewer必须是能够承担后续响应的人，脚本或AI不能代替。
- Beta surface一旦冻结，就需要按semver处理兼容性，不能把14个Developer Preview candidate
  自动全部升级。
- 正式tag必须在上述决定完成后一次性指向同一批准source revision，不能重写失败tag。

Owner确认后，实施动作应限制为：提交正式License/NOTICE、写入具名责任、冻结批准的API
snapshot、生成同revision九module tag并执行一次最终release gate。不得借机继续迁移Scene、
重跑全仓inventory或扩大公共surface。
