import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vitepress'

const configDir = path.dirname(fileURLToPath(import.meta.url))
const navigationPath = path.resolve(configDir, '../.generated/navigation.json')

if (!fs.existsSync(navigationPath)) {
  throw new Error('missing portal/.generated/navigation.json; run npm run docs:prepare first')
}

const navigation = JSON.parse(fs.readFileSync(navigationPath, 'utf8'))

function packageItems(prefix) {
  return navigation.packages
    .filter((item) => item.route.startsWith(prefix))
    .map((item) => ({
      text: `${item.package} · ${item.badge}`,
      link: item.route,
    }))
}

export default defineConfig({
  lang: 'zh-CN',
  title: 'AgentX Go',
  description: 'AgentX Go 中文 Developer Preview API 与接入文档',
  srcDir: '.generated',
  cleanUrls: true,
  lastUpdated: true,
  ignoreDeadLinks: false,
  markdown: {
    lineNumbers: true,
  },
  themeConfig: {
    logo: false,
    siteTitle: 'AgentX Go',
    search: {
      provider: 'local',
      options: {
        translations: {
          button: { buttonText: '搜索文档', buttonAriaLabel: '搜索文档' },
          modal: {
            noResultsText: '没有找到相关结果',
            resetButtonTitle: '清除查询',
            footer: { selectText: '选择', navigateText: '切换', closeText: '关闭' },
          },
        },
      },
    },
    nav: [
      { text: '开始', link: '/docs/quickstart' },
      { text: '接入指南', link: '/docs/guides/capability-map' },
      { text: 'Package API', link: '/packages' },
      { text: '成熟度', link: '/docs/maturity' },
      { text: '分发边界', link: '/docs/reference/distribution-readiness' },
    ],
    sidebar: {
      '/docs/': [
        {
          text: '开始',
          items: [
            { text: '文档导航', link: '/docs/' },
            { text: 'Quickstart', link: '/docs/quickstart' },
            { text: '七类能力矩阵', link: '/docs/guides/capability-map' },
            { text: '执行模型', link: '/docs/concepts/execution-model' },
          ],
        },
        {
          text: '标准构造与控制路径',
          items: [
            { text: '自定义 ExecutionAdapter', link: '/docs/guides/custom-adapter' },
            { text: 'Model / Tool Host Kit', link: '/docs/guides/model-tool-hostkit' },
            { text: 'Workflow Host Kit', link: '/docs/guides/workflow-hostkit' },
            { text: 'Objective Host Kit', link: '/docs/guides/objective-hostkit' },
            { text: 'Session / Subagent Host Kit', link: '/docs/guides/session-subagent-hostkit' },
          ],
        },
        {
          text: '合同与运维边界',
          items: [
            { text: '生命周期与错误', link: '/docs/guides/lifecycle-and-errors' },
            { text: '安装与多 Module', link: '/docs/guides/installation-and-modules' },
            { text: '版本、升级与回滚', link: '/docs/guides/versioning-and-upgrades' },
            { text: 'Package 成熟度', link: '/docs/reference/package-maturity' },
            { text: 'Distribution Readiness', link: '/docs/reference/distribution-readiness' },
          ],
        },
      ],
      '/components/': [{ text: 'Components API', items: packageItems('/components/') }],
      '/runtime/workflow/': [{ text: 'Workflow API', items: packageItems('/runtime/workflow/') }],
      '/runtime/': [{ text: 'Runtime API', items: packageItems('/runtime/') }],
      '/extensions/': [{ text: 'Extensions API', items: packageItems('/extensions/') }],
      '/providers/': [{ text: 'Providers API', items: [{ text: 'Providers 总览', link: '/providers/API' }] }],
      '/tools/': [{ text: 'Tools API', items: [{ text: 'Tools 总览', link: '/tools/API' }] }],
      '/browser/': [{ text: 'Browser API', items: [{ text: 'Browser 总览', link: '/browser/API' }] }],
      '/document/': [{ text: 'Document API', items: [{ text: 'Document 总览', link: '/document/API' }] }],
      '/scenes/': [{ text: 'Portable Scenes API', items: [{ text: '成熟度矩阵', link: '/docs/reference/package-maturity' }] }],
      '/packages': [
        {
          text: 'Package Reference',
          items: [
            { text: '全部 Package', link: '/packages' },
            { text: '成熟度规则', link: '/docs/reference/package-maturity' },
          ],
        },
      ],
    },
    outline: { level: [2, 3], label: '本页内容' },
    docFooter: { prev: '上一页', next: '下一页' },
    lastUpdated: { text: '最后更新' },
    darkModeSwitchLabel: '外观',
    sidebarMenuLabel: '目录',
    returnToTopLabel: '返回顶部',
    socialLinks: [{ icon: 'github', link: 'https://github.com/wsnacj/agentx-go' }],
    footer: {
      message: 'Private Developer Preview · 不构成 Public/Beta/Stable 或生产 SLA',
      copyright: 'AgentX Go',
    },
  },
})
