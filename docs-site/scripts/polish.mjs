/**
 * 共享内容美化模块（docs-site）
 *
 * 修复抓取/转换流水线拍平的代码块、丢失的标题层级与 token 修复伪影：
 *   1. 按空行切块；已有围栏、表格、列表、引用、JSX 保持原子
 *   2. 代码段识别（强锚点起始 + 全行代码形态校验）→ 包围栏 + 语言探测
 *   3. 代码块内伪影清理（反引号、func ( 空格、Runtime Error 拆分等）
 *   4. 相邻同语言代码块合并
 *   5. 短标签段落恢复为 ### 子标题（允许清单 + 形状守卫）
 *   6. 散文连写修复（URL/模型名/路径与单词粘连）
 *
 * 由 scripts/polish-pages.mjs（一次性修库）与 scripts/build-content.mjs（管线）共用。
 */

// ---------------------------------------------------------------------------
// 分块
// ---------------------------------------------------------------------------
export function splitBlocks(text) {
  const lines = text.split('\n');
  const blocks = [];
  let i = 0;

  // frontmatter
  if (lines[0] && lines[0].trim() === '---') {
    let j = 1;
    while (j < lines.length && lines[j].trim() !== '---') j++;
    blocks.push({ kind: 'fm', raw: lines.slice(0, Math.min(j + 1, lines.length)).join('\n') });
    i = j + 1;
  }

  while (i < lines.length) {
    const l = lines[i];
    if (l.trim() === '') { i++; continue; }
    if (l.trimStart().startsWith('```')) {
      // 已有围栏：整块原子化（含内部空行）
      const buf = [l];
      let j = i + 1;
      while (j < lines.length && !lines[j].trimStart().startsWith('```')) buf.push(lines[j++]);
      if (j < lines.length) buf.push(lines[j++]);
      blocks.push({ kind: 'fence', raw: buf.join('\n') });
      i = j;
      continue;
    }
    // 普通段落：到空行或围栏起点为止
    const buf = [];
    let j = i;
    while (j < lines.length && lines[j].trim() !== '' && !lines[j].trimStart().startsWith('```')) buf.push(lines[j++]);
    blocks.push({ kind: 'para', raw: buf.join('\n') });
    i = j;
  }
  return blocks;
}

