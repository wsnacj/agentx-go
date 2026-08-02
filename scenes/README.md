# AgentX Portable Scenes

`github.com/wsnacj/agentx-go/scenes`承载可移植的领域Kit：领域合同、Pack、Evaluator与
deterministic coordination。它可以组合AgentX Core和获准的optional modules，但不得反向
依赖HS、旧`scene/...`包、Runner、credential或具体生产backend。

当前所有package均为Experimental / Developer Preview candidate，不构成Public、Beta或
Stable兼容承诺。

## 当前目录

- [`astock`](./astock/API.md)：A股领域合同、三套Pack、evidence evaluator、工具目录与
  deterministic Host Kit；真实行情/研报provider、credential和产品策略由Host注入。
- [`publicnews`](./publicnews/API.md)：公开新闻意图、证据、来源质量、回答边界、Pack与
  无provider Host Kit；搜索、页面读取和站点策略由Host注入。
- [`companyresearch`](./companyresearch/API.md)：公司研究合同、任务分解、证据guard、
  Pack与无provider Host Kit；财报、行情、新闻和主体解析backend由Host注入。
- [`docparse`](./docparse/API.md)：文档 profile/planner/adapter/fusion/understanding、
  evidence evaluator、Pack与无文件/无provider Host Kit；OCR/PDF、私有schema和真实文件由
  Host显式注入。
- [`browserops`](./browserops/API.md)：浏览器操作Pack、证据合同、确定性evaluator与
  [Host Kit](./browserops/hostkit/API.md)；真实浏览器、profile/login、credential、审批、
  文件/artifact与站点副作用策略由Host显式注入。

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
