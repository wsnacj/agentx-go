# `extensions/domainmodule` 中文 API Reference

成熟度：**Experimental / Developer Preview foundation**。本页描述当前可验证合同，
不构成 Public、Beta、Stable或 semver兼容承诺。

## 适用范围

本包拥有编译期 Domain Module的可移植合同和顺序注册编排：manifest规范化、
module-scoped config、display-safe diagnostics、host resolver、重复ID拒绝、注册顺序
与 report聚合。它不拥有 Runner、Scene、pack registry、tool executor、provider、
credential、网络或产品策略。

## 最小接入

```go
report, err := domainmodule.RegisterAll(ctx, []domainmodule.Registration{{
    Manifest: domainmodule.Manifest{ID: "quotes", Tools: []string{"quote"}},
    Apply: func(ctx context.Context, manifest domainmodule.Manifest, cfg domainmodule.Config) (domainmodule.Diagnostics, error) {
        // 在 host adapter 内注册具体 executor；不要从本层读取环境密钥。
        return nil, nil
    },
}}, domainmodule.RegisterOptions{})
```

`Registration.Apply`按输入顺序串行调用。全部 manifest会先规范化并完成重复ID检查，
因此重复ID不会产生部分注册。`Preflight`随后验证 host-owned约束；config resolver在
任何 apply前运行。

## Config与resolver

- `Config.With`返回包含规范化 module ID的新 map，不修改原 map；
- `Config.Value/Has`优先按规范化ID查找，并保留旧 host原始key兼容；
- 显式 `Config`优先于 resolver；已有值不会再次解析；
- resolver只派生 typed config，不授予能力，也不应隐式读取 ambient secret；
- resolver失败返回 `resolve domain module <id> config`错误和
  `config_resolve_error`诊断。

## Diagnostics与错误

`Diagnostic`的 JSON字段为 `module_id`、`severity`、`code`、`message`和`details`。
code和严重级别会规范化，空 detail会被移除。`Report.Diagnostics()`保持 module与
回调产生诊断的顺序；`Report.HasErrors()`只检查 error severity。

注册不是事务：如果第二个 module失败，第一个 module已经产生的 host mutation不会
自动回滚，report仍保留此前的成功和失败诊断。调用方不得把 `RegisterAll`描述成
原子操作。需要回滚的产品必须由 host adapter提供显式补偿。

## 并发、取消与安全边界

单次 `RegisterAll`是同步、串行调用；本包不承诺同一 host target可被多个调用并发
修改。`context.Context`原样传给 resolver和 apply，具体取消响应取决于 host callback。
本包不会访问网络、文件系统、环境变量或凭据。

## Non-goal

- 动态 Go plugin或运行时下载；
- Scene业务逻辑、provider和credential管理；
- pack/tool/skill的具体安装策略；
- 原子注册、自动回滚或 durable lifecycle；
- 完整 A股 Scene或其它 extension发行。
