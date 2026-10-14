---
title: 图像 API 技能页（面向 AI 助手）
---

本页面是waner~ API站 图像能力的完整调用手册（Skill），专为 AI 助手和自动化脚本编写：把整页内容连同图像生成分组的一个 API Key 交给你的 AI（Claude、ChatGPT、Cursor……），它就能正确完成所有文生图、图生图和异步任务调用——无需其他文档。

给 AI 的一行指令：「阅读下面的 API 手册，用这个密钥帮我生成/编辑图片：sk-xxx」

### 0. 基础信息

| 项目 | 值 |
| --- | --- |
| Base URL | `https://api.aiwanai.cc` |
| 鉴权头 | `Authorization: Bearer sk-your-api-key` |
| 密钥来源 | 控制台 → API Keys，选择图像生成分组 |
| 响应格式 | 图片始终以 `url` 字段返回（即使你请求 `b64_json`）；链接有效期约 1 小时，请及时下载保存 |
| 计费 | 按次固定计费；费用和图片 URL 记录在用量日志中；失败请求不计费 |

### 1. 端点决策表（从这里开始）

| 需求 | 端点 |
| --- | --- |
| 文生图，普通尺寸，约 30 秒内完成 | `POST /v1/images/generations`（同步） |
| 2k/4k 文生图、高质量或 n>1 | `POST /v1/images/generations/async` + 轮询（推荐默认） |
| 图生图 / 编辑 / 风格迁移（带参考图） | `POST /v1/images/edits`（multipart，同步）或 `POST /v1/images/edits/async` |
| 查询异步任务 | `GET /v1/images/tasks/{task_id}` |

### 硬性规则

绝不通过聊天端点（`/v1/chat/completions`、`/v1/responses`）调用 gpt-image-* 模型——会返回 400 并引导你使用本页的端点。

同步调用经过 CDN，连接时长上限约 100 秒；任何可能超过 60 秒的生成（4k、高质量、n>1）都应使用异步端点，否则可能遭遇 524 超时。

异步提交立即返回、不占用连接；每个账号同时最多排队 20 个任务。

### 2. 模型能力表

| 模型 | 说明 | size 选项 | quality | 支持 edits |
| --- | --- | --- | --- | --- |
| `gpt-image-2` | 推荐默认，1024 档 | auto, 1024x1024, 1536x864, 864x1536 | auto/low/medium/high | ✅（最多 4 张参考图） |
| `gpt-image-2-2k` | 2k 档，建议异步 | auto, 2048x2048, 2560x1440, 1440x2560 | 同上 | ✅ |
| `gpt-image-2-4k` | 4k/UHD 档，仅异步 | auto, 3840x2160, 2160x3840, 2880x2880 | 同上 | ✅ |

补充：gpt-image-2 系列还接受自由尺寸（16 的倍数，总像素 ≤ 3840×2160，长宽比 ≤ 3:1）。n 每次请求上限为 4；对 gpt-image-2 系列，n>1 会在服务端并行展开为多个单图请求。

### 3. 同步文生图

```bash
curl https://api.aiwanai.cc/v1/images/generations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-api-key" \
  -d '{
    "model": "gpt-image-2",
    "prompt": "a shiba inu wearing an astronaut helmet, film grain",
    "size": "1024x1024",
    "quality": "medium",
    "n": 1,
    "response_format": "url"
  }'
```

### 成功响应

```json
{
  "created": 1710000000,
  "data": [
    { "url": "https://images.example.com/xxxx.png", "revised_prompt": "..." }
  ]
}
```

### 4. 异步文生图（推荐）

### 第 1 步——提交（请求体与同步端点完全一致）

```bash
curl https://api.aiwanai.cc/v1/images/generations/async \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-api-key" \
  -d '{"model": "gpt-image-2-4k", "prompt": "cyberpunk city at night, neon rain", "size": "3840x2160", "quality": "high"}'
```

### 立即返回

```json
{ "task_id": "3f2b9c1e-....", "status": "queued" }
```

