# runtime/mediaartifact API

导入路径：

```go
import mediaartifact "github.com/wsnacj/agentx-go/runtime/mediaartifact"
```

成熟度：**Experimental / private validation**。

该 package 定义 browser、PDF、video、nodes 等 Runtime capability 输出共享的
媒体产物元数据。它只提供数据合同，不负责采集媒体、生成路径、推断 MIME、
持久化 artifact、维护 lineage 或执行 backend。

## `Descriptor`

```go
type Descriptor struct {
    Source        string `json:"source"`
    Kind          string `json:"kind"`
    Path          string `json:"path,omitempty"`
    URL           string `json:"url,omitempty"`
    MIMEType      string `json:"mime_type,omitempty"`
    Format        string `json:"format,omitempty"`
    Bytes         int64  `json:"bytes,omitempty"`
    Width         int    `json:"width,omitempty"`
    Height        int    `json:"height,omitempty"`
    DurationMs    int64  `json:"duration_ms,omitempty"`
    TimestampMs   int64  `json:"timestamp_ms,omitempty"`
    FPS           int64  `json:"fps,omitempty"`
    HasAudio      *bool  `json:"has_audio,omitempty"`
    ScreenIndex   int64  `json:"screen_index,omitempty"`
    CaptureScope  string `json:"capture_scope,omitempty"`
    CaptureWidth  int    `json:"capture_width,omitempty"`
    CaptureHeight int    `json:"capture_height,omitempty"`
    Facing        string `json:"facing,omitempty"`
    Index         int    `json:"index,omitempty"`
    CreatedAt     string `json:"created_at,omitempty"`
}
```

### 字段约定

- `Source`：产物来源，例如 `browser`、`pdf`、`video` 或 `nodes`。
- `Kind`：调用方定义的媒体产物类型，例如 `screenshot`、`rendered_page`。
- `Path`、`URL`：产物的本地路径或来源 URL；本 package 不验证可访问性。
- `MIMEType`、`Format`：调用方推断或声明的格式信息。
- `Bytes`、尺寸、时长、时间戳、FPS、屏幕和摄像头字段：可选媒体元数据。
- `HasAudio` 使用指针区分“未知”“明确无音频”“明确有音频”。
- `CreatedAt` 保持字符串合同；本 package 不解析或强制时间格式。

`Source` 与 `Kind` 在 JSON 中始终出现；其它零值字段使用 `omitempty`。该
package 不提供 constructor、normalization 或 validation，调用方仍拥有字段
赋值与业务语义。

## 非目标

- 不提供 artifact registry、持久化、lineage 或 runstore；
- 不生成文件、不访问网络、不调用 provider/backend；
- 不定义媒体类型枚举或业务校验；
- 不构成 Public、Beta、Stable 或 production-ready 声明。
