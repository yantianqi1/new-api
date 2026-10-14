/**
 * waner~ API站 文档站内容构建脚本
 *
 * 把抓取的 SHUAI API 接入文档（shuaiapi-docs.md，new-api 同构站点的标准接入文档）
 * 转换为 Starlight 内容目录下的单页 Markdown：
 *   1. 按 ## 拆分章节
 *   2. 重建被抓取抹平的表格
 *   3. 修复抓取导致的丢空格 artifact
 *   4. 品牌 / 域名换血（SHUAI API → waner~ API站，shuai 域名 → api.aiwanai.cc）
 *   5. 删除站点特有内容（QQ 群、公告、shuai 模型表）
 *
 * 用法：node scripts/build-content.mjs
 */

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const SITE = path.resolve(__dirname, '..');
const REPO = path.resolve(SITE, '..');
const SRC = path.join(REPO, 'shuaiapi-docs.md');
const OUT = path.join(SITE, 'src', 'content', 'docs');

// ---------------------------------------------------------------------------
// 章节 → 页面映射（slug / 侧边栏标题，中英双语）
// ---------------------------------------------------------------------------
const SECTION_MAP = {
  '站点信息与 API 地址': { mergeInto: 'quick-start' },
  'API 接入参考（Endpoints）': { slug: 'api-reference', title: { zh: 'Endpoint 列表', en: 'API Reference' } },
  '快速开始：5 分钟完成第一个请求': { slug: 'quick-start', title: { zh: '5 分钟快速开始', en: 'Quick Start' } },
  '图片生成：返回 URL 而非 base64': { slug: 'image-generation', title: { zh: '图片生成：返回 URL 而非 base64', en: 'Image Generation (URL, not base64)' } },
  '图片 API 技能页（给 AI 助手）': { slug: 'image-api-skill', title: { zh: '图片 API 技能页（给 AI 助手）', en: 'Image API Skill Page (for AI assistants)' } },
  // 本站未启用 auto 模式与用户自定义分组链（auto_groups 仅 default，new-api 无链管理功能）
  // → 页面只讲分组，标题去掉「自动模式」
  '分组、auto 模式与分组链': { slug: 'groups', title: { zh: '分组', en: 'Groups' } },
  'Endpoint 与本地测速': { slug: 'local-speed-test', title: { zh: 'Endpoint 与本地测速', en: 'Endpoints & Local Speed Test' } },
  '支持中心：工单与回复': { slug: 'support', title: { zh: '支持与工单', en: 'Support & Tickets' } },
  '平台导览：钱包、日志与安全': { slug: 'platform-tour', title: { zh: '平台导览：钱包 / 日志 / 安全', en: 'Platform Tour: Wallet / Logs / Security' } },
  'XiaoBaiSwitch 一键导入': { slug: 'xiaobai-switch', title: { zh: 'XiaoBaiSwitch 一键导入', en: 'XiaoBaiSwitch One-Click Import' } },
  'CC Switch 一键导入': { slug: 'cc-switch', title: { zh: 'CC Switch 一键导入', en: 'CC Switch One-Click Import' } },
  '配合 Claude Code 使用': { slug: 'claude-code', title: { zh: 'Claude Code', en: 'Claude Code' } },
  '配合 Codex CLI 使用': { slug: 'codex-cli', title: { zh: 'Codex CLI', en: 'Codex CLI' } },
  'DeepSeek + OpenCode / ZCode': { slug: 'deepseek-coding', title: { zh: 'DeepSeek + OpenCode / ZCode', en: 'DeepSeek + OpenCode / ZCode' } },
  '配合 AQBot 使用': { slug: 'aqbot', title: { zh: 'AQBot', en: 'AQBot' } },
  '配合 Cherry Studio 使用': { slug: 'cherry-studio', title: { zh: 'Cherry Studio', en: 'Cherry Studio' } },
  '配合沉浸式翻译使用': { slug: 'immersive-translate', title: { zh: '沉浸式翻译', en: 'Immersive Translate' } },
  'OpenAI SDK 接入': { slug: 'openai-sdk', title: { zh: 'OpenAI SDK 接入', en: 'OpenAI SDK Integration' } },
  'Anthropic SDK 接入': { slug: 'anthropic-sdk', title: { zh: 'Anthropic SDK 接入', en: 'Anthropic SDK Integration' } },
  '错误码与排障': { slug: 'troubleshooting', title: { zh: '错误码与排障', en: 'Error Codes & Troubleshooting' } },
  // 丢弃：站点特有内容
  '目录': { drop: true },
  '站点公告摘要': { drop: true },
  '可用模型与计费': { drop: true },
  'Enabled Models': { drop: true, note: '旧模型表残留子标题' },
};

// ---------------------------------------------------------------------------
// 逐行外科手术：{start, end} 为定位用唯一子串（行内包含即可），替换该区间所有行
// ---------------------------------------------------------------------------
const T = (s) => s; // 仅为可读性

