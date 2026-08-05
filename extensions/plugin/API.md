# Portable Plugin Manifest API（Experimental）

`extensions/plugin`定义AgentX可安装能力包的中立manifest合同。Plugin是分发和装配单元，
不是Tool、Skill、Pack、Domain Module或第二套Agent Runtime。

## 合同

- `Manifest`：名称、schema版本、版本、描述、trust boundary、contained roots和entrypoints；
- `Dependency`：只描述对其它Plugin或Connector的依赖请求，不负责解析或安装；
- `PermissionRequest`：只描述能力请求，不表示Host已经授权；
- `Parse`：从有界JSON解析并拒绝policy、approval、sandbox、credential和secret等Host字段；
- `Normalize`：规范化identity、路径、依赖和权限请求，不修改调用方切片；
- `Error`/`ErrorCode`：支持`errors.Is/As`的display-safe错误合同。

```go
manifest, err := plugin.Parse(content)
if err != nil {
    return err
}
for _, request := range manifest.RequestedPermissions {
    // Host独立评估；不能把request直接当作grant。
    _ = request
}
```

entrypoint必须是plugin root内的相对目录，并落在`Roots`声明范围内。manifest不得保存凭据、
租户身份、授权结果、tool handler、backend句柄或运行时状态。

## Owner边界

- 本包拥有portable manifest、规范化和typed error；
- `components/tool`拥有Tool schema、Handler与Executor；
- `extensions/skills`拥有Skill加载与可移植语义；
- Platform/业务Host拥有安装、完整性检查、权限决策、activation、credential和进程隔离；
- Workflow/Pack继续拥有业务闭环，Plugin不得替代它们。

本包不扫描默认目录、不复制文件、不运行command/hook、不读取环境、不访问网络，也不实例化
Connector、Expert或Team。当前为Experimental，不构成Public/Beta/Stable或安全沙箱承诺。
