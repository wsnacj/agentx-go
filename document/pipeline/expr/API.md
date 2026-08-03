# Document Expression API（Experimental）

`document/pipeline/expr`提供文档字段派生与validation使用的有限数字/布尔表达式求值。

- `EvalNumericExpr`支持`+ - * /`、括号和`abs/approx/between/round`；
- `EvalComparison`与`EvalBooleanExpr`支持比较以及`&&`、`||`组合；
- `LookupGlobal`从`types.DocumentResult`读取`chapter.field`数值；
- `NormalizeNumber`、`ParseNumber`处理常见货币、单位和数字字符串；
- `Parser`是低层递归下降解析器，普通调用方优先使用上述函数。

解析/求值失败返回`ok=false`或布尔false，不执行任意代码。该语言不是通用脚本引擎；客户
公式、精度和合规判定必须由Host审阅。函数无文件/网络副作用，当前为Experimental。