const EDITS = [
  // ---- 站点信息：4 线路表 + 站点特有注意事项 → 单地址 + waner~ API站 真实政策 ----
  {
    start: '**API 地址（4 条线路）：**',
    end: '售后 QQ 群：1040560829',
    replace: T(`| Address | Notes |
| --- | --- |
| \`https://api.aiwanai.cc\` | The single site-wide endpoint, usable globally — no need to pick a line |

**Authentication:**

\`\`\`http
Authorization: Bearer <token>
\`\`\`

Create an API token in the console. Never hand your console session cookie to third-party AI tools.

**Good to know:**
- New accounts get a **¥1 bonus** (500,000 quota) on signup to try the models
- Daily check-in is worth about **¥0.9**, and both sides get **¥1** on invitations; bonus quota is limited to \`[free]\`-prefixed models and \`gpt-5.6-sol\`
- Top-ups via **Alipay (YiPay)** at 10 / 20 / 50 / 100 / 200 / 500 / 1000
- For live model and pricing details, see the Pricing page in the console`),
  },
  // ---- API 接入参考：清理抓取碎片，重写为端点说明 ----
  {
    start: 'Back to docsEndpoints',
    end: '具体可用模型取决于已启用的渠道、用户分组、令牌限制与计费配置。',
    replace: T(`The waner~ API站 API is compatible with the official OpenAI / Anthropic / Gemini protocols and uses Bearer tokens — one \`sk-\` key works across every endpoint.

### Main endpoints

- **\`POST /v1/chat/completions\`** — generates a model reply from a conversation history; OpenAI Chat Completions compatible, with streaming when the model supports it. This is the primary endpoint for OpenAI-compatible clients; streamed responses use SSE, so the client must handle incremental chunks and the final usage; when forwarding, the backend preserves explicit zero and \`false\` values the client sends.
- **\`POST /v1/responses\`** — the OpenAI Responses protocol route; \`/v1/responses/compact\` is its slimmed-down variant.
- **\`POST /v1/messages\`** — the Anthropic-compatible message route.
- **\`POST /v1beta/models/{model}:generateContent\`** — the Gemini-compatible route.
- **\`POST /v1/images/generations\`** — image generation; always returns an image URL rather than base64, see [Image generation](/en/image-generation/).
- **\`POST /v1/embeddings\`** / **\`POST /v1/rerank\`** — embeddings and reranking.
- **\`GET /v1/models\`** — the list of currently available models.

**Full endpoint list:**

| Type | Method | Path |
| --- | --- | --- |
| models | GET | \`/v1/models\` |
| anthropic | POST | \`/v1/messages\` |
| embeddings | POST | \`/v1/embeddings\` |
| gemini | POST | \`/v1beta/models/{model}:generateContent\` |
| image-generation | POST | \`/v1/images/generations\` |
| jina-rerank | POST | \`/v1/rerank\` |
| openai | POST | \`/v1/chat/completions\` |
| openai-response | POST | \`/v1/responses\` |
| openai-response-compact | POST | \`/v1/responses/compact\` |

**Compatibility notes:**
- \`/v1/chat/completions\` follows OpenAI-compatible chat-completion semantics.
- \`/v1/responses\` follows OpenAI-compatible responses semantics where configured.
- \`/v1/messages\` provides the Anthropic-compatible message route.
- \`/v1beta/models/{model}:generateContent\` provides the Gemini-compatible route.
- The models actually available to you depend on the enabled channels, your group, token restrictions and billing configuration.`),
  },
  // ---- 快速开始：碎句修复 ----
  {
    start: 'Go from sign-up to your first successful call',
    end: 'APIs while the site relays your requests to the upstream models.',
    replace: T(`Go from sign-up to your first successful call in five minutes. waner~ API站 is an AI API gateway: you create an \`sk-\` token here, then call it exactly like the official OpenAI / Anthropic APIs while the site relays your requests to the upstream models.`),
  },
  {
    start: 'After signing in, the "API key injection" bar',
    end: 'copy and run.',
    replace: '',
  },
  {
    start: 'Verify by chatting in PlaygroundReplay the interactive setup guide',
    end: 'The setup guide walks through key creation, wallet, and your first chat on the real pages.',
    replace: '',
  },
  // ---- 图片 API 技能页：重建 4 张被抹平的表 ----
  {
    start: 'ItemValue',
    end: 'failed requests are not billed',
    replace: T(`| Item | Value |
| --- | --- |
| Base URL | \`https://api.aiwanai.cc\` |
| Auth header | \`Authorization: Bearer sk-your-api-key\` |
| Key source | Console → API Keys, choose the image-generation group |
| Response format | Images are always returned as a \`url\` field (even if you ask for \`b64_json\`); links stay valid for about 1 hour — download promptly to keep them |
| Billing | Per-call flat pricing; cost and image URLs are recorded in the usage log; failed requests are not billed`),
  },
  {
    start: 'What you wantEndpoint',
    end: 'Check an async taskGET /v1/images/tasks/{task_id}',
    replace: T(`| What you want | Endpoint |
| --- | --- |
| Text-to-image, normal size, done within ~30s | \`POST /v1/images/generations\` (sync) |
| Text-to-image at 2k/4k, high quality, or n>1 | \`POST /v1/images/generations/async\` + polling (recommended default) |
| Image-to-image / edits / style transfer (with reference images) | \`POST /v1/images/edits\` (multipart, sync) or \`POST /v1/images/edits/async\` |
| Check an async task | \`GET /v1/images/tasks/{task_id}\` |`),
  },
  {
    start: 'ModelNotessize options',
    end: 'async onlyauto,3840x2160,2160x3840,2880x2880same✅',
    replace: T(`| Model | Notes | size options | quality | edits support |
| --- | --- | --- | --- | --- |
| \`gpt-image-2\` | Recommended default, 1024 tier | auto, 1024x1024, 1536x864, 864x1536 | auto/low/medium/high | ✅ (up to 4 reference images) |
| \`gpt-image-2-2k\` | 2k tier, prefer async | auto, 2048x2048, 2560x1440, 1440x2560 | same | ✅ |
| \`gpt-image-2-4k\` | 4k/UHD tier, async only | auto, 3840x2160, 2160x3840, 2880x2880 | same | ✅ |`),
  },
  {
    start: 'Status / symptomMeaningWhat the AI should do',
    end: 'Back off 10–30s and retry once',
    replace: T(`| Status / symptom | Meaning | What the AI should do |
| --- | --- | --- |
| 400 "image-generation-only model" | Image model sent to a chat endpoint | Switch to the images endpoints on this page |
| 400 missing model | No \`model\` in body | Add it and retry |
| 401 | Invalid/disabled key | Ask the user to check the key |
| 403 | Group has no access to this model | Use a key from the image-generation group, or switch models |
| 413 | Async body over 15MB | Compress references, or use the sync endpoint |
| 429 (chat/sync) | Rate limited | Wait 10–30s and retry |
| 429 "queued image tasks reached the limit of 20" | Async queue full | Wait for existing tasks to finish |
| 524 / timeout | Sync generation exceeded the CDN limit | Use the async endpoints (that is what they are for) |
| Task \`failed\` | See \`error.message\` | Handle per message; unbilled failures can simply be resubmitted |
| 5xx | Upstream turbulence | Back off 10–30s and retry once |`),
  },
  // ---- 分组：重建模式表（本站未启用 auto 模式与分组链，仅保留两种模式）----
  {
    start: 'ModeMeaning',
    end: 'Group chainTries multiple groups in an order you define yourself',
    replace: T(`| Mode | Meaning |
| --- | --- |
| Follow account group (default) | The token has no group of its own and always follows your account's current group |
| A specific group | Locked to that group; requests fail if it becomes unavailable |`),
  },
  // ---- 分组：删除 auto 模式 / 分组链 / 跨分组重试 / 平滑迁移（本站均未启用，原文来自 SHUAI API）----
  {
    start: "A token'sgroupdecides",
    end: 'andsmooth group migration',
    replace: T(`A token's group decides which pool of upstream channels serves your requests and which price multiplier applies.`),
  },
  {
    start: 'There are four modes',
    end: 'There are four modes',
    replace: 'There are two modes:',
  },
  {
    start: 'auto: automatic group selection',
    end: 'the request simply fails when that group is unavailable',
    replace: '',
  },
  {
    start: 'Smooth migration: your keys survive group changes',
    end: 'depends on whether continuity or price stability matters more to you',
    replace: '',
  },
  // ---- 支持中心 / 平台导览：移除「auto 模式与分组链」引用、平滑迁移开关引用 ----
  {
    start: 'Group and billing questions:',
    end: 'Group and billing questions:',
    replace: 'Group and billing questions:Groups',
  },
  {
    start: 'Notification channels(Profile',
    end: 'seeGroups, auto mode & group chains',
    replace: T('Notification channels (Profile → Notifications): low-balance alerts and more via email / webhook / Bark / Gotify, with a dedicated notification address and threshold.'),
  },
  // ---- 本地测速：重建徽标表 ----
  {
    start: 'BadgeMeaning',
    end: 'Red "connection failed"Unreachable',
    replace: T(`| Badge | Meaning |
| --- | --- |
| Green 3/3 · 128ms | All 3 probes passed — stable |
| Yellow 1/3 · … | Partially passed — intermittent |
| Blue "reachable, unverified" | Connects, but the response is not this site's gateway (possibly intercepted by a middlebox) — use with caution |
| Red "connection failed" | Unreachable |`),
  },
  // ---- 支持中心：重建工单字段表 ----
  {
    start: 'FieldNotes',
    end: 'paste or drag-and-drop',
    replace: T(`| Field | Notes |
| --- | --- |
| Title | One-line summary, up to 120 characters |
| Category | Quota, platform error, frontend error, API error, refund, account, other |
| Verification level | Unverified / Suspected / Reproducible / Confirmed — choose honestly; it helps staff triage |
| Body | Full description, up to 8,000 characters |
| Screenshots | Up to 5 images per message, 8MB each (JPEG/PNG/WebP/GIF), paste or drag-and-drop |`),
  },
  // ---- DeepSeek：重建连接参数表 ----
  {
    start: 'SettingValue',
    end: 'Lower-cost modeldeepseek-v4-flash',
    replace: T(`| Setting | Value |
| --- | --- |
| Provider type | Custom / OpenAI Compatible |
| API Base URL | \`https://api.aiwanai.cc/v1\` |
| API key | A key created under Console → API Keys on this site |
| Recommended endpoint | \`POST https://api.aiwanai.cc/v1/responses\` |
| Recommended model | \`deepseek-v4-pro\` |
| Lower-cost model | \`deepseek-v4-flash\` |`),
  },
  // ---- DeepSeek：重建 ZCode 自定义供应商字段表 ----
  {
    start: 'NameSHUAI API - DeepSeek',
    end: 'API formatResponses (/responses)',
    replace: T(`| Field | Value |
| --- | --- |
| Name | waner~ API站 - DeepSeek |
| Base URL | \`https://api.aiwanai.cc/v1\` |
| API Key | Your key from this site |
| API format | Responses (/responses) |`),
  },
  // ---- DeepSeek：重建工具选择表 ----
  {
    start: 'PreferenceRecommendation',
    end: 'ZCode is more approachable',
    replace: T(`| Preference | Recommendation |
| --- | --- |
| Terminal-first, auditable config, config alongside projects | OpenCode |
| Visual project, model, and conversation management | ZCode |
| Reuse one setup across many repositories | OpenCode global config |
| First coding-agent experience | ZCode is more approachable |`),
  },
  // ---- CC Switch：删除站点特有的一键导入挂件区块 ----
  {
    start: 'Pick an API key and a target app below',
    end: 'Sign inDownload CC Switch',
    replace: T(`In **Console → API Keys**, open any key's action menu and pick **Import to CC Switch**. Your browser opens CC Switch with a confirmation dialog; after confirming, this site is written into the tool's config as a provider and activated.`),
  },
  // ---- XiaoBaiSwitch：删除站点特有的一键导入挂件区块 ----
  {
    start: 'Pick an API key below and click "Import to XiaoBaiSwitch"',
    end: 'Sign inDownload XiaoBaiSwitch',
    replace: T(`In **Console → API Keys**, open any key's action menu and pick **Import to XiaoBaiSwitch**. Your browser opens the app with a confirmation dialog; check the site name, routes, and key prefix, then confirm to write this site in as an upstream site.`),
  },
  // ---- 排障：重建错误码表 ----
  {
    start: 'StatusTypical errorCause & fix',
    end: 'report the model + time if it repeats',
    replace: T(`| Status | Typical error | Cause & fix |
| --- | --- | --- |
| 401 | \`invalid_api_key\` | Wrong/deleted key or missing \`sk-\` prefix; or the Base URL points at another site. Recreate the key and confirm \`https://api.aiwanai.cc\` |
| 403 | \`Forbidden\` | The key has model/IP restrictions, or its group can't access that model. Check key settings or create an unrestricted key |
| 404 | \`model_not_found\` | Misspelled model, model not in your group, or missing \`/v1\` in the path. Copy the exact name from Pricing |
| 408 | Request timeout | Raise the client timeout; use streaming; or pick a faster model |
| 413 | \`context_length_exceeded\` | Input or history exceeds the model's context. Trim history, lower \`max_tokens\`, or use a long-context model |
| 429 | \`rate_limit_exceeded\` | Concurrency/frequency limit hit. Reduce parallelism and add exponential-backoff retries |
| 429 | \`insufficient_quota\` | Balance or key quota exhausted. Top up in Wallet or raise the key quota |
| 500/502 | Upstream error | Temporary site/upstream fault — retry shortly; report if persistent |
| 503 | \`no available channel\` | No channel currently serves that model. Retry later, switch model or group; report the model + time if it repeats |`),
  },
];

