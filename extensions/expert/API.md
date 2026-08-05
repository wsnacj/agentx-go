# Expert Specification API（Experimental）

`extensions/expert`定义可移植Expert角色资产，不选择模型、不创建Agent Session，也不执行Tool。

- `Spec`保存identity、展示信息、untrusted `Instructions`、Tags和显式capability requirements；
- `Parse`拒绝model、provider、credential、budget、session、sandbox等Host-owned字段；
- `Normalize`验证输入上限、规范ID并确定性排序requirements/Tags；
- `Project`只输出display-safe `catalog.KindExpert`资产，不泄漏instruction或requirements；
- `Error/ErrorCode`支持`errors.Is/As`且展示文本不包含原始诊断。

Requirement只允许Tool、Skill、Plugin和Connector，既不是授权，也不是路由或自动安装请求。可选
requirement失败时是否继续由Platform明确处理；必需资产必须fail closed。

Expert instructions是未经信任的内容。Host必须在授权、Prompt assembly和Session/Subagent实例化前
独立复核；本包不会将其自动注入system prompt、memory或catalog。

当前为Experimental，不构成Public/Beta/Stable或生产安全承诺。
