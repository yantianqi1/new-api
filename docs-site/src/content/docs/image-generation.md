---
title: 图片生成（URL，非 base64）
---

在waner~ API站 上生成图片时，接口返回的是图片 URL，而不是一大段 base64 数据。这不是功能缺失，而是有意为之的设计。本指南说明其中的原因，以及集成时的注意事项。

### 为什么返回 URL 而不是 base64？

base64 会把图片“转录”成 JSON 响应中的纯文本。一张高分辨率图片会变成 5–20MB 的文本，带来三个实际问题：

- **慢**：整个响应下载完成前什么都渲染不出来——普通网络环境下，这意味着对着白屏干等好几秒；
- **卡**：许多客户端在渲染超长 base64 字符串时会卡顿，甚至截断；
- **脆弱**：过大的响应在传输过程中更容易触发超时。

改用 URL 后，API 响应只有几百字节，瞬间就能返回；图片本身存放在对象存储中，通过 CDN 下载——又快又稳。主流图片 API 都是这么做的（包括 OpenAI 自家的 DALL·E URL 模式）。

因此在waner~ API站 上：图片接口始终通过 `url` 字段返回图片。即使你把 `response_format` 设为 `"b64_json"`，收到的也仍然是 URL——这不是 bug，而是平台策略，目的是让所有人的图片生成都保持快速。

### 快速开始

```bash
curl https://api.aiwanai.cc/v1/images/generations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-api-key" \
  -d '{
    "model": "gpt-image-2",
    "prompt": "A shiba inu wearing an astronaut helmet, film grain",
    "size": "1024x1024",
    "response_format": "url"
  }'
```

成功的响应类似这样：

```json
{
  "created": 1710000000,
  "data": [
    {
      "url": "https://images.example.com/xxxx.png",
      "revised_prompt": "A shiba inu wearing an astronaut helmet..."
    }
  ]
}
```

在浏览器中打开 `url`，或用 `<img>` 标签引用它，也可以直接下载——随你选择。

### 图片链接约 1 小时后失效

链接在生成后约 1 小时内可访问，之后图片会从存储中清理。这样既能合理控制存储成本，也避免无限期保留你生成的内容。

如果需要长期保存图片，请及时下载。例如用 Python：

```python
import requests

url = resp.json()["data"][0]["url"]
with open("image.png", "wb") as f:
    f.write(requests.get(url, timeout=60).content)
```

在控制台的「图片生成」页面，每张未保存的图片都会显示剩余保留时间，并提供一个「Keep locally」按钮：点击后图片数据只会保存在你浏览器的本地存储中（绝不会上传到我们的服务器），从而永久留在你的历史记录中。未保存的图片在链接失效后会自动从历史记录中移除。

### 关于流式（stream）

图片接口不需要流式。图片是一次性生成的——不存在“逐 token 输出”的过程：

- 不用 `stream`：得到的就是上面的普通 JSON，这也是推荐的方式；
- 设置 `stream: true`：请求同样有效——你会收到一个携带最终 URL 的 SSE 完成事件，只是没有中间预览帧。

一句话：忽略流式，直接发普通 JSON 请求。

### 我的客户端不显示图片

大多数主流客户端（以及 OpenAI 官方 SDK）都能正确处理 `url` 字段。如果你的客户端收到了响应却不显示图片，多半是它只读取 `b64_json` 字段：

- 建议换用（或升级到）支持 `url` 返回格式的客户端；
- 如果是你自己的应用，按上面的方法先把图片下载下来——几行代码就能兼容。

### 内容政策

请勿使用图片生成功能制作违法或不良内容。此类请求会被拦截并记录，多次违规可能影响你的账户。

### 下一步

- 还没有 API 密钥？参见[快速开始](/quick-start/)
- 遇到报错？参见[错误码与故障排除](/troubleshooting/)
- 完整的参数说明在 API 参考的 Images 分组中
