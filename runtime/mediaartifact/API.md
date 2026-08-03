# runtime/mediaartifact API

导入路径：

```go
import mediaartifact "github.com/wsnacj/agentx-go/runtime/mediaartifact"
```

成熟度：**Experimental / private validation**。

该 package 定义 browser、PDF、video、nodes 等 Runtime capability 输出共享的
媒体产物元数据，并提供无副作用的松散工具结果归一化。它不负责采集媒体、
生成路径、推断 MIME、持久化 artifact、维护 lineage 或执行 backend。

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
package 不提供 constructor 或 validation，调用方仍拥有字段赋值与业务语义。

## `Ref`

```go
type Ref struct {
    Raw            string   `json:"raw"`
    Display        string   `json:"display,omitempty"`
    Labels         []string `json:"labels,omitempty"`
    ModeHint       string   `json:"mode_hint,omitempty"`
    ArtifactSource string   `json:"artifact_source,omitempty"`
    ArtifactKind   string   `json:"artifact_kind,omitempty"`
}
```

`Ref` 是从工具结果中提取出的媒体输入引用。`Raw` 保存路径、URL 或 data URI；
其它字段只是后续分析与展示使用的提示，不构成访问授权、真实性或可用性保证。

## `RefsFromValue`

```go
func RefsFromValue(value any) ([]Ref, error)
```

`RefsFromValue` 接受以下输入：

- 单个字符串；字符串看起来像 JSON object/array 且能解码时继续递归解析；
- `[]string` 或 `[]any`；
- `map[string]any`，包括 browser、PDF、video、nodes 常见的嵌套
  `media`、`artifacts`、`rendered_pages`、`frames` 等字段。

函数保持子项顺序，裁剪字符串，复制 `Labels`，并根据显式字段或稳定的工具输出
约定补充 `ArtifactSource`、`ArtifactKind` 与 `ModeHint`。无法解码但以 `{` 或 `[`
开头的字符串仍按普通原始引用处理。其它顶层 Go 类型返回错误。

该归一化只处理内存中的数据结构，不访问本地文件、网络、provider、OCR 或模型，
也不负责去重、校验路径或执行 artifact 持久化。调用方仍须在 Host 边界实施路径、
网络、凭据和副作用策略。

## 非目标

- 不提供 artifact registry、持久化、lineage 或 runstore；
- 不生成文件、不访问网络、不调用 provider/backend；
- 不提供 image analysis、OCR、video frame extraction 或 ffmpeg adapter；
- 不定义媒体类型枚举或业务校验；
- 不构成 Public、Beta、Stable 或 production-ready 声明。
