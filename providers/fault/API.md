# Fault API（Experimental）

`Classify` 将 context、network、HTTP 和显式 `Wrap` 错误映射为稳定 `Kind`、HTTP status 与
retryability。它不替代面向终端用户的安全错误文案，也不决定产品重试策略。
