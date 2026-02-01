import { defineConfig } from 'vitepress'

// Use /api-gateway/ base path for production (GitHub Pages), / for development
// In development, access at http://localhost:5173/
// In production, it will be deployed at https://go-zoox.github.io/api-gateway/
function getBase(): string {
  // Check if we're building for production
  // In dev mode, VitePress uses 'dev', in build mode it's 'production'
  // We can also check for GitHub Actions environment
  if (typeof process === 'undefined') {
    return '/'
  }
  try {
    const env = (process as any).env
    if (env && (env.NODE_ENV === 'production' || env.GITHUB_ACTIONS === 'true')) {
      return '/api-gateway/'
    }
  } catch (e) {
    // Ignore errors
  }
  return '/'
}

export default defineConfig({
  title: 'API Gateway',
  description: 'An Easy, Powerful, Flexible API Gateway',
  base: getBase(),
  
  locales: {
    root: {
      label: 'English',
      lang: 'en',
      title: 'API Gateway',
      description: 'An Easy, Powerful, Flexible API Gateway',
      themeConfig: {
        nav: [
          { text: 'Home', link: '/' },
          { text: 'Guide', link: '/guide/' },
          { text: 'API Reference', link: '/api/' },
          { text: 'Examples', link: '/guide/examples' },
          { text: 'TODO', link: '/TODO' },
          { text: 'GitHub', link: 'https://github.com/go-zoox/api-gateway' }
        ],
        sidebar: {
          '/guide/': [
            {
              text: 'Getting Started',
              items: [
                { text: 'Introduction', link: '/guide/' },
                { text: 'Quick Start', link: '/guide/getting-started' },
                { text: 'Installation', link: '/guide/installation' }
              ]
            },
            {
              text: 'Configuration',
              items: [
                { text: 'Configuration', link: '/guide/configuration' },
                { text: 'Routing', link: '/guide/routing' },
                { text: 'Load Balancing', link: '/guide/load-balancing' },
                { text: 'Health Check', link: '/guide/health-check' }
              ]
            },
            {
              text: 'Advanced',
              items: [
                { text: 'Plugins', link: '/guide/plugins' },
                { text: 'Examples', link: '/guide/examples' }
              ]
            },
            {
              text: 'Development',
              items: [
                { text: 'TODO List', link: '/TODO' }
              ]
            }
          ],
          '/api/': [
            {
              text: 'API Reference',
              items: [
                { text: 'Overview', link: '/api/' },
                { text: 'Config', link: '/api/config' },
                { text: 'Route', link: '/api/route' },
                { text: 'Plugin', link: '/api/plugin' }
              ]
            }
          ]
        }
      }
    },
    zh: {
      label: '中文',
      lang: 'zh-CN',
      title: 'API Gateway',
      description: '一个简单、强大、灵活的 API 网关',
      link: '/zh/',
      themeConfig: {
        nav: [
          { text: '首页', link: '/zh/' },
          { text: '指南', link: '/zh/guide/' },
          { text: 'API 参考', link: '/zh/api/' },
          { text: '示例', link: '/zh/guide/examples' },
          { text: 'TODO', link: '/TODO' },
          { text: 'GitHub', link: 'https://github.com/go-zoox/api-gateway' }
        ],
        sidebar: {
          '/zh/guide/': [
            {
              text: '快速开始',
              items: [
                { text: '介绍', link: '/zh/guide/' },
                { text: '快速开始', link: '/zh/guide/getting-started' },
                { text: '安装', link: '/zh/guide/installation' }
              ]
            },
            {
              text: '配置',
              items: [
                { text: '配置说明', link: '/zh/guide/configuration' },
                { text: '路由配置', link: '/zh/guide/routing' },
                { text: '负载均衡', link: '/zh/guide/load-balancing' },
                { text: '健康检查', link: '/zh/guide/health-check' }
              ]
            },
            {
              text: '高级',
              items: [
                { text: '插件系统', link: '/zh/guide/plugins' },
                { text: '使用示例', link: '/zh/guide/examples' }
              ]
            },
            {
              text: '开发',
              items: [
                { text: 'TODO 列表', link: '/zh/TODO' }
              ]
            }
          ],
          '/zh/api/': [
            {
              text: 'API 参考',
              items: [
                { text: '概览', link: '/zh/api/' },
                { text: '配置', link: '/zh/api/config' },
                { text: '路由', link: '/zh/api/route' },
                { text: '插件', link: '/zh/api/plugin' }
              ]
            }
          ]
        }
      }
    }
  },

  themeConfig: {
    search: {
      provider: 'local'
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/go-zoox/api-gateway' }
    ],
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024 GoZoox Team'
    }
  }
})
