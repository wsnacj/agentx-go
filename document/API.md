# Document Module API（Experimental）

`github.com/wsnacj/agentx-go/document` 是 AgentX 的可选重型文档模块，目标覆盖 OCR、
deterministic document pipeline、PDF adapter 和推荐工具入口。当前处于 Experimental，
不是 Public/Beta/Stable 兼容承诺。

## 目录

- `contracts`：page、block、table、artifact 等 provider-neutral 数据合同；
- [`ocr`](ocr/API.md)：OCR split/cache/worker/diff/processor/pipeline 与显式 provider；
- [`pipeline`](pipeline/API.md)：spec、expression、derive、extract/preprocess 和 document orchestration；
- `pdf`：PDF Go mechanism 及显式 Python/native adapter；
- [`tools`](./tools/API.md)：AgentX document/PDF 推荐工具入口、显式 Host ports 与统一 PDF 协调。

## Host 边界

module 不拥有 credential 发现、默认 provider、客户 schema、商业 SDK license、文件授权、
网络授权或进程授权。真实 HTTP、LLM、artifact、Python/native 和客户脚本能力必须由 Host
显式构造；默认测试和 conformance consumer 不访问网络、不读取 credential、不启动进程。

## 当前状态

P3-B已经完成`contracts -> ocr -> pdf -> pipeline -> tools` source-authority、fixed consumer
与HS production cutover。P5-B又为全部可外部import的低层OCR/Pipeline package补齐中文
Experimental Reference；这不把它们升级为Developer Preview或稳定API。
