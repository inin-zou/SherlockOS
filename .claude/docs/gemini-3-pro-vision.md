# Gemini 3 Pro Vision API Documentation

> **用途**：SherlockOS 场景分析与图像理解（Scene Analysis Worker）
> **官方文档**：https://ai.google.dev/gemini-api/docs/gemini-3
> **图像理解文档**：https://ai.google.dev/gemini-api/docs/image-understanding

## 概述

Gemini 3 Pro 是 Google 最先进的推理模型，具有强大的视觉理解能力。它可以从简单的图像识别跃升到真正的视觉和空间推理。

### 核心能力

| 特性 | 值 |
|------|-----|
| Model ID | `gemini-3-pro-preview` |
| 上下文窗口 | 1M tokens 输入，64K tokens 输出 |
| 图像支持 | 最多 900 张图片/请求 |
| 图像格式 | PNG, JPEG, WebP, HEIC, HEIF |
| 最大文件大小 | 7MB (控制台) / 30MB (API) |
| 知识截止 | 2025 年 1 月 |

### 视觉特性

- **文档理解**：OCR + 复杂视觉推理
- **空间推理**：理解物体位置和关系
- **指向能力**：输出像素级精确坐标
- **多图像分析**：同时处理多张图片

## 安装

```bash
pip install google-genai>=1.52.0
```

## 环境配置

```bash
export GEMINI_API_KEY="your-api-key"
# 或
export GOOGLE_API_KEY="your-api-key"
```

## 基础用法

### 单图像分析

```python
from google import genai
from PIL import Image

client = genai.Client()

# 加载图像
image = Image.open("crime_scene.jpg")

# 分析图像
response = client.models.generate_content(
    model="gemini-3-pro-preview",
    contents=[image, "详细描述这张图片中的所有物体，包括位置、状态和任何异常。"]
)

print(response.text)
```

### 多图像分析

```python
from google import genai
from PIL import Image

client = genai.Client()

# 加载多张图像
images = [
    Image.open("scene_view_1.jpg"),
    Image.open("scene_view_2.jpg"),
    Image.open("scene_view_3.jpg"),
]

# 分析多张图像
response = client.models.generate_content(
    model="gemini-3-pro-preview",
    contents=[
        *images,
        "分析这些犯罪现场照片，识别所有物体和潜在证据，比较不同视角中的信息。"
    ]
)

print(response.text)
```

### 使用 URL 图像

```python
from google import genai
from google.genai import types

client = genai.Client()

# 从 URL 加载图像
image_url = "https://storage.example.com/scene.jpg"

response = client.models.generate_content(
    model="gemini-3-pro-preview",
    contents=[
        types.Part.from_uri(image_url, mime_type="image/jpeg"),
        "分析这张图片"
    ]
)
```

### Base64 编码图像

```python
from google import genai
from google.genai import types
import base64

client = genai.Client()

# 读取并编码图像
with open("evidence.jpg", "rb") as f:
    image_data = base64.standard_b64encode(f.read()).decode("utf-8")

response = client.models.generate_content(
    model="gemini-3-pro-preview",
    contents=[
        types.Part.from_bytes(
            data=base64.standard_b64decode(image_data),
            mime_type="image/jpeg"
        ),
        "这张图片中有什么证据？"
    ]
)
```

## Media Resolution 参数

`media_resolution` 参数控制图像处理的精度和 token 使用量：

| 级别 | Tokens/图像 | 适用场景 |
|------|-------------|----------|
| `low` | ~280 | 快速预览，低成本 |
| `medium` | ~560 | 平衡质量与成本 |
| `high` | ~1120 | 高精度分析（默认） |
| `ultra_high` | ~2240 | 最高精度，细节敏感任务 |

```python
from google import genai
from google.genai import types

client = genai.Client()

response = client.models.generate_content(
    model="gemini-3-pro-preview",
    contents=[image, "分析图像中的细节"],
    config=types.GenerateContentConfig(
        media_resolution="high"  # 或 "low", "medium", "ultra_high"
    )
)
```

## Thinking 配置

Gemini 3 Pro 支持 thinking 模式来增强推理能力：

```python
from google import genai
from google.genai import types

client = genai.Client()

response = client.models.generate_content(
    model="gemini-3-pro-preview",
    contents=[image, "分析这个犯罪现场，推断可能发生了什么"],
    config=types.GenerateContentConfig(
        thinking_config=types.ThinkingConfig(
            thinking_level="high"  # "low" 或 "high"
        )
    )
)
```

