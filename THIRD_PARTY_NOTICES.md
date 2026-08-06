# Third-Party Notices

AgentX Go 自身以 Apache License 2.0 提供。本文档记录九个 library module 当前直接引用的第三方依赖，以及本仓库文档工具的直接依赖；它不是完整的软件物料清单，也不改变任何第三方项目的许可条款。Go 依赖未 vendored，使用者仍应以对应固定版本中携带的 LICENSE 和 NOTICE 为准。

| 依赖 | 固定版本 | 许可证 |
|---|---|---|
| `github.com/fsnotify/fsnotify` | `v1.9.0` | BSD-3-Clause |
| `gopkg.in/yaml.v3` | `v3.0.1` | MIT / Apache-2.0（按文件） |
| `codeberg.org/readeck/go-readability/v2` | `v2.1.1` | MIT |
| `github.com/PuerkitoBio/goquery` | `v1.10.2` | BSD-3-Clause |
| `github.com/sergi/go-diff` | `v1.4.0` | MIT |
| `golang.org/x/net` | `v0.56.0` | BSD-3-Clause |
| `github.com/agext/levenshtein` | `v1.2.3` | Apache-2.0 |
| `github.com/cenkalti/backoff/v5` | `v5.0.3` | MIT |
| `github.com/prometheus/client_golang` | `v1.22.0` | Apache-2.0 |
| `github.com/stretchr/testify` | `v1.10.0` | MIT |
| `go.uber.org/zap` | `v1.27.0` | MIT |
| `playwright` | `1.55.1` | Apache-2.0 |
| `vitepress` | `1.6.4` | MIT |

## 上游 NOTICE 归属

下列归属来自当前固定版本随附的 NOTICE 文件，保留在此便于二进制或派生发行物的维护者履行相应义务。

### github.com/agext/levenshtein v1.2.3

Alrux Go EXTensions (AGExt) - package levenshtein

Copyright 2016 ALRUX Inc.

This product includes software developed at ALRUX Inc.

(http://www.alrux.com/).

### github.com/prometheus/client_golang v1.22.0

Prometheus instrumentation library for Go applications

Copyright 2012-2015 The Prometheus Authors

This product includes software developed at SoundCloud Ltd. (http://soundcloud.com/).

该依赖自身还列出了随附组件及其许可；最终分发二进制时应重新依据实际依赖闭包生成并复核第三方归属，不能只复制本文件。

### gopkg.in/yaml.v3 v3.0.1

Copyright 2011-2016 Canonical Ltd.

该项目的文件按 MIT 或 Apache-2.0 授权；完整条款以固定版本携带的 LICENSE 与 NOTICE 为准。
