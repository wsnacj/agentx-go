# Web Tools fixed-version consumer

该隔离 module 固定 `github.com/wsnacj/agentx-go/tools` pseudo-version，不使用 `replace`，也不
导入 HS、Runner 或 Scene。它通过 fake `retrieval.Preparer` 注册并执行 `search`、`open_page`
和 `find_in_page`，证明 provider 协议、正文提取和共享页面缓存可以由外部 Host 组合。

fake port 不访问真实 provider、credential 或网络；该 consumer 是接入合同证据，不是发行或
线上安全认证。
