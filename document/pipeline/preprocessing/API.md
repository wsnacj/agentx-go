# Document Preprocessing API（Experimental）

`document/pipeline/preprocessing`提供页眉页脚清理的确定性机制和可选LLM接缝。

- `ParseCleanupMode`解析`none/programmatic/llm/auto`；
- `RemoveHeaderFooter`按`CleanupMode`处理页面文本；
- `LLMRequestFunc`是唯一模型port，Host显式拥有provider、credential、retry和网络；
- `HeaderFooterPattern`与`CleanupStats`描述识别模式和清理结果。

programmatic路径不产生外部副作用；LLM/auto路径只有在调用方提供request函数时才调用模型。
context cancellation/deadline会传给Host函数。输入页面不会被原地修改；结果规则仍为
Experimental，不应替代人工合规审阅。
