# `tools/videoframes` 中文 API Reference

成熟度：`Experimental extension`

`videoframes` 提供显式 opt-in 的本地视频帧适配器。它拥有 `video_frames` 的参数兼容、
ffprobe 探测、ffmpeg 帧提取、输入快照、artifact 生命周期、输出预算和媒体描述合同。

该包不会默认注册工具，也不负责安装 ffmpeg/ffprobe。授权、approval、sandbox、可执行文件
allowlist、依赖安装和是否向模型暴露该工具，始终由 Host 决定。

## 构造与注册

```go
type LocalOptions struct {
    Root          string
    MaxFrames     int
    MaxInputBytes int64
    FFmpegPath    string
    FFprobePath   string
}

type LocalAdapter struct { /* unexported fields */ }

func NewLocalAdapter(LocalOptions) *LocalAdapter
func (*LocalAdapter) Available() bool
func Register(tool.Registrar, *LocalAdapter)
func Definition() tool.Definition
```

构造 adapter 不执行命令、不创建目录，也不注册工具。`Register` 仅在 adapter 非空且两个命令
均可解析时注册 `video_frames`；缺少本地依赖时保持隐藏。

`Root` 是所有输入和输出的路径约束边界，不是 sandbox 或授权机制。模型不能覆盖 `root`、
`out_dir` 或 `output_dir`。每次成功执行会在
`.agentx/artifacts/video_frames/<source>-<unique>/` 下保留独立输出；失败输出由 adapter 清理，
既有成功 artifact 不会被后续调用删除。

## 结果合同

```go
type Result struct {
    Tool         string
    Action       string
    Status       string
    SourceVideo  string
    Strategy     string
    IntervalSec  float64
    OutputDir    string
    FrameCount   int
    FilesTouched []string
    Probe        *Probe
    Frames       []Frame
    Warning      string
}

type Probe struct {
    DurationSec float64
    Width       int
    Height      int
    FPS         float64
}

type Frame struct {
    FramePath    string
    Index        int
    TimestampSec float64
    Media        *mediaartifact.Descriptor
}

func FilesTouched([]Frame) []string
```

支持 `interval` 与 `keyframe` 两种策略。未识别的 strategy 兼容旧行为并回落到 `interval`；
默认间隔为 5 秒，默认最多 24 帧，硬上限为 120 帧，默认单输入上限为 512 MiB。

输入必须是 Root 内身份稳定的普通文件；symlink、FIFO 和超限输入分别返回可用
`errors.Is` 识别的 `ErrUnsafeFile` 或 `ErrFileTooLarge`。缺少 path、非法 JSON、模型尝试覆盖
Host-owned 路径时返回 canonical `runtime/toolerrors.ToolArgumentError`。

## 并发、取消与安全边界

`LocalAdapter` 构造后可并发调用。每个调用拥有唯一 artifact 目录；取消会传递给输入复制与
子进程。在 artifact 建立前已经取消的 context 不产生输出目录。命令 stdout/stderr 采用有界
捕获，错误只返回退出状态和预算摘要，不把原始 stderr 作为模型可见错误泄漏。

该 adapter 会读取本地普通文件、执行 Host 指定的 ffprobe/ffmpeg 并写入 Root 下的 artifact，
因此属于显式副作用能力。外部应用必须在注册前完成自己的授权、approval 与 sandbox 决策。
