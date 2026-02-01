# Hunyuan3D-2 API Documentation

> **用途**：SherlockOS 关键物品/证据物 3D 资产生成
> **GitHub**：https://github.com/Tencent-Hunyuan/Hunyuan3D-2
> **HuggingFace**：https://huggingface.co/tencent/Hunyuan3D-2
> **Replicate API**：https://replicate.com/tencent/hunyuan3d-2

## 概述

Hunyuan3D-2 是腾讯的大规模 3D 资产生成系统，支持从文本或图像生成高分辨率纹理 3D 模型。

### 系统组成

| 组件 | 说明 |
|------|------|
| **Hunyuan3D-DiT** | Shape 生成模型 |
| **Hunyuan3D-Paint** | 纹理合成模型 |

### 模型版本

| 版本 | VRAM 需求 | 特点 |
|------|-----------|------|
| Hunyuan3D-2 | ~12GB | 完整版本 |
| Hunyuan3D-2mini | ~5GB | 轻量版本，适合消费级 GPU |
| Hunyuan3D-2-Fast | ~12GB | 推理速度提升 50% |

## 安装

### 本地部署

```bash
# 克隆仓库
git clone https://github.com/Tencent-Hunyuan/Hunyuan3D-2
cd Hunyuan3D-2

# 安装 PyTorch (CUDA 12.1)
pip install torch==2.3.0 torchvision==0.18.0 --index-url https://download.pytorch.org/whl/cu121

# 安装依赖
pip install -r requirements.txt

# 下载模型
huggingface-cli download tencent/Hunyuan3D-2 --local-dir ./models/Hunyuan3D-2
```

## 基础用法（本地 Python API）

### 图像到 3D

```python
from hy3dgen.shapegen import Hunyuan3DDiTFlowMatchingPipeline
from hy3dgen.texgen import Hunyuan3DPaintPipeline

# 1. Shape 生成
shape_pipeline = Hunyuan3DDiTFlowMatchingPipeline.from_pretrained(
    "tencent/Hunyuan3D-2",
    subfolder="hunyuan3d-dit-v2-0"
)

# 从图像生成 mesh
mesh = shape_pipeline(image="evidence_item.png")[0]
mesh.export("evidence_shape.glb")

# 2. 纹理生成
texture_pipeline = Hunyuan3DPaintPipeline.from_pretrained(
    "tencent/Hunyuan3D-2"
)

# 为 mesh 生成纹理
textured_mesh = texture_pipeline(
    mesh=mesh,
    image="evidence_item.png"
)[0]
textured_mesh.export("evidence_textured.glb")
```

### 文本到 3D

```python
from hy3dgen.shapegen import Hunyuan3DDiTFlowMatchingPipeline

pipeline = Hunyuan3DDiTFlowMatchingPipeline.from_pretrained(
    "tencent/Hunyuan3D-2",
    subfolder="hunyuan3d-dit-v2-0"
)

# 从文本生成
mesh = pipeline(
    prompt="A realistic kitchen knife with wooden handle, forensic evidence style"
)[0]

mesh.export("knife_3d.glb")
```

### 低显存模式

```python
from hy3dgen.shapegen import Hunyuan3DDiTFlowMatchingPipeline

# 使用 mini 版本
pipeline = Hunyuan3DDiTFlowMatchingPipeline.from_pretrained(
    "tencent/Hunyuan3D-2mini",
    subfolder="hunyuan3d-dit-v2-mini"
)

# 或启用低显存优化
pipeline.enable_model_cpu_offload()
pipeline.enable_attention_slicing()
```

## Replicate API 用法（云端推理）

无需本地 GPU，直接调用云端 API。

### 安装

```bash
pip install replicate
```

### 基础调用

```python
import replicate
import base64
import requests

# 设置 API Token
# export REPLICATE_API_TOKEN="your-token"

# 读取图像
with open("evidence_photo.jpg", "rb") as f:
    image_data = base64.b64encode(f.read()).decode()

# 调用 API
output = replicate.run(
    "tencent/hunyuan3d-2:latest",
    input={
        "image": f"data:image/jpeg;base64,{image_data}",
        "texture": True,
        "seed": 42,
        "output_format": "glb"
    }
)

# 下载结果
response = requests.get(output["mesh"])
with open("output.glb", "wb") as f:
    f.write(response.content)
```

### 完整参数

```python
output = replicate.run(
    "tencent/hunyuan3d-2:latest",
    input={
        "image": image_data_uri,        # 输入图像 (data URI 或 URL)
        "prompt": "optional text",       # 可选文本描述
        "texture": True,                 # 是否生成纹理
        "seed": 42,                      # 随机种子
        "num_inference_steps": 50,       # 推理步数
        "guidance_scale": 7.5,           # 引导强度
        "output_format": "glb",          # 输出格式: glb, obj, ply
        "texture_resolution": 1024,      # 纹理分辨率
    }
)
```

## FastAPI Server 模式

### 启动服务器

```bash
python api_server.py --port 7860
```

### HTTP 调用

```python
import requests
import base64

# 准备请求
with open("evidence.jpg", "rb") as f:
    image_b64 = base64.b64encode(f.read()).decode()

request_data = {
    "image": image_b64,
    "texture": True,
    "seed": 42,
    "output_format": "glb"
}

# 发送请求
response = requests.post(
    "http://localhost:7860/generate",
    json=request_data,
    timeout=300  # 3D 生成可能需要较长时间
)

if response.status_code == 200:
    with open("output.glb", "wb") as f:
        f.write(response.content)
```

