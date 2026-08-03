# Document Derived Field API（Experimental）

`document/pipeline/derive`根据`configs.DocSpec`中的derived表达式更新
`types.DocumentResult`。

唯一入口`EvaluateDerived(spec, result)`按spec字段顺序计算派生值，并把无法求值的情况写入
既有diagnostics合同。它不调用模型、不读取文件或网络，也不决定业务校验是否通过。

函数会原地修改传入result；调用方不得在同一result上并发执行。表达式语义依赖
`document/pipeline/expr`，当前为Experimental。
