# Transport API（Experimental）

`Resolve` 把 Host 的 `Config` 与 `llm.RequestOptions` 合并为不可变的 `Settings`。请求级
header 覆盖默认 header，provider 已显式设置的 header 不被覆盖。`ApplyPayload` 只补充尚未
存在的 `session_id` 与 `extra_body.cache_control`；payload/response hook 由调用方负责并发安全。
