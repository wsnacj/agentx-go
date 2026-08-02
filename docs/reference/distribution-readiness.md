# Developer Preview分发Readiness

本页区分“授权开发者可以固定版本验证”和“可以向客户或公共proxy发布Beta”。

## 当前结论

| 结论 | 状态 | 含义 |
| --- | --- | --- |
| `private_validation_ready` | `true` | 已通过空缓存私有VCS获取与clean-room运行；只面向已获私有仓库权限的开发者，不是客户发行 |
| `public_beta_ready` | `false` | 不得创建Beta tag、公开推广或宣称production-ready |

## 已关闭项

- 四module统一fixed pseudo-version；
- 8个Developer Preview candidate API snapshot、公开类型闭包和双平台签名gate；
- 三条标准construction和A股推荐extension的无HS fixed consumer；
- 中文Reference、安装、升级和回滚说明；
- CODEOWNERS、贡献流程、支持边界和安全报告入口；
- 本地distribution preflight与module cache/zip provenance检查。
- 空`GOMODCACHE`经已批准的GitHub SSH传输获取四个固定版本module，并完成无HS
  clean-room consumer构建与运行；验证过程未输出或写入credential。

## Public Beta阻断

| 阻断 | 当前状态 | 解除决定 |
| --- | --- | --- |
| License/NOTICE | 未选择、仓库无`LICENSE*` | Owner与必要法律审阅选择开源、商业或双许可策略，并提交对应正式文本 |
| Security approval | 有报告入口，无具名独立approver签署 | 指定security approver并接受响应/残余风险 |
| Ubuntu运行证据 | 只有linux/amd64 CGO-disabled API/build对账 | 在Ubuntu运行批准的module/consumer验证矩阵 |
| Release authorization | 未批准tag、首次版本、Public proxy或客户分发 | release owner与独立reviewer批准release train和回滚 |
| 正式兼容等级 | 当前仅Developer Preview candidate | 完成Pre-Beta threshold并显式批准Beta surface |

License选项目前只作为决策输入，不代表法律建议或默认选择：可选择适合公共SDK采用的
开源许可证、仅授权客户的商业/专有许可，或双许可。M6A不得替Owner写入任何一种正式
`LICENSE`文本。

## 复跑入口

```bash
GOWORK=off go run scripts/check_developer_preview_distribution.go
GOWORK=off go run scripts/check_developer_preview_distribution.go -fresh-cache
```

第二条会使用空临时`GOMODCACHE`和`GOPROXY=direct`从私有远端获取固定版本；需要调用方
已经配置GitHub私有仓库读取权限，但不会把credential写入仓库或输出。
