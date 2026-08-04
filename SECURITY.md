# 安全政策

## 报告漏洞

请不要在公开 Issue、讨论、日志或提交中披露尚未修复的漏洞、凭据或可利用细节。

优先使用本仓库的 GitHub Private Vulnerability Reporting / Security Advisory。若该功能
不可用，请通过 CODEOWNERS 中维护者公布的私密联系方式提交：

- 受影响的 module、package 和版本；
- 最小复现步骤；
- 可能的影响和攻击前提；
- 已知缓解方式；
- 可以安全公开的时间范围。

安全分诊由`@wsnacj`负责，目标是在3个工作日内确认收到报告，再评估影响、修复范围和
披露计划。该目标不是修复SLA，也不承诺固定修复时间。首版暂无backup security owner；
维护者不可用时，响应可能延迟，这也是Developer Preview的已知运营风险。

## 安全边界

AgentX 不应默认发现或持久化 credential。Provider、网络、进程、文件、浏览器、OCR 和
生产 backend 必须由 Host 显式配置并承担授权边界。

以下内容不应提交到仓库、示例或构建日志：

- API key、token、cookie、私钥和真实生产 endpoint credential；
- 客户数据、真实会话内容和未脱敏 artifact；
- 本机绝对路径、临时 secret 文件和 credential store 导出；
- 含敏感值的 HTTP request/response dump。

## 支持范围

当前代码属于Developer Preview。只有仓库中正式发布并列入支持矩阵的版本才进入安全
分诊范围；开发分支和任意commit不自动形成长期支持或安全回补承诺。

安全修复可能要求调用方升级依赖或调整配置，但不会以安全名义隐瞒 error、JSON、授权或
副作用语义变化。无法安全兼容时，发布说明必须明确迁移和缓解步骤。

## 依赖与供应链

- Go module、Node lockfile 和嵌入资产变更必须经过差异审阅；
- 发布候选应执行依赖漏洞扫描、module zip/readback 和 clean-room consumer；
- 生成文件、缓存、测试 credential 和本地 replace 不得进入发行 artifact；
- License/NOTICE 或第三方归属不完整、未完成安全复核或未获得发行授权时，不得声明
  正式公开发行。