// ---------------------------------------------------------------------------
// 代码行识别
// ---------------------------------------------------------------------------
const STRONG_START = [
  /^(?:pip3?|npm|npx|yarn|pnpm|bunx?|curl|docker|git|node|python3?|brew|make)\b/,
  /^(?:claude|codex|opencode|zcode)\b/,
  /^(?:import|from)\s/,
  /^(?:const|let|var|class|def|async|await|new|return|raise|with|for|while|if|elif|else|try|except|finally|print|console)\b/,
  /^(?:client|message|response|stream|task|deadline|result|data|url|r|H|BASE|KEY)\s*[=[(]/,
  /^[\w.-]+\s*\(.*\)\s*;?\s*$/,
  /^[\w`'"]+\s*=\s*[{('"]/,
  /^[A-Za-z_][\w.,\s]{0,40}=\s*["'{\[]/,   // 多目标赋值 BASE, KEY = "..."
  /^\{/,
  /^#:schema/,
  /^\[[a-zA-Z_]/,
  /^Authorization:\s*Bearer/,
  /^"[\w$-]+"\s*:/,
  /^(?:model_provider|requires_openai|wire_api|web_search|model_reasoning_effort)\b/,
];

// 散文指示词：出现 ≥2 个则几乎不可能是代码行
const PROSE_WORDS = /\b(?:the|and|with|that|this|which|your|our|their|they|them|have|been|you|are|was|will|can|may|could|should|would|about|after|before|between|during|without|within|some|both|these|those|while|because|then|there|where|what|who|why|how|does|done|only|also|very|just|like|need|must|same|other|another|many|much|first|last|next|into|over|not|from|its|his|her)\b/gi;

// 代码行内不应出现的排版字符（CJK 句末标点；散文行另有 CODELINE 形态校验兜底）
const NON_CODE_MARK = /[，。；！？]/;

const CODELINE = [
  /^```/,
  /^\s*-[HFd]\b/,                    // curl 标志 -H / -d / -F
  /^\s*--[\w-]+/,                     // 长选项
  /^\s*[}\])]+[';,]?\s*$/,
  /^\s*\}'?\s*$/,
  /^\s*}[}()\[\],\w.=\s]*$/,          // }, timeout=30).json() 等闭合后接代码
  /^\s*[{\[]/,
  /^\s*"[^"]+"\s*:/,                 // json pair
  /^\s*'[^']+'\s*:/,
  /^\s*[\w`'"]+\s*:\s*['"{\[]?/,    // js 对象键 / 标签参数（`apiKey`: / model:）
  /^\s*[\w`'"]+\s*=\s*[{('"]?/,      // toml pair / 赋值
  /^\s*[A-Za-z_][\w.]*\s*\(/,        // 函数调用
  /^\s*(?:if|elif|else|for|while|with|try|except|finally|raise|return|def|class|import|from|print|break|continue|pass)\b/,
  /^\s*#\s*$/,                        // 空注释行
  /^[A-Za-z_][\w.,\s]*=\s*/,          // 多目标赋值 BASE, KEY = ...
];

export function lineIsCode(line) {
  if (NON_CODE_MARK.test(line)) return false;
  if (STRONG_START.some((re) => re.test(line))) {
    const words = line.match(PROSE_WORDS) || [];
    return words.length < 2;
  }
  return CODELINE.some((re) => re.test(line));
}

export function isCodeParagraph(par) {
  const lines = par.split('\n').filter((l) => l.trim() !== '');
  if (lines.length === 0) return false;
  if (!STRONG_START.some((re) => re.test(lines[0]))) return false;
  if (NON_CODE_MARK.test(lines[0])) return false;
  if ((lines[0].match(PROSE_WORDS) || []).length >= 2) return false;
  return lines.every(lineIsCode);
}

// ---------------------------------------------------------------------------
// 语言探测
// ---------------------------------------------------------------------------
function detectLanguage(par) {
  const cleaned = cleanCode(par);
  const first = cleaned.split('\n')[0];
  if (/^curl\b/.test(first) || /^(?:pip3?|npm|npx|brew|docker|git|node|python3?|claude|codex|opencode)\b/.test(first)) return 'bash';
  if (/^#:schema|^\[[a-zA-Z]/.test(first) || /^(?:model_provider|requires_openai|wire_api|web_search)\b/.test(first)) return 'toml';
  if (/^Authorization:/.test(first)) return 'http';
  if (/^import\s+\w+\s+from\b/.test(first) || /^(?:const|let)\s/.test(first) || /\bconsole\.log|await\s|=>|\bnew\s+[A-Z]/.test(cleaned)) return 'javascript';
  if (/^import\b|^from\b/.test(first) || /\bprint\b|requests\.|time\.|\.json\s*\(|\bdef\s|\braise\b|\bTrue\b|\bFalse\b|f"/.test(cleaned)) return 'python';
  // 赋值一个调用结果（无 const/new/await）→ python 风格
  if (/^[a-z_]\w*\s*=\s*[A-Za-z_][\w.]*\s*\(/.test(first) && !/\bnew\b|await/.test(cleaned)) return 'python';
  if (/^\{/.test(first) || /"[\w$-]+"\s*:/.test(cleaned)) return 'json';
  return 'text';
}

// ---------------------------------------------------------------------------
// 代码块内伪影清理
// ---------------------------------------------------------------------------
export function cleanCode(par) {
  let t = par.replace(/`/g, '');
  t = t.replace(/\bRuntime Error\b/g, 'RuntimeError');
  t = t.replace(/([A-Za-z0-9_\)\]])\s+\(/g, '$1(');
  t = t.replace(/([A-Za-z0-9_\]])\s+\[/g, '$1[');
  t = t.replace(/(?<=\S) {2,}/g, ' ');
  return t.replace(/\n{3,}/g, '\n\n').trim();
}

// ---------------------------------------------------------------------------
// 子标题允许清单（形状守卫：单行、≤72 字符、不以句号结尾）
// ---------------------------------------------------------------------------
export const HEADING_EN = [
  /^Step \d+\s*[:.:：—–-]/,
  /^Option \d+\s*[:.:：]/,
  /^\d+\.\s+[A-Z0-9]/,
  /^(?:Python|JavaScript(?:\s*\/\s*TypeScript)?|TypeScript|Bash|Shell)$/,
  /^(?:Prerequisites|Troubleshooting|Notes|Key points|Next steps|Related reading|Good to know|Quick start|Content policy|Data retention|Rate limits|Official links|One-click import|Manual setup|Import from the keys page|Trusted sources and self-checks|Sources to verify|Reading the results|Install from scratch|Migrate an existing installation)\b/,
  /^(?:Advanced|Optional|Important|Hard rules|Success response|Immediate response|State machine|Python polling template|Why |How to |About |My client |Which one should you use|Still stuck|What to fill in|Creating a ticket|Status and follow-up|Before you start|When something breaks|Integrate & debug|Balance & perks|Usage & observability|Account & security|Pick models per target|Connection details|Downloads:|Endpoint decision|Model capability|Sync text-to-image|Async text-to-image|Image-to-image|Error table|Checklist|Basics|What groups are|Three-step self-check|Error code reference|Frequent cases|Streams cut off|Content refused|Image links expire|Streaming)\b/,
];

export const HEADING_ZH = [
  /^第\s*\d+\s*步/,
  /^步骤\s*\d+/,
  /^\d+\.\s*[一-鿿A-Z0-9]/,
  /^(?:Python|JavaScript(?:\s*\/\s*TypeScript)?|TypeScript)$/,
  /^(?:前提条件|故障排查|排障|注意事项|小结|相关阅读|快速开始|内容政策|为什么|如何|进阶|可选|重要|成功响应|立即响应|状态机|硬性规则|挑选模型|从密钥页导入|可信来源与自查|验证|安装|迁移|阅读结果|端点决策|模型能力|同步文生图|异步文生图|图片编辑|错误表|清单|基础|常见案例|创建工单|状态与跟进|填写|绑定邮箱|集成与调试|余额|用量|账户与安全|出问题时|三步自查|错误码参考|流式中断|内容被拒|仍然卡住|连接参数|官方链接|一键导入|手动设置|图片链接|关于流式|该选哪个|分组是什么|数据保留|频率限制|平台导览|连接信息|开始之前|方式一|方式二|从零安装|升级已有安装|该选哪一个|可核对的来源|成功响应|立即响应)/,
];

function headingMatch(par, allowlist) {
  const line = par.trim().split('\n')[0];
  if (par.includes('\n')) return false;
  if (line.length > 72) return false;
  if (/[.]$|。$/.test(line)) return false;
  return allowlist.some((re) => re.test(line));
}

// ---------------------------------------------------------------------------
// 散文连写修复（仅英文散文；中文翻译本身无此问题）
// ---------------------------------------------------------------------------
const PROSE_FIXES = [
  [/(api\.aiwanai\.cc(?:\/v1)?)([a-z(])/g, '$1 $2'],
  [/(\baiwanai\.cc)([a-z])/g, '$1 $2'],
  [/([a-z])(https?:\/\/)/g, '$1 $2'],
  [/([a-z])(b64_json|response_format|web_search|base_url|wire_api|model_reasoning_effort|max_tokens|prompt_cache_key)/g, '$1 $2'],
  [/(b64_json|response_format|web_search|base_url|wire_api|max_tokens|prompt_cache_key)([a-z])/g, '$1 $2'],
  [/(deepseek-v4-(?:pro|flash)|gpt-image-2-\d?k|gpt-image-2|gpt-5\.6|claude-fable-5|claude-sonnet-5|claude-haiku-4-5|claude-opus-4-8|grok-4\.5)([a-z])/g, '$1 $2'],
  [/([A-Za-z])(deepseek-v4|gpt-image-2|gpt-5\.6|claude-fable-5|claude-sonnet-5|grok-4\.5)/g, '$1 $2'],
  [/([a-z])(sk-)/g, '$1 $2'],
  [/\b(Enter|enter|through|Run|Select|select|call|see|use|via|in|to|at|from|under)(\/(?:models|connect|responses|status|chat|messages|api))\b/g, '$1 $2'],
  [/\b(API|SDK|URL)s? s\b/g, '$1s'],
  [/(5–20MB|bytes|hundred bytes)([a-z])/g, '$1 $2'],
  [/(settings|config|auth|opencode|models)\.JSON\b/g, '$1.json'],
  [/\bconfig\.TOML\b/g, 'config.toml'],
  [/%user profile%/g, '%userprofile%'],
  [/@ai-sdk\/OpenAI/g, '@ai-sdk/openai'],
  [/\b(Pricing|Usage Logs|Console|Wallet|Support Center|Notifications|Center|Groups)(is|and|the|to|of|page|pages|shows?|decide|confirm|entries|tab|bell)\b/g, '$1 $2'],
  [/base_urlis\b/g, 'base_url is'], [/base_urlmust/g, 'base_url must'], [/base_urlat\b/g, 'base_url at'],
  [/wire_apiuses/g, 'wire_api uses'], [/thev1suffix/g, 'the /v1 suffix'],
  [/(\/v1)(suffix|endpoint|url)/g, '$1 $2'],
  [/([a-z])(auth\.json|config\.toml|settings\.json|opencode\.json|models\.json)\b/g, '$1 $2'],
  [/\b(to|for|kept|up|at|of|is|worth|capped|valid|than)(\d+)/g, '$1 $2'],
  [/([—–])([a-z])/g, '$1 $2'],
  [/\basa\b/g, 'as a'],
  [/Bearerwork/g, 'Bearer work'],
  [/Withoutstream/g, 'Without stream'], [/Withstream:/g, 'With stream:'],
  [/tiervia/g, 'tier via'],
  [/Runopencode/g, 'Run opencode'],
  [/CC Switchis/g, 'CC Switch is'], [/sono manual/g, 'so no manual'],
  [/Model namescome/g, 'Model names come'],
  [/thesk-/g, 'the sk-'],
];

// 驼峰连写拆分时需保护的专有名词
const CAMEL_PROTECTED = ['LangChain', 'LlamaIndex', 'OpenCode', 'OpenClaw', 'XiaoBaiSwitch', 'ZCode',
  'PowerShell', 'GitHub', 'ChatGPT', 'DeepSeek', 'baseURL', 'apiKey', 'WebP', 'WinGet'];

function splitCamel(text) {
  const masks = [];
  let t = text;
  for (const p of CAMEL_PROTECTED) {
    t = t.replace(new RegExp(p.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g'), () => {
      masks.push(p);
      return `\u0000${masks.length - 1}\u0000`;
    });
  }
  t = t.replace(/([a-z])([A-Z][a-z]{2,})/g, '$1 $2');
  t = t.replace(/\u0000(\d+)\u0000/g, (m, n) => masks[Number(n)]);
  return t;
}

function fixProse(text, enabled) {
  if (!enabled) return text;
  let t = text;
  for (const [from, to] of PROSE_FIXES) t = t.replace(from, to);
  return splitCamel(t);
}

// ---------------------------------------------------------------------------
// 单块分类（供中英文结构同步使用）
// ---------------------------------------------------------------------------
export function classifyBlock(block, allowlist) {
  if (block.kind !== 'para') return null;
  const first = block.raw.trim().split('\n')[0];
  if (/^\|/.test(first)) return null;                       // 表格
  if (/^>/.test(first)) return null;                         // 引用
  if (/^#{1,6}\s/.test(first)) return null;                  // 已有标题
  if (/^<[A-Z]/.test(first)) return null;                    // JSX
  if (isCodeParagraph(block.raw)) return 'fence:' + detectLanguage(block.raw);
  if (headingMatch(block.raw, allowlist)) return 'heading';
  if (/^[-*+]\s|^\d+\.\s/.test(first)) return null;          // 列表（标题判定之后）
  return 'prose';
}

// ---------------------------------------------------------------------------
// 主入口
// ---------------------------------------------------------------------------
export function polish(text, { lang = 'en', fixProseEnabled = true } = {}) {
  const blocks = splitBlocks(text);
  const allowlist = lang === 'zh' ? HEADING_ZH : HEADING_EN;
  const report = { fenced: [], headings: [], cleaned: 0 };

  const processed = blocks.map((b) => {
    if (b.kind === 'fm') return b;

    if (b.kind === 'fence') {
      // 清理已有围栏内部的伪影
      const m = b.raw.match(/^```(\w*)\n([\s\S]*?)\n```$/);
      if (m) {
        report.cleaned++;
        return { kind: 'fence', raw: '```' + (m[1] || '') + '\n' + cleanCode(m[2]) + '\n```' };
      }
      return b;
    }

    const action = classifyBlock(b, allowlist);
    if (action && action.startsWith('fence:')) {
      const langTag = action.slice('fence:'.length);
      report.fenced.push(`[${langTag}] ${b.raw.trim().split('\n')[0].slice(0, 60)}`);
      return { kind: 'fence', raw: '```' + langTag + '\n' + cleanCode(b.raw) + '\n```' };
    }
    if (action === 'heading') {
      report.headings.push(b.raw.trim().slice(0, 60));
      return { kind: 'heading', raw: '### ' + b.raw.trim().replace(/[:：]$/, '') };
    }
    if (action === null) return b;   // 表格 / 列表 / 引用 / 已有标题 / JSX：原样保留

    return { kind: 'para', raw: fixProse(b.raw, fixProseEnabled && lang === 'en') };
  });

  // 合并相邻同语言围栏
  const merged = [];
  for (const b of processed) {
    const prev = merged[merged.length - 1];
    if (
      b.kind === 'fence' && prev && prev.kind === 'fence' &&
      prev.raw.match(/^```(\w*)/)[1] === b.raw.match(/^```(\w*)/)[1]
    ) {
      prev.raw = prev.raw.replace(/```$/, '') + '\n' + b.raw.replace(/^```(\w*)\n/, '');
      report.merged = (report.merged || 0) + 1;
    } else {
      merged.push(b);
    }
  }

  return { text: merged.map((b) => b.raw).join('\n\n') + '\n', report };
}

// 中文也需要的少量安全散文修复（模型名/URL 与中文粘连）
export function polishZhLight(text) {
  return text
    .replace(/([一-鿿，。：；])(https?:\/\/)/g, '$1 $2')
    .replace(/(https?:\/\/[^\s，。：；）]+)([一-鿿])/g, '$1 $2')
    .replace(/%user profile%/g, '%userprofile%');
}
