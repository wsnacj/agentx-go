# AgentX Core Developer Portal工程

本目录只拥有站点主题、导航和首页呈现，不拥有AgentX API正文。

```bash
npm ci
npm run docs:audit
npm run docs:check
npm run docs:dev
```

- 正文来自`docs/`、根政策文件和各package `API.md`；
- package分类来自`docs/reference/developer-preview-packages.tsv`；
- `portal/.generated/`、缓存和静态构建产物均被忽略；
- Portal Candidate不是Public/Beta/Stable或production-ready发布。
- `docs:audit`需要访问npm registry；production build和46/9 coverage gate可在依赖已经
  安装后本地重复运行。
