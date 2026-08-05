# Unified Capability Catalog API（Experimental）

`extensions/catalog` 为 Tool、Skill、Plugin、Connector、Expert、Team 提供同一份只读发现
合同。它解决“有哪些能力、如何稳定过滤和展示变化”，不解决“由谁执行、能否授权、如何安装”。

## Owner 边界

- `components/tool` 继续拥有 Tool schema、Handler 与 Executor；
- `extensions/skills` 继续拥有 Skill loader、activation path 与 requested semantics；
- Plugin/Connector/Expert/Team 的具体 manifest、凭据、实例和生命周期归 Platform 或业务 Host；
- 本包不读取环境、扫描默认目录、发送网络、执行命令或保存用户/租户状态。

`Asset` 只包含 display-safe 的 `Identity`、名称、描述、版本、来源引用、标签和关键词。不得把
credential、prompt 正文、Team 拓扑、安装命令、Handler、backend 句柄或租户身份塞进这些字段。

## 推荐入口

```go
assets := catalog.ProjectTools("canonical:tools", toolDefinitions)
assets = append(assets, catalog.ProjectSkills("workspace:skills", loadedSkills)...)
assets = append(assets, catalog.Asset{
    Identity:  catalog.Identity{Kind: catalog.KindExpert, ID: "researcher"},
    Name:      "Researcher",
    SourceRef: "host:experts",
    Tags:      []string{"research"},
})

index, err := catalog.Build(catalog.DefaultPolicy(), assets)
if err != nil {
    return err
}
result, err := index.Search(catalog.Query{
    Text:  "research",
    Kinds: []catalog.Kind{catalog.KindSkill, catalog.KindExpert},
    Limit: 10,
})
```

- `Build` 规范化 kind、identity、标签和关键词，拒绝重复 `Kind+ID`，按稳定次序生成 immutable
  `Catalog` 与 SHA-256 fingerprint；
- `Snapshot` 和所有 Search hit 都是 detached copy，可并发读取；
- `Search` 使用显式 lexical 规则。文本 token 为 AND，`AnyTags` 为 OR；同分时按 Kind/ID
  排序；`Score` 只解释目录排序，不能直接作为 Tool/Skill/Expert 执行路由；
- `Diff` 报告 added、removed、changed identity；
- `ProjectTools` 不复制参数 schema/handler，`ProjectSkills` 不复制正文/install/dispatch/resources。

## 错误和边界

`ErrorCode` 当前包含 `invalid_policy`、`invalid_asset`、`duplicate_asset` 与 `invalid_query`。
`Error` 的展示文本稳定且不包含 asset 内容、来源或底层 cause；调用方可使用 `errors.Is`、
`errors.As` 或 `AsError` 判断 code。

构建与检索都是受 `Policy` 限制的纯内存确定性操作，没有后台 goroutine、网络和
`Shutdown`。动态来源、revision、并发发布、可见性和服务关闭属于 Host catalog owner。

## 明确不包含

本包不包含统一执行 router、activation plan executor、Skill eligibility 产品策略、Plugin
安装/升级/签名、Connector credential、Expert/Team 实例化或运行编排、RBAC、远程 marketplace、
持久化、provider、Runner、Scene 或真实副作用。

当前为 Experimental，不构成 Public/Beta/Stable、semver 或 production-ready 承诺。
