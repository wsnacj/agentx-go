# OCR Cache API（Experimental）

`document/ocr/cache`定义OCR pipeline使用的最小缓存合同，并提供可选文件系统实现。

## 合同

- `Store.Get(ctx, key)`返回`Entry`、命中状态和错误；未命中不是错误；
- `Store.Set(ctx, key, entry)`写入payload；`Entry.Attributes`只属于调用方metadata；
- `Builder`从`config.CacheConfig`构造Store；`DefaultBuilder`在未启用缓存时返回noop store，
  `kind=fs`时使用文件缓存。

文件缓存会在显式`BaseDir`或系统临时目录下创建文件，支持TTL和最大容量裁剪；它不是durable
数据库，不保证跨进程事务。`Get`/`Set`检查context并在单实例内并发安全。生产Host应显式
配置可写目录、权限、保留期和敏感数据策略。

本包不发现credential、不访问网络，也不拥有业务cache key。当前为Experimental，不承诺
Public/Beta/Stable兼容性。
