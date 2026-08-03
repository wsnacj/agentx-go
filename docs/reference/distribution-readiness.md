# Developer Preview分发Readiness

本页区分“授权开发者可以固定版本验证”和“可以向客户或公共proxy发布Beta”。

## 当前结论

| 结论 | 状态 | 含义 |
| --- | --- | --- |
| `private_validation_ready` | `true` | 已通过空缓存私有VCS获取与clean-room运行；只面向已获私有仓库权限的开发者，不是客户发行 |
| `developer_portal_build_ready` | `true` | 当前九module正文已构建为207页本地Portal，77/14 coverage与本地搜索通过；不是公共托管结论 |
| `historical_core_ubuntu_snapshot_ready` | `true` | M6D曾在Ubuntu 24.04/amd64/Go 1.25.12/CGO=1复验四module快照；不自动覆盖当前九module |
| `current_nine_module_pre_beta_ready` | `false` | 当前九module未形成同一批准release train，也未获得新的全量Ubuntu、安全、license或发行签署 |
| `public_docs_hosting_ready` | `false` | 未批准域名、访问策略、部署或正式版本视图 |
| `public_beta_ready` | `false` | 不得创建Beta tag、公开推广或宣称production-ready |

## 已关闭项

- 历史四module统一fixed pseudo-version，以及当前九module的独立fixed consumer；
- 14个Developer Preview candidate API snapshot、公开类型闭包和双平台签名gate；
- 五条标准construction和已选Portable Scene入口的无HS fixed consumer；
- 中文Reference、安装、升级和回滚说明；
- CODEOWNERS、贡献流程、支持边界和安全报告入口；
- 本地distribution preflight与module cache/zip provenance检查。
- 空`GOMODCACHE`经已批准的GitHub SSH传输获取四个固定版本module，并完成无HS
  clean-room consumer构建与运行；验证过程未输出或写入credential。
- Core中文Developer Portal的clean install、零漏洞审计、93页production build、48/10
  coverage、本地搜索、响应式和浏览器交互验证。
- M6C GitHub Actions run `30732109611`在Ubuntu 24.04.4/amd64、Go 1.25.5、CGO=1
  实机上完成四module normal/race/vet/tidy/list、44/8双平台API、fresh private VCS、
  read-only module cache consumer和四module artifact Origin检查；精确source revision为
  `3c3c7fa46a2857b0ce04efffd046b352bd142e83`。
- M6D本地门禁已在Go 1.25.12上从当前tracked source生成四个`v0.0.0-m6d.0`临时
  module zip，完成四module test/vet/tidy、固定`govulncheck@v1.6.0`、无replace consumer
  和`GOPROXY=off`只读cache复验；0个可达漏洞。该模拟版本和zip仅用于临时技术验证，
  不进入正式分发。
- M6D GitHub Actions run `30733996721`在Ubuntu 24.04/amd64、Go 1.25.12、CGO=1与
  revision `2034622f991e0b704f2ac65fd2dfb36b0e97d2b2`上再次通过M6C runtime gate、
  44 package/8 candidate/2 target API gate、72文件/263链接文档gate、四module候选、
  漏洞扫描和离线consumer。artifact `8829013105`的归档digest为
  `sha256:7c97adf45c301856ad7c160e2c47326170d26bda4dee4d1d950004677cf69237`；
  四个module zip SHA-256与本地结果逐一一致。
- extensions依赖图仍包含`golang.org/x/sys@v0.13.0`的Windows-only
  `GO-2026-5024` module-level finding；当前10个extension package没有import或调用该
  漏洞。它不阻断当前darwin/linux技术候选，但必须由具名security approver在Beta前决定
  升级依赖或显式接受平台残余边界。

## Public Beta阻断

| 阻断 | 当前状态 | 解除决定 |
| --- | --- | --- |
| License/NOTICE | 未选择、仓库无`LICENSE*` | Owner与必要法律审阅选择开源、商业或双许可策略，并提交对应正式文本 |
| Security approval | 有报告入口，无具名独立approver签署 | 指定security approver并接受响应/残余风险 |
| Release authorization | 未批准tag、首次版本、Public proxy或客户分发 | release owner与独立reviewer批准release train和回滚 |
| 正式兼容等级 | 当前仅Developer Preview candidate | 完成Pre-Beta threshold并显式批准Beta surface |

License选项目前只作为决策输入，不代表法律建议或默认选择：可选择适合公共SDK采用的
开源许可证、仅授权客户的商业/专有许可，或双许可。M6A不得替Owner写入任何一种正式
`LICENSE`文本。

## 复跑入口

```bash
GOWORK=off go run scripts/check_developer_preview_distribution.go
GOWORK=off go run scripts/check_developer_preview_distribution.go -fresh-cache -read-only-cache
GOWORK=off go run scripts/check_developer_preview_distribution.go -portal
GOWORK=off CGO_ENABLED=1 go run scripts/check_ubuntu_runtime.go
GOWORK=off go run scripts/check_pre_beta_candidate.go
```

第二条会使用空临时`GOMODCACHE`从私有VCS获取固定版本与公共依赖，随后冻结cache并在
`GOPROXY=off`下消费；需要调用方已经配置GitHub私有仓库读取权限，但不会把credential
写入仓库或输出。
第三条是可选Developer Portal lane；需先运行`npm ci`，不会改变普通Go consumer或四
module验证对Node的零依赖边界。第四条只能在Linux amd64实机运行，并额外执行四module
race；GitHub-hosted复跑入口为`.github/workflows/m6c-ubuntu-runtime.yml`。
第五条构建可删除的同版候选、运行固定漏洞扫描并输出value-safe manifest；完整远端入口
为`.github/workflows/m6d-pre-beta-candidate.yml`。
