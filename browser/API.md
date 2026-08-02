# `browser` 中文 API 总览

成熟度：`Experimental extension`

`browser`是AgentX可选重型Browser module。它不属于Core Foundation的默认依赖，也不提供
默认provider、credential、代理、登录态、生产网络或真实副作用授权。调用方必须显式选择
runtime/host/tools能力并注入具体backend与安全策略。

当前已落地：

- [`browser/runtime`](./runtime/API.md)：provider-neutral session/action/route/snapshot/
  capability/result合同，以及state、recovery、selection、observation和watch implementation。
- [`browser/host/browserd`](./host/browserd/API.md)：显式Plan和StatusProbe驱动的browserd
  process manager、内置Node资产、Playwright bootstrap/cache与ownership校验。
- [`browser/tools`](./tools/API.md)：Browser Tool schema、handler、session协调、能力投影和
  Browser Local Planner机制；具体backend与host ports由调用方显式注入。

browserd Host的构造不会启动进程、下载依赖或访问网络；只有调用方显式执行
`EnsureStarted(ctx)`才进入进程/网络副作用边界。Browser Tools未注入backend时fail closed，
不会回退到隐式系统浏览器。

本页面与各package `API.md`记录private validation当前事实，不构成Public/Beta/Stable、
semver兼容承诺、正式module tag或发行授权。
