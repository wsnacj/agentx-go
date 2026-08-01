# ProductShell Observation / Host Handoff 固定版本 Consumer

这个隔离 consumer只依赖已推送的 `agentx-go/extensions`固定 pseudo-version，不使用
HS、Runner、Scene、长期 `replace`、provider、网络、凭据或生产副作用。

示例由宿主显式提供 typed session与host process事实，并验证以下完整路径：

1. `BuildSessionObservation`归一化session事件、分支和compaction摘要；
2. `BuildHostProcessProgressObservation`归一化Host拥有的process view；
3. Host从typed事实构造 `HostDiagnosticOperatorLineObservation`；
4. canonical helper只把operator line投影为display-safe handoff envelope；
5. 独立consumer检查conformance并记录runtime-use evidence；
6. 最终log字段由Host adapter渲染和交付。

它不读取或解析raw Host diagnostics、tool output、RunStore或transcript backend，也不拥有
ObservationSnapshot、process inventory、readback、真实日志/HTTP/UI delivery。它只证明
外部module能够消费Experimental observation/handoff合同，不构成正式发行或兼容承诺。

```bash
GOWORK=off GOPROXY=off go test ./...
GOWORK=off GOPROXY=off go run .
```
