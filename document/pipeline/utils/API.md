# Document Pipeline Utilities（Experimental）

`document/pipeline/utils`提供pipeline内部和Host adapter可复用的纯函数：

- `Flatten`按tree顺序展开`section.Node`；
- `TakePages`按页索引选择文本；
- `JoinAndClip`拼接并限制最大字符数；
- `UniqueKeepOrder`稳定去重；
- `Ternary`返回两个字符串之一。

函数不访问文件、网络或全局状态。页索引、字符裁剪和nil输入遵循当前实现，调用方不得把这些
helper当作安全边界或业务validation。当前为Experimental。
