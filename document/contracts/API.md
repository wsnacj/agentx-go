# Document Contracts API（Experimental）

`contracts` 只定义 OCR、PDF、pipeline 与 tool adapter 之间共享的 provider-neutral 数据，
不负责文件读取、网络、credential、artifact 持久化或业务字段解释。

- `BoundingBox`：页码、坐标、单位、坐标系和来源；
- `TextBlock`：有序文本观察、置信度与可选几何信息；
- `TableCell`：表格/合并单元格位置与文本；
- `ArtifactRef`：Host 已发布 artifact 的 display-safe 引用。

这些类型当前为 Experimental。客户 schema、年报字段、review 规则和真实存储 URI 不属于本包。
