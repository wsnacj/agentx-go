# MCP Client与Tool Adapter API（Experimental）

`extensions/mcp`实现当前稳定MCP `2025-11-25`的provider-neutral生命周期与Tool子集。具体
stdio/HTTP、进程、endpoint、credential、authorization和产品策略由Platform或业务Host拥有。

## 首版能力

- `Transport`：并发安全、context-aware、bounded message与幂等`Shutdown(ctx)`port；
- `Client`：`Initialize`、`DiscoverTools`、`CallTool`、`Snapshot`与`Shutdown`；
- `ToolSet`：实现现有`components/tool.DefinitionProvider`和`Executor`；
- `Request/Notification/Response/RPCError`：JSON-RPC 2.0 wire；
- `Tool`、`CallToolResult`与`ContentBlock`：MCP Tool声明和完整JSON结果；
- `Error/ErrorCode`：display-safe typed error，远端message不进入展示文本。

初始化固定遵循`initialize`后发送`notifications/initialized`。`DiscoverTools`有页数、Tool数量、
重复名称和cursor循环上限，并按名称稳定排序。Tool annotation、server instructions和execution
metadata均视为不可信描述，不自动进入prompt、授权或调度。

`ToolSet.Execute`只接受发现快照中的Tool和JSON object参数，并返回完整`CallToolResult` JSON。
MCP `isError=true`保持普通Tool结果，不转换成协议错误，确保模型能够观察并自我修正。

## 明确不包含

- stdio process或Streamable HTTP实现；
- OAuth、credential、proxy、retry或默认网络；
- Resources、Prompts、Sampling、Elicitation、Roots和Logging；
- experimental MCP Tasks及其与AgentX Task/Scheduler的映射；
- server-to-client requests、tool list hot reload或自动执行Server Instructions；
- Tool授权、approval或sandbox。

当前为Experimental，不构成完整MCP SDK、Public/Beta/Stable或production-ready承诺。
