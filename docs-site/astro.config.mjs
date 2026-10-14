import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
  site: 'https://docs.inkwan.app',
  integrations: [
    starlight({
      title: 'waner~ API站 文档',
      defaultLocale: 'root',
      locales: {
        root: { label: '简体中文', lang: 'zh-CN', dir: 'ltr' },
        en: { label: 'English', lang: 'en', dir: 'ltr' },
      },
      logo: {
        src: './src/assets/logo.svg',
        replacesTitle: false,
      },
      customCss: ['./src/styles/custom.css'],
      components: {
        Hero: './src/components/overrides/Hero.astro',
        PageTitle: './src/components/overrides/PageTitle.astro',
        Header: './src/components/overrides/Header.astro',
        Sidebar: './src/components/overrides/Sidebar.astro',
      },
      sidebar: [
        {
          label: '开始使用',
          translations: { en: 'Getting started' },
          items: [
            { label: '5 分钟快速开始', slug: 'quick-start', translations: { en: 'Quick Start' } },
            { label: '分组', slug: 'groups', translations: { en: 'Groups' } },
            {
              label: '平台导览：钱包 / 日志 / 安全', slug: 'platform-tour',
              translations: { en: 'Platform Tour: Wallet / Logs / Security' },
            },
          ],
        },
        {
          label: 'API 参考',
          translations: { en: 'API reference' },
          items: [
            { label: 'Endpoint 列表', slug: 'api-reference', translations: { en: 'API Reference' } },
            { label: 'OpenAI SDK 接入', slug: 'openai-sdk', translations: { en: 'OpenAI SDK Integration' } },
            { label: 'Anthropic SDK 接入', slug: 'anthropic-sdk', translations: { en: 'Anthropic SDK Integration' } },
            { label: '图片生成（URL 而非 base64）', slug: 'image-generation', translations: { en: 'Image Generation (URL, not base64)' } },
            { label: '图片 API 技能页（给 AI 助手）', slug: 'image-api-skill', translations: { en: 'Image API Skill Page (for AI assistants)' } },
            { label: 'Endpoint 与本地测速', slug: 'local-speed-test', translations: { en: 'Endpoints & Local Speed Test' } },
          ],
        },
        {
          label: '编码工具',
          translations: { en: 'Coding tools' },
          items: [
            { label: 'Claude Code', slug: 'claude-code' },
            { label: 'Codex CLI', slug: 'codex-cli' },
            { label: 'DeepSeek + OpenCode / ZCode', slug: 'deepseek-coding' },
            { label: 'CC Switch 一键导入', slug: 'cc-switch', translations: { en: 'CC Switch One-Click Import' } },
            { label: 'XiaoBaiSwitch 一键导入', slug: 'xiaobai-switch', translations: { en: 'XiaoBaiSwitch One-Click Import' } },
          ],
        },
        {
          label: '客户端与工具',
          translations: { en: 'Clients & tools' },
          items: [
            { label: 'Cherry Studio', slug: 'cherry-studio' },
            { label: 'AQBot', slug: 'aqbot' },
            { label: '沉浸式翻译', slug: 'immersive-translate', translations: { en: 'Immersive Translate' } },
          ],
        },
        {
          label: '排障',
          translations: { en: 'Troubleshooting' },
          items: [
            { label: '错误码与排障', slug: 'troubleshooting', translations: { en: 'Error Codes & Troubleshooting' } },
          ],
        },
      ],
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/yantianqi1/new-api',
        },
      ],
      editLink: {
        baseUrl: 'https://github.com/yantianqi1/new-api/edit/main/docs-site',
      },
    }),
  ],
});