| thinking_level | 说明 |
|----------------|------|
| `low` | 最小化延迟和成本 |
| `high` | 最大化推理深度（默认） |

## 结构化输出（JSON）

```python
from google import genai
from google.genai import types

client = genai.Client()

# 定义输出 Schema
scene_analysis_schema = {
    "type": "object",
    "properties": {
        "objects": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "id": {"type": "string"},
                    "type": {"type": "string"},
                    "label": {"type": "string"},
                    "position": {"type": "string"},
                    "confidence": {"type": "number"},
                    "is_suspicious": {"type": "boolean"},
                    "notes": {"type": "string"}
                }
            }
        },
        "potential_evidence": {
            "type": "array",
            "items": {"type": "string"}
        },
        "scene_description": {"type": "string"},
        "anomalies": {
            "type": "array",
            "items": {"type": "string"}
        }
    }
}

response = client.models.generate_content(
    model="gemini-3-pro-preview",
    contents=[image, "分析这个犯罪现场"],
    config=types.GenerateContentConfig(
        response_mime_type="application/json",
        response_schema=scene_analysis_schema,
        media_resolution="high"
    )
)

import json
result = json.loads(response.text)
```

## SherlockOS 集成示例

### Scene Analysis Worker (Go HTTP Client)

```go
package clients

import (
    "bytes"
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/sherlockos/backend/internal/models"
)

const (
    geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"
    gemini3ProModel = "gemini-3-pro-preview"
)

// GeminiSceneAnalysisClient implements SceneAnalysisClient using Gemini 3 Pro
type GeminiSceneAnalysisClient struct {
    apiKey     string
    httpClient *http.Client
}

// NewGeminiSceneAnalysisClient creates a new scene analysis client
func NewGeminiSceneAnalysisClient(apiKey string) *GeminiSceneAnalysisClient {
    return &GeminiSceneAnalysisClient{
        apiKey: apiKey,
        httpClient: &http.Client{
            Timeout: 120 * time.Second,
        },
    }
}

// AnalyzeScene processes images and returns detected objects/evidence
func (c *GeminiSceneAnalysisClient) AnalyzeScene(ctx context.Context, input models.SceneAnalysisInput) (*models.SceneAnalysisOutput, error) {
    startTime := time.Now()

    // Build request with images
    prompt := c.buildAnalysisPrompt(input)

    // Make API request
    response, err := c.analyzeWithVision(ctx, input.ImageKeys, prompt)
    if err != nil {
        return nil, fmt.Errorf("gemini vision API error: %w", err)
    }

    // Parse response
    output, err := c.parseAnalysisResponse(response)
    if err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    output.AnalysisTime = time.Since(startTime).Milliseconds()
    output.ModelUsed = "gemini-3-pro-preview"
    return output, nil
}

func (c *GeminiSceneAnalysisClient) buildAnalysisPrompt(input models.SceneAnalysisInput) string {
    basePrompt := `分析这些犯罪现场/场景图片。

## 任务
1. 识别所有可见物体及其位置
2. 标记任何可疑或异常的元素
3. 识别潜在证据
4. 描述整体场景布局

