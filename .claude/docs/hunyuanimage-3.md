# HunyuanImage-3.0 API Documentation

> **用途**：SherlockOS 图像理解与结构化分析（可选，Gemini Vision 的降级备选）
> **GitHub**：https://github.com/Tencent-Hunyuan/HunyuanImage-3.0
> **HuggingFace**：https://huggingface.co/tencent/HunyuanImage-3.0

## 概述

HunyuanImage-3.0 是腾讯的 80B 参数原生多模态模型，统一了图像理解与生成能力。

### 核心特点

| 特性 | 说明 |
|------|------|
| 参数量 | 80B 总参数，13B 激活参数/token |
| 架构 | MoE (64 experts) + 自回归 |
| 能力 | 图像理解 + 图像生成 + CoT 推理 |
| 文字渲染 | 中英文高质量渲染 |

### 在 SherlockOS 中的用途

主要用于**图像理解**，作为 Gemini Vision 的备选：
- 从扫描图中提取物体信息
- 分析场景布局
- 识别潜在证据

## 安装

### 环境要求

- Python 3.10+
- CUDA 12.8+
- GPU: 至少 40GB VRAM（完整版），24GB（Distil 版）

### 安装步骤

```bash
# 安装 PyTorch (CUDA 12.8)
pip install torch==2.7.1 torchvision==0.22.1 --index-url https://download.pytorch.org/whl/cu128

# 克隆仓库
git clone https://github.com/Tencent-Hunyuan/HunyuanImage-3.0
cd HunyuanImage-3.0

# 安装依赖
pip install -r requirements.txt

# 可选：性能优化（3x 加速）
pip install flash-attn==2.8.3
pip install flashinfer-python
```

### 模型下载

```bash
# 完整版
huggingface-cli download tencent/HunyuanImage-3.0 --local-dir ./HunyuanImage-3

# 蒸馏版（更快）
huggingface-cli download tencent/HunyuanImage-3.0-Instruct-Distil --local-dir ./HunyuanImage-3-Instruct-Distil
```

## 图像理解用法

### 基础图像分析

```python
from transformers import AutoModelForCausalLM, AutoTokenizer
from PIL import Image
import torch

# 加载模型
model_path = "./HunyuanImage-3"
model = AutoModelForCausalLM.from_pretrained(
    model_path,
    torch_dtype=torch.bfloat16,
    device_map="auto",
    trust_remote_code=True
)
tokenizer = AutoTokenizer.from_pretrained(model_path, trust_remote_code=True)

# 加载图像
image = Image.open("crime_scene.jpg")

# 构建对话
messages = [
    {
        "role": "user",
        "content": [
            {"type": "image", "image": image},
            {"type": "text", "text": "请详细描述这张图片中的所有物体，包括它们的位置、状态和任何异常之处。"}
        ]
    }
]

# 推理
inputs = tokenizer.apply_chat_template(messages, return_tensors="pt").to(model.device)
outputs = model.generate(inputs, max_new_tokens=1024)
response = tokenizer.decode(outputs[0], skip_special_tokens=True)

print(response)
```

### 结构化物体提取

```python
def extract_objects_from_image(model, tokenizer, image_path: str) -> list:
    """从图像中提取结构化物体信息"""

    image = Image.open(image_path)

    prompt = """分析这张场景图片，以 JSON 格式输出所有可识别的物体。

对于每个物体，提供：
- type: 物体类型（furniture/door/window/evidence_item/other）
- label: 物体名称
- position: 在图中的大致位置（left/center/right, top/middle/bottom）
- state: 状态描述
- notable: 是否有异常或值得注意的地方

输出格式：
```json
{
  "objects": [
    {"type": "...", "label": "...", "position": "...", "state": "...", "notable": "..."}
  ],
  "scene_description": "...",
  "potential_evidence": ["..."]
}
```"""

    messages = [
        {
            "role": "user",
            "content": [
                {"type": "image", "image": image},
                {"type": "text", "text": prompt}
            ]
        }
    ]

    inputs = tokenizer.apply_chat_template(messages, return_tensors="pt").to(model.device)
    outputs = model.generate(inputs, max_new_tokens=2048, temperature=0.2)
    response = tokenizer.decode(outputs[0], skip_special_tokens=True)

    # 解析 JSON
    import json
    import re
    json_match = re.search(r'```json\s*(.*?)\s*```', response, re.DOTALL)
    if json_match:
        return json.loads(json_match.group(1))

    return {"raw_response": response}
```

