# `scenes/docparse` 中文 API Reference

成熟度：**Experimental extension**。本包随 `scenes/v0.2.2`提供，但不进入9包核心
兼容候选面；调用方应固定精确版本，并在升级时运行合同测试。

本包是 AgentX 文档理解 Domain Kit 的推荐入口，拥有可移植 Pack Definition、编译期
skill/tool 资产和稳定 identity。真实文件、OCR/PDF provider、私有 schema、credential、
artifact policy 与人工复核由 Host 持有。

## 推荐入口

- `PackDefinition`、`Definition`：返回 caller-owned `extensions/pack.Definition`。
- `RegisterPacksIntoRegistry`、`RegisterInto`：向显式提供的 canonical Pack registry 注册；
  nil registry 保持 no-op。
- `ExtensionFS`：读取只读、可分发的 `document-operations` skill 与七个 tool manifest。
- `ToolNames`、`SkillNames`：返回领域 identity。
- `LocateAssetsAt`：只接受 Host 显式提供的根目录；`LocateAssets` 与 `DomainRoot` 已弃用并
  fail closed，不会猜测源码 checkout。

## 两条接入路径

1. 已有结构化 parse result：使用 [`hostkit`](./hostkit/API.md) 的 `New`，把
   `parse_result` 传给 `ExtractFields`、`ExtractTable`、`Validate` 或 `Guard`。
2. 完整理解编排：由 Host 组合 `representation`、`profile`、`planner`、`adapters`、
   `fusion` 与 `understanding`，并显式注入 parser adapter。

## 行为边界

- 未匹配 profile 不会猜测业务 schema，只生成 review-required proposal。
- 缺失 page/bbox/table-cell evidence 时不会自动宣称 answer-ready。
- context cancellation 由 adapter/Executor 原样保留；本包不创建后台 goroutine。
- `result_path` 只有配置 `hostkit.ResultLoader` 时可用，canonical 包不直接读文件。

## 非目标

- 本地文件 allowlist、OCR/PDF/LLM client 与重试、缓存、费用策略；
- `core/card`、`core/table` concrete adapter、客户模板和私有 corpus；
- Scene lifecycle、Runner、credential、网络、真实副作用或业务结果背书；
- Public/Beta/Stable、tag 或正式发行承诺。
