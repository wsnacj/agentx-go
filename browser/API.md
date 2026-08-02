# `browser` 中文 API 总览

成熟度：`Experimental extension`

`browser`是AgentX可选重型Browser module。它不属于Core Foundation的默认依赖，也不提供
默认provider、credential、代理、登录态、生产网络或真实副作用授权。调用方必须显式选择
runtime/host/tools能力并注入具体backend与安全策略。

当前已落地：

- [`browser/runtime`](./runtime/API.md)：provider-neutral session/action/route/snapshot/
  capability/result合同，以及state、recovery、selection、observation和watch implementation。

后续P3-A节点将在真实source ready后增加`browser/host/browserd`、`browser/tools`与统一
fixed-version consumer；不会提前创建空package。

本页面与各package `API.md`记录private validation当前事实，不构成Public/Beta/Stable、
semver兼容承诺、正式module tag或发行授权。