### CoT 推理模式

```python
def analyze_with_reasoning(model, tokenizer, image_path: str, question: str) -> dict:
    """使用 Chain-of-Thought 推理分析图像"""

    image = Image.open(image_path)

    prompt = f"""请仔细分析这张图片，并回答以下问题。

在回答之前，请先进行推理思考（用 <think> 标签包裹），然后给出最终答案。

问题：{question}

格式：
<think>
[你的推理过程]
</think>

<answer>
[最终答案]
</answer>"""

    messages = [
        {
            "role": "user",
            "content": [
                {"type": "image", "image": image},
                {"type": "text", "text": prompt}
            ]
        }
    ]

    inputs = tokenizer.apply_chat_template(messages, return_tensors="pt").to(model.device)
    outputs = model.generate(inputs, max_new_tokens=2048)
    response = tokenizer.decode(outputs[0], skip_special_tokens=True)

    # 解析思考和答案
    import re
    think_match = re.search(r'<think>(.*?)</think>', response, re.DOTALL)
    answer_match = re.search(r'<answer>(.*?)</answer>', response, re.DOTALL)

    return {
        "thinking": think_match.group(1).strip() if think_match else None,
        "answer": answer_match.group(1).strip() if answer_match else response
    }
```

## 图像生成用法

虽然 SherlockOS 主要使用 Nano Banana 进行图像生成，但 HunyuanImage-3.0 也支持：

```bash
# 命令行生成
python3 run_image_gen.py \
    --model-id ./HunyuanImage-3 \
    --prompt "A professional forensic evidence board with labeled items" \
    --seed 42 \
    --diff-infer-steps 50 \
    --image-size 1280x768 \
    --save output.png
```

### 使用蒸馏版（更快）

```bash
python3 run_image_gen.py \
    --model-id ./HunyuanImage-3-Instruct-Distil \
    --prompt "Your prompt" \
    --diff-infer-steps 8 \
    --save output.png
```

## SherlockOS 集成示例

### Image Understanding Worker

