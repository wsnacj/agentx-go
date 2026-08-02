# Usage API（Experimental）

`Collector` 接收 `llm.UsageRecord`；`NoopCollector` 用于显式关闭记录。文件、数据库、计费、
遥测和租户归属均由 Host 实现，当前 module 不提供隐式持久化。
