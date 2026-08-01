# Portable Skills Core API（Experimental）

`extensions/skills`拥有 AgentX Skill的可移植数据合同与确定性加载机制。它支持从
本地目录或只读 `fs.FS`加载 `SKILL.md`，解析 frontmatter与资源目录，并提供缓存、
activation path、requested-skill semantics、deep clone和资源引用检查。

完整接入示例见[Portable Skills 接入指南](../../docs/guides/portable-skills.md)和
[fixed-version consumer](../conformance/skills-consumer)。

## 数据合同

- `Source`标记 `custom`、`extra`、`bundled`、`managed`和 `workspace`来源；它描述
  provenance，不授权来源中的脚本或安装指令。
- `Skill`是解析后的主合同，包含名称、描述、内容、location/base dir、路径与工具提示、
  invocation/dispatch、requires/install、resources和 metadata。`InvocationPolicy`、
  `Requires`、`InstallSpec`、`Resources`、`DispatchSpec`只是数据；尤其
  `InstallSpec`不执行命令。
- `SkillConfig`、`Eligibility`、`RuntimeEligibility`和 `Decision`为 HS兼容所需的
  eligibility数据合同；本包不提供产品化 filter evaluator，调用方必须在 Host层实现。
- `FSSource`以 `ID`、`FS`和 `Fingerprint`声明只读来源；`Valid`只检查字段完整，只有
  `runtime/assetfs` provider能够进一步证明 immutable identity并启用共享缓存。
- `LoadOptions`声明目录/FS来源、strict/fail-fast模式及三个加载上限；零值上限使用实现的
  有界默认值。
- `LoadIssue`和 `LoadReport`保留加载阶段、路径、稳定 code/message、计数、cache hit和
  generation；`HasIssues`同时检查 issue列表与 parse-failed计数。
- `RequestedSkillSemantic`只投影显式请求的 name、execution context、allowed tools和
  effort，不代表授权已经通过。

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
- `ResolveSkillKey`及 `NormalizeToolHintsMatch`、`EffectiveToolHintsMatch`、
  `NormalizeSkillExecutionContext`、`NormalizeSkillAllowedTools`、
  `NormalizeSkillExecutionEffort`：提供确定性的 authoring normalization。

`Skill.ExecutionContext`、`AllowedTools`与 `Effort`只是请求和提示元数据，不拥有
authorization、approval或 sandbox语义。

## 最小目录加载

```go
items, report, err := skills.LoadWithReport(skills.LoadOptions{
    ExtraDirs:                []string{"./skills"},
    StrictFrontmatter:        true,
    MaxCandidatesPerRoot:     64,
    MaxSkillsLoadedPerSource: 32,
    MaxSkillFileBytes:        256 << 10,
})
if err != nil {
    return err
}
if report.HasIssues() {
    // Host决定记录、拒绝或降级；loader不替 Host做产品策略。
}
_ = items
```

若资源来自 `embed.FS`或其它内存文件系统，推荐先用 `runtime/assetfs.New`创建 immutable
provider，再把 `provider.ID()`、`provider.FS()`和 `provider.Fingerprint()`放入
`FSSource`。不要用调用方自报 fingerprint伪装 immutable source。

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
