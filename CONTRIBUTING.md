# 贡献指南

感谢你改进 AgentX Go。架构合理性和可验证的外部合同优先于快速增加功能。

## 开始之前

1. 先确认能力应属于 root、components、runtime、extensions、providers、tools、
   browser、document 或 scenes 中的哪一个 owner；
2. 领域规则、客户 schema、credential、安全策略和生产 backend 应留在调用方 Host，
   不得下沉到通用 Runtime；
3. 新的进程、文件、网络或系统副作用必须显式构造、默认关闭并可替换；
4. 修改推荐 API 时，同时更新中文 Reference、examples、签名快照和升级说明。

## 开发流程

- 从当前开发分支创建短期分支；
- 保持提交聚焦，使用 `<area>: <imperative summary>` 风格；
- 不提交凭据、机器路径、缓存、构建产物或真实客户数据；
- 可以通过 Pull Request 或提交范围审阅协作；无论采用哪种方式，都必须保留测试证据。

## 九个 Library Module

```text
.
├── components
├── runtime
├── extensions
├── providers
├── tools
├── browser
├── document
└── scenes
```

根目录本身也是一个 module。`examples` 是独立教学 module，不属于 library release
surface。修改多个 module 时，应从依赖图底部向上验证，禁止 owner package 反向依赖
Host、具体 provider 或 Scene。

## 测试

在每个受影响 module 中运行：

```bash
GOWORK=off go test ./...
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
GOWORK=off go mod tidy -diff
GOWORK=off go list ./...
```

涉及公开合同或文档时，再运行：

```bash
GOWORK=off go run ./scripts/check_developer_preview_api.go
GOWORK=off go run ./scripts/check_package_api_docs.go
GOWORK=off go run ./scripts/check_docs_links.go
npm run docs:check
```

依赖浏览器、OCR、CGO 或系统命令的改动需要记录实际验证平台。默认测试不得要求真实
credential 或生产网络。

## API 与兼容性

- Go exported 不自动等于承诺兼容的公共 API；
- Developer Preview candidate 受到签名和中文 Reference gate；
- Experimental package 可以调整，但必须在 CHANGELOG 中说明调用方可见影响；
- 不得静默改变 error code、JSON、取消、状态转换、durable write 顺序或 `Shutdown` 语义；
- 删除入口前应提供替代路径和迁移说明。

## 贡献许可

AgentX Go 使用 [Apache License 2.0](LICENSE)。除非贡献者另行明确声明，主动提交并被
项目接收的贡献按照该许可证第 5 节处理，不附加额外条款。提交代码前还应确认自己有权
提供相关实现、文档、测试和素材；不得复制许可证不兼容或来源不明的代码。

安全问题请按照 [SECURITY.md](SECURITY.md) 私下报告。
