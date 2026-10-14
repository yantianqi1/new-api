/**
 * 一次性修库脚本：把 polish.mjs 应用到全部已生成页面（中英双语）。
 *
 *   node scripts/polish-pages.mjs --dry      # 只报告，不写文件
 *   node scripts/polish-pages.mjs            # 写入
 *
 * 英文页（源）：代码围栏 + 子标题 + 散文修复。
 * 中文页（译）：代码围栏（分类器自决）+ 子标题（ZH 允许清单；段落并行时按英文动作移植）。
 * index.mdx / models.md 跳过（手写 / 数据生成）。
 */
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { splitBlocks, classifyBlock, polish, cleanCode as cleanCodeBlock, polishZhLight, HEADING_EN, HEADING_ZH } from './polish.mjs';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const SITE = path.resolve(__dirname, '..');
const DOCS = path.join(SITE, 'src', 'content', 'docs');
const DRY = process.argv.includes('--dry');

const SKIP = new Set(['index.mdx', 'models.md']);

function shapeValidHeading(raw) {
  const line = raw.trim().split('\n')[0];
  return !raw.includes('\n') && line.length <= 72 && !/[.:。]$/.test(line);
}

const files = fs.readdirSync(path.join(DOCS, 'en')).filter((f) => f.endsWith('.md') && !SKIP.has(f)).sort();
let totalFenced = 0, totalHeadings = 0;

for (const f of files) {
  const enPath = path.join(DOCS, 'en', f);
  const zhPath = path.join(DOCS, f);
  const enText = fs.readFileSync(enPath, 'utf8');
  const zhText = fs.existsSync(zhPath) ? fs.readFileSync(zhPath, 'utf8') : null;

  // ---- 英文 ----
  const { text: enOut, report: enReport } = polish(enText, { lang: 'en' });
  totalFenced += enReport.fenced.length;
  totalHeadings += enReport.headings.length;
  console.log(`\n══════ en/${f}  fences=${enReport.fenced.length} headings=${enReport.headings.length}${enReport.merged ? ` merged=${enReport.merged}` : ''}${enReport.cleaned ? ` cleaned=${enReport.cleaned}` : ''}`);
  for (const x of enReport.fenced) console.log(`  🔒 ${x}`);
  for (const x of enReport.headings) console.log(`  📌 ${x}`);

  // ---- 中文 ----
  if (zhText !== null) {
    const enBlocks = splitBlocks(enText);
    const zhBlocks = splitBlocks(zhText);
    const parallel = enBlocks.length === zhBlocks.length;

    // 中文侧：自身分类（围栏 + ZH 允许清单标题）；并行页再按英文动作补移植标题
    let transplanted = 0;
    for (let i = 0; i < zhBlocks.length; i++) {
      // 已有围栏：同样清理内部伪影
      if (zhBlocks[i].kind === 'fence') {
        const m = zhBlocks[i].raw.match(/^```(\w*)\n([\s\S]*?)\n```$/);
        if (m) zhBlocks[i] = { kind: 'fence', raw: '```' + (m[1] || '') + '\n' + cleanCodeBlock(m[2]) + '\n```' };
        continue;
      }
      if (zhBlocks[i].kind !== 'para') continue;
      const zhAction = classifyBlock(zhBlocks[i], HEADING_ZH);
      if (zhAction && zhAction.startsWith('fence:')) {
        const langTag = zhAction.slice('fence:'.length);
        zhBlocks[i] = { kind: 'fence', raw: '```' + langTag + '\n' + cleanCodeBlock(zhBlocks[i].raw) + '\n```' };
        continue;
      }
      const enAction = parallel ? classifyBlock(enBlocks[i], HEADING_EN) : null;
      if ((zhAction === 'heading' || (enAction === 'heading' && zhAction === 'prose')) && shapeValidHeading(zhBlocks[i].raw)) {
        if (enAction === 'heading' && zhAction === 'prose') transplanted++;
        zhBlocks[i] = { kind: 'heading', raw: '### ' + zhBlocks[i].raw.trim().replace(/[:：]$/, '') };
      }
    }

    // 合并相邻同语言围栏
    const merged = [];
    for (const b of zhBlocks) {
      const prev = merged[merged.length - 1];
      if (b.kind === 'fence' && prev && prev.kind === 'fence' &&
          prev.raw.match(/^```(\w*)/)[1] === b.raw.match(/^```(\w*)/)[1]) {
        prev.raw = prev.raw.replace(/```$/, '') + '\n' + b.raw.replace(/^```(\w*)\n/, '');
      } else merged.push(b);
    }
    const finalText = polishZhLight(merged.map((b) => b.raw).join('\n\n') + '\n');
    console.log(`══════    ${f}  zh headings transplanted=${transplanted}${parallel ? '' : ' (non-parallel)'}`);
    if (!DRY) fs.writeFileSync(zhPath, finalText);
  }

  if (!DRY) fs.writeFileSync(enPath, enOut);
}

console.log(`\n${DRY ? '[DRY] ' : ''}done: ${files.length} pages, ${totalFenced} code blocks fenced, ${totalHeadings} headings restored.`);
