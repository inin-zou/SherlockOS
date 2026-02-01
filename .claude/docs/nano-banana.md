# Nano Banana / Nano Banana Pro API Documentation

> **用途**：SherlockOS 嫌疑人画像生成、Evidence Board 渲染
> **官方文档**：https://ai.google.dev/gemini-api/docs/image-generation

## 概述

Nano Banana 是 Gemini 的原生图像生成模型系列，支持文本到图像、图像编辑等功能。

| 模型 | Model ID | 特点 | 适用场景 |
|------|----------|------|----------|
| **Nano Banana** | `gemini-2.5-flash-image` | 快速、低成本 | 画像迭代、快速预览 |
| **Nano Banana Pro** | `gemini-3-pro-image-preview` | 高清、文字清晰 | 最终报告、4K 输出 |

## 安装

```bash
pip install google-genai>=1.52.0
pip install Pillow  # 图像处理
```

## 环境配置

```bash
export GEMINI_API_KEY="your-api-key"
```

## 基础用法

### 文本到图像

```python
from google import genai
from PIL import Image
import io

client = genai.Client()

prompt = "A professional police sketch portrait of a male suspect, approximately 30 years old, short black hair, wearing glasses, neutral expression, high detail"

response = client.models.generate_content(
    model="gemini-2.5-flash-image",  # Nano Banana
    contents=[prompt],
)

# 提取生成的图像
for part in response.parts:
    if part.inline_data is not None:
        image = part.as_image()
        image.save("portrait.png")
        print("Image saved!")
    elif part.text is not None:
        print("Text response:", part.text)
```

### 图像编辑（带参考图）

```python
from google import genai
from PIL import Image

client = genai.Client()

# 加载参考图像
reference_image = Image.open("previous_portrait.png")

prompt = "Edit this portrait to add a beard and remove the glasses, keep other features the same"

response = client.models.generate_content(
    model="gemini-2.5-flash-image",
    contents=[prompt, reference_image],
)

for part in response.parts:
    if part.inline_data is not None:
        edited_image = part.as_image()
        edited_image.save("edited_portrait.png")
```

## 高清输出（Nano Banana Pro）

```python
from google import genai
from google.genai import types

client = genai.Client()

prompt = "Create a detailed evidence board showing the crime scene layout with labeled evidence markers, professional forensic style, 4K quality"

response = client.models.generate_content(
    model="gemini-3-pro-image-preview",  # Nano Banana Pro
    contents=[prompt],
    config=types.GenerateContentConfig(
        # Pro 模型支持更高分辨率
        generation_config={
            "response_modalities": ["IMAGE"],
        }
    )
)

for part in response.parts:
    if part.inline_data is not None:
        image = part.as_image()
        image.save("evidence_board_4k.png")
```

## SherlockOS 集成示例

### ImageGen Worker 实现

