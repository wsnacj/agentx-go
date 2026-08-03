# Document Pipeline Config API（Experimental）

`document/pipeline/configs`定义文档解析spec的YAML合同、默认值和确定性推荐机制。

## 主要入口

- `DocSpec`聚合`MetaSpec`、`ChapterSpec`和`ValidationSpec`；
- `FieldSpec`与`ExtractorSpec`描述字段、候选提取、派生表达式和policy；
- `LoadSpec(path)`读取指定YAML文件，或目录下的`main.yaml`，并填充显式默认值；
- `RecommendSpecsForText`只根据调用方给出的文本和spec计算`SpecRecommendation`。

`LoadSpec`不会搜索仓库、Scene或环境变量；文件授权和客户schema由Host负责。spec内容可能
包含prompt或业务规则，因此不属于AgentX默认catalog。结构与默认值保持Experimental。
