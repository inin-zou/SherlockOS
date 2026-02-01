# Gemini 2.5 Flash API Documentation

> **用途**：SherlockOS 轨迹推理（Reasoning Worker）
> **官方文档**：https://ai.google.dev/gemini-api/docs

## 概述

Gemini 2.5 Flash 是 Google 的混合推理模型，支持 "thinking" 功能，适合复杂推理任务。

| 属性 | 值 |
|------|-----|
| Model ID | `gemini-2.5-flash` |
| 上下文窗口 | 1M tokens |
| Thinking Budget | 0-24576 tokens |
| 输出模态 | Text, Code, JSON |

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

从 [Google AI Studio](https://aistudio.google.com/) 获取 API Key。

## 基础用法

### 简单文本生成

```python
from google import genai

client = genai.Client()  # 自动读取环境变量

response = client.models.generate_content(
    model="gemini-2.5-flash",
    contents="Explain quantum computing in simple terms"
)

print(response.text)
```

### 带配置的生成

```python
from google import genai
from google.genai import types

client = genai.Client()

response = client.models.generate_content(
    model="gemini-2.5-flash",
    contents="Analyze this crime scene data...",
    config=types.GenerateContentConfig(
        temperature=0.2,      # 低温度，更确定性输出
        top_p=0.95,
        top_k=20,
        max_output_tokens=8192,
    ),
)
```

## Thinking 模式（推荐用于推理）

Thinking 模式让模型在回答前进行"思考"，提升复杂推理质量。

### 启用 Thinking

```python
from google import genai
from google.genai import types

client = genai.Client()

response = client.models.generate_content(
    model="gemini-2.5-flash",
    contents=reasoning_prompt,
    config=types.GenerateContentConfig(
        thinking_config=types.ThinkingConfig(
            thinking_budget=8192  # 思考 token 预算
        )
    )
)

# 获取思考过程（如果可用）
if hasattr(response, 'thinking'):
    print("Thinking:", response.thinking)
print("Answer:", response.text)
```

### Thinking Budget 配置

| 值 | 说明 |
|----|------|
| `0` | 禁用 thinking |
| `-1` | 动态 thinking（模型自动决定） |
| `1024-24576` | 固定 thinking token 数量 |

**SherlockOS 推荐**：`8192`（平衡质量与延迟）

## 结构化输出（JSON）

### 使用 response_schema

```python
from google import genai
from google.genai import types

# 定义输出 Schema
trajectory_schema = {
    "type": "object",
    "properties": {
        "trajectories": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "id": {"type": "string"},
                    "confidence": {"type": "number"},
                    "segments": {"type": "array"},
                    "explanation": {"type": "string"}
                }
            }
        },
        "uncertainty_areas": {"type": "array"},
        "suggestions": {"type": "array"}
    }
}

response = client.models.generate_content(
    model="gemini-2.5-flash",
    contents=prompt,
    config=types.GenerateContentConfig(
        response_mime_type="application/json",
        response_schema=trajectory_schema,
        thinking_config=types.ThinkingConfig(thinking_budget=8192)
    )
)

import json
result = json.loads(response.text)
```

## 流式输出

```python
from google import genai

client = genai.Client()

for chunk in client.models.generate_content_stream(
    model="gemini-2.5-flash",
    contents="Analyze the trajectory..."
):
    print(chunk.text, end="", flush=True)
```

## SherlockOS 集成示例

### Reasoning Worker 实现

```python
import json
from google import genai
from google.genai import types

REASONING_PROMPT = """
你是一个专业的案件分析助手。基于以下场景数据，推断嫌疑人可能的移动轨迹。

## 场景数据
{scenegraph_json}

## 约束条件
{constraints_json}

## 任务
1. 分析场景中的证据关系
2. 推断 Top-{max_trajectories} 条可能的移动轨迹
3. 为每段轨迹提供证据引用和置信度
4. 标注不确定区域
5. 给出下一步侦查建议

请严格按照 JSON 格式输出。
"""

class ReasoningWorker:
    def __init__(self):
        self.client = genai.Client()

    def run(self, input_data: dict) -> dict:
        prompt = REASONING_PROMPT.format(
            scenegraph_json=json.dumps(input_data["scenegraph"], ensure_ascii=False),
            constraints_json=json.dumps(input_data.get("constraints", []), ensure_ascii=False),
            max_trajectories=input_data.get("max_trajectories", 3)
        )

        response = self.client.models.generate_content(
            model="gemini-2.5-flash",
            contents=prompt,
            config=types.GenerateContentConfig(
                temperature=0.3,
                response_mime_type="application/json",
                thinking_config=types.ThinkingConfig(
                    thinking_budget=input_data.get("thinking_budget", 8192)
                )
            )
        )

        return {
            "result": json.loads(response.text),
            "model_stats": {
                "thinking_tokens": getattr(response, 'thinking_tokens', None),
                "output_tokens": response.usage_metadata.candidates_token_count,
            }
        }
```

## 定价与配额

### 免费层（Google AI Studio）

| 限制 | 值 |
|------|-----|
| RPM (请求/分钟) | 10 |
| TPM (Token/分钟) | 250,000 |
| RPD (请求/天) | 500 |

### 付费定价

| 模型 | 输入 (per 1M tokens) | 输出 (per 1M tokens) |
|------|---------------------|---------------------|
| Gemini 2.5 Flash | $0.15 | $0.60 |
| Gemini 2.5 Pro | $1.25 | $5.00 |

## 错误处理

```python
from google.api_core import exceptions

try:
    response = client.models.generate_content(...)
except exceptions.ResourceExhausted:
    # 429 Rate Limit - 等待重试
    pass
except exceptions.InvalidArgument:
    # 400 请求无效
    pass
except exceptions.InternalServerError:
    # 500 服务器错误 - 重试
    pass
```

## 最佳实践

1. **使用 Thinking 模式**：对于复杂推理任务，始终启用 thinking
2. **结构化输出**：使用 `response_schema` 确保输出格式一致
3. **温度设置**：推理任务使用 0.2-0.4，创意任务使用 0.7-0.9
4. **重试策略**：对 5xx 错误使用指数退避重试

## 参考链接

- [Gemini API 文档](https://ai.google.dev/gemini-api/docs)
- [Thinking 模式文档](https://ai.google.dev/gemini-api/docs/thinking)
- [Python SDK GitHub](https://github.com/googleapis/python-genai)
