# Host HTTP 资源策略 API（Experimental）

`resourcepolicy` 强制“请求只能收窄、不能扩大 Host 权限与预算”。它不拥有具体
Scene 的路径、DSN、凭据或业务限额。

## 数值与布尔策略

- `NarrowPositiveInt`：`0` 继承 Host 上限，正值只能小于等于 Host 上限；
- `NarrowDurationMilliseconds`：避免溢出并拒绝超过 Host timeout 的值；
- `NarrowPermission`：请求可以关闭 Host 已允许能力，不能开启 Host 已禁止能力；
- `NarrowRequirement`：请求可以增加 Host requirement，不能取消已有 requirement。

非法扩张通过 `ErrBudgetNotAllowed`报告，可用 `errors.Is`识别。

## 路径与不透明值

`PathPolicy.Resolve`只接受 Host默认路径、显式 allowlist路径或显式 allowed root
下的路径，并在比较前规范化符号链接，阻止 symlink escape。

`ValuePolicy.Resolve`只接受 Host默认值或显式 allowlist值。它适合保护由 Host
拥有的 DSN或其它 opaque binding；调用方不得把 secret写入错误或诊断。

本包不读取 Scene配置，不决定 production安全策略，也不代表数据库或 filesystem
已经 ready。