```python
import io
import base64
from google import genai
from PIL import Image
from typing import Optional

class ImageGenWorker:
    def __init__(self):
        self.client = genai.Client()

    def generate_portrait(
        self,
        attributes: dict,
        reference_image_path: Optional[str] = None,
        resolution: str = "1k"
    ) -> dict:
        """生成嫌疑人画像"""

        # 构建 prompt
        prompt = self._build_portrait_prompt(attributes)

        # 选择模型
        model = "gemini-2.5-flash-image" if resolution == "1k" else "gemini-3-pro-image-preview"

        # 准备输入
        contents = [prompt]
        if reference_image_path:
            ref_image = Image.open(reference_image_path)
            contents.append(ref_image)

        # 生成
        response = self.client.models.generate_content(
            model=model,
            contents=contents,
        )

        # 处理输出
        for part in response.parts:
            if part.inline_data is not None:
                image = part.as_image()

                # 转为 bytes 用于存储
                buffer = io.BytesIO()
                image.save(buffer, format="PNG")
                image_bytes = buffer.getvalue()

                return {
                    "image_bytes": image_bytes,
                    "width": image.width,
                    "height": image.height,
                    "model_used": "nano-banana" if resolution == "1k" else "nano-banana-pro"
                }

        raise Exception("No image generated")

    def _build_portrait_prompt(self, attrs: dict) -> str:
        """根据结构化属性构建 prompt"""
        parts = ["Professional police sketch portrait of a suspect:"]

        if "age_range" in attrs:
            parts.append(f"- Age: {attrs['age_range']['min']}-{attrs['age_range']['max']} years old")

        if "height_range_cm" in attrs:
            parts.append(f"- Height: {attrs['height_range_cm']['min']}-{attrs['height_range_cm']['max']} cm")

        if "build" in attrs:
            parts.append(f"- Build: {attrs['build']['value']}")

        if "skin_tone" in attrs:
            parts.append(f"- Skin tone: {attrs['skin_tone']['value']}")

        if "hair" in attrs:
            hair = attrs["hair"]
            parts.append(f"- Hair: {hair.get('color', '')} {hair.get('style', '')}")

        if "glasses" in attrs:
            parts.append(f"- Glasses: {attrs['glasses']['type']}")

        if "facial_hair" in attrs:
            parts.append(f"- Facial hair: {attrs['facial_hair']['type']}")

        if "distinctive_features" in attrs:
            for feature in attrs["distinctive_features"]:
                parts.append(f"- Distinctive feature: {feature['description']}")

        parts.append("\nStyle: Realistic pencil sketch, neutral expression, front-facing, high detail, professional forensic quality")

        return "\n".join(parts)

    def generate_evidence_board(
        self,
        objects: list,
        layout: str = "grid"
    ) -> dict:
        """生成证据板"""

        prompt = f"""Create a professional evidence board with the following items arranged in a {layout} layout:

Items:
{chr(10).join([f"- {obj['label']}: {obj.get('description', '')}" for obj in objects])}

Style: Clean forensic evidence board, labeled markers, professional presentation, white background"""

        response = self.client.models.generate_content(
            model="gemini-3-pro-image-preview",  # 使用 Pro 获得更好质量
            contents=[prompt],
        )

        for part in response.parts:
            if part.inline_data is not None:
                image = part.as_image()
                buffer = io.BytesIO()
                image.save(buffer, format="PNG")
                return {
                    "image_bytes": buffer.getvalue(),
                    "width": image.width,
                    "height": image.height,
                    "model_used": "nano-banana-pro"
                }

        raise Exception("No image generated")
```

## 多版本画像对比

```python
def generate_portrait_variants(self, base_attributes: dict, variations: list) -> list:
    """生成多个画像变体用于对比"""
    results = []

    for variation in variations:
        # 合并属性
        attrs = {**base_attributes, **variation}

        result = self.generate_portrait(attrs, resolution="1k")
        result["variation"] = variation
        results.append(result)

    return results

# 使用示例
variations = [
    {"facial_hair": {"type": "beard"}},
    {"facial_hair": {"type": "none"}},
    {"glasses": {"type": "none"}},
]

portraits = worker.generate_portrait_variants(base_attrs, variations)
```

## 定价

| 模型 | 分辨率 | 价格 |
|------|--------|------|
| Nano Banana | 1K | ~$0.04/image |
| Nano Banana Pro | 2K | ~$0.134/image |
| Nano Banana Pro | 4K | ~$0.24/image |

## 配额（免费层）

| 限制 | 值 |
|------|-----|
| RPD (请求/天) | 500 |
| TPM (Token/分钟) | 250,000 |

## 注意事项

1. **SynthID 水印**：所有生成的图像包含不可见的 SynthID 水印
2. **内容安全**：模型会拒绝生成有害内容
3. **人脸生成**：对于真实人脸生成有限制，建议使用"sketch"风格
4. **文字渲染**：Nano Banana Pro 的文字渲染准确率达 94%

## 错误处理

```python
from google.api_core import exceptions

try:
    response = client.models.generate_content(...)
except exceptions.InvalidArgument as e:
    if "safety" in str(e).lower():
        print("Content blocked by safety filters")
    else:
        raise
except exceptions.ResourceExhausted:
    print("Rate limit exceeded, please wait")
```

## 最佳实践

1. **迭代用 Nano Banana**：快速预览使用 `gemini-2.5-flash-image`
2. **最终输出用 Pro**：报告和打印使用 `gemini-3-pro-image-preview`
3. **结构化 Prompt**：使用清晰的列表格式描述特征
4. **参考图编辑**：需要保持一致性时使用参考图

## 参考链接

- [Nano Banana 官方文档](https://ai.google.dev/gemini-api/docs/image-generation)
- [Nano Banana Hackathon Kit](https://github.com/google-gemini/nano-banana-hackathon-kit)
- [Google AI Studio](https://aistudio.google.com/)
