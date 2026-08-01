# Immutable AssetFS API（Experimental）

`assetfs`是 AgentX编译期资源的 immutable、content-addressed文件系统机制 owner。
它只依赖 Go标准库，不拥有 Scene业务语义或 host文件系统策略。

## Provider

```go
provider, err := assetfs.New("scene.demo", embeddedFS)
fingerprint := provider.Fingerprint()
skills, err := provider.Sub("skills")
```

- `New`快照输入 `fs.FS`，后续输入变更不会影响 provider；
- fingerprint由路径、节点类型和文件内容确定，provider ID不改变内容摘要；
- 只接受规范 ID、普通文件和合法 `fs.ValidPath`；
- `ReadFile`返回 detached byte slice；零值 provider fail closed；
- `MustNew`只适用于 build-time固定的 package-level embedded assets。

## Resolver

```go
resolver := assetfs.NewResolver()
_ = resolver.AddAll(parent, child)
resolver.Seal()
file, err := resolver.Open("assetfs://scene.demo/skills/demo/SKILL.md")
```

- `AddAll`对重复 ID、fingerprint冲突和 sealed状态原子失败；
- 相同 ID/fingerprint重复注册幂等；
- 最长 provider ID匹配，拒绝 query、fragment、反斜杠和路径逃逸；
- `Seal`幂等，seal后可并发读取；
- `ErrResolverSealed`与 `fs.PathError`可用 `errors.Is/As`判断。

本包不负责从源码 checkout发现资源、写磁盘、下载资源、执行脚本、加载 Scene
manifest或决定 host override。调用方仍负责 embedded source和显式 override政策。
