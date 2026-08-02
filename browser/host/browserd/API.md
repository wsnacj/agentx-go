# Browserd Host API（Developer Preview）

`github.com/wsnacj/agentx-go/browser/host/browserd` 提供可选的 browserd 进程 Host。
它负责显式配置下的进程生命周期、内置 Node 资产、Playwright bootstrap/cache 和目录
ownership 校验，不负责产品路由、凭据发现、授权、代理、企业网络或客户策略。

## 最小构造

调用方必须提供：

- `Plan`：命令、endpoint、token 与 state/profiles/artifacts/logs 目录；
- `StatusProbe`：由调用方选择的 transport/backend 读取 daemon 状态；
- 可选 timeout：transport、health、probe interval 与 bootstrap deadline。

```go
manager, err := browserd.NewManager(browserd.ManagerOptions{
    Plan: browserd.Plan{
        Enabled: true,
        Command: "agentx-browserd",
        Endpoint: "http://127.0.0.1:43123",
        StateRoot: stateRoot,
        ProfilesRoot: profilesRoot,
        ArtifactsRoot: artifactsRoot,
        LogsRoot: logsRoot,
    },
    Probe: probe,
})
```

构造不会启动进程、下载依赖、创建目录或访问网络。`EnsureStarted(ctx)` 是显式副作用
边界；调用后可能进行 status probe、Node/Playwright bootstrap、目录创建和子进程启动。

## 生命周期合同

- `EnsureStarted(ctx)` 服从调用方 cancellation/deadline；没有显式 `StatusProbe` 时 fail closed；
- `Probe(ctx)` 返回 canonical `browser/runtime.BrowserProfileStatusResult`，并校验四类 ownership root；
- `Close()` 可重复调用，并终止本实例启动的进程；关闭后再次启动返回稳定错误；
- bundled bootstrap 对命令输出做有界捕获，错误信息不回显原始 stdout/stderr；
- `ManagedStarter` 为重复调用提供单实例协调，但不决定业务授权。

## 能力与非目标

`CapabilitiesForNodeBackendPlan` 只在显式 managed bundled plan 且调用方未声明更窄能力时，
给出 browserd 已验证的保守能力集合。cookie、credential、console、request capture 等高权限
能力不会被默认打开。

本包当前为 Developer Preview 候选，不是 Public/Beta/Stable 承诺，也不会自动安装浏览器、
读取环境凭据或选择真实网络策略。