### 第 2 步——轮询（每 3–5 秒一次，最长 10 分钟）

```bash
curl https://api.aiwanai.cc/v1/images/tasks/3f2b9c1e-.... \
  -H "Authorization: Bearer sk-your-api-key"
```

状态机：queued→running→succeeded 或 failed，没有其他取值。

succeeded：result 字段是完整的同步端点响应（取 result.data[0].url）；

failed：error.message 说明原因；失败任务不计费（除非消息明确说明本次生成已计费——此时可从用量日志详情中找回图片 URL）。

```json
{
  "task_id": "3f2b9c1e-....",
  "status": "succeeded",
  "progress": "100%",
  "model": "gpt-image-2-4k",
  "created_at": 1710000000,
  "finished_at": 1710000123,
  "result": { "created": 1710000123, "data": [{ "url": "https://..." }] }
}
```

### Python 轮询模板（AI 可直接照搬）

```python
import time, requests

BASE, KEY = "https://api.aiwanai.cc", "sk-your-api-key"
H = {"Authorization": f"Bearer {KEY}"}

task = requests.post(f"{BASE}/v1/images/generations/async", headers=H, json={
    "model": "gpt-image-2-4k", "prompt": "cyberpunk city at night", "quality": "high",
}, timeout=30).json()

deadline = time.time() + 600
while time.time() < deadline:
    r = requests.get(f"{BASE}/v1/images/tasks/{task['task_id']}", headers=H, timeout=30).json()
    if r["status"] == "succeeded":
        print(r["result"]["data"][0]["url"]); break
    if r["status"] == "failed":
        raise RuntimeError(r["error"]["message"])
    time.sleep(3)
```

### 5. 图生图 / 编辑

使用 multipart 表单；通过 image[] 字段名最多传入 4 张参考图：

```bash
curl https://api.aiwanai.cc/v1/images/edits \
  -H "Authorization: Bearer sk-your-api-key" \
  -F model="gpt-image-2" \
  -F prompt="repaint in watercolor style, keep the composition" \
  -F "image[]=@reference1.png" \
  -F "image[]=@reference2.png" \
  -F size="1024x1024" \
  -F response_format="url"
```

响应格式与文生图一致（data[].url）。

参考图较大或需要高分辨率输出时，请使用 POST /v1/images/edits/async（同样的 multipart 请求体，返回 task_id，轮询方式同上）；异步提交总量上限 15MB。

### 6. 错误表

| 状态 / 现象 | 含义 | AI 应如何处理 |
| --- | --- | --- |
| 400 "image-generation-only model" | 图像模型被发到了聊天端点 | 改用本页的 images 端点 |
| 400 missing model | 请求体中没有 `model` | 补上后重试 |
| 401 | 密钥无效或已禁用 | 请用户检查密钥 |
| 403 | 当前分组无权访问该模型 | 换用图像生成分组的密钥，或更换模型 |
| 413 | 异步请求体超过 15MB | 压缩参考图，或改用同步端点 |
| 429（聊天/同步） | 触发限流 | 等待 10–30 秒后重试 |
| 429 "queued image tasks reached the limit of 20" | 异步队列已满 | 等待已有任务完成 |
| 524 / 超时 | 同步生成超过 CDN 限制 | 改用异步端点（这正是它的用途） |
| 任务 `failed` | 见 `error.message` | 按消息内容处理；未计费的失败直接重新提交即可 |
| 5xx | 上游波动 | 退避 10–30 秒后重试一次 |

### 7. AI 自查清单

已持有图像生成分组的密钥；Base URL 为 https://api.aiwanai.cc；

选择模型：日常出图 → gpt-image-2；2k/4k → 通过异步使用对应档位；编辑 → 选择支持 edits 的模型；

绝不把 gpt-image-* 发到 chat/responses 端点；

异步流程：提交 → 每 3 秒轮询 → succeeded 时读 result.data[*].url，failed 时读 error.message；

### 图片链接约 1 小时后失效：用户想留存时立即下载；

失败请求不计费；费用和历史图片 URL 都在用量日志中。
