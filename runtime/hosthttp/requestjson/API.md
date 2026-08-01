# Host HTTP JSON 请求 API（Experimental）

`requestjson` 提供 Host-deployed HTTP adapter 共用的严格 JSON 解码边界。它只
依赖 Go 标准库，不拥有 Scene DTO、认证、路由或业务错误投影。

## `Decode`

```go
func Decode(body io.Reader, maxBytes int64, target any) error
```

合同：

- 最多读取 `maxBytes + 1` 字节，以确定性方式拒绝超限 body；
- 拒绝未知字段和第一个 JSON value 后的任何额外数据；
- `nil`、空 body 和纯空白 body 保留 `target` 的零值；
- `maxBytes <= 0` 返回 `ErrInvalidLimit`；
- 可通过 `errors.Is` 识别 `ErrBodyTooLarge`、`ErrInvalidLimit` 与
  `ErrTrailingData`。

本包不负责 HTTP status、display-safe message、request identity、认证或
resource policy；这些责任分别属于 handler 和 shared host server。
