# Python PDF Adapter（Experimental）

该 package 把显式 `PythonPath` 与 `ScriptPath` 适配为 `pdf.Runner`。调用方负责：

- 安装并选择兼容的 Python 与依赖；
- 决定是否使用 `BundledScriptPath`；
- 显式传入允许的环境变量和工作目录；
- 在 Host 层完成进程授权、审计、资源隔离和 artifact 策略。

`New` 只校验构造参数；`Run` 才启动进程。stderr 保留给上层的兼容错误投影，但 API 文档和
业务日志不应输出 credential。