## SherlockOS 集成示例

### 3D Asset Worker

```python
import replicate
import requests
import base64
from typing import Optional
import os

class Asset3DWorker:
    """使用 Replicate API 生成 3D 资产"""

    def __init__(self):
        self.api_token = os.environ.get("REPLICATE_API_TOKEN")

    def generate_from_image(
        self,
        image_path: str,
        description: Optional[str] = None,
        with_texture: bool = True,
        seed: int = 42
    ) -> dict:
        """从图像生成 3D 模型"""

        # 读取图像
        with open(image_path, "rb") as f:
            image_data = base64.b64encode(f.read()).decode()

        # 确定 MIME 类型
        ext = image_path.lower().split(".")[-1]
        mime_type = {"jpg": "jpeg", "jpeg": "jpeg", "png": "png"}.get(ext, "jpeg")
        data_uri = f"data:image/{mime_type};base64,{image_data}"

        # 调用 Replicate API
        input_params = {
            "image": data_uri,
            "texture": with_texture,
            "seed": seed,
            "output_format": "glb"
        }

        if description:
            input_params["prompt"] = description

        try:
            output = replicate.run(
                "tencent/hunyuan3d-2:latest",
                input=input_params
            )

            # 下载生成的模型
            mesh_url = output.get("mesh") or output
            response = requests.get(mesh_url)

            return {
                "success": True,
                "mesh_bytes": response.content,
                "format": "glb",
                "has_texture": with_texture
            }

        except Exception as e:
            return {
                "success": False,
                "error": str(e)
            }

    def generate_evidence_item(
        self,
        image_path: str,
        item_type: str,
        case_id: str
    ) -> dict:
        """生成证据物 3D 模型"""

        # 根据类型构建描述
        descriptions = {
            "weapon": "forensic evidence weapon, high detail, realistic",
            "footprint": "3D footprint impression, forensic quality",
            "tool": "forensic tool mark evidence, detailed surface",
            "other": "forensic evidence item, detailed and realistic"
        }

        description = descriptions.get(item_type, descriptions["other"])

        result = self.generate_from_image(
            image_path=image_path,
            description=description,
            with_texture=True
        )

        if result["success"]:
            # 保存到临时位置
            import uuid
            output_path = f"/tmp/{case_id}_{uuid.uuid4().hex}.glb"
            with open(output_path, "wb") as f:
                f.write(result["mesh_bytes"])
            result["output_path"] = output_path

        return result
```

## Gradio Demo

```bash
# 启动完整版
python3 gradio_app.py \
    --model_path tencent/Hunyuan3D-2 \
    --subfolder hunyuan3d-dit-v2-0 \
    --texgen_model_path tencent/Hunyuan3D-2

# 启动轻量版（低显存）
python3 gradio_app.py \
    --model_path tencent/Hunyuan3D-2mini \
    --subfolder hunyuan3d-dit-v2-mini \
    --texgen_model_path tencent/Hunyuan3D-2 \
    --low_vram_mode
```

## Blender 集成

```bash
# 1. 启动 API 服务器
python api_server.py --port 7860

# 2. 在 Blender 中安装插件
# - 下载 blender_addon.zip
# - Blender -> Edit -> Preferences -> Add-ons -> Install
# - 启用 Hunyuan3D Addon
```

## 定价与配额

### Tencent Cloud

| 项目 | 值 |
|------|-----|
| 免费额度 | 20 次/天 |
| 企业额度 | 200 credits |

### Replicate

| 项目 | 值 |
|------|-----|
| 价格 | ~$0.05-0.10/次（按 GPU 时间计费） |
| 超时 | 最长 5 分钟 |

## 输出格式

### GLB (推荐)

```python
mesh.export("output.glb")  # 包含纹理的二进制 glTF
```

### OBJ

```python
mesh.export("output.obj")  # 需要单独的 .mtl 和纹理文件
```

### PLY

```python
mesh.export("output.ply")  # 点云格式，带顶点颜色
```

## 性能参考

| 配置 | Shape 生成 | 纹理生成 | 总计 |
|------|------------|----------|------|
| RTX 4090 | ~30s | ~20s | ~50s |
| RTX 3080 | ~60s | ~40s | ~100s |
| Replicate API | ~60-120s | 包含 | ~60-120s |

## 注意事项

1. **输入图像质量**：清晰的物体照片效果最好，避免复杂背景
2. **物体完整性**：尽量提供物体的完整视角
3. **纹理生成**：对于需要精确纹理的场景，建议提供多角度图像
4. **模型精度**：生成的模型适合可视化，不适合精确测量

## 降级策略

| 情况 | 降级方案 |
|------|----------|
| Replicate 不可用 | 使用本地 mini 版本 |
| 显存不足 | 使用 Hunyuan3D-2mini |
| 生成失败 | 返回简化几何体占位符 |

## 参考链接

- [GitHub 仓库](https://github.com/Tencent-Hunyuan/Hunyuan3D-2)
- [HuggingFace 模型](https://huggingface.co/tencent/Hunyuan3D-2)
- [Replicate API](https://replicate.com/tencent/hunyuan3d-2)
- [Hunyuan3D-2.1 (新版)](https://github.com/Tencent-Hunyuan/Hunyuan3D-2.1)
