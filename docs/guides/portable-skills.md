# Portable Skills 接入

`github.com/wsnacj/agentx-go/extensions/skills`提供 Experimental 的 Skill数据合同与
portable加载机制。它适合需要读取 `SKILL.md`、检查资源、按路径激活并把显式请求语义
交给自身 Runtime的 Host；它不是完整的技能商店、安装器或授权系统。

## 安装

私有仓库访问方式见[安装与多 Module 引用](installation-and-modules.md)。目录加载只需：

```bash
go get github.com/wsnacj/agentx-go/extensions@v0.1.0
```

如果调用方直接使用 `runtime/assetfs`构造 immutable source，还应固定 Runtime：

```bash
go get github.com/wsnacj/agentx-go/runtime@v0.1.0
```

可重复构建的项目应把解析结果固定在`go.mod`和`go.sum`中。

## 路径一：从目录加载

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
for _, issue := range report.Issues {
    log.Printf("skill load issue: code=%s path=%s message=%s", issue.Code, issue.Path, issue.Message)
}
_ = items
```

默认模式把单项读取、解析和大小问题写入 report并继续；`FailFast`按原始顺序返回首个
错误。`StrictFrontmatter`拒绝不合法 YAML，默认 tolerant模式仅用于兼容现有内容，
新内容不应依赖未文档化的修复行为。

## 路径二：从 immutable `fs.FS`加载

```go
provider, err := assetfs.New("example.product.skills", embeddedSkills)
if err != nil {
    return err
}
opts := skills.LoadOptions{
    BundledFS: skills.FSSource{
        ID:          provider.ID(),
        FS:          provider.FS(),
        Fingerprint: provider.Fingerprint(),
    },
    StrictFrontmatter: true,
}
items, first, err := skills.LoadWithReport(opts)
if err != nil {
    return err
}
_, second, err := skills.LoadWithReport(opts)
if err != nil {
    return err
}
_ = items
_ = first.Generation
_ = second.CacheHit
```

只有通过 canonical `assetfs` identity attestation的 source会进入 immutable共享缓存。
普通 `fs.FS`仍可加载，但调用方提供的 ID/fingerprint本身不构成不可变证明。

## 激活、请求语义和资源

```go
active, reason, matched := skills.EvaluateSkillActivationPaths(
    items[0],
    []string{"src/main.go"},
)
requested := skills.ResolveRequestedSkillSemantics(items, []string{"repo-review"})
missing := skills.MissingReferencedResourcePaths(items[0])
_, _, _ = active, reason, matched
_, _ = requested, missing
```

- activation只判断 authoring path scope，不执行工具或授权；
- requested semantics只投影 fork/allowed-tools/effort提示，Host仍要实施 authorization、
  sandbox和资源预算；
- resource refs检查正文引用是否出现在已加载的 scripts/references/assets清单中，不执行
  这些文件；
- `Clone`和 `CloneLoadReport`返回 detached copy，避免调用方修改污染缓存快照。

## 并发与生命周期

返回的 Skill/report可由调用方并发读取；需要修改时先 clone。目录 source通过
fingerprint/generation和 `fsnotify`失效，immutable source按 asset identity复用缓存。
当前目录 watcher没有公共 `Shutdown`合同，因此 package保持 Experimental。长期进程
应避免无界创建不同目录集合，也不应把当前 watcher行为当作兼容性承诺。

## Host继续拥有的责任

以下能力没有进入 canonical Skills Core：prompt catalog/ranking、memory/browser路由
启发式、eligibility/filter、safety规则、安装计划执行、approval、rollback、命令执行、
bundled Skill内容、Runner/ProductShell集成、provider、credential和网络。调用方应在
最窄 Host adapter或产品层实现它们，不要向 `extensions/skills`反向注入业务策略。

可运行的固定版本示例位于
[`extensions/conformance/skills-consumer`](../../extensions/conformance/skills-consumer)：

```bash
GOWORK=off GOPROXY=off go -C extensions/conformance/skills-consumer test ./...
GOWORK=off GOPROXY=off go -C extensions/conformance/skills-consumer run .
```