// ---------------------------------------------------------------------------
// 全局行过滤（删掉站点文档系统的交互式碎句）
// ---------------------------------------------------------------------------
const DROP_LINE_PATTERNS = [
  /^Sign in to inject your real API key into every sample on this page\.$/,
  /^\*\*原文：\*\*\s*https?:\/\//,
  /^Sign in to import your key into .+ with one click\.$/,
  // 站点信息章节的中文导语（英文版不需要；品牌替换前执行，此时可能还是 SHUAI API）
  /^.+是统一 AI API 网关与管理控制台。/,
];

// ---------------------------------------------------------------------------
// 抓取丢空格 artifact 修复表（人工校验版）
// ---------------------------------------------------------------------------
const TOKEN_REPAIRS = {
  // —— 含 OpenAI / DeepSeek 等专有名词的连接 ——
  CompletionsorOpenAI: 'Completions or OpenAI', LlamaIndexand: 'LlamaIndex and',
  OpenOpenAI: 'Open OpenAI', anOpenAI: 'an OpenAI', anOpenAIprovider: 'an OpenAI provider',
  chooseOpenAI: 'choose OpenAI', theOpenAI: 'the OpenAI', theOpenAItype: 'the OpenAI type',
  notDeepSeek: 'not DeepSeek',
  // —— 站点名 / 工具名连接 ——
  NameSHUAI: 'Name', ChatCompletions: 'Chat Completions', docsEndpoints: 'docs Endpoints',
  XiaoBaiSwitchis: 'XiaoBaiSwitch is', OpenCodeis: 'OpenCode is',
  projectsOpenCode: 'projects OpenCode', repositoriesOpenCode: 'repositories OpenCode',
  recommendOpenCodeorZCode: 'recommend OpenCode or ZCode', experienceZCode: 'experience ZCode',
  managementZCode: 'management ZCode', immersivetranslate: 'Immersive Translate',
  builtxiaobaiswitch: 'built xiaobaiswitch', itsxiaobaiswitch: 'its xiaobaiswitch',
  andxiaobaiswitch: 'and xiaobaiswitch', thexiaobaiswitch: 'the xiaobaiswitch',
  andhaikualiases: 'and Haiku aliases',
  // —— 界面元素连接（控制台 / Pricing / Wallet 等）——
  OpenConsole: 'Open Console', UnderConsole: 'Under Console', inConsole: 'in Console',
  onConsole: 'on Console', underConsole: 'under Console', sourceConsole: 'source Console',
  OpenSettings: 'open Settings', openSettings: 'open Settings', toSettings: 'to Settings',
  editingsettings: 'editing settings', createsettings: 'create settings',
  usingsettings: 'using settings', inconfig: 'in config', configurationin: 'configuration in',
  intoConfiguration: 'into Configuration', userprofile: 'user profile', underProfile: 'under Profile',
  sidebarPersonal: 'sidebar Personal', sidebarThird: 'sidebar Third',
  OpenWalletto: 'Open Wallet to', inWalletor: 'in Wallet or', onPricing: 'on Pricing',
  thePricingpage: 'the Pricing page', sPricingpage: 's Pricing page', pageandmodel: 'page and model',
  pagebefore: 'page before', inUsage: 'in Usage', fromUsage: 'from Usage', TheUsage: 'The Usage',
  sUsage: 's Usage', statusin: 'status in', statusto: 'status to',
  openPlayground: 'open Playground', openPlaygroundand: 'open Playground and',
  PlaygroundReplay: 'Playground Replay', theSupport: 'the Support', seeGroups: 'see Groups',
  SeeError: 'See Error', SeeQuick: 'See Quick', seeError: 'see Error', withError: 'with Error',
  theCC: 'the CC', duringCC: 'during CC', throughCC: 'through CC', withCC: 'with CC',
  theCodex: 'the Codex', theClaude: 'the Claude', theAnthropic: 'the Anthropic',
  theREADME: 'the README', theAPI: 'the API', TheAI: 'The AI', anAPI: 'an API', anIP: 'an IP',
  // —— 动词 / 界面动作连接 ——
  clickAdd: 'click Add', ClickAdd: 'Click Add', clickCopy: 'click Copy', clickFetch: 'click Fetch',
  clickImport: 'click Import', clickLocal: 'click Local', clickManage: 'click Manage',
  chooseAnthropic: 'choose Anthropic', chooseChat: 'choose Chat', choosePiin: 'choose Pi in',
  chooseUse: 'choose Use', SelectOther: 'Select Other', selectDisable: 'select Disable',
  SelectResponses: 'Select Responses', selectResponses: 'select Responses',
  confirmResponses: 'confirm Responses', explicitResponses: 'explicit Responses',
  formatResponses: 'format Responses', formatImages: 'format Images',
  autoAutomatically: 'auto Automatically', autoor: 'auto or', toautoor: 'to auto or',
  doesnotauto: 'does not auto', importapplies: 'import applies',
  endpointSwitch: 'endpoint Switch', endpointPOST: 'endpoint POST', acceptsPOST: 'accepts POST',
  orPOST: 'or POST', sPOST: 's POST', usePOST: 'use POST', taskGET: 'task GET',
  wantEndpoint: 'want Endpoint', levelUnverified: 'level Unverified',
  // —— 代码 / 协议术语连接 ——
  andbaseURL: 'and baseURL', apiKey: 'apiKey', modelformat: 'model format',
  formatmenu: 'format menu', formatin: 'format in', tomllocation: 'TOML location',
  jsonin: 'JSON in', jsonthere: 'JSON there', jsoncomplete: 'JSON complete',
  completionsentry: 'completions entry', responsesand: 'responses and',
  responsesin: 'responses in', responsesis: 'responses is', responsespath: 'responses path',
  responsesendpoint: 'responses endpoint', responsestohttps: 'responses to https',
  responsesunderhttps: 'responses under https', missingsk: 'missing sk-',
  runxattr: 'run xattr', usehttporhttpsand: 'use http or https and',
  // —— 其他散文连接 ——
  chainTries: 'chain Tries', chainsper: 'chains per', groupLocked: 'group Locked',
  channelNo: 'channel No', messageHandle: 'message Handle', messagesitself: 'messages itself',
  timeoutRaise: 'timeout Raise', timeoutSync: 'timeout Sync', turbulenceBack: 'turbulence Back',
  errorCause: 'error Cause', errorTemporary: 'error Temporary', modelUse: 'model Use',
  modelNomodelin: 'model No model in', typeCustom: 'type Custom', underAPI: 'under API',
  bodyAdd: 'body Add', keyA: 'key A', keyAsk: 'key Ask', linkOpen: 'link Open',
  limitUse: 'limit Use', limitedWait: 'limited Wait', fullWait: 'full Wait',
  inDownload: 'in Download', openThird: 'open Third', pressWin: 'press Win',
  latencyonly: 'latency only', measuresgateway: 'measures gateway',
  themedianof: 'the median of', theofficial: 'the official', onyourmachine: 'on your machine',
  assistantat: 'assistant at', asanthropic: 'as Anthropic', itsofficial: 'its official',
  flashor: 'flash or', totemporarily: 'to temporarily', notificationlisting: 'notification listing',
  supportupgrade: 'support upgrade', testhelps: 'test helps', sectionfor: 'section for',
  storedonly: 'stored only', succeededorfailed: 'succeeded or failed',
  ofgpt: 'of gpt', Keysand: 'Keys and', tooland: 'tool and', docsand: 'docs and',
  andAuthorization: 'and Authorization', defaultx: 'default x',
  onsucceededreadresult: 'on succeeded read result', onfailedreaderror: 'on failed read error',
  OpenAIprovider: 'OpenAI provider', OpenAItype: 'OpenAI type', toPi: 'to Pi', useCopy: 'use Copy',
  asupport: 'a support', agroup: 'a group', afew: 'a few', aurl: 'a url', asgpt: 'as gpt',
  "token'sgroupdecides": "token's group decides", beenappliedto: 'been applied to',
  theurlfield: 'the `url` field', aurlfield: 'a `url` field',
  theurlin: 'the `url` in', theurlreturn: 'the `url` return',
  Nevercall: 'Never call', "SDK'sbase_urlat": "SDK's base_url at",
  // —— 残余散文连接（虚词/介词类）——
  addweb_search: 'add web_search', ahighermultiplier: 'a higher multiplier',
  coveringintegration: 'covering integration', hourafter: 'hour after',
  installationandcustom: 'installation and custom', levelweb_searchkey: 'level web_search key',
  openaihere: 'OpenAI here', sendresponse_format: 'send response_format',
  setweb_search: 'set web_search', therecommended: 'the recommended',
  theresultfield: 'the result field', thewhole: 'the whole',
  v1automatically: 'v1 automatically', v1manually: 'v1 manually',
  Keyshas: 'Keys has',
  '小白Switchentry': 'XiaoBaiSwitch entry',
  // —— 表头连接（未被表格重建消费的残余）——
  FieldNotes: 'Field Notes', FieldValue: 'Field Value', ItemValue: 'Item Value',
  BadgeMeaning: 'Badge Meaning', ModeMeaning: 'Mode Meaning', SettingValue: 'Setting Value',
  StatusTypical: 'Status Typical', TitleOne: 'Title One', CategoryQuota: 'Category Quota',
  BodyFull: 'Body Full', ScreenshotsUp: 'Screenshots Up', BillingPer: 'Billing Per',
  ModelNotessize: 'Model Notes size', symptomMeaningWhat: 'symptom Meaning What',
  TaskfailedSeeerror: 'Task failed See error', PreferenceRecommendation: 'Preference Recommendation',
  AuthheaderAuthorization: 'Auth header Authorization', KeyYour: 'Key Your',
};

