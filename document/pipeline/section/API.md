# Document Section API（Experimental）

`document/pipeline/section`定义pipeline与Host章节切分器之间的provider-neutral tree合同。

- `Node`描述section identity、标题、层级、页范围、正文与children；
- `SectionNode`是兼容alias，与`Node`保持同一类型identity。

本包只有数据结构，不执行切分、模型调用或文件访问。章节识别算法、客户规则和标题policy由
Host提供的`pipeline.Sectioner`拥有。当前为Experimental。
