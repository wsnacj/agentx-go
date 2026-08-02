# 版本、升级与回滚

AgentX Go当前仍是private Developer Preview，没有tag、semver或Public/Beta/Stable兼容
承诺。为让调用方可以复现同一组合同，根、components、runtime和extensions四个module
使用同一个不可变pseudo-version：

```text
v0.0.0-20260802080954-21919fd8e06a
```

该版本是P1-A已验证的当前消费基线，不是正式发行版本。M5S/M5T历史回滚点仍保留在
changelog和成熟度记录中。

M6D另外使用`v0.0.0-m6d.0`在临时file proxy中验证四module同版发行列车。该版本不会
写入tracked `go.mod`、tag或下载入口，脚本退出后候选zip被删除；它不能作为项目依赖。
M6D manifest继续把上述fixed pseudo-version记录为回滚点。

## 为什么四个 Module 使用同一版本

四个module位于同一仓库，但Go会分别选择它们的版本。只升级其中一个，可能让调用方
继续组合不同历史checkpoint；代码未必立即失败，但API差分、问题复现和回滚会变得不
确定。建议一次显式指定四个版本，让Go的Minimal Version Selection选择同一commit：

```bash
export GOPRIVATE=github.com/wsnacj/agentx-go
export GONOSUMDB=github.com/wsnacj/agentx-go
export GOWORK=off

go get github.com/wsnacj/agentx-go@v0.0.0-20260802080954-21919fd8e06a \
  github.com/wsnacj/agentx-go/components@v0.0.0-20260802080954-21919fd8e06a \
  github.com/wsnacj/agentx-go/runtime@v0.0.0-20260802080954-21919fd8e06a \
  github.com/wsnacj/agentx-go/extensions@v0.0.0-20260802080954-21919fd8e06a
```

只直接import部分module的项目，可以只保留实际使用项；如果同时使用多个module，则必须
检查最终选择版本，而不能只看`go.mod`中某一行：

```bash
go list -m -f '{{.Path}} {{.Version}}' all | grep '^github.com/wsnacj/agentx-go'
```

## 升级检查顺序

1. 保存升级前的`go.mod`、`go.sum`和focused测试结果；
2. 同时指定四个候选版本，检查`go mod graph`和`git diff -- go.mod go.sum`；
3. 运行本项目直接使用package的测试、race和vet；
4. 对根Run/error/cancellation/Shutdown、LLM JSON/tool、Workflow状态/持久化顺序和
   extension公开DTO执行差分；
5. 再运行调用方完整回归，确认没有新增失败；
6. 记录最终`go list -m`输出和可回滚的提交。

如果调用方不能访问私有仓库，应先修复Git/GOPRIVATE配置。不要用长期本地`replace`
绕过访问或版本问题；那样只能证明本机源码可编译，不能证明其他项目可消费。

## API 差分边界

- 8个Developer Preview candidate有`go doc -all`可读snapshot和hash gate；升级前后应
  审阅签名变化，而不是默认“有文档”等于兼容；
- 其余Experimental package仍可能在后续Owner审阅中调整；新项目优先使用三条标准
  construction和成熟度矩阵中的推荐入口；
- type alias可以保持源码、字段和JSON兼容，但反射得到的定义package可能随source
  authority迁移而变化。依赖反射package identity的调用方必须单独验证；
- provider、credential、authorization、具体backend、Scene业务规则和真实网络不属于
  Core版本合同。

## 回滚

回滚必须回到升级前实际记录的四个版本，而不是猜测某个旧commit。执行与升级相同的
四项`go get module@version`，然后恢复对应`go.sum`并重新运行focused测试。若升级已
伴随调用方代码适配，应使用一个完整Git回滚提交，使代码与module版本同时恢复。

不要删除module cache、改写Git历史或使用本地`replace`伪造回滚成功。无法在旧版本上
恢复既有行为时，应保留失败证据并停止升级。

## 当前验证证据

根[`conformance/consumer`](../../conformance/consumer)直接固定四module统一版本，并
运行自定义ExecutionAdapter、Model/Tool Host Kit、Workflow Host Kit与A股extension
fixture路径。`scripts/check_cleanroom_consumer.go`会把该consumer复制到仓库外临时目录，
在`GOWORK=off`、`GOPROXY=off`、`-mod=readonly`下从module cache验证，不读取本仓源码。

这项证据只说明当前固定版本可独立消费；正式tag、license/NOTICE、security/legal、
release owner和生产SLA仍保持fail closed。

允许/禁止变更、四module未来tag前缀、版本epoch和维护责任见
[Developer Preview兼容与分发政策](developer-preview-policy.md)。