// ---------------------------------------------------------------------------
// 品牌 / 域名换血（顺序敏感：长串先替换）
// ---------------------------------------------------------------------------
const BRAND_REPLACEMENTS = [
  ['https://cdn.shuaiapi.com', 'https://api.aiwanai.cc'],
  ['https://api.shuaiapi.com', 'https://api.aiwanai.cc'],
  ['https://api.oai.sb', 'https://api.aiwanai.cc'],
  ['https://oai.sb', 'https://api.aiwanai.cc'],
  ['shuaiapi.com', 'aiwanai.cc'],
  ['oai.sb', 'aiwanai.cc'],
  ['SHUAI API · DeepSeek', 'waner~ API站 · DeepSeek'],
  ['SHUAI API - DeepSeek', 'waner~ API站 - DeepSeek'],
  ['SHUAI API', 'waner~ API站'],
  ['帅API', 'waner~ API站'],
  ['帅站', '本站'],
  ['小白Switch', 'XiaoBaiSwitch'],
  ['"shuaiapi"', '"aiwanai"'],
  ['shuaiapi/', 'aiwanai/'],
  ['Entershuaiapias', 'Enter shuaiapi as'],
  ['beshuaiapi', 'be shuaiapi'],
  // 兜底：此时残余的 shuaiapi 只剩 OpenCode provider id（域名已全部替换完毕）
  ['shuaiapi', 'aiwanai'],
];

