# 版本、升级与回滚

AgentX Go当前仍是private Developer Preview，没有tag、semver或Public/Beta/Stable兼容
承诺。当前九个module按各自最后一次真实source-authority checkpoint使用不可变
pseudo-version；准确矩阵以机器可读
[`developer-preview-module-versions.txt`](../reference/developer-preview-module-versions.txt)为
事实源，并投影到[安装与多Module引用](installation-and-modules.md#当前固定验证版本)。调用方不得假设它们
来自同一个commit，也不得把任意单个pseudo-version推广为整个仓库的版本。

根、runtime与extensions当前仍使用`v0.0.0-20260802113655-f41de95ec5be`；components
已经因后续LLM/tool合同收口使用`v0.0.0-20260802130858-34ec103e09d9`。这两个值都是
当前消费基线的一部分，不是正式发行版本。providers、tools、browser、document和scenes
也各自使用安装指南记录的固定版本。M5S/M5T历史回滚点仍保留在changelog和成熟度记录中。

Pre-Beta技术门禁使用`v0.0.0-m6d.0`在临时file proxy中验证九module同版发行列车。该版本不会
写入tracked `go.mod`、tag或下载入口，脚本退出后候选zip被删除；它不能作为项目依赖。
manifest继续把当前root fixed pseudo-version记录为回滚点。

正式Beta建议使用`v0.1.0-beta.1`，九个module分别使用Go多module约定的目录tag前缀；
该建议尚未获得发行授权，完整矩阵与fail-closed责任见
[Pre-Beta准入合同](../reference/pre-beta-admission.md)。

## 历史同版列车与当前独立版本

M5T/M6D曾让根、components、runtime和extensions四个Core module使用同一版本，以验证
同版发行列车。后续迁移已经让components和五个可选module前进到不同checkpoint，因此
“四个Core module必须同commit”不再是当前事实。升级时应一次显式指定项目直接使用的
全部AgentX module，让Go的Minimal Version Selection得到可复现组合：

```bash
export GOPRIVATE=github.com/wsnacj/agentx-go
export GONOSUMDB=github.com/wsnacj/agentx-go
export GOWORK=off

go get github.com/wsnacj/agentx-go@v0.0.0-20260802113655-f41de95ec5be \
  github.com/wsnacj/agentx-go/components@v0.0.0-20260802130858-34ec103e09d9 \
  github.com/wsnacj/agentx-go/runtime@v0.0.0-20260802113655-f41de95ec5be \
  github.com/wsnacj/agentx-go/extensions@v0.0.0-20260802113655-f41de95ec5be
```

只直接import部分module的项目，可以只保留实际使用项；如果同时使用多个module，则必须
检查最终选择版本，而不能只看`go.mod`中某一行：

```bash
go list -m -f '{{.Path}} {{.Version}}' all | grep '^github.com/wsnacj/agentx-go'
```

任何module都不应仅为“看起来同版”强行回退到旧commit。升级任一module时，应同时检查
其`go.mod`最终选择的Core/peer module版本，并按安装指南中的九module矩阵记录实际回滚点。

## 升级检查顺序

1. 保存升级前的`go.mod`、`go.sum`和focused测试结果；
2. 显式指定直接使用的全部AgentX module，并检查`go mod graph`和
   `git diff -- go.mod go.sum`；
3. 运行本项目直接使用package的测试、race和vet；
4. 对根Run/error/cancellation/Shutdown、LLM JSON/tool、Workflow状态/持久化顺序和
   extension公开DTO执行差分；
5. 再运行调用方完整回归，确认没有新增失败；
6. 记录最终`go list -m`输出和可回滚的提交。

如果调用方不能访问私有仓库，应先修复Git/GOPRIVATE配置。不要用长期本地`replace`
绕过访问或版本问题；那样只能证明本机源码可编译，不能证明其他项目可消费。

## API 差分边界

- 14个Developer Preview candidate有`go doc -all`可读snapshot和hash gate；升级前后应
  审阅签名变化，而不是默认“有文档”等于兼容；
- 其余Experimental package仍可能在后续Owner审阅中调整；新项目优先使用七类能力矩阵的
  construction和成熟度矩阵中的推荐入口；
- type alias可以保持源码、字段和JSON兼容，但反射得到的定义package可能随source
  authority迁移而变化。依赖反射package identity的调用方必须单独验证；
- provider、credential、authorization、具体backend、Scene业务规则和真实网络不属于
  Core版本合同。

## 回滚

回滚必须回到升级前实际记录的module版本集合，而不是猜测某个旧commit。执行与升级
相同的`go get module@version`，然后恢复对应`go.sum`并重新运行focused测试。若升级已
伴随调用方代码适配，应使用一个完整Git回滚提交，使代码与module版本同时恢复。

不要删除module cache、改写Git历史或使用本地`replace`伪造回滚成功。无法在旧版本上
恢复既有行为时，应保留失败证据并停止升级。

## 当前验证证据

`scripts/check_developer_preview_version.go`会核对九module矩阵、代表性fixed-version
consumer和安装文档；历史四module脚本不再被当作当前九module版本证据。

根[`conformance/consumer`](../../conformance/consumer)直接固定Core四module及scenes版本，并
运行自定义ExecutionAdapter、Model/Tool Host Kit、Workflow Host Kit与A股extension
fixture路径。`scripts/check_cleanroom_consumer.go`会把该consumer复制到仓库外临时目录，
在`GOWORK=off`、`GOPROXY=off`、`-mod=readonly`下从module cache验证，不读取本仓源码。

Providers、Tools、Browser、Document与其它Scene分别由各module的`conformance`
consumer验证，不通过根consumer伪装成单一“全功能开箱即用”入口。示例和consumer映射见
[`examples/README.md`](../../examples/README.md)。这些证据只说明当前固定版本可独立消费；
正式tag、license/NOTICE、security/legal、
release owner和生产SLA仍保持fail closed。

允许/禁止变更、九module未来tag前缀、版本epoch和维护责任见
[Developer Preview兼容与分发政策](developer-preview-policy.md)。
