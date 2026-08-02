# AgentX Portable Scenes

`github.com/wsnacj/agentx-go/scenes`承载可移植的领域Kit：领域合同、Pack、Evaluator与
deterministic coordination。它可以组合AgentX Core和获准的optional modules，但不得反向
依赖HS、旧`scene/...`包、Runner、credential或具体生产backend。

当前所有package均为Experimental / Developer Preview candidate，不构成Public、Beta或
Stable兼容承诺。

## 当前目录

- [`astock`](./astock/API.md)：A股领域合同、三套Pack、evidence evaluator、工具目录与
  deterministic Host Kit；真实行情/研报provider、credential和产品策略由Host注入。
- `publicnews`、`companyresearch`、`docparse`：P4-A后续顺序迁入，目录只在首个真实
  implementation ready时创建。

## 依赖方向

```text
AgentX Core / optional mechanisms
              ↓
          scenes/*
              ↓
       Host adapter / Product
```

Core和optional mechanisms不得import本module；HS可以通过固定版本消费本module并保留真实
provider、安全策略、网络/文件副作用、部署与业务结果权威。