// ---------------------------------------------------------------------------
// 工具函数
// ---------------------------------------------------------------------------
function applyEdits(lines) {
  const out = [...lines];
  for (const edit of EDITS) {
    let startIdx = -1, endIdx = -1;
    for (let i = 0; i < out.length; i++) {
      if (startIdx < 0 && out[i].includes(edit.start)) { startIdx = i; }
      if (startIdx >= 0 && out[i].includes(edit.end)) { endIdx = i; break; }
    }
    if (startIdx < 0 || endIdx < 0) {
      throw new Error(`edit anchor not found: "${edit.start}" .. "${edit.end}"`);
    }
    const replacement = edit.replace ? edit.replace.split('\n') : [];
    out.splice(startIdx, endIdx - startIdx + 1, ...replacement);
  }
  return out;
}

function applyTokenRepairs(text) {
  // 仅在代码围栏外修复
  const segments = text.split(/(```[\s\S]*?```)/);
  return segments
    .map((seg, i) => {
      if (i % 2 === 1) return seg; // 代码块
      // 先掩码保护专有名词，防止大写串规则破坏 AQBot / xAI / baseURL 等
      const PROTECTED = [
        'AQBot', 'xAI', 'baseURL', 'apiKey', 'ChatGPT', 'GitHub', 'DeepSeek', 'WeChat',
        'PowerShell', 'JavaScript', 'TypeScript', 'LangChain', 'LlamaIndex', 'OpenCode',
        'OpenClaw', 'XiaoBaiSwitch', 'ZCode', 'WinGet', 'WebP', 'gpt-image', 'antml', 'generateContent',
      ];
      const masks = [];
      let t = seg;
      for (const p of PROTECTED) {
        const re = new RegExp(p.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g');
        t = t.replace(re, () => {
          masks.push(p);
          return `\u0000${masks.length - 1}\u0000`;
        });
      }
      for (const [from, to] of Object.entries(TOKEN_REPAIRS)) {
        t = t.split(from).join(to);
      }
      // 标点后紧贴字母（抓取剥离链接/加粗后丢空格）
      t = t.replace(/([:;,])([A-Z])/g, '$1 $2');
      // 标签:值 连接（如 Downloads:https:// / API Key:sk- / WSL:~/path / failed:error）
      t = t.replace(/([a-zA-Z]):(https?:\/\/|sk-[a-z]|~|[a-z])/g, '$1: $2');
      // 逗号/分号后紧贴小写字母或 / @（英文里必有空格，数字不受影响）
      t = t.replace(/([,;])(?=[/[@a-z])/g, '$1 ');
      // 左括号前紧贴字母（如 Logs(with）
      t = t.replace(/([a-zA-Z])\(/g, '$1 (');
      // 右括号/方括号/@ 前后紧贴字母（如 (/responses)API、inside[...]、through@ai-sdk）
      t = t.replace(/\)([A-Za-z])/g, ') $1');
      t = t.replace(/([a-zA-Z])([[@])/g, '$1 $2');
      // 小写词 + 大写环境变量名（如 asANTHROPIC_BASE_URL、exportCLAUDE_CODE）
      t = t.replace(/([a-z])([A-Z]{2,})(?=[a-z_]|$)/g, '$1 $2');
      // 独立大写串 + 小写词（如 MODELis、APIformat）
      t = t.replace(/(?<![a-zA-Z])([A-Z][A-Z_0-9]{2,})([a-z])/g, '$1 $2');
      // 小写词 + 紧跟的 URL（如 exactlyhttps://）
      t = t.replace(/([a-z])(https?:\/\/)/g, '$1 $2');
      // 数字 + 虚词（如 4.5andprompt、5is）；必须以数字开头，避免域名 aqbot.top 被误拆
      t = t.replace(/(\d\.?)(and|or|is|to|in|of|for)(?=[a-z]|\s)/g, '$1 $2');
      // 撇号所有格 + 小写词（如 SDK'sbase_url）
      t = t.replace(/(['’]s)([a-z]{4,})/g, '$1 $2');
      // 星号 + 小写（如 *models）
      t = t.replace(/\*([a-z])/g, '* $1');
      // 字母 + 路径（如 trailing/v1、Use/status）；token 级匹配，URL 内的 /v1 不动
      t = t.replace(/(^|\s)([A-Za-z][\w-]*)\/(v1|api|docs|users|settings|status)(?![\w])/g, '$1$2 /$3');
      // —— 通用补丁：抓取残留的丢空格模式（不依赖人工词典）——
      // /v1 后紧贴字母（the/v1URL → the /v1 URL）
      t = t.replace(/(\/v1)(?=[A-Za-z])/g, '$1 ');
      // 小写词 + 版本号（Usev3.20.3 → Use v3.20.3）
      t = t.replace(/([a-z])(v\d+\.\d+)/g, '$1 $2');
      // scheme://词（theccswitch://protocol → theccswitch:// protocol）；http(s):// 保持不动
      t = t.replace(/(?<![\w])([a-z][a-z0-9+.-]*):\/\/([a-z])/g, (m, scheme, rest) =>
        /^(https?)$/.test(scheme) ? m : `${scheme}:// ${rest}`);
      // 小写句末标点 + 大写句首（site.The → site. The）
      t = t.replace(/([a-z])\.(?=[A-Z])/g, '$1. ');
      // 小写 + em/en dash（url— that → url — that；change— no → change — no）
      t = t.replace(/([a-z])([—–])(?=[a-z])/g, '$1 $2 ');
      t = t.replace(/([a-z])([—–])(?=\s)/g, '$1 $2');
      // 小写 + 首字母大写词（如 recommendOpenCode、theOpus）；macOS/xAI 等已被掩码保护
      t = t.replace(/([a-z])([A-Z][a-z])/g, '$1 $2');
      // 掩码边界：被保护专有名词与相邻单词之间补空格（如 andbaseURL、Downloads:OpenCode）
      t = t.replace(/([a-z])(\u0000\d+\u0000)/g, '$1 $2');
      t = t.replace(/(\u0000\d+\u0000)([a-z])/g, '$1 $2');
      t = t.replace(/([:;,])(\u0000\d+\u0000)/g, '$1 $2');
      t = t.replace(/(\u0000\d+\u0000):/g, '$1: ');
      // 还原被掩码的专有名词
      t = t.replace(/\u0000(\d+)\u0000/g, (m, n) => masks[Number(n)]);
      return t;
    })
    .join('');
}

// ---------------------------------------------------------------------------
// 字典驱动的通用拆分：散文中不在词典内的词，若能唯一地二分为两个已知词则拆分
// ---------------------------------------------------------------------------
const DICT_WORDS = new Set(`
a about above accept accepts add adds adding after again against all also although always am
among amount an and another any
anyone anything applies apply app apps append appends are area areas around as ask at
attach attached attention audit author auto automatically available await awaiting
back background backup balance base be became become becomes been before begin behavior behind
being below belong belongs between beyond both breakdown bug bugs but button buttons
cache call called calling calls can cancel cannot case cases center certain chain change check
cherry choice choose claude clear click client clients code complete completions config
configuration confirm connection console contact content continue control copy core cost costs
could create created creates credit cross curl current custom customer
dash data day days default define delete deleted deliver description detail details detect
dialog differ directory disable disabled discount do does doesn done double down download
downloads downtime draft during each edge edit edits either email emails else empty enable
enabled end endpoint endpoints enough enter entire error errors even event every everyday
everyone everything example examples existing expand expired extension
fail failed fails failure fallback false faq feature features feed few field fields file
files filter first fixed flag flat for format formats found four free from full fund
gallery gap gateway generate generated generation get gets give given global globally go goes
group groups had has have had header headers here higher his how html http https id ids
if image images img in include included including increase independent index info information
input inside install installed installer instance instead instruction instructions interface
internal interrupt into invalid is issue issues issued it its itself
just key keys know known knowledge
label language large last latest later layer level line link links list live local locally
log logs long look lower
mail main maintain major make manages many map mapping markdown master max maybe me means
member members memory merge message messages method might min minor mode model models more
most much must my name names native need needs new newer next no node none non note notes
notice notified now number numbers object of off offer offered offers official often on once
one only open opened order org original other others out over own owner
page pages parse part parts pass password past patch path paths per perform performance period
person personal pick pin place plan plans platform please point points pool pop port post
prefer preset price pricing primary priority private process produce product profile provider
providers public push query question
rate rather raw read reader receive received recommend recommended record records redirect
refers register registered releases remain remote remove removed rename reorder repeatedly
replace request requests require required reset response responses result results retry return
returns review right root route routes row run running
same save scope script search second section see select selected send sends separate serial
server session set setting settings setup several share short should show shown side sign
simple simply site sites sk slot slow so some something sonnet sort source sources speed
stable stage standalone standard start state status step steps still store stored stream
string strings sub subscription success such support supported switch sync system
target task tasks team temp template term test text than that the their them then there
these they thing things this those though three time times title to today together token
tokens tool toolbox tools top toward track trailing translate tree true try two type types
under unified until up update updated upgrade url us use used user users using
value values variant vendor version very via view visible
wait wallet want was way ways we web when where which while who why will window windows
with within without work works would write written
aiwanai aqbot anthropic codex ccswitch xiaobaiswitch zcode opencode openai zhipu qwen kimi
moonshot deepseek gemini grok glm b64 toml applied decide decides decided smooth
explain explains explained mkdir prompt gpt
`.trim().split(/\s+/));

// 真复合词，不应被拆分
const DICT_EXCLUDE = new Set([
  'anytime', 'backoff', 'baseline', 'behave', 'inline', 'lowercase', 'mainstream', 'nothing',
  'online', 'timeout', 'upstream', 'uptime', 'website', 'feedback', 'placeholder', 'standalone',
  'smartphone', 'homepage', 'username', 'filesystem', 'middleware', 'offine',
  'backend', 'frontend', 'signup', 'signin', 'login', 'logout', 'fallback', 'output', 'input',
  'without', 'further', 'overview', 'workflow', 'github', 'metadata', 'namespace', 'ongoing',
  'username', 'passkey', 'subagent', 'multiplier', 'endpoint', 'playground', 'sandbox',
  'toolset', 'dropdown', 'checkbox', 'readonly', 'json', 'toml', 'localhost',
]);

function splitJoinedWord(word) {
  if (word.length < 6) return null;
  if (DICT_WORDS.has(word) || DICT_EXCLUDE.has(word)) return null;
  let split = null;
  for (let i = 2; i < word.length - 1; i++) {
    const a = word.slice(0, i);
    const b = word.slice(i);
    if (DICT_WORDS.has(a) && DICT_WORDS.has(b)) {
      if (split) return null; // 多种拆分 → 不动
      split = a + ' ' + b;
    }
  }
  return split;
}

const splitterStats = [];

function applyDictSplit(text) {
  const segments = text.split(/(```[\s\S]*?```)/);
  return segments
    .map((seg, i) => {
      if (i % 2 === 1) return seg;
      return seg.replace(/\b[a-z]{4,}\b/g, (word) => {
        const fixed = splitJoinedWord(word);
        if (fixed && !splitterStats.includes(word)) splitterStats.push(word);
        return fixed || word;
      });
    })
    .join('');
}

// ---------------------------------------------------------------------------
// 行内代码美化：散文中的裸代码术语包反引号（代码块与已有行内代码段内不处理）
// ---------------------------------------------------------------------------
const INLINE_CODE = [
  'web_search', 'response_format', 'model_reasoning_effort', 'b64_json', 'max_tokens',
  'wire_api', 'auth.json', 'config.toml', 'settings.json', 'opencode.json', 'models.json',
  'base_url', 'baseURL', 'apiKey', 'sk-', 'xattr', '/v1/responses', '/v1/chat/completions',
  '/v1/messages', '/v1/models', 'invalid_api_key', 'model_not_found', 'insufficient_quota',
  'rate_limit_exceeded', 'context_length_exceeded', 'no available channel', 'generateContent',
];

function applyInlineCode(text) {
  const segments = text.split(/(```[\s\S]*?```)/);
  return segments
    .map((seg, i) => {
      if (i % 2 === 1) return seg; // 代码块
      // 跳过已有的行内代码段，只处理普通散文
      return seg
        .split(/(`[^`\n]+`)/)
        .map((part, j) => {
          if (j % 2 === 1) return part;
          let t = part;
          for (const term of INLINE_CODE) {
            const escaped = term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
            t = t.replace(new RegExp(`(?<![\\\`a-zA-Z_.-])${escaped}(?![a-zA-Z_\`])`, 'g'), '`' + term + '`');
          }
          // 闭合反引号后紧贴左括号或小写字母时补空格（如 `config.toml`(not）；
          // 用 lookbehind 确保只处理闭合反引号（前面是单词字符），不破坏开头的反引号
          t = t.replace(/(?<=\w)`(?=[a-z(])/g, '` ');
          return t;
        })
        .join('');
    })
    .join('');
}

function applyBrand(text) {
  let t = text;
  for (const [from, to] of BRAND_REPLACEMENTS) t = t.split(from).join(to);
  return t;
}

function splitSections(lines) {
  const sections = [];
  let cur = null;
  for (const ln of lines) {
    if (/^## /.test(ln)) {
      if (cur) sections.push(cur);
      cur = { title: ln.slice(3).trim(), lines: [] };
    } else if (cur) {
      cur.lines.push(ln);
    }
  }
  if (cur) sections.push(cur);
  return sections;
}

function trimBody(lines) {
  // 去掉首尾空行与残余的 --- 分隔线
  let body = lines.join('\n').replace(/^\n+/, '').replace(/\n+$/, '');
  body = body.split('\n').filter((ln) => ln.trim() === '---' ? false : true).join('\n');
  return body.trim() + '\n';
}

// ---------------------------------------------------------------------------
// 读取并清洗
// ---------------------------------------------------------------------------
const raw = fs.readFileSync(SRC, 'utf8');
let lines = raw.split('\n');
lines = applyEdits(lines);
lines = lines.filter((ln) => !DROP_LINE_PATTERNS.some((re) => re.test(ln.trim())));
let cleaned = lines.join('\n');
cleaned = applyTokenRepairs(cleaned);
cleaned = applyDictSplit(cleaned);
cleaned = applyInlineCode(cleaned);
// 行内代码段之间的逗号补空格（如 `a`,`b`）
cleaned = cleaned.replace(/,(?=`)/g, ', ');
cleaned = applyBrand(cleaned);
// 二次清理：brand 替换可能引入的多余空行
cleaned = cleaned.replace(/\n{3,}/g, '\n\n');

const sections = splitSections(cleaned.split('\n'));
const pages = new Map(); // slug -> { title, body }
const pendingMerges = []; // { into, body }
const leftovers = [];

for (const sec of sections) {
  const cfg = SECTION_MAP[sec.title];
  if (!cfg) { leftovers.push(sec.title); continue; }
  if (cfg.drop) continue;
  const body = trimBody(sec.lines);
  if (cfg.mergeInto) {
    // 站点信息并入快速开始页（循环结束后统一前置）
    pendingMerges.push({ into: cfg.mergeInto, body });
    continue;
  }
  if (pages.has(cfg.slug)) {
    throw new Error(`duplicate page slug: ${cfg.slug}`);
  }
  pages.set(cfg.slug, { title: cfg.title.en, body });
}

if (leftovers.length) {
  throw new Error(`unmapped sections: ${leftovers.join(' | ')}`);
}

for (const { into, body } of pendingMerges) {
  const target = pages.get(into);
  if (!target) throw new Error(`merge target not found: ${into}`);
  target.body = body + '\n\n' + target.body;
}

// 快速开始页：合并的站点信息章节补上小标题
const qs = pages.get('quick-start');
if (qs && !qs.body.startsWith('## ')) {
  qs.body = '## API address & authentication\n\n' + qs.body;
}
// ---------------------------------------------------------------------------
// 写出：英文指南 → src/content/docs/en/
// ---------------------------------------------------------------------------
const OUT_EN = path.join(SITE, 'src', 'content', 'docs', 'en');
fs.mkdirSync(OUT_EN, { recursive: true });
for (const [slug, page] of pages) {
  const filePath = path.join(OUT_EN, `${slug}.md`);
  // frontmatter 值含冒号/引号时需 YAML 转义
  const safeTitle = /[:#&*!|>'"%@`]/.test(page.title)
    ? JSON.stringify(page.title)
    : page.title;
  const content = page.raw ?? `---\ntitle: ${safeTitle}\n---\n\n${page.body}`;
  fs.writeFileSync(filePath, content);
  console.log(`wrote ${path.relative(SITE, filePath)}`);
}
if (splitterStats.length) {
  console.log(`\ndict-split applied to ${splitterStats.length} tokens:`);
  console.log('  ' + splitterStats.sort().join(', '));
}
console.log(`\ndone: ${pages.size} pages`);
