# Developer Portal生成器决策

状态：Developer Preview 文档工程。

## 结论

中文Developer Portal使用VitePress 1.6.4稳定线。版本通过npm registry的
`latest`标签核对并由`package-lock.json`固定；不使用2.0 alpha。VitePress 1.6.4的
传递Vite范围会解析到存在开发服务器安全公告的版本，因此lockfile显式覆盖到已修复的
Vite 6.4.3；该组合必须同时通过`npm audit`零漏洞和实际production build，否则不得接受。

选择依据：

- 现有正文全部是Markdown，VitePress直接把Markdown构建为静态HTML；
- 默认主题面向技术文档，内置代码高亮、响应式导航和本地全文搜索；
- 不需要React应用、CMS、远端搜索服务或运行时backend；
- 可把当前目录结构直接映射为页面，避免复制第二套API Reference；
- Node依赖只服务Portal构建，不进入九个Go library module或普通consumer依赖闭包。

官方能力依据见[VitePress简介](https://vitepress.dev/guide/what-is-vitepress)、
[本地搜索](https://vitepress.dev/reference/default-theme-search)和
[文件路由](https://vitepress.dev/guide/routing)。

## 有界对比

| 候选 | 当前适配结论 |
| --- | --- |
| VitePress | 采用；Markdown-native、单一静态站点、本地搜索，满足当前最小产品面 |
| Docusaurus | 不采用；版本化、多插件React站点能力当前不是必要条件，集成面更大 |
| Astro Starlight | 不采用；同样适合文档，但引入Astro层不能为现有Markdown/API gate增加必要收益 |

这不是对其它框架的通用优劣判断。若未来公共多版本站点需要复杂国际化、CMS或独立Web
应用能力，应创建新的替换ADR，不在现有生成器中顺手扩张。

## 单一正文源

`docs/*.md`、根政策文件、各package `API.md`、
`docs/reference/developer-preview-packages.tsv`与API snapshot继续拥有事实。
`scripts/prepare_docs_portal.go`只把这些文件投影到ignored的`portal/.generated/`，并从TSV
生成package导航、成熟度提示和source backlink。该目录和静态输出均不得提交。

## 运行边界

```bash
npm ci
npm run docs:check
npm run docs:dev
```

`docs:check`只构建本地静态站点，不访问provider、credential、业务网络或生产endpoint。
公共托管、域名、analytics、交互API控制台和发布授权不属于本生成器。
