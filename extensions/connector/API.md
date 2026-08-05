# Portable Connector Specification API（Experimental）

`extensions/connector`定义外部能力连接的credential-free声明。Connector是Host中的连接实例，
不是Tool、Plugin、MCP transport或第二套Runtime。

- `Spec`只包含identity、显示信息、protocol和transport；
- `Normalize`规范化ID并拒绝未知protocol/transport；
- `Project`生成display-safe `catalog.KindConnector`资产，不连接或授权；
- `Error`/`ErrorCode`支持`errors.Is/As`。

首版识别MCP以及stdio/Streamable HTTP两种声明，但P7-E2只实现Platform stdio。command、args、
endpoint、credential、proxy、retry、tenant、visibility和连接状态必须留在Host，不能塞进`Spec`。

本包不启动进程、不读取环境、不访问网络、不解析MCP，也不实例化Plugin/Expert/Team。当前为
Experimental，不构成Public/Beta/Stable承诺。
