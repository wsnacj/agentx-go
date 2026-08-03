# `scenes/finance/brief` 中文 API Reference

成熟度：**Experimental extension**。

本包拥有财报简报 Pack、source-neutral evidence/result、确定性 evaluator 和显式 legacy case-frame helper。

- `Definition` / `RegisterInto` / `MaterializedDefaultWorkflow`：只读简报 Pack 与 Workflow；
- `EvaluateBrief`：检查主体、报告期间、来源 URL、重点与 guard，不调用模型；
- `BuildBriefCaseInput`：在 Host 已选择 Pack 后从明确输入构造兼容 case frame；它不执行路由、
  source discovery、下载或报告解析。

产品文案、免责声明、投资判断和真实报告处理不属于本包。

