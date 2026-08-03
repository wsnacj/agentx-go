# Process Runtime fixed-version consumer

该独立module固定依赖已推送的`github.com/wsnacj/agentx-go/tools` pseudo-version，不使用
`replace`，也不导入HS、Runner或Scene。它显式构造`tools/process.LocalAdapter`，验证真实本地
前台命令、bounded output和adapter-owned timeout。

该consumer只证明portable local adapter可被新项目直接组合。authorization、approval、
sandbox、production allowlist、signal policy和后台process lifecycle仍由Host负责。
