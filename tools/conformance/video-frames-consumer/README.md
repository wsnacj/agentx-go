# Video Frames fixed-version consumer

该独立module固定依赖已推送的`github.com/wsnacj/agentx-go/tools` pseudo-version，不使用
`replace`，也不导入HS、Runner或Scene。它显式构造`tools/videoframes.LocalAdapter`，使用隔离的
ffprobe/ffmpeg stub验证注册、interval提取、媒体描述、artifact保留和取消语义。

stub用于避免测试依赖机器本地媒体工具，不代表真实provider或生产命令已经获准执行。新项目
仍须显式选择二进制、Root、authorization、approval、sandbox与工具暴露策略。