```python
from transformers import AutoModelForCausalLM, AutoTokenizer
from PIL import Image
import torch
import json
from typing import List, Dict, Any

class ImageUnderstandingWorker:
    """使用 HunyuanImage-3.0 进行图像理解"""

    def __init__(self, model_path: str = "./HunyuanImage-3"):
        self.model = AutoModelForCausalLM.from_pretrained(
            model_path,
            torch_dtype=torch.bfloat16,
            device_map="auto",
            trust_remote_code=True
        )
        self.tokenizer = AutoTokenizer.from_pretrained(
            model_path,
            trust_remote_code=True
        )

    def analyze_scene(self, image_paths: List[str]) -> Dict[str, Any]:
        """分析场景图片，提取物体和证据信息"""

        all_objects = []
        all_evidence = []

        for path in image_paths:
            result = self._analyze_single_image(path)
            all_objects.extend(result.get("objects", []))
            all_evidence.extend(result.get("potential_evidence", []))

        # 去重和合并
        return {
            "objects": self._merge_objects(all_objects),
            "potential_evidence": list(set(all_evidence)),
            "source_images": image_paths
        }

    def _analyze_single_image(self, image_path: str) -> dict:
        """分析单张图片"""

        image = Image.open(image_path)

        prompt = """分析这张犯罪现场/场景图片。

请输出：
1. 所有可识别的物体及其属性
2. 任何可能的证据或异常
3. 场景的整体布局描述

以 JSON 格式输出：
{
  "objects": [
    {
      "id": "obj_001",
      "type": "furniture|door|window|evidence_item|weapon|footprint|other",
      "label": "物体名称",
      "position_description": "位置描述",
      "state": "状态描述",
      "confidence": 0.0-1.0,
      "is_suspicious": true/false,
      "notes": "备注"
    }
  ],
  "potential_evidence": ["可能的证据描述"],
  "scene_layout": "场景布局描述"
}"""

        messages = [
            {
                "role": "user",
                "content": [
                    {"type": "image", "image": image},
                    {"type": "text", "text": prompt}
                ]
            }
        ]

        inputs = self.tokenizer.apply_chat_template(
            messages,
            return_tensors="pt"
        ).to(self.model.device)

        with torch.no_grad():
            outputs = self.model.generate(
                inputs,
                max_new_tokens=2048,
                temperature=0.2
            )

        response = self.tokenizer.decode(outputs[0], skip_special_tokens=True)

        # 解析 JSON
        try:
            import re
            json_match = re.search(r'\{[\s\S]*\}', response)
            if json_match:
                return json.loads(json_match.group())
        except json.JSONDecodeError:
            pass

        return {"raw_response": response, "objects": [], "potential_evidence": []}

    def _merge_objects(self, objects: List[dict]) -> List[dict]:
        """合并来自多张图片的物体（简化实现）"""
        # 实际实现中需要更复杂的去重逻辑
        seen = set()
        merged = []
        for obj in objects:
            key = (obj.get("type"), obj.get("label"))
            if key not in seen:
                seen.add(key)
                merged.append(obj)
        return merged
```

## API 服务（AI/ML API）

如果不想自己部署，可以使用第三方 API：

```python
import requests

response = requests.post(
    "https://api.aimlapi.com/v1/images/generations",
    headers={
        "Authorization": "Bearer YOUR_API_KEY",
        "Content-Type": "application/json"
    },
    json={
        "model": "tencent/hunyuan-image-v3",
        "prompt": "Your prompt",
        "n": 1,
        "size": "1024x1024"
    }
)

result = response.json()
```

## 性能参考

| 配置 | 理解任务 | 生成任务 |
|------|----------|----------|
| A100 80GB | ~5s | ~30s |
| RTX 4090 24GB | 需要量化 | 需要蒸馏版 |
| Distil 版 (8 步) | - | ~10s |

## 与 Gemini Vision 对比

| 特性 | HunyuanImage-3.0 | Gemini 2.5 Flash |
|------|------------------|------------------|
| 部署 | 自部署 | 云 API |
| 成本 | GPU 成本 | 按 token 计费 |
| 延迟 | 依赖硬件 | ~1-3s |
| 中文支持 | 原生优秀 | 良好 |
| 推荐场景 | 离线/高隐私 | 在线/快速迭代 |

**SherlockOS 建议**：优先使用 Gemini 2.5 Flash Vision，HunyuanImage-3.0 作为离线或高隐私场景的备选。

## 降级策略

| 情况 | 降级方案 |
|------|----------|
| GPU 不足 | 使用 Gemini Vision API |
| 模型加载失败 | 使用 AI/ML API |
| 推理超时 | 降低输出 token 数 |

## 注意事项

1. **显存需求高**：完整版需要 40GB+ VRAM
2. **首次加载慢**：模型较大，首次加载需要几分钟
3. **生成质量**：图像生成质量很高，但可能需要调整 prompt
4. **许可证**：商业使用需注意 Tencent Hunyuan Community License

## 参考链接

- [GitHub 仓库](https://github.com/Tencent-Hunyuan/HunyuanImage-3.0)
- [HuggingFace 模型](https://huggingface.co/tencent/HunyuanImage-3.0)
- [技术报告](https://huggingface.co/papers/2509.23951)
- [AI/ML API 文档](https://docs.aimlapi.com/api-references/image-models/tencent/hunyuan-image-v3-text-to-image)
