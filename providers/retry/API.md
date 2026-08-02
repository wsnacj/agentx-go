# Retry API（Experimental）

`Do` 按 `Options` 执行有限重试，遵守总 deadline、单次 attempt timeout、backoff 与 context
取消。默认分类使用 `fault.IsRetryable`。调用方仍须确保被重试操作满足自己的幂等边界。
