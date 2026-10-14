# waner~ API站 文档站（docs-site）

独立 Astro + Starlight 文档站，内容改编自 SHUAI API 公开文档，按 [docs/superpowers/specs/2026-09-15-aiwanai-docs-site-design.md](../docs/superpowers/specs/2026-09-15-aiwanai-docs-site-design.md) 重品牌、重政策；已剔除本站未启用的 auto 模式、分组链、平滑迁移与可用模型表。

## 结构

- `astro.config.mjs` — Starlight 配置；双语 locale（`root` = 简体中文默认，`en` = English），侧边栏 5 组，标签通过 `translations` 按语言切换；`components` 注册四个组件覆盖
- `src/components/overrides/Hero.astro` — 首页 Hero 覆盖：点阵网格 + 径向辉光背景、玻璃徽章、渐变 CTA、可复制的终端代码窗口（文案随 locale 切换）
- `src/components/overrides/PageTitle.astro` — 文档页页头覆盖：渐变眉标 + 紧字距大标题 + 描述行
- `src/components/overrides/Header.astro` — 顶栏覆盖：在原生 SiteTitle / Search / ThemeSelect / LanguageSelect 之外加品牌色「控制台」CTA（新窗口打开 api.aiwanai.cc）
- `src/components/overrides/Sidebar.astro` + `AwSidebarSublist.astro` — 侧边栏覆盖：为 5 个一级分组注入 mask 图标（rocket / code / terminal / grid / lifebuoy，见 `GROUP_ICONS`），含当前页的分组标题点亮品牌色；分组图标的 mask 定义在 `custom.css`，新增分组时需同时补图标
- `src/styles/custom.css` — 全站视觉层：品牌渐变变量、顶栏毛玻璃 + 渐变发丝线、侧边栏分组图标 mask / 当前页高亮条 / 活跃分组点亮、h2 渐变装饰、表格 / 代码块 / 卡片 / 翻页样式（亮暗双主题），以及首页 `.aw-panel` / `.aw-cta` 组件样式
- `src/content/docs/` — 中文站：19 篇指南 + 首页 `index.mdx`（政策信息在 `quick-start.md`）
- `src/content/docs/en/` — 英文站：同构 19 篇 + `index.mdx`
- `src/starlight-schema.ts` — Starlight 内容集合 schema 的项目内副本（见下文「Node 26 说明」）
- `scripts/build-content.mjs` — 内容转换器：读取根目录 `shuaiapi-docs.md` → 英文指南输出到 `en/`

> 组件覆盖注意：Starlight 0.35 的包 exports 不含 `./constants` 与 `user-components/Icon.astro`，覆盖组件里把 `PAGE_TITLE_ID` 内联为 `'_top'`、行动按钮箭头用内联 SVG。升级 `@astrojs/starlight` 后需复核这两处。

## 命令

```bash
npm install --cache /tmp/npm-cache-docsite   # 本机 ~/.npm 为 root 所有，需换缓存目录
npm run content      # 重新生成 en/*.md（改完转换器后执行）
node scripts/build-content.mjs zh-page <slug>  # 单页重译见下文「中文页翻译」
npm run dev          # http://127.0.0.1:4321（中文）/ http://127.0.0.1:4321/en/（英文）
npm run build        # 产物在 dist/
npm run preview      # 预览生产构建
```

## 内容再生成

`scripts/build-content.mjs` 的处理链（顺序敏感）：

1. `applyEdits` — 按锚点行区间重建被抓平的表格、改写站点信息政策段、删除站点专属内容
2. `applyTokenRepairs` — 抓取拼接修复（约 200 条 curated 词典 + 通用正则：标点补空格、小写+大写拆分、ENV 变量拆分等；专有名词先掩码保护）
3. `applyDictSplit` — 字典驱动的通用二分拆分（散文里不在词典内、且能唯一二分为两个已知词的 token）
4. `applyInlineCode` — 散文里的裸代码术语（`web_search`、`response_format`、`/v1/...` 等）包反引号
5. `applyBrand` — 域名/品牌替换（SHUAI API → waner~ API站，4 个 shuai 域名 → 仅 `api.aiwanai.cc`）

转换器内置自检：未映射章节会直接抛错，因此任何 SECTION_MAP / 标题改动必须同步更新。

## 中文页翻译

`src/content/docs/*.md`（19 篇指南 + 首页）由英文版 `src/content/docs/en/*.md` 翻译而来，规则：

- 散文、标题、frontmatter title/description 译为中文；代码块、模型名、工具名、配置键、HTTP 路径保持原样
- 内部链接去掉 `/en/` 前缀（`/en/quick-start/` → `/quick-start/`）
- 重译单页：读 `en/<slug>.md` → 翻译 → 写 `src/content/docs/<slug>.md`，保持 markdown 结构逐行对应

修改英文页后需同步重译对应中文页，否则两语言会漂移。

## Node 26 说明

本机 Node 为 v26（无版本管理器，brew 不可用），拒绝为 `node_modules` 下的 `.ts` 做类型剥离，导致 `@astrojs/starlight/schema` 无法加载。后果是集合 schema 不生效 → frontmatter 默认值缺失 → 生产构建按 `draft === false` 过滤掉全部页面、dev 渲染 `data.head` 崩溃。

`src/starlight-schema.ts` 是 Starlight frontmatter schema 的自包含副本（IconSchema 的图标枚举放宽为 `z.string()`），`src/content.config.ts` 改为引用它。升级 `@astrojs/starlight` 后需同步此文件。

## 部署

静态站走 Cloudflare Workers Assets（无 SSR、无 D1/R2）。`astro.config.mjs` 的 `site` 为 `https://docs.inkwan.app`。

```bash
npm run build
npm run deploy          # wrangler deploy，Worker 名 aiwanai-docs
```

线上：

- 正式域名：https://docs.inkwan.app （中文默认）/ https://docs.inkwan.app/en/
- workers.dev：https://aiwanai-docs.yantianqi275.workers.dev

自定义域名通过 Workers domains API 绑到 `aiwanai-docs`（production），不要在 `wrangler.toml` 里写 zone `[[routes]]`：当前 API Token 不能写 Zone DNS / workers routes。凭证来自 `~/.pi/cloudflare.env`（`CLOUDFLARE_API_TOKEN`），不要提交。
