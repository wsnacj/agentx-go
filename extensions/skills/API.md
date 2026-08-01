# Portable Skills Core API（Experimental）

`extensions/skills`拥有 AgentX Skill的可移植数据合同与确定性加载机制。它支持从
本地目录或只读 `fs.FS`加载 `SKILL.md`，解析 frontmatter与资源目录，并提供缓存、
activation path、requested-skill semantics、deep clone和资源引用检查。

## 推荐入口

- `Load`、`LoadWithReport`：按 `LoadOptions`加载多类 source；
- `LoadFromDirs`、`LoadFromDirsWithReport`：只加载显式目录；
- `FSSource`：以 ID、只读文件系统和 fingerprint描述 immutable source；
- `Clone`、`CloneLoadReport`、`LoadGeneration`：隔离调用方修改并观察 loader generation；
- `NormalizeSkillPathScopes`、`EvaluateSkillActivationPaths`、`SkillRequestedByName`：
  处理路径激活与显式请求；
- `ResolveRequestedSkillSemantics`、`MergeRequestedSkillSemantics`：投影 fork/allowed
  tools/effort提示元数据；
- `ExtractReferencedResourcePaths`、`MissingReferencedResourcePaths`：检查正文中的
  scripts/references/assets引用。

`Skill.ExecutionContext`、`AllowedTools`与 `Effort`只是请求和提示元数据，不拥有
authorization、approval或 sandbox语义。

## Source顺序与错误

`LoadOptions`的 source顺序固定为：Extra目录、Extra FS、Bundled目录、Bundled FS、
Managed目录、Custom目录、Workspace目录。同名 Skill由后出现的 source覆盖，最终按
名称排序。`MaxCandidatesPerRoot`、`MaxSkillsLoadedPerSource`和
`MaxSkillFileBytes`提供有界加载。

默认模式把读取、解析和大小问题记录到 `LoadReport.Issues`并继续；`FailFast`会按原始
顺序立即返回首个错误。`StrictFrontmatter`要求严格 YAML；默认模式保留历史 tolerant
解析兼容。调用方不得依赖未文档化的 YAML修复细节扩张格式。

## 缓存与并发

目录 source使用 fingerprint/generation和 `fsnotify`失效；只有能通过 canonical
`runtime/assetfs` identity attestation的 `FSSource`才启用 immutable cache。返回的
Skill和 report均为 detached copy，可由多个调用方并发读取。

当前 watcher没有公共 `Shutdown`合同，因此本包保持 Experimental。调用方不应把
当前内部 watcher生命周期解释为 Developer Preview稳定承诺。

## 明确不包含

本包不包含：prompt catalog/ranking、memory/browser路由启发式、eligibility/filter
产品策略、安装计划执行、命令/rollback、approval、安全规则、bundled Skill内容、
Runner调度、ProductShell、provider、credential、网络或真实副作用。这些责任由 HS
或调用方 Host显式拥有。

当前没有 Public/Beta/Stable、semver或正式发行承诺。