## 输出格式
以 JSON 格式输出：
{
  "objects": [
    {
      "id": "obj_001",
      "type": "furniture|door|window|evidence_item|weapon|footprint|bloodstain|other",
      "label": "物体名称",
      "position_description": "位置描述（如：房间中央、窗户旁边）",
      "confidence": 0.0-1.0,
      "is_suspicious": true/false,
      "notes": "备注"
    }
  ],
  "potential_evidence": ["潜在证据描述列表"],
  "scene_description": "场景整体描述",
  "anomalies": ["异常情况列表"]
}`

    if input.Query != "" {
        basePrompt += fmt.Sprintf("\n\n## 特定问题\n%s", input.Query)
    }

    return basePrompt
}

func (c *GeminiSceneAnalysisClient) analyzeWithVision(ctx context.Context, imageKeys []string, prompt string) (string, error) {
    url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", geminiBaseURL, gemini3ProModel, c.apiKey)

    // Build parts array with images and prompt
    var parts []map[string]interface{}

    // Add images (would need to fetch from storage in production)
    for _, key := range imageKeys {
        // In production, fetch image from Supabase Storage
        // For now, assume we have base64 data
        parts = append(parts, map[string]interface{}{
            "inlineData": map[string]interface{}{
                "mimeType": "image/jpeg",
                "data":     "", // Base64 encoded image data
            },
        })
    }

    // Add text prompt (must come after images)
    parts = append(parts, map[string]interface{}{
        "text": prompt,
    })

    reqBody := map[string]interface{}{
        "contents": []map[string]interface{}{
            {"parts": parts},
        },
        "generationConfig": map[string]interface{}{
            "temperature":      0.2,
            "topP":             0.95,
            "maxOutputTokens":  8192,
            "responseMimeType": "application/json",
        },
    }

    jsonBody, _ := json.Marshal(reqBody)
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
    }

    // Parse response
    var result struct {
        Candidates []struct {
            Content struct {
                Parts []struct {
                    Text string `json:"text"`
                } `json:"parts"`
            } `json:"content"`
        } `json:"candidates"`
    }

    if err := json.Unmarshal(body, &result); err != nil {
        return "", fmt.Errorf("failed to parse response: %w", err)
    }

    if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
        return "", fmt.Errorf("empty response from API")
    }

    return result.Candidates[0].Content.Parts[0].Text, nil
}

func (c *GeminiSceneAnalysisClient) parseAnalysisResponse(response string) (*models.SceneAnalysisOutput, error) {
    var output models.SceneAnalysisOutput
    if err := json.Unmarshal([]byte(response), &output); err != nil {
        // Return partial result on parse failure
        return &models.SceneAnalysisOutput{
            SceneDescription: response,
            DetectedObjects:  []models.DetectedObject{},
        }, nil
    }
    return &output, nil
}
```

## 定价

| 模型 | 输入 (per 1M tokens) | 输出 (per 1M tokens) |
|------|---------------------|---------------------|
| gemini-3-pro-preview | $2.00 | $12.00 |
| gemini-3-pro-preview (>200k tokens) | $4.00 | $18.00 |
| gemini-3-flash-preview | $0.50 | $3.00 |

## 配额（免费层）

| 限制 | 值 |
|------|-----|
| RPM (请求/分钟) | 10 |
| RPD (请求/天) | 500 |
| TPM (Token/分钟) | 250,000 |

## 与其他模型对比

| 特性 | Gemini 3 Pro | Gemini 2.5 Flash | HunyuanImage-3.0 |
|------|--------------|------------------|------------------|
| 视觉推理深度 | 最强 | 良好 | 良好 |
| 速度 | 较慢 | 快 | 依赖硬件 |
| 成本 | 较高 | 低 | GPU 成本 |
| 空间理解 | 最强 | 良好 | 良好 |
| 指向能力 | 支持 | 不支持 | 不支持 |
| 文档理解 | 最强 | 良好 | 良好 |

**SherlockOS 建议**：
- **详细场景分析**：使用 `gemini-3-pro-preview`（高精度）
- **快速预览/迭代**：使用 `gemini-2.5-flash`（低成本）

## 最佳实践

1. **图像质量**：确保图像清晰、光线充足、无模糊
2. **图像方向**：确保图像正确旋转
3. **Prompt 顺序**：图像部分放在文本 prompt 之前
4. **分辨率选择**：
   - 快速分析用 `low`
   - 标准分析用 `high`
   - 细节敏感任务用 `ultra_high`
5. **Thinking 模式**：复杂推理任务启用 `thinking_level: high`

## 错误处理

```python
from google.api_core import exceptions

try:
    response = client.models.generate_content(...)
except exceptions.ResourceExhausted:
    # 429 Rate Limit - 等待重试
    pass
except exceptions.InvalidArgument:
    # 400 请求无效（图像格式错误等）
    pass
except exceptions.InternalServerError:
    # 500 服务器错误 - 重试
    pass
```

## 参考链接

- [Gemini 3 Developer Guide](https://ai.google.dev/gemini-api/docs/gemini-3)
- [Image Understanding Documentation](https://ai.google.dev/gemini-api/docs/image-understanding)
- [Gemini 3 Pro Overview](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/models/gemini/3-pro)
- [Gemini 3 Pro Vision Blog](https://blog.google/innovation-and-ai/technology/developers-tools/gemini-3-pro-vision/)
